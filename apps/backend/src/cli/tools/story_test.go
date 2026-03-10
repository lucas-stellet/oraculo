package tools_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

// ---------- Story Init ----------

func TestStoryInit(t *testing.T) {
	setupTestDir(t)

	// Create the parent epic first.
	_, err := executeCmd(t, "tools", "epic", "init", "my-epic", "--description", "parent epic")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	out, err := executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic", "--description", "Login flow")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["name"] != "login" {
		t.Errorf("name = %v, want %q", result["name"], "login")
	}
	if result["epic"] != "my-epic" {
		t.Errorf("epic = %v, want %q", result["epic"], "my-epic")
	}
	expectedPath := filepath.Join(".oraculo", "epics", "my-epic", "stories", "login")
	if result["path"] != expectedPath {
		t.Errorf("path = %v, want %q", result["path"], expectedPath)
	}
	if result["created"] != true {
		t.Errorf("created = %v, want true", result["created"])
	}
}

func TestStoryInit_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "story", "init", "login", "--epic", "nonexistent", "--description", "x")
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

func TestStoryInit_Idempotent(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	_, err := executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic", "--description", "first")
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	out, err := executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic", "--description", "second")
	if err != nil {
		t.Fatalf("second init: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["created"] != false {
		t.Errorf("created = %v, want false", result["created"])
	}
}

// ---------- Story Save ----------

func TestStorySave(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	content := "# Story Requirements\n\nLogin flow details."
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "story", "save", "login", "--epic", "my-epic"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("save: %v\noutput: %s", err, buf.String())
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["saved"] != true {
		t.Errorf("saved = %v, want true", result["saved"])
	}
	if result["epic"] != "my-epic" {
		t.Errorf("epic = %v, want %q", result["epic"], "my-epic")
	}
	expectedPath := filepath.Join(".oraculo", "epics", "my-epic", "stories", "login", "requirements.md")
	if result["path"] != expectedPath {
		t.Errorf("path = %v, want %q", result["path"], expectedPath)
	}

	// Verify the file was actually written.
	got, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != content {
		t.Errorf("file content = %q, want %q", got, content)
	}
}

func TestStorySave_CreatesDBRecord(t *testing.T) {
	setupTestDir(t)

	// Create the parent epic but do NOT call story init.
	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	content := "# Requirements\n\nSave-only story."
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "story", "save", "new-story", "--epic", "my-epic"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("save: %v\noutput: %s", err, buf.String())
	}

	// Verify the DB record was created by listing stories.
	out, err := executeCmd(t, "tools", "story", "list", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("list: %v\noutput: %s", err, out)
	}

	var stories []map[string]any
	if err := json.Unmarshal([]byte(out), &stories); err != nil {
		t.Fatalf("invalid JSON array: %v\noutput: %s", err, out)
	}
	if len(stories) != 1 {
		t.Fatalf("len = %d, want 1", len(stories))
	}
	if stories[0]["name"] != "new-story" {
		t.Errorf("Name = %v, want %q", stories[0]["name"], "new-story")
	}
}

func TestStorySave_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("content"))
	cmd.SetArgs([]string{"tools", "story", "save", "login", "--epic", "nonexistent"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent epic")
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["error"] != "not_found" {
		t.Errorf("error = %q, want %q", result["error"], "not_found")
	}
}

// ---------- Story Get ----------

func TestStoryGet(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	// Write requirements file manually.
	content := "# Login Story\n\nRequirements here."
	dir := filepath.Join(".oraculo", "epics", "my-epic", "stories", "login")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "requirements.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCmd(t, "tools", "story", "get", "login", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("get: %v\noutput: %s", err, out)
	}

	// get returns raw markdown, NOT JSON.
	if out != content {
		t.Errorf("output = %q, want %q", out, content)
	}
}

func TestStoryGet_NotFound(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "story", "get", "nonexistent", "--epic", "my-epic")
	if err == nil {
		t.Fatal("expected error for nonexistent story file")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] == "" {
		t.Error("expected non-empty error code")
	}
}

func TestStoryGet_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "story", "get", "login", "--epic", "nonexistent")
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

// ---------- Story List ----------

func TestStoryList(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "alpha", "--epic", "my-epic", "--description", "first")
	_, _ = executeCmd(t, "tools", "story", "init", "beta", "--epic", "my-epic", "--description", "second")

	out, err := executeCmd(t, "tools", "story", "list", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("list: %v\noutput: %s", err, out)
	}

	var stories []map[string]any
	if err := json.Unmarshal([]byte(out), &stories); err != nil {
		t.Fatalf("invalid JSON array: %v\noutput: %s", err, out)
	}
	if len(stories) != 2 {
		t.Fatalf("len = %d, want 2", len(stories))
	}
}

func TestStoryList_Empty(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "story", "list", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("list: %v\noutput: %s", err, out)
	}

	out = strings.TrimSpace(out)
	if out != "null" && out != "[]" {
		t.Errorf("expected null or [], got: %s", out)
	}
}

func TestStoryList_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "story", "list", "--epic", "nonexistent")
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

// ---------- Story Update ----------

func TestStoryUpdate(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic", "--description", "old")

	out, err := executeCmd(t, "tools", "story", "update", "login", "--epic", "my-epic", "--description", "new desc")
	if err != nil {
		t.Fatalf("update: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["description"] != "new desc" {
		t.Errorf("Description = %v, want %q", result["description"], "new desc")
	}
}

func TestStoryUpdate_NotFound(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "story", "update", "nonexistent", "--epic", "my-epic", "--description", "x")
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

func TestStoryUpdate_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "story", "update", "login", "--epic", "nonexistent", "--description", "x")
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

// ---------- Story Delete ----------

func TestStoryDelete(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "doomed", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "story", "delete", "doomed", "--epic", "my-epic")
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
	if result["deleted"] != true {
		t.Errorf("deleted = %v, want true", result["deleted"])
	}
}

func TestStoryDelete_NotFound(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "story", "delete", "nonexistent", "--epic", "my-epic")
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

func TestStoryDelete_EpicNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "story", "delete", "login", "--epic", "nonexistent")
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

// ---------- Story Update Status ----------

func TestStoryUpdateStatus(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "story", "update-status", "login", "--epic", "my-epic", "--status", "approved")
	if err != nil {
		t.Fatalf("update-status: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["approval_status"] != "approved" {
		t.Errorf("ApprovalStatus = %v, want %q", result["approval_status"], "approved")
	}
}

func TestStoryUpdateStatus_InvalidStatus(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "story", "update-status", "login", "--epic", "my-epic", "--status", "bogus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if !strings.Contains(result["message"], "invalid status") {
		t.Errorf("message = %q, expected to contain 'invalid status'", result["message"])
	}
}

func TestStoryUpdateStatus_NotFound(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "story", "update-status", "nonexistent", "--epic", "my-epic", "--status", "approved")
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

// ---------- Story Version ----------

func TestStoryVersion(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	content := "# Story v1\n\nFirst version."
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "story", "version", "login", "--epic", "my-epic"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version: %v\noutput: %s", err, buf.String())
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["id"] == nil || result["id"].(float64) < 1 {
		t.Errorf("expected valid ID, got %v", result["id"])
	}
	if result["number"] != float64(1) {
		t.Errorf("Number = %v, want 1", result["number"])
	}
	if result["story_id"] == nil {
		t.Error("expected StoryID to be set")
	}
}

func TestStoryVersions(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "story", "init", "login", "--epic", "my-epic")

	for _, content := range []string{"v1", "v2"} {
		var buf bytes.Buffer
		cmd := cli.NewRoot("test")
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetIn(strings.NewReader(content))
		cmd.SetArgs([]string{"tools", "story", "version", "login", "--epic", "my-epic"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("version: %v", err)
		}
	}

	out, err := executeCmd(t, "tools", "story", "versions", "login", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("versions: %v\noutput: %s", err, out)
	}

	var versions []map[string]any
	if err := json.Unmarshal([]byte(out), &versions); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(versions) != 2 {
		t.Fatalf("len = %d, want 2", len(versions))
	}
}
