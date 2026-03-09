# Oraculo UI — Design

## 1. Overview

The UI layer provides a browser-based dashboard for monitoring and interacting with Oraculo projects. It surfaces the data that the CLI Trust Layer manages — epics, stories, tasks, DAG dependencies, agent activity, QA verdicts, and accumulated knowledge — through a real-time web interface.

The dashboard never bypasses the CLI. Every data operation flows through the CLI's internal Go functions — the same validated, contracted operations that agents and humans use from the terminal. The UI is a read-heavy, write-light consumer: it reads frequently (polling + file watching), writes only for approvals and settings.

```
oraculo (single Go binary)
├── CLI commands          (already exists)
├── HTTP server           (serves REST API + embedded frontend assets)
├── WebSocket server      (real-time push to browser)
├── MCP server (stdio)    (communication with Claude Code)
└── Auto-open browser     (zero-friction startup)

Human (Browser) ←→ HTTP/WS ←→ Go HTTP server ←→ CLI internals ←→ SQLite + Markdown
Claude Code     ←→ stdio   ←→ Go MCP server  ←→ CLI internals ←→ SQLite + Markdown
Claude Code     ──POST──>  HTTP server ──> SQLite + broadcast WebSocket (HTTP hooks)
```

**Two-channel architecture:** The binary operates two communication channels. Channel 1 — HTTP Hooks — handles automatic telemetry: Claude Code fires hooks on system events (session start/end, agent start/stop, tool use), the HTTP server persists metadata to SQLite and broadcasts to WebSocket clients. These are fire-and-forget; hooks never block agents. Channel 2 — MCP — handles interactive approval gates: agents call `request_approval`, which blocks until a human responds through the dashboard. Each channel does what it does best: telemetry needs automatic and non-blocking (HTTP hooks), approvals need explicit and blocking (MCP).

The entire system — CLI, HTTP server, MCP server, and embedded frontend — lives in a single Go binary. The frontend is built at compile time (Next.js + shadcn/ui generates static assets) and embedded via Go's `embed.FS`. At runtime, there is no Node.js, no npm, no external dependencies.

## 2. Architecture

### 2.1 Dashboard Server

The Go binary includes an HTTP server that serves both the REST API and the embedded frontend assets. Each API endpoint calls CLI internals directly (in-process function calls, not subprocess spawning) and returns JSON to the frontend. The server holds no state of its own — the CLI and SQLite are the single source of truth.

| Endpoint pattern | CLI command(s) | Purpose |
|---|---|---|
| `GET /api/epics` | `oraculo tools epic list` | List all epics |
| `GET /api/epics/:name` | `oraculo tools epic get <name>` | Epic requirements markdown |
| `GET /api/epics/:epic/stories` | `oraculo tools story list --epic <epic>` | Stories for an epic |
| `GET /api/epics/:epic/stories/:story` | `oraculo tools story get <name> --epic <epic>` | Story requirements markdown |
| `GET /api/epics/:epic/stories/:story/tasks` | `oraculo tools task list --epic <epic> --story <story>` | Tasks with status and dependencies |
| `GET /api/epics/:epic/stories/:story/tasks/:task` | `oraculo tools task get <name> --epic <epic> --story <story>` | Single task detail with result |
| `GET /api/knowledge` | `oraculo tools memory search <query>` | Knowledge search |
| `GET /api/knowledge/domains` | `oraculo tools memory domains` | Available domains |
| `GET /api/status` | `oraculo status` (parsed) | Aggregate project statistics |
| `GET /api/sessions` | `sessions` table (SQLite) | Session history |
| `GET /api/sessions/:id/agents` | `agents` table (SQLite) | Agents for a session |
| `GET /api/sessions/:id/tool-events` | `tool_events` table (SQLite) | Tool events for a session |
| `GET /api/epics/:name/versions` | `oraculo tools epic versions <name>` | List epic versions |
| `GET /api/epics/:epic/stories/:story/versions` | `oraculo tools story versions <story> --epic <epic>` | List story versions |
| `GET /api/reviews/:versionId` | `oraculo tools review list <version-id> --type <epic\|story>` | List reviews for a version |
| `POST /api/reviews` | `oraculo tools review create <version-id> --type <type> --verdict <verdict> --comment <comment>` | Submit document review |
| `POST /api/approvals/:id/verdict` | (internal state) | Submit operational approval decision |

The server manages one stateful concern: the **approval queue**. Approval requests arrive via MCP tool calls from agents and are held in memory until the human responds through the dashboard. The verdict is then relayed back to the requesting agent.

For telemetry data (sessions, agents, tool events), the server queries SQLite tables directly — these records are inserted by HTTP hooks, not by CLI commands.

### 2.2 Startup Flow

When Claude Code starts a session, it launches `oraculo mcp` as a registered MCP server (configured during `oraculo install`). The startup sequence:

1. `oraculo mcp` starts the MCP server on stdio (standard MCP transport)
2. The MCP server reads the project's allocated port from `.oraculo/config` → `dashboard.port`
3. If no port is configured, it allocates the first available port in the range 3100-3199
4. The HTTP server attempts to bind the configured port
5. If the port is occupied, it performs a sequential scan within 3100-3199 for the first free port
6. The new port is saved back to `.oraculo/config`
7. If no port is available in the range, the server exits with a clear error: "All ports in range 3100-3199 are in use"
8. The HTTP server begins serving the embedded frontend and REST API
9. The binary opens the user's default browser to `http://localhost:<port>`

The `SessionStart` command hook (`oraculo hook session-start`) fires after the server is ready. It sends a health check to confirm the server is online, then sends session metadata to `/hooks/session-start`. If the server is offline, it prints a warning to stderr and exits with code 0 — the session is never blocked.

**Port System:**

- **Range:** 3100-3199 (hardcoded constants in the Go binary)
- **Configuration:** `.oraculo/config` per project stores `dashboard.port`
- **Stability:** Each project gets a dedicated, stable port — the dashboard URL is bookmarkable across sessions
- **No global config:** The port range lives in the binary, port allocation lives in the project

If the HTTP server is already running (from another session), the new MCP server connects to the existing instance instead of starting a duplicate. The human never needs to run a separate command.

### 2.3 MCP Server

The MCP server exposes tools that AI agents call during orchestration. It runs in the same process as the HTTP server and shares the WebSocket broadcast channel.

| MCP Tool | Direction | Purpose |
|---|---|---|
| `request_approval` | Agent -> Dashboard | Submit a document for human review (requirements, story, QA escalation) |
| `approval_status` | Agent -> Dashboard | Poll for verdict (non-blocking reconnection fallback) |

`request_approval` accepts a payload with: document type, markdown content, epic, optional story. The dashboard holds the request until the human acts. This is the only blocking operation in the system — the agent waits until a verdict is received.

`approval_status` is a non-blocking polling fallback for crash recovery. If an agent disconnects and reconnects, it can check whether a pending approval has already been decided.

**Removed tools:** `notify_agent_state` and `register_project` are no longer part of the MCP server. Agent state telemetry is now handled automatically by HTTP hooks (`SubagentStart`/`SubagentStop`) — agents do not need to make explicit MCP calls for telemetry. Project registration is handled by `oraculo install`.

### 2.4 Real-time Communication

WebSocket provides push updates from server to browser. Two channels feed the WebSocket:

1. **HTTP Hooks** — Claude Code fires hooks on system events. The HTTP server persists metadata to SQLite and broadcasts to all connected WebSocket clients. Hooks are fire-and-forget: `200` empty body, non-blocking. Event types:

   | WebSocket event | Hook endpoint | Claude Code hook | Trigger |
   |---|---|---|---|
   | `session_started` | `POST /hooks/session-start` | `SessionStart` (command) | Session begins |
   | `agent_started` | `POST /hooks/agent-start` | `SubagentStart` | Agent spawned |
   | `agent_stopped` | `POST /hooks/agent-stop` | `SubagentStop` | Agent completed/failed |
   | `tool_used` | `POST /hooks/tool-used` | `PostToolUse` | Mutation tool used (Bash\|Edit\|Write\|NotebookEdit) |
   | `task_completed` | `POST /hooks/task-completed` | `TaskCompleted` | Task marked complete |
   | `agent_stopping` | `POST /hooks/stop` | `Stop` | Agent stopping |
   | `teammate_idle` | `POST /hooks/teammate-idle` | `TeammateIdle` | Teammate idle |
   | `session_ended` | `POST /hooks/session-end` | `SessionEnd` | Session ended |

2. **MCP + CLI** — Approval requests and version reviews are broadcast when they occur.

   | WebSocket event | Origin |
   |---|---|
   | `approval_requested` | MCP `request_approval` tool (operational gates) |
   | `version_created` | CLI `epic version` / `story version` (document versioning) |
   | `review_submitted` | CLI `review create` (document reviews) |

WebSocket message format:

```json
{
  "type": "agent_started" | "agent_stopped" | "tool_used" | "task_completed" | "session_started" | "session_ended" | "approval_requested" | "version_created" | "review_submitted" | "teammate_idle" | "agent_stopping",
  "payload": { ... }
}
```

The frontend subscribes to event types per screen. The DAG View subscribes to `task_completed`. The Agent Monitor subscribes to `agent_started` and `agent_stopped`. The Approvals screen subscribes to `approval_requested`, `version_created`, and `review_submitted`. The Activity Feed subscribes to `tool_used`. The Sessions screen subscribes to `session_started` and `session_ended`.

### 2.5 Data Flow

Most data flows through CLI internals. The HTTP server calls the same Go functions that the CLI commands use — validated, contracted operations with strict preconditions. For telemetry data (sessions, agents, tool events), the server queries SQLite tables directly — these records are inserted by HTTP hooks, not by CLI commands. This preserves the CLI as the single Trust Layer for all business logic while allowing the HTTP server to read hook-persisted telemetry directly.

```
[Browser] --HTTP--> [Go HTTP server] --in-process--> [CLI internals] --SQLite--> [.oraculo/oraculo.db]
[Browser] --HTTP--> [Go HTTP server] --SQLite query-> [sessions, agents, tool_events tables]
[Browser] <--WS---- [Go HTTP server] <--HTTP hooks--- [Claude Code automatic hooks]
[Agent]   --stdio-> [Go MCP server]  --WS broadcast-> [Browser]
```

## 3. Screens

### 3.1 Landing (Epic Selection)

**Purpose:** Entry point for the dashboard when no Epic is selected. The user selects which Epic to work with. All other screens are scoped to the selected Epic.

**Data sources:**
- `oraculo tools epic list` — All epics with phase, story count, task status

**Layout:** A page title "Select an Epic" with subtitle "Choose an Epic to view its stories, tasks, and agent activity." Below, a responsive grid of Epic cards. Each card shows: Epic name (header, semibold), phase badge (Discover / Plan / Execute / Validate), progress bar showing task completion, stats row (story count, task count, completion percentage), status badge (completed / in_progress / pending), and an "Open" button.

**Key interactions:** Click a card or its "Open" button to select the Epic. The sidebar's Epic dropdown updates and the view navigates to Home / Overview for that Epic. The sidebar nav items become active once an Epic is selected.

**When no Epic is selected:** Sidebar Epic dropdown shows "Select Epic...", nav items (Home / Overview through Sessions) are visually muted/disabled, and Settings remains always accessible.

### 3.2 Home / Overview

**Purpose:** Epic-level dashboard showing all stories at a glance. This is the default view when an Epic is selected.

**Data sources:**
- `oraculo tools story list --epic <epic>` — Stories for the selected Epic
- `oraculo tools task list --epic <epic> --story <story>` — Task status for each story

**Layout:** Grid of story cards. Each card shows: story name, status badge (pending, in_progress, completed, failed), task completion progress bar, task count, and agent assignment (if any). Summary metrics at top: total stories, completion percentage, tasks by status.

**Key interactions:** Click a story card to open its Story Detail Page. Filter by status. Sort by name, status, or creation date.

### 3.3 Stories

**Purpose:** Navigate stories within the selected Epic. Each story opens a dedicated detail page with its tasks and documents.

**Data sources:**
- `oraculo tools story list --epic <epic>` — Stories for the selected Epic
- `oraculo tools story get <name> --epic <epic>` — Story requirements markdown
- `oraculo tools task list --epic <epic> --story <story>` — Tasks with status

**Layout:** Grid or vertical list of story cards. Each card shows: story name, status badge (pending, in_progress, completed, failed), task completion progress bar, and task count.

**Key interactions:** Click a story card to open its dedicated detail page. Filter by status. Sort by name, status, or creation date.

### 3.4 Story Detail Page

**Purpose:** Dedicated workspace for a single story, providing access to all its documents and tasks through a tabbed interface.

**Data sources:**
- `oraculo tools story get <name> --epic <epic>` — Story requirements markdown
- `oraculo tools task list --epic <epic> --story <story>` — Tasks with status and dependencies
- `oraculo tools task get <name> --epic <epic> --story <story>` — Individual task detail with result

**Layout:** Breadcrumb navigation at top: `Epic Name → Story Name → [current tab]`. Below the breadcrumb, horizontal tabs:

| Tab | Content |
|---|---|
| **Tasks** | List of tasks with status badges, progress indicators, and agent assignments |
| **Design** | Design document markdown (if present) |
| **Requirements** | Story requirements/definition markdown |
| **QA** | QA verdicts, test results, and validation history |

**Key interactions:** Click tabs to switch views. Click a task to see its full detail and logs. Link from task to DAG View position. Show version history toggle for documents (Requirements, Design).

**Navigation:** Accessed by clicking a story card from the Stories list. Breadcrumb allows navigation back to Epic overview (click Epic name) or Stories list (click Story name).

### 3.5 DAG View

**Purpose:** Visualize task dependency graphs for a story's execution plan.

**Data sources:**
- `oraculo tools task list --epic <epic> --story <story>` — Returns tasks with `depends_on` references

**Update mechanism:** WebSocket push from `task_completed` event (fired by `TaskCompleted` hook). More reliable than WAL file watching — the event fires exactly when a task is marked complete, not on every database write.

**Layout:** Full-width directed graph. Nodes represent tasks, edges represent dependencies. Layout uses a left-to-right topological sort. Node colors encode status: gray (pending), blue (in_progress), green (completed), red (failed). A critical path highlight traces the longest dependency chain. Agent assignment labels appear below each node.

**Key interactions:** Hover a node for task detail tooltip. Click a node to open its detail in a side drawer. Toggle critical path highlighting. Filter by status. Zoom and pan.

**Note:** This is unique to Oraculo. The graph is computed client-side from the task list JSON — no additional CLI command is needed. The `depends_on` field in each task provides the edge list.

### 3.6 Agent Monitor

**Purpose:** Real-time visibility into what agents are doing right now.

**Data sources:**
- `agents` table (SQLite) — Current state and history of agents, populated by HTTP hooks
- WebSocket `agent_started` and `agent_stopped` events — Push from `SubagentStart`/`SubagentStop` hooks

**Update mechanism:** Push only via WebSocket (no polling). Agents do not need to make explicit MCP calls — Claude Code fires the hooks automatically on every agent lifecycle event.

**Layout:** Grid of agent cards, one per active agent. Each card shows: agent type icon (orchestrator/code/qa/research), current task name, elapsed time, status indicator. Below the grid, a scrollable activity feed showing timestamped events (agent started, task completed, QA verdict issued, escalation triggered).

**Key interactions:** Click an agent card to see its full history for the current session. Click a task reference in the feed to navigate to the DAG View. Filter feed by agent type or event type.

### 3.7 Activity Feed

**Purpose:** Real-time visibility into tool mutations performed by agents — which files were edited and which commands were executed.

**Data sources:**
- `tool_events` table (SQLite) — History of mutation tool events, populated by the `PostToolUse` hook
- WebSocket `tool_used` events — Push from the `PostToolUse` hook (Bash|Edit|Write|NotebookEdit only)

**Update mechanism:** Push only via WebSocket. Only mutation tools are tracked; read-only tools (Read, Glob, Grep, WebFetch) are excluded to reduce noise. Only metadata is stored — no content, no diffs, no command text.

**Layout:** Reverse-chronological feed of tool events. Each event shows: timestamp, agent name, tool name (with icon), file path (for Edit/Write/NotebookEdit) or Bash indicator. Filters by tool type and by agent.

**Key interactions:** Click a file path to open the associated task detail. Filter by tool type (Bash, Edit, Write, NotebookEdit) or by agent. Export activity log.

### 3.8 Approvals

**Purpose:** Human review gate for document versions (epic requirements, story definitions) and operational approvals (QA escalations, design, execution-plan).

**Data sources:**
- For document reviews: `oraculo tools epic versions <epic-name>` / `oraculo tools story versions <story-name> --epic <epic-name>` — Load version history
- For operational gates: MCP `request_approval` calls populate the queue
- Diff between versions uses `epic_versions` / `story_versions` tables

**Layout:** Left column: review/approval queue sorted by arrival time, with type badges (epic-version/story-version/qa-escalation/design/execution-plan) and age indicators. Right column: rich markdown viewer with syntax highlighting. For document versions, a toggle switches between rendered view and side-by-side diff against the previous version. Below the viewer: inline comment field and action buttons. For document reviews: two buttons (Approve, Reject). For operational gates: three buttons (Approve, Reject, Needs Revision).

**Key interactions:** Select a queue item to load its content. Toggle diff mode. Add comments (attached to the verdict). Submit verdict — the server relays it back to the requesting agent via MCP callback.

### 3.9 QA Dashboard

**Purpose:** Track validation outcomes across the project.

**Data sources:**
- `oraculo tools task list --epic <epic> --story <story>` — Task status (completed/failed implies QA cycle)
- `oraculo tools task get <name> --epic <epic> --story <story>` — Task result with summary and logs
- WebSocket `agent_stopped` events with QA verdict payloads

**Update mechanism:** Hybrid — initial load via CLI, updates via WebSocket push from `SubagentStop` hooks.

**Layout:** Top row: summary cards (total validations, approval rate, average rejection cycles, active escalations). Below: a table of recent verdicts showing task name, verdict (approved/rejected), attempt number, and timestamp. Rejection count badges highlight tasks approaching the circuit breaker threshold (default 3). An escalation indicator marks tasks that have breached the threshold and require human intervention.

**Key interactions:** Click a verdict row to see the full QA findings. Sort/filter by verdict, story, or date. Link to the related task in DAG View.

### 3.10 Knowledge Base

**Purpose:** Browse and search accumulated project wisdom from the knowledge table.

**Data sources:**
- `oraculo tools memory search <query>` — Full-text search across findings
- `oraculo tools memory domains` — List available domains for filtering

**Update mechanism:** On-demand (user triggers search). No real-time push needed.

**Layout:** Search bar at top with domain and category dropdown filters. Results appear as cards showing: domain badge, category tag, finding text (truncated), confidence indicator, source files, and timestamp. Click a card to expand the full finding.

**Key interactions:** Type-ahead search triggers `memory search` with debounce. Filter by domain (from `memory domains` response) and category (`pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`). Sort by confidence or recency.

### 3.11 Sessions

**Purpose:** History and observability of Claude Code sessions. Shows which sessions were tracked, which models were used, and the associated agent and tool activity.

**Data sources:**
- `sessions` table (SQLite) — Sessions registered via `session-start`/`session-end` hooks
- `agents` table (SQLite) — Agents associated with each session
- `tool_events` table (SQLite) — Tool events per session

**Update mechanism:** REST API + WebSocket push on `session_started` and `session_ended` events.

**SQLite schema:**

```sql
-- sessions: tracks Claude Code sessions observed via hooks
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,   -- Session identifier from Claude Code
    model       TEXT,               -- Model in use (e.g. claude-opus-4-6)
    cwd         TEXT,               -- Working directory
    started_at  TEXT NOT NULL,      -- ISO 8601 timestamp
    ended_at    TEXT,               -- ISO 8601 timestamp (null while active)
    end_reason  TEXT                -- Why the session ended
);

-- agents: tracks agent lifecycle events from SubagentStart/Stop hooks
CREATE TABLE agents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    name        TEXT NOT NULL,      -- Agent name/identifier
    type        TEXT NOT NULL,      -- code, qa, research, orchestrator
    status      TEXT NOT NULL DEFAULT 'active',  -- active, completed, failed
    started_at  TEXT NOT NULL,      -- ISO 8601 timestamp
    stopped_at  TEXT               -- ISO 8601 timestamp (null while active)
);

-- tool_events: tracks mutation tool usage (metadata only — no content, no diffs)
CREATE TABLE tool_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    tool_name   TEXT NOT NULL,      -- Bash, Edit, Write, NotebookEdit
    file_path   TEXT,              -- File path (null for Bash)
    timestamp   TEXT NOT NULL      -- ISO 8601 timestamp
);
```

**Layout:** List of sessions with status indicator (active/ended), model used, working directory, and duration. Clicking a session expands the associated agents and tool events.

**Key interactions:** Filter by status (active/ended), model, or date. View agent and tool details per session.

### 3.12 Settings

**Purpose:** Project configuration and dashboard preferences.

**Layout:** Tabbed form. Tabs: Project (name, directory path, CLI binary path), Dashboard (refresh interval, theme light/dark, notification preferences).

**Key interactions:** Save persists to `.oraculo/config`. The Settings screen is always accessible regardless of Epic selection.

## 4. Navigation and Layout

**Shell:** Persistent left sidebar with icon + label links for each screen. The sidebar collapses to icon-only on narrow viewports. The sidebar header contains the brand name ("Oraculo") and an Epic dropdown selector (Context Switcher).

**Epic dropdown (Context Switcher):** Located at the top of the sidebar, above all navigation items. Shows the currently selected Epic name with a chevron indicator. Clicking opens a dropdown to switch between Epics. When no Epic is selected, displays "Select Epic..." and nav items below are visually muted.

**Sidebar order:**
1. 🏠 Home / Overview — Epic overview (all stories in grid)
2. 📋 Stories — Vertical list or grid of stories
3. 🕸️ DAG View — Task dependency graph
4. 🤖 Agent Monitor — Real-time agent activity
5. 📄 Approvals — Review queue with unread count badge
6. 🧪 QA Dashboard — Validation outcomes
7. 🧠 Knowledge Base — Accumulated project wisdom
8. 🔍 Sessions — Session history and observability
— separator —
9. ⚙️ Settings — Project and dashboard preferences

**Navigation model:**
- **Home / Overview** is the default landing page when an Epic is selected. Shows all stories in a grid with status and progress indicators.
- **Stories** displays a list or grid of all stories in the current Epic. Each story is a dedicated page.
- **Story Detail Page** is accessed by clicking a story. Uses horizontal tabs: [Tasks] [Design] [Requirements] [QA]. Breadcrumb navigation: `Epic → Story → [tab]`.
- **Aprovals** is centralized — all approval types (document versions, QA escalations, operational gates) appear in a single queue.

**When an Epic is selected:** All nav items (1-8) are active. Default view is Home / Overview. Switching Epics via the dropdown navigates to Home / Overview of the new Epic. Agent Monitor, Activity Feed, and Sessions show data from the current session regardless of the selected Epic.

**When no Epic is selected (Landing):** Nav items 1-7 are grayed/disabled. Only Settings is interactive. The main content area shows the Epic selection grid.

**Responsive behavior:** The sidebar collapses below 1024px viewport width. Master-detail views (Stories, Approvals) stack vertically on narrow screens. The DAG View remains full-width with horizontal scroll.

**Theme:** Light and dark modes. The default visual direction is `Mission Control`: `Slate` base, `Blue` accent, cool surfaces, and high contrast tuned for sustained reading. Two alternative directions are documented in the design system for specific workflows: `Operations Warm` (`Stone` base, `Orange` accent) for more editorial review flows, and `Signal Green` (`Zinc` base, `Green` accent) for monitoring and QA. Status-semantic colors remain distinct from the primary accent: blue for in-progress, green for completed, red for failed, amber for attention/pending approval.

### 4.1 Design System Foundations

The dashboard adopts an explicit design system before screen-by-screen implementation. It defines foundations, tokens, and reusable base components so new screens do not accumulate ad-hoc visual decisions.

**Typography:**
- `Space Grotesk` — display family for page titles, section titles, and high-emphasis metrics
- `IBM Plex Sans` — base family for navigation, body copy, labels, buttons, tables, and continuous reading
- `IBM Plex Mono` — operational family for IDs, timestamps, paths, technical counters, and telemetry

**Weights and use:**
- `Space Grotesk 700` — screen titles and highlighted metric values
- `Space Grotesk 600` — section titles and primary card titles
- `IBM Plex Sans 400` — body text, descriptions, and markdown preview
- `IBM Plex Sans 500` — labels, buttons, tabs, and table headers
- `IBM Plex Sans 600` — active nav, selected states, and short emphasis
- `IBM Plex Mono 500` — operational values and technical strings

**Initial type tokens:**
- `display-lg` — `40/48`, `Space Grotesk 700`, used for screen titles
- `heading-md` — `24/32`, `Space Grotesk 600`, used for section headings
- `body-md` — `16/24`, `IBM Plex Sans 400`, used for primary UI body copy
- `label-sm` — `14/20`, `IBM Plex Sans 500`, used for labels and short metadata
- `mono-sm` — `13/18`, `IBM Plex Mono 500`, used for IDs, timestamps, and paths

**Scales and surfaces:**
- Base spacing scale of `8 / 12 / 16 / 24 / 32 / 40`
- Base radius scale in small, medium, and large steps to control density without defaulting to pill shapes
- `background` as the deepest canvas tone, never used to imply interaction
- `card` as the default reading surface, with a thin border and generous internal spacing
- `secondary` as the supporting surface for metadata, inset states, and restrained grouping

**Semantics and iconography:**
- The primary accent must never replace semantic status colors
- Status remains explicit through `pending`, `in_progress`, `completed`, and `failed`
- Iconography stays outline-based, paired with `IBM Plex Sans 500` labels and disciplined semantic color usage

**Initial component inventory:**
- `sidebar item`
- `topbar slot`
- `metric card`
- `status badge`
- `table shell`
- `markdown panel`

Planned future components:
- `DAG node`
- `approval queue item`
- `diff viewer`
- `agent activity card`
- `knowledge result card`

## 5. Technology Stack

| Layer | Technology | Justification |
|---|---|---|
| **Frontend (build-time)** | | |
| Framework | Next.js | Static export generates optimized HTML/CSS/JS bundle |
| Component library | shadcn/ui | High-quality, composable components; no runtime dependency |
| Styling | Tailwind CSS | Utility-first for rapid UI iteration, dark mode built-in |
| Typography | Space Grotesk + IBM Plex Sans + IBM Plex Mono | Separates display hierarchy, interface legibility, and operational telemetry |
| Graph rendering | D3.js or dagre | DAG layout computation and SVG rendering for the dependency graph |
| Markdown rendering | react-markdown + remark-gfm | Render requirements documents with GitHub-flavored markdown |
| **Backend (runtime)** | | |
| HTTP server | Go `net/http` | Standard library, no external dependency, production-grade |
| WebSocket | `gorilla/websocket` or `nhooyr.io/websocket` | Mature Go WebSocket implementations |
| Hook endpoints | Go `net/http` (same server) | Receive fire-and-forget POST requests from Claude Code hooks |
| Static embedding | Go `embed.FS` | Frontend assets compiled into the binary at build time |
| MCP server | Go MCP SDK or custom stdio handler | MCP protocol over stdin/stdout for Claude Code integration |
| CLI integration | In-process function calls | HTTP handlers call the same Go functions as CLI commands — no subprocess overhead |

**Removed dependency:** `fsnotify/fsnotify` — File watcher is no longer needed. Real-time data arrives via HTTP hooks instead of filesystem monitoring.

## 6. Data Sources

Summary of how each screen obtains its data under the two-channel architecture:

| Screen | Primary source | Update mechanism |
|---|---|---|
| Landing | CLI: `epic list` | Polling on mount (unchanged) |
| Home / Overview | CLI: `story list`, `task list` | Polling + WebSocket `task_completed` |
| Stories | CLI: `story list` | Polling |
| Story Detail Page | CLI: `story get`, `task list` | Polling + WebSocket `task_completed` |
| DAG View | CLI: `task list` (includes `depends_on`) | WebSocket push from `TaskCompleted` hook |
| Agent Monitor | `agents` table (SQLite) + WebSocket | WebSocket push from `SubagentStart`/`SubagentStop` hooks |
| Activity Feed | `tool_events` table (SQLite) + WebSocket | WebSocket push from `PostToolUse` hook |
| Approvals | `approvals` table + `epic_versions`/`story_versions` tables (SQLite) + WebSocket | WebSocket push from MCP `request_approval` (operational gates) and `version_created`/`review_submitted` (document reviews) |
| QA Dashboard | CLI: `task list`, `task get` + WebSocket | Hybrid: initial load via CLI, updates via WebSocket |
| Knowledge Base | CLI: `memory search`, `memory domains` | On-demand (user triggers search) |
| Sessions | `sessions` table (SQLite) | REST API + WebSocket push on session start/end |
| Settings | Local config file | Read on mount, write on save |

**What changed from the previous design:**

| Aspect | Before | After |
|---|---|---|
| Agent Monitor | MCP `notify_agent_state` events (required agent instrumentation) | Automatic HTTP hooks `SubagentStart`/`SubagentStop` |
| DAG View | File watcher on `.oraculo/oraculo.db-wal` | `TaskCompleted` hook via WebSocket push |
| Activity Feed | Did not exist | New — `PostToolUse` hook (mutation metadata only) |
| Sessions | Did not exist | New — `sessions` table via `session-start`/`session-end` hooks |
| Approvals | MCP `request_approval` | Unchanged — MCP remains for blocking gates |

## 7. Hook Endpoint Reference

This section documents the HTTP hook endpoints that feed real-time data into the dashboard. Each endpoint is invoked by Claude Code automatically — no agent instrumentation required. The server persists metadata to SQLite and broadcasts a WebSocket event to all connected dashboard clients.

All endpoints accept POST requests with `timeout: 5` (5 seconds) and return `200` with an empty body. If the server is unreachable, Claude Code logs a non-blocking warning and the agent continues unaffected.

| Endpoint | Claude Code hook | WebSocket event | Persists to |
|---|---|---|---|
| `POST /hooks/session-start` | `SessionStart` (command hook) | `session_started` | `sessions` table |
| `POST /hooks/agent-start` | `SubagentStart` | `agent_started` | `agents` table |
| `POST /hooks/agent-stop` | `SubagentStop` | `agent_stopped` | `agents` table |
| `POST /hooks/tool-used` | `PostToolUse` (Bash\|Edit\|Write\|NotebookEdit) | `tool_used` | `tool_events` table |
| `POST /hooks/task-completed` | `TaskCompleted` | `task_completed` | (event metadata) |
| `POST /hooks/stop` | `Stop` | `agent_stopping` | (event metadata) |
| `POST /hooks/teammate-idle` | `TeammateIdle` | `teammate_idle` | (event metadata) |
| `POST /hooks/session-end` | `SessionEnd` | `session_ended` | `sessions` table |

**Note on `session-start`:** The `SessionStart` hook is the only command hook (not a direct HTTP hook). It runs `oraculo hook session-start`, which performs a health check, prints a warning to stderr if the server is offline, and sends session metadata via HTTP if the server is online. It always exits with code 0 — the session is never blocked.

**Note on `tool-used`:** Only mutation tools are tracked (Bash, Edit, Write, NotebookEdit). Read-only tools (Read, Glob, Grep, WebFetch) are excluded to reduce noise. Only metadata is stored — no content, no diffs, no command text.

**Error handling:** All HTTP hook endpoints follow graceful degradation. If the server is offline, Claude Code logs a non-blocking warning and the agent continues. Hooks never block agents — telemetry is best-effort.

## 8. Future Work

Deferred capabilities for later iterations:

- **Inline editing** — Edit requirements markdown directly in the Epic Explorer viewer and save back through `oraculo tools epic save` / `oraculo tools story save`.
- **Timeline view** — Gantt-style visualization of task execution over time, using `started_at` and `completed_at` timestamps.
- **Agent log streaming** — Stream agent stdout/stderr in real-time to the Agent Monitor, beyond structured state events.
- **Notification system** — Desktop notifications for approval requests, QA escalations, and epic completion.
- **Role-based access** — Restrict approval authority and settings access by user role when Oraculo is used in team settings.
- **Embedded terminal** — Launch `oraculo` CLI commands directly from the dashboard interface.
- **Dashboard API for CI** — Expose read-only REST endpoints for CI/CD pipelines to query project status.
- **PreToolUse policy enforcement** — Use the `PreToolUse` hook to block dangerous operations based on rules defined in the dashboard (deferred until the two-channel architecture is validated in production).
