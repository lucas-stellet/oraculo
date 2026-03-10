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

// executeCmd builds a fresh root command, sets args, captures stdout, and executes.
// It returns the captured output and any error from Execute.
func executeCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// setupTestDir creates a temp directory and changes into it. The original
// directory is restored and the temp directory removed when the test ends.
func setupTestDir(t *testing.T) string {
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

func TestEpicInit(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "epic", "init", "my-epic", "--description", "test desc")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["name"] != "my-epic" {
		t.Errorf("name = %v, want %q", result["name"], "my-epic")
	}
	if result["path"] != filepath.Join(".oraculo", "epics", "my-epic") {
		t.Errorf("path = %v, want %q", result["path"], filepath.Join(".oraculo", "epics", "my-epic"))
	}
	if result["created"] != true {
		t.Errorf("created = %v, want true", result["created"])
	}
}

func TestEpicInit_Idempotent(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "my-epic", "--description", "first")
	if err != nil {
		t.Fatalf("first init: %v", err)
	}

	out, err := executeCmd(t, "tools", "epic", "init", "my-epic", "--description", "second")
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

func TestEpicSave(t *testing.T) {
	setupTestDir(t)

	// Init the epic first (creates the DB record).
	_, err := executeCmd(t, "tools", "epic", "init", "my-epic")
	if err != nil {
		t.Fatalf("init: %v", err)
	}

	// Save requirements via stdin.
	content := "# Requirements\n\nSome markdown."
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "epic", "save", "my-epic"})
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
	expectedPath := filepath.Join(".oraculo", "epics", "my-epic", "requirements.md")
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

func TestEpicGet(t *testing.T) {
	setupTestDir(t)

	// Init and save.
	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	content := "# Epic Requirements\n\nDetails here."
	dir := filepath.Join(".oraculo", "epics", "my-epic")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "requirements.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCmd(t, "tools", "epic", "get", "my-epic")
	if err != nil {
		t.Fatalf("get: %v\noutput: %s", err, out)
	}

	// get returns raw markdown, NOT JSON.
	if out != content {
		t.Errorf("output = %q, want %q", out, content)
	}
}

func TestEpicGet_NotFound(t *testing.T) {
	setupTestDir(t)

	// Ensure db is available (PersistentPreRunE creates .oraculo/).
	out, err := executeCmd(t, "tools", "epic", "get", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent epic")
	}

	// Should contain error JSON.
	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] == "" {
		t.Error("expected non-empty error code")
	}
}

func TestEpicList(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "alpha", "--description", "first")
	_, _ = executeCmd(t, "tools", "epic", "init", "beta", "--description", "second")

	out, err := executeCmd(t, "tools", "epic", "list")
	if err != nil {
		t.Fatalf("list: %v\noutput: %s", err, out)
	}

	var epics []map[string]any
	if err := json.Unmarshal([]byte(out), &epics); err != nil {
		t.Fatalf("invalid JSON array: %v\noutput: %s", err, out)
	}
	if len(epics) != 2 {
		t.Fatalf("len = %d, want 2", len(epics))
	}
}

func TestEpicList_Empty(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "epic", "list")
	if err != nil {
		t.Fatalf("list: %v\noutput: %s", err, out)
	}

	// Empty list should be null or empty JSON array.
	out = strings.TrimSpace(out)
	if out != "null" && out != "[]" {
		t.Errorf("expected null or [], got: %s", out)
	}
}

func TestEpicUpdate(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic", "--description", "old")

	out, err := executeCmd(t, "tools", "epic", "update", "my-epic", "--description", "new desc")
	if err != nil {
		t.Fatalf("update: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["description"] != "new desc" {
		t.Errorf("description = %v, want %q", result["description"], "new desc")
	}
}

func TestEpicUpdate_NotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "epic", "update", "nonexistent", "--description", "x")
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

func TestEpicDelete(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "doomed")

	out, err := executeCmd(t, "tools", "epic", "delete", "doomed")
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
	if result["deleted"] != true {
		t.Errorf("deleted = %v, want true", result["deleted"])
	}
}

func TestEpicDelete_NotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "epic", "delete", "nonexistent")
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

// ---------- Epic Version ----------

func TestEpicVersion(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "my-epic")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	content := "# Epic Requirements v1\n\nFirst version."
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "epic", "version", "my-epic"})
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
	if result["epic_id"] == nil {
		t.Error("expected EpicID to be set")
	}
}

func TestEpicVersionMultiple(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	// First version
	var buf1 bytes.Buffer
	cmd1 := cli.NewRoot("test")
	cmd1.SetOut(&buf1)
	cmd1.SetErr(&buf1)
	cmd1.SetIn(strings.NewReader("version 1 content"))
	cmd1.SetArgs([]string{"tools", "epic", "version", "my-epic"})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("version 1: %v", err)
	}

	// Second version
	var buf2 bytes.Buffer
	cmd2 := cli.NewRoot("test")
	cmd2.SetOut(&buf2)
	cmd2.SetErr(&buf2)
	cmd2.SetIn(strings.NewReader("version 2 content"))
	cmd2.SetArgs([]string{"tools", "epic", "version", "my-epic"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("version 2: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf2.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf2.String())
	}
	if result["number"] != float64(2) {
		t.Errorf("Number = %v, want 2", result["number"])
	}
}

func TestEpicVersions(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	// Create two versions
	for _, content := range []string{"v1", "v2"} {
		var buf bytes.Buffer
		cmd := cli.NewRoot("test")
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetIn(strings.NewReader(content))
		cmd.SetArgs([]string{"tools", "epic", "version", "my-epic"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("version: %v", err)
		}
	}

	out, err := executeCmd(t, "tools", "epic", "versions", "my-epic")
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

func TestEpicVersion_NotFound(t *testing.T) {
	setupTestDir(t)

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("content"))
	cmd.SetArgs([]string{"tools", "epic", "version", "nonexistent"})
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
