// src/cli/tools/session.go
package tools

import (
	"encoding/json"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/output"
	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage skill sessions",
	}
	cmd.AddCommand(
		newSessionInitCmd(),
		newSessionStatusCmd(),
		newSessionStateCmd(),
		newSessionCloseCmd(),
	)
	return cmd
}

func newSessionInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Start a new session",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionType, _ := cmd.Flags().GetString("type")
			epicName, _ := cmd.Flags().GetString("epic")
			description, _ := cmd.Flags().GetString("description")

			epicStore := db.NewEpicStore(database)
			epic, _, err := epicStore.Create(epicName, description)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			sessionStore := db.NewSessionStore(database)
			session, created, err := sessionStore.Create(domain.SessionType(sessionType), &epic.ID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			return output.WriteJSON(w, map[string]any{
				"id":      session.ID,
				"type":    session.Type,
				"epic":    epicName,
				"status":  session.Status,
				"created": created,
			})
		},
	}
	cmd.Flags().String("type", "", "Session type (epic, story, plan, execute, validate)")
	cmd.Flags().String("epic", "", "Epic name")
	cmd.Flags().String("description", "", "Epic description")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("epic")
	return cmd
}

func newSessionStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check for an active session",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionType, _ := cmd.Flags().GetString("type")
			epicName, _ := cmd.Flags().GetString("epic")

			epicStore := db.NewEpicStore(database)
			epic, err := epicStore.GetByName(epicName)
			if err != nil {
				return output.WriteJSON(w, map[string]any{"active": false})
			}

			sessionStore := db.NewSessionStore(database)
			session, err := sessionStore.ActiveByEpic(epic.ID, domain.SessionType(sessionType))
			if err != nil {
				return output.WriteJSON(w, map[string]any{"active": false})
			}

			currentPhase, _ := sessionStore.CurrentPhase(session.ID)
			phases, _ := sessionStore.Phases(session.ID)
			completedNames := make([]string, len(phases))
			for i, p := range phases {
				completedNames[i] = p.Name
			}

			return output.WriteJSON(w, map[string]any{
				"active":           true,
				"id":               session.ID,
				"type":             session.Type,
				"epic":             epicName,
				"current_phase":    currentPhase,
				"completed_phases": completedNames,
			})
		},
	}
	cmd.Flags().String("type", "", "Session type")
	cmd.Flags().String("epic", "", "Epic name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("epic")
	return cmd
}

func newSessionStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Get full session state with phase data",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionID, _ := cmd.Flags().GetString("session")
			sessionStore := db.NewSessionStore(database)

			session, err := sessionStore.Get(sessionID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			currentPhase, _ := sessionStore.CurrentPhase(session.ID)
			phases, _ := sessionStore.Phases(session.ID)

			phaseData := make(map[string]any, len(phases))
			for _, p := range phases {
				var parsed any
				if err := json.Unmarshal([]byte(p.Data), &parsed); err != nil {
					phaseData[p.Name] = p.Data // raw string fallback
				} else {
					phaseData[p.Name] = parsed
				}
			}

			return output.WriteJSON(w, map[string]any{
				"id":            session.ID,
				"type":          session.Type,
				"status":        session.Status,
				"current_phase": currentPhase,
				"phases":        phaseData,
			})
		},
	}
	cmd.Flags().String("session", "", "Session ID")
	cmd.MarkFlagRequired("session")
	return cmd
}

func newSessionCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionID, _ := cmd.Flags().GetString("session")
			reason, _ := cmd.Flags().GetString("reason")

			status := domain.SessionCompleted
			if reason == "abandoned" {
				status = domain.SessionAbandoned
			}

			sessionStore := db.NewSessionStore(database)
			if err := sessionStore.Close(sessionID, status); err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"id":     sessionID,
				"status": status,
				"closed": true,
			})
		},
	}
	cmd.Flags().String("session", "", "Session ID")
	cmd.Flags().String("reason", "", "Close reason (abandoned)")
	cmd.MarkFlagRequired("session")
	return cmd
}
