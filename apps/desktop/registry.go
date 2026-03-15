package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// registryEntry represents a running oraculo server instance.
type registryEntry struct {
	Project   string    `json:"project"`
	Path      string    `json:"path"`
	Port      int       `json:"port"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

// registryList reads all entries from the registry file.
func registryList(registryPath string) ([]registryEntry, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var entries []registryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// registryWriteAll atomically replaces all entries in the registry file.
func registryWriteAll(registryPath string, entries []registryEntry) error {
	return registryWithLock(registryPath, func(_ []registryEntry) ([]registryEntry, error) {
		return entries, nil
	})
}

// registryDefaultPath returns ~/.oraculo/servers.json.
func registryDefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oraculo", "servers.json"), nil
}

func registryWithLock(registryPath string, fn func([]registryEntry) ([]registryEntry, error)) error {
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return err
	}

	lock := flock.New(registryPath + ".lock")
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	entries, err := registryList(registryPath)
	if err != nil {
		return err
	}

	entries, err = fn(entries)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(registryPath, data, 0o644)
}
