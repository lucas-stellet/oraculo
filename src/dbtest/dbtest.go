// Package dbtest provides test helpers for opening in-memory databases.
package dbtest

import (
	"testing"

	"github.com/lucas/oraculo/src/db"
)

// Open returns an in-memory *db.DB with all migrations applied.
// The database is closed automatically when the test ends.
func Open(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("dbtest.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
