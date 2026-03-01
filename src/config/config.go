// src/config/config.go
package config

// Config holds the runtime configuration for Oraculo.
type Config struct {
	Port int
}

// Read reads the configuration from the environment and returns a Config.
// If no port is configured, Port is 0 (the caller should default to 3100).
func Read() (*Config, error) {
	return &Config{}, nil
}
