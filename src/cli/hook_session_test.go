package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestHookSessionStart_AlertsWhenServerOffline(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// Write config with port that nothing is listening on
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)
	os.WriteFile(
		filepath.Join(tmp, ".oraculo", "config.json"),
		[]byte(`{"port": 39999}`),
		0o644,
	)

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook should never return error: %v", err)
	}
	stderr := buf.String()
	if !strings.Contains(stderr, "warning") {
		t.Errorf("expected warning in output, got: %s", stderr)
	}
}

func TestHookSessionStart_AttemptsAutoStart(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)
	os.WriteFile(
		filepath.Join(tmp, ".oraculo", "config.json"),
		[]byte(`{"port": 39999}`),
		0o644,
	)

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	cmd.Execute()

	output := buf.String()
	// Should attempt auto-start and report failure (oraculo binary may not exist in test)
	if !strings.Contains(output, "auto-start") && !strings.Contains(output, "starting") {
		t.Errorf("expected auto-start attempt in output, got: %s", output)
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
