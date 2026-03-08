package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

func TestStartCmd_HasSubcommands(t *testing.T) {
	root := cli.NewRoot("test")

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"start", "--help"})
	root.Execute()

	out := buf.String()
	for _, sub := range []string{"mcp", "http"} {
		if !bytes.Contains([]byte(out), []byte(sub)) {
			t.Errorf("start --help missing subcommand %q", sub)
		}
	}
}

func TestStartMCPCmd_Exists(t *testing.T) {
	root := cli.NewRoot("test")

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"start", "mcp", "--help"})
	err := root.Execute()

	if err != nil {
		t.Fatalf("start mcp --help failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "MCP") && !strings.Contains(out, "mcp") {
		t.Errorf("start mcp --help missing MCP reference, got: %s", out)
	}
}

func TestStartHTTPCmd_Exists(t *testing.T) {
	root := cli.NewRoot("test")

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"start", "http", "--help"})
	err := root.Execute()

	if err != nil {
		t.Fatalf("start http --help failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "HTTP") && !strings.Contains(out, "http") {
		t.Errorf("start http --help missing HTTP reference, got: %s", out)
	}
}

func TestStartCmd_BackwardsCompat(t *testing.T) {
	// "start" without subcommand should still be valid (runs both)
	root := cli.NewRoot("test")

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"start", "--help"})
	err := root.Execute()

	if err != nil {
		t.Fatalf("start --help failed: %v", err)
	}
	out := buf.String()
	// Should mention both subcommands and the parent description
	if !strings.Contains(out, "mcp") || !strings.Contains(out, "http") {
		t.Errorf("start --help should list both subcommands, got: %s", out)
	}
}
