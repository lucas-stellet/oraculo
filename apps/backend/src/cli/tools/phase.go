// src/cli/tools/phase.go
package tools

import (
	"io"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

func newPhaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage session phases",
	}
	cmd.AddCommand(newPhaseCompleteCmd())
	return cmd
}

func newPhaseCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <phase>",
		Short: "Complete a phase (reads JSON data from stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())
			sessionID, _ := cmd.Flags().GetString("session")

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			dataStr := string(data)
			if dataStr == "" {
				dataStr = "{}"
			}

			sessionStore := db.NewSessionStore(database)
			if err := sessionStore.CompletePhase(sessionID, phase, dataStr); err != nil {
				output.WriteError(w, err)
				return err
			}

			// Determine next phase
			session, _ := sessionStore.Get(sessionID)
			next, _ := sessionStore.CurrentPhase(sessionID)
			sessionClosed := session.Status == domain.SessionCompleted

			result := map[string]any{
				"phase":     phase,
				"completed": true,
				"next":      next,
			}
			if sessionClosed {
				result["session_closed"] = true
			}
			return output.WriteJSON(w, result)
		},
	}
	cmd.Flags().String("session", "", "Session ID")
	cmd.MarkFlagRequired("session")
	return cmd
}
