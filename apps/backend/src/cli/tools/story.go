// apps/backend/src/cli/tools/story.go
package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

// newStoryCmd returns the "story" parent command with all subcommands.
func newStoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Manage stories",
	}
	cmd.AddCommand(
		newStoryInitCmd(),
		newStorySaveCmd(),
		newStoryGetCmd(),
		newStoryListCmd(),
		newStoryUpdateCmd(),
		newStoryUpdateStatusCmd(),
		newStoryDeleteCmd(),
		newStoryVersionCmd(),
		newStoryVersionsCmd(),
	)
	return cmd
}

// resolveEpic looks up an epic by name and returns it. On error it writes
// the JSON error to w and returns nil plus the error.
func resolveEpic(cmd *cobra.Command) (*db.DB, string, int, error) {
	w := cmd.OutOrStdout()
	epicName, _ := cmd.Flags().GetString("epic")
	database := dbFromContext(cmd.Context())
	epicStore := db.NewEpicStore(database)
	epic, err := epicStore.GetByName(epicName)
	if err != nil {
		output.WriteError(w, err)
		return nil, "", 0, err
	}
	return database, epicName, epic.ID, nil
}

func newStoryInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new story within an epic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")
			w := cmd.OutOrStdout()

			database, epicName, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			store := db.NewStoryStore(database)
			_, created, err := store.Create(epicID, name, description)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"name":    name,
				"epic":    epicName,
				"path":    filepath.Join(".oraculo", "epics", epicName, "stories", name),
				"created": created,
			})
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("description", "", "Story description")
	return cmd
}

func newStorySaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Save requirements markdown for a story (reads stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, epicName, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			// Upsert DB record so downstream commands (e.g. approval request) can find the story.
			store := db.NewStoryStore(database)
			if _, _, err := store.Create(epicID, name, ""); err != nil {
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

			path := filepath.Join(dir, "requirements.md")
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

func newStoryGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get requirements markdown for a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			_, epicName, _, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			path := filepath.Join(".oraculo", "epics", epicName, "stories", name, "requirements.md")
			content, err := os.ReadFile(path)
			if err != nil {
				wrappedErr := fmt.Errorf("story %q requirements: %w", name, err)
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

func newStoryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all stories for an epic",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			store := db.NewStoryStore(database)
			stories, err := store.List(epicID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, stories)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	return cmd
}

func newStoryUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a story's description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")
			w := cmd.OutOrStdout()

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			store := db.NewStoryStore(database)
			story, err := store.Update(epicID, name, description)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, story)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("description", "", "New story description")
	return cmd
}

func newStoryUpdateStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status <name>",
		Short: "Update a story's approval status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			statusStr, _ := cmd.Flags().GetString("status")
			w := cmd.OutOrStdout()

			status := domain.ApprovalStatus(statusStr)
			if !status.Valid() {
				err := fmt.Errorf("invalid status %q: must be one of none, pending, approved, rejected, needs_revision", statusStr)
				output.WriteError(w, err)
				return err
			}

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			store := db.NewStoryStore(database)
			if err := store.UpdateApprovalStatus(epicID, name, status); err != nil {
				output.WriteError(w, err)
				return err
			}

			story, err := store.GetByName(epicID, name)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, story)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("status", "", "Approval status: none|pending|approved|rejected|needs_revision (required)")
	_ = cmd.MarkFlagRequired("status")
	return cmd
}

func newStoryVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version <name>",
		Short: "Create a new version of story definition (reads stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			storyStore := db.NewStoryStore(database)
			story, err := storyStore.GetByName(epicID, name)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			content, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				output.WriteError(w, fmt.Errorf("read stdin: %w", err))
				return err
			}

			versionStore := db.NewVersionStore(database)
			version, err := versionStore.CreateStoryVersion(story.ID, string(content))
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			approvalStore := db.NewApprovalStore(database)
			approval, err := approvalStore.Request(domain.ApprovalDesign, &epicID, &story.ID, string(content))
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			return output.WriteJSON(w, map[string]any{
				"version_id":  version.ID,
				"approval_id": approval.ID,
			})
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	return cmd
}

func newStoryVersionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions <name>",
		Short: "List all versions of a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			storyStore := db.NewStoryStore(database)
			story, err := storyStore.GetByName(epicID, name)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			versionStore := db.NewVersionStore(database)
			versions, err := versionStore.ListStoryVersions(story.ID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, versions)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	return cmd
}

func newStoryDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, epicName, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			store := db.NewStoryStore(database)
			if err := store.Delete(epicID, name); err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"name":    name,
				"epic":    epicName,
				"deleted": true,
			})
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	return cmd
}
