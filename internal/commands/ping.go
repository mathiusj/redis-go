package commands

import (
	"github.com/codecrafters-redis-go/internal/resp"
)

// PingCommand implements the PING command
type PingCommand struct{}

// NewPingCommand creates a new PING command
func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

// Name returns the command name
func (c *PingCommand) Name() string {
	return "PING"
}

// Execute runs the PING command
func (c *PingCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) == 0 {
		return resp.Pong()
	}
	// If argument provided, echo it back as simple string
	return resp.SimpleStringValue(args[0])
}

// MinArgs returns the minimum number of arguments
func (c *PingCommand) MinArgs() int {
	return 0
}

// MaxArgs returns the maximum number of arguments
func (c *PingCommand) MaxArgs() int {
	return 1
}
