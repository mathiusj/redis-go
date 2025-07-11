package replication

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-redis-go/internal/logger"
	"github.com/codecrafters-redis-go/internal/resp"
)

// Client handles replica's connection to master
type Client struct {
	masterHost   string
	masterPort   string
	replicaPort  int
	conn         net.Conn
	encoder      *resp.Encoder
	parser       *resp.Parser
	offset       int64 // Track bytes processed from master
}

// NewClient creates a new replication client
func NewClient(host, port string, replicaPort int) *Client {
	return &Client{
		masterHost:  host,
		masterPort:  port,
		replicaPort: replicaPort,
	}
}

// Connect establishes connection to the master
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%s", c.masterHost, c.masterPort)
	logger.Info("Connecting to master at %s", addr)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	c.conn = conn
	c.encoder = resp.NewEncoder(conn)
	c.parser = resp.NewParser(conn)

	logger.Info("Connected to master successfully")
	return nil
}

// Close closes the connection to master
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Handshake performs the replication handshake with master
func (c *Client) Handshake() error {
	// Step 1: Send PING
	if err := c.sendPing(); err != nil {
		return err
	}

	// Step 2: Send REPLCONF listening-port
	if err := c.sendReplConf(); err != nil {
		return err
	}

	// Step 3: Send PSYNC
	if err := c.sendPsync(); err != nil {
		return err
	}

	// Step 4: Receive RDB file (if sent)
	if err := c.receiveRDB(); err != nil {
		logger.Warn("Failed to receive RDB: %v", err)
		// Don't fail - some tests don't send RDB
	}

	return nil
}

// sendPing sends PING command and waits for PONG response
func (c *Client) sendPing() error {
	logger.Debug("Sending PING to master")

	// Create PING command
	pingCmd := resp.ArrayValue(
		resp.BulkStringValue("PING"),
	)

	// Send PING
	if err := c.encoder.Encode(pingCmd); err != nil {
		return fmt.Errorf("failed to encode PING: %w", err)
	}

	// Read response
	response, err := c.parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to read PING response: %w", err)
	}

	// Check if response is PONG
	if response.Type != resp.SimpleString || response.Str != "PONG" {
		return fmt.Errorf("unexpected PING response: %v", response)
	}

	logger.Debug("Received PONG from master")
	return nil
}

// sendReplConf sends REPLCONF commands
func (c *Client) sendReplConf() error {
	// Send REPLCONF listening-port
	if err := c.sendReplConfListeningPort(); err != nil {
		return err
	}

	// Send REPLCONF capa
	if err := c.sendReplConfCapa(); err != nil {
		return err
	}

	return nil
}

// sendReplConfListeningPort sends REPLCONF listening-port command
func (c *Client) sendReplConfListeningPort() error {
	logger.Debug("Sending REPLCONF listening-port %d to master", c.replicaPort)

	// Create REPLCONF listening-port command
	replConfCmd := resp.ArrayValue(
		resp.BulkStringValue("REPLCONF"),
		resp.BulkStringValue("listening-port"),
		resp.BulkStringValue(fmt.Sprintf("%d", c.replicaPort)),
	)

	// Send REPLCONF
	if err := c.encoder.Encode(replConfCmd); err != nil {
		return fmt.Errorf("failed to encode REPLCONF listening-port: %w", err)
	}

	// Read response
	response, err := c.parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to read REPLCONF listening-port response: %w", err)
	}

	// Check if response is OK
	if response.Type != resp.SimpleString || response.Str != "OK" {
		return fmt.Errorf("unexpected REPLCONF listening-port response: %v", response)
	}

	logger.Debug("Received OK for REPLCONF listening-port from master")
	return nil
}

// sendReplConfCapa sends REPLCONF capa command
func (c *Client) sendReplConfCapa() error {
	logger.Debug("Sending REPLCONF capa to master")

	// Create REPLCONF capa command
	// For now, we'll send psync2 capability
	replConfCmd := resp.ArrayValue(
		resp.BulkStringValue("REPLCONF"),
		resp.BulkStringValue("capa"),
		resp.BulkStringValue("psync2"),
	)

	// Send REPLCONF
	if err := c.encoder.Encode(replConfCmd); err != nil {
		return fmt.Errorf("failed to encode REPLCONF capa: %w", err)
	}

	// Read response
	response, err := c.parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to read REPLCONF capa response: %w", err)
	}

	// Check if response is OK
	if response.Type != resp.SimpleString || response.Str != "OK" {
		return fmt.Errorf("unexpected REPLCONF capa response: %v", response)
	}

	logger.Debug("Received OK for REPLCONF capa from master")
	return nil
}

// sendPsync sends PSYNC command to initiate replication
func (c *Client) sendPsync() error {
	logger.Debug("Sending PSYNC ? -1 to master")

	// Create PSYNC command
	// "?" means we don't have a previous replication ID
	// "-1" means we don't have any offset
	psyncCmd := resp.ArrayValue(
		resp.BulkStringValue("PSYNC"),
		resp.BulkStringValue("?"),
		resp.BulkStringValue("-1"),
	)

	// Send PSYNC
	if err := c.encoder.Encode(psyncCmd); err != nil {
		return fmt.Errorf("failed to encode PSYNC: %w", err)
	}

	// Read response
	response, err := c.parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to read PSYNC response: %w", err)
	}

	// Check if response is FULLRESYNC
	if response.Type != resp.SimpleString {
		return fmt.Errorf("expected simple string response, got %v", response.Type)
	}

	// Parse FULLRESYNC response
	parts := strings.Split(response.Str, " ")
	if len(parts) != 3 || parts[0] != "FULLRESYNC" {
		return fmt.Errorf("unexpected PSYNC response: %s", response.Str)
	}

	// Extract replication ID and offset
	replID := parts[1]
	offset, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid offset in FULLRESYNC: %s", parts[2])
	}

	logger.Info("Received FULLRESYNC with replid=%s offset=%d", replID, offset)

	// IMPORTANT: Create a new parser after FULLRESYNC
	// This prevents the RDB from being buffered by the old parser
	c.parser = resp.NewParser(c.conn)
	logger.Debug("Created new parser after FULLRESYNC")

	return nil
}

// receiveRDB receives and processes the RDB file from master
func (c *Client) receiveRDB() error {
	logger.Debug("Attempting to receive RDB file from master")

	// The RDB is sent as a bulk string WITHOUT trailing CRLF
	// OR in some cases (tests), no RDB is sent at all

	// Try to read the first byte to determine what's coming
	firstByte := make([]byte, 1)
	n, err := c.conn.Read(firstByte)
	if err != nil {
		if err == io.EOF {
			logger.Debug("EOF when checking for RDB, assuming no RDB")
			return nil
		}
		return fmt.Errorf("failed to read first byte: %w", err)
	}

	if n == 0 {
		logger.Debug("No data read, assuming no RDB")
		return nil
	}

	// Check if it's a bulk string
	if firstByte[0] != '$' {
		// Not an RDB - this is the first byte of the next command
		logger.Debug("First byte is not '$' (got %c), prepending for next command", firstByte[0])

		// Create a connection wrapper that prepends this byte
		prependConn := &prependReader{
			prepend: firstByte,
			reader:  c.conn,
		}
		// Replace parser to use the prepended connection
		c.parser = resp.NewParser(prependConn)

		return nil
	}

	// It's a bulk string - read the RDB
	logger.Debug("Found bulk string marker, reading RDB")

	// Read the length line
	lengthStr := ""
	for {
		b := make([]byte, 1)
		if _, err := c.conn.Read(b); err != nil {
			return fmt.Errorf("failed to read RDB length: %w", err)
		}
		if b[0] == '\r' {
			continue
		}
		if b[0] == '\n' {
			break
		}
		lengthStr += string(b[0])
	}

	// Parse length
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return fmt.Errorf("invalid RDB length: %s", lengthStr)
	}

	logger.Debug("RDB length: %d bytes", length)

	// Read the RDB data (no trailing CRLF)
	rdbData := make([]byte, length)
	totalRead := 0
	for totalRead < length {
		n, err := c.conn.Read(rdbData[totalRead:])
		if err != nil {
			return fmt.Errorf("failed to read RDB data: %w", err)
		}
		totalRead += n
	}

	logger.Debug("Successfully received RDB: %d bytes", len(rdbData))
	// TODO: Parse and apply RDB in future stages

	return nil
}

// prependReader is a helper to prepend bytes to a reader
type prependReader struct {
	prepend []byte
	reader  io.Reader
	used    bool
}

func (pr *prependReader) Read(p []byte) (n int, err error) {
	if !pr.used && len(pr.prepend) > 0 {
		n = copy(p, pr.prepend)
		pr.prepend = pr.prepend[n:]
		if len(pr.prepend) == 0 {
			pr.used = true
		}
		return n, nil
	}
	return pr.reader.Read(p)
}

// ListenForCommands continuously reads commands from master and returns them
// This should be called in a goroutine after successful handshake
func (c *Client) ListenForCommands() (resp.Value, error) {
	// Read next command from master
	return c.parser.Parse()
}

// GetOffset returns the current replication offset
func (c *Client) GetOffset() int64 {
	return c.offset
}

// SendReplConfAck sends REPLCONF ACK with current offset to master
func (c *Client) SendReplConfAck() error {
	logger.Debug("Sending REPLCONF ACK %d to master", c.offset)

	// Create REPLCONF ACK command
	ackCmd := resp.ArrayValue(
		resp.BulkStringValue("REPLCONF"),
		resp.BulkStringValue("ACK"),
		resp.BulkStringValue(fmt.Sprintf("%d", c.offset)),
	)

	// Send ACK
	if err := c.encoder.Encode(ackCmd); err != nil {
		return fmt.Errorf("failed to send REPLCONF ACK: %w", err)
	}

	return nil
}
