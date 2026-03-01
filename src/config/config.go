// src/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

const configPath = ".oraculo/config.json"

// Config holds the runtime configuration for Oraculo.
type Config struct {
	Port int `json:"port"`
}

// Read reads the config from .oraculo/config.json in the current directory.
// Returns an empty Config and no error if the file does not exist.
func Read() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Write writes the config to .oraculo/config.json in the current directory.
// The .oraculo/ directory must already exist.
func Write(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// FindPort finds an available TCP port in the inclusive range [min, max].
// It returns an error if no port is available in the range.
func FindPort(min, max int) (int, error) {
	for port := min; port <= max; port++ {
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}
		ln.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no available port in range %d-%d", min, max)
}
