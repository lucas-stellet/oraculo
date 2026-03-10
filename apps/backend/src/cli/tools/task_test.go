package tools_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

// ---------- helpers (reuse executeCmd, setupTestDir from epic_test.go) ----------

// initEpicAndStory creates an epic and story for task tests. Returns any error.
func initEpicAndStory(t *testing.T, epic, story string) {
	t.Helper()
	if _, err := executeCmd(t, "tools", "epic", "init", epic); err != nil {
		t.Fatalf("epic init %q: %v", epic, err)
	}
	if _, err := executeCmd(t, "tools", "story", "init", story, "--epic", epic); err != nil {
		t.Fatalf("story init %q: %v", story, err)
	}
}

// ---------- Task Init ----------

func TestTaskInit(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	out, err := executeCmd(t, "tools", "task", "init", "setup-db",
		"--epic", "my-epic", "--story", "login", "--description", "Set up database")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["name"] != "setup-db" {
		t.Errorf("name = %v, want %q", result["name"], "setup-db")
	}
	if result["epic"] != "my-epic" {
		t.Errorf("epic = %v, want %q", result["epic"], "my-epic")
	}
	if result["story"] != "login" {
		t.Errorf("story = %v, want %q", result["story"], "login")
	}
	if result["created"] != true {
		t.Errorf("created = %v, want true", result["created"])
	}
}

func TestTaskInit_Idempotent(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	_, _ = executeCmd(t, "tools", "task", "init", "setup-db",
		"--epic", "my-epic", "--story", "login")

	out, err := executeCmd(t, "tools", "task", "init", "setup-db",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["created"] != false {
		t.Errorf("created = %v, want false", result["created"])
	}
}

// ---------- Task Lifecycle: init -> start -> complete ----------

func TestTaskLifecycle_Complete(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	// init
	_, err := executeCmd(t, "tools", "task", "init", "build-ui",
		"--epic", "my-epic", "--story", "login", "--description", "Build the UI")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// start
	out, err := executeCmd(t, "tools", "task", "start", "build-ui",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("start: %v\noutput: %s", err, out)
	}

	var started map[string]any
	if err := json.Unmarshal([]byte(out), &started); err != nil {
		t.Fatalf("invalid JSON from start: %v\noutput: %s", err, out)
	}
	if started["status"] != "in_progress" {
		t.Errorf("Status after start = %v, want %q", started["status"], "in_progress")
	}

	// complete (requires stdin JSON)
	resultJSON := `{"summary":"Implemented login form","logs":"Created files","skills_used":["tdd"],"files_modified":["src/login.go"]}`
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(resultJSON))
	cmd.SetArgs([]string{"tools", "task", "complete", "build-ui",
		"--epic", "my-epic", "--story", "login"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("complete: %v\noutput: %s", err, buf.String())
	}

	var completed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &completed); err != nil {
		t.Fatalf("invalid JSON from complete: %v\noutput: %s", err, buf.String())
	}
	if completed["status"] != "completed" {
		t.Errorf("Status after complete = %v, want %q", completed["status"], "completed")
	}
}

// ---------- Task Lifecycle: init -> start -> fail ----------

func TestTaskLifecycle_Fail(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	// init
	_, err := executeCmd(t, "tools", "task", "init", "flaky-task",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// start
	_, err = executeCmd(t, "tools", "task", "start", "flaky-task",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// fail
	out, err := executeCmd(t, "tools", "task", "fail", "flaky-task",
		"--epic", "my-epic", "--story", "login", "--reason", "timeout exceeded")
	if err != nil {
		t.Fatalf("fail: %v\noutput: %s", err, out)
	}

	var failed map[string]any
	if err := json.Unmarshal([]byte(out), &failed); err != nil {
		t.Fatalf("invalid JSON from fail: %v\noutput: %s", err, out)
	}
	if failed["status"] != "failed" {
		t.Errorf("Status after fail = %v, want %q", failed["status"], "failed")
	}
	if failed["failure_reason"] != "timeout exceeded" {
		t.Errorf("FailureReason = %v, want %q", failed["failure_reason"], "timeout exceeded")
	}
}

// ---------- DAG: init with --depends-on ----------

func TestTaskInit_DependsOn(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	// Create the dependency task first.
	_, err := executeCmd(t, "tools", "task", "init", "setup-db",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("init setup-db: %v", err)
	}

	// Create a task that depends on setup-db.
	out, err := executeCmd(t, "tools", "task", "init", "build-ui",
		"--epic", "my-epic", "--story", "login",
		"--depends-on", "setup-db")
	if err != nil {
		t.Fatalf("init build-ui with depends-on: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["created"] != true {
		t.Errorf("created = %v, want true", result["created"])
	}
}

// ---------- Invalid Transitions ----------

func TestTaskStart_NotPending(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	// init + start
	_, _ = executeCmd(t, "tools", "task", "init", "task-a",
		"--epic", "my-epic", "--story", "login")
	_, _ = executeCmd(t, "tools", "task", "start", "task-a",
		"--epic", "my-epic", "--story", "login")

	// start again should fail (already in_progress)
	out, err := executeCmd(t, "tools", "task", "start", "task-a",
		"--epic", "my-epic", "--story", "login")
	if err == nil {
		t.Fatal("expected error for starting non-pending task")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] != "invalid_transition" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_transition")
	}
}

func TestTaskComplete_NotInProgress(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	// init only (pending, not in_progress)
	_, _ = executeCmd(t, "tools", "task", "init", "task-b",
		"--epic", "my-epic", "--story", "login")

	// complete on pending should fail
	resultJSON := `{"summary":"done","logs":"","skills_used":[],"files_modified":[]}`
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(resultJSON))
	cmd.SetArgs([]string{"tools", "task", "complete", "task-b",
		"--epic", "my-epic", "--story", "login"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for completing non-in_progress task")
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["error"] != "invalid_transition" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_transition")
	}
}

// ---------- Task List ----------

func TestTaskList(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	// Create several tasks with different statuses.
	_, _ = executeCmd(t, "tools", "task", "init", "task-1",
		"--epic", "my-epic", "--story", "login")
	_, _ = executeCmd(t, "tools", "task", "init", "task-2",
		"--epic", "my-epic", "--story", "login")
	_, _ = executeCmd(t, "tools", "task", "start", "task-2",
		"--epic", "my-epic", "--story", "login")

	out, err := executeCmd(t, "tools", "task", "list",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("list: %v\noutput: %s", err, out)
	}

	var tasks []map[string]any
	if err := json.Unmarshal([]byte(out), &tasks); err != nil {
		t.Fatalf("invalid JSON array: %v\noutput: %s", err, out)
	}
	if len(tasks) != 2 {
		t.Fatalf("len = %d, want 2", len(tasks))
	}

	// Verify we see both statuses.
	statuses := map[string]bool{}
	for _, task := range tasks {
		if s, ok := task["status"].(string); ok {
			statuses[s] = true
		}
	}
	if !statuses["pending"] {
		t.Error("expected at least one pending task in list")
	}
	if !statuses["in_progress"] {
		t.Error("expected at least one in_progress task in list")
	}
}

// ---------- Task Get ----------

func TestTaskGet(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	_, _ = executeCmd(t, "tools", "task", "init", "my-task",
		"--epic", "my-epic", "--story", "login", "--description", "do stuff")

	out, err := executeCmd(t, "tools", "task", "get", "my-task",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("get: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["name"] != "my-task" {
		t.Errorf("Name = %v, want %q", result["name"], "my-task")
	}
	if result["description"] != "do stuff" {
		t.Errorf("Description = %v, want %q", result["description"], "do stuff")
	}
	if result["status"] != "pending" {
		t.Errorf("Status = %v, want %q", result["status"], "pending")
	}
}

// ---------- Task Delete ----------

func TestTaskDelete(t *testing.T) {
	setupTestDir(t)
	initEpicAndStory(t, "my-epic", "login")

	_, _ = executeCmd(t, "tools", "task", "init", "doomed",
		"--epic", "my-epic", "--story", "login")

	out, err := executeCmd(t, "tools", "task", "delete", "doomed",
		"--epic", "my-epic", "--story", "login")
	if err != nil {
		t.Fatalf("delete: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["name"] != "doomed" {
		t.Errorf("name = %v, want %q", result["name"], "doomed")
	}
	if result["epic"] != "my-epic" {
		t.Errorf("epic = %v, want %q", result["epic"], "my-epic")
	}
	if result["story"] != "login" {
		t.Errorf("story = %v, want %q", result["story"], "login")
	}
	if result["deleted"] != true {
		t.Errorf("deleted = %v, want true", result["deleted"])
	}

	// Verify it's really gone.
	out2, err := executeCmd(t, "tools", "task", "get", "doomed",
		"--epic", "my-epic", "--story", "login")
	if err == nil {
		t.Fatal("expected error getting deleted task")
	}
	var errResult map[string]string
	if err := json.Unmarshal([]byte(out2), &errResult); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out2)
	}
	if errResult["error"] != "not_found" {
		t.Errorf("error = %q, want %q", errResult["error"], "not_found")
	}
}

// ---------- Error: Epic Not Found ----------

func TestTaskInit_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "task", "init", "task-x",
		"--epic", "nonexistent", "--story", "whatever")
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

// ---------- Error: Story Not Found ----------

func TestTaskInit_StoryNotFound(t *testing.T) {
	setupTestDir(t)

	// Create epic but not story.
	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "task", "init", "task-x",
		"--epic", "my-epic", "--story", "nonexistent")
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
