# Oraculo Agents — Design

## 1. Overview

Oraculo's agent layer is the execution workforce that turns plans into validated code. The orchestrator decomposes requirements into a DAG, dispatches tasks to specialized agents, and validates all output through independent QA. No agent acts without a plan, no code ships without review.

Two operating modes mirror the system-level design:

- **Product Engineering** (epics): Discover > Plan > Execute > Validate
- **Software Engineering** (stories): Plan > Execute > Validate

All agents work on the same branch in the same directory. The orchestrator coordinates file access through DAG dependencies. The CLI Trust Layer validates every state transition. Ephemeral SQLite tracks execution state; committed markdowns capture decisions and outcomes.

Full details: [`design/overview.md`](design/overview.md)

## 2. The Orchestrator

The orchestrator is the only agent that sees the full picture. It plans, delegates, and coordinates — it never executes. Its context window is reserved exclusively for strategic reasoning: syntax errors, compilation logs, and debug traces never enter its context.

During the Plan phase, the orchestrator decomposes requirements into a DAG — tasks as nodes, dependencies as edges. It identifies what is parallel, what is sequential, and where the bottleneck lies. The CLI validates the DAG (acyclicity, dependency integrity) and persists it to SQLite.

The orchestrator assigns skills to each agent based on task needs. TDD for code tasks, playwright for E2E validation, frontend-design for UI work. The skill is the containment mechanism — it defines the agent's workflow, constraints, and quality gates.

Dispatch follows the DAG: all unblocked tasks run in parallel when possible, sequential when file-level coordination requires it. QA throughput governs the pace — the orchestrator limits dispatch so code is produced only as fast as QA can validate it, preventing a backlog of unreviewed work.

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

A circuit breaker limits QA rejection cycles (default 3) before escalating to the human. All QA findings and attempt summaries are preserved for human review.

Full details: [`design/qa-agent.md`](design/qa-agent.md)

## 4.5 Research Agent

Research agents investigate the codebase and external references to provide evidence for decision-making. They are dispatched during Discover (when no prior context exists) and during Plan (for technical feasibility analysis).

Each research agent receives a focused investigation scope: relevant directories, specific patterns to look for, or external references to analyze. They return structured findings — not opinions. The orchestrator integrates findings into the dialogue or the planning process.

Research agents do not write code, run tests, or modify files. They observe, analyze, and report.

Full details: [`design/research-agent.md`](design/research-agent.md)

## 5. Runtime

### SQLite — Operational State + Knowledge

A single SQLite database at `.oraculo/oraculo.db` tracks two categories of data. Operational state (tasks, dependencies, QA verdicts) is transient — essential during execution, cleanable after epic completion. The `knowledge` table is persistent — it accumulates lessons learned across all epics. The database uses the schema from [`docs/cli/design.md`](../cli/design.md) §4.3. It is listed in `.gitignore` and not committed, but its knowledge data is the project's long-term memory.

### Single Markdown Artifact

When a story completes, the system generates one committed markdown file summarizing the implementation: what was built, key decisions, files modified, QA outcome. One file per story — task-level detail is too granular for human reading.

### Memory Model

Three sources, three purposes:
- **CLAUDE.md** — Project conventions, architecture, code style (persistent, committed)
- **Epic markdowns** — Requirements, decisions, outcomes per feature (persistent, committed)
- **SQLite** — Operational state (transient) + accumulated knowledge (persistent), not committed

No three-tier memory architecture, no curation pipeline. A simple knowledge table with full-text search — not a promotion scoring system.

Full details: [`design/runtime.md`](design/runtime.md)

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
