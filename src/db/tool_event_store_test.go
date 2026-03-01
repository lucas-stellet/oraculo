package db

import (
	"testing"
)

func TestToolEventStore_RecordAndList(t *testing.T) {
	database := testDB(t)
	store := NewToolEventStore(database)

	err := store.Record("session-1", "Edit", "/src/main.go")
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	err = store.Record("session-1", "Bash", "")
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	events, err := store.ListBySession("session-1", 50)
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
	if events[0].ToolName != "Edit" {
		t.Errorf("ToolName = %q, want %q", events[0].ToolName, "Edit")
	}
}
