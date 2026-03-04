// src/cli/tools/design.go
package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/output"
	"github.com/spf13/cobra"
)

// newDesignCmd returns the "design" parent command with all subcommands.
func newDesignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "design",
		Short: "Manage design artifacts",
	}
	cmd.AddCommand(
		newDesignSaveCmd(),
		newDesignGetCmd(),
	)
	return cmd
}

func newDesignSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <story-name>",
		Short: "Save design markdown for a story (reads stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, epicName, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			// Verify the story exists.
			store := db.NewStoryStore(database)
			if _, err := store.GetByName(epicID, name); err != nil {
				output.WriteError(w, err)
				return err
			}

			content, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				output.WriteError(w, fmt.Errorf("read stdin: %w", err))
				return err
			}

			dir := filepath.Join(".oraculo", "epics", epicName, "stories", name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				output.WriteError(w, fmt.Errorf("create directory: %w", err))
				return err
			}

			path := filepath.Join(dir, "design.md")
			if err := os.WriteFile(path, content, 0o644); err != nil {
				output.WriteError(w, fmt.Errorf("write file: %w", err))
				return err
			}

			return output.WriteJSON(w, map[string]any{
				"name":  name,
				"epic":  epicName,
				"path":  path,
				"saved": true,
			})
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	return cmd
}

func newDesignGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <story-name>",
		Short: "Get design markdown for a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			_, epicName, _, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			path := filepath.Join(".oraculo", "epics", epicName, "stories", name, "design.md")
			content, err := os.ReadFile(path)
			if err != nil {
				wrappedErr := fmt.Errorf("story %q design: %w", name, err)
				output.WriteError(w, wrappedErr)
				return wrappedErr
			}

			_, err = w.Write(content)
			return err
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	return cmd
}
