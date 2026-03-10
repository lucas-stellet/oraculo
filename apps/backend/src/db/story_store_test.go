package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

func TestStoryStore_ListSummaries(t *testing.T) {
	database := testDB(t)
	epicStore := NewEpicStore(database)
	storyStore := NewStoryStore(database)
	taskStore := NewTaskStore(database)

	epic, _, err := epicStore.Create("test-epic", "desc")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}

	story, _, err := storyStore.Create(epic.ID, "story-1", "desc")
	if err != nil {
		t.Fatalf("create story: %v", err)
	}

	// Create 3 tasks: 2 completed, 1 failed
	for _, name := range []string{"task-1", "task-2", "task-3"} {
		if _, _, err := taskStore.Create(story.ID, name, "d", nil); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	if _, err := taskStore.Start(story.ID, "task-1"); err != nil {
		t.Fatalf("start task-1: %v", err)
	}
	if _, err := taskStore.Complete(story.ID, "task-1", domain.TaskResult{Summary: "done"}); err != nil {
		t.Fatalf("complete task-1: %v", err)
	}
	if _, err := taskStore.Start(story.ID, "task-2"); err != nil {
		t.Fatalf("start task-2: %v", err)
	}
	if _, err := taskStore.Complete(story.ID, "task-2", domain.TaskResult{Summary: "done"}); err != nil {
		t.Fatalf("complete task-2: %v", err)
	}
	if _, err := taskStore.Start(story.ID, "task-3"); err != nil {
		t.Fatalf("start task-3: %v", err)
	}
	if _, err := taskStore.Fail(story.ID, "task-3", "oops"); err != nil {
		t.Fatalf("fail task-3: %v", err)
	}

	summaries, err := storyStore.ListSummaries(epic.ID)
	if err != nil {
		t.Fatalf("ListSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.TaskCount != 3 {
		t.Errorf("TaskCount = %d, want 3", s.TaskCount)
	}
	if s.CompletedTaskCount != 2 {
		t.Errorf("CompletedTaskCount = %d, want 2", s.CompletedTaskCount)
	}
	if s.FailedTaskCount != 1 {
		t.Errorf("FailedTaskCount = %d, want 1", s.FailedTaskCount)
	}
}

// createEpic is a test helper that creates an epic and returns its ID.
func createEpic(t *testing.T, database *DB, name string) int {
	t.Helper()
	epic, _, err := NewEpicStore(database).Create(name, "test epic")
	if err != nil {
		t.Fatalf("createEpic(%q): %v", name, err)
	}
	return epic.ID
}

func TestStoryStore_Create(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	story, created, err := store.Create(epicID, "my-story", "a description")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if story.Name != "my-story" {
		t.Errorf("Name = %q, want %q", story.Name, "my-story")
	}
	if story.Description != "a description" {
		t.Errorf("Description = %q, want %q", story.Description, "a description")
	}
	if story.EpicID != epicID {
		t.Errorf("EpicID = %d, want %d", story.EpicID, epicID)
	}
}

func TestStoryStore_CreateIdempotent(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	s1, _, _ := store.Create(epicID, "my-story", "desc")
	s2, created, err := store.Create(epicID, "my-story", "desc")
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if s2.ID != s1.ID {
		t.Errorf("IDs differ: %d vs %d", s2.ID, s1.ID)
	}
}

func TestStoryStore_CreateSameNameDifferentEpic(t *testing.T) {
	database := testDB(t)
	epicID1 := createEpic(t, database, "epic-one")
	epicID2 := createEpic(t, database, "epic-two")
	store := NewStoryStore(database)

	s1, created1, err := store.Create(epicID1, "shared-name", "desc1")
	if err != nil {
		t.Fatalf("Create under epic1: %v", err)
	}
	if !created1 {
		t.Error("expected created=true for epic1")
	}

	s2, created2, err := store.Create(epicID2, "shared-name", "desc2")
	if err != nil {
		t.Fatalf("Create under epic2: %v", err)
	}
	if !created2 {
		t.Error("expected created=true for epic2")
	}

	if s1.ID == s2.ID {
		t.Errorf("expected different IDs, both got %d", s1.ID)
	}
}

func TestStoryStore_GetByName(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	store.Create(epicID, "test-story", "desc")
	story, err := store.GetByName(epicID, "test-story")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if story.Name != "test-story" {
		t.Errorf("Name = %q, want %q", story.Name, "test-story")
	}
	if story.EpicID != epicID {
		t.Errorf("EpicID = %d, want %d", story.EpicID, epicID)
	}
}

func TestStoryStore_GetByName_WrongEpic(t *testing.T) {
	database := testDB(t)
	epicID1 := createEpic(t, database, "epic-one")
	epicID2 := createEpic(t, database, "epic-two")
	store := NewStoryStore(database)

	store.Create(epicID1, "only-in-epic1", "desc")
	_, err := store.GetByName(epicID2, "only-in-epic1")
	if err == nil {
		t.Fatal("expected error when querying wrong epic")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStoryStore_GetByName_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	_, err := store.GetByName(epicID, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent story")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStoryStore_List(t *testing.T) {
	database := testDB(t)
	epicID1 := createEpic(t, database, "epic-one")
	epicID2 := createEpic(t, database, "epic-two")
	store := NewStoryStore(database)

	store.Create(epicID1, "alpha", "")
	store.Create(epicID1, "beta", "")
	store.Create(epicID2, "gamma", "")

	stories, err := store.List(epicID1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(stories) != 2 {
		t.Fatalf("len = %d, want 2", len(stories))
	}
	if stories[0].Name != "alpha" {
		t.Errorf("stories[0].Name = %q, want %q", stories[0].Name, "alpha")
	}
	if stories[1].Name != "beta" {
		t.Errorf("stories[1].Name = %q, want %q", stories[1].Name, "beta")
	}
}

func TestStoryStore_Update(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	store.Create(epicID, "my-story", "old")
	story, err := store.Update(epicID, "my-story", "new description")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if story.Description != "new description" {
		t.Errorf("Description = %q, want %q", story.Description, "new description")
	}
}

func TestStoryStore_Update_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	_, err := store.Update(epicID, "nonexistent", "desc")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStoryStore_Delete(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	store.Create(epicID, "doomed", "")
	if err := store.Delete(epicID, "doomed"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.GetByName(epicID, "doomed")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestStoryStore_Delete_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	err := store.Delete(epicID, "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStoryStore_UpdateApprovalStatus(t *testing.T) {
	database := testDB(t)
	epicID := createEpic(t, database, "my-epic")
	store := NewStoryStore(database)

	store.Create(epicID, "my-story", "")
	if err := store.UpdateApprovalStatus(epicID, "my-story", domain.ApprovalPending); err != nil {
		t.Fatalf("UpdateApprovalStatus: %v", err)
	}
	story, _ := store.GetByName(epicID, "my-story")
	if story.ApprovalStatus != domain.ApprovalPending {
		t.Errorf("ApprovalStatus = %q, want %q", story.ApprovalStatus, domain.ApprovalPending)
	}
}
