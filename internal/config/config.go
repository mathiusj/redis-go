package config

import (
	"flag"
	"sync"
)

// Config holds the Redis server configuration
type Config struct {
	mu         sync.RWMutex
	Dir        string
	DBFilename string
	Port       int
}

// New creates a new configuration with default values
func New() *Config {
	return &Config{
		Dir:        ".",
		DBFilename: "dump.rdb",
		Port:       6379,
	}
}

// ParseFlags parses command-line flags and updates the configuration
func (config *Config) ParseFlags() {
	flag.StringVar(&config.Dir, "dir", config.Dir, "The directory where RDB files are stored")
	flag.StringVar(&config.DBFilename, "dbfilename", config.DBFilename, "The name of the RDB file")
	flag.IntVar(&config.Port, "port", config.Port, "The port to listen on")
	flag.Parse()
}

// Get retrieves a configuration value by key
func (config *Config) Get(key string) (string, bool) {
	config.mu.RLock()
	defer config.mu.RUnlock()

	switch key {
	case "dir":
		return config.Dir, true
	case "dbfilename":
		return config.DBFilename, true
	default:
		return "", false
	}
}

// Set updates a configuration value by key
func (config *Config) Set(key, value string) bool {
	config.mu.Lock()
	defer config.mu.Unlock()

	switch key {
	case "dir":
		config.Dir = value
		return true
	case "dbfilename":
		config.DBFilename = value
		return true
	default:
		return false
	}
}
