// src/cli/logs_test.go
package cli

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			`data: {"time":"2026-01-01T00:00:00Z","level":"INFO","msg":"mcp.connected"}`,
			`2026-01-01T00:00:00Z INFO mcp.connected`,
		},
		{
			`data: {"time":"2026-01-01T00:00:00Z","level":"WARN","msg":"ws.drop"}`,
			`2026-01-01T00:00:00Z WARN ws.drop`,
		},
		{"event: ping", ""},
		{"", ""},
		{"data: not-json", "not-json"},
	}
	for _, tc := range cases {
		got := parseLine(tc.input)
		if got != tc.want {
			t.Errorf("parseLine(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
