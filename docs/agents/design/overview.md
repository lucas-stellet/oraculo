# Agents Design — Overview

## 1. Operating Modes

Oraculo's agent layer supports two operating modes, matching the system-level modes defined in [`docs/design.md`](../../design.md).

**Product Engineering** (epics): Discover > Plan > Execute > Validate

The full mode. An idea enters as a raw problem statement. The orchestrator guides the user through Socratic discovery, produces validated requirements, decomposes them into a DAG of tasks, dispatches agents to execute, and validates all output through independent QA. The Deliver/Merge phase is future work.

**Software Engineering** (stories without an epic): Plan > Execute > Validate

The reduced mode. Context comes directly from the user — a bug report, a feature request with sufficient detail, a refactoring task. Discovery is skipped. The orchestrator decomposes the work into a DAG and proceeds directly to execution and validation.

Both modes share the same agent infrastructure, the same QA pipeline, and the same CLI Trust Layer. The only difference is whether Discover runs.

## 2. Phases from the Agent Perspective

### 2.1 Discover

**Actors:** Orchestrator + Human

The orchestrator asks questions, not agents. No code agents or QA agents are involved. The output is a validated requirements document stored as markdown in `.oraculo/epics/<name>/requirements.md`.

### 2.2 Plan

**Actors:** Orchestrator

The orchestrator decomposes requirements into a DAG — tasks as nodes, dependencies as edges. It identifies what is parallel, what is sequential, and where the bottleneck lies. It assigns skills to each task based on the work required. The CLI validates the DAG (acyclicity, dependency integrity) and persists it to SQLite.

### 2.3 Execute

**Actors:** Code Agents (with skills loaded by the orchestrator)

Each code agent receives a specific task with focused context: relevant files, architectural patterns, conventions from CLAUDE.md. Agents work on the same branch in the same directory — the orchestrator ensures no two agents touch the same files simultaneously. All code tasks use TDD via the TDD skill.

### 2.4 Validate

**Actors:** QA Agent

A dedicated QA agent reviews the implementation with a clean context window. It receives the diff, the specs, and test results — never the code agent's reasoning. QA never fixes; it reports structured findings. If QA rejects, the orchestrator spawns a new code agent with QA's feedback.

## 3. End-to-End Task Lifecycle

A task follows this path through the system:

**1. Birth.** During Plan, the orchestrator emits a DAG. The CLI validates and persists it to SQLite. Each task gets a status of `pending`.

**2. Dispatch.** The orchestrator evaluates the DAG, identifies tasks with all dependencies satisfied, and dispatches them. When possible, multiple tasks run in parallel. When tasks share file dependencies, they run sequentially.

**3. Execution.** A code agent is spawned with the appropriate skills and focused context. It works on the same branch, in the same directory as all other agents. The TDD skill enforces red-green-refactor. On completion, the agent reports its results through the CLI.

**4. Validation.** The QA agent receives the diff, specs, and test results in a clean context. It checks functional correctness, standards compliance, edge cases, and test quality. The CLI records the verdict in SQLite.

**5. Resolution.** If approved, the task is marked complete. If rejected, the orchestrator spawns a new code agent with QA's feedback — fresh context, no memory of the previous attempt. A circuit breaker limits rejection cycles before escalating to the human.

**6. Integration.** Once all tasks in a story are validated, a single markdown summary is generated and committed. The ephemeral SQLite data has served its purpose. The Deliver/Merge phase (merging to mainline) is future work.

## 4. Key Invariants

These hold across all phases and all agents:

1. **CLI validates everything.** Every state transition, every DAG mutation, every verdict passes through the CLI. The CLI is deterministic — its decisions are based on data, not LLM reasoning.

2. **QA is mandatory.** No task is complete without independent QA validation. The QA agent always operates with a clean context window.

3. **Append-only state.** Tasks are never deleted — they are marked failed or completed. QA verdicts are immutable records. The full history is preserved in SQLite for the duration of the epic.

4. **Same-branch execution.** All agents work on the same branch in the same directory. The orchestrator coordinates file access through DAG dependencies. No worktrees, no branch proliferation.

5. **Skills define agent behavior.** The orchestrator assigns skills to agents based on task needs. The skill is the containment mechanism — it defines the workflow, constraints, and quality gates for the agent.

6. **Single markdown artifact.** When a story completes, the system produces one committed markdown file summarizing what was done. The operational detail stays in ephemeral SQLite.
