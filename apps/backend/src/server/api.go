package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lucas/oraculo/apps/backend/src/approval"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
	"github.com/lucas/oraculo/apps/backend/src/ws"
)

// APIHandler handles REST API requests.
type APIHandler struct {
	epics       *db.EpicStore
	stories     *db.StoryStore
	tasks       *db.TaskStore
	approvals   *db.ApprovalStore
	versions    *db.VersionStore
	reviews     *db.ReviewStore
	validations *db.ValidationStore
	bridge      *approval.Bridge
	hub         *ws.Hub
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

// resolveEpicStory resolves epic and story from path values.
func (a *APIHandler) resolveEpicStory(w http.ResponseWriter, r *http.Request) (*domain.Epic, *domain.Story, bool) {
	epicName := r.PathValue("epicName")
	epic, err := a.epics.GetByName(epicName)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "epic not found")
		return nil, nil, false
	}
	storyName := r.PathValue("storyName")
	story, err := a.stories.GetByName(epic.ID, storyName)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "story not found")
		return nil, nil, false
	}
	return epic, story, true
}

// handleListStories returns all stories for an epic with task counts.
// GET /api/epics/{epicName}/stories
func (a *APIHandler) handleListStories(w http.ResponseWriter, r *http.Request) {
	epicName := r.PathValue("epicName")
	epic, err := a.epics.GetByName(epicName)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "epic not found")
		return
	}
	summaries, err := a.stories.ListSummaries(epic.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, summaries)
}

// handleListTasks returns all tasks for a story with dependencies and results.
// GET /api/epics/{epicName}/stories/{storyName}/tasks
func (a *APIHandler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	_, story, ok := a.resolveEpicStory(w, r)
	if !ok {
		return
	}
	enriched, err := a.tasks.ListEnriched(story.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, enriched)
}

// handleListStoryVersions returns all versions for a story.
// GET /api/epics/{epicName}/stories/{storyName}/versions
func (a *APIHandler) handleListStoryVersions(w http.ResponseWriter, r *http.Request) {
	_, story, ok := a.resolveEpicStory(w, r)
	if !ok {
		return
	}
	versions, err := a.versions.ListStoryVersions(story.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if versions == nil {
		versions = []domain.StoryVersion{}
	}
	writeJSON(w, versions)
}

// handleListStoryReviews returns all reviews across all versions of a story.
// GET /api/epics/{epicName}/stories/{storyName}/reviews
func (a *APIHandler) handleListStoryReviews(w http.ResponseWriter, r *http.Request) {
	_, story, ok := a.resolveEpicStory(w, r)
	if !ok {
		return
	}
	reviews, err := a.reviews.ListByStory(story.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, reviews)
}

// handleListValidations returns all QA validations for a story.
// GET /api/epics/{epicName}/stories/{storyName}/validations
func (a *APIHandler) handleListValidations(w http.ResponseWriter, r *http.Request) {
	_, story, ok := a.resolveEpicStory(w, r)
	if !ok {
		return
	}
	validations, err := a.validations.ListByStory(story.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, validations)
}

// handleGetApproval returns a single approval by ID.
// GET /api/approvals/{id}
func (a *APIHandler) handleGetApproval(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "missing approval id")
		return
	}
	appr, err := a.approvals.GetByID(id)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "approval not found")
		return
	}
	writeJSON(w, appr)
}

// handleListApprovals returns all approvals, optionally filtered by pending status and epic.
// GET /api/approvals
// GET /api/approvals?status=pending
// GET /api/approvals?epic_id=1
func (a *APIHandler) handleListApprovals(w http.ResponseWriter, r *http.Request) {
	pendingOnly := r.URL.Query().Get("status") == "pending"
	epicIDStr := r.URL.Query().Get("epic_id")

	if epicIDStr != "" {
		epicID, err := strconv.Atoi(epicIDStr)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid epic_id")
			return
		}
		approvals, err := a.approvals.ListByEpic(epicID, pendingOnly)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, approvals)
		return
	}

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

	if verdict == domain.VerdictApproved {
		a.approvals.DeleteCommentsByApproval(id)
	}

	updated, err := a.bridge.Status(id)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, updated)
}

// handleCreateComment creates a new comment on an approval.
// POST /api/approvals/{id}/comments
// Body: {"selected_text":"...","comment":"..."}
func (a *APIHandler) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "missing approval id")
		return
	}

	var body struct {
		SelectedText string `json:"selected_text"`
		Comment      string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.SelectedText == "" || body.Comment == "" {
		writeAPIError(w, http.StatusBadRequest, "selected_text and comment are required")
		return
	}

	comment, err := a.approvals.CreateComment(id, body.SelectedText, body.Comment)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, comment)
}

// handleListComments returns all comments for an approval.
// GET /api/approvals/{id}/comments
func (a *APIHandler) handleListComments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeAPIError(w, http.StatusBadRequest, "missing approval id")
		return
	}

	comments, err := a.approvals.ListComments(id)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if comments == nil {
		comments = []domain.ApprovalComment{}
	}
	writeJSON(w, comments)
}

// handleDeleteComment deletes a single comment by ID.
// DELETE /api/approvals/{id}/comments/{commentId}
func (a *APIHandler) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("commentId")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	if err := a.approvals.DeleteComment(commentID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
