// apps/backend/src/cli/hook_agent.go
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/spf13/cobra"
)

func newHookAgentStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent-start",
		Short: "Register an agent start with task association",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Always succeed — never block orchestrator
			if err := hookAgentStart(cmd); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook agent-start: %v\n", err)
			}
			return nil
		},
	}
	cmd.Flags().String("session-id", "", "Claude Code session ID")
	cmd.Flags().String("agent-name", "", "Agent name (required)")
	cmd.Flags().String("agent-type", "code", "Agent type (code, research, qa)")
	cmd.Flags().String("task-name", "", "Task name (required)")
	cmd.Flags().String("story-name", "", "Story name (required)")
	cmd.Flags().String("epic-name", "", "Epic name (required)")
	_ = cmd.MarkFlagRequired("agent-name")
	_ = cmd.MarkFlagRequired("task-name")
	_ = cmd.MarkFlagRequired("story-name")
	_ = cmd.MarkFlagRequired("epic-name")
	return cmd
}

func hookAgentStart(cmd *cobra.Command) error {
	agentName, _ := cmd.Flags().GetString("agent-name")
	agentType, _ := cmd.Flags().GetString("agent-type")
	taskName, _ := cmd.Flags().GetString("task-name")
	storyName, _ := cmd.Flags().GetString("story-name")
	epicName, _ := cmd.Flags().GetString("epic-name")
	sessionID, _ := cmd.Flags().GetString("session-id")

	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return fmt.Errorf("server port not configured")
	}
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !isServerHealthy(healthURL) {
		return fmt.Errorf("server not reachable on port %d", port)
	}

	payload, _ := json.Marshal(map[string]string{
		"session_id": sessionID,
		"agent_name": agentName,
		"agent_type": agentType,
		"task_name":  taskName,
		"story_name": storyName,
		"epic_name":  epicName,
	})
	postURL := fmt.Sprintf("http://localhost:%d/hooks/agent-start", port)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(postURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("post agent-start: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent-start returned %d", resp.StatusCode)
	}
	return nil
}

func newHookTaskStartedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task-started",
		Short: "Broadcast a task-started event",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hookTaskStarted(cmd); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook task-started: %v\n", err)
			}
			return nil
		},
	}
	cmd.Flags().String("task-name", "", "Task name (required)")
	cmd.Flags().String("story-name", "", "Story name (required)")
	cmd.Flags().String("epic-name", "", "Epic name (required)")
	_ = cmd.MarkFlagRequired("task-name")
	_ = cmd.MarkFlagRequired("story-name")
	_ = cmd.MarkFlagRequired("epic-name")
	return cmd
}

func hookTaskStarted(cmd *cobra.Command) error {
	taskName, _ := cmd.Flags().GetString("task-name")
	storyName, _ := cmd.Flags().GetString("story-name")
	epicName, _ := cmd.Flags().GetString("epic-name")

	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return fmt.Errorf("server port not configured")
	}
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !isServerHealthy(healthURL) {
		return fmt.Errorf("server not reachable on port %d", port)
	}

	payload, _ := json.Marshal(map[string]string{
		"task_name":  taskName,
		"story_name": storyName,
		"epic_name":  epicName,
	})
	postURL := fmt.Sprintf("http://localhost:%d/hooks/task-started", port)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(postURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("post task-started: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("task-started returned %d", resp.StatusCode)
	}
	return nil
}
