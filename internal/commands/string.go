package commands

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-redis-go/internal/errors"
	"github.com/codecrafters-redis-go/internal/resp"
)

// SetCommand implements the SET command
type SetCommand struct{}

// NewSetCommand creates a new SET command
func NewSetCommand() *SetCommand {
	return &SetCommand{}
}

// Name returns the command name
func (c *SetCommand) Name() string {
	return "SET"
}

// Execute runs the SET command
func (c *SetCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) < 2 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("set").Error())
	}

	key := args[0]
	value := args[1]
	var expiration *time.Time

	// Parse additional arguments for expiry
	argIndex := 2
	for argIndex < len(args) {
		option := strings.ToUpper(args[argIndex])

		switch option {
		case "EX": // Expire in seconds
			if argIndex+1 >= len(args) {
				return resp.ErrorValue(errors.ErrSyntaxError.Error())
			}
			seconds, err := strconv.Atoi(args[argIndex+1])
			if err != nil || seconds <= 0 {
				return resp.ErrorValue(errors.InvalidExpireTime("set").Error())
			}
			expirationTime := time.Now().Add(time.Duration(seconds) * time.Second)
			expiration = &expirationTime
			argIndex += 2

		case "PX": // Expire in milliseconds
			if argIndex+1 >= len(args) {
				return resp.ErrorValue(errors.ErrSyntaxError.Error())
			}
			milliseconds, err := strconv.Atoi(args[argIndex+1])
			if err != nil || milliseconds <= 0 {
				return resp.ErrorValue(errors.InvalidExpireTime("set").Error())
			}
			expirationTime := time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
			expiration = &expirationTime
			argIndex += 2

		default:
			return resp.ErrorValue(errors.ErrSyntaxError.Error())
		}
	}

	context.Storage.Set(key, value, expiration)
	return resp.OK()
}

// MinArgs returns the minimum number of arguments
func (c *SetCommand) MinArgs() int {
	return 2
}

// MaxArgs returns the maximum number of arguments
func (c *SetCommand) MaxArgs() int {
	return -1 // Variable number of arguments
}

// GetCommand implements the GET command
type GetCommand struct{}

// NewGetCommand creates a new GET command
func NewGetCommand() *GetCommand {
	return &GetCommand{}
}

// Name returns the command name
func (c *GetCommand) Name() string {
	return "GET"
}

// Execute runs the GET command
func (c *GetCommand) Execute(args []string, context *Context) resp.Value {
	if len(args) != 1 {
		return resp.ErrorValue(errors.WrongNumberOfArguments("get").Error())
	}

	key := args[0]
	value, ok := context.Storage.Get(key)
	if !ok {
		// Return null bulk string for non-existent keys
		return resp.NullBulkString()
	}

	return resp.BulkStringValue(value)
}

// MinArgs returns the minimum number of arguments
func (c *GetCommand) MinArgs() int {
	return 1
}

// MaxArgs returns the maximum number of arguments
func (c *GetCommand) MaxArgs() int {
	return 1
}
