package db_test

import (
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/dbtest"
)

func TestMigrationV6_SessionEventsTableExists(t *testing.T) {
	database := dbtest.Open(t)
	var name string
	err := database.Conn().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='session_events'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("session_events table not found: %v", err)
	}
}

func TestMigrationV6_EndedAtColumnExists(t *testing.T) {
	database := dbtest.Open(t)
	_, err := database.Conn().Exec(
		"INSERT INTO claude_sessions (id) VALUES ('test-session')",
	)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
	_, err = database.Conn().Exec(
		"UPDATE claude_sessions SET ended_at = datetime('now') WHERE id = 'test-session'",
	)
	if err != nil {
		t.Fatalf("ended_at column missing or broken: %v", err)
	}
}

func seedSession(t *testing.T, database *db.DB, id string) {
	t.Helper()
	_, err := database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES (?)", id)
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}
}

func TestSessionEventStore_Record(t *testing.T) {
	database := dbtest.Open(t)
	seedSession(t, database, "s1")
	store := db.NewSessionEventStore(database)

	err := store.Record("s1", "stop", `{"last_assistant_message":"done"}`)
	if err != nil {
		t.Fatalf("record event: %v", err)
	}

	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "stop" {
		t.Errorf("expected event_type=stop, got %s", events[0].EventType)
	}
	if events[0].SessionID != "s1" {
		t.Errorf("expected session_id=s1, got %s", events[0].SessionID)
	}
}

func TestSessionEventStore_ListBySession_MultipleEvents(t *testing.T) {
	database := dbtest.Open(t)
	seedSession(t, database, "s1")
	seedSession(t, database, "s2")
	store := db.NewSessionEventStore(database)

	store.Record("s1", "stop", "{}")
	store.Record("s1", "task_completed", `{"task_name":"auth"}`)
	store.Record("s2", "stop", "{}")

	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events for s1, got %d", len(events))
	}
}

func TestSessionEventStore_InvalidEventType(t *testing.T) {
	database := dbtest.Open(t)
	seedSession(t, database, "s1")
	store := db.NewSessionEventStore(database)

	err := store.Record("s1", "invalid_type", "{}")
	if err == nil {
		t.Fatal("expected error for invalid event type")
	}
}
