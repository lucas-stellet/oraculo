# Oraculo UI — Philosophy

## 1. Purpose

The UI is Oraculo's **observation and control surface** for humans.

Oraculo's agents work autonomously — decomposing tasks, writing code, running QA — all coordinated by the orchestrator through the CLI Trust Layer. The human needs to see what is happening, understand why, and intervene at critical moments. The UI makes this possible without disrupting the system's operation.

The UI does not introduce new data flows. It consumes data that already exists in the CLI and storage layer, presenting it visually and interactively. Agents and the CLI function identically whether the dashboard is open or not. The UI is an enhancement to human comprehension, never a dependency for system correctness.

## 2. Core Belief

**The human's role is to observe, understand, and approve — the UI exists to make each of those effortless.**

The system already works. Agents execute. The CLI enforces contracts. The DAG manages ordering. What the human needs is not more control — it is more clarity. The UI provides a window into a running system, surfacing the right information at the right moment so the human can make confident decisions without slowing the machine down.

## 3. Theoretical Foundations

The UI's design is grounded in three established concepts from human-computer interaction and systems engineering.

### 3.1 Situational Awareness (Mica Endsley)

Situational awareness operates at three levels: perception of the current state, comprehension of its meaning, and projection of what happens next. The UI must serve all three. It shows which agents are active and what they are doing (perception). It shows how tasks relate to each other in the DAG and what is blocking progress (comprehension). It shows what will happen when the current task completes and what needs human attention next (projection).

**Why it matters:** A dashboard that shows raw data without context forces the human to build a mental model from scratch on every glance. The UI must do that synthesis work — presenting not just facts, but meaning.

### 3.2 Information Radiator (Alistair Cockburn)

An information radiator is a display placed where people can see it effortlessly. It requires no login, no navigation, no active effort. The information is simply there, always current, always visible. The UI operates as Oraculo's information radiator — a persistent, ambient display of system state that the human can absorb at a glance.

**Why it matters:** If checking project status requires navigating menus, running queries, or interpreting logs, it will not be checked often enough. The UI must make status visible by default, not by request.

### 3.3 Human-in-the-Loop

The system handles execution autonomously, but certain decisions require human judgment: approving epic requirements, accepting story definitions, resolving QA escalations, and overriding agent recommendations. The UI is the surface where these decisions are presented, reviewed, and confirmed. The human is in the loop at approval gates — not in the execution path.

**Why it matters:** Oraculo's Socratic discipline demands that critical artifacts receive human validation before work proceeds. The UI must make approval comfortable and informed — presenting all relevant context so the human can decide quickly and confidently, without needing to search for supporting information.

## 4. Design Principles

### 4.1 Observe, Don't Interfere

The UI shows what the system is doing without becoming a bottleneck. Agents do not wait for the UI to render, acknowledge, or sync. The dashboard catches up to system state asynchronously. If the UI crashes, goes offline, or falls behind, the system continues operating normally. The observation layer never constrains the execution layer.

### 4.2 The CLI Remains the Trust Layer

The UI never writes directly to storage or the filesystem. Every mutation flows through the CLI, preserving all validation, contracts, and invariants that the Trust Layer guarantees. The dashboard translates user actions into CLI operations — it is a bridge, not a bypass. This means every action available in the UI is also available at the terminal, and both paths produce identical results.

### 4.3 Approval Is the Human's Moment

The primary purpose of the UI is to make human review comfortable, informed, and decisive. Rich rendering of markdown artifacts, contextual display of related decisions, diff views between document versions, and clear accept/reject actions — everything serves the review decision. For document reviews (epic requirements, story definitions), the dashboard shows versioned snapshots with diffs between versions and accepts `approved`/`rejected` verdicts. For operational gates (design, execution-plan, qa-escalation), the dashboard also supports `needs_revision`. The UI does not rush the human. It presents complete information and waits. This is where the Socratic discipline becomes tangible: the human sees the full reasoning before committing.

### 4.4 Real-Time Without Coupling

Live connections provide updates as agents progress, tasks complete, and QA verdicts arrive. But live updates are a convenience, not a contract. The system produces correct results regardless of whether anyone is watching. The UI can reconnect, refresh, and reconstruct its state entirely from the CLI at any time. No information exists only in the live stream.

This principle is made structurally true by the **two-channel architecture**. Telemetry flows through HTTP hooks — fire-and-forget POST requests that Claude Code sends automatically on every qualifying system event (session start, agent spawn, tool use, task completion). These hooks are non-blocking by design: if the dashboard server is offline or slow, the hook times out silently and the agent continues without interruption. The observation layer cannot interfere with execution even if the server crashes entirely.

Human review gates use a separate channel — MCP tools over stdio — which intentionally *does* block the agent. This is the one moment where the human's response is required before the system proceeds. For operational gates, the agent calls `request_approval` (blocking MCP). For document reviews, the agent creates a version via CLI and polls for reviews. The two channels are not interchangeable: telemetry is automatic and non-blocking (HTTP hooks); review gates are explicit and blocking. Each channel does what it does best, and neither leaks into the other's domain.

The result: the dashboard is a passive observer by default and an active gatekeeper only at approval moments — never accidentally, never by coupling.

### 4.5 Mission Control, Not a Cockpit

The metaphor is NASA Mission Control, not an airplane cockpit. The human observes comprehensive telemetry, makes strategic decisions, and intervenes when escalation demands it. The human does not steer individual agents, edit code through the dashboard, or micromanage task execution. The orchestrator flies the plane. The UI provides the situational awareness that makes strategic oversight possible.

### 4.6 All-in-One

The UI is not a separate application. It lives inside the same binary as the CLI — one artifact that contains the command-line interface, the dashboard server, the agent communication layer, and the frontend assets. There is no separate installation, no additional runtime, no extra process to manage. If the dashboard requires its own setup, its own technology stack, or its own deployment, it becomes a liability that teams will skip. The UI inherits the CLI's zero-dependency distribution principle: one binary, everything included.

### 4.7 Run the Agent, We Do the Rest

The human should never need to start the dashboard manually. When a session begins, the dashboard starts automatically — the server comes up, the browser opens, and the interface is ready before the human asks their first question. The human runs a command, and the dashboard is already there, waiting. This is the Pit of Success applied to the UI: the correct setup is the default setup. No separate terminal tab, no manual server start, no "remember to launch the dashboard first." The system takes care of itself so the human can focus on the work. Each project gets a dedicated port in the 3100-3199 range, persisted in `.oraculo/config`, so the dashboard URL is stable and bookmarkable across sessions.

## 5. Hook-Based Observability

### 5.1 Automatic Telemetry via HTTP Hooks

The dashboard's real-time data arrives through Claude Code's native hook system. When `oraculo install` configures a project, it registers HTTP hooks in `.claude/settings.json` that fire automatically on every qualifying event — no agent instrumentation required. Agents do not call special tools or emit structured logs to feed the dashboard. Claude Code fires the hooks; the dashboard receives them.

The hooks covered by the HTTP channel:

| Hook | Event |
|---|---|
| `SubagentStart` | An agent was spawned by the orchestrator |
| `SubagentStop` | An agent completed or failed |
| `PostToolUse` (mutation tools only) | A file was edited, written, or a shell command ran |
| `TaskCompleted` | A task was marked as completed |
| `Stop` | An agent is stopping |
| `TeammateIdle` | A teammate agent became idle |
| `SessionEnd` | The Claude Code session ended |

`SessionStart` is handled by a command hook (`oraculo hook session-start`) rather than an HTTP hook. This is because session start requires a health check and must be able to print a warning to the user if the server is offline — HTTP hooks cannot produce output visible to the user.

### 5.2 Graceful Degradation

Every HTTP hook operates under a graceful degradation policy. When the server is online, it persists the event to SQLite and broadcasts to connected WebSocket clients. When the server is offline, Claude Code logs a non-blocking warning and the agent continues. Hooks use a 5-second timeout; if the server does not respond in time, the hook is abandoned and the agent proceeds.

This means the dashboard can go offline, restart, or fall behind at any time without affecting the system's operation. When the server comes back up, future events resume. Historical events that occurred while the server was offline are not replayed — the dashboard shows what it observed, not a complete reconstruction of the session. For authoritative state, the CLI remains the source of truth.

### 5.3 What the Dashboard Observes

HTTP hooks deliver metadata, not content. The `PostToolUse` hook records *which* tool ran and *which file* was affected — not what was written, not the diff, not the command text. This is a deliberate privacy and storage decision: the dashboard shows the shape of agent activity without becoming a surveillance system for code content.

The three new SQLite tables that support hook-based observability:

- **`sessions`** — One row per Claude Code session, tracking model, working directory, and lifetime
- **`agents`** — One row per agent start event, tracking name, type, status, and duration
- **`tool_events`** — One row per mutation tool use, tracking which tool and which file (metadata only)

These tables are populated exclusively by HTTP hooks and read exclusively by the dashboard's REST API. They are append-only telemetry — the CLI does not write to them, and the dashboard does not expose them as editable state.

### 5.4 Two-Channel Architecture Summary

```
Channel 1: HTTP Hooks (automatic telemetry)
═══════════════════════════════════════════
Claude Code ──POST──> HTTP server ──> SQLite + WebSocket broadcast
                      (fire-and-forget, 200 empty body)

Channel 2: MCP (interactive approval gates)
═══════════════════════════════════════════
Claude Code ──stdio──> MCP server ──> SQLite + Go channel (blocks)
                                          │
Dashboard ──POST /api/approvals/:id/verdict──> UPDATE SQLite + unblock
                                          │
                       MCP server <── Go channel ──> Claude Code (resumes)
```

Both channels write to the same SQLite database and share the same WebSocket broadcast mechanism. Both run inside the same Go binary. The distinction is behavioral: HTTP hooks are non-blocking telemetry; MCP calls are blocking workflow gates. The dashboard consumes both — observing through the hook stream, acting through the approval API.

## 6. Relationship to Other Layers

The UI sits at the outermost layer of Oraculo's architecture. It depends on every layer below it but nothing depends on it.

**CLI Trust Layer** is the UI's sole interface to system state. The dashboard server calls CLI commands and presents their output. The UI inherits all of the CLI's validation guarantees by never circumventing them.

**Agents** are observed but never directed through the UI. The human sees agent activity, task progress, and QA results. Intervention happens through approval gates defined by the workflow, not through ad-hoc commands to running agents.

**Commands and Skills** define the workflows that produce the artifacts the UI displays. The UI renders epic requirements, story definitions, and task statuses — all generated by skills and persisted through the CLI. The UI does not define new workflows; it makes existing ones visible.

**Both humans and agents go through the same CLI.** The UI makes the CLI's data visual and interactive. It is a different lens on the same source of truth, not a parallel data path.

## 7. What the UI Is Not

The UI is **not a project management tool**. It does not replace Jira, Linear, or any external tracker. It shows Oraculo's internal state — epics, stories, tasks, agent activity — for the humans who are actively working with the system.

The UI is **not a code editor**. Humans do not write or modify code through the dashboard. Code is produced by agents under TDD discipline and reviewed through QA. The UI shows results and diffs, not editable source files.

The UI is **not required for Oraculo to function**. Every capability the UI exposes exists in the CLI first. A team can run Oraculo entirely from the terminal. The UI adds visibility and comfort — it does not add capability.

The UI is **not a data source**. It reads from the CLI and displays what it finds. It never generates, transforms, or stores authoritative data. If the UI disagrees with the CLI, the UI is wrong.
