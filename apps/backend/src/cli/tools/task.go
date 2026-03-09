// apps/backend/src/cli/tools/task.go
package tools

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

// newTaskCmd returns the "task" parent command with all subcommands.
func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks within a story",
	}
	cmd.AddCommand(
		newTaskInitCmd(),
		newTaskStartCmd(),
		newTaskCompleteCmd(),
		newTaskFailCmd(),
		newTaskGetCmd(),
		newTaskListCmd(),
		newTaskDeleteCmd(),
	)
	return cmd
}

// resolveStory resolves both epic and story by name, returning the DB, epic name,
// story name, story ID, and any error. On error, the JSON error is written to
// cmd's stdout so callers can simply return.
func resolveStory(cmd *cobra.Command) (*db.DB, string, string, int, error) {
	w := cmd.OutOrStdout()
	epicName, _ := cmd.Flags().GetString("epic")
	storyName, _ := cmd.Flags().GetString("story")
	database := dbFromContext(cmd.Context())

	epicStore := db.NewEpicStore(database)
	epic, err := epicStore.GetByName(epicName)
	if err != nil {
		output.WriteError(w, err)
		return nil, "", "", 0, err
	}

	storyStore := db.NewStoryStore(database)
	story, err := storyStore.GetByName(epic.ID, storyName)
	if err != nil {
		output.WriteError(w, err)
		return nil, "", "", 0, err
	}

	return database, epicName, storyName, story.ID, nil
}

func newTaskInitCmd() *cobra.Command {
	var dependsOn []string

	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new task within a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			description, _ := cmd.Flags().GetString("description")
			w := cmd.OutOrStdout()

			database, epicName, storyName, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			store := db.NewTaskStore(database)
			_, created, err := store.Create(storyID, name, description, dependsOn)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"name":    name,
				"epic":    epicName,
				"story":   storyName,
				"created": created,
			})
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	cmd.Flags().String("description", "", "Task description")
	cmd.Flags().StringArrayVar(&dependsOn, "depends-on", nil, "Task dependency (repeatable)")
	return cmd
}

func newTaskStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Transition a task from pending to in_progress",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, _, _, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			store := db.NewTaskStore(database)
			task, err := store.Start(storyID, name)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, task)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	return cmd
}

// taskCompleteInput mirrors the JSON structure expected on stdin for task complete.
type taskCompleteInput struct {
	Summary       string   `json:"summary"`
	Logs          string   `json:"logs"`
	SkillsUsed    []string `json:"skills_used"`
	FilesModified []string `json:"files_modified"`
}

func newTaskCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <name>",
		Short: "Transition a task from in_progress to completed (reads JSON from stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, _, _, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				wrappedErr := fmt.Errorf("read stdin: %w", err)
				output.WriteError(w, wrappedErr)
				return wrappedErr
			}

			var input taskCompleteInput
			if err := json.Unmarshal(data, &input); err != nil {
				wrappedErr := fmt.Errorf("parse result JSON: %w", err)
				output.WriteError(w, wrappedErr)
				return wrappedErr
			}

			result := domain.TaskResult{
				Summary:       input.Summary,
				Logs:          input.Logs,
				SkillsUsed:    input.SkillsUsed,
				FilesModified: input.FilesModified,
			}

			store := db.NewTaskStore(database)
			task, err := store.Complete(storyID, name, result)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, task)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	return cmd
}

func newTaskFailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fail <name>",
		Short: "Transition a task from in_progress to failed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			reason, _ := cmd.Flags().GetString("reason")
			w := cmd.OutOrStdout()

			database, _, _, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			store := db.NewTaskStore(database)
			task, err := store.Fail(storyID, name, reason)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, task)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	cmd.Flags().String("reason", "", "Failure reason (required)")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}

func newTaskGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get a task by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, _, _, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			store := db.NewTaskStore(database)
			task, err := store.GetByName(storyID, name)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, task)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	return cmd
}

func newTaskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks for a story",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			database, _, _, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			store := db.NewTaskStore(database)
			tasks, err := store.List(storyID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, tasks)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	return cmd
}

func newTaskDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, epicName, storyName, storyID, err := resolveStory(cmd)
			if err != nil {
				return err
			}

			store := db.NewTaskStore(database)
			if err := store.Delete(storyID, name); err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"name":    name,
				"epic":    epicName,
				"story":   storyName,
				"deleted": true,
			})
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Parent story name (required)")
	_ = cmd.MarkFlagRequired("story")
	return cmd
}
