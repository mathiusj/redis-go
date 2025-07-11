package commands

import (
	"strings"

	"github.com/codecrafters-redis-go/internal/resp"
)

// ReplConfCommand implements the REPLCONF command
type ReplConfCommand struct{}

// NewReplConfCommand creates a new REPLCONF command
func NewReplConfCommand() *ReplConfCommand {
	return &ReplConfCommand{}
}

// Name returns the command name
func (c *ReplConfCommand) Name() string {
	return "REPLCONF"
}

// Execute runs the REPLCONF command
func (c *ReplConfCommand) Execute(ctx Context, args []string) resp.Value {
	if len(args) < 1 {
		return resp.ErrorValue("ERR wrong number of arguments for 'replconf' command")
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "LISTENING-PORT":
		// Just acknowledge for now
		return resp.OK()

	case "CAPA":
		// Just acknowledge capabilities for now
		return resp.OK()

	case "GETACK":
		// Handle GETACK subcommand
		if len(args) < 2 {
			return resp.ErrorValue("ERR wrong number of arguments for REPLCONF GETACK")
		}
		// For now, just acknowledge
		// In later stages, we'll implement proper offset tracking
		return resp.OK()

	default:
		return resp.ErrorValue("ERR Unknown REPLCONF subcommand '" + args[0] + "'")
	}
}

// MinArgs returns the minimum number of arguments
func (c *ReplConfCommand) MinArgs() int {
	return 2 // subcommand and at least one parameter
}

// MaxArgs returns the maximum number of arguments
func (c *ReplConfCommand) MaxArgs() int {
	return -1 // Variable number of arguments depending on subcommand
}
