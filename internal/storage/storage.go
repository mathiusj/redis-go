package storage

import (
	"sync"
	"time"

	"github.com/codecrafters-redis-go/internal/utils"
)

// ValueType interface for different Redis data types
type ValueType interface {
	Type() string
}

// StringValue represents a Redis string value
type StringValue struct {
	Value string
}

func (s StringValue) Type() string {
	return "string"
}

type entry struct {
	value  interface{}
	expiry *time.Time
}

type Storage struct {
	mu      sync.RWMutex
	data    map[string]entry
	done    chan struct{}
	stopped bool
}

func New() *Storage {
	s := &Storage{
		data: make(map[string]entry),
		done: make(chan struct{}),
	}
	go s.cleanupExpired()
	return s
}

func (s *Storage) Set(key string, value interface{}, expiry *time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = entry{value: value, expiry: expiry}
}

func (s *Storage) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.data[key]
	if !exists {
		return nil, false
	}

	if e.expiry != nil && time.Now().After(*e.expiry) {
		// Key has expired, remove it
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		s.mu.RLock()
		return nil, false
	}

	return e.value, true
}

// GetString gets a value and returns it as a string if it's a string type
func (s *Storage) GetString(key string) (string, bool) {
	val, exists := s.Get(key)
	if !exists {
		return "", false
	}

	switch v := val.(type) {
	case string:
		return v, true
	case StringValue:
		return v.Value, true
	default:
		return "", false
	}
}

func (s *Storage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Storage) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys []string
	now := time.Now()

	for key, e := range s.data {
		// Skip expired keys
		if e.expiry != nil && now.After(*e.expiry) {
			continue
		}

		if pattern == "*" || utils.MatchPattern(pattern, key) {
			keys = append(keys, key)
		}
	}

	return keys
}

func (s *Storage) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for key, e := range s.data {
				if e.expiry != nil && now.After(*e.expiry) {
					delete(s.data, key)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}

func (s *Storage) Close() {
	s.mu.Lock()
	if !s.stopped {
		s.stopped = true
		close(s.done)
	}
	s.mu.Unlock()
}
