package storage

import (
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
