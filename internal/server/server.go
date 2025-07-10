package server

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/codecrafters-redis-go/internal/handlers"
	"github.com/codecrafters-redis-go/internal/resp"
)

// Server represents a Redis server
type Server struct {
	addr     string
	registry *handlers.Registry
	listener net.Listener
	wg       sync.WaitGroup
	shutdown chan struct{}
}

// New creates a new Redis server
func New(addr string) *Server {
	return &Server{
		addr:     addr,
		registry: handlers.NewRegistry(),
		shutdown: make(chan struct{}),
	}
}

// Start begins listening for connections
func (server *Server) Start() error {
	listener, err := net.Listen("tcp", server.addr)
	if err != nil {
		return fmt.Errorf("failed to bind to %s: %w", server.addr, err)
	}

	server.listener = listener
	fmt.Printf("Redis server listening on %s\n", server.addr)

	// Accept connections in a goroutine
	go server.acceptConnections()

	return nil
}

// Stop gracefully shuts down the server
func (server *Server) Stop() error {
	close(server.shutdown)

	if server.listener != nil {
		server.listener.Close()
	}

	// Wait for all connections to finish
	server.wg.Wait()

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
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
		}

		server.wg.Add(1)
		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		server.wg.Done()
	}()

	parser := resp.NewParser(conn)
	encoder := resp.NewEncoder(conn)

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
		response := server.registry.HandleCommand(value)

		// Send the response
		if err := encoder.Encode(response); err != nil {
			fmt.Printf("Error sending response: %v\n", err)
			return
		}
	}
}

// RegisterHandler adds a custom command handler
func (server *Server) RegisterHandler(command string, handler handlers.Handler) {
	server.registry.Register(command, handler)
}
