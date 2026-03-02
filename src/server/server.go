package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/lucas/oraculo/src/applog"
	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/ws"
)

// Server is the HTTP server that exposes hooks, API, and WebSocket endpoints.
type Server struct {
	mux      *http.ServeMux
	database *db.DB
}

// New constructs a Server wired with all stores, bridge, and hub.
func New(database *db.DB, bridge *approval.Bridge, hub *ws.Hub, logs *applog.Broadcaster) *Server {
	hook := &HookHandler{
		agents:   db.NewAgentStore(database),
		toolEvts: db.NewToolEventStore(database),
		hub:      hub,
	}

	api := &APIHandler{
		epics:     db.NewEpicStore(database),
		stories:   db.NewStoryStore(database),
		tasks:     db.NewTaskStore(database),
		approvals: db.NewApprovalStore(database),
		bridge:    bridge,
		hub:       hub,
	}

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", handleHealth)

	// Hook endpoints
	mux.HandleFunc("POST /hooks/agent-start", hook.handleAgentStart)
	mux.HandleFunc("POST /hooks/agent-stop", hook.handleAgentStop)
	mux.HandleFunc("POST /hooks/tool-used", hook.handleToolUsed)
	mux.HandleFunc("POST /hooks/task-completed", hook.handleTaskCompleted)
	mux.HandleFunc("POST /hooks/stop", hook.handleStop)
	mux.HandleFunc("POST /hooks/teammate-idle", hook.handleTeammateIdle)
	mux.HandleFunc("POST /hooks/session-start", hook.handleSessionStart)
	mux.HandleFunc("POST /hooks/session-end", hook.handleSessionEnd)

	// API endpoints
	mux.HandleFunc("GET /api/epics", api.handleListEpics)
	mux.HandleFunc("GET /api/approvals", api.handleListApprovals)
	mux.HandleFunc("POST /api/approvals/{id}/verdict", api.handleVerdict)

	// WebSocket
	mux.HandleFunc("GET /ws", hub.ServeWS)

	// SSE log stream
	if logs != nil {
		mux.HandleFunc("GET /logs", logs.ServeSSE)
	}

	return &Server{mux: mux, database: database}
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on the given port.
// It returns when ctx is cancelled or an error occurs.
func (s *Server) ListenAndServe(ctx context.Context, port int) error {
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.mux,
	}

	go func() {
		<-ctx.Done()
		httpSrv.Shutdown(context.Background())
	}()

	ln, err := net.Listen("tcp", httpSrv.Addr)
	if err != nil {
		return err
	}
	if err := httpSrv.Serve(ln); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}
