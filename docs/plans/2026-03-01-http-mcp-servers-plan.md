# HTTP + MCP Servers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement `oraculo start` (MCP + HTTP + WebSocket), `oraculo install`, and `oraculo uninstall` as defined in `docs/plans/2026-03-01-http-mcp-servers-design.md`.

**Architecture:** Single-process server launched by Claude Code via `oraculo start`. MCP runs on stdio, HTTP + WebSocket on a configured port. Approval gates use in-process Go channels. errgroup coordinates lifecycle.

**Tech Stack:** Go 1.24, coder/websocket, modelcontextprotocol/go-sdk, golang.org/x/sync/errgroup, net/http (stdlib routing), modernc.org/sqlite

---

### Task 1: Add dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add new dependencies**

Run: `cd /Users/lucas/dev/projects/oraculo && go get github.com/coder/websocket@latest github.com/modelcontextprotocol/go-sdk@latest golang.org/x/sync@latest`

**Step 2: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add websocket, mcp-go-sdk, and errgroup dependencies"
```

---

### Task 2: Export db.OpenMemory and create dbtest package

**Files:**
- Modify: `src/db/db.go`
- Create: `src/dbtest/dbtest.go`
- Test: `src/dbtest/dbtest_test.go`

**Step 1: Write the failing test**

Create `src/dbtest/dbtest_test.go`:

```go
package dbtest_test

import (
	"testing"

	"github.com/lucas/oraculo/src/dbtest"
)

func TestOpen(t *testing.T) {
	database := dbtest.Open(t)
	if database == nil {
		t.Fatal("expected non-nil database")
	}
	// Verify we can query — migrations ran
	var count int
	err := database.Conn().QueryRow("SELECT count(*) FROM epics").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/dbtest/ -v`
Expected: FAIL — package doesn't exist yet

**Step 3: Export OpenMemory in db.go**

In `src/db/db.go`, add a public wrapper for the existing `openPath(":memory:")`:

```go
// OpenMemory opens an in-memory SQLite database with all migrations applied.
// Intended for testing.
func OpenMemory() (*DB, error) {
	return openPath(":memory:")
}
```

**Step 4: Create dbtest package**

Create `src/dbtest/dbtest.go`:

```go
package dbtest

import (
	"testing"

	"github.com/lucas/oraculo/src/db"
)

// Open returns an in-memory SQLite database with all migrations applied.
// The database is closed automatically when the test finishes.
func Open(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("dbtest.Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./src/dbtest/ -v`
Expected: PASS

**Step 6: Run all existing tests to verify no regressions**

Run: `make test`
Expected: All tests pass

**Step 7: Commit**

```bash
git add src/db/db.go src/dbtest/
git commit -m "feat: export db.OpenMemory and add dbtest helper package"
```

---

### Task 3: Config package

**Files:**
- Create: `src/config/config.go`
- Test: `src/config/config_test.go`
- Modify: `src/cli/hook_session.go` (use config.Read instead of inline)

**Step 1: Write the failing tests**

Create `src/config/config_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/src/config"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })
	return tmp
}

func TestReadMissing(t *testing.T) {
	setupTestDir(t)
	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if cfg.Port != 0 {
		t.Errorf("Port = %d, want 0", cfg.Port)
	}
}

func TestWriteAndRead(t *testing.T) {
	tmp := setupTestDir(t)
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)

	err := config.Write(&config.Config{Port: 3142})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	cfg, err := config.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if cfg.Port != 3142 {
		t.Errorf("Port = %d, want 3142", cfg.Port)
	}
}

func TestFindPort(t *testing.T) {
	port, err := config.FindPort(30000, 30099)
	if err != nil {
		t.Fatalf("FindPort: %v", err)
	}
	if port < 30000 || port > 30099 {
		t.Errorf("Port %d outside range 30000-30099", port)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/config/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement config package**

Create `src/config/config.go`:

```go
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const configPath = ".oraculo/config.json"

// Config holds Oraculo project configuration.
type Config struct {
	Port int `json:"port"`
}

// Read loads .oraculo/config.json from the working directory.
// Returns zero-value Config if the file does not exist.
func Read() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// Write persists cfg to .oraculo/config.json atomically.
func Write(cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(configPath)
	tmp, err := os.CreateTemp(dir, "config-*.json")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp config: %w", err)
	}
	if err := os.Rename(tmpName, configPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

// FindPort returns the first available TCP port in [start, end].
func FindPort(start, end int) (int, error) {
	for p := start; p <= end; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			l.Close()
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available port in %d-%d", start, end)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./src/config/ -v`
Expected: PASS

**Step 5: Update hook_session.go to use config.Read()**

In `src/cli/hook_session.go`, replace the inline `configFile` struct and `readConfigPort()` function with a call to `config.Read()`. Remove the duplicated struct and function.

**Step 6: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 7: Commit**

```bash
git add src/config/ src/cli/hook_session.go
git commit -m "feat: add config package, refactor hook_session to use it"
```

---

### Task 4: Schema migration v3 + stores

**Files:**
- Modify: `src/db/migrations.go`
- Create: `src/db/agent_store.go`
- Create: `src/db/tool_event_store.go`
- Test: `src/db/agent_store_test.go`
- Test: `src/db/tool_event_store_test.go`

**Step 1: Write the failing tests for AgentStore**

Create `src/db/agent_store_test.go`:

```go
package db

import (
	"testing"
)

func TestAgentStore_StartAndStop(t *testing.T) {
	database := testDB(t)
	store := NewAgentStore(database)

	agent, err := store.Start("session-1", "code-agent", "code")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if agent.Name != "code-agent" {
		t.Errorf("Name = %q, want %q", agent.Name, "code-agent")
	}
	if agent.Status != "active" {
		t.Errorf("Status = %q, want %q", agent.Status, "active")
	}

	stopped, err := store.Stop(agent.ID, "completed")
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stopped.Status != "completed" {
		t.Errorf("Status = %q, want %q", stopped.Status, "completed")
	}
	if stopped.StoppedAt == nil {
		t.Error("StoppedAt should not be nil")
	}
}

func TestAgentStore_ListBySession(t *testing.T) {
	database := testDB(t)
	store := NewAgentStore(database)

	store.Start("session-1", "agent-a", "code")
	store.Start("session-1", "agent-b", "qa")
	store.Start("session-2", "agent-c", "research")

	agents, err := store.ListBySession("session-1")
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2", len(agents))
	}
}
```

**Step 2: Write the failing tests for ToolEventStore**

Create `src/db/tool_event_store_test.go`:

```go
package db

import (
	"testing"
)

func TestToolEventStore_RecordAndList(t *testing.T) {
	database := testDB(t)
	store := NewToolEventStore(database)

	err := store.Record("session-1", "Edit", "/src/main.go")
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	err = store.Record("session-1", "Bash", "")
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	events, err := store.ListBySession("session-1", 50)
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
	if events[0].ToolName != "Edit" {
		t.Errorf("ToolName = %q, want %q", events[0].ToolName, "Edit")
	}
}
```

**Step 3: Run tests to verify they fail**

Run: `go test ./src/db/ -run TestAgentStore -v && go test ./src/db/ -run TestToolEventStore -v`
Expected: FAIL — types and functions don't exist

**Step 4: Add migration v3**

In `src/db/migrations.go`, add the third migration to the migrations slice:

```go
// Migration v3: Hook telemetry tables
func(tx *sql.Tx) error {
    _, err := tx.Exec(`
        CREATE TABLE agents (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id  TEXT NOT NULL,
            name        TEXT NOT NULL,
            type        TEXT NOT NULL DEFAULT 'unknown',
            status      TEXT NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'completed', 'failed')),
            started_at  TEXT NOT NULL,
            stopped_at  TEXT
        );
        CREATE INDEX idx_agents_session_id ON agents(session_id);

        CREATE TABLE tool_events (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id  TEXT NOT NULL,
            tool_name   TEXT NOT NULL,
            file_path   TEXT,
            timestamp   TEXT NOT NULL
        );
        CREATE INDEX idx_tool_events_session_id ON tool_events(session_id);
        CREATE INDEX idx_tool_events_timestamp ON tool_events(timestamp);
    `)
    return err
},
```

**Step 5: Implement AgentStore**

Create `src/db/agent_store.go` following the existing store pattern (constructor takes `*DB`, methods return domain-style structs, re-read after write).

Define an `Agent` struct in the store file (or in domain if preferred):

```go
type Agent struct {
    ID        int
    SessionID string
    Name      string
    Type      string
    Status    string
    StartedAt time.Time
    StoppedAt *time.Time
}
```

Methods: `Start(sessionID, name, agentType string) (*Agent, error)`, `Stop(id int, status string) (*Agent, error)`, `ListBySession(sessionID string) ([]Agent, error)`.

**Step 6: Implement ToolEventStore**

Create `src/db/tool_event_store.go`:

```go
type ToolEvent struct {
    ID        int
    SessionID string
    ToolName  string
    FilePath  string
    Timestamp time.Time
}
```

Methods: `Record(sessionID, toolName, filePath string) error`, `ListBySession(sessionID string, limit int) ([]ToolEvent, error)`.

**Step 7: Run tests to verify they pass**

Run: `go test ./src/db/ -v`
Expected: All tests pass (including existing tests — migration v3 applied to in-memory DB)

**Step 8: Commit**

```bash
git add src/db/migrations.go src/db/agent_store.go src/db/agent_store_test.go src/db/tool_event_store.go src/db/tool_event_store_test.go
git commit -m "feat: add migration v3, AgentStore and ToolEventStore"
```

---

### Task 5: Add ErrApprovalDecided sentinel error

**Files:**
- Modify: `src/domain/errors.go`
- Modify: `src/output/json.go` (add mapping)

**Step 1: Add the error**

In `src/domain/errors.go`:

```go
var ErrApprovalDecided = errors.New("approval already decided")
```

**Step 2: Add CLI output mapping**

In `src/output/json.go`, add to the `WriteError` switch:

```go
case errors.Is(err, domain.ErrApprovalDecided):
    code = "approval_decided"
```

**Step 3: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 4: Commit**

```bash
git add src/domain/errors.go src/output/json.go
git commit -m "feat: add ErrApprovalDecided sentinel error"
```

---

### Task 6: WebSocket hub

**Files:**
- Create: `src/ws/hub.go`
- Test: `src/ws/hub_test.go`

**Step 1: Write the failing tests**

Create `src/ws/hub_test.go`:

```go
package ws_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/lucas/oraculo/src/ws"
)

func TestHub_BroadcastToClient(t *testing.T) {
	hub := ws.NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start hub in background
	go hub.Run(ctx)

	// Start test HTTP server with WS endpoint
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	// Connect a WS client
	wsURL := "ws" + srv.URL[4:] // http -> ws
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.CloseNow()

	// Give client time to register
	time.Sleep(50 * time.Millisecond)

	// Broadcast a message
	msg := map[string]string{"type": "test", "payload": "hello"}
	data, _ := json.Marshal(msg)
	hub.Broadcast(data)

	// Read message from client
	_, received, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	var got map[string]string
	json.Unmarshal(received, &got)
	if got["type"] != "test" {
		t.Errorf("type = %q, want %q", got["type"], "test")
	}
}

func TestHub_BroadcastNonBlocking(t *testing.T) {
	hub := ws.NewHub()
	// Don't start Run — broadcast should not block even without consumers
	hub.Broadcast([]byte("should not block"))
	// If we reach here, the test passes (non-blocking)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./src/ws/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement the Hub**

Create `src/ws/hub.go`:

```go
package ws

import (
	"context"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub manages WebSocket connections and broadcasts messages.
type Hub struct {
	mu        sync.Mutex
	clients   map[*client]struct{}
	broadcast chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*client]struct{}),
		broadcast: make(chan []byte, 64),
	}
}

// Run processes broadcasts until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) error {
	for {
		select {
		case msg := <-h.broadcast:
			h.mu.Lock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// Client too slow, drop message
				}
			}
			h.mu.Unlock()
		case <-ctx.Done():
			h.mu.Lock()
			for c := range h.clients {
				c.conn.Close(websocket.StatusGoingAway, "server shutdown")
				close(c.send)
				delete(h.clients, c)
			}
			h.mu.Unlock()
			return nil
		}
	}
}

// Broadcast queues a message. Non-blocking: drops if buffer is full.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
	}
}

// ServeWS upgrades an HTTP connection to WebSocket.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Accept all origins (dashboard may be on different port)
	})
	if err != nil {
		return
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 16),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	// Writer goroutine
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, c)
			h.mu.Unlock()
			conn.CloseNow()
		}()

		ctx := r.Context()
		for {
			select {
			case msg, ok := <-c.send:
				if !ok {
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Reader goroutine (just drains reads to detect disconnect)
	conn.CloseRead(r.Context())
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./src/ws/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/ws/
git commit -m "feat: add WebSocket hub with broadcast and connection management"
```

---

### Task 7: ApprovalBridge

**Files:**
- Create: `src/approval/bridge.go`
- Test: `src/approval/bridge_test.go`

**Step 1: Write the failing tests**

Create `src/approval/bridge_test.go`:

```go
package approval_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
	"github.com/lucas/oraculo/src/domain"
)

type stubBroadcaster struct {
	msgs [][]byte
	mu   sync.Mutex
}

func (b *stubBroadcaster) Broadcast(msg []byte) {
	b.mu.Lock()
	b.msgs = append(b.msgs, msg)
	b.mu.Unlock()
}

func TestBridge_RequestAndDecide(t *testing.T) {
	database := dbtest.Open(t)
	epicStore := db.NewEpicStore(database)
	epicStore.Create("my-epic", "desc")

	bridge := approval.NewBridge(db.NewApprovalStore(database), &stubBroadcaster{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *approval.VerdictResult
	var reqErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		result, reqErr = bridge.Request(ctx, approval.ApprovalRequest{
			Type:    domain.ApprovalEpicRequirements,
			Epic:    "my-epic",
			Content: "# Requirements\nSome content",
		})
	}()

	// Wait for approval to appear in DB
	time.Sleep(100 * time.Millisecond)

	// Get the pending approval
	approvals, _ := db.NewApprovalStore(database).List(true)
	if len(approvals) == 0 {
		t.Fatal("no pending approvals found")
	}

	// Decide
	err := bridge.Decide(approvals[0].ID, domain.VerdictApproved, "looks good")
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}

	<-done
	if reqErr != nil {
		t.Fatalf("Request: %v", reqErr)
	}
	if result.Verdict != domain.VerdictApproved {
		t.Errorf("Verdict = %q, want %q", result.Verdict, domain.VerdictApproved)
	}
}

func TestBridge_ContextCancellation(t *testing.T) {
	database := dbtest.Open(t)
	epicStore := db.NewEpicStore(database)
	epicStore.Create("my-epic", "desc")

	bridge := approval.NewBridge(db.NewApprovalStore(database), &stubBroadcaster{})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := bridge.Request(ctx, approval.ApprovalRequest{
			Type:    domain.ApprovalEpicRequirements,
			Epic:    "my-epic",
			Content: "content",
		})
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done
}

func TestBridge_Status(t *testing.T) {
	database := dbtest.Open(t)
	epicStore := db.NewEpicStore(database)
	epicStore.Create("my-epic", "desc")

	store := db.NewApprovalStore(database)
	bridge := approval.NewBridge(store, &stubBroadcaster{})

	appr, _ := store.Request(domain.ApprovalEpicRequirements, nil, nil, "content")

	status, err := bridge.Status(appr.ID)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Status != domain.ApprovalPending {
		t.Errorf("Status = %q, want %q", status.Status, domain.ApprovalPending)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/approval/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement the bridge**

Create `src/approval/bridge.go`:

```go
package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
)

// Broadcaster pushes messages to connected dashboard clients.
type Broadcaster interface {
	Broadcast(msg []byte)
}

// ApprovalRequest contains the data needed to create an approval.
type ApprovalRequest struct {
	Type    domain.ApprovalType
	Epic    string
	Story   string
	Content string
}

// VerdictResult is returned when an approval is decided.
type VerdictResult struct {
	ID      string
	Verdict domain.Verdict
	Comment string
}

// Bridge coordinates approvals between MCP tools and HTTP handlers.
type Bridge struct {
	store       *db.ApprovalStore
	broadcaster Broadcaster
	mu          sync.Mutex
	pending     map[string]chan VerdictResult
}

func NewBridge(store *db.ApprovalStore, b Broadcaster) *Bridge {
	return &Bridge{
		store:       store,
		broadcaster: b,
		pending:     make(map[string]chan VerdictResult),
	}
}

// Request creates an approval and blocks until a verdict arrives.
func (b *Bridge) Request(ctx context.Context, req ApprovalRequest) (*VerdictResult, error) {
	// Resolve epic/story IDs via store as needed
	approval, err := b.store.Request(req.Type, nil, nil, req.Content)
	if err != nil {
		return nil, fmt.Errorf("create approval: %w", err)
	}

	ch := make(chan VerdictResult, 1)
	b.mu.Lock()
	b.pending[approval.ID] = ch
	b.mu.Unlock()

	// Broadcast to dashboard
	msg, _ := json.Marshal(map[string]any{
		"type":    "approval_requested",
		"payload": map[string]any{"id": approval.ID, "approval_type": req.Type, "epic": req.Epic},
	})
	b.broadcaster.Broadcast(msg)

	select {
	case v := <-ch:
		return &v, nil
	case <-ctx.Done():
		b.mu.Lock()
		delete(b.pending, approval.ID)
		b.mu.Unlock()
		return nil, ctx.Err()
	}
}

// Decide delivers a verdict and unblocks the pending Request.
func (b *Bridge) Decide(id string, verdict domain.Verdict, comment string) error {
	_, err := b.store.Verdict(id, verdict, comment)
	if err != nil {
		return err
	}

	b.mu.Lock()
	ch, ok := b.pending[id]
	if ok {
		delete(b.pending, id)
	}
	b.mu.Unlock()

	if ok {
		ch <- VerdictResult{ID: id, Verdict: verdict, Comment: comment}
	}

	// Broadcast to dashboard
	msg, _ := json.Marshal(map[string]any{
		"type":    "approval_decided",
		"payload": map[string]any{"id": id, "verdict": verdict, "comment": comment},
	})
	b.broadcaster.Broadcast(msg)

	return nil
}

// Status returns the current state of an approval.
func (b *Bridge) Status(id string) (*domain.Approval, error) {
	return b.store.GetByID(id)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./src/approval/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/approval/
git commit -m "feat: add ApprovalBridge with Broadcaster interface and Go channels"
```

---

### Task 8: HTTP server — hooks and health

**Files:**
- Create: `src/server/server.go`
- Create: `src/server/hooks.go`
- Create: `src/server/errors.go`
- Test: `src/server/hooks_test.go`

**Step 1: Write the failing tests**

Create `src/server/hooks_test.go`:

```go
package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

func testServer(t *testing.T) *server.Server {
	t.Helper()
	database := dbtest.Open(t)
	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	return server.New(database, bridge, hub)
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

func TestAgentStartHook(t *testing.T) {
	srv := testServer(t)

	body := `{"session_id":"s1","agent_name":"code-agent","agent_type":"code"}`
	req := httptest.NewRequest("POST", "/hooks/agent-start", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status %d, want 200", rec.Code)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/server/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement server and hooks**

Create `src/server/errors.go` with `writeAPIError` and `writeJSON` helpers.

Create `src/server/hooks.go` with `HookHandler` struct and methods for each hook endpoint.

Create `src/server/server.go`:

```go
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/ws"
)

type Server struct {
	mux      *http.ServeMux
	database *db.DB
}

func New(database *db.DB, bridge *approval.Bridge, hub *ws.Hub) *Server {
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

	return &Server{mux: mux, database: database}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./src/server/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/server/
git commit -m "feat: add HTTP server with hook endpoints and health check"
```

---

### Task 9: HTTP server — REST API endpoints

**Files:**
- Create: `src/server/api.go`
- Test: `src/server/api_test.go`

**Step 1: Write the failing tests**

Create `src/server/api_test.go` with tests for:
- `GET /api/epics` — returns list of epics
- `GET /api/approvals` — returns list of approvals
- `GET /api/approvals?status=pending` — filters pending only
- `POST /api/approvals/{id}/verdict` — submits verdict, returns updated approval

**Step 2: Run test to verify it fails**

Run: `go test ./src/server/ -run TestAPI -v`

**Step 3: Implement API handlers**

Create `src/server/api.go` with the `APIHandler` struct and handler methods. Each method calls the corresponding store method and returns JSON.

The verdict handler calls `bridge.Decide()` which writes to SQLite and unblocks any pending MCP request.

**Step 4: Run tests to verify they pass**

Run: `go test ./src/server/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/server/api.go src/server/api_test.go
git commit -m "feat: add REST API endpoints for dashboard"
```

---

### Task 10: MCP server

**Files:**
- Create: `src/mcp/server.go`
- Test: `src/mcp/server_test.go`

**Step 1: Write the failing tests**

Create `src/mcp/server_test.go` testing the handler methods directly (not over stdio):
- `handleRequestApproval` — creates approval, verify it appears in SQLite
- `handleApprovalStatus` — returns current status of an approval
- Validation: missing required fields return error

**Step 2: Run test to verify it fails**

Run: `go test ./src/mcp/ -v`
Expected: FAIL

**Step 3: Implement MCP server**

Create `src/mcp/server.go`:

```go
package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
)

type Server struct {
	bridge *approval.Bridge
	store  *db.ApprovalStore
}

func New(bridge *approval.Bridge, store *db.ApprovalStore) *Server {
	return &Server{bridge: bridge, store: store}
}

func (s *Server) ServeStdio(ctx context.Context) error {
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "oraculo"}, nil)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "request_approval",
		Description: "Submit an artifact for human review. Blocks until a verdict is received.",
	}, s.handleRequestApproval)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "approval_status",
		Description: "Check the current status of an approval request.",
	}, s.handleApprovalStatus)

	return srv.Run(ctx, &mcpsdk.StdioTransport{})
}
```

Implement `handleRequestApproval` and `handleApprovalStatus` as methods on `*Server`.

**Step 4: Run tests to verify they pass**

Run: `go test ./src/mcp/ -v`
Expected: PASS

**Step 5: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 6: Commit**

```bash
git add src/mcp/
git commit -m "feat: add MCP server with request_approval and approval_status tools"
```

---

### Task 11: `oraculo start` command

**Files:**
- Create: `src/cli/start.go`
- Modify: `src/cli/root.go` (add start command)

**Step 1: Implement start command**

Create `src/cli/start.go`:

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/config"
	"github.com/lucas/oraculo/src/db"
	mcpserver "github.com/lucas/oraculo/src/mcp"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Oraculo (MCP + HTTP + WebSocket)",
		Long:  "Starts the MCP server on stdio and the HTTP/WebSocket server on the configured port. Launched by Claude Code as an MCP server.",
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	cfg, err := config.Read()
	if err != nil {
		return err
	}

	port := cfg.Port
	if port == 0 {
		port = 3100
	}

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	srv := server.New(database, bridge, hub)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database))

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		fmt.Fprintf(os.Stderr, "Oraculo HTTP server listening on :%d\n", port)
		return srv.ListenAndServe(ctx, port)
	})
	g.Go(func() error { return mcpSrv.ServeStdio(ctx) })

	return g.Wait()
}
```

**Step 2: Add to root command**

In `src/cli/root.go`, add `newStartCmd()` to the root command's `AddCommand` call.

**Step 3: Verify build**

Run: `make build`
Expected: Build succeeds

**Step 4: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 5: Commit**

```bash
git add src/cli/start.go src/cli/root.go
git commit -m "feat: add oraculo start command (MCP + HTTP + WebSocket)"
```

---

### Task 12: `oraculo install` command

**Files:**
- Modify: `src/cli/install.go`
- Test: `src/cli/install_test.go`

**Step 1: Write the failing tests**

Create `src/cli/install_test.go` (in `package cli_test` if following existing patterns, or use the `executeCmd` test helper from existing tests):

Tests should verify:
- Running `oraculo install` creates `.oraculo/config.json` with a port
- Running `oraculo install` creates `.claude/settings.json` with hooks and MCP config
- Running `oraculo install` creates `.oraculo/oraculo.db`
- Port in settings.json matches port in config.json

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestInstall -v`

**Step 3: Implement install**

Replace the stub in `src/cli/install.go` with the full implementation:
1. Create `.oraculo/` directory
2. Open and close DB (triggers migration)
3. Find available port via `config.FindPort`
4. Write `config.Write(&config.Config{Port: port})`
5. Generate and write `.claude/settings.json`
6. Copy skills (skip if `claude-kit/skills/oraculo/` is empty)

**Step 4: Run tests to verify they pass**

Run: `go test ./src/cli/ -run TestInstall -v`
Expected: PASS

**Step 5: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 6: Commit**

```bash
git add src/cli/install.go src/cli/install_test.go
git commit -m "feat: implement oraculo install command"
```

---

### Task 13: `oraculo uninstall` command

**Files:**
- Create: `src/cli/uninstall.go`
- Modify: `src/cli/root.go` (add uninstall command)
- Test: `src/cli/uninstall_test.go`

**Step 1: Write the failing tests**

Tests should verify:
- After install + uninstall, `.claude/settings.json` no longer has Oraculo entries
- `.oraculo/` directory is preserved (without `--purge`)
- With `--purge`, `.oraculo/` is removed

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestUninstall -v`

**Step 3: Implement uninstall**

Create `src/cli/uninstall.go`:
1. Read `.claude/settings.json`
2. Remove Oraculo hook entries and MCP server config
3. Write back (or remove if empty)
4. Remove `.claude/skills/oraculo/`
5. If `--purge`: remove `.oraculo/` entirely

**Step 4: Run tests to verify they pass**

Run: `go test ./src/cli/ -run TestUninstall -v`
Expected: PASS

**Step 5: Run all tests**

Run: `make test`
Expected: All tests pass

**Step 6: Commit**

```bash
git add src/cli/uninstall.go src/cli/root.go src/cli/uninstall_test.go
git commit -m "feat: add oraculo uninstall command"
```

---

### Task 14: Integration test

**Files:**
- Create: `src/integration_test.go` (or `tests/integration_test.go`)

**Step 1: Write an integration test**

Test the full approval flow in-process:
1. Create an in-memory DB
2. Create hub, bridge, server, MCP server
3. Seed an epic
4. In a goroutine, call `bridge.Request()` (simulating MCP tool call)
5. Wait for the approval to appear in the store
6. Submit verdict via HTTP POST to `/api/approvals/{id}/verdict`
7. Verify `bridge.Request()` returns the correct verdict

**Step 2: Run test**

Run: `go test ./src/ -run TestIntegration -v` (or appropriate path)
Expected: PASS

**Step 3: Run ALL tests**

Run: `make test`
Expected: All tests pass

**Step 4: Commit**

```bash
git add tests/ # or src/
git commit -m "test: add integration test for approval flow"
```

---

### Task 15: Final verification

**Step 1: Run full test suite**

Run: `make test`
Expected: All tests pass

**Step 2: Build binary**

Run: `make build`
Expected: Build succeeds

**Step 3: Verify binary has new commands**

Run: `./oraculo start --help`
Run: `./oraculo install --help`
Run: `./oraculo uninstall --help`
Expected: All three show usage information

**Step 4: Run vet and format checks**

Run: `make vet`
Expected: No issues
