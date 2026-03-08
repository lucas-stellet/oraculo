package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/ws"
)

// HookHandler handles inbound webhook calls from Claude Code hooks.
type HookHandler struct {
	agents   *db.AgentStore
	toolEvts *db.ToolEventStore
	sessEvts *db.SessionEventStore
	hub      *ws.Hub
	logger   *slog.Logger
}

func (h *HookHandler) broadcast(event string, payload any) {
	msg, err := json.Marshal(map[string]any{"event": event, "data": payload})
	if err != nil {
		return
	}
	h.hub.Broadcast(msg)
	h.logger.Info("hook.broadcast", "event", event)
}

// handleAgentStart records a new agent start event.
// POST /hooks/agent-start
// Body: {"session_id":"...","agent_name":"...","agent_type":"..."}
func (h *HookHandler) handleAgentStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
		AgentName string `json:"agent_name"`
		AgentType string `json:"agent_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.agent_started", "agent", body.AgentName, "type", body.AgentType)
	agent, err := h.agents.Start(body.SessionID, body.AgentName, body.AgentType)
	if err != nil {
		h.logger.Warn("hook.agent_started.error", "err", err)
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.broadcast("agent_started", agent)
	writeJSON(w, agent)
}

// handleAgentStop records an agent stop event.
// POST /hooks/agent-stop
// Body: {"agent_id":1,"status":"completed"}
func (h *HookHandler) handleAgentStop(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AgentID int    `json:"agent_id"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.agent_stopped", "agent_id", body.AgentID, "status", body.Status)
	agent, err := h.agents.Stop(body.AgentID, body.Status)
	if err != nil {
		h.logger.Warn("hook.agent_stopped.error", "err", err)
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.broadcast("agent_stopped", agent)
	writeJSON(w, agent)
}

// handleToolUsed records a tool usage event.
// POST /hooks/tool-used
// Body: {"session_id":"...","tool_name":"...","file_path":"..."}
func (h *HookHandler) handleToolUsed(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
		ToolName  string `json:"tool_name"`
		FilePath  string `json:"file_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.tool_used", "tool", body.ToolName, "file", body.FilePath)
	if err := h.toolEvts.Record(body.SessionID, body.ToolName, body.FilePath); err != nil {
		h.logger.Warn("hook.tool_used.error", "err", err)
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.broadcast("tool_used", body)
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleTaskCompleted broadcasts a task completion event.
// POST /hooks/task-completed
// Body: {"session_id":"...","task_name":"...","status":"..."}
func (h *HookHandler) handleTaskCompleted(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
		TaskName  string `json:"task_name"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.task_completed", "task", body.TaskName, "status", body.Status)

	payload, _ := json.Marshal(map[string]string{
		"task_name": body.TaskName,
		"status":    body.Status,
	})
	if err := h.sessEvts.Record(body.SessionID, "task_completed", string(payload)); err != nil {
		h.logger.Warn("hook.task_completed.record_error", "err", err)
	}

	h.broadcast("task_completed", body)
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleStop broadcasts a server stop event.
// POST /hooks/stop
func (h *HookHandler) handleStop(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID            string `json:"session_id"`
		LastAssistantMessage string `json:"last_assistant_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.stop", "session_id", body.SessionID)

	payload, _ := json.Marshal(map[string]string{
		"last_assistant_message": body.LastAssistantMessage,
	})
	if err := h.sessEvts.Record(body.SessionID, "stop", string(payload)); err != nil {
		h.logger.Warn("hook.stop.record_error", "err", err)
	}

	h.broadcast("stop", body)
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleTeammateIdle broadcasts a teammate idle event.
// POST /hooks/teammate-idle
func (h *HookHandler) handleTeammateIdle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID    string `json:"session_id"`
		TeammateName string `json:"teammate_name"`
		TeamName     string `json:"team_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.teammate_idle", "teammate", body.TeammateName)

	payload, _ := json.Marshal(map[string]string{
		"teammate_name": body.TeammateName,
		"team_name":     body.TeamName,
	})
	if err := h.sessEvts.Record(body.SessionID, "teammate_idle", string(payload)); err != nil {
		h.logger.Warn("hook.teammate_idle.record_error", "err", err)
	}

	h.broadcast("teammate_idle", body)
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleSessionStart broadcasts a session start event.
// POST /hooks/session-start
// Body: {"session_id":"..."}
func (h *HookHandler) handleSessionStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.session_started", "session_id", body.SessionID)
	h.broadcast("session_started", body)
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleSessionEnd broadcasts a session end event.
// POST /hooks/session-end
// Body: {"session_id":"..."}
func (h *HookHandler) handleSessionEnd(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	h.logger.Info("hook.session_ended", "session_id", body.SessionID)
	h.broadcast("session_ended", body)
	writeJSON(w, map[string]string{"status": "ok"})
}
