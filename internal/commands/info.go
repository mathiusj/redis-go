package commands

import (
	"strings"

	"github.com/codecrafters-redis-go/internal/resp"
)

// InfoCommand implements the INFO command
type InfoCommand struct{}

// NewInfoCommand creates a new INFO command
func NewInfoCommand() *InfoCommand {
	return &InfoCommand{}
}

// Name returns the command name
func (c *InfoCommand) Name() string {
	return "INFO"
}

// Execute runs the INFO command
func (c *InfoCommand) Execute(args []string, context *Context) resp.Value {
	// Default to all sections if no section specified
	section := "all"
	if len(args) > 0 {
		section = strings.ToLower(args[0])
	}

	var output []string

	switch section {
	case "replication":
		output = c.getReplicationInfo(context)
	case "all":
		// For now, we only support replication
		output = c.getReplicationInfo(context)
	default:
		// Return empty bulk string for unknown sections
		return resp.BulkStringValue("")
	}

	// Join all lines with CRLF
	result := strings.Join(output, "\r\n")
	return resp.BulkStringValue(result)
}

// getReplicationInfo returns replication information
func (c *InfoCommand) getReplicationInfo(context *Context) []string {
	// For now, we're always a master with no slaves
	return []string{
		"# Replication",
		"role:master",
		"connected_slaves:0",
		"master_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		"master_replid2:0000000000000000000000000000000000000000",
		"master_repl_offset:0",
		"second_repl_offset:-1",
		"repl_backlog_active:0",
		"repl_backlog_size:1048576",
		"repl_backlog_first_byte_offset:0",
		"repl_backlog_histlen:0",
	}
}

// MinArgs returns the minimum number of arguments
func (c *InfoCommand) MinArgs() int {
	return 0
}

// MaxArgs returns the maximum number of arguments
func (c *InfoCommand) MaxArgs() int {
	return 1
}
