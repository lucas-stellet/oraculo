// Package dbtest provides test helpers for creating in-memory SQLite databases.
package dbtest

import (
	"testing"

	"github.com/lucas/oraculo/src/db"
)

// Open opens an in-memory SQLite database with all migrations applied.
// The database is closed automatically when the test ends.
func Open(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.OpenInMemory()
	if err != nil {
		t.Fatalf("dbtest.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
