package cli

import (
	"fmt"
	"io"
	"net"
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

// killServer finds and stops all processes listening on port.
// Waits until the port is confirmed free before returning.
func killServer(w io.Writer, port int) error {
	pids, err := pidsOnPort(port)
	if err != nil {
		return fmt.Errorf("find processes on port %d: %w", port, err)
	}
	if len(pids) == 0 {
		fmt.Fprintf(w, "No process found on port %d.\n", port)
		return nil
	}

	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			continue // already dead
		}
		fmt.Fprintf(w, "Stopping Oraculo (pid %d, port %d)...\n", pid, port)
		_ = proc.Signal(syscall.SIGTERM)
	}

	// Wait up to 5 s for the port to be released.
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		if !portInUse(port) {
			fmt.Fprintln(w, "Oraculo stopped.")
			return nil
		}
	}

	// Force-kill any survivors and wait a bit more.
	for _, pid := range pids {
		if proc, err := os.FindProcess(pid); err == nil {
			_ = proc.Signal(syscall.SIGKILL)
		}
	}
	time.Sleep(300 * time.Millisecond)
	fmt.Fprintln(w, "Oraculo force-stopped.")
	return nil
}

// pidsOnPort returns all PIDs listening on the given TCP port.
func pidsOnPort(port int) ([]int, error) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil {
		// lsof exits 1 when no process matches — not an error for us.
		return nil, nil
	}
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func portInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}
