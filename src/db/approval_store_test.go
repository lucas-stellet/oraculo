package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func TestApprovalStore_Request(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	approval, err := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "some content")
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if approval.ID == "" {
		t.Error("expected non-empty ID")
	}
	if approval.Type != domain.ApprovalEpicRequirements {
		t.Errorf("Type = %q, want %q", approval.Type, domain.ApprovalEpicRequirements)
	}
	if approval.EpicID == nil || *approval.EpicID != epic.ID {
		t.Errorf("EpicID = %v, want %d", approval.EpicID, epic.ID)
	}
	if approval.StoryID != nil {
		t.Errorf("StoryID = %v, want nil", approval.StoryID)
	}
	if approval.Content != "some content" {
		t.Errorf("Content = %q, want %q", approval.Content, "some content")
	}
	if approval.Status != domain.ApprovalPending {
		t.Errorf("Status = %q, want %q", approval.Status, domain.ApprovalPending)
	}
	if approval.RequestedAt.IsZero() {
		t.Error("expected non-zero RequestedAt")
	}
	if approval.DecidedAt != nil {
		t.Errorf("DecidedAt = %v, want nil", approval.DecidedAt)
	}
}

func TestApprovalStore_GetByID(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	storyStore := NewStoryStore(database)
	story, _, _ := storyStore.Create(epic.ID, "test-story", "desc")

	created, _ := store.Request(domain.ApprovalStoryDefinition, &epic.ID, &story.ID, "story content")

	approval, err := store.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if approval.ID != created.ID {
		t.Errorf("ID = %q, want %q", approval.ID, created.ID)
	}
	if approval.Type != domain.ApprovalStoryDefinition {
		t.Errorf("Type = %q, want %q", approval.Type, domain.ApprovalStoryDefinition)
	}
	if approval.EpicID == nil || *approval.EpicID != epic.ID {
		t.Errorf("EpicID = %v, want %d", approval.EpicID, epic.ID)
	}
	if approval.StoryID == nil || *approval.StoryID != story.ID {
		t.Errorf("StoryID = %v, want %d", approval.StoryID, story.ID)
	}
	if approval.Content != "story content" {
		t.Errorf("Content = %q, want %q", approval.Content, "story content")
	}
}

func TestApprovalStore_GetByID_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	_, err := store.GetByID("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent approval")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestApprovalStore_List(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "content 1")
	store.Request(domain.ApprovalExecutionPlan, &epic.ID, nil, "content 2")

	approvals, err := store.List(false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(approvals) != 2 {
		t.Fatalf("len = %d, want 2", len(approvals))
	}
}

func TestApprovalStore_List_PendingOnly(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	a1, _ := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "content 1")
	store.Request(domain.ApprovalExecutionPlan, &epic.ID, nil, "content 2")

	// Approve the first one so only the second remains pending
	store.Verdict(a1.ID, domain.VerdictApproved, "looks good")

	approvals, err := store.List(true)
	if err != nil {
		t.Fatalf("List(pendingOnly=true): %v", err)
	}
	if len(approvals) != 1 {
		t.Fatalf("len = %d, want 1", len(approvals))
	}
	if approvals[0].Status != domain.ApprovalPending {
		t.Errorf("Status = %q, want %q", approvals[0].Status, domain.ApprovalPending)
	}
}

func TestApprovalStore_Verdict_Approved(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	created, _ := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "content")

	approval, err := store.Verdict(created.ID, domain.VerdictApproved, "lgtm")
	if err != nil {
		t.Fatalf("Verdict: %v", err)
	}
	if approval.Status != domain.ApprovalApproved {
		t.Errorf("Status = %q, want %q", approval.Status, domain.ApprovalApproved)
	}
	if approval.VerdictComment != "lgtm" {
		t.Errorf("VerdictComment = %q, want %q", approval.VerdictComment, "lgtm")
	}
	if approval.DecidedAt == nil {
		t.Error("expected non-nil DecidedAt")
	}
	if approval.PreviousVersion != "" {
		t.Errorf("PreviousVersion = %q, want empty", approval.PreviousVersion)
	}
}

func TestApprovalStore_Verdict_Rejected(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	created, _ := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "content")

	approval, err := store.Verdict(created.ID, domain.VerdictRejected, "not acceptable")
	if err != nil {
		t.Fatalf("Verdict: %v", err)
	}
	if approval.Status != domain.ApprovalRejected {
		t.Errorf("Status = %q, want %q", approval.Status, domain.ApprovalRejected)
	}
	if approval.VerdictComment != "not acceptable" {
		t.Errorf("VerdictComment = %q, want %q", approval.VerdictComment, "not acceptable")
	}
}

func TestApprovalStore_Verdict_NeedsRevision(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	created, _ := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "original content")

	approval, err := store.Verdict(created.ID, domain.VerdictNeedsRevision, "needs work")
	if err != nil {
		t.Fatalf("Verdict: %v", err)
	}
	if approval.Status != domain.ApprovalNeedsRevision {
		t.Errorf("Status = %q, want %q", approval.Status, domain.ApprovalNeedsRevision)
	}
	if approval.PreviousVersion != "original content" {
		t.Errorf("PreviousVersion = %q, want %q", approval.PreviousVersion, "original content")
	}
}

func TestApprovalStore_Verdict_NonPending_Error(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	created, _ := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "content")
	store.Verdict(created.ID, domain.VerdictApproved, "done")

	// Attempt to verdict again on non-pending approval
	_, err := store.Verdict(created.ID, domain.VerdictRejected, "too late")
	if err == nil {
		t.Fatal("expected error for non-pending verdict")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestApprovalStore_Verdict_PropagatesEpicApprovalStatus(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)
	epicStore := NewEpicStore(database)

	epic, _, _ := epicStore.Create("test-epic", "desc")
	created, _ := store.Request(domain.ApprovalEpicRequirements, &epic.ID, nil, "content")

	store.Verdict(created.ID, domain.VerdictApproved, "lgtm")

	updated, err := epicStore.GetByName("test-epic")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if updated.ApprovalStatus != domain.ApprovalApproved {
		t.Errorf("epic ApprovalStatus = %q, want %q", updated.ApprovalStatus, domain.ApprovalApproved)
	}
}

func TestApprovalStore_Verdict_PropagatesStoryApprovalStatus(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)
	epicStore := NewEpicStore(database)
	storyStore := NewStoryStore(database)

	epic, _, _ := epicStore.Create("test-epic", "desc")
	story, _, _ := storyStore.Create(epic.ID, "test-story", "desc")
	created, _ := store.Request(domain.ApprovalStoryDefinition, &epic.ID, &story.ID, "content")

	store.Verdict(created.ID, domain.VerdictApproved, "lgtm")

	updated, err := storyStore.GetByName(epic.ID, "test-story")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if updated.ApprovalStatus != domain.ApprovalApproved {
		t.Errorf("story ApprovalStatus = %q, want %q", updated.ApprovalStatus, domain.ApprovalApproved)
	}
}

func TestApprovalStore_Verdict_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	_, err := store.Verdict("nonexistent-id", domain.VerdictApproved, "nope")
	if err == nil {
		t.Fatal("expected error for nonexistent approval")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
