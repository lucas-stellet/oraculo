// src/cli/tools/tools.go
package tools

import (
	"github.com/lucas/oraculo/src/db"
	"github.com/spf13/cobra"
)

// NewToolsCmd returns the "tools" parent command.
func NewToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Agent-facing commands (JSON output)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.Open()
			if err != nil {
				return err
			}
			cmd.SetContext(withDB(cmd.Context(), database))
			return nil
		},
	}
	cmd.AddCommand(newEpicCmd(), newStoryCmd(), newTaskCmd(), newMemoryCmd(), newApprovalCmd(), newSessionCmd(), newPhaseCmd(), newDesignCmd(), newReviewCmd())
	return cmd
}
