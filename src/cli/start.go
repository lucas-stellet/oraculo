package cli

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Oraculo (MCP + HTTP + WebSocket)",
		Long:  "Starts the MCP server on stdio and the HTTP/WebSocket server on the configured port. Launched by Claude Code as an MCP server.",
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, _ []string) error {
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
	srv := server.New(database, bridge, hub, broadcaster)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

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
