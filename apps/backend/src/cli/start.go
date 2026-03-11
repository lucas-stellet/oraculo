package cli

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/lucas/oraculo/apps/backend/src/applog"
	"github.com/lucas/oraculo/apps/backend/src/approval"
	"github.com/lucas/oraculo/apps/backend/src/config"
	"github.com/lucas/oraculo/apps/backend/src/db"
	mcpserver "github.com/lucas/oraculo/apps/backend/src/mcp"
	"github.com/lucas/oraculo/apps/backend/src/registry"
	"github.com/lucas/oraculo/apps/backend/src/server"
	"github.com/lucas/oraculo/apps/backend/src/ws"
)

const defaultIdleTimeout = 15 * time.Minute

// openLogFile opens .oraculo/server.log in the working directory (append mode).
// Falls back to stderr if the file cannot be opened.
func openLogFile() io.Writer {
	wd, err := os.Getwd()
	if err != nil {
		return os.Stderr
	}
	logPath := filepath.Join(wd, ".oraculo", "server.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return os.Stderr
	}
	return io.MultiWriter(f, os.Stderr)
}

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

	srv := server.New(database, bridge, hub, broadcaster, cfg.ProjectName(), version)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

	// Register in global server registry.
	regPath, regErr := registry.DefaultPath()
	if regErr == nil {
		wd, _ := os.Getwd()
		_ = registry.Register(regPath, registry.Entry{
			Project: cfg.ProjectName(),
			Path:    wd,
			Port:    port,
			PID:     os.Getpid(),
		})
		defer registry.Unregister(regPath, wd)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		logger.Info("server.started", "port", port)
		return srv.ListenAndServe(ctx, port, 0)
	})
	g.Go(func() error { return mcpSrv.RunStdio(ctx) })

	err = g.Wait()
	logger.Info("server.stopping")
	return err
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

	logOut := openLogFile()
	broadcaster := applog.NewBroadcaster(logOut)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)

	srv := server.New(database, bridge, hub, broadcaster, cfg.ProjectName(), version)

	// Register in global server registry.
	regPath, regErr := registry.DefaultPath()
	if regErr == nil {
		wd, _ := os.Getwd()
		_ = registry.Register(regPath, registry.Entry{
			Project: cfg.ProjectName(),
			Path:    wd,
			Port:    port,
			PID:     os.Getpid(),
		})
		defer registry.Unregister(regPath, wd)
	}

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
