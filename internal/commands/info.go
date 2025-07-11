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
func (c *InfoCommand) Execute(ctx Context, args []string) resp.Value {
	section := "all"
	if len(args) > 0 {
		section = strings.ToLower(args[0])
	}

	info := c.buildInfo(ctx, section)
	return resp.BulkStringValue(info)
}

// buildInfo constructs the INFO response
func (c *InfoCommand) buildInfo(ctx Context, section string) string {
	var info strings.Builder

	if section == "all" || section == "replication" {
		info.WriteString("# Replication\r\n")

		if ctx.Config.IsReplica() {
			// Replica mode
			info.WriteString("role:slave\r\n")
			// TODO: Add more replica-specific info in later stages
		} else {
			// Master mode
			info.WriteString("role:master\r\n")
			info.WriteString("master_replid:")
			info.WriteString(c.getMasterReplID())
			info.WriteString("\r\n")
			info.WriteString("master_repl_offset:0\r\n")
		}
	}

	return strings.TrimSpace(info.String())
}

// getMasterReplID returns the master replication ID
func (c *InfoCommand) getMasterReplID() string {
	// Fixed replication ID for now
	return "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
}

// MinArgs returns the minimum number of arguments
func (c *InfoCommand) MinArgs() int {
	return 0
}

// MaxArgs returns the maximum number of arguments
func (c *InfoCommand) MaxArgs() int {
	return 1
}
