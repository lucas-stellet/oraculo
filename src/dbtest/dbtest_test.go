package dbtest_test

import (
	"testing"

	"github.com/lucas/oraculo/src/dbtest"
)

func TestOpen(t *testing.T) {
	database := dbtest.Open(t)
	if database == nil {
		t.Fatal("expected non-nil database")
	}
	// Verify we can query — migrations ran
	var count int
	err := database.Conn().QueryRow("SELECT count(*) FROM epics").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
}
