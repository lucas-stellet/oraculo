// apps/backend/src/output/table_test.go
package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable(t *testing.T) {
	var buf bytes.Buffer
	err := WriteTable(&buf, []string{"Name", "Status"}, [][]string{
		{"alpha", "pending"},
		{"beta", "completed"},
	})
	if err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Name") {
		t.Error("missing header")
	}
	if !strings.Contains(got, "alpha") {
		t.Error("missing row data")
	}
	if !strings.Contains(got, "beta") {
		t.Error("missing second row")
	}
}

func TestWriteTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := WriteTable(&buf, []string{}, nil)
	if err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	if buf.Len() != 0 {
		t.Error("expected empty output for no headers")
	}
}

func TestWriteTable_ColumnAlignment(t *testing.T) {
	var buf bytes.Buffer
	WriteTable(&buf, []string{"Name", "Description"}, [][]string{
		{"a", "short"},
		{"longer-name", "also long"},
	})
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected at least 4 lines (header + separator + 2 rows), got %d", len(lines))
	}
}
