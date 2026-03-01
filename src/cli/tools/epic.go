// src/cli/tools/epic.go
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

// newEpicCmd returns the "epic" parent command with all subcommands.
func newEpicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epic",
		Short: "Manage epics",
	}
	cmd.AddCommand(
		newEpicInitCmd(),
		newEpicSaveCmd(),
		newEpicGetCmd(),
		newEpicListCmd(),
		newEpicUpdateCmd(),
		newEpicDeleteCmd(),
	)
	return cmd
}

func newEpicInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")
			w := cmd.OutOrStdout()

			store := db.NewEpicStore(dbFromContext(cmd.Context()))
			_, created, err := store.Create(name, description)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"name":    name,
				"path":    filepath.Join(".oraculo", "epics", name),
				"created": created,
			})
		},
	}
	cmd.Flags().String("description", "", "Epic description")
	return cmd
}

func newEpicSaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save <name>",
		Short: "Save requirements markdown for an epic (reads stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			content, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				output.WriteError(w, fmt.Errorf("read stdin: %w", err))
				return err
			}

			dir := filepath.Join(".oraculo", "epics", name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				output.WriteError(w, fmt.Errorf("create directory: %w", err))
				return err
			}

			path := filepath.Join(dir, "requirements.md")
			if err := os.WriteFile(path, content, 0o644); err != nil {
				output.WriteError(w, fmt.Errorf("write file: %w", err))
				return err
			}

			return output.WriteJSON(w, map[string]any{
				"name":  name,
				"path":  path,
				"saved": true,
			})
		},
	}
}

func newEpicGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get requirements markdown for an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			path := filepath.Join(".oraculo", "epics", name, "requirements.md")
			content, err := os.ReadFile(path)
			if err != nil {
				wrappedErr := fmt.Errorf("epic %q requirements: %w", name, err)
				output.WriteError(w, wrappedErr)
				return wrappedErr
			}

			_, err = w.Write(content)
			return err
		},
	}
}

func newEpicListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all epics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			store := db.NewEpicStore(dbFromContext(cmd.Context()))
			epics, err := store.List()
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, epics)
		},
	}
}

func newEpicUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update an epic's description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")
			w := cmd.OutOrStdout()

			store := db.NewEpicStore(dbFromContext(cmd.Context()))
			epic, err := store.Update(name, description)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, epic)
		},
	}
	cmd.Flags().String("description", "", "New epic description")
	return cmd
}

func newEpicDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			store := db.NewEpicStore(dbFromContext(cmd.Context()))
			if err := store.Delete(name); err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"name":    name,
				"deleted": true,
			})
		},
	}
}
