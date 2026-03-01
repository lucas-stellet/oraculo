// src/cli/tools/approval.go
package tools

import (
	"fmt"
	"io"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/output"
	"github.com/spf13/cobra"
)

// newApprovalCmd returns the "approval" parent command with all subcommands.
func newApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: "Manage approval gates",
	}
	cmd.AddCommand(
		newApprovalRequestCmd(),
		newApprovalStatusCmd(),
		newApprovalListCmd(),
		newApprovalVerdictCmd(),
	)
	return cmd
}

func newApprovalRequestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "Create a new approval request (reads content from stdin)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			approvalType, _ := cmd.Flags().GetString("type")
			at := domain.ApprovalType(approvalType)
			if !at.Valid() {
				err := fmt.Errorf("invalid approval type %q: %w", approvalType, domain.ErrMissingRequired)
				output.WriteError(w, err)
				return err
			}

			database, _, epicID, err := resolveEpic(cmd)
			if err != nil {
				return err
			}

			var storyIDPtr *int
			storyName, _ := cmd.Flags().GetString("story")
			if storyName != "" {
				storyStore := db.NewStoryStore(database)
				story, err := storyStore.GetByName(epicID, storyName)
				if err != nil {
					output.WriteError(w, err)
					return err
				}
				storyIDPtr = &story.ID
			}

			content, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				output.WriteError(w, fmt.Errorf("read stdin: %w", err))
				return err
			}

			epicIDPtr := &epicID
			store := db.NewApprovalStore(database)
			approval, err := store.Request(at, epicIDPtr, storyIDPtr, string(content))
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, approval)
		},
	}
	cmd.Flags().String("type", "", "Approval type (epic-requirements|story-definition|qa-escalation|execution-plan)")
	_ = cmd.MarkFlagRequired("type")
	cmd.Flags().String("epic", "", "Parent epic name (required)")
	_ = cmd.MarkFlagRequired("epic")
	cmd.Flags().String("story", "", "Story name (optional)")
	return cmd
}

func newApprovalStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <id>",
		Short: "Get the status of an approval request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			id := args[0]

			store := db.NewApprovalStore(dbFromContext(cmd.Context()))
			approval, err := store.GetByID(id)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, approval)
		},
	}
}

func newApprovalListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List approval requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			pendingOnly, _ := cmd.Flags().GetBool("pending")

			store := db.NewApprovalStore(dbFromContext(cmd.Context()))
			approvals, err := store.List(pendingOnly)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, approvals)
		},
	}
	cmd.Flags().Bool("pending", false, "Show only pending approvals")
	return cmd
}

func newApprovalVerdictCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verdict <id>",
		Short: "Record a verdict on an approval request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			id := args[0]

			verdictStr, _ := cmd.Flags().GetString("verdict")
			v := domain.Verdict(verdictStr)
			if !v.Valid() {
				err := fmt.Errorf("invalid verdict %q: %w", verdictStr, domain.ErrMissingRequired)
				output.WriteError(w, err)
				return err
			}

			comment, _ := cmd.Flags().GetString("comment")

			store := db.NewApprovalStore(dbFromContext(cmd.Context()))
			approval, err := store.Verdict(id, v, comment)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, approval)
		},
	}
	cmd.Flags().String("verdict", "", "Verdict (approved|rejected|needs_revision)")
	_ = cmd.MarkFlagRequired("verdict")
	cmd.Flags().String("comment", "", "Optional comment")
	return cmd
}
