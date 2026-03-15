# Backend Migration: Go → TypeScript/Bun + Claude Agent SDK

**Date:** 2026-03-15
**Status:** Approved
**Approach:** Full rewrite (no users in production)

## Summary

Replace the entire Go backend (`apps/backend/`) with a TypeScript/Bun application (`apps/orchestrator/`) that:

1. Replicates all CLI, HTTP, WebSocket, SQLite, and MCP functionality
2. Adds a new orchestration layer using the Claude Agent SDK
3. Compiles to a single standalone binary via `bun build --compile`
4. Preserves all contracts with the desktop app (Wails) and frontend (Next.js)

The desktop app (`apps/desktop/`) remains Go/Wails — only ~180 LOC of inlined utility code changes.

## Motivation

- **Agent SDK integration**: The Claude Agent SDK (`@anthropic-ai/claude-agent-sdk`) is TypeScript-native. Using it from Go would require shelling out to a Node/Bun process, adding unnecessary indirection.
- **Deterministic orchestration**: The current skill-based orchestration relies on an LLM to evaluate the DAG and dispatch agents. Moving to code-driven orchestration (TypeScript) eliminates token cost, latency, and non-determinism from the dispatch loop.
- **Unified codebase**: Frontend is TypeScript (Next.js). Backend becomes TypeScript. One language, one toolchain, one team.
- **No migration risk**: No users in production. Full rewrite is safe.

## Proof of Concept

Three tests validated the approach (all successful):

| Test | What | Result | Time |
|------|------|--------|------|
| Simple query | Agent SDK `query()` from within Claude Code | `HELLO_FROM_AGENT` returned | 6.5s |
| Tools | Agent with `Read`, `Glob` tools | Read `package.json`, analyzed deps | 7.6s |
| Full CLI flow | Bash → Agent SDK → tools → JSON result | Listed 11 .ts files, 504 LOC | 16.2s |

Key finding: The `CLAUDECODE=1` environment variable must be stripped from the env passed to `query()` to allow nested Claude Code sessions.

## Architecture

### What Changes

| Component | Before (Go) | After (TypeScript/Bun) |
|-----------|-------------|------------------------|
| `apps/backend/` | Go binary | **Removed** |
| `apps/orchestrator/` | Did not exist | **New** — replaces backend |
| `apps/desktop/` | Imports `spa`, `registry` packages from backend | Inlines ~180 LOC |
| `apps/frontend/` | No change | No change |
| `claude-kit/skills/` | Embedded via `embed.FS`, copied to `.claude/skills/` | **Removed** — prompts built programmatically |
| `npm/cli-*/` | Ships Go binary | Ships Bun binary |

### What Stays the Same

These contracts are preserved for desktop and frontend compatibility:

- `~/.oraculo/servers.json` — same JSON schema `{ project, path, port, pid, started_at }`
- CLI subcommands: `oraculo start` (runs HTTP + MCP via `Promise.all`), `oraculo start http`, `oraculo start mcp`, `oraculo kill`, `oraculo restart`, `oraculo status`
- REST API: all `/api/*` endpoints with same paths and payloads
- WebSocket: `ws://localhost:{port}/ws` with same event types
- Health check: `GET /health`
- Hook endpoints: `POST /hooks/*` with same payloads

### What Does Not Change in Desktop

- `apps/desktop/projects.go` — manages `~/.oraculo/projects.json`, self-contained, no backend imports

## Project Structure

```
apps/orchestrator/
├── src/
│   ├── index.ts                    — CLI entry point
│   ├── cli/
│   │   ├── root.ts                 — Command tree (commander)
│   │   ├── context.ts              — DB injection into subcommands
│   │   ├── lifecycle.ts            — start, kill (lsof+SIGTERM→SIGKILL), restart, status (dashboard), logs (SSE client)
│   │   ├── install.ts              — install (--global/--local/--lang), setup (plugin-mode), uninstall (--purge)
│   │   ├── version.ts              — version command
│   │   ├── hooks/                  — Claude Code command-type hooks (called via .claude/settings.json)
│   │   │   ├── session-start.ts    — Upsert claude_sessions, collect git metadata, auto-start daemon if offline
│   │   │   ├── session-end.ts      — Update claude_sessions ended_at, record session_event
│   │   │   ├── agent-start.ts      — Resolve epic/story/task, POST to /hooks/agent-start
│   │   │   └── task-started.ts     — POST to /hooks/task-started
│   │   └── tools/                  — Agent-facing subcommands (JSON output)
│   │       ├── epic.ts
│   │       ├── story.ts
│   │       ├── task.ts
│   │       ├── memory.ts
│   │       ├── approval.ts
│   │       ├── session.ts
│   │       ├── phase.ts
│   │       ├── design.ts
│   │       ├── review.ts
│   │       └── validation.ts
│   ├── db/
│   │   ├── database.ts             — bun:sqlite, WAL, foreign keys, busy_timeout
│   │   ├── migrations.ts           — PRAGMA user_version (same 9 migrations, each in transaction)
│   │   └── stores/
│   │       ├── agent-store.ts
│   │       ├── approval-store.ts
│   │       ├── epic-store.ts
│   │       ├── memory-store.ts     — FTS5 full-text search
│   │       ├── review-store.ts
│   │       ├── session-event-store.ts
│   │       ├── session-store.ts
│   │       ├── story-store.ts
│   │       ├── task-store.ts
│   │       ├── tool-event-store.ts
│   │       ├── validation-store.ts
│   │       └── version-store.ts
│   ├── domain/
│   │   ├── entities.ts             — TypeScript interfaces (incl. EpicSummary, StorySummary, TaskEnriched)
│   │   ├── enums.ts                — Union types (incl. SessionTypeToPhase mapping)
│   │   ├── errors.ts               — DomainError with codes and context
│   │   └── state.ts                — State machine transitions + phase sequences per session type
│   ├── server/
│   │   ├── http.ts                 — Hono on Bun.serve() — REST + WebSocket upgrade
│   │   ├── ws.ts                   — WebSocket hub (broadcast, per-client buffering)
│   │   ├── broadcaster.ts          — SSE fan-out + 200-entry ring buffer for /api/logs
│   │   ├── routes.ts               — Hono route definitions
│   │   ├── middleware.ts           — CORS (localhost:*), error handling
│   │   └── hooks.ts                — Claude Code lifecycle hook endpoints
│   ├── mcp/
│   │   └── server.ts               — MCP server (request_approval, approval_status)
│   ├── orchestrator/
│   │   ├── dag.ts                  — DAG evaluation, wave calculation
│   │   ├── dispatch.ts             — Agent SDK query() per task
│   │   ├── execute.ts              — Wave loop (main orchestration)
│   │   ├── agents.ts               — Agent definitions (code, research, QA)
│   │   ├── prompts.ts              — Prompt builders (replaces embedded skills)
│   │   └── approval-bridge.ts      — Promise-based blocking gate
│   ├── registry/
│   │   └── registry.ts             — ~/.oraculo/servers.json (mkdirSync atomic lock)
│   ├── config/
│   │   └── config.ts               — .oraculo/config.json
│   ├── logging/
│   │   ├── logger.ts               — pino setup (operational log + sync fallback)
│   │   └── execution-log.ts        — Per-run JSONL files
│   └── utils/
│       ├── output.ts               — writeJSON, writeError, writeTable
│       ├── env.ts                   — cleanEnv() for Agent SDK
│       └── uuid.ts                 — crypto.randomUUID()
├── build.ts                        — Cross-compile script (uses bun-plugin-pino)
├── package.json
└── tsconfig.json
```

## Technology Mapping

### Dependencies

| Purpose | Go Library | TypeScript Replacement |
|---------|-----------|----------------------|
| SQLite | `modernc.org/sqlite` via `database/sql` | `bun:sqlite` (native, sync API) |
| HTTP + WebSocket | `net/http` + `coder/websocket` | `Bun.serve()` + **Hono** (router, ~14KB) |
| CLI framework | `spf13/cobra` | `commander` + `@commander-js/extra-typings` |
| MCP server | `modelcontextprotocol/go-sdk` | `@modelcontextprotocol/sdk` |
| File lock | `gofrs/flock` | `mkdirSync` atomic POSIX lock (zero deps) |
| Concurrency | `sync/errgroup` | `Promise.all()` |
| Agent dispatch | N/A (LLM-driven via Task tool) | `@anthropic-ai/claude-agent-sdk` |
| Logging | `log/slog` | `pino` + `pino-roll` + `pino-pretty` |
| Schema validation | struct tags | `zod` |

### Why Hono

The Go backend uses Go 1.22's built-in pattern routing (`mux.HandleFunc("GET /path/{param}")`). Raw `new URL()` parsing in TypeScript is verbose for 20+ routes. Hono provides:
- ~14KB, ~400K ops/sec — negligible overhead
- Built-in CORS middleware, typed route params
- WebSocket support via `hono/bun` adapter
- SSE support via `hono/streaming`
- Runs natively on `Bun.serve()`

### Why mkdirSync Instead of proper-lockfile

`proper-lockfile` is unmaintained (last publish: v4.1.2, 5+ years old). `mkdirSync` is atomic on POSIX, zero dependencies, and cannot break with Bun updates. Stale lock detection via PID file inside the lock directory + `process.kill(pid, 0)` validation.

### Key API Differences

| Pattern | Go | TypeScript |
|---------|-----|-----------|
| SQLite queries | `row.Scan(&f1, &f2)` | `query().get(id) as Type` |
| Nullable fields | `sql.NullString` | `string \| null` |
| Transactions | `tx.Begin()` / `defer tx.Rollback()` | `db.transaction(() => { ... })` |
| UUID | `crypto/rand` manual | `crypto.randomUUID()` |
| HTTP routing | `mux.HandleFunc("GET /path/{param}")` | Hono `app.get("/path/:param", handler)` |
| WebSocket | Library with hub goroutine | `Bun.serve({ websocket: { ... } })` |
| Blocking gate | Go channel `<-ch` | `new Promise()` + deferred resolve |
| Daemon spawn | `os.StartProcess` + `syscall.Setsid` | `Bun.spawn({ detached: true, stdio: ["ignore","ignore","ignore"] })` |
| Process kill | `syscall.SIGTERM` | `process.kill(pid, "SIGTERM")` |
| Embed files | `//go:embed` + `embed.FS` | Not needed — prompts are code |

## SQLite + Domain Layer

### Database Setup

```typescript
import { Database } from "bun:sqlite";

function openDatabase(path: string): Database {
  const db = new Database(path);
  db.run("PRAGMA journal_mode = WAL");
  db.run("PRAGMA foreign_keys = ON");
  db.run("PRAGMA busy_timeout = 5000");
  db.run("PRAGMA synchronous = NORMAL");
  migrate(db);
  return db;
}
```

Additional PRAGMAs vs Go:
- `busy_timeout = 5000` — handles write contention when HTTP server and orchestrator write concurrently
- `synchronous = NORMAL` — safe with WAL mode, faster than default FULL

### Migrations

Same system: `PRAGMA user_version` as schema version counter. Same 9 migrations creating the same tables: `epics`, `stories`, `tasks`, `task_dependencies`, `task_results`, `validations`, `knowledge` (FTS5), `approvals`, `sessions`, `session_phases`, `claude_sessions`, `agents`, `tool_events`, `epic_versions`, `story_versions`, `reviews`, `session_events`, `approval_comments`.

Each migration runs inside `db.transaction()` for atomicity — if a migration fails mid-way, `user_version` is not incremented.

### Domain Model

TypeScript interfaces mirror Go structs. Union types replace Go typed enums (`type TaskStatus = "pending" | "in_progress" | "completed" | "failed"` — zero runtime cost, better inference than `enum`). `DomainError` with codes and optional context replaces error sentinels. State machine (`canTransition`) is a pure function over a `Record<TaskStatus, TaskStatus[]>`.

### Stores

12 stores, one per domain concept. Each takes a `Database` instance in constructor. Pure SQL queries — no ORM. Same query logic as Go, adapted to bun:sqlite's sync API.

The `MemoryStore` uses FTS5 for full-text search on the `knowledge` table, with INSERT/DELETE triggers to sync the FTS index — same as Go.

### DB Context Injection for CLI Tools

Go uses `context.Context` with `withDB`/`dbFromContext` to inject the database into subcommands. In TypeScript, commander's hook system provides the equivalent:

```typescript
// cli/context.ts
const toolsCommand = new Command("task");
toolsCommand.hook("preAction", (cmd) => {
  cmd.setOptionValue("db", openDatabase(dbPath()));
});
```

All tool subcommands receive the DB instance via the command's option values, matching the Go pattern without global state.

## Orchestrator + Agent SDK

### Execution Model

```
oraculo execute --story <id>
  │
  ├── DAG evaluation (pure code — no LLM)
  │   └── nextWave(): tasks with status=pending and all deps completed
  │
  ├── Wave dispatch (parallel via Promise.all)
  │   ├── query() → code-agent (task 1)
  │   ├── query() → code-agent (task 2)
  │   └── query() → research-agent (task 3)
  │
  ├── Collect results → update SQLite → broadcast WebSocket
  │
  └── Next wave (loop until DAG complete or approval gate blocks)
```

### Agent Types

| Type | Model | Max Turns | Tools | Purpose |
|------|-------|-----------|-------|---------|
| code | claude-sonnet-4-6 | 30 | Read, Write, Edit, Bash, Glob, Grep | Implement tasks (TDD) |
| research | claude-sonnet-4-6 | 15 | Read, Glob, Grep, WebSearch, WebFetch | Codebase analysis |
| qa | claude-sonnet-4-6 | 20 | Read, Glob, Grep, Bash | Validation (clean context) |

### Agent SDK Integration

Each task dispatch:
1. Builds prompt programmatically from task + story context + dependency outputs
2. Calls `query()` with agent definition (model, tools, system prompt, cwd)
3. Streams events — broadcasts progress via WebSocket
4. Collects result — updates SQLite (complete/fail) + saves TaskResult

Environment: `cleanEnv()` strips `CLAUDECODE` from `process.env` to allow nested Claude Code sessions.

### Approval Bridge

Promise-based blocking gate replacing Go channels:
- `request(taskId, type)` creates approval in SQLite, broadcasts WebSocket event, returns `Promise<Verdict>`
- `decide(approvalId, verdict)` resolves the pending Promise, broadcasts decision
- `Map<approvalId, { resolve, reject }>` replaces `map[string]chan VerdictResult`

### QA Circuit Breaker

Deterministic counter in SQLite (not LLM context). After 3 failures on the same task, triggers `requestApproval("qa-escalation")` which blocks the wave loop until human decides.

### Prompt Builders

Skills (`claude-kit/skills/`) become TypeScript functions that build prompts with data from SQLite:
- `buildTaskPrompt(task, storyContext)` — task description + story context + dependency outputs
- `buildQAPrompt(task, results)` — validation criteria + implementation output
- `buildResearchPrompt(task, codebase)` — analysis scope + questions

No more embedded markdown files. Prompts are code — typed, testable, composable.

## Logging System

### Two Streams

| Stream | Purpose | Format | Location | Retention |
|--------|---------|--------|----------|-----------|
| Operational | System health, errors, lifecycle | JSON (pino, rotating) | `~/.oraculo/logs/oraculo.log` | 7 days, max 50MB/file, 7 files |
| Execution | Full trace per agent per run | JSONL (per file) | `~/.oraculo/runs/<runId>/` | 3 days |

### Directory Layout

```
~/.oraculo/
├── logs/
│   ├── oraculo.log                ← active operational log
│   ├── oraculo.2026-03-14.log     ← rotated
│   └── current.log → oraculo.log  ← symlink for tail -f
└── runs/
    └── run_20260315_143022_abc/
        ├── orchestrator.jsonl     ← DAG decisions, wave events
        ├── agent-task-42.jsonl    ← code agent full trace
        ├── agent-task-43.jsonl    ← research agent full trace
        └── agent-task-44.jsonl    ← QA agent full trace
```

### Stack

- **`pino`** — structured JSON logger, child loggers with inherited context
- **`pino-roll`** — rotation by day + size, symlink to `current.log`
- **`pino-pretty`** — colorized output in development only
- **Sync fallback** — if `pino.transport()` worker threads fail on Bun, fall back to `pino.destination()` (sync SonicBoom, no worker thread)

### Bundling Requirement

`bun-plugin-pino` is **required** in `build.ts`. Pino uses `worker_threads` internally for transports, and the worker thread file paths break when bundled without this plugin.

### Correlation Hierarchy

```
runId (execution-level)
  └── waveIndex (wave-level)
       └── taskId + agentType (agent-level)
```

Every log line carries the full context chain. Filterable with `jq`.

### Logged Events

| Category | Events | Level |
|----------|--------|-------|
| Lifecycle | server_started, server_stopped, shutdown_signal | info |
| DAG | wave_calculated, wave_empty, deadlock_detected | info/warn |
| Agent | agent_started, agent_progress, agent_completed, agent_failed | info/error |
| State transitions | task status changes | info |
| Approval gates | approval_requested, approval_decided | info/warn |
| QA circuit breaker | failure_count_incremented, escalation_triggered | warn |
| SDK events | sdk_init, sdk_tool_call, sdk_rate_limit | debug/warn |
| HTTP/WS | request_received, ws_client_connected | debug |
| DB | migration_applied, query_slow (>100ms) | info/warn |

### Rules

- All log writes wrapped in try/catch — logging never crashes an agent
- Automatic pruning of runs older than 3 days on startup
- Never use pino-pretty in production
- SSE ring buffer (200 entries) in `server/broadcaster.ts` feeds `/api/logs` for dashboard real-time view

## CLI + HTTP Server

### CLI Framework

`commander` with `@commander-js/extra-typings` for full type inference. Same command tree as Go/cobra. Same subcommands, same flags, same JSON output format for agent-facing tools.

### HTTP Server

Hono on `Bun.serve()`. Same REST API paths and payloads. CORS middleware (`localhost:*` origins). Idle timeout watchdog. Graceful shutdown on SIGTERM/SIGINT.

### SSE Broadcaster

Dedicated `server/broadcaster.ts` — implements the same pattern as Go's `applog.Broadcaster`:
- 200-entry ring buffer for replay to new SSE subscribers
- Fan-out to all active SSE subscriber channels
- Writes plain-text to stderr and JSON to SSE clients simultaneously
- Integrates with pino via custom transport or `pino.destination()` write hook

### MCP Server

`@modelcontextprotocol/sdk` over stdio. Same two tools: `request_approval` (blocking) and `approval_status` (non-blocking). Zod schemas instead of Go struct tags. May need `--packages=external` in build if dynamic requires break bundling — test early.

### Daemon Management

`Bun.spawn({ detached: true, stdio: ["ignore", "ignore", "ignore"] })` replaces `os.StartProcess` + `syscall.Setsid`. The `stdio: ["ignore", ...]` is critical — without it the parent process will not exit until the child closes its stdio.

Port discovery via `Bun.listen()` probing 3100-3199. Process killing via `process.kill(pid, "SIGTERM")`.

Stale PID validation: before assuming a server is alive from `servers.json`, validate with `process.kill(pid, 0)` (signal 0 tests existence without sending a signal). Clean stale entries on startup.

## Testing Strategy

### Test Runner

**bun:test** — zero config, built into the runtime, 3-10x faster than Vitest for synchronous tests (which SQLite store tests are, since bun:sqlite is sync). Same runtime as production — no impedance mismatch.

### Testing Layers

| Layer | What to Test | Pattern |
|-------|-------------|---------|
| Domain | State machine, error codes, entities | Pure unit tests, no I/O |
| Stores | SQL queries, migrations, FTS5 | In-memory SQLite (`new Database(":memory:")`) |
| CLI | Command parsing, output format | Spawn `bun run src/index.ts` as subprocess, assert JSON output |
| HTTP/WS | Routes, SSE, WebSocket events | `fetch()` against `Bun.serve()` in test, `beforeAll`/`afterAll` for server lifecycle |
| Orchestrator | DAG evaluation, wave dispatch | Mock Agent SDK `query()` via `bun:test` `mock.module()`, assert task state transitions |
| Integration | Full flows (story execute) | Real SQLite + mocked Agent SDK |

### Agent SDK Mocking

```typescript
import { mock } from "bun:test";
mock.module("@anthropic-ai/claude-agent-sdk", () => ({
  query: mock(() => asyncGenerator([
    { type: "result", result: "task completed", is_error: false, num_turns: 3 }
  ])),
}));
```

## Distribution

### Binary Compilation

```typescript
// build.ts
import pinoPlugin from "bun-plugin-pino";

const targets = [
  "bun-darwin-arm64",
  "bun-darwin-x64",
  "bun-linux-x64",
  "bun-linux-arm64",
];

for (const target of targets) {
  await Bun.build({
    entrypoints: ["./src/index.ts"],
    outdir: `./dist/${target}`,
    target,
    plugins: [pinoPlugin()],
  });
}
```

Binary size: ~60MB (vs ~8MB Go). Trade-off accepted — Bun runtime is included. Claude Code itself ships as a Bun binary of similar size.

Test the compiled binary on a clean machine early — Bun can hardcode absolute paths from the build machine.

### npm Packages

Same structure as today:
- `@oraculo/cli` — JS launcher (resolves platform package)
- `@oraculo/cli-{darwin,linux}-{arm64,x64}` — platform-specific Bun binary

The launcher `bin/cli.js` does not change. Only the binary inside platform packages changes from Go to Bun.

### Desktop App Changes

Inline ~180 LOC in `apps/desktop/`:
1. `spa.WithPlaceholders()` — ~40 LOC regex/string replace
2. `spa.Shell()` — ~30 LOC path mapping
3. `registry.DefaultPath()` / `registry.List()` / `registry.WriteAll()` — ~110 LOC JSON file read/write (including stale entry cleanup)
4. Remove Go module dependency on `apps/backend/src/`

The desktop app continues calling `oraculo start http` and `oraculo kill` as subprocess — binary path resolution (`findBinary`) works unchanged.

### Install Command

Simplified: no more copying skills from `embed.FS` to `.claude/skills/`. Just configures `.claude/settings.json` with hooks and MCP server reference.

## What Gets Deleted

- `apps/backend/` — entire Go backend
- `claude-kit/embed.go` — Go embed directive
- `claude-kit/skills/execute/` — replaced by `orchestrator/dispatch.ts` + `prompts.ts`
- Hook-based orchestration in `.claude/settings.json` — replaced by Agent SDK hooks

## Complete CLI Command Reference

Every command from the Go backend must be replicated. This is the authoritative list.

### Lifecycle Commands (`cli/lifecycle.ts`)

| Command | Flags | Behavior |
|---------|-------|----------|
| `oraculo start` | | Runs HTTP + MCP servers concurrently via `Promise.all`, signal handling (SIGTERM/SIGINT) |
| `oraculo start http` | | Spawns detached HTTP daemon, registers in `servers.json`, polls `/health` until ready |
| `oraculo start mcp` | | Starts MCP server on stdio only (no HTTP) |
| `oraculo kill` | | `lsof -ti tcp:<port>` to find PIDs → SIGTERM → wait 5s → SIGKILL → poll until port free |
| `oraculo restart` | | Kill + wait port free + spawn new daemon (self re-exec via own binary path) |
| `oraculo status` | | Dashboard: queries DB for epics list, task status counts per story, pending approvals. Renders ASCII table via `output.writeTable()` |
| `oraculo logs` | | SSE client: connects to running server's `/logs` endpoint, parses `data:` lines, formats `[time] [level] msg` to terminal |
| `oraculo version` | | Prints version string |

### Install Commands (`cli/install.ts`)

| Command | Flags | Behavior |
|---------|-------|----------|
| `oraculo install` | `--global`, `--local`, `--lang <bcp47>` | Creates `.oraculo/`, DB, finds port, writes config, writes `.claude/settings.json` (hooks + MCP) |
| `oraculo setup` | `--lang <bcp47>` | Plugin-mode: hooks-only (no skills), writes `.mcp.json` separately, preserves existing config/port |
| `oraculo uninstall` | `--purge` | Removes oraculo entries from settings.json, removes skill dirs. `--purge` also removes `.oraculo/` directory |

### Hook CLI Commands (`cli/hooks/`)

These are the command-type hooks configured in `.claude/settings.json` and invoked by Claude Code. They read JSON from stdin and proxy to the HTTP server.

| Command | Stdin Fields | Behavior |
|---------|-------------|----------|
| `oraculo hook session-start` | `{ session_id, source }` | Upserts `claude_sessions` in DB, collects metadata (cwd, `git rev-parse --abbrev-ref HEAD`, source). **Auto-starts daemon if offline**: health check → `SpawnDaemon()` → `pollHealth()` until ready. Then POSTs to `/hooks/session-start` |
| `oraculo hook session-end` | `{ session_id }` | Updates `claude_sessions.ended_at`, records `session_event`, POSTs to `/hooks/session-end` |
| `oraculo hook agent-start` | (flags: `--session-id`, `--agent-name`, `--agent-type`, `--task-name`, `--story-name`, `--epic-name`) | POSTs to `/hooks/agent-start` |
| `oraculo hook task-started` | (flags: `--task-name`, `--story-name`, `--epic-name`) | POSTs to `/hooks/task-started` |

### Tool Commands (`cli/tools/`)

All output JSON via `output.writeJSON()`. DB injected via `context.ts`.

**Epic**: `init <name> --description`, `save <name>` (stdin→writes `.oraculo/epics/{name}/requirements.md`), `get <name>` (reads requirements.md), `list`, `update <name> --description`, `delete <name>`, `version <name>` (stdin→creates version + design approval), `versions <name>`

**Story**: `init <name> --epic --description`, `save <name> --epic` (stdin→writes requirements.md), `get <name> --epic` (reads requirements.md), `list --epic`, `update <name> --epic --description`, `update-status <name> --epic --status`, `delete <name> --epic`, `version <name> --epic` (stdin→creates version + design approval), `versions <name> --epic`

**Task**: `init <name> --epic --story --description --depends-on` (with cycle detection), `start <name> --epic --story`, `complete <name> --epic --story` (stdin JSON: summary, logs, skills_used, files_modified → inserts TaskResult), `fail <name> --epic --story --reason`, `get <name> --epic --story`, `list --epic --story`, `delete <name> --epic --story`

**Memory**: `store --domain --category --finding --source --confidence`, `search <query> --domain --limit`, `domains`

**Approval**: `request --type --epic --story` (stdin), `status <id>`, `list --pending`, `verdict <id> --verdict --comment`

**Session**: `init --type --epic --description`, `status --type --epic`, `state --session`, `close --session --reason`

**Phase**: `complete <phase> --session` (stdin, auto-closes session on last phase)

**Design**: `save <story> --epic` (stdin→writes `.oraculo/epics/{epic}/stories/{story}/design.md`), `get <story> --epic` (reads design.md)

**Review**: `create <version-id> --type --verdict --comment` (propagates approval_status to parent entity), `get <review-id>`, `list <version-id> --type`

**Validation**: `save <story> --epic --verdict --findings`

### Filesystem Operations

Several commands read/write markdown files to `.oraculo/epics/` — these are NOT just DB operations:

- `epic save` → writes `.oraculo/epics/{name}/requirements.md`
- `epic get` → reads `.oraculo/epics/{name}/requirements.md`
- `story save` → writes `.oraculo/epics/{epic}/stories/{story}/requirements.md`
- `story get` → reads `.oraculo/epics/{epic}/stories/{story}/requirements.md`
- `design save` → writes `.oraculo/epics/{epic}/stories/{story}/design.md`
- `design get` → reads `.oraculo/epics/{epic}/stories/{story}/design.md`

## Complete HTTP Endpoint Reference

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| `GET` | `/health` | healthHandler | Returns `{"status":"ok","project_name":"..."}` |
| `GET` | `/api/system/status` | systemStatusHandler | Returns `{version, update_available, project_commit, started_at, new_version}`. Binary update detection via file mtime vs server start time. Project commit via `git rev-parse --short HEAD` |
| `GET` | `/api/epics` | listEpics | ListSummaries: complex JOIN with active session, phase, story/task counts |
| `POST` | `/api/epics` | createEpic | Create new epic |
| `GET` | `/api/epics/:epicName/stories` | listStories | ListSummaries: JOIN with task counts (total, completed, failed) |
| `GET` | `/api/epics/:epicName/stories/:storyName/tasks` | listTasks | ListEnriched: batch fetch deps + results |
| `GET` | `/api/epics/:epicName/stories/:storyName/versions` | listVersions | Story versions |
| `GET` | `/api/epics/:epicName/stories/:storyName/reviews` | listReviews | ListByStory (cross-version) |
| `GET` | `/api/epics/:epicName/stories/:storyName/validations` | listValidations | ListByStory |
| `GET` | `/api/approvals` | listApprovals | Supports `?status=pending` and `?epic_id=` filters (uses ListByEpic) |
| `GET` | `/api/approvals/:id` | getApproval | Single approval with details |
| `POST` | `/api/approvals/:id/verdict` | handleVerdict | Submit verdict via bridge. Deletes comments on approve |
| `POST` | `/api/approvals/:id/comments` | createComment | Create inline comment |
| `GET` | `/api/approvals/:id/comments` | listComments | List comments for approval |
| `DELETE` | `/api/approvals/:id/comments/:commentId` | deleteComment | Delete comment |
| `POST` | `/hooks/agent-start` | hookAgentStart | Creates agent record, broadcasts WS |
| `POST` | `/hooks/agent-stop` | hookAgentStop | Stops agent, broadcasts WS |
| `POST` | `/hooks/tool-used` | hookToolUsed | Records tool event, broadcasts WS |
| `POST` | `/hooks/task-completed` | hookTaskCompleted | Records session event, broadcasts WS |
| `POST` | `/hooks/task-started` | hookTaskStarted | Broadcasts WS |
| `POST` | `/hooks/stop` | hookStop | Records session event, broadcasts WS |
| `POST` | `/hooks/teammate-idle` | hookTeammateIdle | Records session event, broadcasts WS |
| `POST` | `/hooks/session-start` | hookSessionStart | Broadcasts session_started |
| `POST` | `/hooks/session-end` | hookSessionEnd | Broadcasts session_ended |
| `GET` | `/ws` | WebSocket upgrade | Hub with broadcast |
| `GET` | `/logs` | SSE stream | Ring buffer replay + live fan-out |

## Complete Store Method Reference

All specialized store methods must be replicated:

- **EpicStore**: Create, GetByName, List, **ListSummaries** (JOIN: active session, phase, story/task counts), Update, Delete, UpdateApprovalStatus
- **StoryStore**: Create, GetByName, List, **ListSummaries** (JOIN: task total/completed/failed counts), Update, Delete, UpdateApprovalStatus
- **TaskStore**: Create (with tx + deps + cycle detection), GetByName, List, **ListEnriched** (batch deps + results), Start, Complete (with tx + result insert), Fail, Delete, addDependencies, **detectCycle** (BFS)
- **MemoryStore**: Store, **Search** (FTS5), Domains
- **ApprovalStore**: Request, GetByID, List, **ListByEpic** (used by `?epic_id=` filter), Verdict, CreateComment, ListComments, DeleteComment, **DeleteCommentsByApproval** (on approve verdict)
- **SessionStore**: Create, Get, Close, CompletePhase, Phases, CurrentPhase, **ActiveByEpic**
- **VersionStore**: CreateEpicVersion (auto-increment + set pending), GetEpicVersion, ListEpicVersions, **LatestEpicVersion**, CreateStoryVersion, GetStoryVersion, ListStoryVersions, **LatestStoryVersion**
- **ReviewStore**: Create (propagates approval_status to parent), GetByID, ListByVersion, **ListByStory** (cross-version, used by API)
- **ValidationStore**: Save, ListByStory
- **AgentStore**: Start (with optional taskID), Stop, ListBySession
- **ToolEventStore**: Record, ListBySession
- **SessionEventStore**: Record, ListBySession

## Config Fields

```typescript
interface Config {
  port: number;              // HTTP server port (default: auto-discovered 3100-3199)
  name: string;              // Project name (default: cwd basename)
  preferred_language: string; // BCP 47 language tag
  skills: {
    design_agent: string[];  // Design agent skill commands
    code_agent: string[];    // Code agent skill commands
  };
}
```

Atomic write: temp file + rename (same as Go).

## Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Bun binary size (~60MB vs ~8MB Go) | Acceptable for dev tooling; Claude Code itself is a Bun binary |
| bun:sqlite FTS5 compatibility | Verified: Bun's bundled SQLite includes FTS5 |
| Agent SDK nested session block | Validated: stripping `CLAUDECODE` env var works (PoC tested) |
| Pino worker threads in compiled binary | Use `bun-plugin-pino` in build; sync `pino.destination()` fallback |
| MCP SDK bundling | May need `--packages=external`; test early in build.ts |
| Sync SQLite blocking event loop | Acceptable for local dev tool; sub-ms queries on SSD; log slow queries >100ms |
| File lock reliability | `mkdirSync` atomic lock + PID file + stale detection via `process.kill(pid, 0)` |
| Desktop app breakage | Minimal coupling: ~180 LOC to inline, all contracts preserved, `projects.go` untouched |
| Hardcoded paths in compiled binary | Test on clean machine; use relative paths |
