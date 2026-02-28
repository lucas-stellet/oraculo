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
```

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
| `POST /api/approvals/:id/verdict` | (internal state) | Submit approval decision |

The server manages one stateful concern: the **approval queue**. Approval requests arrive via MCP tool calls from agents and are held in memory until the human responds through the dashboard. The verdict is then relayed back to the requesting agent.

### 2.2 Startup Flow

When Claude Code starts a session, it launches `oraculo mcp` as a registered MCP server (configured during `oraculo install`). The startup sequence:

1. `oraculo mcp` starts the MCP server on stdio (standard MCP transport)
2. The MCP server starts the HTTP server on an available port
3. The HTTP server begins serving the embedded frontend and REST API
4. The binary opens the user's default browser to `http://localhost:<port>`
5. File watchers start monitoring `.oraculo/` for changes

If the HTTP server is already running (from another session), the new MCP server connects to the existing instance instead of starting a duplicate. The human never needs to run a separate command.

### 2.3 MCP Server

The MCP server exposes tools that AI agents call during orchestration. It runs in the same process as the HTTP server and shares the WebSocket broadcast channel.

| MCP Tool | Direction | Purpose |
|---|---|---|
| `request_approval` | Agent -> Dashboard | Submit a document for human review (requirements, story, QA escalation) |
| `notify_agent_state` | Agent -> Dashboard | Report agent start, completion, failure, or QA verdict |
| `register_project` | Agent -> Dashboard | Register a project directory for dashboard discovery |

`request_approval` accepts a payload with: document type, markdown content, optional previous version (for diff), and a callback mechanism. The dashboard holds the request until the human acts.

`notify_agent_state` pushes real-time updates that the Agent Monitor screen consumes. Payload includes: agent ID, agent type (orchestrator/code/qa/research), task reference, event type (started/completed/failed), and optional metadata.

### 2.4 Real-time Communication

WebSocket provides push updates from server to browser. Two event sources feed the WebSocket channel:

1. **File watcher** — Monitors `.oraculo/epics/**/*.md` for markdown changes and `.oraculo/oraculo.db-wal` for SQLite write-ahead log activity. On change, the server re-queries the relevant CLI commands and pushes deltas.

2. **MCP notifications** — Agent state changes and approval requests are broadcast immediately upon receipt.

WebSocket message format:

```json
{
  "type": "task_updated" | "approval_requested" | "agent_state" | "knowledge_added" | "file_changed",
  "payload": { ... }
}
```

The frontend subscribes to event types per screen. The DAG View subscribes to `task_updated`. The Agent Monitor subscribes to `agent_state`. The Approvals screen subscribes to `approval_requested`.

### 2.5 Data Flow

All data flows through CLI internals. The HTTP server calls the same Go functions that the CLI commands use — validated, contracted operations with strict preconditions. The server never opens SQLite directly and never writes to the filesystem outside CLI functions. This preserves the CLI as the single Trust Layer — all validation, lifecycle enforcement, and schema migration remain in one place.

```
[Browser] --HTTP--> [Go HTTP server] --in-process--> [CLI internals] --SQLite--> [.oraculo/oraculo.db]
[Browser] <--WS---- [Go HTTP server] <--file watch-- [.oraculo/epics/**/*.md]
[Agent]   --stdio-> [Go MCP server]  --WS broadcast-> [Browser]
```

## 3. Screens

### 3.1 Home/Dashboard

**Purpose:** Project overview at a glance — what exists, what is in progress, what needs attention.

**Data sources:**
- `oraculo tools epic list` — Epic count and names
- `oraculo tools story list --epic <epic>` — Story count per epic
- `oraculo tools task list --epic <epic> --story <story>` — Task status distribution

**Layout:** Four summary cards across the top (total epics, total stories, task completion percentage, active agents). Below, a table of epics with inline progress bars showing story/task completion. A sidebar panel shows the current phase indicator (Discover/Plan/Execute/Validate) for the most recently active epic.

**Key interactions:** Click any epic row to navigate to Epic Explorer. Click the active agents count to jump to Agent Monitor.

### 3.2 Epic Explorer

**Purpose:** Navigate the Epic > Story > Task hierarchy and read requirements documents.

**Data sources:**
- `oraculo tools epic list` + `oraculo tools epic get <name>` — Epic metadata and markdown
- `oraculo tools story list --epic <epic>` + `oraculo tools story get <name> --epic <epic>` — Story metadata and markdown
- `oraculo tools task list --epic <epic> --story <story>` — Task list with status

**Layout:** Master-detail split. Left panel: collapsible tree showing Epics > Stories > Tasks. Right panel: markdown viewer rendering the selected entity's requirements document or task detail. Phase badges (pending, in_progress, completed, failed) appear next to each tree node.

**Key interactions:** Select a tree node to load its content in the viewer. Expand/collapse levels. Filter tree by status. Link from any task to its DAG View position.

### 3.3 DAG View

**Purpose:** Visualize task dependency graphs for a story's execution plan.

**Data sources:**
- `oraculo tools task list --epic <epic> --story <story>` — Returns tasks with `depends_on` references

**Layout:** Full-width directed graph. Nodes represent tasks, edges represent dependencies. Layout uses a left-to-right topological sort. Node colors encode status: gray (pending), blue (in_progress), green (completed), red (failed). A critical path highlight traces the longest dependency chain. Agent assignment labels appear below each node.

**Key interactions:** Hover a node for task detail tooltip. Click a node to open its detail in a side drawer. Toggle critical path highlighting. Filter by status. Zoom and pan.

**Note:** This is unique to Oraculo. The graph is computed client-side from the task list JSON — no additional CLI command is needed. The `depends_on` field in each task provides the edge list.

### 3.4 Agent Monitor

**Purpose:** Real-time visibility into what agents are doing right now.

**Data sources:**
- WebSocket `agent_state` events from MCP `notify_agent_state` calls

**Layout:** Grid of agent cards, one per active agent. Each card shows: agent type icon (orchestrator/code/qa/research), current task name, elapsed time, status indicator. Below the grid, a scrollable activity feed showing timestamped events (agent started, task completed, QA verdict issued, escalation triggered).

**Key interactions:** Click an agent card to see its full history for the current session. Click a task reference in the feed to navigate to the DAG View. Filter feed by agent type or event type.

### 3.5 Approvals

**Purpose:** Human review and approval gate for requirements documents, story definitions, and QA escalations.

**Data sources:**
- MCP `request_approval` calls populate the queue
- `oraculo tools epic get` / `oraculo tools story get` — Load current document versions for diff

**Layout:** Left column: approval queue sorted by arrival time, with type badges (requirements/story/qa-escalation) and age indicators. Right column: rich markdown viewer with syntax highlighting. When a previous version exists, a toggle switches between rendered view and side-by-side diff. Below the viewer: inline comment field, and three action buttons (Approve, Reject, Needs Revision).

**Key interactions:** Select a queue item to load its content. Toggle diff mode. Add comments (attached to the verdict). Submit verdict — the server relays it back to the requesting agent via MCP callback.

### 3.6 QA Dashboard

**Purpose:** Track validation outcomes across the project.

**Data sources:**
- `oraculo tools task list --epic <epic> --story <story>` — Task status (completed/failed implies QA cycle)
- `oraculo tools task get <name> --epic <epic> --story <story>` — Task result with summary and logs
- WebSocket `agent_state` events with QA verdict payloads

**Layout:** Top row: summary cards (total validations, approval rate, average rejection cycles, active escalations). Below: a table of recent verdicts showing task name, verdict (approved/rejected), attempt number, and timestamp. Rejection count badges highlight tasks approaching the circuit breaker threshold (default 3). An escalation indicator marks tasks that have breached the threshold and require human intervention.

**Key interactions:** Click a verdict row to see the full QA findings. Sort/filter by verdict, story, or date. Link to the related task in DAG View.

### 3.7 Knowledge Base

**Purpose:** Browse and search accumulated project wisdom from the knowledge table.

**Data sources:**
- `oraculo tools memory search <query>` — Full-text search across findings
- `oraculo tools memory domains` — List available domains for filtering

**Layout:** Search bar at top with domain and category dropdown filters. Results appear as cards showing: domain badge, category tag, finding text (truncated), confidence indicator, source files, and timestamp. Click a card to expand the full finding.

**Key interactions:** Type-ahead search triggers `memory search` with debounce. Filter by domain (from `memory domains` response) and category (`pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`). Sort by confidence or recency.

### 3.8 Settings

**Purpose:** Project configuration and dashboard preferences.

**Layout:** Tabbed form. Tabs: Project (name, directory path, CLI binary path), Dashboard (refresh interval, theme light/dark, notification preferences), Connected Projects (multi-project list with add/remove).

**Key interactions:** Save persists to a local config file (`~/.oraculo/dashboard.json`). Connected projects allow the dashboard to serve multiple Oraculo-enabled repositories from a single instance.

## 4. Navigation and Layout

**Shell:** Persistent left sidebar with icon + label links for each screen. The sidebar collapses to icon-only on narrow viewports. Top bar shows the current project name and a project switcher dropdown (for multi-project mode).

**Sidebar order:**
1. Home/Dashboard
2. Epic Explorer
3. DAG View
4. Agent Monitor
5. Approvals (with unread count badge)
6. QA Dashboard
7. Knowledge Base
8. Settings

**Responsive behavior:** The sidebar collapses below 1024px viewport width. Master-detail views (Epic Explorer, Approvals) stack vertically on narrow screens. The DAG View remains full-width with horizontal scroll.

**Theme:** Light and dark modes. Colors follow a neutral base with status-semantic accents: blue for in-progress, green for completed, red for failed, amber for pending approval.

## 5. Technology Stack

| Layer | Technology | Justification |
|---|---|---|
| **Frontend (build-time)** | | |
| Framework | Next.js | Static export generates optimized HTML/CSS/JS bundle |
| Component library | shadcn/ui | High-quality, composable components; no runtime dependency |
| Styling | Tailwind CSS | Utility-first for rapid UI iteration, dark mode built-in |
| Graph rendering | D3.js or dagre | DAG layout computation and SVG rendering for the dependency graph |
| Markdown rendering | react-markdown + remark-gfm | Render requirements documents with GitHub-flavored markdown |
| **Backend (runtime)** | | |
| HTTP server | Go `net/http` | Standard library, no external dependency, production-grade |
| WebSocket | `gorilla/websocket` or `nhooyr.io/websocket` | Mature Go WebSocket implementations |
| File watching | `fsnotify/fsnotify` | Cross-platform filesystem watching for `.oraculo/` changes |
| Static embedding | Go `embed.FS` | Frontend assets compiled into the binary at build time |
| MCP server | Go MCP SDK or custom stdio handler | MCP protocol over stdin/stdout for Claude Code integration |
| CLI integration | In-process function calls | HTTP handlers call the same Go functions as CLI commands — no subprocess overhead |

## 6. Data Sources

Summary of how each screen obtains its data:

| Screen | Primary source | Update mechanism |
|---|---|---|
| Home/Dashboard | CLI: `epic list`, `story list`, `task list` | Polling on mount + WebSocket `task_updated` |
| Epic Explorer | CLI: `epic get`, `story get`, `task list` | File watcher on `.oraculo/epics/**/*.md` |
| DAG View | CLI: `task list` (includes `depends_on`) | WebSocket `task_updated` |
| Agent Monitor | WebSocket: `agent_state` events | Push only (no polling) |
| Approvals | MCP: `request_approval` | Push only (no polling) |
| QA Dashboard | CLI: `task list`, `task get` + WebSocket `agent_state` | Hybrid: initial load via CLI, updates via WebSocket |
| Knowledge Base | CLI: `memory search`, `memory domains` | On-demand (user triggers search) |
| Settings | Local config file | Read on mount, write on save |

## 7. Future Work

Deferred capabilities for later iterations:

- **Inline editing** — Edit requirements markdown directly in the Epic Explorer viewer and save back through `oraculo tools epic save` / `oraculo tools story save`.
- **Timeline view** — Gantt-style visualization of task execution over time, using `started_at` and `completed_at` timestamps.
- **Agent log streaming** — Stream agent stdout/stderr in real-time to the Agent Monitor, beyond structured state events.
- **Notification system** — Desktop notifications for approval requests, QA escalations, and epic completion.
- **Role-based access** — Restrict approval authority and settings access by user role when Oraculo is used in team settings.
- **Embedded terminal** — Launch `oraculo` CLI commands directly from the dashboard interface.
- **Dashboard API for CI** — Expose read-only REST endpoints for CI/CD pipelines to query project status.
