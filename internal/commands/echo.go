package commands

import (
	"github.com/codecrafters-redis-go/internal/errors"
	"github.com/codecrafters-redis-go/internal/resp"
)

// EchoCommand implements the ECHO command
type EchoCommand struct{}

// NewEchoCommand creates a new ECHO command
func NewEchoCommand() *EchoCommand {
	return &EchoCommand{}
}

// Name returns the command name
func (c *EchoCommand) Name() string {
	return "ECHO"
}

// Execute runs the ECHO command
func (c *EchoCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) == 0 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("echo").Error())
	}
	// ECHO returns the argument as a bulk string
	return resp.BulkStringValue(args[0])
}

// MinArgs returns the minimum number of arguments
func (c *EchoCommand) MinArgs() int {
	return 1
}

// MaxArgs returns the maximum number of arguments
func (c *EchoCommand) MaxArgs() int {
	return 1
}
