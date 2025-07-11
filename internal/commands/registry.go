package commands

import (
	"strings"
	"sync"

	"github.com/codecrafters-redis-go/internal/config"
	"github.com/codecrafters-redis-go/internal/errors"
	"github.com/codecrafters-redis-go/internal/resp"
	"github.com/codecrafters-redis-go/internal/storage"
)

// Registry manages command implementations
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Command
	context  *Context
}

// NewRegistry creates a new command registry
func NewRegistry(cfg *config.Config, store *storage.Storage) *Registry {
	registry := &Registry{
		commands: make(map[string]Command),
		context: &Context{
			Config:  cfg,
			Storage: store,
		},
	}

	// Register default commands
	registry.RegisterCommand(NewPingCommand())
	registry.RegisterCommand(NewEchoCommand())
	registry.RegisterCommand(NewSetCommand())
	registry.RegisterCommand(NewGetCommand())
	registry.RegisterCommand(NewConfigCommand())
	registry.RegisterCommand(NewKeysCommand())
	registry.RegisterCommand(NewInfoCommand())
	registry.RegisterCommand(NewReplConfCommand())
	registry.RegisterCommand(NewPsyncCommand())

	return registry
}

// RegisterCommand adds a new command to the registry
func (r *Registry) RegisterCommand(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[strings.ToUpper(cmd.Name())] = cmd
}

// GetCommand retrieves a command by name
func (r *Registry) GetCommand(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.commands[strings.ToUpper(name)]
	return cmd, ok
}

// HandleCommand processes a command and returns a response
func (r *Registry) HandleCommand(cmdValue resp.Value) resp.Value {
	commandName, err := cmdValue.GetCommand()
	if err != nil {
		return resp.ErrorValue("ERR invalid command format")
	}

	cmd, ok := r.GetCommand(commandName)
	if !ok {
		return resp.ErrorValue(errors.UnknownCommand(commandName).Error())
	}

	args := cmdValue.GetArgs()

	// Validate argument count
	if cmd.MinArgs() > 0 && len(args) < cmd.MinArgs() {
		return resp.ErrorValue(errors.WrongNumberOfArguments(strings.ToLower(commandName)).Error())
	}

	if cmd.MaxArgs() >= 0 && len(args) > cmd.MaxArgs() {
		return resp.ErrorValue(errors.WrongNumberOfArguments(strings.ToLower(commandName)).Error())
	}

	// Execute the command
	return cmd.Execute(args, r.context)
}

// GetContext returns the command context
func (r *Registry) GetContext() *Context {
	return r.context
}

// SetPropagateFunc sets the propagation function for command replication
func (r *Registry) SetPropagateFunc(propagateFunc func(resp.Value)) {
	r.context.PropagateFunc = propagateFunc
}
