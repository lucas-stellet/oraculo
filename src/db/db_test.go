package db

import "testing"

func TestOpen_CreatesSchemaOnFreshDB(t *testing.T) {
	database := testDB(t)

	tables := []string{"epics", "stories", "tasks", "task_dependencies",
		"task_results", "validations", "knowledge", "approvals"}
	for _, table := range tables {
		var name string
		err := database.conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestOpen_MigrationIsIdempotent(t *testing.T) {
	database := testDB(t)

	if err := database.migrate(); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}

func TestOpen_TracksSchemaVersion(t *testing.T) {
	database := testDB(t)

	var version int
	err := database.conn.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		t.Fatalf("failed to read user_version: %v", err)
	}
	if version != len(migrations) {
		t.Errorf("user_version = %d, want %d", version, len(migrations))
	}
}
