package commands

import (
	"github.com/codecrafters-redis-go/internal/resp"
	"github.com/codecrafters-redis-go/internal/storage"
)

// XAddCommand implements the XADD command
type XAddCommand struct{}

func NewXAddCommand() *XAddCommand {
	return &XAddCommand{}
}

func (c *XAddCommand) Name() string {
	return "XADD"
}

func (c *XAddCommand) Execute(ctx Context, args []string) resp.Value {
	key := args[0]
	id := args[1]

	// Check if ID is 0-0
	if id == "0-0" {
		return resp.ErrorValue("ERR The ID specified in XADD must be greater than 0-0")
	}

	// Parse field-value pairs
	if len(args[2:])%2 != 0 {
		return resp.ErrorValue("ERR wrong number of arguments for 'xadd' command")
	}

	fields := make(map[string]string)
	for i := 2; i < len(args); i += 2 {
		fields[args[i]] = args[i+1]
	}

	// Get or create stream
	val, exists := ctx.Storage.Get(key)
	var stream *storage.Stream

	if exists {
		// Check if it's a stream
		var ok bool
		stream, ok = val.(*storage.Stream)
		if !ok {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value")
		}

		// Validate ID is greater than the last entry
		lastEntry := stream.GetLastEntry()
		if lastEntry != nil {
			comparison := storage.CompareStreamIDs(id, lastEntry.ID)
			if comparison <= 0 {
				return resp.ErrorValue("ERR The ID specified in XADD is equal or smaller than the target stream top item")
			}
		}
	} else {
		// Create new stream
		stream = storage.NewStream()
		ctx.Storage.Set(key, stream, nil)
	}

	// Add entry to stream
	stream.AddEntry(id, fields)

	// Return the ID that was added
	return resp.BulkStringValue(id)
}

func (c *XAddCommand) MinArgs() int {
	return 4 // key id field value
}

func (c *XAddCommand) MaxArgs() int {
	return -1 // Variable number of field-value pairs
}
