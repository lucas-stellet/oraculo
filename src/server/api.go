package server

import (
	"encoding/json"
	"net/http"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/ws"
)

// APIHandler handles REST API requests.
type APIHandler struct {
	epics     *db.EpicStore
	stories   *db.StoryStore
	tasks     *db.TaskStore
	approvals *db.ApprovalStore
	bridge    *approval.Bridge
	hub       *ws.Hub
}

// handleListEpics returns all epics with aggregated summary data.
// GET /api/epics
func (a *APIHandler) handleListEpics(w http.ResponseWriter, r *http.Request) {
	summaries, err := a.epics.ListSummaries()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, summaries)
}

// handleCreateEpic creates a new epic.
// POST /api/epics
// Body: {"name":"...","description":"..."}
func (a *APIHandler) handleCreateEpic(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		writeAPIError(w, http.StatusBadRequest, "name is required")
		return
	}

	epic, created, err := a.epics.Create(body.Name, body.Description)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !created {
		writeAPIError(w, http.StatusConflict, "epic already exists")
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, epic)
}

// handleListApprovals returns all approvals, optionally filtered by pending status.
// GET /api/approvals
// GET /api/approvals?status=pending
func (a *APIHandler) handleListApprovals(w http.ResponseWriter, r *http.Request) {
	pendingOnly := r.URL.Query().Get("status") == "pending"
	approvals, err := a.approvals.List(pendingOnly)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if approvals == nil {
		approvals = []domain.Approval{}
	}
	writeJSON(w, approvals)
}

// handleVerdict submits a verdict for an approval.
// POST /api/approvals/{id}/verdict
// Body: {"verdict":"approved","comment":"..."}
func (a *APIHandler) handleVerdict(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "missing approval id")
		return
	}

	var body struct {
		Verdict string `json:"verdict"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	verdict := domain.Verdict(body.Verdict)
	if !verdict.Valid() {
		writeAPIError(w, http.StatusBadRequest, "invalid verdict: must be approved, rejected, or needs_revision")
		return
	}

	if err := a.bridge.Decide(id, verdict, body.Comment); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := a.bridge.Status(id)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, updated)
}
