// src/output/json_test.go
package output

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	err := WriteJSON(&buf, map[string]any{"name": "test", "created": true})
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"name"`) {
		t.Error("missing name key")
	}
	if !strings.Contains(got, `"created"`) {
		t.Error("missing created key")
	}
	if got[len(got)-1] != '\n' {
		t.Error("missing trailing newline")
	}
}

func TestWriteError_NotFound(t *testing.T) {
	var buf bytes.Buffer
	err := fmt.Errorf("epic %q: %w", "test", domain.ErrNotFound)
	WriteError(&buf, err)
	got := buf.String()
	if !strings.Contains(got, `"not_found"`) {
		t.Errorf("expected not_found code in: %s", got)
	}
}

func TestWriteError_InvalidTransition(t *testing.T) {
	var buf bytes.Buffer
	err := fmt.Errorf("task: %w", domain.ErrInvalidTransition)
	WriteError(&buf, err)
	got := buf.String()
	if !strings.Contains(got, `"invalid_transition"`) {
		t.Errorf("expected invalid_transition code in: %s", got)
	}
}

func TestWriteError_CyclicDependency(t *testing.T) {
	var buf bytes.Buffer
	WriteError(&buf, domain.ErrCyclicDependency)
	got := buf.String()
	if !strings.Contains(got, `"cyclic_dependency"`) {
		t.Errorf("expected cyclic_dependency code in: %s", got)
	}
}

func TestWriteError_MissingRequired(t *testing.T) {
	var buf bytes.Buffer
	WriteError(&buf, domain.ErrMissingRequired)
	got := buf.String()
	if !strings.Contains(got, `"missing_required"`) {
		t.Errorf("expected missing_required code in: %s", got)
	}
}

func TestWriteError_UnknownError(t *testing.T) {
	var buf bytes.Buffer
	WriteError(&buf, fmt.Errorf("something unexpected"))
	got := buf.String()
	if !strings.Contains(got, `"unknown_error"`) {
		t.Errorf("expected unknown_error code in: %s", got)
	}
}
