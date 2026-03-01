# HTTP Hooks Integration — Design

## Context

Oraculo's dashboard currently relies on MCP notifications and file watchers for real-time data. This design introduces a **two-channel architecture** that uses Claude Code's native hook system for automatic telemetry (HTTP hooks) while preserving MCP for interactive approval gates. The result is reliable, fire-and-forget observability with zero agent-side instrumentation — the hooks fire automatically on every qualifying event.

---

## 1. Two-Channel Architecture

The Oraculo Go binary serves two communication channels in a single process:

```
┌─────────────────────────────────────────────────────────────────────┐
│                        oraculo (single Go binary)                   │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  CLI commands │  │  HTTP server │  │  MCP server  │              │
│  │  (existing)   │  │  (REST + WS) │  │  (stdio)     │              │
│  └──────────────┘  └──────┬───────┘  └──────┬───────┘              │
│                           │                  │                      │
│                     ┌─────┴──────────────────┴─────┐               │
│                     │        CLI internals          │               │
│                     │     (Trust Layer functions)    │               │
│                     └──────────────┬────────────────┘               │
│                                    │                                │
│                              ┌─────┴─────┐                         │
│                              │   SQLite   │                         │
│                              └───────────┘                         │
└─────────────────────────────────────────────────────────────────────┘

Channel 1: HTTP Hooks (automatic telemetry)
═══════════════════════════════════════════
Claude Code ──POST──> HTTP server ──> SQLite + WebSocket broadcast
                      (fire-and-forget, 200 empty body)

Channel 2: MCP (interactive approval gates)
═══════════════════════════════════════════
Claude Code ──stdio──> MCP server ──> SQLite + Go channel (blocks)
                                          │
Dashboard ──POST /api/approvals/:id/verdict──> UPDATE SQLite + send on channel
                                          │
                       MCP server <── Go channel ──> Claude Code (unblocks)
```

**Channel 1 — HTTP Hooks** handle automatic telemetry. Claude Code fires these hooks on system events (session start/end, agent start/stop, tool use). The HTTP server receives the event, persists metadata to SQLite, and broadcasts to connected WebSocket clients. These are fire-and-forget: the hook returns `200` with an empty body. If the server is unreachable, the hook fails silently — agents are never blocked by telemetry.

**Channel 2 — MCP** handles interactive approval gates. When an agent needs human approval, it calls the `request_approval` MCP tool. This blocks the agent until a verdict is received. The dashboard displays the artifact, the human decides, and the verdict flows back through the MCP tool response to unblock the agent.

**Why two channels:** HTTP hooks are automatic — they fire without any agent instrumentation. But they cannot block. MCP tools are interactive — they can block and return responses. But they require explicit agent calls. Telemetry needs automatic + non-blocking (HTTP hooks). Approvals need explicit + blocking (MCP). Each channel does what it does best.

---

## 2. Hook Endpoints

All HTTP hook endpoints accept POST requests, persist metadata to SQLite, broadcast to WebSocket clients, and return `200` with an empty body. Connection failures are non-blocking — if the server is offline, Claude Code logs a warning and continues.

All HTTP hooks are configured with `timeout: 5` (5 seconds).

### 2.1 Endpoint Summary

| Endpoint | Claude Code Hook | Matcher | Purpose |
|---|---|---|---|
| `POST /hooks/session-start` | Command hook (via `oraculo hook session-start`) | — | Register new session |
| `POST /hooks/agent-start` | `SubagentStart` | — | Agent spawned |
| `POST /hooks/agent-stop` | `SubagentStop` | — | Agent completed/failed |
| `POST /hooks/tool-used` | `PostToolUse` | `Bash\|Edit\|Write\|NotebookEdit` | Mutation event |
| `POST /hooks/task-completed` | `TaskCompleted` | — | Task finished |
| `POST /hooks/stop` | `Stop` | — | Agent stopping |
| `POST /hooks/teammate-idle` | `TeammateIdle` | — | Teammate idle |
| `POST /hooks/session-end` | `SessionEnd` | — | Session ended |

### 2.2 Endpoint Details

#### POST /hooks/session-start

**Trigger:** Command hook executing `oraculo hook session-start` (see Section 4.2).

Unlike the other hooks (which are HTTP hooks called by Claude Code directly), this is a **command hook** — Claude Code executes the Go binary as a subprocess. The `oraculo hook session-start` command:

1. Performs a health check on the HTTP server
2. If the server is offline, prints a warning to stderr (visible to the user)
3. If the server is online, sends session metadata via HTTP

**Payload persisted:**
- `session_id` — Unique identifier for this Claude Code session
- `model` — The model being used
- `cwd` — Working directory
- `started_at` — Timestamp

#### POST /hooks/agent-start

**Trigger:** `SubagentStart` hook — fires when the orchestrator spawns a subagent.

**Payload persisted:**
- `session_id` — Parent session
- `agent_name` — Agent identifier
- `agent_type` — Inferred type (code, qa, research, orchestrator)
- `started_at` — Timestamp

**WebSocket broadcast:** `{ "type": "agent_started", "payload": { ... } }`

#### POST /hooks/agent-stop

**Trigger:** `SubagentStop` hook — fires when a subagent completes or fails.

**Payload persisted:**
- `session_id` — Parent session
- `agent_name` — Agent identifier
- `status` — Final status
- `stopped_at` — Timestamp

**WebSocket broadcast:** `{ "type": "agent_stopped", "payload": { ... } }`

#### POST /hooks/tool-used

**Trigger:** `PostToolUse` hook with matcher `Bash|Edit|Write|NotebookEdit`.

Only mutation tools are tracked. Read-only tools (Read, Glob, Grep, WebFetch) are excluded to reduce noise and storage. **Only metadata is stored — no content, no diffs, no command text.**

**Payload persisted:**
- `session_id` — Parent session
- `tool_name` — Which tool was used (Bash, Edit, Write, NotebookEdit)
- `file_path` — For Edit/Write/NotebookEdit: the file path. For Bash: null (no command content stored)
- `timestamp` — When the tool was used

**WebSocket broadcast:** `{ "type": "tool_used", "payload": { ... } }`

#### POST /hooks/task-completed

**Trigger:** `TaskCompleted` hook — fires when Claude Code marks a task as completed.

**Payload persisted:** Metadata about the task completion event.

**WebSocket broadcast:** `{ "type": "task_completed", "payload": { ... } }`

#### POST /hooks/stop

**Trigger:** `Stop` hook — fires when an agent is stopping.

**Payload persisted:** Stop event metadata.

**WebSocket broadcast:** `{ "type": "agent_stopping", "payload": { ... } }`

#### POST /hooks/teammate-idle

**Trigger:** `TeammateIdle` hook — fires when a teammate agent becomes idle.

**Payload persisted:** Idle event metadata.

**WebSocket broadcast:** `{ "type": "teammate_idle", "payload": { ... } }`

#### POST /hooks/session-end

**Trigger:** `SessionEnd` hook — fires when the Claude Code session ends.

**Payload persisted:**
- `session_id` — Session being ended
- `ended_at` — Timestamp
- `end_reason` — Why the session ended

**WebSocket broadcast:** `{ "type": "session_ended", "payload": { ... } }`

### 2.3 Error Handling

All HTTP hook endpoints follow a **graceful degradation** policy:

- Server online: persist to SQLite, broadcast to WebSocket, return `200 {}`
- Server offline: Claude Code logs a non-blocking warning, agent continues unaffected
- Server error (5xx): Return error, Claude Code treats as non-blocking failure
- Timeout (>5s): Claude Code cancels the hook, agent continues

The hooks never block agents. Telemetry is best-effort — the system operates correctly without it.

---

## 3. MCP Approval Gates

The MCP server exposes two tools for approval workflows. These are the only blocking operations in the system — an agent that requests approval waits until a human responds.

### 3.1 MCP Tools

| MCP Tool | Purpose | Blocking |
|---|---|---|
| `request_approval` | Submit artifact for human review | Yes — blocks until verdict |
| `approval_status` | Poll for verdict (reconnection fallback) | No — returns current status |

### 3.2 request_approval

Called by the orchestrator when a workflow reaches an approval gate.

**Input:**
```json
{
  "type": "epic-requirements | story-definition | execution-plan | qa-escalation",
  "epic": "epic-name",
  "story": "story-name (optional)",
  "content": "markdown content of the artifact"
}
```

**Flow:**

```
Agent calls request_approval
        │
        ▼
MCP server: INSERT INTO approvals (status='pending') → SQLite
        │
        ▼
MCP server: create Go channel for this approval ID
        │
        ▼
HTTP server: broadcast WebSocket { "type": "approval_requested", ... }
        │
        ▼
Dashboard: renders artifact for human review
        │
        ▼
Human clicks Approve / Reject / Needs Revision
        │
        ▼
Dashboard: POST /api/approvals/:id/verdict { "verdict": "approved", "comment": "..." }
        │
        ▼
HTTP server: UPDATE approvals SET status=:verdict, comment=:comment → SQLite
        │
        ▼
HTTP server: send verdict on Go channel
        │
        ▼
MCP server: receives from Go channel, responds to agent
        │
        ▼
Agent unblocks with verdict
```

**Output (returned to agent):**
```json
{
  "id": "approval-uuid",
  "verdict": "approved | rejected | needs_revision",
  "comment": "Human's feedback (optional)"
}
```

### 3.3 approval_status

A non-blocking polling fallback for crash recovery. If an agent disconnects and reconnects, it can check whether a pending approval has already been decided.

**Input:**
```json
{
  "id": "approval-uuid"
}
```

**Output:**
```json
{
  "id": "approval-uuid",
  "status": "pending | approved | rejected | needs_revision",
  "comment": "Human's feedback (if decided)"
}
```

### 3.4 Crash Recovery

SQLite is the source of truth for approval state, not in-memory channels.

**Scenario: Server crashes while an approval is pending.**
1. The approval record exists in SQLite with `status = 'pending'`
2. When the server restarts, there is no in-memory Go channel for this approval
3. If the human submits a verdict before the agent reconnects, it is persisted to SQLite
4. When the agent reconnects and calls `approval_status`, it reads the verdict from SQLite
5. If the agent calls `request_approval` again for the same artifact, the MCP server can detect the duplicate and return the existing verdict

**Scenario: Agent crashes while an approval is pending.**
1. The approval record exists in SQLite with `status = 'pending'`
2. The human can still submit a verdict through the dashboard (persisted to SQLite)
3. When the agent restarts and reaches the same approval gate, it calls `approval_status` first
4. If a verdict exists, the agent uses it and continues without creating a duplicate request

---

## 4. `oraculo install`

The `oraculo install` command configures a project for Oraculo by setting up the infrastructure, hooks, and MCP server registration.

### 4.1 What It Does

1. **Creates `.oraculo/` directory** — Project infrastructure root
2. **Initializes SQLite** — Creates `.oraculo/oraculo.db` with the full schema (see Section 5)
3. **Allocates port** — Finds the first available port in range 3100-3199, saves to `.oraculo/config.json`
4. **Writes `.claude/settings.json`** — Registers all hooks and the MCP server
5. **Copies skills** — Copies `claude-kit/skills/oraculo/` to `.claude/skills/oraculo/`

### 4.2 Generated `.claude/settings.json`

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
        "url": "http://localhost:3100/hooks/agent-start",
        "timeout": 5
      }
    ],
    "SubagentStop": [
      {
        "type": "http",
        "url": "http://localhost:3100/hooks/agent-stop",
        "timeout": 5
      }
    ],
    "PostToolUse": [
      {
        "type": "http",
        "url": "http://localhost:3100/hooks/tool-used",
        "matcher": "Bash|Edit|Write|NotebookEdit",
        "timeout": 5
      }
    ],
    "TaskCompleted": [
      {
        "type": "http",
        "url": "http://localhost:3100/hooks/task-completed",
        "timeout": 5
      }
    ],
    "Stop": [
      {
        "type": "http",
        "url": "http://localhost:3100/hooks/stop",
        "timeout": 5
      }
    ],
    "TeammateIdle": [
      {
        "type": "http",
        "url": "http://localhost:3100/hooks/teammate-idle",
        "timeout": 5
      }
    ],
    "SessionEnd": [
      {
        "type": "http",
        "url": "http://localhost:3100/hooks/session-end",
        "timeout": 5
      }
    ]
  },
  "mcpServers": {
    "oraculo": {
      "command": "oraculo",
      "args": ["mcp"],
      "env": {}
    }
  }
}
```

**Notes:**
- The port `3100` in the example is replaced with the actual allocated port during install
- `SessionStart` uses a **command hook** (not HTTP) because it needs to perform a health check and print a warning if the server is offline — HTTP hooks cannot output to the user
- All HTTP hooks use `timeout: 5` (5 seconds)
- The `PostToolUse` hook uses a matcher to restrict to mutation tools only

### 4.3 SessionStart as Command Hook

The `SessionStart` hook is the only command hook. It runs `oraculo hook session-start`, which is a subcommand of the Go binary (not a shell script). This command:

1. Reads the project port from `.oraculo/config.json`
2. Sends a health check HTTP request to `http://localhost:<port>/health`
3. If the server responds: sends session-start metadata to `/hooks/session-start`
4. If the server is unreachable: prints a warning to stderr:
   ```
   ⚠ Oraculo dashboard is offline. Run 'oraculo server' to start it.
   ```
5. Always exits with code 0 (never blocks the session)

This is implemented as a Go binary subcommand rather than a shell script to maintain the zero-dependency distribution principle — no shell compatibility issues, no path resolution problems, no external dependencies.

### 4.4 `oraculo uninstall`

Removes Oraculo configuration from the project:

- Removes `.claude/settings.json` hook entries and MCP server config
- Removes `.claude/skills/oraculo/`
- **Preserves** `.oraculo/` directory (database, config, knowledge)
- With `--purge` flag: also removes `.oraculo/` entirely

---

## 5. SQLite Schema

### 5.1 New Tables

These tables are added to the existing `.oraculo/oraculo.db` schema to support the HTTP hooks telemetry data.

#### sessions

Tracks Claude Code sessions observed via hooks.

```sql
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,          -- Session identifier from Claude Code
    model       TEXT,                      -- Model being used (e.g. claude-opus-4-6)
    cwd         TEXT,                      -- Working directory
    started_at  TEXT NOT NULL,             -- ISO 8601 timestamp
    ended_at    TEXT,                      -- ISO 8601 timestamp (null while active)
    end_reason  TEXT                       -- Why the session ended
);
```

#### agents

Tracks agent lifecycle events from SubagentStart/Stop hooks.

```sql
CREATE TABLE agents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    name        TEXT NOT NULL,             -- Agent name/identifier
    type        TEXT NOT NULL,             -- code, qa, research, orchestrator
    status      TEXT NOT NULL DEFAULT 'active',  -- active, completed, failed
    started_at  TEXT NOT NULL,             -- ISO 8601 timestamp
    stopped_at  TEXT                       -- ISO 8601 timestamp (null while active)
);

CREATE INDEX idx_agents_session_id ON agents(session_id);
```

#### tool_events

Tracks mutation tool usage from PostToolUse hooks. Only metadata — no content, no diffs, no commands.

```sql
CREATE TABLE tool_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    tool_name   TEXT NOT NULL,             -- Bash, Edit, Write, NotebookEdit
    file_path   TEXT,                      -- File path (null for Bash)
    timestamp   TEXT NOT NULL              -- ISO 8601 timestamp
);

CREATE INDEX idx_tool_events_session_id ON tool_events(session_id);
CREATE INDEX idx_tool_events_timestamp ON tool_events(timestamp);
```

### 5.2 Existing Table: approvals

The `approvals` table already exists in the schema. This design confirms its structure and clarifies the fields used by the MCP approval gate flow.

```sql
CREATE TABLE approvals (
    id          TEXT PRIMARY KEY,          -- UUID
    type        TEXT NOT NULL,             -- epic-requirements, story-definition,
                                          -- execution-plan, qa-escalation
    epic        TEXT NOT NULL,             -- Epic name
    story       TEXT,                      -- Story name (null for epic-level approvals)
    content     TEXT NOT NULL,             -- Markdown content of the artifact
    status      TEXT NOT NULL DEFAULT 'pending',  -- pending, approved, rejected, needs_revision
    comment     TEXT,                      -- Human feedback (null until decided)
    created_at  TEXT NOT NULL,             -- ISO 8601 timestamp
    decided_at  TEXT                       -- ISO 8601 timestamp (null until decided)
);
```

### 5.3 Schema Notes

- All timestamps use ISO 8601 format (`2026-02-28T14:30:00Z`)
- Foreign keys reference `sessions(id)` — if a session record is missing (e.g., server was offline during session start), tool events and agent records for that session are still insertable with a synthetic session ID
- The `agents` table uses an auto-incrementing integer ID (not the agent name) because the same agent name can appear across multiple sessions
- No content or diff data is stored in `tool_events` — this is a deliberate privacy and storage decision. The dashboard shows *what* was touched, not *what was written*

---

## 6. Dashboard Data Sources

This section maps each dashboard screen to its data source under the new two-channel architecture. HTTP hooks replace file watchers and MCP notifications as the primary real-time data source.

### 6.1 Data Flow Summary

| Screen | Primary Data Source | Update Mechanism |
|---|---|---|
| Landing (Epic Selection) | CLI: `epic list` | Polling on mount (unchanged) |
| Stories | CLI: `story list`, `story get`, `task list` | Polling + WebSocket `task_completed` |
| DAG View | CLI: `task list` (includes `depends_on`) | WebSocket push from `TaskCompleted` hook |
| Agent Monitor | `agents` table + WebSocket | WebSocket push from `SubagentStart`/`SubagentStop` hooks |
| Activity Feed | `tool_events` table + WebSocket | WebSocket push from `PostToolUse` hook |
| Approvals | `approvals` table + WebSocket | WebSocket push from MCP `request_approval` |
| QA Dashboard | CLI: `task list`, `task get` + WebSocket | Hybrid: initial load via CLI, updates via WebSocket |
| Knowledge Base | CLI: `memory search`, `memory domains` | On-demand (unchanged) |
| Sessions | `sessions` table | REST API + WebSocket push on session start/end |
| Settings | Local config file | Read on mount, write on save (unchanged) |

### 6.2 What Changes From the Current Design

**Before (file watcher + MCP):**
- Agent Monitor relied on MCP `notify_agent_state` calls (required agent-side instrumentation)
- DAG View relied on file watcher monitoring `.oraculo/oraculo.db-wal`
- No activity feed existed (no tool-level telemetry)
- No session tracking existed

**After (HTTP hooks + MCP):**
- Agent Monitor receives events automatically from `SubagentStart`/`SubagentStop` hooks — no agent instrumentation required
- DAG View receives task completion events from `TaskCompleted` hook — more reliable than WAL file watching
- Activity Feed shows real-time tool mutations from `PostToolUse` hook — new capability
- Sessions screen shows session history from `sessions` table — new capability
- Approvals remain on MCP (blocking workflow unchanged)

**Key change:** The `notify_agent_state` MCP tool is **replaced** by HTTP hooks. Agents no longer need to call MCP tools for telemetry — Claude Code fires hooks automatically. The MCP server's tool set is reduced to:

| MCP Tool | Status |
|---|---|
| `request_approval` | **Kept** — interactive approval gates |
| `approval_status` | **Kept** — polling fallback for crash recovery |
| `notify_agent_state` | **Removed** — replaced by SubagentStart/Stop HTTP hooks |
| `register_project` | **Removed** — handled by `oraculo install` |

---

## 7. Server Architecture

### 7.1 Unified Process

The Oraculo Go binary runs everything in a single process:

```
oraculo server
    ├── HTTP server (net/http)
    │   ├── /hooks/*          ← Hook endpoints (fire-and-forget)
    │   ├── /api/*            ← REST API for dashboard
    │   └── /                 ← Embedded frontend (static assets)
    │
    ├── WebSocket server (gorilla/websocket or nhooyr.io/websocket)
    │   └── /ws               ← Real-time push to browser
    │
    └── MCP server (stdio)
        ├── request_approval  ← Blocking approval tool
        └── approval_status   ← Polling fallback
```

The MCP server runs on stdio (standard MCP transport, launched by Claude Code). The HTTP server and WebSocket server run on the allocated port. All three share:

- **CLI internals** — In-process function calls to the same Go functions that CLI commands use
- **SQLite connection** — Single database connection pool shared across all servers
- **WebSocket broadcast channel** — Go channel that all servers can write to for pushing events to the browser

### 7.2 Startup Sequence

When Claude Code starts a session:

1. Claude Code launches `oraculo mcp` as a registered MCP server (from `.claude/settings.json`)
2. `oraculo mcp` reads the port from `.oraculo/config.json`
3. If no other instance is running on that port, it starts the HTTP + WebSocket servers
4. If another instance is already running (port occupied by Oraculo), it connects to it
5. The MCP server begins handling stdio communication with Claude Code
6. The `SessionStart` command hook fires, running `oraculo hook session-start`
7. The health check confirms the server is online; session metadata is sent

### 7.3 Port Allocation

- **Range:** 3100-3199 (hardcoded constants in the Go binary)
- **Allocation:** `oraculo install` finds the first available port in the range
- **Persistence:** Saved to `.oraculo/config.json` as `dashboard.port`
- **Stability:** Each project gets a dedicated, stable port — the dashboard URL is bookmarkable
- **Conflict resolution:** If the configured port is occupied by a non-Oraculo process, sequential scan within the range finds the next free port; the new port is saved back to config

---

## 8. Future: PreToolUse Policy Enforcement

This section documents a **designed but not implemented** capability for future iterations.

### 8.1 Concept

The `PreToolUse` hook fires *before* a tool executes. Claude Code can block the tool execution based on the hook's response. This enables **policy enforcement from the dashboard** — a human could define rules that prevent certain operations in real-time.

### 8.2 Example Policies

- Block file writes to protected directories (e.g., `migrations/`, `config/`)
- Block Bash commands matching dangerous patterns (e.g., `rm -rf`, `DROP TABLE`)
- Require explicit approval before modifying critical files
- Rate-limit tool executions per agent

### 8.3 Why Deferred

Policy enforcement adds complexity to the critical path. A PreToolUse hook that blocks tool execution must be fast and reliable — any latency or failure directly impacts agent productivity. The current priority is reliable telemetry (PostToolUse) and approval gates (MCP). Policy enforcement will be implemented after the two-channel architecture is validated in production.

---

## Summary of Changes to Existing Documents

This section lists what each existing document needs to update to reflect the HTTP hooks integration. Phase 2 agents should use this as their change manifest.

| Document | Changes Required |
|---|---|
| `docs/design.md` | Section 3.3 (Hooks): update from "automatic guardians" to two-channel architecture description; add HTTP hooks as telemetry channel alongside MCP for approvals; update Section 3.5 (Dashboard) to mention hook-based real-time data |
| `docs/agents/design.md` | Add new section on HTTP hooks integration; document that `notify_agent_state` MCP tool is replaced by automatic SubagentStart/Stop hooks; update runtime section to mention new SQLite tables (sessions, agents, tool_events) |
| `docs/agents/philosophy.md` | Update communication model — agents no longer need explicit MCP calls for telemetry; hooks fire automatically; the observation layer becomes truly non-intrusive |
| `docs/ui/design.md` | Section 2.3 (MCP Server): remove `notify_agent_state` and `register_project` tools; update Section 2.4 (Real-time Communication): replace file watcher source with HTTP hooks source; update Section 6 (Data Sources) table with new update mechanisms; add Sessions screen |
| `docs/ui/philosophy.md` | Section 4.4 (Real-Time Without Coupling): strengthen with HTTP hooks — telemetry is now truly fire-and-forget; the observation layer cannot interfere with execution even if the server crashes |
