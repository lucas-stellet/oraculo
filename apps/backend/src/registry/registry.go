package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// Entry represents a running oraculo server instance.
type Entry struct {
	Project   string    `json:"project"`
	Path      string    `json:"path"`
	Port      int       `json:"port"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

// Register adds or updates an entry in the registry file.
func Register(registryPath string, entry Entry) error {
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now()
	}
	return withLock(registryPath, func(entries []Entry) ([]Entry, error) {
		for i, e := range entries {
			if e.Path == entry.Path {
				entries[i] = entry
				return entries, nil
			}
		}
		return append(entries, entry), nil
	})
}

// Unregister removes the entry matching the given project path.
func Unregister(registryPath string, projectPath string) error {
	return withLock(registryPath, func(entries []Entry) ([]Entry, error) {
		for i, e := range entries {
			if e.Path == projectPath {
				return append(entries[:i], entries[i+1:]...), nil
			}
		}
		return entries, nil
	})
}

// List reads all entries from the registry file.
func List(registryPath string) ([]Entry, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// WriteAll atomically replaces all entries in the registry file.
func WriteAll(registryPath string, entries []Entry) error {
	return withLock(registryPath, func(_ []Entry) ([]Entry, error) {
		return entries, nil
	})
}

// DefaultPath returns ~/.oraculo/servers.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oraculo", "servers.json"), nil
}

func withLock(registryPath string, fn func([]Entry) ([]Entry, error)) error {
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return err
	}

	lock := flock.New(registryPath + ".lock")
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	entries, err := List(registryPath)
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
