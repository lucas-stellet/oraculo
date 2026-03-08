// src/cli/tools/memory.go
package tools

import (
	"fmt"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

// newMemoryCmd returns the "memory" parent command with store, search, and domains subcommands.
func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage codebase knowledge memory",
	}
	cmd.AddCommand(
		newMemoryStoreCmd(),
		newMemorySearchCmd(),
		newMemoryDomainsCmd(),
	)
	return cmd
}

func newMemoryStoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Store a knowledge finding",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			domainName, _ := cmd.Flags().GetString("domain")
			categoryStr, _ := cmd.Flags().GetString("category")
			finding, _ := cmd.Flags().GetString("finding")
			source, _ := cmd.Flags().GetString("source")
			confidenceStr, _ := cmd.Flags().GetString("confidence")

			category := domain.Category(categoryStr)
			if !category.Valid() {
				err := fmt.Errorf("invalid category %q: must be one of pattern, convention, constraint, dependency, test, architecture", categoryStr)
				output.WriteError(w, err)
				return err
			}

			confidence := domain.Confidence(confidenceStr)
			if !confidence.Valid() {
				err := fmt.Errorf("invalid confidence %q: must be one of high, medium, low", confidenceStr)
				output.WriteError(w, err)
				return err
			}

			store := db.NewMemoryStore(dbFromContext(cmd.Context()))
			knowledge, err := store.Store(domainName, category, finding, source, confidence)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, knowledge)
		},
	}
	cmd.Flags().String("domain", "", "Knowledge domain (required)")
	_ = cmd.MarkFlagRequired("domain")
	cmd.Flags().String("category", "", "Finding category: pattern|convention|constraint|dependency|test|architecture (required)")
	_ = cmd.MarkFlagRequired("category")
	cmd.Flags().String("finding", "", "The knowledge finding text (required)")
	_ = cmd.MarkFlagRequired("finding")
	cmd.Flags().String("source", "", "Source file(s) related to the finding")
	cmd.Flags().String("confidence", "medium", "Confidence level: high|medium|low")
	return cmd
}

func newMemorySearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search knowledge findings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			w := cmd.OutOrStdout()

			domainFilter, _ := cmd.Flags().GetString("domain")
			limit, _ := cmd.Flags().GetInt("limit")

			store := db.NewMemoryStore(dbFromContext(cmd.Context()))
			results, err := store.Search(query, domainFilter, limit)
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, results)
		},
	}
	cmd.Flags().String("domain", "", "Filter results by domain")
	cmd.Flags().Int("limit", 0, "Maximum number of results (0 = unlimited)")
	return cmd
}

func newMemoryDomainsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "domains",
		Short: "List all knowledge domains",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			store := db.NewMemoryStore(dbFromContext(cmd.Context()))
			domains, err := store.Domains()
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, domains)
		},
	}
}
