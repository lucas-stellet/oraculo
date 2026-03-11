package registry_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/registry"
)

func TestRegister_CreatesFileAndAddsEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	err := registry.Register(path, registry.Entry{
		Project: "test-project",
		Path:    "/tmp/test",
		Port:    3100,
		PID:     os.Getpid(),
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	entries, err := registry.List(path)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Project != "test-project" {
		t.Errorf("expected project 'test-project', got %q", entries[0].Project)
	}
	if entries[0].Port != 3100 {
		t.Errorf("expected port 3100, got %d", entries[0].Port)
	}
}

func TestUnregister_RemovesEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	_ = registry.Register(path, registry.Entry{Project: "a", Path: "/a", Port: 3100, PID: os.Getpid()})
	_ = registry.Register(path, registry.Entry{Project: "b", Path: "/b", Port: 3101, PID: os.Getpid()})

	err := registry.Unregister(path, "/a")
	if err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	entries, _ := registry.List(path)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "/b" {
		t.Errorf("expected /b, got %q", entries[0].Path)
	}
}

func TestRegister_UpdatesExistingEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	_ = registry.Register(path, registry.Entry{Project: "a", Path: "/a", Port: 3100, PID: 1})
	_ = registry.Register(path, registry.Entry{Project: "a", Path: "/a", Port: 3200, PID: 2})

	entries, _ := registry.List(path)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 3200 {
		t.Errorf("expected port 3200, got %d", entries[0].Port)
	}
}

func TestList_ReturnsNilForMissingFile(t *testing.T) {
	entries, err := registry.List("/nonexistent/servers.json")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}
