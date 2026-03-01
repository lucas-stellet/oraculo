# HTTP + MCP Servers — Implementation Design

## Context

This document defines the implementation design for Oraculo's HTTP server, MCP server, and the `install`/`uninstall` commands. It builds on the architectural decisions in `docs/plans/2026-02-28-http-hooks-integration-design.md` and refines them with concrete implementation choices.

The CLI Trust Layer (28+ commands) is complete. This design covers what gets built on top of it: the servers that make real-time telemetry and approval gates work, and the install command that wires everything together.

---

## 1. Commands

### New Commands

| Command | Purpose | Audience |
|---|---|---|
| `oraculo all-in` | MCP stdio + HTTP + WebSocket in one process | Claude Code (configured in settings.json) |
| `oraculo serve` | HTTP + WebSocket only (dashboard without Claude Code) | Humans |
| `oraculo mcp` | MCP stdio only (debug/testing) | Developers |
| `oraculo install [--global\|--local]` | Configure hooks, MCP, copy skills | Humans |
| `oraculo uninstall [--purge]` | Remove configuration (preserve DB by default) | Humans |

### Command Relationship

```
oraculo all-in = oraculo mcp + oraculo serve (in one process)
```

Claude Code launches `oraculo all-in` as its MCP server. This is the primary path — one binary, one process, everything runs together. The separate `mcp` and `serve` commands exist for when a user needs a single piece: `serve` for viewing the dashboard without an active Claude Code session, `mcp` for debugging or testing MCP tools in isolation.

---

## 2. Package Structure

### New Packages

```
src/
├── server/
│   ├── server.go          # HTTP server setup, router, lifecycle
│   ├── hooks.go           # POST /hooks/* handlers (fire-and-forget telemetry)
│   ├── api.go             # REST API handlers for dashboard
│   └── ws.go              # WebSocket hub (broadcast, connection management)
├── mcp/
│   ├── mcp.go             # MCP server setup (go-sdk wiring)
│   └── tools.go           # request_approval + approval_status tool handlers
├── approval/
│   └── bridge.go          # Approval coordination (Go channels + SQLite polling)
└── config/
    └── config.go          # Config read/write (.oraculo/config.json)
```

### Modified Packages

```
src/
├── cli/
│   ├── root.go            # Add all-in, serve, mcp, uninstall commands
│   ├── allin.go           # oraculo all-in command handler
│   ├── serve.go           # oraculo serve command handler
│   ├── mcp_cmd.go         # oraculo mcp command handler
│   ├── install.go         # Replace stub with full implementation
│   ├── uninstall.go       # oraculo uninstall command handler
│   └── hook_session.go    # Extract configFile to config/ package
├── db/
│   └── migrations.go      # Add migration v3 (agents, tool_events tables)
└── domain/
    └── (add agent, tool_event types if needed)
```

### Dependency Rules

- `server/` depends on `db/`, `domain/`, `approval/`, `config/`
- `mcp/` depends on `db/`, `domain/`, `approval/`
- `approval/` is pure — Go channels + interface. No HTTP or MCP dependency.
- `config/` is pure — reads/writes `.oraculo/config.json`. No other dependency.
- `cli/` commands are thin wiring: parse flags → create dependencies → start servers.

### New External Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/coder/websocket` | WebSocket server (minimal, idiomatic, context-native) |
| `github.com/modelcontextprotocol/go-sdk` | Official MCP Go SDK (JSON-RPC over stdio) |

---

## 3. Server Lifecycle

### `oraculo all-in` (primary path, launched by Claude Code)

```
Claude Code starts "oraculo all-in"
       │
       ▼
  Open SQLite (.oraculo/oraculo.db)
       │
       ▼
  Read port from .oraculo/config.json
       │
       ▼
  Attempt to bind port
       ├── Success → start HTTP + WebSocket on port
       │   └── Create ApprovalBridge with Go channels (in-process)
       │
       └── Port occupied → check if Oraculo (GET /health)
           ├── Is Oraculo → use ApprovalBridge with SQLite polling
           └── Not Oraculo → try next port in range 3100-3199
               └── Save new port to config.json
       │
       ▼
  Start MCP server on stdio (go-sdk)
       │
       ▼
  Block until stdio closes (Claude Code terminates)
       │
       ▼
  Graceful shutdown: stop HTTP server, close WebSocket connections, close DB
```

### `oraculo serve` (standalone dashboard)

```
Open SQLite → Read port → Start HTTP + WebSocket → Block until SIGINT/SIGTERM
```

No MCP. Dashboard only.

### `oraculo mcp` (standalone MCP)

```
Open SQLite → Start MCP server on stdio → Block until stdio closes
```

No HTTP, no WebSocket. Approvals use SQLite polling if a separate `oraculo serve` is running, or block indefinitely if no server is available to serve the dashboard.

---

## 4. ApprovalBridge

The `ApprovalBridge` is the coordination layer between MCP tools (where agents request approvals) and HTTP handlers (where humans submit verdicts). Two implementations handle the in-process and cross-process cases.

### Interface

```go
// approval/bridge.go

type Bridge interface {
    // Request creates an approval in SQLite and blocks until a verdict arrives.
    // Returns the verdict and any human feedback.
    Request(ctx context.Context, req ApprovalRequest) (Verdict, error)

    // Decide receives a verdict from the dashboard and unblocks the pending Request.
    Decide(id string, verdict Verdict) error

    // Status returns the current state of an approval (non-blocking polling fallback).
    Status(id string) (ApprovalStatus, error)
}
```

### In-Process Implementation (common case)

Used when `oraculo all-in` successfully binds the port (MCP + HTTP in same process).

- `Request`: inserts approval in SQLite → broadcasts WebSocket notification → creates `chan Verdict` → blocks on channel
- `Decide`: writes verdict to SQLite → sends on channel → blocked `Request` unblocks
- Zero latency between verdict submission and agent unblock.

Internal state: `sync.Map` of `approval_id → chan Verdict`. Channels are created on `Request` and cleaned up after verdict or context cancellation.

### Cross-Process Implementation (fallback)

Used when `oraculo all-in` detects another Oraculo instance on the port, or when `oraculo mcp` runs standalone.

- `Request`: inserts approval in SQLite → POSTs to running HTTP server at `POST /internal/approvals/notify` → polls SQLite every 500ms for verdict
- `Decide`: not called (the HTTP server's own in-process bridge handles it)
- `Status`: reads directly from SQLite

The 500ms polling interval is imperceptible for human approvals (which take seconds to minutes).

### Crash Recovery

SQLite is the source of truth, not in-memory channels.

- **Server crashes while approval pending**: Verdict persisted in SQLite by the time server restarts. Agent calls `approval_status` to check.
- **Agent crashes while approval pending**: Human can still submit verdict via dashboard (persisted to SQLite). When agent restarts and calls `request_approval` again, bridge detects existing verdict and returns it without creating a duplicate.

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
| `POST /api/approvals/:id/verdict` | Submit verdict (human-in-the-loop) |
| `GET /api/sessions` | List active/recent sessions |
| `GET /api/agents?session=<id>` | List agents for a session |
| `GET /api/activity?session=<id>` | List recent tool events |
| `GET /health` | Health check (returns `{"status": "ok"}`) |

### Internal Endpoints

| Endpoint | Purpose |
|---|---|
| `POST /internal/approvals/notify` | Cross-process: MCP notifies server of new approval |

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

### API Design Rules

- REST API endpoints consume the Trust Layer: they call the same `db.*Store` methods that CLI commands use. No direct SQLite queries.
- Hook endpoints use dedicated stores for the new telemetry tables (`agents`, `tool_events`).
- All errors return `{"error": "<code>", "message": "..."}` with appropriate HTTP status codes (400, 404, 500).

---

## 6. MCP Tools

The MCP server exposes exactly 2 tools. Both use the `ApprovalBridge` for coordination.

### `request_approval` (blocking)

**Input:**
```json
{
  "type": "epic-requirements | story-definition | execution-plan | qa-escalation",
  "epic": "epic-name",
  "story": "story-name (optional)",
  "content": "markdown artifact"
}
```

**Output (when human decides):**
```json
{
  "id": "approval-uuid",
  "verdict": "approved | rejected | needs_revision",
  "comment": "human feedback (optional)"
}
```

**Flow:**
1. MCP handler calls `bridge.Request(ctx, req)`
2. Bridge inserts approval in SQLite, broadcasts to WebSocket, blocks
3. Human sees approval in dashboard, submits verdict
4. HTTP handler calls `bridge.Decide(id, verdict)`
5. Bridge writes verdict to SQLite, unblocks the pending Request
6. MCP handler returns verdict to agent

**Duplicate detection:** If `request_approval` is called for an approval that already exists with a verdict, return the existing verdict without creating a duplicate. This covers crash + retry scenarios.

### `approval_status` (non-blocking)

**Input:**
```json
{ "id": "approval-uuid" }
```

**Output:**
```json
{
  "id": "approval-uuid",
  "status": "pending | approved | rejected | needs_revision",
  "comment": "human feedback (if decided)"
}
```

Polling fallback for crash recovery. If an agent reconnects after a crash, it can check whether a pending approval has already been decided.

---

## 7. Schema Migration v3

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

## 8. `oraculo install`

### What It Does

1. Creates `.oraculo/` directory if it doesn't exist
2. Initializes SQLite with the full schema (runs all migrations)
3. Allocates a port (first available in range 3100-3199)
4. Saves port to `.oraculo/config.json`
5. Writes `.claude/settings.json` with hooks and MCP server configuration
6. Copies skills from `claude-kit/skills/oraculo/` to `.claude/skills/oraculo/`

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
      "args": ["all-in"],
      "env": {}
    }
  }
}
```

`<PORT>` is replaced with the allocated port during install.

### Config File Format

`.oraculo/config.json`:

```json
{
  "port": 3100
}
```

### Port Allocation Logic

1. Scan range 3100-3199
2. For each port, attempt TCP listen
3. First available port is allocated
4. If port is later found occupied by a non-Oraculo process, `oraculo all-in` scans for the next free port and updates config

---

## 9. `oraculo uninstall`

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

## 10. Decisions Log

| Decision | Choice | Rationale |
|---|---|---|
| WebSocket library | `github.com/coder/websocket` | Minimal, idiomatic, context-native, actively maintained by Coder |
| MCP library | `github.com/modelcontextprotocol/go-sdk` | Official SDK, maintained by MCP org + Google, spec-aligned |
| Process model | `all-in` embeds both servers | Non-developer users (PMs on Windows) need single-binary simplicity |
| Separate commands | `mcp`, `serve`, `all-in` | Each does exactly one thing; `all-in` is the configured default |
| Approval coordination | Go channels (in-process) + SQLite polling (cross-process) | Optimal latency in common case, resilient fallback |
| Telemetry FK | No FK on agents.session_id | Server may be offline during session start; events should not be rejected |
| Port range | 3100-3199 | Avoids common ports; 100 ports enough for multi-project setups |
| Polling interval | 500ms | Imperceptible for human approvals (seconds/minutes) |
