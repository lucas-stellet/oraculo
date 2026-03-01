// src/cli/install.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Oraculo skills and hooks into Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			scope := "local (.claude/)"
			if global {
				scope = "global (~/.claude/)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "oraculo install: %s (not yet implemented)\n", scope)
			return nil
		},
	}
	cmd.Flags().Bool("global", false, "Install globally for all projects")
	cmd.Flags().Bool("local", false, "Install locally for current project")
	return cmd
}
