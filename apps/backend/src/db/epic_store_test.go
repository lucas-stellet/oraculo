package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

func TestEpicStore_Create(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	epic, created, err := store.Create("my-epic", "a description")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if epic.Name != "my-epic" {
		t.Errorf("Name = %q, want %q", epic.Name, "my-epic")
	}
	if epic.Description != "a description" {
		t.Errorf("Description = %q, want %q", epic.Description, "a description")
	}
}

func TestEpicStore_CreateIdempotent(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	e1, _, _ := store.Create("my-epic", "desc")
	e2, created, err := store.Create("my-epic", "desc")
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if e2.ID != e1.ID {
		t.Errorf("IDs differ: %d vs %d", e2.ID, e1.ID)
	}
}

func TestEpicStore_GetByName(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("test-epic", "desc")
	epic, err := store.GetByName("test-epic")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if epic.Name != "test-epic" {
		t.Errorf("Name = %q, want %q", epic.Name, "test-epic")
	}
}

func TestEpicStore_GetByName_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	_, err := store.GetByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent epic")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_List(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("alpha", "")
	store.Create("beta", "")

	epics, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(epics) != 2 {
		t.Fatalf("len = %d, want 2", len(epics))
	}
}

func TestEpicStore_Update(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("my-epic", "old")
	epic, err := store.Update("my-epic", "new description")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if epic.Description != "new description" {
		t.Errorf("Description = %q, want %q", epic.Description, "new description")
	}
}

func TestEpicStore_Update_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	_, err := store.Update("nonexistent", "desc")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_Delete(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("doomed", "")
	if err := store.Delete("doomed"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.GetByName("doomed")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestEpicStore_Delete_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	err := store.Delete("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_UpdateApprovalStatus(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("my-epic", "")
	if err := store.UpdateApprovalStatus("my-epic", domain.ApprovalPending); err != nil {
		t.Fatalf("UpdateApprovalStatus: %v", err)
	}
	epic, _ := store.GetByName("my-epic")
	if epic.ApprovalStatus != domain.ApprovalPending {
		t.Errorf("ApprovalStatus = %q, want %q", epic.ApprovalStatus, domain.ApprovalPending)
	}
}

func TestEpicStore_ListSummaries_Empty(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	summaries, err := store.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestEpicStore_ListSummaries_NoSessions(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("my-epic", "desc")

	summaries, err := store.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.Name != "my-epic" {
		t.Errorf("Name = %q, want %q", s.Name, "my-epic")
	}
	if s.Phase != "discover" {
		t.Errorf("Phase = %q, want %q", s.Phase, "discover")
	}
	if s.PhaseStatus != "pending" {
		t.Errorf("PhaseStatus = %q, want %q", s.PhaseStatus, "pending")
	}
	if s.StoryCount != 0 {
		t.Errorf("StoryCount = %d, want 0", s.StoryCount)
	}
}

func TestEpicStore_ListSummaries_WithActiveSession(t *testing.T) {
	database := testDB(t)
	epicStore := NewEpicStore(database)
	sessionStore := NewSessionStore(database)

	epic, _, _ := epicStore.Create("my-epic", "desc")
	sessionStore.Create(domain.SessionEpic, &epic.ID)

	summaries, err := epicStore.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	s := summaries[0]
	if s.Phase != "discover" {
		t.Errorf("Phase = %q, want %q", s.Phase, "discover")
	}
	if s.PhaseStatus != "active" {
		t.Errorf("PhaseStatus = %q, want %q", s.PhaseStatus, "active")
	}
}

func TestEpicStore_ListSummaries_WithCompletedSession(t *testing.T) {
	database := testDB(t)
	epicStore := NewEpicStore(database)
	sessionStore := NewSessionStore(database)

	epic, _, _ := epicStore.Create("my-epic", "desc")
	sess, _, _ := sessionStore.Create(domain.SessionPlan, &epic.ID)
	sessionStore.Close(sess.ID, domain.SessionCompleted)

	summaries, err := epicStore.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	s := summaries[0]
	if s.Phase != "plan" {
		t.Errorf("Phase = %q, want %q", s.Phase, "plan")
	}
	if s.PhaseStatus != "completed" {
		t.Errorf("PhaseStatus = %q, want %q", s.PhaseStatus, "completed")
	}
}

func TestEpicStore_ListSummaries_WithStoriesAndTasks(t *testing.T) {
	database := testDB(t)
	epicStore := NewEpicStore(database)
	storyStore := NewStoryStore(database)
	taskStore := NewTaskStore(database)

	epic, _, _ := epicStore.Create("my-epic", "desc")
	story, _, _ := storyStore.Create(epic.ID, "story-1", "desc")
	storyStore.Create(epic.ID, "story-2", "desc")

	taskStore.Create(story.ID, "task-1", "desc", nil)
	taskStore.Create(story.ID, "task-2", "desc", nil)
	taskStore.Start(story.ID, "task-2")
	taskStore.Complete(story.ID, "task-2", domain.TaskResult{Summary: "done"})

	summaries, err := epicStore.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	s := summaries[0]
	if s.StoryCount != 2 {
		t.Errorf("StoryCount = %d, want 2", s.StoryCount)
	}
	if s.TaskCount != 2 {
		t.Errorf("TaskCount = %d, want 2", s.TaskCount)
	}
	if s.CompletedTaskCount != 1 {
		t.Errorf("CompletedTaskCount = %d, want 1", s.CompletedTaskCount)
	}
}

func TestEpicStore_ListSummaries_ActiveSessionTakesPriority(t *testing.T) {
	database := testDB(t)
	epicStore := NewEpicStore(database)
	sessionStore := NewSessionStore(database)

	epic, _, _ := epicStore.Create("my-epic", "desc")

	// Completed plan session
	sess1, _, _ := sessionStore.Create(domain.SessionPlan, &epic.ID)
	sessionStore.Close(sess1.ID, domain.SessionCompleted)

	// Active execute session
	sessionStore.Create(domain.SessionExecute, &epic.ID)

	summaries, err := epicStore.ListSummaries()
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	s := summaries[0]
	if s.Phase != "execute" {
		t.Errorf("Phase = %q, want %q (active session should take priority)", s.Phase, "execute")
	}
	if s.PhaseStatus != "active" {
		t.Errorf("PhaseStatus = %q, want %q", s.PhaseStatus, "active")
	}
}
