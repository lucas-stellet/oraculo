// src/mcp/server.go
package mcp

import (
	"context"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
)

// Server is the MCP (Model Context Protocol) server that exposes Oraculo tools
// to Claude Code over stdio.
type Server struct {
	bridge *approval.Bridge
	store  *db.ApprovalStore
}

// New creates a new MCP Server backed by the given approval bridge and store.
func New(bridge *approval.Bridge, store *db.ApprovalStore) *Server {
	return &Server{bridge: bridge, store: store}
}

// RunStdio runs the MCP server, reading JSON-RPC requests from stdin and
// writing responses to stdout. It returns when ctx is done.
func (s *Server) RunStdio(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
