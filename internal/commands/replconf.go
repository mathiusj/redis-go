package commands

import (
	"strings"

	"github.com/codecrafters-redis-go/internal/errors"
	"github.com/codecrafters-redis-go/internal/logger"
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
func (c *ReplConfCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) < 2 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("replconf").Error())
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "listening-port":
		// Handle listening-port subcommand
		port := args[1]
		logger.Debug("Received REPLCONF listening-port %s", port)
		// TODO: In future stages, we might want to store replica information
		return resp.SimpleStringValue("OK")

	case "capa":
		// Handle capa subcommand
		capability := args[1]
		logger.Debug("Received REPLCONF capa %s", capability)
		// TODO: In future stages, we might want to track replica capabilities
		return resp.SimpleStringValue("OK")

	default:
		return resp.ErrorValue("ERR unsupported REPLCONF subcommand '" + subcommand + "'")
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
