package cli

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/spf13/cobra"
)

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the Oraculo HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()

			cfg, err := config.Read()
			if err != nil {
				return err
			}
			port := cfg.Port
			if port == 0 {
				port = 3100
			}

			// Kill the running instance (no-op if none).
			if err := killServer(w, port); err != nil {
				return err
			}

			// Wait for the port to be free (up to 5 s).
			for range 50 {
				if !portInUse(port) {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			if portInUse(port) {
				return fmt.Errorf("port %d still in use after shutdown", port)
			}

			// Re-exec the current binary as "start http" (detached).
			bin, err := os.Executable()
			if err != nil {
				return fmt.Errorf("locate binary: %w", err)
			}

			child, err := os.StartProcess(bin, []string{bin, "start", "http"},
				&os.ProcAttr{
					Files: []*os.File{nil, nil, nil},
					Sys:   &syscall.SysProcAttr{Setsid: true},
					Env:   os.Environ(),
				},
			)
			if err != nil {
				return fmt.Errorf("start Oraculo: %w", err)
			}
			// Detach: let the child live after we exit.
			child.Release()

			fmt.Fprintf(w, "Oraculo started (port %d).\n", port)
			return nil
		},
	}
}

