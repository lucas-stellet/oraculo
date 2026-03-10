# WebSocket Dashboard + Agent↔Task Association Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect the dashboard to the backend WebSocket for real-time updates (approvals, task progress, agent activity) and persist the agent↔task association so the Tasks tab shows which agent is executing each task.

**Architecture:** A single `WebSocketProvider` in the epic layout opens one WS connection; each `_client.tsx` subscribes via `useWebSocket(handler)`. The backend gains a `task_id` column on `agents`, two new CLI hook commands (`agent-start`, `task-started`) called by the execute skill, and a new `POST /hooks/task-started` endpoint.

**Tech Stack:** Go (backend), React/Next.js static export (frontend), SQLite (persistence), `github.com/coder/websocket` (WS), Cobra (CLI)

**Spec:** `docs/superpowers/specs/2026-03-10-websocket-dashboard-design.md`

---

## Chunk 1: Backend — DB Migration + AgentStore

### Files
- Modify: `apps/backend/src/db/migrations.go`
- Modify: `apps/backend/src/db/agent_store.go`
- Modify: `apps/backend/src/db/agent_store_test.go`

---

### Task 1: Migration V8 — add `task_id` to `agents`

**Files:**
- Modify: `apps/backend/src/db/migrations.go`

- [ ] **Step 1: Add `migrateV8` and register it**

  In `migrations.go`, append `migrateV8` to the `migrations` slice and add the function:

  ```go
  var migrations = []func(*sql.Tx) error{
      migrateV1, migrateV2, migrateV3, migrateV4,
      migrateV5, migrateV6, migrateV7, migrateV8,
  }

  func migrateV8(tx *sql.Tx) error {
      _, err := tx.Exec(`ALTER TABLE agents ADD COLUMN task_id INTEGER REFERENCES tasks(id)`)
      if err != nil {
          return fmt.Errorf("migration v8: %w", err)
      }
      return nil
  }
  ```

- [ ] **Step 2: Run existing DB tests to confirm migration applies cleanly**

  ```bash
  cd apps/backend && go test ./src/db/... -v -run TestDB
  ```
  Expected: all existing tests pass; schema version advances to 8 on fresh DB.

- [ ] **Step 3: Commit**

  ```bash
  git add apps/backend/src/db/migrations.go
  git commit -m "feat(db): add task_id column to agents (migration v8)"
  ```

---

### Task 2: Update `AgentStore.Start()` to accept `taskID *int`

**Files:**
- Modify: `apps/backend/src/db/agent_store.go`
- Modify: `apps/backend/src/db/agent_store_test.go`

- [ ] **Step 1: Write the failing test**

  In `agent_store_test.go`, add after `TestAgentStore_ListBySession`:

  ```go
  func TestAgentStore_StartWithTaskID(t *testing.T) {
      database := testDB(t)

      // Seed an epic, story, and task so we have a valid task_id.
      epicStore := NewEpicStore(database)
      epic, _ := epicStore.Create("gastos", "")
      storyStore := NewStoryStore(database)
      story, _ := storyStore.Create(epic.ID, "registro", "")
      taskStore := NewTaskStore(database)
      task, _, _ := taskStore.Create(story.ID, "implement-api", "", nil)

      store := NewAgentStore(database)
      taskID := task.ID
      agent, err := store.Start("s1", "code-01", "code", &taskID)
      if err != nil {
          t.Fatalf("Start: %v", err)
      }
      if agent.TaskID == nil {
          t.Fatal("TaskID should not be nil")
      }
      if *agent.TaskID != task.ID {
          t.Errorf("TaskID = %d, want %d", *agent.TaskID, task.ID)
      }
  }

  func TestAgentStore_StartWithoutTaskID(t *testing.T) {
      database := testDB(t)
      store := NewAgentStore(database)
      agent, err := store.Start("s1", "code-01", "code", nil)
      if err != nil {
          t.Fatalf("Start: %v", err)
      }
      if agent.TaskID != nil {
          t.Errorf("TaskID should be nil, got %d", *agent.TaskID)
      }
  }
  ```

- [ ] **Step 2: Run tests to confirm they fail**

  ```bash
  cd apps/backend && go test ./src/db/... -v -run TestAgentStore_StartWith
  ```
  Expected: compile error — `Start` does not accept 4 args yet.

- [ ] **Step 3: Update `Agent` struct and `Start()` in `agent_store.go`**

  Add `TaskID *int` to the `Agent` struct:

  ```go
  type Agent struct {
      ID        int
      SessionID string
      Name      string
      Type      string
      Status    string
      StartedAt time.Time
      StoppedAt *time.Time
      TaskID    *int
  }
  ```

  Update `scanAgent` to scan the new column:

  ```go
  func scanAgent(row interface{ Scan(...any) error }) (*Agent, error) {
      var (
          a         Agent
          startedAt string
          stoppedAt sql.NullString
          taskID    sql.NullInt64
      )
      if err := row.Scan(
          &a.ID, &a.SessionID, &a.Name, &a.Type, &a.Status,
          &startedAt, &stoppedAt, &taskID,
      ); err != nil {
          return nil, err
      }
      var err error
      if a.StartedAt, err = time.Parse(sqliteTimeFmt, startedAt); err != nil {
          return nil, fmt.Errorf("parse started_at: %w", err)
      }
      if stoppedAt.Valid {
          t, err := time.Parse(sqliteTimeFmt, stoppedAt.String)
          if err != nil {
              return nil, fmt.Errorf("parse stopped_at: %w", err)
          }
          a.StoppedAt = &t
      }
      if taskID.Valid {
          id := int(taskID.Int64)
          a.TaskID = &id
      }
      return &a, nil
  }
  ```

  Update the SELECT queries to include `task_id`. Find the two query strings in `Start`, `Stop`, `ListBySession`, `getByID` that select from `agents` and add `task_id` to the column list:

  ```go
  // Before (in each query):
  "SELECT id, session_id, name, type, status, started_at, stopped_at FROM agents ..."
  // After:
  "SELECT id, session_id, name, type, status, started_at, stopped_at, task_id FROM agents ..."
  ```

  Update `Start()` signature and INSERT:

  ```go
  func (s *AgentStore) Start(sessionID, name, agentType string, taskID *int) (*Agent, error) {
      now := time.Now().UTC().Format(sqliteTimeFmt)
      var res sql.Result
      var err error
      if taskID != nil {
          res, err = s.db.conn.Exec(
              "INSERT INTO agents (session_id, name, type, status, started_at, task_id) VALUES (?, ?, ?, 'active', ?, ?)",
              sessionID, name, agentType, now, *taskID,
          )
      } else {
          res, err = s.db.conn.Exec(
              "INSERT INTO agents (session_id, name, type, status, started_at) VALUES (?, ?, ?, 'active', ?)",
              sessionID, name, agentType, now,
          )
      }
      if err != nil {
          return nil, fmt.Errorf("start agent: %w", err)
      }
      id, err := res.LastInsertId()
      if err != nil {
          return nil, fmt.Errorf("last insert id: %w", err)
      }
      return s.getByID(int(id))
  }
  ```

- [ ] **Step 4: Fix the existing `TestAgentStore_StartAndStop` — update `Start` call**

  The existing test calls `store.Start("session-1", "code-agent", "code")` — update to pass `nil` as 4th arg:

  ```go
  agent, err := store.Start("session-1", "code-agent", "code", nil)
  ```

  Do the same for all other calls in `TestAgentStore_ListBySession`.

- [ ] **Step 5: Run all DB tests**

  ```bash
  cd apps/backend && go test ./src/db/... -v
  ```
  Expected: all pass.

- [ ] **Step 6: Commit**

  ```bash
  git add apps/backend/src/db/agent_store.go apps/backend/src/db/agent_store_test.go
  git commit -m "feat(db): agent Store.Start accepts optional task_id"
  ```

---

## Chunk 2: Backend — Hooks Update + New Endpoint

### Files
- Modify: `apps/backend/src/server/hooks.go`
- Modify: `apps/backend/src/server/hooks_test.go`
- Modify: `apps/backend/src/server/server.go`

---

### Task 3: Update `handleAgentStart` — require task fields + resolve task_id

**Files:**
- Modify: `apps/backend/src/server/hooks.go`
- Modify: `apps/backend/src/server/server.go`

- [ ] **Step 1: Write failing tests**

  In `hooks_test.go`, replace `TestAgentStartHook` and add new ones:

  ```go
  func seedEpicStoryTask(t *testing.T, database *db.DB) (epicName, storyName, taskName string) {
      t.Helper()
      epicName, storyName, taskName = "gastos", "registro", "implement-api"
      epicStore := db.NewEpicStore(database)
      epic, err := epicStore.Create(epicName, "")
      if err != nil {
          t.Fatalf("seed epic: %v", err)
      }
      storyStore := db.NewStoryStore(database)
      story, err := storyStore.Create(epic.ID, storyName, "")
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
  ```

  Also update `TestAgentStopHook` which currently starts an agent first — the agent-start call there will fail without task fields. Update its setup to use `seedEpicStoryTask` and include the task fields. Leave the old `TestAgentStartHook` deleted (it will be replaced by the two new ones above).

- [ ] **Step 2: Run tests to confirm they fail**

  ```bash
  cd apps/backend && go test ./src/server/... -v -run TestAgentStart
  ```
  Expected: compile errors or 200 where 400 expected — confirms change needed.

- [ ] **Step 3: Add epic/story/task stores to `HookHandler` in `hooks.go`**

  ```go
  type HookHandler struct {
      agents   *db.AgentStore
      toolEvts *db.ToolEventStore
      sessEvts *db.SessionEventStore
      epics    *db.EpicStore
      stories  *db.StoryStore
      tasks    *db.TaskStore
      hub      *ws.Hub
      logger   *slog.Logger
  }
  ```

- [ ] **Step 4: Update `handleAgentStart` to require and resolve task fields**

  ```go
  func (h *HookHandler) handleAgentStart(w http.ResponseWriter, r *http.Request) {
      var body struct {
          SessionID string `json:"session_id"`
          AgentName string `json:"agent_name"`
          AgentType string `json:"agent_type"`
          TaskName  string `json:"task_name"`
          StoryName string `json:"story_name"`
          EpicName  string `json:"epic_name"`
      }
      if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
          writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
          return
      }
      if body.TaskName == "" || body.StoryName == "" || body.EpicName == "" {
          writeAPIError(w, http.StatusBadRequest, "task_name, story_name, and epic_name are required")
          return
      }

      epic, err := h.epics.GetByName(body.EpicName)
      if err != nil {
          writeAPIError(w, http.StatusBadRequest, "epic not found: "+body.EpicName)
          return
      }
      story, err := h.stories.GetByName(epic.ID, body.StoryName)
      if err != nil {
          writeAPIError(w, http.StatusBadRequest, "story not found: "+body.StoryName)
          return
      }
      task, err := h.tasks.GetByName(story.ID, body.TaskName)
      if err != nil {
          writeAPIError(w, http.StatusBadRequest, "task not found: "+body.TaskName)
          return
      }

      taskID := task.ID
      h.logger.Info("hook.agent_started", "agent", body.AgentName, "type", body.AgentType, "task_id", taskID)
      agent, err := h.agents.Start(body.SessionID, body.AgentName, body.AgentType, &taskID)
      if err != nil {
          h.logger.Warn("hook.agent_started.error", "err", err)
          writeAPIError(w, http.StatusInternalServerError, err.Error())
          return
      }
      h.broadcast("agent_started", agent)
      writeJSON(w, agent)
  }
  ```

- [ ] **Step 5: Wire new stores in `server.go`**

  In `New()`, update the `HookHandler` initialization:

  ```go
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
  ```

- [ ] **Step 6: Run hook tests**

  ```bash
  cd apps/backend && go test ./src/server/... -v -run TestAgentStart
  ```
  Expected: `TestAgentStartHook_WithTask` passes, `TestAgentStartHook_MissingTaskFields_Returns400` passes.

---

### Task 4: Add `POST /hooks/task-started` endpoint

**Files:**
- Modify: `apps/backend/src/server/hooks.go`
- Modify: `apps/backend/src/server/hooks_test.go`
- Modify: `apps/backend/src/server/server.go`

- [ ] **Step 1: Write the failing test**

  In `hooks_test.go`:

  ```go
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
  ```

- [ ] **Step 2: Run test to confirm it fails**

  ```bash
  cd apps/backend && go test ./src/server/... -v -run TestTaskStartedHook
  ```
  Expected: FAIL — 405 Method Not Allowed (route not registered).

- [ ] **Step 3: Add `handleTaskStarted` to `hooks.go`**

  Add after `handleTaskCompleted`:

  ```go
  // handleTaskStarted broadcasts a task start event.
  // POST /hooks/task-started
  // Body: {"session_id":"...","task_name":"...","story_name":"...","epic_name":"..."}
  func (h *HookHandler) handleTaskStarted(w http.ResponseWriter, r *http.Request) {
      var body struct {
          SessionID string `json:"session_id"`
          TaskName  string `json:"task_name"`
          StoryName string `json:"story_name"`
          EpicName  string `json:"epic_name"`
      }
      if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
          writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
          return
      }
      h.logger.Info("hook.task_started", "task", body.TaskName, "story", body.StoryName, "epic", body.EpicName)
      h.broadcast("task_started", body)
      writeJSON(w, map[string]string{"status": "ok"})
  }
  ```

- [ ] **Step 4: Register the route in `server.go`**

  Add after the `task-completed` route:

  ```go
  mux.HandleFunc("POST /hooks/task-started", hook.handleTaskStarted)
  ```

- [ ] **Step 5: Run all server tests**

  ```bash
  cd apps/backend && go test ./src/server/... -v
  ```
  Expected: all pass.

- [ ] **Step 6: Commit**

  ```bash
  git add apps/backend/src/server/hooks.go apps/backend/src/server/hooks_test.go apps/backend/src/server/server.go
  git commit -m "feat(server): require task context in agent-start hook; add task-started hook"
  ```

---

## Chunk 3: Backend CLI — New Hook Commands

### Files
- Create: `apps/backend/src/cli/hook_agent.go`
- Modify: `apps/backend/src/cli/root.go`

---

### Task 5: Add `oraculo hook agent-start` and `oraculo hook task-started` CLI commands

**Files:**
- Create: `apps/backend/src/cli/hook_agent.go`
- Modify: `apps/backend/src/cli/root.go`

These CLI commands are called by the execute skill's orchestrator. They POST to the HTTP server. They follow the same pattern as `hook_session.go`.

- [ ] **Step 1: Write the test**

  In `apps/backend/src/cli/`, create `hook_agent_test.go`:

  ```go
  package cli_test

  import (
      "bytes"
      "testing"
  )

  func TestHookAgentStartCmd_MissingRequired(t *testing.T) {
      root := newTestRoot()
      buf := &bytes.Buffer{}
      root.SetOut(buf)
      root.SetErr(buf)
      root.SetArgs([]string{"hook", "agent-start"})
      err := root.Execute()
      // Should fail: required flags missing
      if err == nil {
          t.Fatal("expected error for missing required flags")
      }
  }

  func TestHookTaskStartedCmd_MissingRequired(t *testing.T) {
      root := newTestRoot()
      buf := &bytes.Buffer{}
      root.SetOut(buf)
      root.SetErr(buf)
      root.SetArgs([]string{"hook", "task-started"})
      err := root.Execute()
      if err == nil {
          t.Fatal("expected error for missing required flags")
      }
  }
  ```

  Check if `newTestRoot` exists in cli tests; if not, use `NewRoot("test")` directly.

- [ ] **Step 2: Run to confirm failure**

  ```bash
  cd apps/backend && go test ./src/cli/... -v -run TestHookAgent
  ```
  Expected: compile error — commands don't exist yet.

- [ ] **Step 3: Create `hook_agent.go`**

  ```go
  // apps/backend/src/cli/hook_agent.go
  package cli

  import (
      "encoding/json"
      "fmt"
      "net/http"
      "strings"
      "time"

      "github.com/lucas/oraculo/apps/backend/src/config"
      "github.com/spf13/cobra"
  )

  func newHookAgentStartCmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "agent-start",
          Short: "Register an agent start with task association",
          RunE: func(cmd *cobra.Command, args []string) error {
              // Always succeed — never block orchestrator
              if err := hookAgentStart(cmd); err != nil {
                  fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook agent-start: %v\n", err)
              }
              return nil
          },
      }
      cmd.Flags().String("session-id", "", "Claude Code session ID")
      cmd.Flags().String("agent-name", "", "Agent name (required)")
      cmd.Flags().String("agent-type", "code", "Agent type (code, research, qa)")
      cmd.Flags().String("task-name", "", "Task name (required)")
      cmd.Flags().String("story-name", "", "Story name (required)")
      cmd.Flags().String("epic-name", "", "Epic name (required)")
      _ = cmd.MarkFlagRequired("agent-name")
      _ = cmd.MarkFlagRequired("task-name")
      _ = cmd.MarkFlagRequired("story-name")
      _ = cmd.MarkFlagRequired("epic-name")
      return cmd
  }

  func hookAgentStart(cmd *cobra.Command) error {
      agentName, _ := cmd.Flags().GetString("agent-name")
      agentType, _ := cmd.Flags().GetString("agent-type")
      taskName, _ := cmd.Flags().GetString("task-name")
      storyName, _ := cmd.Flags().GetString("story-name")
      epicName, _ := cmd.Flags().GetString("epic-name")
      sessionID, _ := cmd.Flags().GetString("session-id")

      cfg, _ := config.Read()
      port := cfg.Port
      if port == 0 {
          return fmt.Errorf("server port not configured")
      }
      healthURL := fmt.Sprintf("http://localhost:%d/health", port)
      if !isServerHealthy(healthURL) {
          return fmt.Errorf("server not reachable on port %d", port)
      }

      payload, _ := json.Marshal(map[string]string{
          "session_id": sessionID,
          "agent_name": agentName,
          "agent_type": agentType,
          "task_name":  taskName,
          "story_name": storyName,
          "epic_name":  epicName,
      })
      postURL := fmt.Sprintf("http://localhost:%d/hooks/agent-start", port)
      client := &http.Client{Timeout: 2 * time.Second}
      resp, err := client.Post(postURL, "application/json", strings.NewReader(string(payload)))
      if err != nil {
          return fmt.Errorf("post agent-start: %w", err)
      }
      defer resp.Body.Close()
      if resp.StatusCode != http.StatusOK {
          return fmt.Errorf("agent-start returned %d", resp.StatusCode)
      }
      return nil
  }

  func newHookTaskStartedCmd() *cobra.Command {
      cmd := &cobra.Command{
          Use:   "task-started",
          Short: "Broadcast a task-started event",
          RunE: func(cmd *cobra.Command, args []string) error {
              if err := hookTaskStarted(cmd); err != nil {
                  fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook task-started: %v\n", err)
              }
              return nil
          },
      }
      cmd.Flags().String("task-name", "", "Task name (required)")
      cmd.Flags().String("story-name", "", "Story name (required)")
      cmd.Flags().String("epic-name", "", "Epic name (required)")
      _ = cmd.MarkFlagRequired("task-name")
      _ = cmd.MarkFlagRequired("story-name")
      _ = cmd.MarkFlagRequired("epic-name")
      return cmd
  }

  func hookTaskStarted(cmd *cobra.Command) error {
      taskName, _ := cmd.Flags().GetString("task-name")
      storyName, _ := cmd.Flags().GetString("story-name")
      epicName, _ := cmd.Flags().GetString("epic-name")

      cfg, _ := config.Read()
      port := cfg.Port
      if port == 0 {
          return fmt.Errorf("server port not configured")
      }
      healthURL := fmt.Sprintf("http://localhost:%d/health", port)
      if !isServerHealthy(healthURL) {
          return fmt.Errorf("server not reachable on port %d", port)
      }

      payload, _ := json.Marshal(map[string]string{
          "task_name":  taskName,
          "story_name": storyName,
          "epic_name":  epicName,
      })
      postURL := fmt.Sprintf("http://localhost:%d/hooks/task-started", port)
      client := &http.Client{Timeout: 2 * time.Second}
      resp, err := client.Post(postURL, "application/json", strings.NewReader(string(payload)))
      if err != nil {
          return fmt.Errorf("post task-started: %w", err)
      }
      defer resp.Body.Close()
      if resp.StatusCode != http.StatusOK {
          return fmt.Errorf("task-started returned %d", resp.StatusCode)
      }
      return nil
  }
  ```

- [ ] **Step 4: Register commands in `root.go`**

  ```go
  hookCmd.AddCommand(newHookSessionStartCmd())
  hookCmd.AddCommand(newHookSessionEndCmd())
  hookCmd.AddCommand(newHookAgentStartCmd())
  hookCmd.AddCommand(newHookTaskStartedCmd())
  ```

- [ ] **Step 5: Run CLI tests**

  ```bash
  cd apps/backend && go test ./src/cli/... -v -run TestHookAgent
  ```
  Expected: both tests pass (required flag validation works).

- [ ] **Step 6: Build to verify no compile errors**

  ```bash
  cd apps/backend && go build ./...
  ```
  Expected: clean build.

- [ ] **Step 7: Commit**

  ```bash
  git add apps/backend/src/cli/hook_agent.go apps/backend/src/cli/root.go
  git commit -m "feat(cli): add hook agent-start and hook task-started commands"
  ```

---

## Chunk 4: Execute Skill Update

### Files
- Modify: `claude-kit/skills/oraculo/execute/phases/01-team-assembly.md`
- Modify: `claude-kit/skills/oraculo/execute/references/agent-dispatch.md`

---

### Task 6: Add hook calls to execute skill

**Files:**
- Modify: `claude-kit/skills/oraculo/execute/phases/01-team-assembly.md`
- Modify: `claude-kit/skills/oraculo/execute/references/agent-dispatch.md`

- [ ] **Step 1: Update `01-team-assembly.md` — add CLI hook calls before each dispatch**

  In the `## Execution` section, after the current instructions for assembling the agent prompt, add:

  ```markdown
  Before dispatching each agent, call both hooks in this order:

  1. Notify task start:
     ```bash
     oraculo hook task-started \
       --task-name "<task_name>" \
       --story-name "<story_name>" \
       --epic-name "<epic_name>"
     ```

  2. Register agent with task association:
     ```bash
     oraculo hook agent-start \
       --agent-name "<agent_name>" \
       --agent-type "<code|research>" \
       --task-name "<task_name>" \
       --story-name "<story_name>" \
       --epic-name "<epic_name>"
     ```

  These calls are best-effort. If they fail (server offline), log a warning and proceed with dispatch.
  ```

- [ ] **Step 2: Update `agent-dispatch.md` — document new hooks in a new section**

  Add after `## Retry Discipline`:

  ```markdown
  ## Pre-Dispatch Hooks

  Before dispatching each agent via the Agent tool, the orchestrator MUST call two hooks:

  ### 1. `oraculo hook task-started`
  Broadcasts `task_started` WS event so the dashboard re-fetches and shows the task as in-progress.

  ```bash
  oraculo hook task-started \
    --task-name "<task_name>" \
    --story-name "<story_name>" \
    --epic-name "<epic_name>"
  ```

  ### 2. `oraculo hook agent-start`
  Registers the agent with its task association. The dashboard uses this to show "executing · {agent_name}" on the task row.

  ```bash
  oraculo hook agent-start \
    --agent-name "<agent_name>" \
    --agent-type "<code|research>" \
    --task-name "<task_name>" \
    --story-name "<story_name>" \
    --epic-name "<epic_name>"
  ```

  `--session-id` is optional. If provided, it should be the Claude Code session ID of the dispatched agent (not always known before dispatch — omit if unknown).

  Both calls are best-effort: failure logs a warning but does not block dispatch.
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add claude-kit/skills/oraculo/execute/phases/01-team-assembly.md \
          claude-kit/skills/oraculo/execute/references/agent-dispatch.md
  git commit -m "feat(skill): add pre-dispatch hook calls in execute team-assembly"
  ```

---

## Chunk 5: Frontend — WebSocket Provider

### Files
- Create: `apps/dashboard/src/lib/ws.tsx`
- Modify: `apps/dashboard/src/app/epics/[id]/layout.tsx`

---

### Task 7: Create `WebSocketProvider` and `useWebSocket` hook

**Files:**
- Create: `apps/dashboard/src/lib/ws.tsx`

- [ ] **Step 1: Create `apps/dashboard/src/lib/ws.tsx`**

  ```tsx
  "use client";

  import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useRef,
  } from "react";

  export interface WSEvent {
    event: string;
    data?: unknown;
    id?: string;
  }

  type Handler = (evt: WSEvent) => void;

  interface WSContextValue {
    subscribe: (handler: Handler) => () => void;
  }

  const WSContext = createContext<WSContextValue | null>(null);

  export function WebSocketProvider({ children }: { children: React.ReactNode }) {
    const handlersRef = useRef<Set<Handler>>(new Set());
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const connect = useCallback(() => {
      // Determine WS URL: same origin, /ws path.
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${protocol}//${window.location.host}/ws`;

      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onmessage = (msg) => {
        try {
          const evt: WSEvent = JSON.parse(msg.data);
          handlersRef.current.forEach((h) => h(evt));
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onclose = () => {
        // Reconnect after 2s unless component is unmounting
        reconnectTimerRef.current = setTimeout(connect, 2000);
      };

      ws.onerror = () => {
        ws.close();
      };
    }, []);

    useEffect(() => {
      connect();
      return () => {
        if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
        wsRef.current?.close();
      };
    }, [connect]);

    const subscribe = useCallback((handler: Handler) => {
      handlersRef.current.add(handler);
      return () => {
        handlersRef.current.delete(handler);
      };
    }, []);

    return (
      <WSContext.Provider value={{ subscribe }}>
        {children}
      </WSContext.Provider>
    );
  }

  // useWebSocket registers a handler for incoming WS events.
  // The handler is called for every event received while the component is mounted.
  // Re-renders do NOT re-subscribe — use a stable callback (useCallback) if needed.
  export function useWebSocket(handler: Handler) {
    const ctx = useContext(WSContext);
    const handlerRef = useRef(handler);
    handlerRef.current = handler;

    useEffect(() => {
      if (!ctx) return;
      // Wrap in stable ref so the subscription never changes identity.
      const stable: Handler = (evt) => handlerRef.current(evt);
      return ctx.subscribe(stable);
    }, [ctx]);
  }
  ```

- [ ] **Step 2: Update `layout.tsx` to wrap with `WebSocketProvider`**

  In `apps/dashboard/src/app/epics/[id]/layout.tsx`:

  ```tsx
  import { WebSocketProvider } from "@/lib/ws";

  // Inside the return:
  return (
    <SidebarProvider>
      <WebSocketProvider>
        <div className="flex h-screen overflow-hidden bg-[#020617]">
          <Sidebar ... />
          <div className="flex-1 overflow-y-auto">{children}</div>
        </div>
      </WebSocketProvider>
    </SidebarProvider>
  );
  ```

- [ ] **Step 3: Build dashboard to catch type errors**

  ```bash
  cd apps/dashboard && bun run build
  ```
  Expected: clean build (ws.tsx is new code, no existing consumers yet).

- [ ] **Step 4: Commit**

  ```bash
  git add apps/dashboard/src/lib/ws.tsx apps/dashboard/src/app/epics/[id]/layout.tsx
  git commit -m "feat(dashboard): add WebSocketProvider and useWebSocket hook"
  ```

---

## Chunk 6: Frontend — Per-Page Subscriptions

### Files
- Modify: `apps/dashboard/src/app/epics/[id]/approvals/_client.tsx`
- Modify: `apps/dashboard/src/app/epics/[id]/approvals/[approvalId]/review/_client.tsx`
- Modify: `apps/dashboard/src/app/epics/[id]/_client.tsx`
- Modify: `apps/dashboard/src/app/epics/[id]/stories/[storyId]/_client.tsx`

---

### Task 8: Approvals list — real-time pending/resolved updates

**Files:**
- Modify: `apps/dashboard/src/app/epics/[id]/approvals/_client.tsx`

- [ ] **Step 1: Add `useWebSocket` to `approvals/_client.tsx`**

  Import `useWebSocket` and `useCallback`, then add inside the component after the existing `useEffect`:

  ```tsx
  import { useCallback } from "react";
  import { useWebSocket } from "@/lib/ws";
  import type { WSEvent } from "@/lib/ws";

  // Inside ApprovalsPage component, after the initial useEffect:
  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event !== "approval_requested" && evt.event !== "approval_decided") return;
    if (!evt.id) return;
    api.getApproval(evt.id).then((approval) => {
      if (approval.status === "pending") {
        setPending((prev) => {
          // Replace if exists, prepend if new
          const exists = prev.some((a) => a.id === approval.id);
          return exists
            ? prev.map((a) => (a.id === approval.id ? approval : a))
            : [approval, ...prev];
        });
        setResolved((prev) => prev.filter((a) => a.id !== approval.id));
      } else {
        setResolved((prev) => {
          const exists = prev.some((a) => a.id === approval.id);
          return exists
            ? prev.map((a) => (a.id === approval.id ? approval : a))
            : [approval, ...prev];
        });
        setPending((prev) => prev.filter((a) => a.id !== approval.id));
      }
    }).catch(() => {});
  }, []);

  useWebSocket(handleWS);
  ```

  Also update the sidebar pending count in the layout: the `approval_requested` event should update `pendingCount`. Since `layout.tsx` is separate, add the same subscription pattern there too:

  In `layout.tsx`, add inside `EpicLayout`:
  ```tsx
  import { useCallback } from "react";
  import { useWebSocket } from "@/lib/ws";

  const handleWS = useCallback((evt: { event: string }) => {
    if (evt.event === "approval_requested" || evt.event === "approval_decided") {
      api.listApprovals(undefined, "pending").then((approvals) => {
        setPendingCount(approvals.length);
      }).catch(() => {});
    }
  }, []);

  useWebSocket(handleWS);
  ```

- [ ] **Step 2: Build dashboard**

  ```bash
  cd apps/dashboard && bun run build
  ```
  Expected: clean build.

- [ ] **Step 3: Commit**

  ```bash
  git add apps/dashboard/src/app/epics/[id]/approvals/_client.tsx \
          apps/dashboard/src/app/epics/[id]/layout.tsx
  git commit -m "feat(dashboard): real-time approval updates via WebSocket"
  ```

---

### Task 9: Approval review page — "already decided" banner

**Files:**
- Modify: `apps/dashboard/src/app/epics/[id]/approvals/[approvalId]/review/_client.tsx`

- [ ] **Step 1: Add WS subscription for `approval_decided`**

  In `review/_client.tsx`, add state and subscription:

  ```tsx
  import { useCallback, useState as useStateAlreadyImported } from "react";
  import { useWebSocket } from "@/lib/ws";
  import type { WSEvent } from "@/lib/ws";

  // Add state inside component:
  const [decidedElsewhere, setDecidedElsewhere] = useState(false);

  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event === "approval_decided" && evt.id === approvalId && !submitting) {
      setDecidedElsewhere(true);
    }
  }, [approvalId, submitting]);

  useWebSocket(handleWS);
  ```

  Add a banner in the JSX above the action buttons (only shown when `decidedElsewhere` is true):

  ```tsx
  {decidedElsewhere && (
    <div className="flex items-center gap-3 rounded-lg border border-amber-700 bg-amber-950/40 px-4 py-3 text-sm text-amber-300 font-[family-name:var(--font-sans)]">
      This approval was decided in another session. Refresh to see the latest status.
    </div>
  )}
  ```

- [ ] **Step 2: Build dashboard**

  ```bash
  cd apps/dashboard && bun run build
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add "apps/dashboard/src/app/epics/[id]/approvals/[approvalId]/review/_client.tsx"
  git commit -m "feat(dashboard): show banner when approval decided in another session"
  ```

---

### Task 10: Epic overview — real-time story progress bars

**Files:**
- Modify: `apps/dashboard/src/app/epics/[id]/_client.tsx`

- [ ] **Step 1: Add WS subscription for task events**

  In `epics/[id]/_client.tsx`:

  ```tsx
  import { useCallback } from "react";
  import { useWebSocket } from "@/lib/ws";
  import type { WSEvent } from "@/lib/ws";

  // Add inside EpicOverviewPage, after the initial useEffect:
  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event !== "task_started" && evt.event !== "task_completed") return;
    api.listStories(epicName).then(setStories).catch(() => {});
  }, [epicName]);

  useWebSocket(handleWS);
  ```

- [ ] **Step 2: Build dashboard**

  ```bash
  cd apps/dashboard && bun run build
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add "apps/dashboard/src/app/epics/[id]/_client.tsx"
  git commit -m "feat(dashboard): real-time story progress bars via WebSocket"
  ```

---

### Task 11: Story detail — real-time task status + agent badge

**Files:**
- Modify: `apps/dashboard/src/app/epics/[id]/stories/[storyId]/_client.tsx`

- [ ] **Step 1: Add agent activity state and WS subscription**

  In `stories/[storyId]/_client.tsx`, add new state for active agents keyed by task_id:

  ```tsx
  import { useCallback } from "react";
  import { useWebSocket } from "@/lib/ws";
  import type { WSEvent } from "@/lib/ws";

  // Add state inside StoryDetailPage:
  const [activeAgents, setActiveAgents] = useState<Record<number, string>>({});
  // Maps task_id -> agent_name for currently executing tasks

  // WS handler:
  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event === "task_started" || evt.event === "task_completed") {
      api.listTasks(epicName, storyName).then(setTasks).catch(() => {});
      // Also refresh story for updated task counts
      api.listStories(epicName).then((stories) => {
        const found = stories.find((s) => s.name === storyName);
        if (found) setStory(found);
      }).catch(() => {});
    }
    if (evt.event === "agent_started") {
      const data = evt.data as { task_id?: number; name?: string } | undefined;
      if (data?.task_id && data?.name) {
        setActiveAgents((prev) => ({ ...prev, [data.task_id!]: data.name! }));
      }
    }
    if (evt.event === "agent_stopped") {
      const data = evt.data as { task_id?: number } | undefined;
      if (data?.task_id) {
        setActiveAgents((prev) => {
          const next = { ...prev };
          delete next[data.task_id!];
          return next;
        });
      }
    }
  }, [epicName, storyName]);

  useWebSocket(handleWS);
  ```

- [ ] **Step 2: Pass `activeAgents` to `TasksTab`**

  Update the `TasksTab` call:

  ```tsx
  {activeTab === "Tasks" && <TasksTab tasks={tasks} activeAgents={activeAgents} />}
  ```

- [ ] **Step 3: Update `TasksTab` to show agent badge**

  In `apps/dashboard/src/app/epics/[id]/stories/[storyId]/_components/tasks-tab.tsx`, update the component to accept and display `activeAgents`:

  Read the file first to understand its current structure, then add the prop and badge. The badge should appear on `in_progress` tasks that have an entry in `activeAgents`. It should look like:

  ```tsx
  // Add to TasksTab props:
  activeAgents?: Record<number, string>;

  // In the task row, after the status badge, if task.status === "in_progress" and activeAgents[task.id]:
  <span className="inline-flex items-center gap-1 rounded-md bg-blue-900/50 px-2 py-0.5 text-[11px] text-blue-300 font-[family-name:var(--font-mono)]">
    <span className="h-1.5 w-1.5 rounded-full bg-blue-400 animate-pulse" />
    {activeAgents[task.id]}
  </span>
  ```

- [ ] **Step 4: Read `tasks-tab.tsx` first to understand its structure, then apply changes**

  Run: `cat apps/dashboard/src/app/epics/[id]/stories/[storyId]/_components/tasks-tab.tsx`

  Then edit accordingly following the existing code style.

- [ ] **Step 5: Build dashboard**

  ```bash
  cd apps/dashboard && bun run build
  ```
  Expected: clean build.

- [ ] **Step 6: Commit**

  ```bash
  git add "apps/dashboard/src/app/epics/[id]/stories/[storyId]/_client.tsx" \
          "apps/dashboard/src/app/epics/[id]/stories/[storyId]/_components/tasks-tab.tsx"
  git commit -m "feat(dashboard): real-time task status and agent badge via WebSocket"
  ```

---

## Final Build Verification

- [ ] **Build the full backend**

  ```bash
  cd apps/backend && go build ./...
  ```
  Expected: clean.

- [ ] **Run all backend tests**

  ```bash
  cd apps/backend && go test ./...
  ```
  Expected: all pass.

- [ ] **Build the full dashboard**

  ```bash
  cd apps/dashboard && bun run build
  ```
  Expected: clean export, no type errors.

- [ ] **Run `make build` to rebuild the Go binary with embedded dashboard**

  ```bash
  make build
  ```
  Expected: binary built successfully with updated static assets.
