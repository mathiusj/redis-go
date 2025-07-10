package storage

import (
	"sync"
	"time"
)

// Entry represents a stored value with optional expiration
type Entry struct {
	Value      string
	Expiration *time.Time
}

// Storage provides thread-safe key-value storage
type Storage struct {
	mu    sync.RWMutex
	data  map[string]*Entry
}

// New creates a new storage instance
func New() *Storage {
	return &Storage{
		data: make(map[string]*Entry),
	}
}

// Set stores a key-value pair
func (s *Storage) Set(key, value string, expiration *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = &Entry{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value by key
func (s *Storage) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok {
		return "", false
	}

	// Check if expired
	if entry.Expiration != nil && time.Now().After(*entry.Expiration) {
		// Remove expired entry
		delete(s.data, key)
		return "", false
	}

	return entry.Value, true
}

// Delete removes a key from storage
func (s *Storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, existed := s.data[key]
	delete(s.data, key)
	return existed
}

// Exists checks if a key exists and is not expired
func (s *Storage) Exists(key string) bool {
	_, ok := s.Get(key)
	return ok
}

// Size returns the number of keys in storage
func (s *Storage) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}
