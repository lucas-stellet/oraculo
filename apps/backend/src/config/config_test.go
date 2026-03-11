package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/config"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })
	return tmp
}

func TestReadMissing(t *testing.T) {
	setupTestDir(t)
	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if cfg.Port != 0 {
		t.Errorf("Port = %d, want 0", cfg.Port)
	}
}

func TestWriteAndRead(t *testing.T) {
	tmp := setupTestDir(t)
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)

	err := config.Write(&config.Config{Port: 3142})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if cfg.Port != 3142 {
		t.Errorf("Port = %d, want 3142", cfg.Port)
	}
}

func TestPreferredLanguageRoundTrip(t *testing.T) {
	tmp := setupTestDir(t)
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)

	err := config.Write(&config.Config{Port: 3142, PreferredLanguage: "pt-BR"})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if cfg.PreferredLanguage != "pt-BR" {
		t.Errorf("PreferredLanguage = %q, want %q", cfg.PreferredLanguage, "pt-BR")
	}
}

func TestPreferredLanguageOmitEmpty(t *testing.T) {
	tmp := setupTestDir(t)
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)

	err := config.Write(&config.Config{Port: 3142})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".oraculo", "config.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got := string(data); strings.Contains(got, "preferred_language") {
		t.Errorf("expected preferred_language to be omitted, got: %s", got)
	}
}

func TestConfig_ProjectName_UsesConfiguredName(t *testing.T) {
	cfg := &config.Config{Name: "my-project"}
	if got := cfg.ProjectName(); got != "my-project" {
		t.Errorf("expected 'my-project', got %q", got)
	}
}

func TestConfig_ProjectName_FallsBackToBasename(t *testing.T) {
	cfg := &config.Config{}
	name := cfg.ProjectName()
	if name == "" || name == "unknown" {
		t.Errorf("expected a non-empty basename, got %q", name)
	}
}

func TestFindPort(t *testing.T) {
	port, err := config.FindPort(30000, 30099)
	if err != nil {
		t.Fatalf("FindPort: %v", err)
	}
	if port < 30000 || port > 30099 {
		t.Errorf("Port %d outside range 30000-30099", port)
	}
}
