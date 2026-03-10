package db

import (
	"testing"
)

func TestAgentStore_StartAndStop(t *testing.T) {
	database := testDB(t)
	store := NewAgentStore(database)

	agent, err := store.Start("session-1", "code-agent", "code", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if agent.Name != "code-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "code-agent")
	}
	if agent.Status != "active" {
		t.Errorf("Status = %q, want %q", agent.Status, "active")
	}

	stopped, err := store.Stop(agent.ID, "completed")
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stopped.Status != "completed" {
		t.Errorf("Status = %q, want %q", stopped.Status, "completed")
	}
	if stopped.StoppedAt == nil {
		t.Error("StoppedAt should not be nil")
	}
}

func TestAgentStore_ListBySession(t *testing.T) {
	database := testDB(t)
	store := NewAgentStore(database)

	store.Start("session-1", "agent-a", "code", nil)
	store.Start("session-1", "agent-b", "qa", nil)
	store.Start("session-2", "agent-c", "research", nil)

	agents, err := store.ListBySession("session-1")
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2", len(agents))
	}
}

func TestAgentStore_StartWithTaskID(t *testing.T) {
	database := testDB(t)

	// Seed an epic, story, and task so we have a valid task_id.
	epicStore := NewEpicStore(database)
	epic, _, _ := epicStore.Create("gastos", "")
	storyStore := NewStoryStore(database)
	story, _, _ := storyStore.Create(epic.ID, "registro", "")
	taskStore := NewTaskStore(database)
	task, _, _ := taskStore.Create(story.ID, "implement-api", "", nil)

	store := NewAgentStore(database)
	taskID := task.ID
	agent, err := store.Start("s1", "code-01", "code", &taskID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if agent.TaskID == nil {
		t.Fatal("TaskID should not be nil")
	}
	if *agent.TaskID != task.ID {
		t.Errorf("TaskID = %d, want %d", *agent.TaskID, task.ID)
	}
}

func TestAgentStore_StartWithoutTaskID(t *testing.T) {
	database := testDB(t)
	store := NewAgentStore(database)
	agent, err := store.Start("s1", "code-01", "code", nil)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if agent.TaskID != nil {
		t.Errorf("TaskID should be nil, got %d", *agent.TaskID)
	}
}
