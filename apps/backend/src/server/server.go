package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/applog"
	"github.com/lucas/oraculo/apps/backend/src/approval"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/ws"
)

// Server is the HTTP server that exposes hooks, API, and WebSocket endpoints.
type Server struct {
	mux          *http.ServeMux
	handler      http.Handler
	database     *db.DB
	lastActivity time.Time
	mu           sync.Mutex
	projectName  string
}

// New constructs a Server wired with all stores, bridge, and hub.
func New(database *db.DB, bridge *approval.Bridge, hub *ws.Hub, logs *applog.Broadcaster, projectName string, version string) *Server {
	var logger *slog.Logger
	if logs != nil {
		logger = slog.New(logs)
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

	hook := &HookHandler{
		agents:   db.NewAgentStore(database),
		toolEvts: db.NewToolEventStore(database),
		sessEvts: db.NewSessionEventStore(database),
		epics:    db.NewEpicStore(database),
		stories:  db.NewStoryStore(database),
		tasks:    db.NewTaskStore(database),
		hub:      hub,
		logger:   logger,
	}

	api := &APIHandler{
		epics:       db.NewEpicStore(database),
		stories:     db.NewStoryStore(database),
		tasks:       db.NewTaskStore(database),
		approvals:   db.NewApprovalStore(database),
		versions:    db.NewVersionStore(database),
		reviews:     db.NewReviewStore(database),
		validations: db.NewValidationStore(database),
		bridge:      bridge,
		hub:         hub,
	}

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "project_name": projectName})
	})

	// Hook endpoints
	mux.HandleFunc("POST /hooks/agent-start", hook.handleAgentStart)
	mux.HandleFunc("POST /hooks/agent-stop", hook.handleAgentStop)
	mux.HandleFunc("POST /hooks/tool-used", hook.handleToolUsed)
	mux.HandleFunc("POST /hooks/task-completed", hook.handleTaskCompleted)
	mux.HandleFunc("POST /hooks/task-started", hook.handleTaskStarted)
	mux.HandleFunc("POST /hooks/stop", hook.handleStop)
	mux.HandleFunc("POST /hooks/teammate-idle", hook.handleTeammateIdle)
	mux.HandleFunc("POST /hooks/session-start", hook.handleSessionStart)
	mux.HandleFunc("POST /hooks/session-end", hook.handleSessionEnd)

	// System endpoints
	sys := newSystemHandler(version)
	mux.HandleFunc("GET /api/system/status", sys.handleStatus)

	// API endpoints
	mux.HandleFunc("GET /api/epics", api.handleListEpics)
	mux.HandleFunc("POST /api/epics", api.handleCreateEpic)
	mux.HandleFunc("GET /api/epics/{epicName}/stories", api.handleListStories)
	mux.HandleFunc("GET /api/epics/{epicName}/stories/{storyName}/tasks", api.handleListTasks)
	mux.HandleFunc("GET /api/epics/{epicName}/stories/{storyName}/versions", api.handleListStoryVersions)
	mux.HandleFunc("GET /api/epics/{epicName}/stories/{storyName}/reviews", api.handleListStoryReviews)
	mux.HandleFunc("GET /api/epics/{epicName}/stories/{storyName}/validations", api.handleListValidations)
	mux.HandleFunc("GET /api/approvals", api.handleListApprovals)
	mux.HandleFunc("GET /api/approvals/{id}", api.handleGetApproval)
	mux.HandleFunc("POST /api/approvals/{id}/verdict", api.handleVerdict)
	mux.HandleFunc("POST /api/approvals/{id}/comments", api.handleCreateComment)
	mux.HandleFunc("GET /api/approvals/{id}/comments", api.handleListComments)
	mux.HandleFunc("DELETE /api/approvals/{id}/comments/{commentId}", api.handleDeleteComment)

	// WebSocket
	mux.HandleFunc("GET /ws", hub.ServeWS)

	// SSE log stream
	if logs != nil {
		mux.HandleFunc("GET /logs", logs.ServeSSE)
	}

	s := &Server{mux: mux, database: database, lastActivity: time.Now(), projectName: projectName}
	s.handler = corsMiddleware(mux)
	return s
}

// LastActivity returns the time of the last HTTP request.
func (s *Server) LastActivity() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastActivity
}

func (s *Server) touchActivity() {
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.touchActivity()
	s.handler.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server on the given port.
// If idleTimeout > 0, the server shuts down after that duration of inactivity.
// It returns when ctx is cancelled, the idle timeout fires, or an error occurs.
func (s *Server) ListenAndServe(ctx context.Context, port int, idleTimeout time.Duration) error {
	addr := fmt.Sprintf(":%d", port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: s,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Idle timeout watchdog
	if idleTimeout > 0 {
		go func() {
			ticker := time.NewTicker(idleTimeout / 4)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if time.Since(s.LastActivity()) > idleTimeout {
						cancel()
						return
					}
				}
			}
		}()
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		httpSrv.Shutdown(shutdownCtx)
	}()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if err := httpSrv.Serve(ln); err != http.ErrServerClosed {
		return err
	}
	return nil
}
