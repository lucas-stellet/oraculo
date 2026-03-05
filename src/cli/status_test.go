package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/cli"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
)

// statusCmd builds a fresh root command, sets args to "status", captures stdout, and executes.
func statusCmd(t *testing.T) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"status"})
	err := cmd.Execute()
	return buf.String(), err
}

func TestStatus_NoProject(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	out, err := statusCmd(t)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, "No Oraculo project found") {
		t.Errorf("expected 'No Oraculo project found' message, got: %s", out)
	}
}

func TestStatus_EmptyProject(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Create the DB so the project exists.
	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	database.Close()

	out, err := statusCmd(t)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, "=== Oraculo Status ===") {
		t.Error("missing header")
	}
	if !strings.Contains(out, "Epics") {
		t.Error("missing Epics section")
	}
	if !strings.Contains(out, "(none)") {
		t.Error("expected (none) for empty epics")
	}
	if !strings.Contains(out, "Tasks") {
		t.Error("missing Tasks section")
	}
	if !strings.Contains(out, "Pending Approvals: 0") {
		t.Errorf("expected 'Pending Approvals: 0', got: %s", out)
	}
}

func TestStatus_WithData(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	// Set up the database with test data.
	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Create epics.
	epicStore := db.NewEpicStore(database)
	epic1, _, err := epicStore.Create("gastos-control", "expense tracking")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	if err := epicStore.UpdateApprovalStatus("gastos-control", domain.ApprovalApproved); err != nil {
		t.Fatalf("update approval: %v", err)
	}
	_, _, err = epicStore.Create("user-auth", "auth module")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	if err := epicStore.UpdateApprovalStatus("user-auth", domain.ApprovalPending); err != nil {
		t.Fatalf("update approval: %v", err)
	}

	// Create a story under the first epic.
	storyStore := db.NewStoryStore(database)
	story, _, err := storyStore.Create(epic1.ID, "setup-db", "database setup")
	if err != nil {
		t.Fatalf("create story: %v", err)
	}

	// Create tasks under the story.
	taskStore := db.NewTaskStore(database)
	_, _, err = taskStore.Create(story.ID, "task-a", "do A", nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	_, _, err = taskStore.Create(story.ID, "task-b", "do B", nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	_, _, err = taskStore.Create(story.ID, "task-c", "do C", nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	// Start task-b (pending -> in_progress).
	_, err = taskStore.Start(story.ID, "task-b")
	if err != nil {
		t.Fatalf("start task: %v", err)
	}
	// Complete task-b (in_progress -> completed).
	_, err = taskStore.Complete(story.ID, "task-b", domain.TaskResult{Summary: "done"})
	if err != nil {
		t.Fatalf("complete task: %v", err)
	}

	// Create pending approvals.
	approvalStore := db.NewApprovalStore(database)
	epicID := epic1.ID
	_, err = approvalStore.Request(domain.ApprovalQAEscalation, &epicID, nil, "review reqs")
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}
	_, err = approvalStore.Request(domain.ApprovalDesign, &epicID, nil, "review story")
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	database.Close()

	// Run the status command.
	out, err := statusCmd(t)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	// Verify header.
	if !strings.Contains(out, "=== Oraculo Status ===") {
		t.Error("missing header")
	}

	// Verify epics.
	if !strings.Contains(out, "gastos-control") {
		t.Error("missing epic gastos-control")
	}
	if !strings.Contains(out, "user-auth") {
		t.Error("missing epic user-auth")
	}
	if !strings.Contains(out, "approved") {
		t.Error("missing approved status")
	}

	// Verify task counts.
	if !strings.Contains(out, "Tasks") {
		t.Error("missing Tasks section")
	}
	// 2 pending (task-a, task-c), 0 in_progress, 1 completed (task-b), 0 failed.
	if !strings.Contains(out, "pending") {
		t.Error("missing pending status in task table")
	}
	if !strings.Contains(out, "completed") {
		t.Error("missing completed status in task table")
	}

	// Verify pending approvals count.
	if !strings.Contains(out, "Pending Approvals: 2") {
		t.Errorf("expected 'Pending Approvals: 2', got: %s", out)
	}
}
