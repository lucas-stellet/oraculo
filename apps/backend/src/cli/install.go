// apps/backend/src/cli/install.go
package cli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	claudekit "github.com/lucas/oraculo/claude-kit"
	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/spf13/cobra"
)

// claudeSettings is the structure written to .claude/settings.json.
type claudeSettings struct {
	Hooks      map[string][]hookGroup `json:"hooks"`
	MCPServers map[string]mcpServer   `json:"mcpServers"`
}

// hookGroup is the Claude Code hooks format: an optional regex matcher paired
// with an array of hook definitions. Matcher filters which tools trigger the hook.
type hookGroup struct {
	Matcher string    `json:"matcher,omitempty"`
	Hooks   []hookDef `json:"hooks"`
}

// hookDef is a single hook action — either a command or an HTTP request.
type hookDef struct {
	Type    string `json:"type"`
	Command string `json:"command,omitempty"` // type "command"
	URL     string `json:"url,omitempty"`     // type "http"
	Timeout int    `json:"timeout,omitempty"` // type "http"
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
	cmd.Flags().String("lang", "", "Preferred language for Oraculo sessions (BCP 47, e.g. pt-BR, en-US)")
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
	lang, _ := cmd.Flags().GetString("lang")
	if lang == "" {
		lang = "pt-BR"
	}
	cfg := &config.Config{Port: port, PreferredLanguage: lang}
	if err := config.Write(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Fprintf(w, "created .oraculo/config.json (port %d)\n", port)

	// Step 5: Create .claude/ directory.
	if err := os.MkdirAll(".claude", 0o755); err != nil {
		return fmt.Errorf("create .claude: %w", err)
	}
	fmt.Fprintln(w, "created .claude/")

	// Step 6: Write .claude/settings.json with hooks and MCP config.
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	httpGroup := func(url string) []hookGroup {
		return []hookGroup{{Hooks: []hookDef{{Type: "http", URL: url, Timeout: 5}}}}
	}
	settings := claudeSettings{
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

	// Step 7: Copy embedded skills to .claude/skills/oraculo-<name>/.
	skillsSrcRoot := filepath.Join("skills", "oraculo")
	entries, err := fs.ReadDir(claudekit.SkillsFS, skillsSrcRoot)
	if err != nil {
		return fmt.Errorf("read embedded skills: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		src := filepath.Join(skillsSrcRoot, e.Name())
		dst := filepath.Join(".claude", "skills", "oraculo-"+e.Name())
		if err := copyFromFS(claudekit.SkillsFS, src, dst); err != nil {
			return fmt.Errorf("copy skill %s: %w", e.Name(), err)
		}
	}
	fmt.Fprintln(w, "copied skills to .claude/skills/")

	fmt.Fprintln(w, "Oraculo installed successfully.")
	return nil
}

// copyFromFS recursively copies srcDir from fsys into dst on the local filesystem.
func copyFromFS(fsys fs.FS, srcDir, dst string) error {
	return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
