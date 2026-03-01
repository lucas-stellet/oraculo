// Package server implements the Oraculo HTTP server that exposes a REST API for the
// dashboard and WebSocket endpoint for real-time push.
package server

import (
	"net/http"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
)

// Hub is satisfied by *ws.Hub.
type Hub interface {
	Broadcast([]byte)
	ServeWS(w http.ResponseWriter, r *http.Request)
}

// Server is the Oraculo HTTP server. It implements http.Handler.
type Server struct {
	database *db.DB
	bridge   *approval.Bridge
	hub      Hub
	mux      *http.ServeMux
}

// New creates a new Server and registers all routes.
func New(database *db.DB, bridge *approval.Bridge, hub Hub) *Server {
	s := &Server{
		database: database,
		bridge:   bridge,
		hub:      hub,
		mux:      http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// registerRoutes registers all HTTP routes.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.hub.ServeWS(w, r)
	})
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/approvals", s.handleApprovals)
	s.mux.HandleFunc("/api/approvals/", s.handleApprovalByID)
}
