// apps/backend/src/cli/uninstall.go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove Oraculo from the current project",
		RunE:  runUninstall,
	}
	cmd.Flags().Bool("purge", false, "Also remove .oraculo/ directory and database")
	return cmd
}

func runUninstall(cmd *cobra.Command, args []string) error {
	w := cmd.OutOrStdout()

	if err := cleanSettings(w); err != nil {
		return err
	}

	if err := removeSkills(w); err != nil {
		return err
	}

	purge, _ := cmd.Flags().GetBool("purge")
	if purge {
		if err := removeOraculoDir(w); err != nil {
			return err
		}
	}

	fmt.Fprintln(w, "Oraculo uninstalled.")
	return nil
}

// cleanSettings reads .claude/settings.json and removes all Oraculo-specific
// entries (hooks whose command starts with "oraculo" and the "oraculo" MCP
// server), then writes the cleaned file back.
func cleanSettings(w interface{ Write([]byte) (int, error) }) error {
	settingsPath := filepath.Join(".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if os.IsNotExist(err) {
		// Nothing to clean — not an error.
		return nil
	}
	if err != nil {
		return fmt.Errorf("read settings.json: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse settings.json: %w", err)
	}

	// Remove "oraculo" from mcpServers.
	if mcpRaw, ok := settings["mcpServers"]; ok {
		if mcp, ok := mcpRaw.(map[string]any); ok {
			delete(mcp, "oraculo")
			settings["mcpServers"] = mcp
		}
	}

	// Remove hook entries that belong to Oraculo (those starting with "oraculo").
	if hooksRaw, ok := settings["hooks"]; ok {
		if hooks, ok := hooksRaw.(map[string]any); ok {
			for event, entriesRaw := range hooks {
				entries, ok := entriesRaw.([]any)
				if !ok {
					continue
				}
				filtered := make([]any, 0, len(entries))
				for _, e := range entries {
					s, ok := e.(string)
					if !ok || !strings.HasPrefix(s, "oraculo") {
						filtered = append(filtered, e)
					}
				}
				hooks[event] = filtered
			}
			settings["hooks"] = hooks
		}
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings.json: %w", err)
	}
	if err := os.WriteFile(settingsPath, out, 0o644); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	return nil
}

// removeSkills removes all .claude/skills/oraculo-*/ directories installed by
// Oraculo, and also the legacy .claude/skills/oraculo/ directory if it exists.
func removeSkills(w interface{ Write([]byte) (int, error) }) error {
	matches, _ := filepath.Glob(filepath.Join(".claude", "skills", "oraculo-*"))
	for _, m := range matches {
		if err := os.RemoveAll(m); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove skill %s: %w", m, err)
		}
	}
	// Also remove legacy .claude/skills/oraculo/ if it exists.
	legacy := filepath.Join(".claude", "skills", "oraculo")
	if err := os.RemoveAll(legacy); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove legacy skills: %w", err)
	}
	return nil
}

// removeOraculoDir removes the .oraculo/ directory entirely.
func removeOraculoDir(w interface{ Write([]byte) (int, error) }) error {
	if err := os.RemoveAll(".oraculo"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove .oraculo/: %w", err)
	}
	return nil
}
