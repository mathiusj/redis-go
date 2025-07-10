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
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to bind to %s: %w", s.addr, err)
	}

	s.listener = l
	fmt.Printf("Redis server listening on %s\n", s.addr)

	// Accept connections in a goroutine
	go s.acceptConnections()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all connections to finish
	s.wg.Wait()

	return nil
}

// Wait blocks until the server is shut down
func (s *Server) Wait() {
	<-s.shutdown
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				return
			default:
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	parser := resp.NewParser(conn)
	encoder := resp.NewEncoder(conn)

	for {
		// Check for shutdown
		select {
		case <-s.shutdown:
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
		response := s.registry.HandleCommand(value)

		// Send the response
		if err := encoder.Encode(response); err != nil {
			fmt.Printf("Error sending response: %v\n", err)
			return
		}
	}
}

// RegisterHandler adds a custom command handler
func (s *Server) RegisterHandler(command string, handler handlers.Handler) {
	s.registry.Register(command, handler)
}
