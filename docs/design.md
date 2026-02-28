# Oraculo — System Design

## 1. Operating Model

Oraculo operates in two modes depending on the scope of the work.

**Product Engineering (epics):** Discover > Plan > Execute > Validate

**Software Engineering (stories without an epic):** Plan > Execute > Validate

Stories without an associated epic skip Discover — context comes directly from the user. The minimum path is always Plan > Execute > Validate. Standalone stories use a lightweight epic created automatically by the CLI, preserving the hierarchical data model transparently.

### 1.1 Discover

Oraculo instigates the user with questions. It explores the idea, identifies risks, and surfaces edge cases. The output is a requirements document validated by the user.

The Discover phase ends in an **approval gate** (`epic-requirements`): the agent calls `oraculo tools approval request --type epic-requirements`, the dashboard displays the requirements document for review, and the agent enters `awaiting_approval`. Plan does not start until a verdict is received.

_Mandatory for epics. Skipped when a story is submitted directly with sufficient context._

### 1.2 Plan

Requirements are decomposed into tasks. Oraculo models the dependencies as a DAG — identifies what is parallel, what is sequential, and where the constraint lies (TOC). The output is an optimized execution plan.

An optional **approval gate** (`execution-plan`) exists between Plan and Execute: the agent may call `oraculo tools approval request --type execution-plan` to present the DAG for human review before agents are dispatched. This gate is recommended for large or high-risk epics.

### 1.3 Execute

Oraculo assembles a team of agents and delegates. Each agent receives a specific task with clear context: project patterns, existing architecture, expected tests. Agents work in parallel following the DAG. All code is written with TDD.

### 1.4 Validate

A dedicated QA agent reviews the implementation. It verifies: tests pass, project standards were followed, edge cases are covered, the implementation meets the documented requirements. The QA agent is independent from the executing agents — fresh eyes, no bias. If QA rejects, Oraculo returns to the appropriate phase — never forces through, never accepts with caveats.

When the QA agent identifies a critical defect that it cannot resolve autonomously, it escalates via `oraculo tools approval request --type qa-escalation`. This triggers an **approval gate** (`qa-escalation`) that surfaces the issue to a human reviewer through the dashboard. The agent enters `awaiting_approval` until the reviewer delivers a verdict directing the next action.

**Golden rule:** Oraculo never skips the core phases (Plan, Execute, Validate). For epics, Discover is mandatory. Approval gates between phases are mandatory — workflow never advances past a gate without an explicit verdict.

## 2. Documentation as Project Memory

Everything Oraculo produces is recorded. Nothing is lost, but without polluting the project.

### SQLite — Operational State + Accumulated Knowledge

A single SQLite database (`.oraculo/oraculo.db`) within the project serves two purposes:

**Transient operational data** — Task status, dependencies, QA verdicts, approval requests and verdicts, and execution state. This data is essential during active development but can be cleaned after an epic completes.

**Persistent knowledge** — The `knowledge` table accumulates lessons learned, codebase patterns, and conventions discovered across all epics. This data survives epic lifecycle and grows richer over time.

On epic/story completion:
1. Generate a markdown overview (summary of what was implemented, key decisions, QA outcome)
2. Extract lessons learned into the `knowledge` table
3. Operational data for that epic may be cleaned (optional)

### Markdown as Phase Output

Each phase of the operating model produces its own markdown artifact:
- **Discover** outputs a requirements document (`requirements.md` at epic level)
- **Plan** consumes requirements and produces the DAG (tracked in SQLite)
- **Validate** triggers generation of an overview markdown summarizing the implementation

SQLite holds operational state and accumulated knowledge. Markdown files capture product definitions and implementation summaries. The project stays clean — one `.db` file, one `requirements.md` per epic/story, and one overview per validated implementation.

## 3. Claude Code Ecosystem

Oraculo is built entirely on the Claude Code ecosystem. Each native capability is a piece of the system.

### Skills (Commands)

The entry points of Oraculo. Each phase of the operating model is an invocable skill — `/oraculo:epic`, `/oraculo:story`, and so on. The user interacts with Oraculo through these commands.

### Team of Agents

The execution engine. Oraculo uses Claude Code's team functionality to assemble a team of specialized agents — code agents, research agents, and QA agents. Each agent receives well-defined context and scope. Oraculo is the team leader, never an executing member.

### Hooks

The automatic guardians. Hooks ensure standards are respected without relying on goodwill — pre-commit validations, quality checks, formatting. They act as gates that code must pass through.

### CLAUDE.md / Memory

The persistent context. Project patterns, code conventions, architecture — everything agents need to know to produce code that fits the project. Oraculo feeds its agents with this context before any delegation.

### Dashboard

The observation and control surface. A browser-based dashboard that provides visibility into agents, tasks, the DAG, approval gates, and accumulated knowledge. It consumes data through the CLI Trust Layer (never bypasses it) and functions as Mission Control: comprehensive situational awareness with strategic human intervention at approval gates. When an agent enters `awaiting_approval`, the dashboard surfaces the artifact for review and collects the human verdict (`approved`, `rejected`, or `needs_revision`).

**Principle:** Oraculo does not reinvent tools. It orchestrates what Claude Code already offers, maximizing every native capability.

## 4. Target Audience and Team Flow

Oraculo is a team tool, not a solo developer tool.

### Who uses it

- **Product** — Brings ideas and features. Oraculo guides discovery, asks the right questions, documents decisions. Product does not need to know code to use Oraculo in the Discover phase.
- **Development** — Receives already-refined requirements and a structured execution plan. Oraculo orchestrates agents to implement with quality. The dev supervises, does not manually execute every line.
- **Anyone on the team** — Can query SQLite or read the Markdown overviews to understand the history of any feature.

### Typical flow

**Epic flow (Product Engineering):**

1. Someone on the team has an idea or identifies a problem
2. Starts Oraculo with `/oraculo:epic` — questions, refinement, edge cases
3. **Approval gate** (`epic-requirements`) — dashboard presents requirements for human review; workflow pauses until verdict
4. Validated requirements become a plan with tasks in a DAG
5. Optional **approval gate** (`execution-plan`) — dashboard presents the DAG for review before agents are dispatched
6. Agents execute in parallel, following TDD and project standards
7. QA agent validates independently — critical defects trigger **approval gate** (`qa-escalation`) for human direction; if rejected, returns to the appropriate phase

**Story flow (Software Engineering):**

1. A work item is already defined or the user supplies direct context
2. Starts Oraculo with `/oraculo:story` — skips Discover, goes straight to Plan
3. **Approval gate** (`story-definition`) — dashboard presents the story definition for human review; workflow pauses until verdict
4. Agents execute in parallel, following TDD and project standards
5. QA agent validates independently — critical defects trigger **approval gate** (`qa-escalation`) for human direction; if rejected, returns to the appropriate phase

**Oraculo reduces the distance between an idea and quality code.** It does not replace the team — it amplifies the team's ability to think well and execute with rigor.
