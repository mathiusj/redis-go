package handlers

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-redis-go/internal/config"
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
	config   *config.Config
}

// NewRegistry creates a new command registry
func NewRegistry(cfg *config.Config) *Registry {
	return NewRegistryWithStorage(cfg, storage.New())
}

// NewRegistryWithStorage creates a new command registry with a provided storage
func NewRegistryWithStorage(cfg *config.Config, store *storage.Storage) *Registry {
	registry := &Registry{
		handlers: make(map[string]Handler),
		storage:  store,
		config:   cfg,
	}

	// Register default handlers
	registry.Register("PING", registry.handlePing)
	registry.Register("ECHO", registry.handleEcho)
	registry.Register("SET", registry.handleSet)
	registry.Register("GET", registry.handleGet)
	registry.Register("CONFIG", registry.handleConfig)
	registry.Register("KEYS", registry.handleKeys)

	return registry
}

// Register adds a new command handler
func (registry *Registry) Register(command string, handler Handler) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.handlers[strings.ToUpper(command)] = handler
}

// Get retrieves a handler for a command
func (registry *Registry) Get(command string) (Handler, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	handler, ok := registry.handlers[strings.ToUpper(command)]
	return handler, ok
}

// HandleCommand processes a command and returns a response
func (registry *Registry) HandleCommand(cmd resp.Value) resp.Value {
	command, err := cmd.GetCommand()
	if err != nil {
		return resp.ErrorValue("ERR invalid command format")
	}

	handler, ok := registry.Get(command)
	if !ok {
		return resp.ErrorValue("ERR unknown command '" + command + "'")
	}

	args := cmd.GetArgs()
	return handler(args)
}

// Command handlers

func (registry *Registry) handlePing(args []string) resp.Value {
	if len(args) == 0 {
		return resp.Pong()
	}
	// If argument provided, echo it back
	return resp.SimpleStringValue(args[0])
}

func (registry *Registry) handleEcho(args []string) resp.Value {
	if len(args) == 0 {
		return resp.ErrorValue("ERR wrong number of arguments for 'echo' command")
	}
	// ECHO returns the argument as a bulk string
	return resp.BulkStringValue(args[0])
}

func (registry *Registry) handleSet(args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'set' command")
	}

	key := args[0]
	value := args[1]
	var expiration *time.Time

	// Parse additional arguments for expiry
	argIndex := 2
	for argIndex < len(args) {
		option := strings.ToUpper(args[argIndex])

		switch option {
		case "EX": // Expire in seconds
			if argIndex+1 >= len(args) {
				return resp.ErrorValue("ERR syntax error")
			}
			seconds, err := strconv.Atoi(args[argIndex+1])
			if err != nil || seconds <= 0 {
				return resp.ErrorValue("ERR invalid expire time in set")
			}
			expirationTime := time.Now().Add(time.Duration(seconds) * time.Second)
			expiration = &expirationTime
			argIndex += 2

		case "PX": // Expire in milliseconds
			if argIndex+1 >= len(args) {
				return resp.ErrorValue("ERR syntax error")
			}
			milliseconds, err := strconv.Atoi(args[argIndex+1])
			if err != nil || milliseconds <= 0 {
				return resp.ErrorValue("ERR invalid expire time in set")
			}
			expirationTime := time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
			expiration = &expirationTime
			argIndex += 2

		default:
			return resp.ErrorValue("ERR syntax error")
		}
	}

	registry.storage.Set(key, value, expiration)

	return resp.OK()
}

func (registry *Registry) handleGet(args []string) resp.Value {
	if len(args) != 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'get' command")
	}

	key := args[0]
	value, ok := registry.storage.Get(key)
	if !ok {
		// Return null bulk string for non-existent keys
		return resp.NullBulkString()
	}

	return resp.BulkStringValue(value)
}

func (registry *Registry) handleConfig(args []string) resp.Value {
	if len(args) < 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'config' command")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "GET":
		if len(args) != 2 {
			return resp.ErrorValue("ERR wrong number of arguments for 'config get' command")
		}

		parameter := strings.ToLower(args[1])
		value, ok := registry.config.Get(parameter)
		if !ok {
			// Return empty array for unknown config parameters
			return resp.ArrayValue()
		}

		// Return array with parameter name and value
		return resp.ArrayValue(
			resp.BulkStringValue(parameter),
			resp.BulkStringValue(value),
		)

	case "SET":
		if len(args) != 3 {
			return resp.ErrorValue("ERR wrong number of arguments for 'config set' command")
		}

		parameter := strings.ToLower(args[1])
		value := args[2]

		if !registry.config.Set(parameter, value) {
			return resp.ErrorValue("ERR Unsupported CONFIG parameter: " + parameter)
		}

		return resp.OK()

	default:
		return resp.ErrorValue("ERR Unknown subcommand or wrong number of arguments")
	}
}

func (registry *Registry) handleKeys(args []string) resp.Value {
	if len(args) != 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0]

	// For now, we'll implement simple pattern matching
	// "*" means all keys
	keys := registry.storage.Keys(pattern)

	// Convert keys to RESP array
	values := make([]resp.Value, len(keys))
	for index, key := range keys {
		values[index] = resp.BulkStringValue(key)
	}

	return resp.ArrayValue(values...)
}
