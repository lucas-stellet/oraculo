package tools_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

// createEpicVersion is a test helper that inits an epic, creates a version via stdin,
// and returns the version ID.
func createEpicVersion(t *testing.T, epicName, content string) float64 {
	t.Helper()

	_, err := executeCmd(t, "tools", "epic", "init", epicName)
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "epic", "version", epicName})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("epic version: %v\noutput: %s", err, buf.String())
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	return result["ID"].(float64)
}

func TestReviewCreateApproved(t *testing.T) {
	setupTestDir(t)

	versionID := createEpicVersion(t, "my-epic", "# Requirements v1")

	out, err := executeCmd(t, "tools", "review", "create", fmt.Sprintf("%d", int(versionID)),
		"--type", "epic", "--verdict", "approved", "--comment", "Looks good")
	if err != nil {
		t.Fatalf("review create: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["ID"] == nil || result["ID"].(float64) < 1 {
		t.Errorf("expected valid ID, got %v", result["ID"])
	}
	if result["Verdict"] != "approved" {
		t.Errorf("Verdict = %v, want %q", result["Verdict"], "approved")
	}
	if result["Comment"] != "Looks good" {
		t.Errorf("Comment = %v, want %q", result["Comment"], "Looks good")
	}
}

func TestReviewCreateRejected(t *testing.T) {
	setupTestDir(t)

	versionID := createEpicVersion(t, "my-epic", "# Requirements v1")

	out, err := executeCmd(t, "tools", "review", "create", fmt.Sprintf("%d", int(versionID)),
		"--type", "epic", "--verdict", "rejected", "--comment", "Needs work")
	if err != nil {
		t.Fatalf("review create: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["Verdict"] != "rejected" {
		t.Errorf("Verdict = %v, want %q", result["Verdict"], "rejected")
	}
}

func TestReviewGet(t *testing.T) {
	setupTestDir(t)

	versionID := createEpicVersion(t, "my-epic", "# Requirements v1")

	// Create a review first
	out, err := executeCmd(t, "tools", "review", "create", fmt.Sprintf("%d", int(versionID)),
		"--type", "epic", "--verdict", "approved", "--comment", "LGTM")
	if err != nil {
		t.Fatalf("review create: %v\noutput: %s", err, out)
	}

	var created map[string]any
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	reviewID := int(created["ID"].(float64))

	// Get the review
	out, err = executeCmd(t, "tools", "review", "get", fmt.Sprintf("%d", reviewID))
	if err != nil {
		t.Fatalf("review get: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["ID"] != float64(reviewID) {
		t.Errorf("ID = %v, want %v", result["ID"], reviewID)
	}
	if result["Verdict"] != "approved" {
		t.Errorf("Verdict = %v, want %q", result["Verdict"], "approved")
	}
	if result["Comment"] != "LGTM" {
		t.Errorf("Comment = %v, want %q", result["Comment"], "LGTM")
	}
}

func TestReviewList(t *testing.T) {
	setupTestDir(t)

	versionID := createEpicVersion(t, "my-epic", "# Requirements v1")
	vid := fmt.Sprintf("%d", int(versionID))

	// Create two reviews
	_, err := executeCmd(t, "tools", "review", "create", vid, "--type", "epic", "--verdict", "rejected", "--comment", "First pass")
	if err != nil {
		t.Fatalf("review create 1: %v", err)
	}
	_, err = executeCmd(t, "tools", "review", "create", vid, "--type", "epic", "--verdict", "approved", "--comment", "Second pass")
	if err != nil {
		t.Fatalf("review create 2: %v", err)
	}

	out, err := executeCmd(t, "tools", "review", "list", vid, "--type", "epic")
	if err != nil {
		t.Fatalf("review list: %v\noutput: %s", err, out)
	}

	var reviews []map[string]any
	if err := json.Unmarshal([]byte(out), &reviews); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(reviews) != 2 {
		t.Fatalf("len = %d, want 2", len(reviews))
	}
}

func TestReviewCreateInvalidType(t *testing.T) {
	setupTestDir(t)

	versionID := createEpicVersion(t, "my-epic", "# Requirements v1")

	out, err := executeCmd(t, "tools", "review", "create", fmt.Sprintf("%d", int(versionID)),
		"--type", "invalid", "--verdict", "approved")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] == nil || result["error"] == "" {
		t.Error("expected error in response")
	}
}

func TestReviewCreateInvalidVerdict(t *testing.T) {
	setupTestDir(t)

	versionID := createEpicVersion(t, "my-epic", "# Requirements v1")

	out, err := executeCmd(t, "tools", "review", "create", fmt.Sprintf("%d", int(versionID)),
		"--type", "epic", "--verdict", "maybe")
	if err == nil {
		t.Fatal("expected error for invalid verdict")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] == nil || result["error"] == "" {
		t.Error("expected error in response")
	}
}
