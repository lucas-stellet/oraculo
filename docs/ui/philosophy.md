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

The primary purpose of the UI is to make human review comfortable, informed, and decisive. Rich rendering of markdown artifacts, contextual display of related decisions, diff views for changes, and clear accept/reject actions — everything serves the approval decision. The UI does not rush the human. It presents complete information and waits. This is where the Socratic discipline becomes tangible: the human sees the full reasoning before committing.

### 4.4 Real-Time Without Coupling

Live connections provide updates as agents progress, tasks complete, and QA verdicts arrive. But live updates are a convenience, not a contract. The system produces correct results regardless of whether anyone is watching. The UI can reconnect, refresh, and reconstruct its state entirely from the CLI at any time. No information exists only in the live stream.

### 4.5 Mission Control, Not a Cockpit

The metaphor is NASA Mission Control, not an airplane cockpit. The human observes comprehensive telemetry, makes strategic decisions, and intervenes when escalation demands it. The human does not steer individual agents, edit code through the dashboard, or micromanage task execution. The orchestrator flies the plane. The UI provides the situational awareness that makes strategic oversight possible.

### 4.6 All-in-One

The UI is not a separate application. It lives inside the same binary as the CLI — one artifact that contains the command-line interface, the dashboard server, the agent communication layer, and the frontend assets. There is no separate installation, no additional runtime, no extra process to manage. If the dashboard requires its own setup, its own technology stack, or its own deployment, it becomes a liability that teams will skip. The UI inherits the CLI's zero-dependency distribution principle: one binary, everything included.

### 4.7 Run the Agent, We Do the Rest

The human should never need to start the dashboard manually. When a session begins, the dashboard starts automatically — the server comes up, the browser opens, and the interface is ready before the human asks their first question. The human runs a command, and the dashboard is already there, waiting. This is the Pit of Success applied to the UI: the correct setup is the default setup. No separate terminal tab, no manual server start, no "remember to launch the dashboard first." The system takes care of itself so the human can focus on the work. Each project gets a dedicated port in the 3100-3199 range, persisted in `.oraculo/config`, so the dashboard URL is stable and bookmarkable across sessions.

## 5. Relationship to Other Layers

The UI sits at the outermost layer of Oraculo's architecture. It depends on every layer below it but nothing depends on it.

**CLI Trust Layer** is the UI's sole interface to system state. The dashboard server calls CLI commands and presents their output. The UI inherits all of the CLI's validation guarantees by never circumventing them.

**Agents** are observed but never directed through the UI. The human sees agent activity, task progress, and QA results. Intervention happens through approval gates defined by the workflow, not through ad-hoc commands to running agents.

**Commands and Skills** define the workflows that produce the artifacts the UI displays. The UI renders epic requirements, story definitions, and task statuses — all generated by skills and persisted through the CLI. The UI does not define new workflows; it makes existing ones visible.

**Both humans and agents go through the same CLI.** The UI makes the CLI's data visual and interactive. It is a different lens on the same source of truth, not a parallel data path.

## 6. What the UI Is Not

The UI is **not a project management tool**. It does not replace Jira, Linear, or any external tracker. It shows Oraculo's internal state — epics, stories, tasks, agent activity — for the humans who are actively working with the system.

The UI is **not a code editor**. Humans do not write or modify code through the dashboard. Code is produced by agents under TDD discipline and reviewed through QA. The UI shows results and diffs, not editable source files.

The UI is **not required for Oraculo to function**. Every capability the UI exposes exists in the CLI first. A team can run Oraculo entirely from the terminal. The UI adds visibility and comfort — it does not add capability.

The UI is **not a data source**. It reads from the CLI and displays what it finds. It never generates, transforms, or stores authoritative data. If the UI disagrees with the CLI, the UI is wrong.
