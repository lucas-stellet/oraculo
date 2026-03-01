package db

import (
	"testing"
)

func TestAgentStore_StartAndStop(t *testing.T) {
	database := testDB(t)
	store := NewAgentStore(database)

	agent, err := store.Start("session-1", "code-agent", "code")
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

	store.Start("session-1", "agent-a", "code")
	store.Start("session-1", "agent-b", "qa")
	store.Start("session-2", "agent-c", "research")

	agents, err := store.ListBySession("session-1")
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2", len(agents))
	}
}
