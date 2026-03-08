package tools_test

import (
	"encoding/json"
	"testing"
)

func TestSessionInit(t *testing.T) {
	setupTestDir(t)

	// Create the epic first
	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["type"] != "epic" {
		t.Errorf("type = %v, want %q", result["type"], "epic")
	}
	if result["epic"] != "my-epic" {
		t.Errorf("epic = %v, want %q", result["epic"], "my-epic")
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want %q", result["status"], "active")
	}
	if result["created"] != true {
		t.Errorf("created = %v, want true", result["created"])
	}
	if result["id"] == nil || result["id"] == "" {
		t.Error("expected non-empty id")
	}
}

func TestSessionInit_Description(t *testing.T) {
	setupTestDir(t)

	// session init creates the epic if it doesn't exist, passing description through
	out, err := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "desc-epic", "--description", "A raw idea about improving auth")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	// Verify description was persisted by reading the epic back via list
	listOut, err := executeCmd(t, "tools", "epic", "list")
	if err != nil {
		t.Fatalf("list error: %v\noutput: %s", err, listOut)
	}

	var epics []map[string]any
	if err := json.Unmarshal([]byte(listOut), &epics); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, listOut)
	}

	var found bool
	for _, e := range epics {
		if e["Name"] == "desc-epic" {
			found = true
			if e["Description"] != "A raw idea about improving auth" {
				t.Errorf("Description = %v, want %q", e["Description"], "A raw idea about improving auth")
			}
			break
		}
	}
	if !found {
		t.Fatal("epic desc-epic not found in list")
	}
}

func TestSessionInit_Idempotent(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["created"] != false {
		t.Errorf("created = %v, want false", result["created"])
	}
}

func TestSessionStatus_Active(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "session", "status", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}
	if result["current_phase"] != "setup" {
		t.Errorf("current_phase = %v, want %q", result["current_phase"], "setup")
	}
}

func TestSessionStatus_NoSession(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "session", "status", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["active"] != false {
		t.Errorf("active = %v, want false", result["active"])
	}
}

func TestSessionState(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	initOut, _ := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	var initResult map[string]any
	json.Unmarshal([]byte(initOut), &initResult)
	sessionID := initResult["id"].(string)

	out, err := executeCmd(t, "tools", "session", "state", "--session", sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["id"] != sessionID {
		t.Errorf("id = %v, want %q", result["id"], sessionID)
	}
	if result["type"] != "epic" {
		t.Errorf("type = %v, want %q", result["type"], "epic")
	}
	if result["current_phase"] != "setup" {
		t.Errorf("current_phase = %v, want %q", result["current_phase"], "setup")
	}
}

func TestSessionClose(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	initOut, _ := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	var initResult map[string]any
	json.Unmarshal([]byte(initOut), &initResult)
	sessionID := initResult["id"].(string)

	out, err := executeCmd(t, "tools", "session", "close", "--session", sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["closed"] != true {
		t.Errorf("closed = %v, want true", result["closed"])
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want %q", result["status"], "completed")
	}
}

func TestSessionClose_Abandoned(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	initOut, _ := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	var initResult map[string]any
	json.Unmarshal([]byte(initOut), &initResult)
	sessionID := initResult["id"].(string)

	out, err := executeCmd(t, "tools", "session", "close", "--session", sessionID, "--reason", "abandoned")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["status"] != "abandoned" {
		t.Errorf("status = %v, want %q", result["status"], "abandoned")
	}
}
