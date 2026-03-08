package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

func createTestEpic(t *testing.T, database *DB, name string) int {
	t.Helper()
	epic, _, err := NewEpicStore(database).Create(name, "test epic")
	if err != nil {
		t.Fatalf("createTestEpic(%q): %v", name, err)
	}
	return epic.ID
}

func createTestStory(t *testing.T, database *DB, epicID int, name string) int {
	t.Helper()
	story, _, err := NewStoryStore(database).Create(epicID, name, "test story")
	if err != nil {
		t.Fatalf("createTestStory(%q): %v", name, err)
	}
	return story.ID
}

func TestVersionStore_CreateEpicVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	store := NewVersionStore(database)

	v, err := store.CreateEpicVersion(epicID, "version 1 content")
	if err != nil {
		t.Fatalf("CreateEpicVersion: %v", err)
	}
	if v.Number != 1 {
		t.Errorf("Number = %d, want 1", v.Number)
	}
	if v.EpicID != epicID {
		t.Errorf("EpicID = %d, want %d", v.EpicID, epicID)
	}
	if v.Content != "version 1 content" {
		t.Errorf("Content = %q, want %q", v.Content, "version 1 content")
	}
	if v.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	// Check that epic approval_status was set to pending.
	epic, err := NewEpicStore(database).GetByName("my-epic")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if epic.ApprovalStatus != domain.ApprovalPending {
		t.Errorf("epic ApprovalStatus = %q, want %q", epic.ApprovalStatus, domain.ApprovalPending)
	}
}

func TestVersionStore_CreateEpicVersion_AutoIncrement(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	store := NewVersionStore(database)

	v1, err := store.CreateEpicVersion(epicID, "v1")
	if err != nil {
		t.Fatalf("CreateEpicVersion v1: %v", err)
	}
	v2, err := store.CreateEpicVersion(epicID, "v2")
	if err != nil {
		t.Fatalf("CreateEpicVersion v2: %v", err)
	}
	v3, err := store.CreateEpicVersion(epicID, "v3")
	if err != nil {
		t.Fatalf("CreateEpicVersion v3: %v", err)
	}

	if v1.Number != 1 || v2.Number != 2 || v3.Number != 3 {
		t.Errorf("Numbers = (%d, %d, %d), want (1, 2, 3)", v1.Number, v2.Number, v3.Number)
	}
}

func TestVersionStore_GetEpicVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	store := NewVersionStore(database)

	created, _ := store.CreateEpicVersion(epicID, "content")
	got, err := store.GetEpicVersion(created.ID)
	if err != nil {
		t.Fatalf("GetEpicVersion: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Content != "content" {
		t.Errorf("Content = %q, want %q", got.Content, "content")
	}
}

func TestVersionStore_GetEpicVersion_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewVersionStore(database)

	_, err := store.GetEpicVersion(999)
	if err == nil {
		t.Fatal("expected error for nonexistent version")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestVersionStore_ListEpicVersions(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	store := NewVersionStore(database)

	store.CreateEpicVersion(epicID, "v1")
	store.CreateEpicVersion(epicID, "v2")
	store.CreateEpicVersion(epicID, "v3")

	versions, err := store.ListEpicVersions(epicID)
	if err != nil {
		t.Fatalf("ListEpicVersions: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("len = %d, want 3", len(versions))
	}
	if versions[0].Number != 1 || versions[1].Number != 2 || versions[2].Number != 3 {
		t.Errorf("Numbers = (%d, %d, %d), want (1, 2, 3)", versions[0].Number, versions[1].Number, versions[2].Number)
	}
}

func TestVersionStore_LatestEpicVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	store := NewVersionStore(database)

	store.CreateEpicVersion(epicID, "v1")
	store.CreateEpicVersion(epicID, "v2")
	store.CreateEpicVersion(epicID, "latest content")

	latest, err := store.LatestEpicVersion(epicID)
	if err != nil {
		t.Fatalf("LatestEpicVersion: %v", err)
	}
	if latest.Number != 3 {
		t.Errorf("Number = %d, want 3", latest.Number)
	}
	if latest.Content != "latest content" {
		t.Errorf("Content = %q, want %q", latest.Content, "latest content")
	}
}

func TestVersionStore_LatestEpicVersion_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	store := NewVersionStore(database)

	_, err := store.LatestEpicVersion(epicID)
	if err == nil {
		t.Fatal("expected error when no versions exist")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestVersionStore_CreateStoryVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	store := NewVersionStore(database)

	v, err := store.CreateStoryVersion(storyID, "story version 1")
	if err != nil {
		t.Fatalf("CreateStoryVersion: %v", err)
	}
	if v.Number != 1 {
		t.Errorf("Number = %d, want 1", v.Number)
	}
	if v.StoryID != storyID {
		t.Errorf("StoryID = %d, want %d", v.StoryID, storyID)
	}
	if v.Content != "story version 1" {
		t.Errorf("Content = %q, want %q", v.Content, "story version 1")
	}

	// Check that story approval_status was set to pending.
	story, err := NewStoryStore(database).GetByName(epicID, "my-story")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if story.ApprovalStatus != domain.ApprovalPending {
		t.Errorf("story ApprovalStatus = %q, want %q", story.ApprovalStatus, domain.ApprovalPending)
	}
}

func TestVersionStore_CreateStoryVersion_AutoIncrement(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	store := NewVersionStore(database)

	v1, _ := store.CreateStoryVersion(storyID, "v1")
	v2, _ := store.CreateStoryVersion(storyID, "v2")
	v3, _ := store.CreateStoryVersion(storyID, "v3")

	if v1.Number != 1 || v2.Number != 2 || v3.Number != 3 {
		t.Errorf("Numbers = (%d, %d, %d), want (1, 2, 3)", v1.Number, v2.Number, v3.Number)
	}
}

func TestVersionStore_GetStoryVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	store := NewVersionStore(database)

	created, _ := store.CreateStoryVersion(storyID, "content")
	got, err := store.GetStoryVersion(created.ID)
	if err != nil {
		t.Fatalf("GetStoryVersion: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
}

func TestVersionStore_GetStoryVersion_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewVersionStore(database)

	_, err := store.GetStoryVersion(999)
	if err == nil {
		t.Fatal("expected error for nonexistent version")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestVersionStore_ListStoryVersions(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	store := NewVersionStore(database)

	store.CreateStoryVersion(storyID, "v1")
	store.CreateStoryVersion(storyID, "v2")

	versions, err := store.ListStoryVersions(storyID)
	if err != nil {
		t.Fatalf("ListStoryVersions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("len = %d, want 2", len(versions))
	}
	if versions[0].Number != 1 || versions[1].Number != 2 {
		t.Errorf("Numbers = (%d, %d), want (1, 2)", versions[0].Number, versions[1].Number)
	}
}

func TestVersionStore_LatestStoryVersion(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	store := NewVersionStore(database)

	store.CreateStoryVersion(storyID, "v1")
	store.CreateStoryVersion(storyID, "latest")

	latest, err := store.LatestStoryVersion(storyID)
	if err != nil {
		t.Fatalf("LatestStoryVersion: %v", err)
	}
	if latest.Number != 2 {
		t.Errorf("Number = %d, want 2", latest.Number)
	}
	if latest.Content != "latest" {
		t.Errorf("Content = %q, want %q", latest.Content, "latest")
	}
}

func TestVersionStore_LatestStoryVersion_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createTestEpic(t, database, "my-epic")
	storyID := createTestStory(t, database, epicID, "my-story")
	store := NewVersionStore(database)

	_, err := store.LatestStoryVersion(storyID)
	if err == nil {
		t.Fatal("expected error when no versions exist")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
