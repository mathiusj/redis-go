package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
		s, ok := val.(*storage.Stream)
		if !ok {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		stream = s
	} else {
		// Create new stream
		stream = storage.NewStream()
		ctx.Storage.Set(key, stream, nil)
	}

	// Parse and validate ID
	entryID, err := parseStreamID(id, stream)
	if err != nil {
		return resp.ErrorValue(err.Error())
	}

	// Add entry to stream
	stream.AddEntry(entryID, fields)

	// Return the ID
	return resp.BulkStringValue(entryID)
}

func (c *XAddCommand) MinArgs() int {
	return 4 // key id field value
}

func (c *XAddCommand) MaxArgs() int {
	return -1 // unlimited
}

// parseStreamID parses and validates a stream ID
func parseStreamID(id string, stream *storage.Stream) (string, error) {
	// Handle special * ID
	if id == "*" {
		ms := time.Now().UnixMilli()
		seq := uint64(0)

		// If we have entries, check if we need to increment sequence
		if lastEntry := stream.GetLastEntry(); lastEntry != nil {
			lastMS, lastSeq := parseID(lastEntry.ID)
			if lastMS == uint64(ms) {
				seq = lastSeq + 1
			}
		}

		return fmt.Sprintf("%d-%d", ms, seq), nil
	}

	// Parse explicit ID
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	// Handle partial IDs with *
	if parts[1] == "*" {
		ms, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
		}

		seq := uint64(0)
		if lastEntry := stream.GetLastEntry(); lastEntry != nil {
			lastMS, lastSeq := parseID(lastEntry.ID)
			if lastMS == ms {
				seq = lastSeq + 1
			}
		}

		return fmt.Sprintf("%d-%d", ms, seq), nil
	}

	// Validate full explicit ID
	ms, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	seq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
	}

	// Check if ID is valid (must be greater than last entry)
	if lastEntry := stream.GetLastEntry(); lastEntry != nil {
		lastMS, lastSeq := parseID(lastEntry.ID)
		if ms < lastMS || (ms == lastMS && seq <= lastSeq) {
			return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	// Check for 0-0 ID
	if ms == 0 && seq == 0 {
		return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	return id, nil
}

// parseID parses a stream ID into its millisecond and sequence components
func parseID(id string) (uint64, uint64) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0
	}

	ms, _ := strconv.ParseUint(parts[0], 10, 64)
	seq, _ := strconv.ParseUint(parts[1], 10, 64)

	return ms, seq
}
