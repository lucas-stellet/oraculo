package spa_test

import (
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/spa"
)

func TestWithPlaceholders(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"epics/gastos/approvals", "epics/__placeholder__/approvals"},
		{"epics/__placeholder__/approvals", "epics/__placeholder__/approvals"},
		{"epics/gastos/approvals/abc-123/review", "epics/__placeholder__/approvals/__placeholder__/review"},
		{"epics/gastos/stories/registro", "epics/__placeholder__/stories/__placeholder__"},
		{"other/path", "other/path"},
		{"epics", "epics"},
	}
	for _, tt := range tests {
		got := spa.WithPlaceholders(tt.input)
		if got != tt.want {
			t.Errorf("WithPlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShell(t *testing.T) {
	tests := []struct {
		path  string
		isRSC bool
		want  string
	}{
		{"/epics/gastos/approvals/abc/review", false, "/epics/__placeholder__/approvals/__placeholder__/review.html"},
		{"/epics/gastos/approvals", false, "/epics/__placeholder__/approvals.html"},
		{"/epics/gastos/stories/registro", false, "/epics/__placeholder__/stories/__placeholder__.html"},
		{"/epics/gastos", false, "/epics/__placeholder__.html"},
		{"/epics/gastos", true, "/epics/__placeholder__.txt"},
		{"/other", false, "/"},
	}
	for _, tt := range tests {
		got := spa.Shell(tt.path, tt.isRSC)
		if got != tt.want {
			t.Errorf("Shell(%q, %v) = %q, want %q", tt.path, tt.isRSC, got, tt.want)
		}
	}
}
