package tools_test

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidationSave(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	out, err := executeCmd(t,
		"tools", "validation", "save", "login",
		"--epic", "my-epic",
		"--verdict", "approved",
		"--findings", `{"blocker":0,"major":0,"minor":1}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["Verdict"] != "approved" {
		t.Errorf("Verdict = %v, want %q", result["Verdict"], "approved")
	}
	if result["Findings"] != `{"blocker":0,"major":0,"minor":1}` {
		t.Errorf("Findings = %v", result["Findings"])
	}
}

func TestValidationSave_Rejected(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	out, err := executeCmd(t,
		"tools", "validation", "save", "login",
		"--epic", "my-epic",
		"--verdict", "rejected",
		"--findings", `{"blocker":2}`,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["Verdict"] != "rejected" {
		t.Errorf("Verdict = %v, want %q", result["Verdict"], "rejected")
	}
}

func TestValidationSave_InvalidVerdict(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	out, err := executeCmd(t,
		"tools", "validation", "save", "login",
		"--epic", "my-epic",
		"--verdict", "maybe",
	)
	if err == nil {
		t.Fatal("expected error for invalid verdict")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if !strings.Contains(result["message"], "invalid verdict") {
		t.Errorf("message = %q, expected to contain 'invalid verdict'", result["message"])
	}
}

func TestValidationSave_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t,
		"tools", "validation", "save", "login",
		"--epic", "nonexistent",
		"--verdict", "approved",
	)
	if err == nil {
		t.Fatal("expected error for nonexistent epic")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] != "not_found" {
		t.Errorf("error = %q, want %q", result["error"], "not_found")
	}
}

func TestValidationSave_StoryNotFound(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t,
		"tools", "validation", "save", "nonexistent",
		"--epic", "my-epic",
		"--verdict", "approved",
	)
	if err == nil {
		t.Fatal("expected error for nonexistent story")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] != "not_found" {
		t.Errorf("error = %q, want %q", result["error"], "not_found")
	}
}
