package commands

import (
	"strings"

	"github.com/codecrafters-redis-go/internal/errors"
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
func (c *ConfigCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) < 1 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("config").Error())
	}

	subcommand := strings.ToUpper(args[0])

	switch subcommand {
	case "GET":
		if len(args) != 2 {
			return resp.ErrorValue(errors.WrongNumberOfArguments("config get").Error())
		}

		parameter := strings.ToLower(args[1])
		value, ok := context.Config.Get(parameter)
		if !ok {
			// Return empty array for unknown config parameters
			return resp.ArrayValue()
		}

		// Return array with parameter name and value
		return resp.ArrayValue(
			resp.BulkStringValue(parameter),
			resp.BulkStringValue(value),
		)

	case "SET":
		if len(args) != 3 {
			return resp.ErrorValue(errors.WrongNumberOfArguments("config set").Error())
		}

		parameter := strings.ToLower(args[1])
		value := args[2]

		if !context.Config.Set(parameter, value) {
			return resp.ErrorValue("ERR Unsupported CONFIG parameter: " + parameter)
		}

		return resp.OK()

	default:
		return resp.ErrorValue("ERR Unknown subcommand or wrong number of arguments")
	}
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
func (c *KeysCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) != 1 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("keys").Error())
	}

	pattern := args[0]
	keys := context.Storage.Keys(pattern)

	// Convert keys to RESP array
	values := make([]resp.Value, len(keys))
	for index, key := range keys {
		values[index] = resp.BulkStringValue(key)
	}

	return resp.ArrayValue(values...)
}

// MinArgs returns the minimum number of arguments
func (c *KeysCommand) MinArgs() int {
	return 1
}

// MaxArgs returns the maximum number of arguments
func (c *KeysCommand) MaxArgs() int {
	return 1
}
