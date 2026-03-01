package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
)

// handleHealth responds with 200 OK for health checks.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleApprovals handles GET /api/approvals.
// Accepts optional query parameter: status=pending
func (s *Server) handleApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pendingOnly := r.URL.Query().Get("status") == "pending"
	store := db.NewApprovalStore(s.database)
	approvals, err := store.List(pendingOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, approvals)
}

// handleApprovalByID dispatches to sub-handlers based on the URL path.
//
//	GET  /api/approvals/{id}          → get by ID
//	POST /api/approvals/{id}/verdict  → submit verdict
func (s *Server) handleApprovalByID(w http.ResponseWriter, r *http.Request) {
	// Strip the "/api/approvals/" prefix.
	path := strings.TrimPrefix(r.URL.Path, "/api/approvals/")
	// path is now either "{id}" or "{id}/verdict"

	parts := strings.SplitN(path, "/", 2)
	id := parts[0]
	if id == "" {
		http.Error(w, "missing approval ID", http.StatusBadRequest)
		return
	}

	if len(parts) == 2 && parts[1] == "verdict" {
		s.handleVerdict(w, r, id)
		return
	}

	s.handleGetApproval(w, r, id)
}

// handleGetApproval handles GET /api/approvals/{id}.
func (s *Server) handleGetApproval(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store := db.NewApprovalStore(s.database)
	approval, err := store.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, approval)
}

// verdictRequest is the JSON body for POST /api/approvals/{id}/verdict.
type verdictRequest struct {
	Verdict string `json:"verdict"`
	Comment string `json:"comment"`
}

// handleVerdict handles POST /api/approvals/{id}/verdict.
func (s *Server) handleVerdict(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req verdictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	verdict := domain.Verdict(req.Verdict)
	if !verdict.Valid() {
		http.Error(w, "invalid verdict: "+req.Verdict, http.StatusBadRequest)
		return
	}

	approval, err := s.bridge.Decide(id, verdict, req.Comment)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, approval)
}

// writeJSON serialises v as JSON and writes it to w with Content-Type application/json.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
