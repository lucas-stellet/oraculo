# Session Events & Hook Purpose — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Give purpose to empty broadcast-only hooks by recording structured events in a `session_events` table, and make `SessionStart` use Claude Code's `session_id` instead of generating its own UUID.

**Architecture:** New migration (V6) adds `session_events` table and `ended_at` column to `claude_sessions`. A new `SessionEventStore` handles persistence. `SessionStart` reads Claude Code's `session_id` from stdin. `SessionEnd` becomes a command hook. HTTP handlers for Stop/TaskCompleted/TeammateIdle write to the new table.

**Tech Stack:** Go, SQLite, cobra CLI, net/http

---

### Task 1: Migration V6 — session_events table and ended_at column

**Files:**
- Modify: `src/db/migrations.go:14` (add `migrateV6` to slice)
- Modify: `src/db/migrations.go` (append `migrateV6` function)

**Step 1: Write the failing test**

Create a test that opens the DB and verifies the `session_events` table and `ended_at` column exist.

File: `src/db/session_event_store_test.go`

```go
package db_test

import (
	"testing"

	"github.com/lucas/oraculo/src/dbtest"
)

func TestMigrationV6_SessionEventsTableExists(t *testing.T) {
	database := dbtest.Open(t)
	var name string
	err := database.Conn().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='session_events'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("session_events table not found: %v", err)
	}
}

func TestMigrationV6_EndedAtColumnExists(t *testing.T) {
	database := dbtest.Open(t)
	// Insert a session and update ended_at to verify the column exists.
	_, err := database.Conn().Exec(
		"INSERT INTO claude_sessions (id) VALUES ('test-session')",
	)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
	_, err = database.Conn().Exec(
		"UPDATE claude_sessions SET ended_at = datetime('now') WHERE id = 'test-session'",
	)
	if err != nil {
		t.Fatalf("ended_at column missing or broken: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestMigrationV6 -v`
Expected: FAIL — table `session_events` not found, column `ended_at` does not exist.

**Step 3: Write the migration**

In `src/db/migrations.go`, add `migrateV6` to the slice and implement:

```go
// Add to migrations slice:
var migrations = []func(*sql.Tx) error{
	migrateV1,
	migrateV2,
	migrateV3,
	migrateV4,
	migrateV5,
	migrateV6,
}

func migrateV6(tx *sql.Tx) error {
	stmts := []string{
		`ALTER TABLE claude_sessions ADD COLUMN ended_at TEXT`,
		`CREATE TABLE session_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES claude_sessions(id),
			event_type TEXT NOT NULL
			           CHECK (event_type IN ('task_completed','stop','teammate_idle','session_end')),
			payload    TEXT DEFAULT '{}',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX idx_session_events_session ON session_events(session_id)`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration v6: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/db/ -run TestMigrationV6 -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/db/migrations.go src/db/session_event_store_test.go
git commit -m "feat: add migration V6 — session_events table and ended_at column"
```

---

### Task 2: SessionEventStore — record and list events

**Files:**
- Create: `src/db/session_event_store.go`
- Modify: `src/db/session_event_store_test.go` (add store tests)

**Step 1: Write the failing tests**

Append to `src/db/session_event_store_test.go`:

```go
import (
	"testing"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
)

func seedSession(t *testing.T, database *db.DB, id string) {
	t.Helper()
	_, err := database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES (?)", id)
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}
}

func TestSessionEventStore_Record(t *testing.T) {
	database := dbtest.Open(t)
	seedSession(t, database, "s1")
	store := db.NewSessionEventStore(database)

	err := store.Record("s1", "stop", `{"last_assistant_message":"done"}`)
	if err != nil {
		t.Fatalf("record event: %v", err)
	}

	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "stop" {
		t.Errorf("expected event_type=stop, got %s", events[0].EventType)
	}
	if events[0].SessionID != "s1" {
		t.Errorf("expected session_id=s1, got %s", events[0].SessionID)
	}
}

func TestSessionEventStore_ListBySession_MultipleEvents(t *testing.T) {
	database := dbtest.Open(t)
	seedSession(t, database, "s1")
	seedSession(t, database, "s2")
	store := db.NewSessionEventStore(database)

	store.Record("s1", "stop", "{}")
	store.Record("s1", "task_completed", `{"task_name":"auth"}`)
	store.Record("s2", "stop", "{}")

	events, err := store.ListBySession("s1", 10)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events for s1, got %d", len(events))
	}
}

func TestSessionEventStore_InvalidEventType(t *testing.T) {
	database := dbtest.Open(t)
	seedSession(t, database, "s1")
	store := db.NewSessionEventStore(database)

	err := store.Record("s1", "invalid_type", "{}")
	if err == nil {
		t.Fatal("expected error for invalid event type")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestSessionEventStore -v`
Expected: FAIL — `db.NewSessionEventStore` undefined.

**Step 3: Write the store**

Create `src/db/session_event_store.go`:

```go
package db

import (
	"fmt"
	"time"
)

// SessionEvent represents a lifecycle event recorded during a Claude Code session.
type SessionEvent struct {
	ID        int
	SessionID string
	EventType string
	Payload   string
	CreatedAt time.Time
}

// SessionEventStore provides persistence for session lifecycle events.
type SessionEventStore struct {
	db *DB
}

// NewSessionEventStore returns a SessionEventStore using the given database.
func NewSessionEventStore(db *DB) *SessionEventStore {
	return &SessionEventStore{db: db}
}

// Record inserts a new session event.
func (s *SessionEventStore) Record(sessionID, eventType, payload string) error {
	_, err := s.db.conn.Exec(
		"INSERT INTO session_events (session_id, event_type, payload) VALUES (?, ?, ?)",
		sessionID, eventType, payload,
	)
	if err != nil {
		return fmt.Errorf("record session event: %w", err)
	}
	return nil
}

// ListBySession returns up to limit session events ordered by creation time.
func (s *SessionEventStore) ListBySession(sessionID string, limit int) ([]SessionEvent, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, session_id, event_type, payload, created_at FROM session_events WHERE session_id = ? ORDER BY created_at LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list session events: %w", err)
	}
	defer rows.Close()

	var events []SessionEvent
	for rows.Next() {
		var (
			e         SessionEvent
			createdAt string
		)
		if err := rows.Scan(&e.ID, &e.SessionID, &e.EventType, &e.Payload, &createdAt); err != nil {
			return nil, fmt.Errorf("scan session event: %w", err)
		}
		if e.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session events: %w", err)
	}
	return events, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/db/ -run TestSessionEventStore -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/db/session_event_store.go src/db/session_event_store_test.go
git commit -m "feat: add SessionEventStore for recording session lifecycle events"
```

---

### Task 3: Refactor SessionStart to use Claude Code's session_id from stdin

**Files:**
- Modify: `src/cli/hook_session.go:34-95`
- Modify: `src/cli/hook_session_test.go`

**Step 1: Write the failing test**

Add to `src/cli/hook_session_test.go`:

```go
func TestHookSessionStart_UsesClaudeSessionID(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// Provide Claude Code JSON on stdin
	stdinJSON := `{"session_id":"claude-abc-123","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify the session was created with Claude's session_id, not a UUID
	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	var id string
	err = database.Conn().QueryRow(
		"SELECT id FROM claude_sessions WHERE id = ?", "claude-abc-123",
	).Scan(&id)
	if err != nil {
		t.Fatalf("session not found with Claude's session_id: %v", err)
	}
}

func TestHookSessionStart_UpsertOnResume(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// First call — startup
	startupJSON := `{"session_id":"claude-abc-123","hook_event_name":"SessionStart","source":"startup"}`
	cmd1 := cli.NewRoot("test")
	cmd1.SetOut(&bytes.Buffer{})
	cmd1.SetErr(&bytes.Buffer{})
	cmd1.SetIn(strings.NewReader(startupJSON))
	cmd1.SetArgs([]string{"hook", "session-start"})
	cmd1.Execute()

	// Second call — resume (same session_id)
	resumeJSON := `{"session_id":"claude-abc-123","hook_event_name":"SessionStart","source":"resume"}`
	cmd2 := cli.NewRoot("test")
	cmd2.SetOut(&bytes.Buffer{})
	cmd2.SetErr(&bytes.Buffer{})
	cmd2.SetIn(strings.NewReader(resumeJSON))
	cmd2.SetArgs([]string{"hook", "session-start"})
	err := cmd2.Execute()

	if err != nil {
		t.Fatalf("resume should succeed: %v", err)
	}

	// Only one row should exist
	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	var count int
	database.Conn().QueryRow("SELECT COUNT(*) FROM claude_sessions WHERE id = ?", "claude-abc-123").Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 session row, got %d", count)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestHookSessionStart_UsesClaudeSessionID -v`
Expected: FAIL — session created with UUID, not `claude-abc-123`.

**Step 3: Refactor hookSessionStart**

Modify `src/cli/hook_session.go`. Replace the function `hookSessionStart`:

```go
// hookInput is the JSON structure Claude Code sends on stdin for all hooks.
type hookInput struct {
	SessionID string `json:"session_id"`
	Source    string `json:"source"`
}

func hookSessionStart(cmd *cobra.Command) error {
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Read Claude Code's JSON input from stdin
	var input hookInput
	if err := json.NewDecoder(cmd.InOrStdin()).Decode(&input); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if input.SessionID == "" {
		return fmt.Errorf("missing session_id in stdin")
	}

	// Collect metadata
	wd, _ := os.Getwd()
	branch := gitBranch()
	metadata := map[string]string{
		"working_dir": wd,
		"git_branch":  branch,
		"source":      input.Source,
		"updated_at":  time.Now().UTC().Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Upsert: INSERT OR IGNORE for first time, then UPDATE metadata
	_, err = database.Conn().Exec(
		"INSERT OR IGNORE INTO claude_sessions (id, metadata) VALUES (?, ?)",
		input.SessionID, string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("register session: %w", err)
	}
	_, err = database.Conn().Exec(
		"UPDATE claude_sessions SET metadata = ? WHERE id = ?",
		string(metadataJSON), input.SessionID,
	)
	if err != nil {
		return fmt.Errorf("update session metadata: %w", err)
	}

	// Health check and auto-start (unchanged logic)
	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return nil
	}

	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	online := isServerHealthy(healthURL)

	if !online {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: Oraculo HTTP server offline — auto-starting on port %d\n", port)
		if err := SpawnDaemon(); err != nil {
			msg := fmt.Sprintf("warning: failed to auto-start Oraculo server: %v", err)
			fmt.Fprintln(cmd.ErrOrStderr(), msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		}
		online = pollHealth(healthURL, 500*time.Millisecond, 10*time.Second)
		if !online {
			msg := "warning: Oraculo server started but not responding. Telemetry unavailable."
			fmt.Fprintln(cmd.ErrOrStderr(), msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		}
	}

	// POST session-start
	postURL := fmt.Sprintf("http://localhost:%d/hooks/session-start", port)
	client := &http.Client{Timeout: 2 * time.Second}
	client.Post(postURL, "application/json", strings.NewReader(string(metadataJSON)))

	return nil
}
```

Also remove the `github.com/google/uuid` import since it's no longer needed.

**Step 4: Update existing tests**

The existing tests in `hook_session_test.go` don't provide stdin, so they'll break. Update them to provide minimal stdin JSON. For `TestHookSessionStart_NoConfig`:

```go
func TestHookSessionStart_NoConfig(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	stdinJSON := `{"session_id":"test-no-config","hook_event_name":"SessionStart","source":"startup"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, buf.String())
	}

	dbPath := filepath.Join(".oraculo", "oraculo.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db to be created")
	}
}
```

Apply the same stdin pattern to `TestHookSessionStart_AlertsWhenServerOffline`, `TestHookSessionStart_AttemptsAutoStart`, and `TestHookSessionStart_AlwaysExitsZero`. Each needs `cmd.SetIn(strings.NewReader(stdinJSON))` with a valid JSON `session_id`.

**Step 5: Run all tests**

Run: `go test ./src/cli/ -run TestHookSessionStart -v`
Expected: PASS

**Step 6: Commit**

```bash
git add src/cli/hook_session.go src/cli/hook_session_test.go
git commit -m "refactor: SessionStart reads session_id from Claude Code stdin"
```

---

### Task 4: SessionEnd command hook — register ended_at and session_event

**Files:**
- Modify: `src/cli/hook_session.go` (add `newHookSessionEndCmd` and `hookSessionEnd`)
- Modify: `src/cli/root.go:19` (register the new subcommand)
- Modify: `src/cli/install.go:109` (change SessionEnd from HTTP to command)
- Create: `src/cli/hook_session_end_test.go`

**Step 1: Write the failing tests**

File: `src/cli/hook_session_end_test.go`

```go
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/cli"
	"github.com/lucas/oraculo/src/db"
)

func TestHookSessionEnd_UpdatesEndedAt(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// Seed a session first
	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES ('s1')")
	database.Close()

	stdinJSON := `{"session_id":"s1","hook_event_name":"SessionEnd","reason":"prompt_input_exit"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-end"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify ended_at was set
	database, err = db.Open()
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer database.Close()

	var endedAt *string
	database.Conn().QueryRow("SELECT ended_at FROM claude_sessions WHERE id = 's1'").Scan(&endedAt)
	if endedAt == nil {
		t.Fatal("expected ended_at to be set")
	}
}

func TestHookSessionEnd_RecordsSessionEvent(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	database, err := db.Open()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES ('s1')")
	database.Close()

	stdinJSON := `{"session_id":"s1","hook_event_name":"SessionEnd","reason":"clear"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-end"})
	cmd.Execute()

	database, err = db.Open()
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer database.Close()

	var eventType, payload string
	err = database.Conn().QueryRow(
		"SELECT event_type, payload FROM session_events WHERE session_id = 's1'",
	).Scan(&eventType, &payload)
	if err != nil {
		t.Fatalf("session event not found: %v", err)
	}
	if eventType != "session_end" {
		t.Errorf("expected event_type=session_end, got %s", eventType)
	}
	if !strings.Contains(payload, "clear") {
		t.Errorf("expected payload to contain reason, got %s", payload)
	}
}

func TestHookSessionEnd_AlwaysExitsZero(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	// No session in DB — should still exit 0
	stdinJSON := `{"session_id":"nonexistent","hook_event_name":"SessionEnd","reason":"other"}`

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(stdinJSON))
	cmd.SetArgs([]string{"hook", "session-end"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook must always exit 0, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestHookSessionEnd -v`
Expected: FAIL — `session-end` subcommand not found.

**Step 3: Implement the session-end hook**

Add to `src/cli/hook_session.go`:

```go
func newHookSessionEndCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-end",
		Short: "Register a Claude Code session end",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hookSessionEnd(cmd); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: hook session-end: %v\n", err)
			}
			return nil
		},
	}
}

func hookSessionEnd(cmd *cobra.Command) error {
	var input struct {
		SessionID string `json:"session_id"`
		Reason    string `json:"reason"`
	}
	if err := json.NewDecoder(cmd.InOrStdin()).Decode(&input); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if input.SessionID == "" {
		return fmt.Errorf("missing session_id in stdin")
	}

	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Update ended_at
	database.Conn().Exec(
		"UPDATE claude_sessions SET ended_at = datetime('now') WHERE id = ?",
		input.SessionID,
	)

	// Record session event
	payload, _ := json.Marshal(map[string]string{"reason": input.Reason})
	database.Conn().Exec(
		"INSERT INTO session_events (session_id, event_type, payload) VALUES (?, 'session_end', ?)",
		input.SessionID, string(payload),
	)

	// Notify HTTP server if online
	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return nil
	}
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	if !isServerHealthy(healthURL) {
		return nil
	}
	body, _ := json.Marshal(map[string]string{"session_id": input.SessionID, "reason": input.Reason})
	postURL := fmt.Sprintf("http://localhost:%d/hooks/session-end", port)
	client := &http.Client{Timeout: 2 * time.Second}
	client.Post(postURL, "application/json", strings.NewReader(string(body)))

	return nil
}
```

**Step 4: Register the subcommand**

In `src/cli/root.go`, add after line 19:

```go
hookCmd.AddCommand(newHookSessionEndCmd())
```

**Step 5: Update install.go — SessionEnd becomes command hook**

In `src/cli/install.go`, replace line 109:

```go
// Before:
"SessionEnd":     httpGroup(baseURL + "/hooks/session-end"),

// After:
"SessionEnd": {{
	Hooks: []hookDef{{Type: "command", Command: "oraculo hook session-end"}},
}},
```

**Step 6: Run tests**

Run: `go test ./src/cli/ -run "TestHookSessionEnd|TestInstall" -v`
Expected: PASS

**Step 7: Commit**

```bash
git add src/cli/hook_session.go src/cli/hook_session_end_test.go src/cli/root.go src/cli/install.go
git commit -m "feat: SessionEnd command hook records ended_at and session_event"
```

---

### Task 5: HTTP handlers — Stop, TaskCompleted, TeammateIdle write to session_events

**Files:**
- Modify: `src/server/hooks.go:12-18` (add `sessEvts` field to `HookHandler`)
- Modify: `src/server/hooks.go:99-136` (update `handleTaskCompleted`, `handleStop`, `handleTeammateIdle`)
- Modify: `src/server/server.go:35-40` (wire `SessionEventStore` into `HookHandler`)
- Modify: `src/server/hooks_test.go` (update tests to verify DB writes)

**Step 1: Write the failing tests**

Update tests in `src/server/hooks_test.go`. Add a helper to query session_events:

```go
func seedClaudeSession(t *testing.T, database *db.DB, id string) {
	t.Helper()
	_, err := database.Conn().Exec("INSERT INTO claude_sessions (id) VALUES (?)", id)
	if err != nil {
		t.Fatalf("seed claude session: %v", err)
	}
}
```

Update `testServer` to also return the database so tests can query it:

```go
func testServerWithDB(t *testing.T) (*server.Server, *db.DB) {
	t.Helper()
	database := dbtest.Open(t)
	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	return server.New(database, bridge, hub, nil), database
}
```

Add new tests:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/server/ -run "WritesSessionEvent" -v`
Expected: FAIL — `HookHandler` has no `sessEvts` field, no DB writes happening.

**Step 3: Wire SessionEventStore into HookHandler**

In `src/server/hooks.go`, add the field:

```go
type HookHandler struct {
	agents   *db.AgentStore
	toolEvts *db.ToolEventStore
	sessEvts *db.SessionEventStore
	hub      *ws.Hub
	logger   *slog.Logger
}
```

In `src/server/server.go`, wire it in the `New` function:

```go
hook := &HookHandler{
	agents:   db.NewAgentStore(database),
	toolEvts: db.NewToolEventStore(database),
	sessEvts: db.NewSessionEventStore(database),
	hub:      hub,
	logger:   logger,
}
```

**Step 4: Update handleStop**

```go
func (h *HookHandler) handleStop(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID           string `json:"session_id"`
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
```

**Step 5: Update handleTaskCompleted**

```go
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
```

**Step 6: Update handleTeammateIdle**

```go
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
```

**Step 7: Run all tests**

Run: `go test ./src/server/ -v`
Expected: PASS

Also run: `go test ./... -count=1` to ensure nothing else broke.

**Step 8: Commit**

```bash
git add src/server/hooks.go src/server/server.go src/server/hooks_test.go
git commit -m "feat: Stop, TaskCompleted, TeammateIdle handlers write to session_events"
```

---

### Task 6: Full integration verification

**Files:** None — verification only.

**Step 1: Run the full test suite**

Run: `go test ./... -count=1 -v`
Expected: All tests PASS.

**Step 2: Build the binary**

Run: `go build -o /dev/null ./...`
Expected: Clean build, no errors.

**Step 3: Verify install output includes SessionEnd as command hook**

Run the install test specifically:

Run: `go test ./src/cli/ -run TestInstall -v`
Expected: All install tests PASS, including `TestInstall_HooksConfiguration` which validates SessionEnd.

**Step 4: Commit (if any fixups needed)**

Only if previous tasks needed adjustments.
