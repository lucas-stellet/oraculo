// src/server/server.go
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/ws"
)

// Server is the HTTP/WebSocket server for the Oraculo UI and approval API.
type Server struct {
	database *db.DB
	bridge   *approval.Bridge
	hub      *ws.Hub
	mux      *http.ServeMux
}

// New creates a new Server wiring together the database, approval bridge, and
// WebSocket hub.
func New(database *db.DB, bridge *approval.Bridge, hub *ws.Hub) *Server {
	s := &Server{
		database: database,
		bridge:   bridge,
		hub:      hub,
		mux:      http.NewServeMux(),
	}
	s.mux.HandleFunc("/ws", hub.ServeWS)
	return s
}

// ServeHTTP implements http.Handler so the server can be used in tests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on the given port. It returns when
// ctx is done or when the listener encounters a fatal error.
func (s *Server) ListenAndServe(ctx context.Context, port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	httpSrv := &http.Server{Handler: s}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		return httpSrv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}
