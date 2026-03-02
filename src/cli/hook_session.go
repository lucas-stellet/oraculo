// src/cli/hook_session.go
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

	"github.com/google/uuid"
	"github.com/lucas/oraculo/src/config"
	"github.com/lucas/oraculo/src/db"
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

func hookSessionStart(cmd *cobra.Command) error {
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Collect metadata
	id := uuid.New().String()
	wd, _ := os.Getwd()
	branch := gitBranch()
	metadata := map[string]string{
		"session_id":  id,
		"working_dir": wd,
		"git_branch":  branch,
		"started_at":  time.Now().UTC().Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Register in SQLite
	_, err = database.Conn().Exec(
		"INSERT INTO claude_sessions (id, metadata) VALUES (?, ?)",
		id, string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("register session: %w", err)
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
		if err := startHTTPDaemon(); err != nil {
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

func isServerHealthy(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	return err == nil && resp.StatusCode == http.StatusOK
}

func startHTTPDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		exe = "oraculo"
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
