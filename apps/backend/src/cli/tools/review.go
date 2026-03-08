// src/cli/tools/review.go
package tools

import (
	"fmt"
	"strconv"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

// newReviewCmd returns the "review" parent command with all subcommands.
func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Manage document reviews",
	}
	cmd.AddCommand(
		newReviewCreateCmd(),
		newReviewGetCmd(),
		newReviewListCmd(),
	)
	return cmd
}

func newReviewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <version-id>",
		Short: "Create a review for a document version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			versionID, err := strconv.Atoi(args[0])
			if err != nil {
				err = fmt.Errorf("invalid version-id %q: %w", args[0], err)
				output.WriteError(w, err)
				return err
			}

			typeName, _ := cmd.Flags().GetString("type")
			versionType := domain.VersionType(typeName)
			if versionType != domain.VersionTypeEpic && versionType != domain.VersionTypeStory {
				err := fmt.Errorf("invalid type %q: must be %q or %q", typeName, domain.VersionTypeEpic, domain.VersionTypeStory)
				output.WriteError(w, err)
				return err
			}

			verdictName, _ := cmd.Flags().GetString("verdict")
			verdict := domain.ReviewVerdict(verdictName)
			if !verdict.Valid() {
				err := fmt.Errorf("invalid verdict %q: must be %q or %q", verdictName, domain.ReviewApproved, domain.ReviewRejected)
				output.WriteError(w, err)
				return err
			}

			comment, _ := cmd.Flags().GetString("comment")

			store := db.NewReviewStore(dbFromContext(cmd.Context()))
			review, err := store.Create(versionID, versionType, verdict, comment)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, review)
		},
	}
	cmd.Flags().String("type", "", "Version type (epic or story)")
	_ = cmd.MarkFlagRequired("type")
	cmd.Flags().String("verdict", "", "Review verdict (approved or rejected)")
	_ = cmd.MarkFlagRequired("verdict")
	cmd.Flags().String("comment", "", "Optional review comment")
	return cmd
}

func newReviewGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <review-id>",
		Short: "Get a review by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			id, err := strconv.Atoi(args[0])
			if err != nil {
				err = fmt.Errorf("invalid review-id %q: %w", args[0], err)
				output.WriteError(w, err)
				return err
			}

			store := db.NewReviewStore(dbFromContext(cmd.Context()))
			review, err := store.GetByID(id)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, review)
		},
	}
}

func newReviewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <version-id>",
		Short: "List reviews for a document version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			versionID, err := strconv.Atoi(args[0])
			if err != nil {
				err = fmt.Errorf("invalid version-id %q: %w", args[0], err)
				output.WriteError(w, err)
				return err
			}

			typeName, _ := cmd.Flags().GetString("type")
			versionType := domain.VersionType(typeName)
			if versionType != domain.VersionTypeEpic && versionType != domain.VersionTypeStory {
				err := fmt.Errorf("invalid type %q: must be %q or %q", typeName, domain.VersionTypeEpic, domain.VersionTypeStory)
				output.WriteError(w, err)
				return err
			}

			store := db.NewReviewStore(dbFromContext(cmd.Context()))
			reviews, err := store.ListByVersion(versionID, versionType)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, reviews)
		},
	}
	cmd.Flags().String("type", "", "Version type (epic or story)")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}
