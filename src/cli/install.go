// src/cli/install.go
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lucas/oraculo/src/config"
	"github.com/lucas/oraculo/src/db"
	"github.com/spf13/cobra"
)

// claudeSettings is the structure written to .claude/settings.json.
type claudeSettings struct {
	Hooks      map[string][]hookEntry `json:"hooks"`
	MCPServers map[string]mcpServer   `json:"mcpServers"`
}

type hookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type mcpServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Oraculo skills and hooks into Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(cmd)
		},
	}
	cmd.Flags().Bool("global", false, "Install globally for all projects")
	cmd.Flags().Bool("local", false, "Install locally for current project")
	return cmd
}

func runInstall(cmd *cobra.Command) error {
	w := cmd.OutOrStdout()

	// Step 1: Create .oraculo/ directory.
	if err := os.MkdirAll(".oraculo", 0o755); err != nil {
		return fmt.Errorf("create .oraculo: %w", err)
	}
	fmt.Fprintln(w, "created .oraculo/")

	// Step 2: Open and close DB to trigger migration.
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	database.Close()
	fmt.Fprintln(w, "created .oraculo/oraculo.db")

	// Step 3: Find an available port.
	port, err := config.FindPort(3100, 3199)
	if err != nil {
		return fmt.Errorf("find port: %w", err)
	}

	// Step 4: Write config.
	if err := config.Write(&config.Config{Port: port}); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Fprintf(w, "created .oraculo/config.json (port %d)\n", port)

	// Step 5: Create .claude/ directory.
	if err := os.MkdirAll(".claude", 0o755); err != nil {
		return fmt.Errorf("create .claude: %w", err)
	}
	fmt.Fprintln(w, "created .claude/")

	// Step 6: Write .claude/settings.json with hooks and MCP config.
	settings := claudeSettings{
		Hooks: map[string][]hookEntry{
			"PreToolUse":  {{Type: "command", Command: "oraculo hook pre-tool $TOOL_NAME"}},
			"PostToolUse": {{Type: "command", Command: "oraculo hook post-tool $TOOL_NAME"}},
			"SessionStart": {{Type: "command", Command: "oraculo hook session-start"}},
			"SessionEnd":  {{Type: "command", Command: "oraculo hook session-end"}},
		},
		MCPServers: map[string]mcpServer{
			"oraculo": {
				Command: "oraculo",
				Args:    []string{"start"},
			},
		},
	}
	settingsData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(filepath.Join(".claude", "settings.json"), settingsData, 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}
	fmt.Fprintln(w, "created .claude/settings.json")

	// Step 7: Copy skills from claude-kit/skills/oraculo/ to .claude/skills/oraculo/ if source exists.
	skillsSrc := filepath.Join("claude-kit", "skills", "oraculo")
	if _, err := os.Stat(skillsSrc); err == nil {
		skillsDst := filepath.Join(".claude", "skills", "oraculo")
		if err := copyDir(skillsSrc, skillsDst); err != nil {
			return fmt.Errorf("copy skills: %w", err)
		}
		fmt.Fprintln(w, "copied skills to .claude/skills/oraculo/")
	}

	fmt.Fprintln(w, "Oraculo installed successfully.")
	return nil
}

// copyDir recursively copies src directory to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
