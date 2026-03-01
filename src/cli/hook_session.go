// src/cli/hook_session.go
package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
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

	// Health check if config exists
	port := readConfigPort()
	if port > 0 {
		healthURL := fmt.Sprintf("http://localhost:%d/health", port)
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(healthURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: Oraculo dashboard is offline. Run 'oraculo server' to start it.\n")
			return nil
		}
		// POST session-start
		postURL := fmt.Sprintf("http://localhost:%d/hooks/session-start", port)
		client.Post(postURL, "application/json", strings.NewReader(string(metadataJSON)))
	}
	return nil
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

type configFile struct {
	Port int `json:"port"`
}

func readConfigPort() int {
	data, err := os.ReadFile(".oraculo/config.json")
	if err != nil {
		return 0
	}
	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return 0
	}
	return cfg.Port
}
