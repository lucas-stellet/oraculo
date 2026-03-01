# HTTP + MCP Servers — Implementation Design

## Context

This document defines the implementation design for Oraculo's HTTP server, MCP server, and the `install`/`uninstall` commands. It builds on the architectural decisions in `docs/plans/2026-02-28-http-hooks-integration-design.md` and refines them with concrete implementation choices.

The CLI Trust Layer (28+ commands) is complete. This design covers what gets built on top of it: the servers that make real-time telemetry and approval gates work, and the install command that wires everything together.

---

## 1. Commands

### New Commands

| Command | Purpose | Audience |
|---|---|---|
| `oraculo start` | MCP stdio + HTTP + WebSocket in one process | Claude Code (configured in settings.json) |
| `oraculo install [--global\|--local]` | Configure hooks, MCP, copy skills | Humans |
| `oraculo uninstall [--purge]` | Remove configuration (preserve DB by default) | Humans |

### Single Entry Point

There is exactly **one way** to start Oraculo: `oraculo start`. This command starts everything — MCP server on stdio, HTTP server on the configured port, WebSocket for real-time dashboard updates — in a single process.

```
Claude Code launches "oraculo start" (configured as MCP server in settings.json)
       │
       ▼
  Single process: MCP (stdio) + HTTP (port) + WebSocket (port)
```

This is an MVP simplification. Future versions may introduce separate `serve` (HTTP-only) and `mcp` (MCP-only) commands for advanced use cases, but for now a single command eliminates edge cases around cross-process coordination.

**Consequence:** Approvals always use Go channels in-process. No SQLite polling fallback needed. No cross-process bridge. No port-occupied detection for connecting to another Oraculo instance.

---

## 2. Package Structure

### New Packages

```
src/
├── server/
│   ├── server.go          # HTTP server setup, router, lifecycle
│   ├── hooks.go           # HookHandler struct with POST /hooks/* methods
│   ├── api.go             # APIHandler struct with REST API methods
│   └── errors.go          # HTTP error response helper (writeAPIError)
├── ws/
│   └── hub.go             # WebSocket hub: broadcast, connection pool, Run(ctx)
├── mcp/
│   └── server.go          # MCP server struct, tool handlers, ServeStdio(ctx)
├── approval/
│   └── bridge.go          # Bridge struct, Broadcaster interface, Go channels
├── config/
│   └── config.go          # Config struct, Read, Write (atomic), FindPort
└── dbtest/
    └── dbtest.go          # Shared test helper: Open(t) *db.DB (in-memory)
```

### Modified Packages

```
src/
├── cli/
│   ├── root.go            # Add start, uninstall commands
│   ├── start.go           # oraculo start command handler (errgroup wiring)
│   ├── install.go         # Replace stub with full implementation
│   ├── uninstall.go       # oraculo uninstall command handler
│   └── hook_session.go    # Replace inline configFile with config.Read()
├── db/
│   ├── db.go              # Export OpenMemory() for dbtest package
│   ├── migrations.go      # Add migration v3 (agents, tool_events tables)
│   ├── agent_store.go     # AgentStore: Start, Stop, ListBySession
│   └── tool_event_store.go # ToolEventStore: Record, ListBySession
└── domain/
    └── errors.go          # Add ErrApprovalDecided sentinel error
```

### Dependency Rules

- `server/` depends on `db/`, `domain/`, `approval/`, `ws/`, `config/`
- `mcp/` depends on `db/`, `domain/`, `approval/`
- `ws/` is standalone — no dependency on server, mcp, or approval
- `approval/` defines a `Broadcaster` interface; depends on `db/` only. `ws.Hub` satisfies the interface but is not imported.
- `config/` is pure — reads/writes `.oraculo/config.json`. No other dependency.
- `dbtest/` depends on `db/` only — test support package.
- `cli/` commands are thin wiring: parse flags → construct dependencies → start servers.

### New External Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/coder/websocket` | WebSocket server (minimal, idiomatic, context-native) |
| `github.com/modelcontextprotocol/go-sdk` | Official MCP Go SDK (JSON-RPC over stdio) |
| `golang.org/x/sync/errgroup` | Concurrent server lifecycle coordination |

---

## 3. Server Lifecycle

### `oraculo start` (launched by Claude Code)

```
Claude Code starts "oraculo start"
       │
       ▼
  Open SQLite (.oraculo/oraculo.db)
       │
       ▼
  Read port from config.Read()
       │
       ▼
  Create ws.Hub, approval.Bridge, server.Server, mcp.Server
       │
       ▼
  errgroup.WithContext(ctx):
    g.Go(hub.Run)           — WebSocket broadcast loop
    g.Go(srv.ListenAndServe) — HTTP server
    g.Go(mcpSrv.ServeStdio)  — MCP stdio loop
       │
       ▼
  g.Wait() — blocks until any goroutine exits or SIGINT/SIGTERM
       │
       ▼
  Context cancelled → all servers drain and exit
       │
       ▼
  Close DB
```

### Wiring Pattern

```go
// cli/start.go
func runStart(cmd *cobra.Command, _ []string) error {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    database, err := db.Open()
    if err != nil { return err }
    defer database.Close()

    cfg, err := config.Read()
    if err != nil { return err }

    hub    := ws.NewHub()
    bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
    srv    := server.New(database, bridge, hub, cfg.Port)
    mcpSrv := mcp.New(bridge, db.NewApprovalStore(database))

    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error { return hub.Run(ctx) })
    g.Go(func() error { return srv.ListenAndServe(ctx) })
    g.Go(func() error { return mcpSrv.ServeStdio(ctx) })
    return g.Wait()
}
```

Each component's `Run`/`ListenAndServe`/`ServeStdio` method accepts `context.Context` and returns `error`. When any goroutine exits (or ctx is cancelled by a signal), errgroup cancels the shared context and all others drain gracefully.

### Port Conflict Handling

If the configured port is occupied:

1. Try next available port in range 3100-3199
2. Save new port to `.oraculo/config.json` atomically
3. Log a warning to stderr: `"Port 3100 in use, using 3101 instead"`
4. Continue startup normally

---

## 4. ApprovalBridge

The `ApprovalBridge` coordinates approvals between MCP tools (where agents request) and HTTP handlers (where humans decide). Since everything runs in one process, this is always in-memory Go channels.

### Broadcaster Interface

The bridge needs to push notifications to connected dashboards but must not import `ws/` directly. A narrow interface decouples the packages and simplifies testing:

```go
// approval/bridge.go

// Broadcaster pushes messages to connected dashboard clients.
// ws.Hub satisfies this; tests use a stub.
type Broadcaster interface {
    Broadcast(msg []byte)
}
```

### Bridge Struct

```go
type Bridge struct {
    store       *db.ApprovalStore
    broadcaster Broadcaster
    mu          sync.Mutex
    pending     map[string]chan domain.Verdict
}

func NewBridge(store *db.ApprovalStore, b Broadcaster) *Bridge {
    return &Bridge{
        store:       store,
        broadcaster: b,
        pending:     make(map[string]chan domain.Verdict),
    }
}
```

`sync.Mutex` + plain map instead of `sync.Map`: the approval channel map has high churn (create, write once, delete) and low concurrency (one goroutine per approval). A mutex is clearer, safer, and avoids type assertions.

### Methods

```go
// Request creates an approval in SQLite, broadcasts to dashboard, and blocks
// until a verdict arrives or context is cancelled.
func (b *Bridge) Request(ctx context.Context, req ApprovalRequest) (domain.Verdict, error)

// Decide receives a verdict from the dashboard, persists to SQLite, and
// unblocks the pending Request.
func (b *Bridge) Decide(id string, verdict domain.Verdict, comment string) error

// Status returns the current state of an approval (non-blocking).
func (b *Bridge) Status(id string) (*domain.Approval, error)
```

### Flow

1. Agent calls MCP `request_approval` → MCP handler calls `bridge.Request()`
2. Bridge inserts approval in SQLite (`status = 'pending'`)
3. Bridge creates `chan domain.Verdict` (buffered, size 1) in `pending` map
4. Bridge broadcasts `{"type": "approval_requested", ...}` to WebSocket
5. Bridge blocks on channel via `select { case v := <-ch: ... case <-ctx.Done(): ... }`
6. Human sees approval in dashboard, submits verdict
7. Dashboard POSTs to `POST /api/approvals/{id}/verdict`
8. HTTP handler calls `bridge.Decide(id, verdict, comment)`
9. Bridge updates approval in SQLite, sends verdict on channel, deletes from map
10. `bridge.Request()` unblocks, returns verdict to MCP handler
11. MCP handler returns verdict to agent

### Lock Scope

The mutex is held only for brief map operations — never while blocking on a channel or writing to SQLite:

```go
func (b *Bridge) Request(ctx context.Context, req ApprovalRequest) (domain.Verdict, error) {
    // 1. Write to DB (no lock)
    approval, err := b.store.Request(...)
    if err != nil { return domain.Verdict{}, err }

    // 2. Register channel (brief lock)
    ch := make(chan domain.Verdict, 1)
    b.mu.Lock()
    b.pending[approval.ID] = ch
    b.mu.Unlock()

    // 3. Broadcast (no lock)
    b.broadcaster.Broadcast(...)

    // 4. Block on channel
    select {
    case v := <-ch:
        return v, nil
    case <-ctx.Done():
        b.mu.Lock()
        delete(b.pending, approval.ID)
        b.mu.Unlock()
        return domain.Verdict{}, ctx.Err()
    }
}
```

### Duplicate Detection

If `request_approval` is called for an approval that already exists with a verdict, return the existing verdict without blocking. This covers crash + retry scenarios.

### Context Cancellation

If the agent's context is cancelled (Claude Code terminates), the blocked `Request` returns `context.Canceled`. The approval remains pending in SQLite — a future session can pick it up via `approval_status`.

---

## 5. HTTP Endpoints

### Hook Endpoints (automatic telemetry)

All hook endpoints accept POST, persist metadata to SQLite, broadcast to WebSocket, and return `200 {}`. Connection failures are non-blocking.

| Endpoint | Claude Code Hook | Persists To |
|---|---|---|
| `POST /hooks/session-start` | Command hook (via `oraculo hook session-start`) | `claude_sessions` |
| `POST /hooks/agent-start` | `SubagentStart` | `agents` |
| `POST /hooks/agent-stop` | `SubagentStop` | `agents` (update) |
| `POST /hooks/tool-used` | `PostToolUse` (matcher: `Bash\|Edit\|Write\|NotebookEdit`) | `tool_events` |
| `POST /hooks/task-completed` | `TaskCompleted` | — (broadcast only) |
| `POST /hooks/stop` | `Stop` | — (broadcast only) |
| `POST /hooks/teammate-idle` | `TeammateIdle` | — (broadcast only) |
| `POST /hooks/session-end` | `SessionEnd` | `claude_sessions` (update) |

### REST API Endpoints (dashboard)

| Endpoint | Purpose |
|---|---|
| `GET /api/epics` | List epics |
| `GET /api/stories?epic=<name>` | List stories for an epic |
| `GET /api/tasks?epic=<name>&story=<name>` | List tasks for a story |
| `GET /api/approvals` | List approvals (optionally filter `?status=pending`) |
| `POST /api/approvals/{id}/verdict` | Submit verdict (human-in-the-loop) |
| `GET /api/sessions` | List active/recent sessions |
| `GET /api/agents?session={id}` | List agents for a session |
| `GET /api/activity?session={id}` | List recent tool events |
| `GET /health` | Health check (returns `{"status": "ok"}`) |

Path parameters use Go 1.22+ `{id}` syntax, accessed via `r.PathValue("id")`. No external router library needed.

### WebSocket

| Endpoint | Purpose |
|---|---|
| `GET /ws` | Upgrade to WebSocket for real-time push |

Broadcast message format (server → browser):

```json
{
  "type": "agent_started | agent_stopped | tool_used | task_completed | agent_stopping | teammate_idle | session_ended | approval_requested | approval_decided",
  "payload": { ... }
}
```

### Handler Design

HTTP handlers are methods on structs, not closures. This groups shared dependencies and makes handlers independently testable:

```go
// server/hooks.go
type HookHandler struct {
    sessions *db.ClaudeSessionStore
    agents   *db.AgentStore
    toolEvts *db.ToolEventStore
    hub      *ws.Hub
}

// server/api.go
type APIHandler struct {
    epics     *db.EpicStore
    stories   *db.StoryStore
    tasks     *db.TaskStore
    approvals *db.ApprovalStore
    bridge    *approval.Bridge
    hub       *ws.Hub
}
```

Router wiring in `server.go`:

```go
mux := http.NewServeMux()
mux.HandleFunc("POST /hooks/agent-start",          hook.handleAgentStart)
mux.HandleFunc("POST /hooks/agent-stop",           hook.handleAgentStop)
// ...
mux.HandleFunc("GET /api/approvals",               api.handleListApprovals)
mux.HandleFunc("POST /api/approvals/{id}/verdict",  api.handleVerdict)
mux.HandleFunc("GET /ws",                          hub.ServeWS)
mux.HandleFunc("GET /health",                      handleHealth)
```

### Error → HTTP Status Mapping

A dedicated `writeAPIError` function in `server/errors.go` maps domain sentinel errors to HTTP status codes. This is separate from `output.WriteError` (which is for CLI output without status codes):

| Domain Error | HTTP Status | JSON `error` code |
|---|---|---|
| `ErrNotFound` | 404 | `not_found` |
| `ErrInvalidTransition` | 409 | `conflict` |
| `ErrApprovalDecided` | 409 | `approval_decided` |
| `ErrAlreadyExists` | 409 | `already_exists` |
| `ErrMissingRequired` | 400 | `bad_request` |
| (other) | 500 | `internal_error` |

---

## 6. WebSocket Hub

The `ws.Hub` manages WebSocket connections and broadcasts. It runs as its own goroutine via errgroup.

```go
// ws/hub.go

type Hub struct {
    mu        sync.Mutex
    clients   map[*client]struct{}
    broadcast chan []byte           // buffered (64), non-blocking send
}

func NewHub() *Hub

// Run processes broadcasts until ctx is cancelled. Called by errgroup.
func (h *Hub) Run(ctx context.Context) error

// Broadcast queues a message. Non-blocking: drops if buffer is full.
// This ensures MCP handlers and HTTP handlers never block on slow WebSocket clients.
func (h *Hub) Broadcast(msg []byte)

// ServeWS upgrades an HTTP connection and registers the client.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request)
```

The buffered channel with non-blocking send on `Broadcast` is critical: callers (MCP handler, hook handler) must not block waiting for a slow WebSocket client. If the buffer is full, the message is dropped — telemetry is best-effort.

---

## 7. MCP Server

The MCP server is a struct with handler methods, mirroring the HTTP handler pattern.

```go
// mcp/server.go

type Server struct {
    bridge *approval.Bridge
    store  *db.ApprovalStore
}

func New(bridge *approval.Bridge, store *db.ApprovalStore) *Server

// ServeStdio starts the MCP server on stdin/stdout. Blocks until ctx is cancelled.
func (s *Server) ServeStdio(ctx context.Context) error
```

Internally, `ServeStdio` creates the go-sdk server, registers tools, and runs on `StdioTransport`:

```go
func (s *Server) ServeStdio(ctx context.Context) error {
    srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "oraculo"}, nil)
    mcpsdk.AddTool(srv, requestApprovalTool, s.handleRequestApproval)
    mcpsdk.AddTool(srv, approvalStatusTool,  s.handleApprovalStatus)
    return srv.Run(ctx, &mcpsdk.StdioTransport{})
}
```

### Tools

**`request_approval` (blocking):**

Input:
```json
{
  "type": "epic-requirements | story-definition | execution-plan | qa-escalation",
  "epic": "epic-name",
  "story": "story-name (optional)",
  "content": "markdown artifact"
}
```

Output (when human decides):
```json
{
  "id": "approval-uuid",
  "verdict": "approved | rejected | needs_revision",
  "comment": "human feedback (optional)"
}
```

**`approval_status` (non-blocking):**

Input: `{ "id": "approval-uuid" }`

Output: `{ "id": "...", "status": "pending|approved|rejected|needs_revision", "comment": "..." }`

---

## 8. Schema Migration v3

New migration adding telemetry tables for hook data.

```sql
-- Migration v3: Hook telemetry tables

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
```

**Design decisions:**
- No FK from `agents.session_id` to `claude_sessions.id` — if the server was offline during session start, we still want to accept agent events for that session.
- `agents.type` defaults to `'unknown'` — the type is inferred from the hook payload, and we don't want to reject events with missing type information.
- `tool_events` stores only metadata — no content, no diffs, no commands. Privacy and storage by design.

---

## 9. `oraculo install`

### What It Does

1. Creates `.oraculo/` directory if it doesn't exist
2. Initializes SQLite with the full schema (runs all migrations)
3. Allocates a port via `config.FindPort(3100, 3199)`
4. Saves port to `.oraculo/config.json` via `config.Write()` (atomic)
5. Writes `.claude/settings.json` with hooks and MCP server configuration
6. Copies skills from `claude-kit/skills/oraculo/` to `.claude/skills/oraculo/`

### Config Package API

```go
// config/config.go

type Config struct {
    Port int `json:"port"`
}

// Read loads .oraculo/config.json. Returns zero-value Config if file doesn't exist.
func Read() (*Config, error)

// Write persists cfg to .oraculo/config.json atomically (temp file + rename).
func Write(cfg *Config) error

// FindPort returns the first available TCP port in [start, end].
func FindPort(start, end int) (int, error)
```

### Flags

- `--global`: writes to `~/.claude/` instead of `.claude/` (all projects gain Oraculo)
- `--local` (default): writes to `.claude/` in the current project

### Generated `.claude/settings.json`

```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "oraculo hook session-start"
      }
    ],
    "SubagentStart": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/agent-start",
        "timeout": 5
      }
    ],
    "SubagentStop": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/agent-stop",
        "timeout": 5
      }
    ],
    "PostToolUse": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/tool-used",
        "matcher": "Bash|Edit|Write|NotebookEdit",
        "timeout": 5
      }
    ],
    "TaskCompleted": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/task-completed",
        "timeout": 5
      }
    ],
    "Stop": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/stop",
        "timeout": 5
      }
    ],
    "TeammateIdle": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/teammate-idle",
        "timeout": 5
      }
    ],
    "SessionEnd": [
      {
        "type": "http",
        "url": "http://localhost:<PORT>/hooks/session-end",
        "timeout": 5
      }
    ]
  },
  "mcpServers": {
    "oraculo": {
      "command": "oraculo",
      "args": ["start"],
      "env": {}
    }
  }
}
```

`<PORT>` is replaced with the allocated port during install.

---

## 10. `oraculo uninstall`

### What It Does

1. Removes Oraculo hook entries from `.claude/settings.json`
2. Removes MCP server config from `.claude/settings.json`
3. Removes `.claude/skills/oraculo/`
4. **Preserves** `.oraculo/` directory (database, config, knowledge, markdowns)

### Flags

- `--purge`: also removes `.oraculo/` entirely (destructive)

### Safety

- Warns before `--purge` with a confirmation prompt
- Only removes Oraculo-specific entries from settings.json — does not touch other hooks or MCP servers

---

## 11. New Domain Errors

Add to `domain/errors.go`:

```go
var ErrApprovalDecided = errors.New("approval already decided")
```

Used by `bridge.Decide` when a verdict is submitted for an approval that is not pending. Maps to HTTP 409 Conflict.

---

## 12. Test Patterns

### Shared Test Helper

`src/dbtest/dbtest.go` exports `Open(t) *db.DB` using an in-memory SQLite. Requires exposing `db.OpenMemory()` as a public function.

```go
// dbtest/dbtest.go
func Open(t *testing.T) *db.DB {
    t.Helper()
    database, err := db.OpenMemory()
    if err != nil { t.Fatalf("dbtest.Open: %v", err) }
    t.Cleanup(func() { database.Close() })
    return database
}
```

### HTTP Handler Tests

Use `net/http/httptest`:

```go
func TestHandleListApprovals(t *testing.T) {
    database := dbtest.Open(t)
    // ... seed data using stores ...
    srv := server.New(database, bridge, hub, 0)

    req := httptest.NewRequest("GET", "/api/approvals", nil)
    rec := httptest.NewRecorder()
    srv.ServeHTTP(rec, req)

    if rec.Code != 200 { t.Fatalf("status %d", rec.Code) }
}
```

### Bridge Tests

Use a `stubBroadcaster` that satisfies the `Broadcaster` interface:

```go
type stubBroadcaster struct {
    msgs [][]byte
    mu   sync.Mutex
}
func (b *stubBroadcaster) Broadcast(msg []byte) {
    b.mu.Lock()
    b.msgs = append(b.msgs, msg)
    b.mu.Unlock()
}
```

### MCP Tool Tests

Test handler methods directly (not over stdio):

```go
func TestHandleRequestApproval(t *testing.T) {
    database := dbtest.Open(t)
    bridge := approval.NewBridge(db.NewApprovalStore(database), &stubBroadcaster{})
    srv := mcp.New(bridge, db.NewApprovalStore(database))
    // call srv.handleRequestApproval directly
}
```

---

## 13. Decisions Log

| Decision | Choice | Rationale |
|---|---|---|
| WebSocket library | `github.com/coder/websocket` | Minimal, idiomatic, context-native, actively maintained by Coder |
| MCP library | `github.com/modelcontextprotocol/go-sdk` | Official SDK, maintained by MCP org + Google, spec-aligned |
| Process model | Single `oraculo start` command | MVP simplification: eliminates cross-process edge cases |
| Server coordination | `errgroup.WithContext` | Idiomatic Go: one context, all goroutines drain on cancellation |
| HTTP routing | Go 1.22+ stdlib `net/http` | `{id}` path params + method routing built in; no external router needed |
| Handler design | Struct methods (`APIHandler`, `HookHandler`) | Groups shared deps, independently testable, matches Go convention |
| Approval coordination | Go channels + `sync.Mutex` map | Single process; mutex beats `sync.Map` for high-churn low-concurrency |
| Bridge → WS coupling | `Broadcaster` interface | Decouples packages; tests use a stub |
| Config writes | Atomic (temp file + rename) | No corrupt config on crash |
| Test DB | `dbtest.Open(t)` shared helper | Avoids duplicating in-memory DB setup across packages |
| Telemetry FK | No FK on agents.session_id | Server may be offline during session start; events should not be rejected |
| Port range | 3100-3199 | Avoids common ports; 100 ports enough for multi-project setups |
| Future expansion | `serve` and `mcp` commands | Can be added later when advanced use cases emerge |
