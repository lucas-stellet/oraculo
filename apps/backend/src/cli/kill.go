package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/spf13/cobra"
)

func newKillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill",
		Short: "Stop the running Oraculo HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			port := cfg.Port
			if port == 0 {
				port = 3100
			}
			return killServer(cmd.OutOrStdout(), port)
		},
	}
}

// killServer finds and stops the process listening on port.
// Returns nil if no process is found or after it is stopped.
func killServer(w io.Writer, port int) error {
	pid, err := pidOnPort(port)
	if err != nil {
		return fmt.Errorf("find process on port %d: %w", port, err)
	}
	if pid == 0 {
		fmt.Fprintf(w, "No process found on port %d.\n", port)
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	// Verify the process is actually running before touching it.
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		fmt.Fprintf(w, "Process %d is not running.\n", pid)
		return nil
	}

	fmt.Fprintf(w, "Stopping Oraculo (pid %d, port %d)...\n", pid, port)

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to %d: %w", pid, err)
	}

	// Wait up to 5 s for graceful shutdown.
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			fmt.Fprintln(w, "Oraculo stopped.")
			return nil
		}
	}

	// Force kill if still alive.
	_ = proc.Signal(syscall.SIGKILL)
	fmt.Fprintln(w, "Oraculo force-stopped.")
	return nil
}

// pidOnPort returns the PID of the process listening on the given TCP port,
// or 0 if no process is found.
func pidOnPort(port int) (int, error) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil {
		// lsof exits 1 when no process matches — not an error for us.
		return 0, nil
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if line == "" {
		return 0, nil
	}
	pid, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("parse pid %q: %w", line, err)
	}
	return pid, nil
}
