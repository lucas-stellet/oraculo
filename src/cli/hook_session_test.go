package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

func TestHookSessionStart_NoConfig(t *testing.T) {
	// Setup temp dir
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	// Must always succeed (exit 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, buf.String())
	}

	// Should have created .oraculo/ and registered session
	dbPath := filepath.Join(".oraculo", "oraculo.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db to be created")
	}
}

func TestHookSessionStart_AlwaysExitsZero(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook must always exit 0, got: %v", err)
	}
}
