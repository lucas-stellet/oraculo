package dbtest

import (
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/db"
)

// Open returns an in-memory SQLite database with all migrations applied.
// The database is closed automatically when the test finishes.
func Open(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("dbtest.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
