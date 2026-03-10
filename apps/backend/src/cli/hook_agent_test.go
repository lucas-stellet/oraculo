package cli_test

import (
	"bytes"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/cli"
)

func TestHookAgentStartCmd_MissingRequired(t *testing.T) {
	root := cli.NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"hook", "agent-start"})
	err := root.Execute()
	// Should fail: required flags missing
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}

func TestHookTaskStartedCmd_MissingRequired(t *testing.T) {
	root := cli.NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"hook", "task-started"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}
