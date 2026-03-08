// src/cli/tools/validation.go
package tools

import (
	"fmt"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

// newValidationCmd returns the "validation" parent command with subcommands.
func newValidationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validation",
		Short: "Manage QA validations",
	}
	cmd.AddCommand(newValidationSaveCmd())
	return cmd
}

func newValidationSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <story-name>",
		Short: "Save a QA validation verdict for a story",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			w := cmd.OutOrStdout()

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			verdictStr, _ := cmd.Flags().GetString("verdict")
			if verdictStr != "approved" && verdictStr != "rejected" {
				err := fmt.Errorf("invalid verdict %q: must be approved or rejected", verdictStr)
				output.WriteError(w, err)
				return err
			}

			findings, _ := cmd.Flags().GetString("findings")

			storyStore := db.NewStoryStore(database)
			story, err := storyStore.GetByName(epicID, name)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			valStore := db.NewValidationStore(database)
			validation, err := valStore.Save(story.ID, verdictStr, findings)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, validation)
		},
	}
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("verdict", "", "Validation verdict: approved|rejected (required)")
	_ = cmd.MarkFlagRequired("verdict")
	cmd.Flags().String("findings", "", "QA findings as JSON string")
	return cmd
}
