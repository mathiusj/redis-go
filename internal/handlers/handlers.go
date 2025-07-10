package handlers

import (
	"strings"
	"sync"

	"github.com/codecrafters-redis-go/internal/resp"
)

// Handler is a function that processes a Redis command
type Handler func(args []string) resp.Value

// Registry manages command handlers
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[string]Handler),
	}

	// Register default handlers
	r.Register("PING", handlePing)
	r.Register("ECHO", handleEcho)

	return r
}

// Register adds a new command handler
func (r *Registry) Register(command string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[strings.ToUpper(command)] = handler
}

// Get retrieves a handler for a command
func (r *Registry) Get(command string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[strings.ToUpper(command)]
	return handler, ok
}

// HandleCommand processes a command and returns a response
func (r *Registry) HandleCommand(cmd resp.Value) resp.Value {
	command, err := cmd.GetCommand()
	if err != nil {
		return resp.ErrorValue("ERR invalid command format")
	}

	handler, ok := r.Get(command)
	if !ok {
		return resp.ErrorValue("ERR unknown command '" + command + "'")
	}

	args := cmd.GetArgs()
	return handler(args)
}

// Command handlers

func handlePing(args []string) resp.Value {
	if len(args) == 0 {
		return resp.Pong()
	}
	// If argument provided, echo it back
	return resp.SimpleStringValue(args[0])
}

func handleEcho(args []string) resp.Value {
	if len(args) == 0 {
		return resp.ErrorValue("ERR wrong number of arguments for 'echo' command")
	}
	// ECHO returns the argument as a bulk string
	return resp.BulkStringValue(args[0])
}
