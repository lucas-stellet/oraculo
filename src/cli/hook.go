// src/cli/hook.go
package cli

import "github.com/spf13/cobra"

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Commands triggered by Claude Code hooks",
	}
}
