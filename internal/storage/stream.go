package storage

import (
	"strconv"
	"strings"
	"sync"
)

// StreamEntry represents a single entry in a stream
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// Stream represents a Redis stream data structure
type Stream struct {
	mu      sync.RWMutex
	entries []StreamEntry
}

// NewStream creates a new stream
func NewStream() *Stream {
	return &Stream{
		entries: make([]StreamEntry, 0),
	}
}

// AddEntry adds a new entry to the stream
func (s *Stream) AddEntry(id string, fields map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = append(s.entries, StreamEntry{
		ID:     id,
		Fields: fields,
	})
}

// GetLastEntry returns the last entry in the stream
func (s *Stream) GetLastEntry() *StreamEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0 {
		return nil
	}

	return &s.entries[len(s.entries)-1]
}

// GetEntries returns all entries in the stream
func (s *Stream) GetEntries() []StreamEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid data races
	result := make([]StreamEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// CompareStreamIDs compares two stream IDs
// Returns -1 if id1 < id2, 0 if id1 == id2, 1 if id1 > id2
func CompareStreamIDs(id1, id2 string) int {
	// Parse ID format: timestamp-sequence
	parts1 := strings.Split(id1, "-")
	parts2 := strings.Split(id2, "-")

	// Compare timestamps
	ts1, _ := strconv.ParseInt(parts1[0], 10, 64)
	ts2, _ := strconv.ParseInt(parts2[0], 10, 64)

	if ts1 < ts2 {
		return -1
	} else if ts1 > ts2 {
		return 1
	}

	// Timestamps are equal, compare sequences
	seq1, _ := strconv.ParseInt(parts1[1], 10, 64)
	seq2, _ := strconv.ParseInt(parts2[1], 10, 64)

	if seq1 < seq2 {
		return -1
	} else if seq1 > seq2 {
		return 1
	}

	return 0
}

// Len returns the number of entries in the stream
func (s *Stream) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.entries)
}

// Type returns the type of this value (for the TYPE command)
func (s *Stream) Type() string {
	return "stream"
}
