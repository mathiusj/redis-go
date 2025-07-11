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

	logger.Debug("WAIT command: numreplicas=%d, timeout=%d", numReplicas, timeout)

	// For now, if we're a replica, WAIT is not supported
	if ctx.Config.ReplicaOf != "" {
		return resp.ErrorValue("ERR WAIT cannot be used with replica instances")
	}

	// Get the number of connected replicas
	connectedReplicas := 0
	if ctx.Server != nil {
		connectedReplicas = len(ctx.Server.GetReplicas())
	}
	logger.Debug("Connected replicas: %d", connectedReplicas)

	// If numreplicas is 0, return immediately with the count of connected replicas
	if numReplicas == 0 {
		return resp.Value{
			Type:    resp.Integer,
			Integer: connectedReplicas,
		}
	}

	// TODO: In future stages, implement actual waiting for ACKs
	// For now, with no replicas or no propagation tracking, return 0
	return resp.Value{
		Type:    resp.Integer,
		Integer: 0,
	}
}

func (c *WaitCommand) MinArgs() int {
	return 2
}

func (c *WaitCommand) MaxArgs() int {
	return 2
}
