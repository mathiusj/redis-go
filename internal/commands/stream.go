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
		var ok bool
		stream, ok = val.(*storage.Stream)
		if !ok {
			return resp.ErrorValue("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		// Create new stream
		stream = storage.NewStream()
		ctx.Storage.Set(key, stream, nil)
	}

	// Parse and generate ID if needed
	generatedID, err := parseStreamID(id, stream)
	if err != nil {
		return resp.ErrorValue(err.Error())
	}

	// Add entry to stream
	stream.AddEntry(generatedID, fields)

	// Return the generated ID
	return resp.BulkStringValue(generatedID)
}

func (c *XAddCommand) MinArgs() int {
	return 4 // key id field value
}

func (c *XAddCommand) MaxArgs() int {
	return -1 // Variable number of field-value pairs
}

// parseStreamID parses and generates a stream ID
func parseStreamID(id string, stream *storage.Stream) (string, error) {
	// Check for special case 0-0
	if id == "0-0" {
		return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	// Handle full auto-generation with *
	if id == "*" {
		ms := time.Now().UnixMilli()
		seq := uint64(0)

		// If we have entries, check if we need to increment sequence
		if lastEntry := stream.GetLastEntry(); lastEntry != nil {
			lastMS, lastSeq := parseExistingID(lastEntry.ID)
			if lastMS == uint64(ms) {
				seq = lastSeq + 1
			}
		}

		return fmt.Sprintf("%d-%d", ms, seq), nil
	}

	// Handle partial auto-generation (e.g., "123-*")
	if strings.Contains(id, "*") {
		parts := strings.Split(id, "-")
		if len(parts) != 2 || parts[1] != "*" {
			return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
		}

		ms, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return "", fmt.Errorf("ERR Invalid stream ID specified as stream command argument")
		}

		seq := uint64(0)
		if lastEntry := stream.GetLastEntry(); lastEntry != nil {
			lastMS, lastSeq := parseExistingID(lastEntry.ID)
			// If the timestamp matches the last entry, increment the sequence
			if lastMS == ms {
				seq = lastSeq + 1
			}
			// If the timestamp is less than the last entry, it's invalid
			if ms < lastMS {
				return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
			}
		} else if ms == 0 {
			// Special case: for timestamp 0 with no entries, start at sequence 1
			seq = 1
		}

		return fmt.Sprintf("%d-%d", ms, seq), nil
	}

	// Handle explicit ID - validate it's greater than the last entry
	if lastEntry := stream.GetLastEntry(); lastEntry != nil {
		comparison := storage.CompareStreamIDs(id, lastEntry.ID)
		if comparison <= 0 {
			return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	return id, nil
}

// parseExistingID parses a stream ID into its components
func parseExistingID(id string) (uint64, uint64) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0
	}

	ms, _ := strconv.ParseUint(parts[0], 10, 64)
	seq, _ := strconv.ParseUint(parts[1], 10, 64)

	return ms, seq
}
