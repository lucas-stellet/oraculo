package server_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/approval"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/dbtest"
	"github.com/lucas/oraculo/apps/backend/src/server"
	"github.com/lucas/oraculo/apps/backend/src/ws"
)

func testServer(t *testing.T) *server.Server {
	t.Helper()
	database := dbtest.Open(t)
	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	return server.New(database, bridge, hub, nil, "", "test")
}

func TestHealthEndpoint(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func seedEpicStoryTask(t *testing.T, database *db.DB) (epicName, storyName, taskName string) {
	t.Helper()
	epicName, storyName, taskName = "gastos", "registro", "implement-api"
	epicStore := db.NewEpicStore(database)
	epic, _, err := epicStore.Create(epicName, "")
	if err != nil {
		t.Fatalf("seed epic: %v", err)
	}
	storyStore := db.NewStoryStore(database)
	story, _, err := storyStore.Create(epic.ID, storyName, "")
	if err != nil {
		t.Fatalf("seed story: %v", err)
	}
	taskStore := db.NewTaskStore(database)
	_, _, err = taskStore.Create(story.ID, taskName, "", nil)
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}
	return
}

func TestAgentStartHook_WithTask(t *testing.T) {
	srv, database := testServerWithDB(t)
	epicName, storyName, taskName := seedEpicStoryTask(t, database)

	body := fmt.Sprintf(`{"session_id":"s1","agent_name":"code-01","agent_type":"code","task_name":%q,"story_name":%q,"epic_name":%q}`,
		taskName, storyName, epicName)
	req := httptest.NewRequest("POST", "/hooks/agent-start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200: %s", rec.Code, rec.Body.String())
	}
}

func TestAgentStartHook_MissingTaskFields_Returns400(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1","agent_name":"code-01","agent_type":"code"}`
	req := httptest.NewRequest("POST", "/hooks/agent-start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestAgentStopHook(t *testing.T) {
	srv, database := testServerWithDB(t)
	epicName, storyName, taskName := seedEpicStoryTask(t, database)

	// First start an agent.
	startBody := fmt.Sprintf(`{"session_id":"s1","agent_name":"code-agent","agent_type":"code","task_name":%q,"story_name":%q,"epic_name":%q}`,
		taskName, storyName, epicName)
	startReq := httptest.NewRequest("POST", "/hooks/agent-start", strings.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	srv.ServeHTTP(startRec, startReq)
	if startRec.Code != 200 {
		t.Fatalf("agent-start status %d, want 200: %s", startRec.Code, startRec.Body.String())
	}

	// Now stop agent with id=1.
	stopBody := `{"agent_id":1,"status":"completed"}`
	stopReq := httptest.NewRequest("POST", "/hooks/agent-stop", strings.NewReader(stopBody))
	stopReq.Header.Set("Content-Type", "application/json")
	stopRec := httptest.NewRecorder()
	srv.ServeHTTP(stopRec, stopReq)
	if stopRec.Code != 200 {
		t.Fatalf("agent-stop status %d, want 200", stopRec.Code)
	}
}

func TestToolUsedHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1","tool_name":"Read","file_path":"/src/main.go"}`
	req := httptest.NewRequest("POST", "/hooks/tool-used", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func TestTaskCompletedHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1","task_name":"implement-feature","status":"completed"}`
	req := httptest.NewRequest("POST", "/hooks/task-completed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func TestStopHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1","last_assistant_message":"done"}`
	req := httptest.NewRequest("POST", "/hooks/stop", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func TestTeammateIdleHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1","agent_name":"qa-agent"}`
	req := httptest.NewRequest("POST", "/hooks/teammate-idle", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func TestSessionStartHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1"}`
	req := httptest.NewRequest("POST", "/hooks/session-start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func seedClaudeSession(t *testing.T, database *db.DB, id string) {
	t.Helper()
	_, err := database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES (?)", id)
	if err != nil {
		t.Fatalf("seed claude session: %v", err)
	}
}

func TestTaskCompletedHook_WritesSessionEvent(t *testing.T) {
	srv, database := testServerWithDB(t)
	seedClaudeSession(t, database, "s1")

	body := `{"session_id":"s1","task_name":"implement-auth","status":"completed"}`
	req := httptest.NewRequest("POST", "/hooks/task-completed", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}

	store := db.NewSessionEventStore(database)
	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "task_completed" {
		t.Errorf("expected task_completed, got %s", events[0].EventType)
	}
}

func TestStopHook_WritesSessionEvent(t *testing.T) {
	srv, database := testServerWithDB(t)
	seedClaudeSession(t, database, "s1")

	body := `{"session_id":"s1","last_assistant_message":"I've completed the work."}`
	req := httptest.NewRequest("POST", "/hooks/stop", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}

	store := db.NewSessionEventStore(database)
	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "stop" {
		t.Errorf("expected stop, got %s", events[0].EventType)
	}
}

func TestTeammateIdleHook_WritesSessionEvent(t *testing.T) {
	srv, database := testServerWithDB(t)
	seedClaudeSession(t, database, "s1")

	body := `{"session_id":"s1","teammate_name":"researcher","team_name":"my-project"}`
	req := httptest.NewRequest("POST", "/hooks/teammate-idle", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}

	store := db.NewSessionEventStore(database)
	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "teammate_idle" {
		t.Errorf("expected teammate_idle, got %s", events[0].EventType)
	}
}

func TestSessionEndHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1"}`
	req := httptest.NewRequest("POST", "/hooks/session-end", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}

func TestTaskStartedHook(t *testing.T) {
	srv := testServer(t)
	body := `{"session_id":"s1","task_name":"implement-feature","story_name":"registro","epic_name":"gastos"}`
	req := httptest.NewRequest("POST", "/hooks/task-started", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}
