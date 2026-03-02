package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

// installCmd builds a fresh root command, sets args to "install", captures stdout, and executes.
func installCmd(t *testing.T) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"install"})
	err := cmd.Execute()
	return buf.String(), err
}

func setupInstallDir(t *testing.T) string {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	return tmp
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestInstall_CreatesOraculoConfig(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	configPath := filepath.Join(".oraculo", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected .oraculo/config.json to exist: %v", err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.json is not valid JSON: %v", err)
	}

	port, ok := cfg["port"]
	if !ok {
		t.Fatal("config.json missing 'port' field")
	}
	portNum, ok := port.(float64)
	if !ok || portNum <= 0 {
		t.Fatalf("config.json 'port' is not a positive number: %v", port)
	}
}

func TestInstall_CreatesDatabase(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	dbPath := filepath.Join(".oraculo", "oraculo.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db to be created")
	}
}

func TestInstall_CreatesClaudeSettings(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	settingsPath := filepath.Join(".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("expected .claude/settings.json to exist: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}

	// Verify hooks section exists
	if _, ok := settings["hooks"]; !ok {
		t.Error("settings.json missing 'hooks' section")
	}

	// Verify mcpServers section exists
	if _, ok := settings["mcpServers"]; !ok {
		t.Error("settings.json missing 'mcpServers' section")
	}
}

func TestInstall_PortConsistency(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Read config.json port
	configData, err := os.ReadFile(filepath.Join(".oraculo", "config.json"))
	if err != nil {
		t.Fatalf("read config.json: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(configData, &cfg); err != nil {
		t.Fatalf("parse config.json: %v", err)
	}
	configPort := cfg["port"].(float64)

	// Read settings.json and verify oraculo MCP server args contain "start"
	settingsData, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	// Verify port is in valid range
	if configPort < 3100 || configPort > 3199 {
		t.Errorf("config port %v is not in range [3100, 3199]", configPort)
	}

	// Verify mcpServers has oraculo entry
	mcpServers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers is not an object")
	}
	oraculoServer, ok := mcpServers["oraculo"]
	if !ok {
		t.Fatal("mcpServers missing 'oraculo' entry")
	}
	serverMap, ok := oraculoServer.(map[string]any)
	if !ok {
		t.Fatal("oraculo server entry is not an object")
	}
	if serverMap["command"] != "oraculo" {
		t.Errorf("oraculo MCP server command = %v, want 'oraculo'", serverMap["command"])
	}
}

func TestInstall_HooksConfiguration(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	settingsData, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks is not an object")
	}

	requiredHooks := []string{"PreToolUse", "PostToolUse", "SessionStart", "SessionEnd"}
	for _, hook := range requiredHooks {
		if _, ok := hooks[hook]; !ok {
			t.Errorf("hooks missing %q entry", hook)
		}
	}
}

func TestInstall_Idempotent(t *testing.T) {
	setupInstallDir(t)

	// Run install twice
	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	_, err = installCmd(t)
	if err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	// Both files should still exist
	if _, err := os.Stat(filepath.Join(".oraculo", "config.json")); os.IsNotExist(err) {
		t.Error("config.json missing after second install")
	}
	if _, err := os.Stat(filepath.Join(".claude", "settings.json")); os.IsNotExist(err) {
		t.Error("settings.json missing after second install")
	}
}

func TestInstall_CopiesLocalOraculoSkills(t *testing.T) {
	setupInstallDir(t)

	sourcePath := filepath.Join("claude-kit", "skills", "oraculo", "epic", "SKILL.md")
	sourceContent := "---\nname: oraculo:epic\n---\n# Epic\n"
	writeFixtureFile(t, sourcePath, sourceContent)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	destPath := filepath.Join(".claude", "skills", "oraculo", "epic", "SKILL.md")
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", destPath, err)
	}
	if string(data) != sourceContent {
		t.Fatalf("copied skill content mismatch:\nwant:\n%s\ngot:\n%s", sourceContent, string(data))
	}
}
