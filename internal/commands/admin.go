package commands

import (
	"strings"

	"github.com/codecrafters-redis-go/internal/resp"
)

// ConfigCommand implements the CONFIG command
type ConfigCommand struct{}

// NewConfigCommand creates a new CONFIG command
func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

// Name returns the command name
func (c *ConfigCommand) Name() string {
	return "CONFIG"
}

// Execute runs the CONFIG command
func (c *ConfigCommand) Execute(ctx Context, args []string) resp.Value {
	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "GET":
		if len(args) < 2 {
			return resp.ErrorValue("ERR wrong number of arguments for 'config get' command")
		}
		return c.handleConfigGet(ctx, args[1])
	default:
		return resp.ErrorValue("ERR Unknown subcommand '" + args[0] + "'")
	}
}

// handleConfigGet handles CONFIG GET subcommand
func (c *ConfigCommand) handleConfigGet(ctx Context, pattern string) resp.Value {
	result := []resp.Value{}

	// For now, only support exact matches for "dir" and "dbfilename"
	if pattern == "dir" || pattern == "*" {
		result = append(result, resp.BulkStringValue("dir"))
		result = append(result, resp.BulkStringValue(ctx.Config.Dir))
	}

	if pattern == "dbfilename" || pattern == "*" {
		result = append(result, resp.BulkStringValue("dbfilename"))
		result = append(result, resp.BulkStringValue(ctx.Config.DBFilename))
	}

	return resp.ArrayValue(result...)
}

// MinArgs returns the minimum number of arguments
func (c *ConfigCommand) MinArgs() int {
	return 1
}

// MaxArgs returns the maximum number of arguments
func (c *ConfigCommand) MaxArgs() int {
	return 3
}

// KeysCommand implements the KEYS command
type KeysCommand struct{}

// NewKeysCommand creates a new KEYS command
func NewKeysCommand() *KeysCommand {
	return &KeysCommand{}
}

// Name returns the command name
func (c *KeysCommand) Name() string {
	return "KEYS"
}

// Execute runs the KEYS command
func (c *KeysCommand) Execute(ctx Context, args []string) resp.Value {
	pattern := args[0]

	// Get all matching keys from storage
	keys := ctx.Storage.Keys(pattern)

	// Convert to array of bulk strings
	result := make([]resp.Value, len(keys))
	for i, key := range keys {
		result[i] = resp.BulkStringValue(key)
	}

	return resp.ArrayValue(result...)
}

// MinArgs returns the minimum number of arguments
func (c *KeysCommand) MinArgs() int {
	return 1
}

// MaxArgs returns the maximum number of arguments
func (c *KeysCommand) MaxArgs() int {
	return 1
}
