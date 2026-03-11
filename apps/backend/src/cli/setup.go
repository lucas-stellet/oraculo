// apps/backend/src/cli/setup.go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Initialize Oraculo infrastructure (for plugin users)",
		Long: `Set up .oraculo/ directory, database, and hooks without copying skills
or configuring MCP. Use this when skills and MCP are provided by
the Claude Code plugin.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd)
		},
	}
	cmd.Flags().String("lang", "", "Preferred language for Oraculo sessions (BCP 47, e.g. pt-BR, en-US)")
	return cmd
}

type claudeSettingsHooksOnly struct {
	Hooks map[string][]hookGroup `json:"hooks"`
}

func runSetup(cmd *cobra.Command) error {
	w := cmd.OutOrStdout()

	if err := os.MkdirAll(".oraculo", 0o755); err != nil {
		return fmt.Errorf("create .oraculo: %w", err)
	}
	fmt.Fprintln(w, "created .oraculo/")

	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	database.Close()
	fmt.Fprintln(w, "created .oraculo/oraculo.db")

	existing, _ := config.Read()

	port := 0
	if existing != nil {
		port = existing.Port
	}
	if port == 0 {
		port, err = config.FindPort(3100, 3199)
		if err != nil {
			return fmt.Errorf("find port: %w", err)
		}
	}

	lang, _ := cmd.Flags().GetString("lang")
	if lang == "" {
		if existing != nil && existing.PreferredLanguage != "" {
			lang = existing.PreferredLanguage
		} else {
			lang = "pt-BR"
		}
	}
	cfg := &config.Config{Port: port, PreferredLanguage: lang}
	if existing != nil {
		cfg.Skills = existing.Skills
	}
	if err := config.Write(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Fprintf(w, "created .oraculo/config.json (port %d)\n", port)

	if err := os.MkdirAll(".claude", 0o755); err != nil {
		return fmt.Errorf("create .claude: %w", err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	httpGroup := func(url string) []hookGroup {
		return []hookGroup{{Hooks: []hookDef{{Type: "http", URL: url, Timeout: 5}}}}
	}
	settings := claudeSettingsHooksOnly{
		Hooks: map[string][]hookGroup{
			"SessionStart": {{
				Hooks: []hookDef{{Type: "command", Command: "oraculo hook session-start"}},
			}},
			"SubagentStart":  httpGroup(baseURL + "/hooks/agent-start"),
			"SubagentStop":   httpGroup(baseURL + "/hooks/agent-stop"),
			"TaskCompleted":  httpGroup(baseURL + "/hooks/task-completed"),
			"Stop":           httpGroup(baseURL + "/hooks/stop"),
			"TeammateIdle":   httpGroup(baseURL + "/hooks/teammate-idle"),
			"SessionEnd": {{
				Hooks: []hookDef{{Type: "command", Command: "oraculo hook session-end"}},
			}},
			"PostToolUse": {{
				Matcher: "Bash|Edit|Write|NotebookEdit",
				Hooks:   []hookDef{{Type: "http", URL: baseURL + "/hooks/tool-used", Timeout: 5}},
			}},
		},
	}
	settingsData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(filepath.Join(".claude", "settings.json"), settingsData, 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}
	fmt.Fprintln(w, "created .claude/settings.json (hooks only)")

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}
	mcpConfig := map[string]any{
		"mcpServers": map[string]any{
			"oraculo": map[string]any{
				"command": binaryPath,
				"args":    []string{"start"},
			},
		},
	}
	mcpData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mcp config: %w", err)
	}
	if err := os.WriteFile(".mcp.json", mcpData, 0o644); err != nil {
		return fmt.Errorf("write .mcp.json: %w", err)
	}
	fmt.Fprintf(w, "created .mcp.json (command: %s start)\n", binaryPath)

	fmt.Fprintln(w, "Oraculo setup complete. Install the plugin for skills and MCP.")
	return nil
}
