package commands

import (
	"strconv"

	"github.com/codecrafters-redis-go/internal/logger"
	"github.com/codecrafters-redis-go/internal/resp"
)

// WaitCommand implements the WAIT command
type WaitCommand struct{}

func NewWaitCommand() *WaitCommand {
	return &WaitCommand{}
}

func (c *WaitCommand) Name() string {
	return "WAIT"
}

func (c *WaitCommand) Execute(ctx Context, args []string) resp.Value {
	// WAIT numreplicas timeout
	if len(args) < 2 {
		return resp.ErrorValue("ERR wrong number of arguments for 'wait' command")
	}

	// Parse numreplicas
	numReplicas, err := strconv.Atoi(args[0])
	if err != nil || numReplicas < 0 {
		return resp.ErrorValue("ERR invalid numreplicas")
	}

	// Parse timeout (in milliseconds)
	timeout, err := strconv.Atoi(args[1])
	if err != nil || timeout < 0 {
		return resp.ErrorValue("ERR invalid timeout")
	}

	// Get current replicas
	var replicas []interface{}
	if ctx.Server != nil {
		replicas = ctx.Server.GetReplicas()
	}

	logger.Debug("WAIT command: numreplicas=%d, timeout=%d, connected_replicas=%d",
		numReplicas, timeout, len(replicas))

	// When no write commands have been sent since replicas connected,
	// all replicas are considered synchronized
	// For now, we assume all connected replicas are synchronized
	// In later stages, we'll track actual synchronization status
	synchronizedCount := len(replicas)

	// Return the count of synchronized replicas
	return resp.Value{
		Type:    resp.Integer,
		Integer: synchronizedCount,
	}
}

// MinArgs returns the minimum number of arguments
func (c *WaitCommand) MinArgs() int {
	return 2
}

// MaxArgs returns the maximum number of arguments
func (c *WaitCommand) MaxArgs() int {
	return 2
}
