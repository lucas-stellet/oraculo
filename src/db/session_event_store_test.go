package db_test

import (
	"testing"

	"github.com/lucas/oraculo/src/dbtest"
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
