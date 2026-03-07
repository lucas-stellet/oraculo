package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/lucas/oraculo/src/applog"
	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/config"
	"github.com/lucas/oraculo/src/db"
	mcpserver "github.com/lucas/oraculo/src/mcp"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

const defaultIdleTimeout = 15 * time.Minute

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Oraculo services",
		Long:  "Start Oraculo services. Without subcommand, starts both MCP and HTTP servers.",
		RunE:  runStartAll,
	}
	cmd.AddCommand(newStartMCPCmd())
	cmd.AddCommand(newStartHTTPCmd())
	return cmd
}

func newStartMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server on stdio (managed by Claude Code)",
		RunE:  runStartMCP,
	}
}

func newStartHTTPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "http",
		Short: "Start HTTP + WebSocket server as daemon",
		RunE:  runStartHTTP,
	}
}

// runStartAll starts both MCP and HTTP servers (backwards compatible).
func runStartAll(cmd *cobra.Command, _ []string) error {
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

	// Resolve dashboard path relative to executable
	staticPath := dashboardStaticPath()
	logger.Info("server.config", "port", port, "static_path", staticPath)
	srv := server.New(database, bridge, hub, broadcaster, staticPath)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		logger.Info("server.started", "port", port)
		return srv.ListenAndServe(ctx, port, 0)
	})

	// Open browser after server starts
	go func() {
		time.Sleep(500 * time.Millisecond)
		url := fmt.Sprintf("http://localhost:%d", port)
		if err := openBrowser(url); err != nil {
			logger.Error("browser.open_failed", "url", url, "error", err)
		} else {
			logger.Info("browser.opened", "url", url)
		}
	}()

	g.Go(func() error { return mcpSrv.RunStdio(ctx) })

	err = g.Wait()
	logger.Info("server.stopping")
	return err
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

// dashboardStaticPath returns the absolute path to the dashboard static files.
// It resolves the path relative to the executable's directory.
func dashboardStaticPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "./apps/dashboard/out"
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "apps", "dashboard", "out")
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

// runStartHTTP starts the HTTP + WebSocket server as a daemon with idle timeout.
func runStartHTTP(cmd *cobra.Command, _ []string) error {
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

	staticPath := dashboardStaticPath()
	srv := server.New(database, bridge, hub, broadcaster, staticPath)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		logger.Info("server.started", "port", port, "idle_timeout", defaultIdleTimeout)
		return srv.ListenAndServe(ctx, port, defaultIdleTimeout)
	})

	err = g.Wait()
	logger.Info("server.stopping")
	return err
}
