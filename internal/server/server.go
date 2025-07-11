package server

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/codecrafters-redis-go/internal/commands"
	"github.com/codecrafters-redis-go/internal/config"
	"github.com/codecrafters-redis-go/internal/logger"
	"github.com/codecrafters-redis-go/internal/rdb"
	"github.com/codecrafters-redis-go/internal/replication"
	"github.com/codecrafters-redis-go/internal/resp"
	"github.com/codecrafters-redis-go/internal/storage"
)

// Replica represents a connected replica
type Replica struct {
	conn    net.Conn
	encoder *resp.Encoder
}

// Server represents a Redis server
type Server struct {
	addr              string
	config            *config.Config
	storage           *storage.Storage
	registry          *commands.Registry
	listener          net.Listener
	wg                sync.WaitGroup
	shutdown          chan struct{}
	replicationClient *replication.Client
	replicas          []*Replica
	replicasMu        sync.RWMutex
}

// New creates a new Redis server
func New(cfg *config.Config) *Server {
	store := storage.New()
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	server := &Server{
		addr:     addr,
		config:   cfg,
		storage:  store,
		registry: commands.NewRegistry(cfg, store),
		shutdown: make(chan struct{}),
		replicas: make([]*Replica, 0),
	}

	// Set the propagation function in the registry
	server.registry.SetPropagateFunc(server.propagateCommand)

	// Set the server reference in the registry
	server.registry.SetServer(server)

	return server
}

// Start begins listening for connections
func (server *Server) Start() error {
	// Load RDB file if it exists
	if err := rdb.LoadFile(server.config.Dir, server.config.DBFilename, server.storage); err != nil {
		logger.Warn("Failed to load RDB file: %v", err)
	}

	listener, err := net.Listen("tcp", server.addr)
	if err != nil {
		return fmt.Errorf("failed to bind to %s: %w", server.addr, err)
	}

	server.listener = listener
	logger.Info("Redis server listening on %s", server.addr)

	// Accept connections in a goroutine
	go server.acceptConnections()

		// If configured as replica, connect to master
	if server.config.IsReplica() {
		host, port := server.config.GetReplicaInfo()
		if host != "" && port != "" {
			server.replicationClient = replication.NewClient(host, port, server.config.Port)

			// Connect to master in a goroutine
			go func() {
				if err := server.connectToMaster(); err != nil {
					logger.Error("Failed to connect to master: %v", err)
				}
			}()
		}
	}

	return nil
}

// Stop gracefully shuts down the server
func (server *Server) Stop() error {
	close(server.shutdown)

	if server.listener != nil {
		server.listener.Close()
	}

	// Close replication client if exists
	if server.replicationClient != nil {
		server.replicationClient.Close()
	}

	// Wait for all connections to finish
	server.wg.Wait()

	// Close storage to stop background cleanup
	server.storage.Close()

	logger.Info("Server stopped gracefully")
	return nil
}

// Wait blocks until the server is shut down
func (server *Server) Wait() {
	<-server.shutdown
}

func (server *Server) acceptConnections() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			select {
			case <-server.shutdown:
				return
			default:
				logger.Error("Error accepting connection: %v", err)
				continue
			}
		}

		logger.Debug("Accepted connection from %s", conn.RemoteAddr())
		server.wg.Add(1)
		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		server.wg.Done()
		// Remove replica if this was a replica connection
		server.removeReplica(conn)
		logger.Debug("Closed connection from %s", conn.RemoteAddr())
	}()

	parser := resp.NewParser(conn)
	encoder := resp.NewEncoder(conn)
	isReplica := false

	for {
		// Check for shutdown
		select {
		case <-server.shutdown:
			return
		default:
		}

		// Parse the next command
		value, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				// Client disconnected
				return
			}
			// Send error response
			encoder.Encode(resp.ErrorValue("ERR " + err.Error()))
			continue
		}

				// Handle the command
		cmdName, _ := value.GetCommand()
		logger.Debug("Handling command: %s", cmdName)
		response := server.registry.HandleCommand(value)

		// Special handling for PSYNC command
		if strings.ToUpper(cmdName) == "PSYNC" {
			// Check if this is a FULLRESYNC response
			if response.Type == resp.SimpleString && strings.HasPrefix(response.Str, "FULLRESYNC") {
				// Send the FULLRESYNC response first
				if err := encoder.Encode(response); err != nil {
					logger.Error("Error sending FULLRESYNC response: %v", err)
					return
				}

				// Send empty RDB file as bulk string
				emptyRDB := server.getEmptyRDB()
				logger.Debug("Sending RDB file: %d bytes", len(emptyRDB))

				// Send RDB as bulk string directly to connection
				// without the trailing CRLF (non-standard RESP for replication)
				header := fmt.Sprintf("$%d\r\n", len(emptyRDB))
				if _, err := conn.Write([]byte(header)); err != nil {
					logger.Error("Error sending RDB header: %v", err)
					return
				}

				// Send RDB data
				if _, err := conn.Write(emptyRDB); err != nil {
					logger.Error("Error sending RDB data: %v", err)
					return
				}

				// Note: NOT sending trailing CRLF as expected by replication protocol
				logger.Debug("Successfully sent RDB file without trailing CRLF")

				// Mark this connection as a replica
				isReplica = true
				server.addReplica(conn)
				continue
			}
		}

		// Send the response
		logger.Debug("Sending normal response for command: %s", cmdName)
		if err := encoder.Encode(response); err != nil {
			logger.Error("Error sending response: %v", err)
			return
		}

		// Propagate write commands to replicas (only if this is not a replica connection)
		if !isReplica && server.shouldPropagate(cmdName) && response.Type != resp.Error {
			logger.Debug("Propagating command %s to replicas", cmdName)
			server.propagateCommand(value)
		}
	}
}

// RegisterCommand adds a custom command implementation
func (server *Server) RegisterCommand(cmd commands.Command) {
	server.registry.RegisterCommand(cmd)
}

// addReplica adds a new replica to the server's replica list
func (server *Server) addReplica(conn net.Conn) {
	server.replicasMu.Lock()
	defer server.replicasMu.Unlock()

	replica := &Replica{
		conn:    conn,
		encoder: resp.NewEncoder(conn),
	}
	server.replicas = append(server.replicas, replica)
	logger.Info("Added new replica: %s", conn.RemoteAddr())
}

// removeReplica removes a replica from the server's replica list
func (server *Server) removeReplica(conn net.Conn) {
	server.replicasMu.Lock()
	defer server.replicasMu.Unlock()

	for i, replica := range server.replicas {
		if replica.conn == conn {
			server.replicas = append(server.replicas[:i], server.replicas[i+1:]...)
			logger.Info("Removed replica: %s", conn.RemoteAddr())
			break
		}
	}
}

// GetReplicas returns a copy of the current replicas list
// Implements commands.ServerAccessor interface
func (server *Server) GetReplicas() []interface{} {
	server.replicasMu.RLock()
	defer server.replicasMu.RUnlock()

	// Return as []interface{} to implement ServerAccessor
	replicas := make([]interface{}, len(server.replicas))
	for i, r := range server.replicas {
		replicas[i] = r
	}
	return replicas
}

// propagateCommand sends a command to all connected replicas
func (server *Server) propagateCommand(command resp.Value) {
	server.replicasMu.RLock()
	defer server.replicasMu.RUnlock()

	for _, replica := range server.replicas {
		if err := replica.encoder.Encode(command); err != nil {
			logger.Error("Failed to propagate command to replica %s: %v", replica.conn.RemoteAddr(), err)
			// TODO: Remove failed replica
		}
	}
}

// shouldPropagate returns true if the command should be propagated to replicas
func (server *Server) shouldPropagate(cmdName string) bool {
	// List of write commands that should be propagated
	writeCommands := map[string]bool{
		"SET":    true,
		"DEL":    true,
		"EXPIRE": true,
		"INCR":   true,
		"DECR":   true,
		"RPUSH":  true,
		"LPUSH":  true,
		"SADD":   true,
		"SREM":   true,
		"HSET":   true,
		"HDEL":   true,
	}

	return writeCommands[strings.ToUpper(cmdName)]
}

// connectToMaster establishes connection to master and performs handshake
func (server *Server) connectToMaster() error {
	logger.Debug("connectToMaster started")

	// Connect to master
	if err := server.replicationClient.Connect(); err != nil {
		return err
	}

	// Perform handshake
	logger.Debug("Starting handshake...")
	if err := server.replicationClient.Handshake(); err != nil {
		return err
	}
	logger.Debug("Handshake completed, starting processReplicationStream...")

	// Start listening for commands from master immediately (no goroutine delay)
	// This will block, so the original goroutine in Start() serves this purpose
	server.processReplicationStream()

	return nil
}

// processReplicationStream continuously reads and executes commands from master
func (server *Server) processReplicationStream() {
	logger.Info("Started processing replication stream from master")

	// Add a debug log to see if we're ready immediately
	logger.Debug("Ready to receive commands from master")

	for {
		// Check for shutdown
		select {
		case <-server.shutdown:
			return
		default:
		}

		// Listen for command from master
		command, err := server.replicationClient.ListenForCommands()
		if err != nil {
			if err == io.EOF {
				logger.Warn("Master connection closed")
				return
			}
			logger.Error("Error reading command from master: %v", err)
			continue
		}

		// Execute the command locally
		cmdName, cmdErr := command.GetCommand()
		if cmdErr != nil {
			logger.Error("Error getting command name: %v", cmdErr)
			continue
		}
		args := command.GetArgs()
		logger.Debug("Received command from master: %s", cmdName)

		// Special handling for REPLCONF GETACK - send ACK before updating offset
		if strings.ToUpper(cmdName) == "REPLCONF" && len(args) > 0 && strings.ToUpper(args[0]) == "GETACK" {
			logger.Debug("Received REPLCONF GETACK, sending ACK")
			// Send ACK with current offset (before processing this command)
			if err := server.replicationClient.SendReplConfAck(); err != nil {
				logger.Error("Failed to send REPLCONF ACK: %v", err)
			}
			// Now update the offset for this command
			server.replicationClient.ProcessCommand(command)
			continue
		}

		// For all other commands, update offset first
		server.replicationClient.ProcessCommand(command)

		// Execute command through registry (this will update local storage)
		response := server.registry.HandleCommand(command)

		// Log any errors but don't stop replication
		if response.Type == resp.Error {
			logger.Error("Error executing replicated command %s: %s", cmdName, response.Str)
		} else {
			logger.Debug("Successfully executed replicated command: %s", cmdName)
		}
	}
}

// getEmptyRDB returns a minimal valid RDB file
func (server *Server) getEmptyRDB() []byte {
	// Minimal RDB format:
	// - Magic string "REDIS" (5 bytes)
	// - Version "0003" (4 bytes)
	// - EOF marker 0xFF (1 byte)
	// No checksum for version 3

	rdb := make([]byte, 0, 10)

	// Magic string
	rdb = append(rdb, []byte("REDIS")...)

	// Version (RDB version 3)
	rdb = append(rdb, []byte("0003")...)

	// EOF marker
	rdb = append(rdb, 0xFF)

	return rdb
}
