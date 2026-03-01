package tools_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

// TestApprovalFullWorkflow tests request -> status (pending) -> verdict (approved) -> status (approved).
func TestApprovalFullWorkflow(t *testing.T) {
	setupTestDir(t)

	// Create an epic first.
	_, err := executeCmd(t, "tools", "epic", "init", "approval-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	// 1. Request an approval via stdin.
	content := "# Requirements\n\nContent here"
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(content))
	cmd.SetArgs([]string{"tools", "approval", "request", "--type", "epic-requirements", "--epic", "approval-epic"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("request: %v\noutput: %s", err, buf.String())
	}

	var requested map[string]any
	if err := json.Unmarshal(buf.Bytes(), &requested); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	approvalID, ok := requested["ID"].(string)
	if !ok || approvalID == "" {
		t.Fatalf("expected non-empty approval ID, got: %v", requested["ID"])
	}
	if requested["Status"] != "pending" {
		t.Errorf("Status = %v, want %q", requested["Status"], "pending")
	}
	if requested["Type"] != "epic-requirements" {
		t.Errorf("Type = %v, want %q", requested["Type"], "epic-requirements")
	}
	if requested["Content"] != content {
		t.Errorf("Content = %v, want %q", requested["Content"], content)
	}

	// 2. Check status — should be pending.
	statusOut, err := executeCmd(t, "tools", "approval", "status", approvalID)
	if err != nil {
		t.Fatalf("status: %v\noutput: %s", err, statusOut)
	}
	var statusResult map[string]any
	if err := json.Unmarshal([]byte(statusOut), &statusResult); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, statusOut)
	}
	if statusResult["Status"] != "pending" {
		t.Errorf("Status = %v, want %q", statusResult["Status"], "pending")
	}

	// 3. Verdict — approve.
	verdictOut, err := executeCmd(t, "tools", "approval", "verdict", approvalID, "--verdict", "approved", "--comment", "looks good")
	if err != nil {
		t.Fatalf("verdict: %v\noutput: %s", err, verdictOut)
	}
	var verdictResult map[string]any
	if err := json.Unmarshal([]byte(verdictOut), &verdictResult); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, verdictOut)
	}
	if verdictResult["Status"] != "approved" {
		t.Errorf("Status = %v, want %q", verdictResult["Status"], "approved")
	}
	if verdictResult["VerdictComment"] != "looks good" {
		t.Errorf("VerdictComment = %v, want %q", verdictResult["VerdictComment"], "looks good")
	}

	// 4. Check status again — should be approved.
	statusOut2, err := executeCmd(t, "tools", "approval", "status", approvalID)
	if err != nil {
		t.Fatalf("status after verdict: %v\noutput: %s", err, statusOut2)
	}
	var statusResult2 map[string]any
	if err := json.Unmarshal([]byte(statusOut2), &statusResult2); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, statusOut2)
	}
	if statusResult2["Status"] != "approved" {
		t.Errorf("Status = %v, want %q", statusResult2["Status"], "approved")
	}
}

// TestApprovalRequestAllTypes tests that all valid approval types are accepted.
func TestApprovalRequestAllTypes(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "types-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	types := []string{"epic-requirements", "story-definition", "qa-escalation", "execution-plan"}
	for _, at := range types {
		t.Run(at, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := cli.NewRoot("test")
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetIn(strings.NewReader("# " + at + "\n\nContent"))
			cmd.SetArgs([]string{"tools", "approval", "request", "--type", at, "--epic", "types-epic"})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("request %s: %v\noutput: %s", at, err, buf.String())
			}

			var result map[string]any
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
			}
			if result["Type"] != at {
				t.Errorf("Type = %v, want %q", result["Type"], at)
			}
		})
	}
}

// TestApprovalRequestInvalidType tests that an invalid type is rejected.
func TestApprovalRequestInvalidType(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "inv-type-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("content"))
	cmd.SetArgs([]string{"tools", "approval", "request", "--type", "invalid-type", "--epic", "inv-type-epic"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid type")
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["error"] != "missing_required" {
		t.Errorf("error = %q, want %q", result["error"], "missing_required")
	}
}

// TestApprovalRequestWithStory tests that --story resolves to a story ID.
func TestApprovalRequestWithStory(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "story-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}
	_, err = executeCmd(t, "tools", "story", "init", "my-story", "--epic", "story-epic")
	if err != nil {
		t.Fatalf("story init: %v", err)
	}

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("# Story Requirements"))
	cmd.SetArgs([]string{"tools", "approval", "request", "--type", "story-definition", "--epic", "story-epic", "--story", "my-story"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("request: %v\noutput: %s", err, buf.String())
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	// StoryID should be non-nil (a float64 in JSON).
	if result["StoryID"] == nil {
		t.Error("expected non-nil StoryID")
	}
	if result["EpicID"] == nil {
		t.Error("expected non-nil EpicID")
	}
}

// TestApprovalVerdictNeedsRevision tests that needs_revision stores previous version.
func TestApprovalVerdictNeedsRevision(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "rev-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	originalContent := "# Original Content\n\nFirst version"
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(originalContent))
	cmd.SetArgs([]string{"tools", "approval", "request", "--type", "epic-requirements", "--epic", "rev-epic"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("request: %v\noutput: %s", err, buf.String())
	}

	var requested map[string]any
	if err := json.Unmarshal(buf.Bytes(), &requested); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	approvalID := requested["ID"].(string)

	// Verdict with needs_revision.
	verdictOut, err := executeCmd(t, "tools", "approval", "verdict", approvalID, "--verdict", "needs_revision", "--comment", "needs more detail")
	if err != nil {
		t.Fatalf("verdict: %v\noutput: %s", err, verdictOut)
	}

	var verdictResult map[string]any
	if err := json.Unmarshal([]byte(verdictOut), &verdictResult); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, verdictOut)
	}
	if verdictResult["Status"] != "needs_revision" {
		t.Errorf("Status = %v, want %q", verdictResult["Status"], "needs_revision")
	}
	if verdictResult["PreviousVersion"] != originalContent {
		t.Errorf("PreviousVersion = %v, want %q", verdictResult["PreviousVersion"], originalContent)
	}
}

// TestApprovalListPending tests that --pending filters correctly.
func TestApprovalListPending(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "list-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	// Create two approvals.
	var ids []string
	for i := 0; i < 2; i++ {
		var buf bytes.Buffer
		cmd := cli.NewRoot("test")
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetIn(strings.NewReader("content"))
		cmd.SetArgs([]string{"tools", "approval", "request", "--type", "epic-requirements", "--epic", "list-epic"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		var result map[string]any
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		ids = append(ids, result["ID"].(string))
	}

	// Approve the first one.
	_, err = executeCmd(t, "tools", "approval", "verdict", ids[0], "--verdict", "approved")
	if err != nil {
		t.Fatalf("verdict: %v", err)
	}

	// List all — should have 2.
	allOut, err := executeCmd(t, "tools", "approval", "list")
	if err != nil {
		t.Fatalf("list all: %v\noutput: %s", err, allOut)
	}
	var allApprovals []map[string]any
	if err := json.Unmarshal([]byte(allOut), &allApprovals); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, allOut)
	}
	if len(allApprovals) != 2 {
		t.Errorf("list all: len = %d, want 2", len(allApprovals))
	}

	// List pending — should have 1.
	pendingOut, err := executeCmd(t, "tools", "approval", "list", "--pending")
	if err != nil {
		t.Fatalf("list pending: %v\noutput: %s", err, pendingOut)
	}
	var pendingApprovals []map[string]any
	if err := json.Unmarshal([]byte(pendingOut), &pendingApprovals); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, pendingOut)
	}
	if len(pendingApprovals) != 1 {
		t.Errorf("list pending: len = %d, want 1", len(pendingApprovals))
	}
	if pendingApprovals[0]["ID"] != ids[1] {
		t.Errorf("pending ID = %v, want %q", pendingApprovals[0]["ID"], ids[1])
	}
}

// TestApprovalVerdictOnNonPending tests that a verdict on a non-pending approval returns invalid_transition.
func TestApprovalVerdictOnNonPending(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "epic", "init", "nonpend-epic", "--description", "test")
	if err != nil {
		t.Fatalf("epic init: %v", err)
	}

	// Create and approve.
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader("content"))
	cmd.SetArgs([]string{"tools", "approval", "request", "--type", "epic-requirements", "--epic", "nonpend-epic"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("request: %v", err)
	}
	var requested map[string]any
	if err := json.Unmarshal(buf.Bytes(), &requested); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	approvalID := requested["ID"].(string)

	_, err = executeCmd(t, "tools", "approval", "verdict", approvalID, "--verdict", "approved")
	if err != nil {
		t.Fatalf("first verdict: %v", err)
	}

	// Try to verdict again — should fail.
	out, err := executeCmd(t, "tools", "approval", "verdict", approvalID, "--verdict", "rejected")
	if err == nil {
		t.Fatal("expected error for verdict on non-pending approval")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] != "invalid_transition" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_transition")
	}
}

// TestApprovalStatusNotFound tests that status on a nonexistent ID returns not_found.
func TestApprovalStatusNotFound(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "approval", "status", "nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent approval")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] != "not_found" {
		t.Errorf("error = %q, want %q", result["error"], "not_found")
	}
}

// TestApprovalVerdictInvalidVerdict tests that an invalid verdict value is rejected.
func TestApprovalVerdictInvalidVerdict(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "approval", "verdict", "some-id", "--verdict", "maybe")
	if err == nil {
		t.Fatal("expected error for invalid verdict")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] != "missing_required" {
		t.Errorf("error = %q, want %q", result["error"], "missing_required")
	}
}
