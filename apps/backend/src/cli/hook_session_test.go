package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
	"github.com/lucas/oraculo/apps/backend/src/db"
)

// stubSpawnDaemon replaces the real daemon spawner with a stub that returns
// an error, preventing fork-bomb behavior when os.Executable() returns the
// test binary and avoiding the 10s pollHealth wait.
func stubSpawnDaemon(t *testing.T) {
	t.Helper()
	orig := cli.SpawnDaemon
	cli.SpawnDaemon = func() error {
		return fmt.Errorf("stub: daemon spawn disabled in tests")
	}
	t.Cleanup(func() { cli.SpawnDaemon = orig })
}

func TestHookSessionStart_NoConfig(t *testing.T) {
	// Setup temp dir
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	stdinJSON := `{"session_id":"test-no-config","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	// Must always succeed (exit 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, buf.String())
	}

	// Should have created .oraculo/ and registered session
	dbPath := filepath.Join(".oraculo", "oraculo.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db to be created")
	}
}

func TestHookSessionStart_AlertsWhenServerOffline(t *testing.T) {
	stubSpawnDaemon(t)
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// Write config with port that nothing is listening on
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)
	os.WriteFile(
		filepath.Join(tmp, ".oraculo", "config.json"),
		[]byte(`{"port": 39999}`),
		0o644,
	)

	stdinJSON := `{"session_id":"test-alerts","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook should never return error: %v", err)
	}
	stderr := buf.String()
	if !strings.Contains(stderr, "warning") {
		t.Errorf("expected warning in output, got: %s", stderr)
	}
}

func TestHookSessionStart_AttemptsAutoStart(t *testing.T) {
	stubSpawnDaemon(t)
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)
	os.WriteFile(
		filepath.Join(tmp, ".oraculo", "config.json"),
		[]byte(`{"port": 39999}`),
		0o644,
	)

	stdinJSON := `{"session_id":"test-autostart","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	cmd.Execute()

	output := buf.String()
	// Should attempt auto-start and report failure (oraculo binary may not exist in test)
	if !strings.Contains(output, "auto-start") && !strings.Contains(output, "starting") {
		t.Errorf("expected auto-start attempt in output, got: %s", output)
	}
}

func TestHookSessionStart_AlwaysExitsZero(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	stdinJSON := `{"session_id":"test-exit-zero","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook must always exit 0, got: %v", err)
	}
}

func TestHookSessionStart_UsesClaudeSessionID(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	stdinJSON := `{"session_id":"claude-abc-123","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	var id string
	err = database.Conn().QueryRow(
		"SELECT id FROM claude_sessions WHERE id = ?", "claude-abc-123",
	).Scan(&id)
	if err != nil {
		t.Fatalf("session not found with Claude's session_id: %v", err)
	}
}

func TestHookSessionStart_UpsertOnResume(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	startupJSON := `{"session_id":"claude-abc-123","hook_event_name":"SessionStart","source":"startup"}`
	cmd1 := cli.NewRoot("test")
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetErr(&bytes.Buffer{})
	cmd1.SetIn(strings.NewReader(startupJSON))
	cmd1.SetArgs([]string{"hook", "session-start"})
	cmd1.Execute()

	resumeJSON := `{"session_id":"claude-abc-123","hook_event_name":"SessionStart","source":"resume"}`
	cmd2 := cli.NewRoot("test")
	cmd2.SetOut(&bytes.Buffer{})
	cmd2.SetErr(&bytes.Buffer{})
	cmd2.SetIn(strings.NewReader(resumeJSON))
	cmd2.SetArgs([]string{"hook", "session-start"})
	err := cmd2.Execute()

	if err != nil {
		t.Fatalf("resume should succeed: %v", err)
	}

	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	var count int
	database.Conn().QueryRow("SELECT COUNT(*) FROM claude_sessions WHERE id = ?", "claude-abc-123").Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 session row, got %d", count)
	}
}
