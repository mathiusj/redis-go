package commands

import (
	"strconv"
	"strings"
	"time"

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

	// Parse options
	var ttl *time.Duration
	var condition string

	for i := 2; i < len(args); i++ {
		opt := strings.ToUpper(args[i])
		switch opt {
		case "PX":
			if i+1 >= len(args) {
				return resp.ErrorValue("ERR syntax error")
			}
			ms, err := strconv.Atoi(args[i+1])
			if err != nil || ms <= 0 {
				return resp.ErrorValue("ERR invalid expire time in set")
			}
			d := time.Duration(ms) * time.Millisecond
			ttl = &d
			i++ // Skip the next argument
		case "NX":
			condition = "NX"
		case "XX":
			condition = "XX"
		default:
			return resp.ErrorValue("ERR syntax error")
		}
	}

	// Apply conditions
	switch condition {
	case "NX":
		if ctx.Storage.Exists(key) {
			return resp.NullBulkString()
		}
	case "XX":
		if !ctx.Storage.Exists(key) {
			return resp.NullBulkString()
		}
	}

	// Set the value
	var expiration *time.Time
	if ttl != nil {
		exp := time.Now().Add(*ttl)
		expiration = &exp
	}
	ctx.Storage.Set(key, value, expiration)

	// Don't propagate here - let the server handle propagation uniformly
	// The server will propagate the original command after successful execution

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
func (c *GetCommand) Execute(ctx Context, args []string) resp.Value {
	key := args[0]
	value, exists := ctx.Storage.Get(key)
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
