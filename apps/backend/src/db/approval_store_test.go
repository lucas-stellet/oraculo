package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

func TestApprovalStore_Request(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	approval, err := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "some content")
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if approval.ID == "" {
		t.Error("expected non-empty ID")
	}
	if approval.Type != domain.ApprovalQAEscalation {
		t.Errorf("Type = %q, want %q", approval.Type, domain.ApprovalQAEscalation)
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

	created, _ := store.Request(domain.ApprovalDesign, &epic.ID, &story.ID, "story content")

	approval, err := store.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if approval.ID != created.ID {
		t.Errorf("ID = %q, want %q", approval.ID, created.ID)
	}
	if approval.Type != domain.ApprovalDesign {
		t.Errorf("Type = %q, want %q", approval.Type, domain.ApprovalDesign)
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

	store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content 1")
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

	a1, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content 1")
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

func TestApprovalStore_ListByEpic(t *testing.T) {
	database := testDB(t)
	epicStore := NewEpicStore(database)
	approvalStore := NewApprovalStore(database)

	epic1, _, err := epicStore.Create("epic-1", "")
	if err != nil {
		t.Fatalf("create epic-1: %v", err)
	}
	epic2, _, err := epicStore.Create("epic-2", "")
	if err != nil {
		t.Fatalf("create epic-2: %v", err)
	}

	if _, err := approvalStore.Request(domain.ApprovalDesign, &epic1.ID, nil, "content-1"); err != nil {
		t.Fatalf("request approval 1: %v", err)
	}
	if _, err := approvalStore.Request(domain.ApprovalDesign, &epic2.ID, nil, "content-2"); err != nil {
		t.Fatalf("request approval 2: %v", err)
	}

	result, err := approvalStore.ListByEpic(epic1.ID, false)
	if err != nil {
		t.Fatalf("ListByEpic: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 approval, got %d", len(result))
	}
}

func TestApprovalStore_Verdict_Approved(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")

	created, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content")

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

	created, _ := store.Request(domain.ApprovalDesign, &epic.ID, nil, "content")

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

	created, _ := store.Request(domain.ApprovalExecutionPlan, &epic.ID, nil, "original content")

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

	created, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content")
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

func TestApprovalStore_CreateComment(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	approval, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content")

	comment, err := store.CreateComment(approval.ID, "some selected text", "this needs clarification")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if comment.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if comment.SelectedText != "some selected text" {
		t.Errorf("SelectedText = %q, want %q", comment.SelectedText, "some selected text")
	}
	if comment.Comment != "this needs clarification" {
		t.Errorf("Comment = %q, want %q", comment.Comment, "this needs clarification")
	}
	if comment.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestApprovalStore_ListComments(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	a1, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content 1")
	a2, _ := store.Request(domain.ApprovalExecutionPlan, &epic.ID, nil, "content 2")

	store.CreateComment(a1.ID, "text A", "comment A")
	store.CreateComment(a1.ID, "text B", "comment B")
	store.CreateComment(a2.ID, "text C", "comment C") // different approval — must not appear

	comments, err := store.ListComments(a1.ID)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("len = %d, want 2", len(comments))
	}
	if comments[0].SelectedText != "text A" {
		t.Errorf("comments[0].SelectedText = %q, want %q", comments[0].SelectedText, "text A")
	}
	if comments[1].SelectedText != "text B" {
		t.Errorf("comments[1].SelectedText = %q, want %q", comments[1].SelectedText, "text B")
	}
}

func TestApprovalStore_ListComments_Empty(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	approval, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content")

	comments, err := store.ListComments(approval.ID)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("len = %d, want 0", len(comments))
	}
}

func TestApprovalStore_DeleteComment(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	approval, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content")

	c1, _ := store.CreateComment(approval.ID, "text A", "comment A")
	store.CreateComment(approval.ID, "text B", "comment B")

	if err := store.DeleteComment(c1.ID); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}

	comments, err := store.ListComments(approval.ID)
	if err != nil {
		t.Fatalf("ListComments after delete: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("len = %d, want 1", len(comments))
	}
	if comments[0].SelectedText != "text B" {
		t.Errorf("remaining comment SelectedText = %q, want %q", comments[0].SelectedText, "text B")
	}
}

func TestApprovalStore_DeleteCommentsByApproval(t *testing.T) {
	database := testDB(t)
	store := NewApprovalStore(database)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	a1, _ := store.Request(domain.ApprovalQAEscalation, &epic.ID, nil, "content 1")
	a2, _ := store.Request(domain.ApprovalExecutionPlan, &epic.ID, nil, "content 2")

	store.CreateComment(a1.ID, "text A", "comment A")
	store.CreateComment(a1.ID, "text B", "comment B")
	store.CreateComment(a2.ID, "text C", "comment C")

	if err := store.DeleteCommentsByApproval(a1.ID); err != nil {
		t.Fatalf("DeleteCommentsByApproval: %v", err)
	}

	// a1 comments must be gone
	comments, err := store.ListComments(a1.ID)
	if err != nil {
		t.Fatalf("ListComments a1 after delete: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments for a1, got %d", len(comments))
	}

	// a2 comments must be intact
	comments, err = store.ListComments(a2.ID)
	if err != nil {
		t.Fatalf("ListComments a2 after delete: %v", err)
	}
	if len(comments) != 1 {
		t.Errorf("expected 1 comment for a2, got %d", len(comments))
	}
}
