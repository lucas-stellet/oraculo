package cli_test

import (
	"bytes"
	"testing"

	"github.com/lucas/oraculo/src/cli"
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
