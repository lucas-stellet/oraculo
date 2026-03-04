# Oraculo Agents — Philosophy

## 1. Purpose

The agent layer is Oraculo's execution workforce.

Oraculo operates through a team of specialized agents — each with a clear role, strict boundaries, and no authority beyond its scope. The orchestrator plans and delegates. Code agents implement. QA agents validate. No agent does another's job.

This document defines the beliefs that govern how agents work together. It is the foundation for all design decisions about agent roles, tools, communication, and coordination.

## 2. Core Belief

**Agents are powerful reasoners but unreliable operators. Every agent must be contained by deterministic boundaries.**

An agent left unconstrained will hallucinate paths, weaken its own test assertions, drift from project conventions, and confidently produce broken output. The system does not prevent this by hoping for better prompts — it prevents this by design. The CLI Trust Layer enforces contracts. Skills enforce discipline (TDD, code standards). The DAG enforces ordering. Approval gates enforce human oversight at critical junctures — mandatory pauses where no work proceeds without an explicit verdict. Same-branch execution with skill-based containment replaces the need for physical isolation — agents receive precise instructions about what files to touch, what tests to run, and what conventions to follow. Every agent operates inside a cage of deterministic constraints that makes the right behavior the only possible behavior.

## 3. The Orchestrator

The orchestrator is the only agent that sees the full picture. It receives project context, decomposes work into a DAG, dispatches tasks, assigns skills to agents based on task needs, and coordinates validation. It never writes code, runs tests, or touches the filesystem.

**The orchestrator's context window is sacred.** Syntax errors, compilation logs, and debug traces destroy strategic reasoning. The orchestrator plans and delegates — tactical execution is pushed entirely to subordinate agents and the CLI Trust Layer. This is not a preference; mixing planning and execution in the same context measurably degrades both.

**The execution plan is a computable graph, not a text list.** Tasks are nodes, dependencies are edges. The orchestrator evaluates the graph continuously, dispatching all unblocked nodes when possible. When an agent encounters an unexpected obstacle, the orchestrator mutates the DAG dynamically — pruning invalid branches and scheduling new nodes — without regenerating the entire plan.

**Code generation is subordinated to QA throughput.** The orchestrator does not dispatch all parallel tasks just because the DAG allows it. It limits dispatch rate so that code is produced only as fast as QA can validate it. Unvalidated work-in-progress accumulates context drift and exponential rework costs. The orchestrator treats QA capacity as the system's governing constraint.

**The orchestrator selects skills for each agent.** Based on the nature of a task, the orchestrator loads the appropriate skills — TDD for code tasks, playwright for E2E validation, frontend-design for UI work. The skill is the containment mechanism: it defines the agent's workflow, constraints, and quality gates.

## 4. Code Agents

Code agents are the hands of the system. They receive a focused task, a constrained context, and clear instructions. They implement, nothing more.

**Less context produces better code.** Code agents receive the minimum viable information for their task — relevant interfaces, architectural patterns, test context, and the architectural design document from the Plan phase — not the entire repository. The TDD skill is always loaded, plus any additional skills specified in the story or task configuration. Targeted context reduces token cost and produces more deterministic output. The CLI builds this working set.

**Agents work on the same branch in the same directory.** All code agents operate on the shared working branch. The orchestrator manages file-level coordination — ensuring no two agents modify the same files simultaneously. This eliminates worktree complexity while the DAG's dependency structure prevents conflicts.

**TDD is enforced through skills, not separate agents.** A single code agent receives the TDD skill, which enforces the red-green-refactor loop: write a failing test, implement to make it pass, refactor. The skill instructions prevent the agent from skipping steps or weakening assertions. This is simpler than separate test-author/implementer agents while preserving TDD discipline.

**TDD is the anti-hallucination mechanism.** The red-green-refactor loop forces the agent to observe a test failing against real dependencies before writing implementation. This anchors a probabilistic model to deterministic reality.

## 5. QA Agents

QA agents are the immune system. They validate all output with fresh eyes, no shared memory with the agent that produced it, and no incentive to agree.

**Independence is architectural, not aspirational.** The QA agent operates with a completely clean context window — no memory of the generation process, no access to the code agent's reasoning. It receives the diff, the specs, and test results. This is what breaks the sycophancy cycle where an agent tends to agree with its own errors. No task can be marked complete without passing through QA.

**QA never fixes — it only reports.** When the QA agent finds issues, it produces a structured finding. It never modifies code. The orchestrator spawns a new code agent with QA's feedback to address the issues. This separation prevents QA from compromising its own verdicts.

**The Trust Layer is the final arbiter, not the QA agent.** The recursive problem of "who watches the watchmen" is not solved by adding more watchers. It is solved by deterministic reality. Compilation, test execution, and contract assertions do not have opinions. The QA agent produces findings and evidence; the CLI verifies them. If the output violates the contract, it is rejected — no debate.

**A circuit breaker triggers a dashboard approval gate.** When QA rejection cycles exceed the configured threshold, the system does not silently fail or spawn more agents. The QA agent submits a `qa-escalation` approval request to the dashboard, enters `awaiting_approval`, and halts further dispatch on that task. The human reviews the accumulated findings and issues a verdict — `approved`, `rejected`, or `needs_revision` — before work resumes.

## 5.5 Research Agents

Research agents investigate the codebase and external references to provide evidence for decision-making. They are dispatched during Discover and during Plan's Design sub-phase.

**Codebase Research Agent.** Investigates the existing codebase — architecture, patterns, conventions, and relevant file structures — to provide evidence for planning and design decisions. Read-only — never modifies files.

**Web Research Agent.** Researches external technologies, best practices, libraries, and integration approaches for features not yet present in the project. Provides structured findings about available solutions, trade-offs, and recommendations. Read-only — never modifies files.

Research agents do not write code, run tests, or modify files. They observe, analyze, and report. Their findings are synthesized by the orchestrator into the architectural design document during the Plan phase.

## 6. Memory

Memory is simple and practical. It serves agents without burdening the system.

**CLAUDE.md is the persistent context.** Project patterns, code conventions, architecture decisions — everything agents need to know lives in CLAUDE.md. It is the single source of project-level context, always available, always up to date.

**Epic markdowns are the accumulated intelligence.** Requirements documents, story definitions, and task summaries capture the decisions and reasoning of each feature. These are committed to the repository and serve as the historical record.

**SQLite tracks operational state and accumulated knowledge.** A single database per project (`.oraculo/oraculo.db`) holds two categories of data. Transient operational data — task status, dependencies, QA verdicts — is essential during execution and can be cleaned after epic completion. Persistent knowledge — lessons learned, codebase patterns, conventions — accumulates across all epics and survives their lifecycle. The database is not committed (`.gitignore`), but the knowledge it contains is the project's long-term memory.

**Simple knowledge store, not a complex memory system.** The `knowledge` table with full-text search provides a single, queryable store for codebase findings. There is no three-tier architecture, no curation pipeline, no promotion scoring. This is a practical knowledge store — not the deferred rich memory system described in future-work.md.

## 7. Communication Model

Agents communicate with the outside world through two dedicated channels — one automatic, one interactive. These channels are designed around a fundamental constraint: telemetry must never block execution, but approvals must.

**Channel 1 — HTTP Hooks (automatic telemetry).** Claude Code fires hooks automatically on system events: session start and end, agent spawn and completion, tool use. These hooks POST to the Oraculo HTTP server with no agent-side instrumentation required. The agent never calls anything. The hook fires, the server persists metadata to SQLite, and connected dashboard clients receive a WebSocket broadcast. If the server is offline, the hook fails silently — the agent continues unaffected. Telemetry is best-effort by design. The observation layer cannot interfere with execution under any failure mode.

**Channel 2 — MCP (interactive approval gates).** When a workflow reaches a mandatory pause, the orchestrator calls the `request_approval` MCP tool. This is a blocking call — the agent suspends and waits. The dashboard displays the artifact, a human reviews it, and a verdict (`approved`, `rejected`, or `needs_revision`) flows back through the MCP tool response to unblock the agent. Approvals are the only blocking operations in the system. Everything else is fire-and-forget.

**Why two channels.** HTTP hooks are automatic — they require no agent code — but they cannot block or return data. MCP tools are interactive — they can block and return responses — but they require explicit agent calls. Telemetry needs automatic and non-blocking (HTTP hooks). Approvals need explicit and blocking (MCP). Each channel does exactly what it does best, and neither channel is asked to do what the other does better.

**Agents are not instrumented for observability.** Earlier designs required agents to call `notify_agent_state` to report their status. This was fragile — a crashed or stuck agent would stop reporting, silently degrading the dashboard. Under the hook-based model, Claude Code itself fires `SubagentStart` and `SubagentStop` events regardless of what the agent is doing. The observation layer is decoupled from agent logic entirely. An agent that crashes is still reported as stopped.

**The dashboard cannot interfere with agents through telemetry.** A slow dashboard, an overloaded WebSocket broadcast, or a full SQLite disk cannot delay agent work. The only path from the dashboard to an agent is through an approval gate — an explicit, intentional blocking point the orchestrator created on purpose. Outside of approval gates, agents and the dashboard are causally independent.

## 8. Standardized Agent Prompts

The orchestrator uses a fixed template for code agent prompts. Rather than composing prompts ad hoc, the template defines a deterministic structure with slots that are filled via CLI commands and configuration: task description, relevant files, design document, CLAUDE.md conventions, skill instructions, and story requirements.

This ensures every code agent receives the same baseline context regardless of which task it handles. The template is not a suggestion — it is the only way code agents are prompted. Deviations would require changing the template itself, not working around it per-task.

The standardized template eliminates a class of orchestrator errors where forgotten context or inconsistent formatting leads to agent drift. When every agent starts from the same structure, the quality floor is deterministic.

## 9. What This Document Is Not

This document defines beliefs, not implementation. It does not specify CLI commands, database schemas, DAG formats, or agent prompt templates. Those belong in the design document.

This document also does not replace the existing philosophy documents. The core principles (docs/philosophy.md) and the CLI Trust Layer philosophy (docs/cli/philosophy.md) remain authoritative. This document extends them into the domain of multi-agent coordination.
