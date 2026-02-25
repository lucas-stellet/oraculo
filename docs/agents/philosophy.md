# Oraculo Agents — Philosophy

## 1. Purpose

The agent layer is Oraculo's execution workforce.

Oraculo operates through a team of specialized agents — each with a clear role, strict boundaries, and no authority beyond its scope. The orchestrator plans and delegates. Code agents implement. QA agents validate. No agent does another's job.

This document defines the beliefs that govern how agents work together. It is the foundation for all design decisions about agent roles, tools, communication, and coordination.

## 2. Core Belief

**Agents are powerful reasoners but unreliable operators. Every agent must be contained by deterministic boundaries.**

An agent left unconstrained will hallucinate paths, weaken its own test assertions, drift from project conventions, and confidently produce broken output. The system does not prevent this by hoping for better prompts — it prevents this by design. The CLI Trust Layer enforces contracts. Git worktrees enforce isolation. The DAG enforces ordering. TDD enforces correctness. Every agent operates inside a cage of deterministic constraints that makes the right behavior the only possible behavior.

## 3. The Orchestrator

The orchestrator is the only agent that sees the full picture. It receives project context, decomposes work into a DAG, assembles agent teams, dispatches tasks, and coordinates validation. It never writes code, runs tests, or touches the filesystem.

**The orchestrator's context window is sacred.** Syntax errors, compilation logs, and debug traces destroy strategic reasoning. The orchestrator plans and delegates — tactical execution is pushed entirely to subordinate agents and the CLI Trust Layer. This is not a preference; mixing planning and execution in the same context measurably degrades both.

**The execution plan is a computable graph, not a text list.** Tasks are nodes, dependencies are edges. The orchestrator evaluates the graph continuously, dispatching all unblocked nodes in parallel. When an agent encounters an unexpected obstacle, the orchestrator mutates the DAG dynamically — pruning invalid branches and scheduling new nodes — without regenerating the entire plan.

**Code generation is subordinated to QA throughput.** The orchestrator does not dispatch all parallel tasks just because the DAG allows it. It limits dispatch rate so that code is produced only as fast as QA can validate it. Unvalidated work-in-progress accumulates context drift and exponential rework costs. The orchestrator treats QA capacity as the system's governing constraint.

**Squads are small.** 4-5 agents per squad deliver production code effectively. Beyond that, coordination overhead grows non-linearly and merge conflicts dominate. The orchestrator decomposes work to minimize overlap between agents and establishes explicit ownership boundaries.

## 4. Code Agents

Code agents are the hands of the system. They receive a focused task, a constrained context, and a set of tests to make pass. They implement, nothing more.

**Less context produces better code.** Code agents receive the minimum viable information for their task — relevant interfaces, architectural patterns, test files — not the entire repository. Targeted context reduces token cost and produces more deterministic output. The CLI builds this working set.

**Each agent works in physical isolation.** Every code agent operates in a dedicated git worktree on a separate branch. No agent sees another agent's in-progress work. This eliminates file-locking contention and prevents agents from hallucinating against half-written code from sibling agents.

**Code agents do not write their own tests.** A separate agent authors tests from the Epic/Story specifications. The code agent receives those tests and writes implementation to make them pass. The code agent has no write permission to test files. This prevents "test hacking" — the documented pathology where an agent silently weakens assertions instead of writing correct implementation.

**TDD is the anti-hallucination mechanism.** The red-green-refactor loop forces the agent to observe a test failing against real dependencies before writing implementation. This anchors a probabilistic model to deterministic reality. The CLI rejects any commit where the execution trace does not show a failing test immediately before the implementation phase.

## 5. QA Agents

QA agents are the immune system. They validate all output with fresh eyes, no shared memory with the agent that produced it, and no incentive to agree.

**Independence is architectural, not aspirational.** The QA agent operates with a completely clean context window — no memory of the generation process, no access to the code agent's reasoning. This is what breaks the sycophancy cycle where an agent tends to agree with its own errors. No DAG node in Execute can be marked complete without passing through Validate.

**Validation is adversarial, not cooperative.** A QA agent that merely confirms "looks good" is worthless. Effective validation requires agents with orthogonal expertise — one that checks functional correctness, another that actively tries to break the system with malicious inputs and edge cases. Consensus without challenge amplifies errors into accepted facts.

**The Trust Layer is the final arbiter, not the QA agent.** The recursive problem of "who watches the watchmen" is not solved by adding more watchers. It is solved by deterministic reality. Compilation, test execution, and contract assertions do not have opinions. The QA agent produces tests and evidence; the CLI verifies them. If the output violates the contract, it is rejected — no debate.

**Different models catch different blind spots.** When the code agent and QA agent use the same LLM architecture, shared training biases cascade through the review chain undetected. Using heterogeneous models across roles prevents this.

## 6. Memory

Memory is the accumulated intelligence of the system. It persists across sessions, grows through curation, and is governed by the same contracts that govern everything else.

**Memory has three layers, not one.** Working memory is the minimal snapshot assembled for the current task — ephemeral, built on demand by the CLI. Episodic memory is the immutable log of everything that happened — events tied to runs, phases, and agents, never modified or deleted. Semantic memory is validated knowledge — patterns, conventions, constraints — versioned and promoted through explicit review. Each layer has different rules. Mixing them corrupts all three.

**Nothing enters memory without a contract.** The CLI validates every write. Schema is mandatory. Provenance is mandatory. Contradictions are stored as versioned proposals with justifications, not forced into premature consensus. Only QA-validated knowledge is promoted to operational truth.

**Wisdom is curated, not accumulated.** Raw accumulation produces noise. The system scores episodes by impact and recurrence, consolidates them into reflections, validates the reflections, and links them relationally. Without this curation pipeline, the knowledge base grows in volume but not in utility.

**Writes are serialized, reads are parallel.** Many agents read memory concurrently. All writes go through the CLI as short, serialized transactions. This is not a limitation — it is the governance model. The CLI is the single gateway that prevents agents from writing incorrect or contradictory knowledge.

## 7. What This Document Is Not

This document defines beliefs, not implementation. It does not specify CLI commands, database schemas, DAG formats, or agent prompt templates. Those belong in the design document.

This document also does not replace the existing philosophy documents. The core principles (docs/philosophy.md) and the CLI Trust Layer philosophy (docs/cli/philosophy.md) remain authoritative. This document extends them into the domain of multi-agent coordination.
