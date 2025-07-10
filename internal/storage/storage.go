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
func (storage *Storage) Set(key, value string, expiration *time.Time) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.data[key] = &Entry{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves a value by key
func (storage *Storage) Get(key string) (string, bool) {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	entry, ok := storage.data[key]
	if !ok {
		return "", false
	}

	// Check if expired
	if entry.Expiration != nil && time.Now().After(*entry.Expiration) {
		// Remove expired entry
		delete(storage.data, key)
		return "", false
	}

	return entry.Value, true
}

// Delete removes a key from storage
func (storage *Storage) Delete(key string) bool {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	_, existed := storage.data[key]
	delete(storage.data, key)
	return existed
}

// Exists checks if a key exists and is not expired
func (storage *Storage) Exists(key string) bool {
	_, ok := storage.Get(key)
	return ok
}

// Size returns the number of keys in storage
func (storage *Storage) Size() int {
	storage.mu.RLock()
	defer storage.mu.RUnlock()
	return len(storage.data)
}
