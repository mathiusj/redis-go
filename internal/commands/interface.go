package commands

import (
	"github.com/codecrafters-redis-go/internal/config"
	"github.com/codecrafters-redis-go/internal/resp"
	"github.com/codecrafters-redis-go/internal/storage"
)

// ServerAccessor provides access to server functionality without circular dependency
type ServerAccessor interface {
	GetReplicas() []interface{} // Returns replica connections
}

// Command represents a Redis command implementation
type Command interface {
	// Name returns the command name (e.g., "SET", "GET")
	Name() string

	// Execute runs the command with the given arguments
	Execute(ctx Context, args []string) resp.Value

	// MinArgs returns the minimum number of arguments required
	MinArgs() int

	// MaxArgs returns the maximum number of arguments (-1 for unlimited)
	MaxArgs() int
}

// Context provides shared resources to commands
type Context struct {
	Storage       *storage.Storage
	Config        *config.Config
	PropagateFunc func(resp.Value) // Function to propagate commands to replicas
	Server        ServerAccessor   // Access to server functions
}

// Validator provides argument validation for commands
type Validator interface {
	Validate(args []string) error
}

// Middleware represents a command middleware function
type Middleware func(Command) Command
