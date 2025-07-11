package commands

import (
	"fmt"

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
func (c *PsyncCommand) Execute(ctx Context, args []string) resp.Value {
	if len(args) < 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'psync' command")
	}

	replID := args[0]
	offset := args[1]

	logger.Debug("Received PSYNC %s %s", replID, offset)

	// For now, we always respond with FULLRESYNC
	// In later stages, we might support partial resyncs
	masterReplID := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	masterOffset := "0"

	response := fmt.Sprintf("FULLRESYNC %s %s", masterReplID, masterOffset)
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
