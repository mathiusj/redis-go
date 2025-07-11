package commands

import (
	"strconv"
	"time"

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

	// Convert timeout to duration
	timeoutDuration := time.Duration(timeout) * time.Millisecond

	logger.Debug("WAIT command: numreplicas=%d, timeout=%d ms", numReplicas, timeout)

	// Check if we have access to the server
	if ctx.Server == nil {
		return resp.ErrorValue("ERR WAIT is not supported in this context")
	}

	// Type assert to get the actual server with WaitForReplicas method
	type serverWaiter interface {
		WaitForReplicas(int, time.Duration) int
	}

	waiter, ok := ctx.Server.(serverWaiter)
	if !ok {
		// Fallback to old behavior if server doesn't implement WaitForReplicas
		replicas := ctx.Server.GetReplicas()
		return resp.Value{
			Type:    resp.Integer,
			Integer: len(replicas),
		}
	}

	// Wait for replicas to acknowledge
	synchronizedCount := waiter.WaitForReplicas(numReplicas, timeoutDuration)

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
