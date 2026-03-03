package cli_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	claudekit "github.com/lucas/oraculo/claude-kit"
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
	for _, hookName := range requiredHooks {
		val, ok := hooks[hookName]
		if !ok {
			t.Errorf("hooks missing %q entry", hookName)
			continue
		}
		// Each hook event must be an array of hook groups (new format).
		groups, ok := val.([]any)
		if !ok || len(groups) == 0 {
			t.Errorf("hooks[%q] must be a non-empty array", hookName)
			continue
		}
		group, ok := groups[0].(map[string]any)
		if !ok {
			t.Errorf("hooks[%q][0] must be an object", hookName)
			continue
		}
		if _, ok := group["hooks"]; !ok {
			t.Errorf("hooks[%q][0] missing 'hooks' field", hookName)
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

func TestInstall_MCPServerArgs(t *testing.T) {
	setupInstallDir(t)
	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	oraculo, ok := settings.MCPServers["oraculo"]
	if !ok {
		t.Fatal("mcpServers.oraculo not found")
	}
	if len(oraculo.Args) != 2 || oraculo.Args[0] != "start" || oraculo.Args[1] != "mcp" {
		t.Errorf("expected args [start mcp], got %v", oraculo.Args)
	}
}

func TestInstall_CopiesEmbeddedOraculoSkills(t *testing.T) {
	setupInstallDir(t)

	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Walk the embedded FS and verify every file was copied to .claude/skills/oraculo-<name>/
	err = fs.WalkDir(claudekit.SkillsFS, filepath.Join("skills", "oraculo"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		// path: skills/oraculo/epic/SKILL.md
		// rel:  epic/SKILL.md
		rel, err := filepath.Rel(filepath.Join("skills", "oraculo"), path)
		if err != nil {
			return err
		}
		// skills/oraculo/epic/SKILL.md → .claude/skills/oraculo-epic/SKILL.md
		parts := strings.SplitN(rel, string(filepath.Separator), 2)
		var destPath string
		if len(parts) == 2 {
			destPath = filepath.Join(".claude", "skills", "oraculo-"+parts[0], parts[1])
		} else {
			destPath = filepath.Join(".claude", "skills", "oraculo-"+parts[0])
		}
		got, err := os.ReadFile(destPath)
		if err != nil {
			t.Errorf("expected %s to exist: %v", destPath, err)
			return nil
		}
		want, err := fs.ReadFile(claudekit.SkillsFS, path)
		if err != nil {
			return err
		}
		if string(got) != string(want) {
			t.Errorf("content mismatch for %s", destPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded FS: %v", err)
	}
}
