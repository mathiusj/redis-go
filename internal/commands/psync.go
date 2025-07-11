package commands

import (
	"fmt"

	"github.com/codecrafters-redis-go/internal/errors"
	"github.com/codecrafters-redis-go/internal/logger"
	"github.com/codecrafters-redis-go/internal/resp"
)

// PsyncCommand implements the PSYNC command
type PsyncCommand struct{}

// NewPsyncCommand creates a new PSYNC command
func NewPsyncCommand() *PsyncCommand {
	return &PsyncCommand{}
}

// Name returns the command name
func (c *PsyncCommand) Name() string {
	return "PSYNC"
}

// Execute runs the PSYNC command
func (c *PsyncCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) < 2 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("psync").Error())
	}

	replicationID := args[0]
	offset := args[1]

	logger.Debug("Received PSYNC %s %s", replicationID, offset)

	// For now, we always respond with FULLRESYNC
	// In a real implementation, we would check if we can do partial sync
	if replicationID == "?" && offset == "-1" {
		// Replica is requesting full sync
		// Generate a replication ID (same one we use in INFO command)
		replID := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		masterOffset := 0

		response := fmt.Sprintf("FULLRESYNC %s %d", replID, masterOffset)
		logger.Info("Sending FULLRESYNC to replica")

		// TODO: In future stages, we'll need to send the RDB file after this response

		return resp.SimpleStringValue(response)
	}

	// For partial sync requests, we would check if we can continue from the given offset
	// For now, always force full sync
	replID := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	masterOffset := 0

	response := fmt.Sprintf("FULLRESYNC %s %d", replID, masterOffset)
	return resp.SimpleStringValue(response)
}

// MinArgs returns the minimum number of arguments
func (c *PsyncCommand) MinArgs() int {
	return 2 // replication_id and offset
}

// MaxArgs returns the maximum number of arguments
func (c *PsyncCommand) MaxArgs() int {
	return 2
}
