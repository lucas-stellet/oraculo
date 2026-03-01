# Oraculo Agents — Design

## 1. Overview

Oraculo's agent layer is the execution workforce that turns plans into validated code. The orchestrator decomposes requirements into a DAG, dispatches tasks to specialized agents, and validates all output through independent QA. No agent acts without a plan, no code ships without review.

Two operating modes mirror the system-level design:

- **Product Engineering** (epics): Discover > Plan > Execute > Validate
- **Software Engineering** (stories): Plan > Execute > Validate

All agents work on the same branch in the same directory. The orchestrator coordinates file access through DAG dependencies. The CLI Trust Layer validates every state transition. SQLite tracks operational state and accumulated knowledge; committed markdowns capture decisions and outcomes.

Full details: [`design/overview.md`](design/overview.md)

## 2. The Orchestrator

The orchestrator is the only agent that sees the full picture. It plans, delegates, and coordinates — it never executes. Its context window is reserved exclusively for strategic reasoning: syntax errors, compilation logs, and debug traces never enter its context.

During the Plan phase, the orchestrator decomposes requirements into a DAG — tasks as nodes, dependencies as edges. It identifies what is parallel, what is sequential, and where the bottleneck lies. The CLI validates the DAG (acyclicity, dependency integrity) and persists it to SQLite.

The orchestrator assigns skills to each agent based on task needs. TDD for code tasks, playwright for E2E validation, frontend-design for UI work. The skill is the containment mechanism — it defines the agent's workflow, constraints, and quality gates.

Dispatch follows the DAG: all unblocked tasks run in parallel when possible, sequential when file-level coordination requires it. QA throughput governs the pace — the orchestrator limits dispatch so code is produced only as fast as QA can validate it, preventing a backlog of unreviewed work. When a task enters `awaiting_approval`, the orchestrator tracks that state and suspends all dependent dispatches until a human verdict is received.

Full details: [`design/orchestrator.md`](design/orchestrator.md)

## 3. Code Agent

A single code agent type handles all implementation tasks. There is no separate test-author or implementer — one agent writes tests and implementation, guided by the TDD skill's red-green-refactor loop.

Each code agent receives focused context: the task description, relevant files, project conventions from CLAUDE.md, and story/epic requirements. Less context produces better code — agents receive the minimum viable information, not the entire repository.

All code agents work on the same branch in the same directory. The orchestrator ensures no two agents touch the same files simultaneously through DAG dependencies. Scope is enforced by skill instructions and task descriptions, not by filesystem ACLs.

On QA rejection, the orchestrator spawns a **new** code agent with QA's findings and a fresh context. Agents that receive rejection feedback on their own work tend to defend previous decisions; a new agent treats QA's findings as ground truth.

Full details: [`design/code-agent.md`](design/code-agent.md)

## 4. QA Agent

The QA agent validates all code output with a completely clean context window — no memory of the code agent's reasoning, no access to its conversational history. It receives exactly the diff, the specs, and test results. This isolation breaks the sycophancy cycle where an agent agrees with its own errors.

QA checks functional correctness, standards compliance, edge cases, test quality, and scope. It **never fixes code** — it produces structured findings that the orchestrator routes to a new code agent. This separation ensures QA can never compromise its own verdicts.

A circuit breaker limits QA rejection cycles (default 3) before the QA agent submits a `qa-escalation` approval request to the dashboard. The agent enters `awaiting_approval` and halts dispatch on that task until the human issues a verdict (`approved`, `rejected`, or `needs_revision`). All QA findings and attempt summaries are preserved for human review.

Full details: [`design/qa-agent.md`](design/qa-agent.md)

## 4.5 Research Agent

Research agents investigate the codebase and external references to provide evidence for decision-making. They are dispatched during Discover (when no prior context exists) and during Plan (for technical feasibility analysis).

Each research agent receives a focused investigation scope: relevant directories, specific patterns to look for, or external references to analyze. They return structured findings — not opinions. The orchestrator integrates findings into the dialogue or the planning process.

Research agents do not write code, run tests, or modify files. They observe, analyze, and report.

Full details: [`design/research-agent.md`](design/research-agent.md)

## 5. Runtime

### SQLite — Operational State + Knowledge

A single SQLite database at `.oraculo/oraculo.db` tracks two categories of data. Operational state (tasks, dependencies, QA verdicts) is transient — essential during execution, cleanable after epic completion. The `knowledge` table is persistent — it accumulates lessons learned across all epics. The database uses the schema from [`docs/cli/design.md`](../cli/design.md) §4.3. It is listed in `.gitignore` and not committed, but its knowledge data is the project's long-term memory.

In addition to the existing schema, three tables support HTTP hook telemetry:

- **`sessions`** — One row per Claude Code session observed. Records the session ID, model, working directory, start time, and end time. Written by the `SessionStart` command hook and the `SessionEnd` HTTP hook.
- **`agents`** — One row per agent lifecycle event. Written by `SubagentStart` and `SubagentStop` HTTP hooks. Tracks agent name, inferred type (code, qa, research, orchestrator), status, and start/stop timestamps.
- **`tool_events`** — One row per mutation tool invocation (Bash, Edit, Write, NotebookEdit). Written by the `PostToolUse` HTTP hook. Stores only metadata: session ID, tool name, file path, and timestamp. No content, no diffs, no command text.

### Single Markdown Artifact

When a story completes, the system generates one committed markdown file summarizing the implementation: what was built, key decisions, files modified, QA outcome. One file per story — task-level detail is too granular for human reading.

### Memory Model

Three sources, three purposes:
- **CLAUDE.md** — Project conventions, architecture, code style (persistent, committed)
- **Epic markdowns** — Requirements, decisions, outcomes per feature (persistent, committed)
- **SQLite** — Operational state (transient) + accumulated knowledge (persistent), not committed

No three-tier memory architecture, no curation pipeline. A simple knowledge table with full-text search — not a promotion scoring system.

Full details: [`design/runtime.md`](design/runtime.md)

## 7. HTTP Hooks Integration

### Overview

Agents are tracked through Claude Code's native HTTP hook system rather than explicit MCP tool calls. This is automatic — the hooks fire on every qualifying event without any agent-side instrumentation. No agent needs to announce itself; the infrastructure observes it.

The previous `notify_agent_state` MCP tool is **removed**. It required agents to call MCP explicitly to report their own status — a form of instrumentation that added noise to agent context and could be forgotten or skipped. `SubagentStart` and `SubagentStop` hooks replace it entirely.

### Hook-to-Agent Mapping

| Claude Code Hook | Agent Event | SQLite Write | WebSocket Broadcast |
|---|---|---|---|
| `SubagentStart` | Agent spawned by orchestrator | INSERT into `agents` (status = active) | `{ "type": "agent_started" }` |
| `SubagentStop` | Agent completed or failed | UPDATE `agents` (status, stopped_at) | `{ "type": "agent_stopped" }` |
| `PostToolUse` (Bash\|Edit\|Write\|NotebookEdit) | Agent mutated a file or ran a command | INSERT into `tool_events` | `{ "type": "tool_used" }` |
| `TaskCompleted` | Agent marked a task complete | (read by existing task system) | `{ "type": "task_completed" }` |
| `SessionStart` (command) | Session begins | INSERT into `sessions` | — |
| `SessionEnd` | Session ends | UPDATE `sessions` (ended_at) | `{ "type": "session_ended" }` |

### Agent Lifecycle via Hooks

When the orchestrator spawns a subagent, Claude Code fires `SubagentStart`. The HTTP server receives the event and inserts a row into the `agents` table with `status = 'active'`. The dashboard's Agent Monitor receives an immediate WebSocket push — no polling required.

When the subagent finishes (successfully or not), Claude Code fires `SubagentStop`. The HTTP server updates the `agents` row with the final status and `stopped_at` timestamp. The dashboard reflects the change in real time.

The orchestrator does not call any MCP tool to report agent state changes. The infrastructure observes them automatically.

### Tool Event Tracking

Every time an agent uses a mutation tool (Bash, Edit, Write, NotebookEdit), the `PostToolUse` hook fires after the tool completes. Only metadata is persisted — **no content, no diffs, no command text**. This is a deliberate privacy and storage decision: the dashboard shows *what was touched*, not *what was written*.

Read-only tools (Read, Glob, Grep, WebFetch) do not fire the hook. The matcher `Bash|Edit|Write|NotebookEdit` restricts tracking to operations that change state.

### Non-Blocking Telemetry

All HTTP hooks use a 5-second timeout and return `200` with an empty body. If the Oraculo server is offline, Claude Code logs a non-blocking warning and the agent continues unaffected. Telemetry is best-effort — agents are never blocked by observation infrastructure.

This is a hard architectural constraint: the observation layer cannot degrade agent throughput. If it can block agents, it will eventually do so at the worst possible moment.

### MCP Tools — Reduced Scope

With HTTP hooks handling telemetry, the MCP server's tool set is reduced to the approval gate workflow only:

| MCP Tool | Status | Purpose |
|---|---|---|
| `request_approval` | Kept | Interactive approval gate — blocks until human verdict |
| `approval_status` | Kept | Polling fallback for crash recovery |
| `notify_agent_state` | Removed | Replaced by SubagentStart/Stop HTTP hooks |
| `register_project` | Removed | Handled by `oraculo install` |

Agents only call MCP tools when they need a human decision. All other communication with the dashboard is automatic.

## 6. Future Work

The research phase produced extensive findings across six topics. Many capabilities were deferred in favor of a simpler architecture. The most significant deferred items:

- **Deliver/Merge phase** — Automated integration of validated code into mainline
- **Worktree isolation** — Physical filesystem isolation for true parallel execution
- **Separate test-author/implementer** — Stronger TDD guarantees through agent separation
- **Adversarial QA** — Active exploitation testing with the Executable Proof pattern
- **Mutation testing** — Automated validation of test quality via code mutation
- **Heterogeneous models** — Different LLM families for different agent roles
- **Rich memory system** — Three-tier architecture with curation pipeline
- **Advanced dispatch** — Critical-path scoring, shifting bottleneck detection
- **Event sourcing** — Append-only event log as single source of truth
- **RBAC governance** — Role-based access control on the SQLite blackboard

Each item includes research findings that will inform future implementation.

Full details: [`design/future-work.md`](design/future-work.md)
