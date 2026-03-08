// src/db/validation_store_test.go
package db

import (
	"testing"
)

func TestValidationStore_Save(t *testing.T) {
	database := testDB(t)

	// Create epic and story first.
	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	storyStore := NewStoryStore(database)
	story, _, _ := storyStore.Create(epic.ID, "test-story", "desc")

	store := NewValidationStore(database)
	v, err := store.Save(story.ID, "approved", `{"blocker":0,"major":0,"minor":1}`)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if v.StoryID != story.ID {
		t.Errorf("StoryID = %d, want %d", v.StoryID, story.ID)
	}
	if v.Verdict != "approved" {
		t.Errorf("Verdict = %q, want %q", v.Verdict, "approved")
	}
	if v.Findings != `{"blocker":0,"major":0,"minor":1}` {
		t.Errorf("Findings = %q", v.Findings)
	}
	if v.TaskID != nil {
		t.Errorf("TaskID = %v, want nil", v.TaskID)
	}
}

func TestValidationStore_Save_Rejected(t *testing.T) {
	database := testDB(t)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	storyStore := NewStoryStore(database)
	story, _, _ := storyStore.Create(epic.ID, "test-story", "desc")

	store := NewValidationStore(database)
	v, err := store.Save(story.ID, "rejected", `{"blocker":1}`)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if v.Verdict != "rejected" {
		t.Errorf("Verdict = %q, want %q", v.Verdict, "rejected")
	}
}

func TestValidationStore_ListByStory(t *testing.T) {
	database := testDB(t)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	storyStore := NewStoryStore(database)
	story, _, _ := storyStore.Create(epic.ID, "test-story", "desc")

	store := NewValidationStore(database)
	store.Save(story.ID, "rejected", `{"blocker":1}`)
	store.Save(story.ID, "approved", `{"blocker":0}`)

	results, err := store.ListByStory(story.ID)
	if err != nil {
		t.Fatalf("ListByStory: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
	if results[0].Verdict != "rejected" {
		t.Errorf("results[0].Verdict = %q, want %q", results[0].Verdict, "rejected")
	}
	if results[1].Verdict != "approved" {
		t.Errorf("results[1].Verdict = %q, want %q", results[1].Verdict, "approved")
	}
}

func TestValidationStore_ListByStory_Empty(t *testing.T) {
	database := testDB(t)

	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("test-epic", "desc")
	storyStore := NewStoryStore(database)
	story, _, _ := storyStore.Create(epic.ID, "test-story", "desc")

	store := NewValidationStore(database)
	results, err := store.ListByStory(story.ID)
	if err != nil {
		t.Fatalf("ListByStory: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len = %d, want 0", len(results))
	}
}
