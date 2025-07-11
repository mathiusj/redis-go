package replication

import (
	"fmt"
	"net"
	"time"

	"github.com/codecrafters-redis-go/internal/logger"
	"github.com/codecrafters-redis-go/internal/resp"
)

// Client handles replica's connection to master
type Client struct {
	masterHost string
	masterPort string
	conn       net.Conn
	encoder    *resp.Encoder
	parser     *resp.Parser
}

// NewClient creates a new replication client
func NewClient(host, port string) *Client {
	return &Client{
		masterHost: host,
		masterPort: port,
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

// Handshake performs the replication handshake with the master
func (c *Client) Handshake() error {
	// Step 1: Send PING
	if err := c.sendPing(); err != nil {
		return fmt.Errorf("failed to send PING: %w", err)
	}

	// TODO: Add more handshake steps (REPLCONF, PSYNC) in future stages

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
