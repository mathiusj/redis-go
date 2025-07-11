package storage

import (
	"sync"
	"time"

	"github.com/codecrafters-redis-go/internal/logger"
	"github.com/codecrafters-redis-go/internal/utils"
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

	// Background cleanup
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
	cleanupDone     sync.WaitGroup
}

// New creates a new storage instance
func New() *Storage {
	return NewWithCleanupInterval(1 * time.Minute)
}

// NewWithCleanupInterval creates a new storage instance with custom cleanup interval
func NewWithCleanupInterval(interval time.Duration) *Storage {
	storage := &Storage{
		data:            make(map[string]*Entry),
		cleanupInterval: interval,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup if interval is positive
	if interval > 0 {
		storage.startCleanup()
	}

	return storage
}

// startCleanup starts the background cleanup goroutine
func (storage *Storage) startCleanup() {
	storage.cleanupDone.Add(1)
	go func() {
		defer storage.cleanupDone.Done()

		ticker := time.NewTicker(storage.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count := storage.cleanupExpired()
				if count > 0 {
					logger.Debug("Cleaned up %d expired keys", count)
				}
			case <-storage.stopCleanup:
				return
			}
		}
	}()
}

// cleanupExpired removes all expired entries and returns the count
func (storage *Storage) cleanupExpired() int {
	storage.mu.Lock()
	defer storage.mu.Unlock()

	now := time.Now()
	count := 0

	for key, entry := range storage.data {
		if entry.Expiration != nil && now.After(*entry.Expiration) {
			delete(storage.data, key)
			count++
		}
	}

	return count
}

// Close stops the background cleanup goroutine
func (storage *Storage) Close() {
	close(storage.stopCleanup)
	storage.cleanupDone.Wait()
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

// Keys returns all keys matching the given pattern
func (storage *Storage) Keys(pattern string) []string {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	keys := make([]string, 0)

	// Iterate through all keys and check pattern match
	for key, entry := range storage.data {
		// Check if expired
		if entry.Expiration != nil && time.Now().After(*entry.Expiration) {
			continue
		}

		// Check if key matches pattern
		if utils.MatchPattern(pattern, key) {
			keys = append(keys, key)
		}
	}

	return keys
}
