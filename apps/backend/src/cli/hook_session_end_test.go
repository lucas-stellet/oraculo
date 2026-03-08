package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
	"github.com/lucas/oraculo/apps/backend/src/db"
)

func TestHookSessionEnd_UpdatesEndedAt(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// Seed a session first
	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES ('s1')")
	database.Close()

	stdinJSON := `{"session_id":"s1","hook_event_name":"SessionEnd","reason":"prompt_input_exit"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-end"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify ended_at was set
	database, err = db.Open()
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer database.Close()

	var endedAt *string
	database.Conn().QueryRow("SELECT ended_at FROM claude_sessions WHERE id = 's1'").Scan(&endedAt)
	if endedAt == nil {
		t.Fatal("expected ended_at to be set")
	}
}

func TestHookSessionEnd_RecordsSessionEvent(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES ('s1')")
	database.Close()

	stdinJSON := `{"session_id":"s1","hook_event_name":"SessionEnd","reason":"clear"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-end"})
	cmd.Execute()

	database, err = db.Open()
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer database.Close()

	var eventType, payload string
	err = database.Conn().QueryRow(
		"SELECT event_type, payload FROM session_events WHERE session_id = 's1'",
	).Scan(&eventType, &payload)
	if err != nil {
		t.Fatalf("session event not found: %v", err)
	}
	if eventType != "session_end" {
		t.Errorf("expected event_type=session_end, got %s", eventType)
	}
	if !strings.Contains(payload, "clear") {
		t.Errorf("expected payload to contain reason, got %s", payload)
	}
}

func TestHookSessionEnd_AlwaysExitsZero(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// No session in DB — should still exit 0
	stdinJSON := `{"session_id":"nonexistent","hook_event_name":"SessionEnd","reason":"other"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-end"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook must always exit 0, got: %v", err)
	}
}
