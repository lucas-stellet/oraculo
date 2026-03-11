package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

func setupCmd(t *testing.T, extraArgs ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(append([]string{"setup"}, extraArgs...))
	err := cmd.Execute()
	return buf.String(), err
}

func TestSetup_CreatesOraculoConfig(t *testing.T) {
	setupInstallDir(t)
	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(".oraculo", "config.json"))
	if err != nil {
		t.Fatalf("expected .oraculo/config.json: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := cfg["port"]; !ok {
		t.Error("missing 'port'")
	}
	if _, ok := cfg["preferred_language"]; !ok {
		t.Error("missing 'preferred_language'")
	}
}

func TestSetup_CreatesDatabase(t *testing.T) {
	setupInstallDir(t)
	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(".oraculo", "oraculo.db")); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db")
	}
}

func TestSetup_CreatesHooksOnly(t *testing.T) {
	setupInstallDir(t)
	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("expected .claude/settings.json: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := settings["hooks"]; !ok {
		t.Error("missing 'hooks'")
	}
	if _, ok := settings["mcpServers"]; ok {
		t.Error("setup should not configure mcpServers")
	}
}

func TestSetup_DoesNotCopySkills(t *testing.T) {
	setupInstallDir(t)
	_, err := setupCmd(t)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	matches, _ := filepath.Glob(filepath.Join(".claude", "skills", "oraculo-*"))
	if len(matches) != 0 {
		t.Errorf("setup should not copy skills, found: %v", matches)
	}
}
