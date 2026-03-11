package cli

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/lucas/oraculo/apps/backend/src/applog"
	"github.com/lucas/oraculo/apps/backend/src/approval"
	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/lucas/oraculo/apps/backend/src/db"
	mcpserver "github.com/lucas/oraculo/apps/backend/src/mcp"
	"github.com/lucas/oraculo/apps/backend/src/server"
	"github.com/lucas/oraculo/apps/backend/src/ws"
)

const defaultIdleTimeout = 15 * time.Minute

func newStartCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Oraculo services",
		Long:  "Start Oraculo services. Without subcommand, starts both MCP and HTTP servers.",
		RunE:  makeStartAll(version),
	}
	cmd.AddCommand(newStartMCPCmd())
	cmd.AddCommand(newStartHTTPCmd(version))
	return cmd
}

func newStartMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server on stdio (managed by Claude Code)",
		RunE:  runStartMCP,
	}
}

func newStartHTTPCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "http",
		Short: "Start HTTP + WebSocket server as daemon",
		RunE:  makeStartHTTP(version),
	}
}

// makeStartAll returns a RunE that starts both MCP and HTTP servers.
func makeStartAll(version string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runStartAll(cmd, version)
	}
}

func runStartAll(cmd *cobra.Command, version string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	cfg, err := config.Read()
	if err != nil {
		return err
	}

	port := cfg.Port
	if port == 0 {
		port = 3100
	}

	broadcaster := applog.NewBroadcaster(os.Stderr)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)

	// Dashboard is embedded in the binary, pass empty string
	srv := server.New(database, bridge, hub, broadcaster, "", version)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })

	if isPortInUse(port) {
		logger.Info("server.http_already_running", "port", port)
	} else {
		go func() {
			logger.Info("server.started", "port", port)
			if err := srv.ListenAndServe(ctx, port, 0); err != nil {
				logger.Warn("server.http_unavailable", "port", port, "error", err)
			}
		}()

		if os.Getenv("ORACULO_NO_BROWSER") == "" {
			go func() {
				time.Sleep(500 * time.Millisecond)
				url := fmt.Sprintf("http://localhost:%d", port)
				if err := openBrowser(url); err != nil {
					logger.Error("browser.open_failed", "url", url, "error", err)
				} else {
					logger.Info("browser.opened", "url", url)
				}
			}()
		}
	}

	g.Go(func() error { return mcpSrv.RunStdio(ctx) })

	err = g.Wait()
	logger.Info("server.stopping")
	return err
}

// isPortInUse reports whether something is already listening on the given TCP port.
func isPortInUse(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// openBrowser opens the given URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// runStartMCP starts only the MCP server on stdio.
func runStartMCP(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	broadcaster := applog.NewBroadcaster(os.Stderr)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

	return mcpSrv.RunStdio(ctx)
}

// makeStartHTTP returns a RunE that starts the HTTP + WebSocket server.
func makeStartHTTP(version string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runStartHTTP(cmd, version)
	}
}

// runStartHTTP starts the HTTP + WebSocket server as a daemon with idle timeout.
func runStartHTTP(cmd *cobra.Command, version string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	cfg, err := config.Read()
	if err != nil {
		return err
	}

	port := cfg.Port
	if port == 0 {
		port = 3100
	}

	broadcaster := applog.NewBroadcaster(os.Stderr)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)

	// Dashboard is embedded in the binary, pass empty string
	srv := server.New(database, bridge, hub, broadcaster, "", version)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		logger.Info("server.started", "port", port, "idle_timeout", defaultIdleTimeout)
		return srv.ListenAndServe(ctx, port, defaultIdleTimeout)
	})

	if os.Getenv("ORACULO_NO_BROWSER") == "" {
		go func() {
			time.Sleep(500 * time.Millisecond)
			url := fmt.Sprintf("http://localhost:%d", port)
			if err := openBrowser(url); err != nil {
				logger.Error("browser.open_failed", "url", url, "error", err)
			} else {
				logger.Info("browser.opened", "url", url)
			}
		}()
	}

	err = g.Wait()
	logger.Info("server.stopping")
	return err
}
