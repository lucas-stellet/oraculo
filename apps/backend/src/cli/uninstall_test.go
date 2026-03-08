package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

// uninstallCmd builds a fresh root command with given args and executes it.
func uninstallCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(append([]string{"uninstall"}, args...))
	err := cmd.Execute()
	return buf.String(), err
}

// writeSettings writes a settings.json file under .claude/ in the current dir.
func writeSettings(t *testing.T, data map[string]any) {
	t.Helper()
	if err := os.MkdirAll(".claude", 0o755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}
	if err := os.WriteFile(filepath.Join(".claude", "settings.json"), b, 0o644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}
}

// readSettings reads .claude/settings.json and returns it as a map.
func readSettings(t *testing.T) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("unmarshal settings.json: %v", err)
	}
	return data
}

func TestUninstall_RemovesHooksAndMCPFromSettings(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Write a settings.json with Oraculo hooks and mcpServers,
	// plus an unrelated entry that must be preserved.
	writeSettings(t, map[string]any{
		"hooks": map[string]any{
			"PreToolUse":   []any{"oraculo hook pre-tool-use"},
			"PostToolUse":  []any{"some-other-hook"},
			"SessionStart": []any{"oraculo hook session-start"},
		},
		"mcpServers": map[string]any{
			"oraculo":    map[string]any{"command": "oraculo mcp"},
			"other-tool": map[string]any{"command": "other mcp"},
		},
	})

	out, err := uninstallCmd(t)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	settings := readSettings(t)

	// Verify mcpServers["oraculo"] was removed.
	mcpServers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("mcpServers should still be a map")
	}
	if _, found := mcpServers["oraculo"]; found {
		t.Error("expected mcpServers['oraculo'] to be removed")
	}
	// Verify unrelated mcpServer was preserved.
	if _, found := mcpServers["other-tool"]; !found {
		t.Error("expected mcpServers['other-tool'] to be preserved")
	}

	// Verify hooks: Oraculo entries removed, other entries preserved.
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks should still be a map")
	}
	// PostToolUse had only a non-oraculo hook — should still have it.
	postToolUse, ok := hooks["PostToolUse"].([]any)
	if !ok {
		t.Fatalf("hooks['PostToolUse'] should be a slice, got %T", hooks["PostToolUse"])
	}
	if len(postToolUse) != 1 || postToolUse[0] != "some-other-hook" {
		t.Errorf("hooks['PostToolUse'] should contain 'some-other-hook', got: %v", postToolUse)
	}
	// PreToolUse had only an Oraculo hook — after removal it should be absent or empty.
	if preToolUse, exists := hooks["PreToolUse"]; exists {
		entries, ok := preToolUse.([]any)
		if ok && len(entries) > 0 {
			t.Errorf("hooks['PreToolUse'] should be empty after uninstall, got: %v", entries)
		}
	}
	// SessionStart had only an Oraculo hook — after removal it should be absent or empty.
	if sessionStart, exists := hooks["SessionStart"]; exists {
		entries, ok := sessionStart.([]any)
		if ok && len(entries) > 0 {
			t.Errorf("hooks['SessionStart'] should be empty after uninstall, got: %v", entries)
		}
	}
}

func TestUninstall_PreservesOraculoDir_WithoutPurge(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Create .oraculo/ directory to simulate an installed state.
	if err := os.MkdirAll(".oraculo", 0o755); err != nil {
		t.Fatalf("mkdir .oraculo: %v", err)
	}
	dbFile := filepath.Join(".oraculo", "oraculo.db")
	if err := os.WriteFile(dbFile, []byte("fake db"), 0o644); err != nil {
		t.Fatalf("create fake db: %v", err)
	}

	out, err := uninstallCmd(t)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	// .oraculo/ must still exist without --purge.
	if _, statErr := os.Stat(".oraculo"); os.IsNotExist(statErr) {
		t.Error("expected .oraculo/ to be preserved (no --purge)")
	}
	if _, statErr := os.Stat(dbFile); os.IsNotExist(statErr) {
		t.Error("expected .oraculo/oraculo.db to be preserved (no --purge)")
	}
}

func TestUninstall_RemovesOraculoDir_WithPurge(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Create .oraculo/ directory to simulate an installed state.
	if err := os.MkdirAll(".oraculo", 0o755); err != nil {
		t.Fatalf("mkdir .oraculo: %v", err)
	}
	dbFile := filepath.Join(".oraculo", "oraculo.db")
	if err := os.WriteFile(dbFile, []byte("fake db"), 0o644); err != nil {
		t.Fatalf("create fake db: %v", err)
	}

	out, err := uninstallCmd(t, "--purge")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	// .oraculo/ must be removed with --purge.
	if _, statErr := os.Stat(".oraculo"); !os.IsNotExist(statErr) {
		t.Error("expected .oraculo/ to be removed with --purge")
	}
}

func TestUninstall_NoSettingsFile_DoesNotError(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// No .claude/settings.json exists — should not error.
	out, err := uninstallCmd(t)
	if err != nil {
		t.Fatalf("expected no error when settings.json is absent, got: %v\noutput: %s", err, out)
	}
}

func TestUninstall_RemovesSkillsDirectory(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Create .claude/skills/oraculo/ to simulate installed skills.
	skillsDir := filepath.Join(".claude", "skills", "oraculo")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	skillFile := filepath.Join(skillsDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte("# skill"), 0o644); err != nil {
		t.Fatalf("create skill file: %v", err)
	}

	out, err := uninstallCmd(t)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	// .claude/skills/oraculo/ must be removed.
	if _, statErr := os.Stat(skillsDir); !os.IsNotExist(statErr) {
		t.Error("expected .claude/skills/oraculo/ to be removed")
	}
}
