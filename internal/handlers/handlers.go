package handlers

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-redis-go/internal/resp"
	"github.com/codecrafters-redis-go/internal/storage"
)

// Handler is a function that processes a Redis command
type Handler func(args []string) resp.Value

// Registry manages command handlers
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	storage  *storage.Storage
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[string]Handler),
		storage:  storage.New(),
	}

	// Register default handlers
	r.Register("PING", r.handlePing)
	r.Register("ECHO", r.handleEcho)
	r.Register("SET", r.handleSet)
	r.Register("GET", r.handleGet)

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

func (r *Registry) handlePing(args []string) resp.Value {
	if len(args) == 0 {
		return resp.Pong()
	}
	// If argument provided, echo it back
	return resp.SimpleStringValue(args[0])
}

func (r *Registry) handleEcho(args []string) resp.Value {
	if len(args) == 0 {
		return resp.ErrorValue("ERR wrong number of arguments for 'echo' command")
	}
	// ECHO returns the argument as a bulk string
	return resp.BulkStringValue(args[0])
}

func (r *Registry) handleSet(args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'set' command")
	}

	key := args[0]
	value := args[1]
	var expiration *time.Time

	// Parse additional arguments for expiry
	i := 2
	for i < len(args) {
		option := strings.ToUpper(args[i])

		switch option {
		case "EX": // Expire in seconds
			if i+1 >= len(args) {
				return resp.ErrorValue("ERR syntax error")
			}
			seconds, err := strconv.Atoi(args[i+1])
			if err != nil || seconds <= 0 {
				return resp.ErrorValue("ERR invalid expire time in set")
			}
			exp := time.Now().Add(time.Duration(seconds) * time.Second)
			expiration = &exp
			i += 2

		case "PX": // Expire in milliseconds
			if i+1 >= len(args) {
				return resp.ErrorValue("ERR syntax error")
			}
			milliseconds, err := strconv.Atoi(args[i+1])
			if err != nil || milliseconds <= 0 {
				return resp.ErrorValue("ERR invalid expire time in set")
			}
			exp := time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
			expiration = &exp
			i += 2

		default:
			return resp.ErrorValue("ERR syntax error")
		}
	}

	r.storage.Set(key, value, expiration)

	return resp.OK()
}

func (r *Registry) handleGet(args []string) resp.Value {
	if len(args) != 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'get' command")
	}

	key := args[0]
	value, ok := r.storage.Get(key)
	if !ok {
		// Return null bulk string for non-existent keys
		return resp.NullBulkString()
	}

	return resp.BulkStringValue(value)
}
