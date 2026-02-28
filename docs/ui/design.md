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
2. The MCP server reads the project's allocated port from `.oraculo/config` → `dashboard.port`
3. If no port is configured, it allocates the first available port in the range 3100-3199
4. The HTTP server attempts to bind the configured port
5. If the port is occupied, it performs a sequential scan within 3100-3199 for the first free port
6. The new port is saved back to `.oraculo/config`
7. If no port is available in the range, the server exits with a clear error: "All ports in range 3100-3199 are in use"
8. The HTTP server begins serving the embedded frontend and REST API
9. The binary opens the user's default browser to `http://localhost:<port>`
10. File watchers start monitoring `.oraculo/` for changes

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

### 3.1 Landing (Epic Selection)

**Purpose:** Entry point for the dashboard. The user selects which Epic to work with. All other screens are scoped to the selected Epic.

**Data sources:**
- `oraculo tools epic list` — All epics with phase, story count, task status

**Layout:** A page title "Select an Epic" with subtitle "Choose an Epic to view its stories, tasks, and agent activity." Below, a responsive grid of Epic cards. Each card shows: Epic name (header, semibold), phase badge (Discover / Plan / Execute / Validate), progress bar showing task completion, stats row (story count, task count, completion percentage), status badge (completed / in_progress / pending), and an "Open" button.

**Key interactions:** Click a card or its "Open" button to select the Epic. The sidebar's Epic dropdown updates and the view navigates to Stories for that Epic. The sidebar nav items become active once an Epic is selected.

**When no Epic is selected:** Sidebar Epic dropdown shows "Select Epic...", nav items (Stories through Knowledge Base) are visually muted/disabled, and Settings remains always accessible.

### 3.2 Stories

**Purpose:** Navigate the Story > Task hierarchy within the selected Epic and read requirements documents.

**Data sources:**
- `oraculo tools story list --epic <epic>` — Stories for the selected Epic
- `oraculo tools story get <name> --epic <epic>` — Story requirements markdown
- `oraculo tools task list --epic <epic> --story <story>` — Tasks with status

**Layout:** Master-detail split. Left panel: tree view of Stories within the selected Epic. Each story expands to show its tasks with status badges (pending, in_progress, completed, failed). Right panel: markdown viewer rendering the selected story's requirements or task detail.

**Key interactions:** Select a tree node to load its content in the viewer. Expand/collapse stories. Filter by status. Link from any task to its DAG View position.

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

**Layout:** Tabbed form. Tabs: Project (name, directory path, CLI binary path), Dashboard (refresh interval, theme light/dark, notification preferences).

**Key interactions:** Save persists to `.oraculo/config`. The Settings screen is always accessible regardless of Epic selection.

## 4. Navigation and Layout

**Shell:** Persistent left sidebar with icon + label links for each screen. The sidebar collapses to icon-only on narrow viewports. The sidebar header contains the brand name ("Oraculo") and an Epic dropdown selector.

**Epic dropdown:** Located in the sidebar header, replacing the project subtitle. Shows the currently selected Epic name with a chevron indicator. Clicking opens a dropdown to switch between Epics. When no Epic is selected, displays "Select Epic..." and nav items below are visually muted.

**Sidebar order:**
1. Stories (icon: folder-tree)
2. DAG View (icon: git-branch)
3. Agent Monitor (icon: bot)
4. Approvals (icon: shield-alert, with unread count badge)
5. QA Dashboard (icon: shield-check)
6. Knowledge Base (icon: brain)
— separator —
7. Settings (icon: settings)

**When an Epic is selected:** All nav items (1-6) are active, and the default view is Stories. Switching Epics via the dropdown navigates to Stories of the new Epic.

**When no Epic is selected (Landing):** Nav items 1-6 are grayed/disabled. Only Settings is interactive. The main content area shows the Epic selection grid.

**Responsive behavior:** The sidebar collapses below 1024px viewport width. Master-detail views (Stories, Approvals) stack vertically on narrow screens. The DAG View remains full-width with horizontal scroll.

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
| Landing | CLI: `epic list` | Polling on mount |
| Stories | CLI: `story list`, `story get`, `task list` | File watcher on `.oraculo/epics/**/*.md` |
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
