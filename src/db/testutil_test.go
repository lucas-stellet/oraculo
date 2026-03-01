package db

import "testing"

// testDB opens an in-memory SQLite database with all migrations applied.
// The database is closed automatically when the test ends.
func testDB(t *testing.T) *DB {
	t.Helper()
	database, err := openPath(":memory:")
	if err != nil {
		t.Fatalf("testDB: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
