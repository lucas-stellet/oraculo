// apps/backend/src/cli/hook_session.go
package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/spf13/cobra"
)

func newHookSessionStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-start",
		Short: "Register a Claude Code session start",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Always succeed — never block Claude Code session
			if err := hookSessionStart(cmd); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook session-start: %v\n", err)
			}
			return nil
		},
	}
}

// hookInput is the JSON structure Claude Code sends on stdin for all hooks.
type hookInput struct {
	SessionID string `json:"session_id"`
	Source    string `json:"source"`
}

func hookSessionStart(cmd *cobra.Command) error {
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Read Claude Code's JSON input from stdin
	var input hookInput
	if err := json.NewDecoder(cmd.InOrStdin()).Decode(&input); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if input.SessionID == "" {
		return fmt.Errorf("missing session_id in stdin")
	}

	// Collect metadata
	wd, _ := os.Getwd()
	branch := gitBranch()
	metadata := map[string]string{
		"working_dir": wd,
		"git_branch":  branch,
		"source":      input.Source,
		"updated_at":  time.Now().UTC().Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Upsert: INSERT OR IGNORE for first time, then UPDATE metadata
	_, err = database.Conn().Exec(
		"INSERT OR IGNORE INTO claude_sessions (id, metadata) VALUES (?, ?)",
		input.SessionID, string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("register session: %w", err)
	}
	_, err = database.Conn().Exec(
		"UPDATE claude_sessions SET metadata = ? WHERE id = ?",
		string(metadataJSON), input.SessionID,
	)
	if err != nil {
		return fmt.Errorf("update session metadata: %w", err)
	}

	// Health check and auto-start
	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return nil
	}

	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	online := isServerHealthy(healthURL)

	if !online {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: Oraculo HTTP server offline — auto-starting on port %d\n", port)
		if err := SpawnDaemon(); err != nil {
			msg := fmt.Sprintf("warning: failed to auto-start Oraculo server: %v", err)
			fmt.Fprintln(cmd.ErrOrStderr(), msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		}
		online = pollHealth(healthURL, 500*time.Millisecond, 10*time.Second)
		if !online {
			msg := "warning: Oraculo server started but not responding. Telemetry unavailable."
			fmt.Fprintln(cmd.ErrOrStderr(), msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		}
	}

	// POST session-start
	postURL := fmt.Sprintf("http://localhost:%d/hooks/session-start", port)
	client := &http.Client{Timeout: 2 * time.Second}
	client.Post(postURL, "application/json", strings.NewReader(string(metadataJSON)))

	return nil
}

func newHookSessionEndCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-end",
		Short: "Register a Claude Code session end",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hookSessionEnd(cmd); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook session-end: %v\n", err)
			}
			return nil
		},
	}
}

func hookSessionEnd(cmd *cobra.Command) error {
	var input struct {
		SessionID string `json:"session_id"`
		Reason    string `json:"reason"`
	}
	if err := json.NewDecoder(cmd.InOrStdin()).Decode(&input); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if input.SessionID == "" {
		return fmt.Errorf("missing session_id in stdin")
	}

	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Update ended_at
	database.Conn().Exec(
		"UPDATE claude_sessions SET ended_at = datetime('now') WHERE id = ?",
		input.SessionID,
	)

	// Record session event
	payload, _ := json.Marshal(map[string]string{"reason": input.Reason})
	database.Conn().Exec(
		"INSERT INTO session_events (session_id, event_type, payload) VALUES (?, 'session_end', ?)",
		input.SessionID, string(payload),
	)

	// Notify HTTP server if online
	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return nil
	}
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !isServerHealthy(healthURL) {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"session_id": input.SessionID, "reason": input.Reason})
	postURL := fmt.Sprintf("http://localhost:%d/hooks/session-end", port)
	client := &http.Client{Timeout: 2 * time.Second}
	client.Post(postURL, "application/json", strings.NewReader(string(body)))

	return nil
}

func isServerHealthy(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	return err == nil && resp.StatusCode == http.StatusOK
}

// SpawnDaemon is the function used to start the HTTP daemon. It is exported
// so tests can replace it with a no-op stub instead of spawning real detached
// processes (which would be fork-bomb-like during go test).
var SpawnDaemon = startHTTPDaemon

func startHTTPDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		exe = "oraculo"
	}
	// Guard: never spawn from a test binary.
	if strings.HasSuffix(exe, ".test") {
		return fmt.Errorf("refusing to spawn daemon from test binary: %s", exe)
	}
	cmd := exec.Command(exe, "start", "http")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func pollHealth(url string, interval, timeout time.Duration) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			if isServerHealthy(url) {
				return true
			}
		}
	}
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
