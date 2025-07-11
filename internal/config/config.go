package config

import (
	"flag"
	"strings"
	"sync"
)

// Config holds the Redis server configuration
type Config struct {
	mu         sync.RWMutex
	Dir        string
	DBFilename string
	Port       int
	ReplicaOf  string // Format: "host port"
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
	flag.StringVar(&config.ReplicaOf, "replicaof", config.ReplicaOf, "Make this server a replica of <host> <port>")
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

// IsReplica returns true if this server is configured as a replica
func (config *Config) IsReplica() bool {
	config.mu.RLock()
	defer config.mu.RUnlock()
	return config.ReplicaOf != ""
}

// GetReplicaInfo parses and returns the master host and port
func (config *Config) GetReplicaInfo() (host string, port string) {
	config.mu.RLock()
	defer config.mu.RUnlock()

	if config.ReplicaOf == "" {
		return "", ""
	}

	// Parse "host port" format
	parts := strings.Fields(config.ReplicaOf)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return "", ""
}
