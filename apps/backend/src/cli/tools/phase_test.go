package tools_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

func initSessionForTest(t *testing.T, epicName, sessionType string) string {
	t.Helper()
	_, _ = executeCmd(t, "tools", "epic", "init", epicName)
	out, err := executeCmd(t, "tools", "session", "init", "--type", sessionType, "--epic", epicName)
	if err != nil {
		t.Fatalf("init session: %v\noutput: %s", err, out)
	}
	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	return result["id"].(string)
}

func executePhaseComplete(t *testing.T, sessionID, phase, data string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(data))
	cmd.SetArgs([]string{"tools", "phase", "complete", phase, "--session", sessionID})
	err := cmd.Execute()
	return buf.String(), err
}

func TestPhaseComplete(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "validate") // 3 phases: setup, qa-dispatch, verdict

	out, err := executePhaseComplete(t, sessionID, "setup", `{"level":"deep"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["phase"] != "setup" {
		t.Errorf("phase = %v, want %q", result["phase"], "setup")
	}
	if result["completed"] != true {
		t.Errorf("completed = %v, want true", result["completed"])
	}
	if result["next"] != "qa-dispatch" {
		t.Errorf("next = %v, want %q", result["next"], "qa-dispatch")
	}
}

func TestPhaseComplete_OutOfOrder(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "epic")

	out, err := executePhaseComplete(t, sessionID, "reframing", "{}")
	if err == nil {
		t.Fatal("expected error for out-of-order phase")
	}

	var result map[string]string
	json.Unmarshal([]byte(out), &result)
	if result["error"] != "invalid_transition" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_transition")
	}
}

func TestPhaseComplete_UnknownPhase(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "epic")

	out, err := executePhaseComplete(t, sessionID, "nonexistent", "{}")
	if err == nil {
		t.Fatal("expected error for unknown phase")
	}

	var result map[string]string
	json.Unmarshal([]byte(out), &result)
	if result["error"] != "invalid_phase" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_phase")
	}
}

func TestPhaseComplete_LastPhase(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "validate") // 3 phases

	executePhaseComplete(t, sessionID, "setup", "{}")
	executePhaseComplete(t, sessionID, "qa-dispatch", "{}")
	out, err := executePhaseComplete(t, sessionID, "verdict", "{}")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["next"] != "" {
		t.Errorf("next = %v, want empty", result["next"])
	}
	if result["session_closed"] != true {
		t.Errorf("session_closed = %v, want true", result["session_closed"])
	}
}
