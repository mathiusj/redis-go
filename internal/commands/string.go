package commands

import (
	"strconv"
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
func (c *SetCommand) Execute(ctx Context, args []string) resp.Value {
	key := args[0]
	value := args[1]

	var expiry *time.Time

	// Parse optional arguments
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "px", "PX":
			if i+1 >= len(args) {
				return resp.ErrorValue(errors.ErrSyntaxError.Error())
			}
			ms, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || ms <= 0 {
				return resp.ErrorValue(errors.ErrInvalidExpireTime.Error())
			}
			exp := time.Now().Add(time.Duration(ms) * time.Millisecond)
			expiry = &exp
			i++ // Skip the next argument
		}
	}

	// Store the value as a string
	ctx.Storage.Set(key, value, expiry)

	// Propagate to replicas - don't do it here, let the server handle it

	return resp.SimpleStringValue("OK")
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
func (c *GetCommand) Execute(ctx Context, args []string) resp.Value {
	key := args[0]

	value, exists := ctx.Storage.GetString(key)
	if !exists {
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
