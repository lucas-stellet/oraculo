// src/cli/root.go
package cli

import (
	"github.com/lucas/oraculo/src/cli/tools"
	"github.com/spf13/cobra"
)

// NewRoot returns the root cobra.Command for oraculo.
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "oraculo",
		Short: "Socratic guide for quality product development",
		Long:  "Oraculo is a Socratic guide and team orchestrator for quality product development.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newVersionCmd(version),
		newInstallCmd(),
		tools.NewToolsCmd(),
	)
	return root
}
