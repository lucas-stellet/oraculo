# Agent Philosophy (DRAFT)

> Consolidated synthesis of 6 parallel research threads on multi-agent orchestration, execution, communication, quality, memory, and coding agents. Based on 7 research documents, ~99 patterns identified, ~38 gaps catalogued.

---

## 1. Core Axioms

These axioms emerged independently across multiple research threads. They are non-negotiable design constraints.

### 1.1 Orchestrate, Never Execute

The main agent (Oraculo) operates exclusively as **Planner**. It ingests project context, assembles teams, generates the DAG, monitors progress, and coordinates validation. It never writes code, runs tests, or touches the filesystem directly.

**Why**: Removing tactical execution from the orchestrator reduces hallucinations and strategic drift by preserving context window integrity. When syntax errors, compilation logs, and debug traces enter the planner's context, strategic reasoning degrades measurably.

**Evidence**: OpenCode (plan/build separation), Cline (Plan/Act modes), MetaGPT (SOP-driven role separation), AutoGen (thinkers vs. executors). The Planner-Executor pattern is the most validated architecture across all surveyed frameworks.

### 1.2 Quality Over Speed (Drum-Buffer-Rope)

The QA Agent is the **Drum** -- the constraint that dictates the maximum tempo of the entire system. Code generation must be mathematically subordinated to QA throughput.

**Why**: When code production outpaces verification capacity, unvalidated Work-In-Progress accumulates, causing "context drift" -- rejected code returns for rework but dependencies and interfaces have already changed. Unbounded parallelism without respecting the verification bottleneck leads to catastrophic context collision and exponential rework costs.

**Mechanism**:
- **Drum**: QA Agent throughput defines maximum system speed
- **Buffer**: Limited queue of completed tasks awaiting QA (prevents QA starvation)
- **Rope**: Signal from QA back to orchestrator to throttle task dispatch

**Evidence**: Theory of Constraints (Goldratt) adapted for agent swarms. TAIS framework. Amdahl's Law for Agents (arXiv:2503.15703) -- global speed is dictated by the longest sequential path.

### 1.3 Ask Before Doing (Constraint Satisfaction)

No execution without deep understanding. The ACONIC framework formalizes this: model requirements as constraint satisfaction problems (3-SAT), use tree decomposition to partition the constraint graph. If the graph cannot be decomposed into low-treewidth Stories, the problem is not yet understood -- more Socratic exploration is required.

**Why**: Heuristic decomposition fails at scale. Formal constraint graphs allow the orchestrator to algorithmically minimize cognitive load on execution agents.

**Evidence**: ACONIC (arXiv:2510.07772), Plan-over-Graph (arXiv:2502.14563).

### 1.4 The Trust Layer is Deterministic Reality

The CLI Trust Layer is the ultimate arbiter. It does not "believe" -- it **verifies**. Compilation, test execution, and contract assertions are the deterministic boundaries that contain probabilistic AI behavior.

**Why**: The recursive problem of "who watches the watchmen?" is solved by transitioning from purely probabilistic validation (LLM judging LLM) to deterministic validation (contracts that break builds). A contract is satisfied if both input and output are successfully validated against specifications -- no opinion required.

**Evidence**: Design by Contract (Bertrand Meyer), SymbolicAI `@contract` decorators, TDFlow empirical results (94.3% with ground-truth tests vs. 69.8% with self-generated tests).

---

## 2. Orchestration Architecture

### 2.1 DAG as First-Class Data Structure

The execution plan is not a text list -- it is a **computable graph** of nodes (subtasks) and edges (dependencies). The orchestrator continuously evaluates graph state in O(|V|+|E|), dispatching parallel executions for all nodes with in-degree zero.

**Key properties**:
- Programmatic cycle detection (no infinite loops)
- Maximum parallelism width identified at any moment
- Topological ordering enforces dependency respect
- Conditional branch nodes allow runtime mutation without full DAG regeneration

**Dynamic mutation**: Static DAGs inevitably fail because the optimal execution path depends on dynamic observations (deprecated API, unexpected legacy schema). The Routine Framework embeds formal branching logic directly into the plan representation. When an agent hits a conditional node, the orchestrator evaluates new state, prunes the invalid branch, and dynamically schedules newly-discovered nodes.

### 2.2 Dynamic Topology Routing

The Performance Convergence Scaling Law (2026) reveals that when foundation models reach reasoning parity, the dominant optimization variable shifts from model selection to **structural composition** -- how agents are coordinated matters more than individual LLM intelligence.

**Four canonical topologies**:

| Topology | When to Use | Oraculo Phase |
|----------|-------------|---------------|
| Parallel | High width, low depth, decoupled features | Execute (independent stories) |
| Sequential | Low width, high depth, strict dependencies | Validate (review chain) |
| Hierarchical | High coupling, manager-worker delegation | Discover (Socratic exploration) |
| Hybrid | Mixed width/depth across phases | Full lifecycle |

The orchestrator analyzes DAG mathematical properties (width, depth, coupling) and routes execution dynamically. Static workflow architectures artificially limit LLM capabilities.

**Evidence**: AdaptOrch Topology Router (arXiv:2602.16873).

### 2.3 Concurrent Re-planning (PIE)

The orchestrator re-plans later DAG stages **concurrently** while current stages are still executing. When an agent encounters an unexpected constraint, the orchestrator absorbs dynamic feedback and re-optimizes the downstream DAG in background -- without pausing current execution.

**Evidence**: PIE Framework (AAAI 2025).

### 2.4 Squad Size Limits

Empirical evidence indicates that specialized squads of **4-5 agents** deliver production code effectively. Larger swarms generate excessive merge conflicts. Coordination overhead grows non-linearly with agent count.

---

## 3. Task Decomposition and Parallel Execution

### 3.1 Plan-over-Graph

Instead of generating sequential text plans, the LLM constructs a formal graph of nodes and edges directly. Dynamic programming computes optimal parallel execution paths. This exposes maximum parallelization opportunities naturally.

### 3.2 Constraint-Induced Decomposition

Transform Epics into constraint graphs. Minimize treewidth to mathematically guarantee resulting Stories don't contain conflicting requirements. High treewidth = problem not yet understood = more Socratic exploration needed.

### 3.3 Critical Path Analysis

Amdahl's Law for Agents: global speed is throttled by the longest sequential dependency chain. The orchestrator must:
- Identify the critical path in the DAG
- Allocate highest-performance models to critical path agents
- Use smaller, faster models for non-critical parallel tasks

### 3.4 Failure Recovery

**Circuit Breakers**: Monitor failure rates and latency at each agent handoff. If a Code Agent consistently fails CLI validation, the circuit breaker trips -- blocking further processing in that DAG branch. Deploy a Debugger Agent to evaluate the snapshot.

**Coordinated Checkpointing**: When one parallel branch fails, ALL dependent agents within the sub-graph revert to a synchronized temporal state. Uncoordinated rollbacks in parallel environments create timeline divergence -- agents collaborating on fundamentally incompatible state assumptions.

**Graceful Degradation**: If a secondary tool fails, the agent should document the limitation rather than crash the node.

---

## 4. Communication and Context

### 4.1 The Blackboard Pattern (SQLite + CLI)

The SQLite database + CLI functions as a **deterministic blackboard**. Agents publish proposals and evidence; the leader/QA promotes what becomes "canonical knowledge." The DAG serves as the controller, defining which writes are permitted in each phase.

**Governance is as important as storage**: The challenge isn't sharing data -- it's governing **who can write what and when**, with explicit policies.

**Roles on the blackboard** (from bMAS/LbMAS research):
- **Planner**: Decomposes and schedules
- **Critic**: Reviews quality
- **Conflict-Resolver**: Mediates contradictions between agents
- **Cleaner**: Maintains integrity and relevance of shared content
- **Decider**: Determines when sufficient consensus is reached

### 4.2 Selective Context Injection (Minimal Viable Information)

Context must be treated as **modular, lazy-loaded dependencies**, not monolithic prompt preambles. Less but highly-targeted information produces significantly more deterministic output.

**Practical targets**:
- Concepts < 100 lines
- Guides < 150 lines
- ~80% token reduction achievable (from 8,000 to ~750 tokens per request)

**Mechanisms**:
- CLI provides deterministic "build working set" command
- FTS5/BM25 retrieves only relevant context per DAG node
- Topological repository mapping (Aider-style): classes, method signatures, dependencies -- without loading implementation bodies
- Graph ranking surfaces only the most-referenced nodes

### 4.3 Handoff Protocol

Handoffs are **first-class inspectable transitions**, not implicit "ask another agent" actions hidden in prompts.

**Information preserved during handoff**:
1. Task description (structured, not conversational)
2. Current state (checkpoint reference)
3. Expected outputs (schema-validated)
4. Constraints and preconditions
5. Prior results (references, not full content)

**Standard Operating Procedures (SOPs)** codify handoff formats as machine-readable templates. A DAG node can only be marked complete when the agent produces a structured output artifact that the QA agent can deterministically parse and validate.

### 4.4 Context Divergence Resolution

When parallel agents develop contradictory understanding:

1. **Primary defense**: SQLite + CLI as single source of truth -- parallel agents always validate against canonical blackboard state
2. **Git Worktrees**: Physical isolation prevents contamination during parallel execution
3. **QA Agent as arbiter**: Resolves divergences with empirical evidence
4. **Truth Maintenance**: Store conflicting conclusions as versioned proposals with justifications -- don't force early consensus. Promote to "operational truth" only after QA validation

---

## 5. Quality Assurance and Validation

### 5.1 Orthogonal Verification

The QA Agent operates with a **completely clean context window**, with no memory of the original generation process. This breaks the destructive cycle of sycophancy where an agent tends to agree with its own errors.

**Architectural requirement**: No edge in the DAG connects Execute directly to Deliver without passing through a Validate node.

### 5.2 TDD as Anti-Hallucination Mechanism

TDD is not merely a testing strategy -- it is a **critical mechanism for anchoring a probabilistic model to deterministic reality**.

**The Red-Green-Refactor loop forced by the CLI**:
1. Agent writes test from Socratic requirements (Red)
2. Agent observes test failing against real dependencies (proof of real validation)
3. Agent writes minimal implementation to satisfy the contract (Green)
4. Agent refactors while keeping tests green (Refactor)

**The CLI must reject any commit if the execution trace in SQLite does not demonstrably show a failing test state immediately before the implementation phase.**

### 5.3 Agent Separation: Test-Author vs. Implementer

The TDFlow framework demonstrated empirically:
- With human ground-truth tests: **94.3%** success rate
- With self-generated tests: **69.8%**
- Without tests: **30-45%**

**"Test hacking"** is a documented pathology: when the agent can't make tests pass, it silently weakens test assertions instead. The architectural solution:

1. A specialized agent writes tests from Epic/Story specifications
2. A separate, isolated agent writes implementation
3. **The implementation agent has NO write permission to test files**
4. The Trust Layer (CLI) enforces this permission boundary

### 5.4 Mutation Testing (Validating the Validator)

An LLM can write `assert True` -- a test that passes but covers nothing. Mutation testing introduces deliberate bugs into code and verifies the QA Agent's tests detect them.

- If tests still pass despite mutation: tests are fragile, QA output is rejected
- Target: **>80% mutation detection rate**
- The CLI can automate this as part of the Trust Layer

### 5.5 Adversarial Validation

Cooperative QA agents are insufficient for robust security. Two types of validation agents:

| Type | Role | What It Catches |
|------|------|----------------|
| **Critic Agent** | Evaluates non-functional criteria | Algorithmic efficiency, pattern adherence, maintainability, security |
| **Adversarial Agent** | Actively tries to break the system | Fuzzing APIs, malicious inputs, prompt injection, silent failures, edge cases |

### 5.6 Heterogeneous Model Cross-Validation

Use different LLM architectures for different roles in the validation cycle. If the Code Agent uses Model X and the QA Agent uses Model Y, shared architectural biases don't cascade through the review chain.

**Evidence**: Recursive Knowledge Synthesis (arXiv:2601.08839) -- tri-agent with heterogeneous models.

### 5.7 Quality Metrics

| Metric | What It Measures | Target |
|--------|-----------------|--------|
| Pass@1 / Pass@k | Functional correctness | >85% |
| Mutation Score | Test robustness | >80% |
| Cyclomatic Complexity | Maintainability | <10 paths/method |
| Execution Efficacy | Fraction of CLI calls returning desired state | High |
| Convergence Score | Efficiency of steps to solution | Low is better |
| Tool Error Rate | Error amplification through DAG | Low |

**Process quality matters**: An agent that needs 50 tool calls to resolve a syntax error is fundamentally less reliable than one that resolves in 3, even if both pass the test suite.

### 5.8 Human-on-the-Loop (Rule 40-60)

AI handles autonomously 40-60% of review burden (syntax, linting, formatting, basic security checks). Humans reserve cognitive load for:
- Complex business logic
- Architectural drift
- AI-specific logical failures (sorts that silently fail on edge cases, inconsistent conventions, hallucinated error handling)

---

## 6. Shared Memory and Knowledge

### 6.1 Three-Tier Memory Architecture

| Tier | What | Storage | Mutability |
|------|------|---------|------------|
| **Working Memory** | Minimal snapshot for current task: problem, goals, constraints, recent decisions, top-k retrieved items | Built by CLI "build working set" command | Ephemeral per DAG node |
| **Episodic Memory** | Immutable events tied to `run_id/thread_id` and phase (Discover/Plan/Execute/Validate/Deliver) | SQLite append-only | Immutable |
| **Semantic Memory** | Validated patterns, conventions, constraints with category/domain, versioning, status | SQLite with FTS5 | Versioned (proposed/validated/deprecated) |

### 6.2 Event Sourcing as Foundation

Store state changes as an **immutable sequence of events**. Never lose the past. Derive "current state" as a validated projection.

**Practical implications**:
- Every "wisdom promotion" must point to a set of source events (provenance)
- Full replay, debug, and audit capability
- Separate immutable events (what happened) from projections (consolidated patterns, current constraints)

### 6.3 Concurrency: N Readers + 1 Writer

SQLite WAL mode supports many concurrent readers but only one writer at a time. The healthy parallelism pattern:

- **Many agents reading in parallel** (context retrieval, validation checks)
- **Writes serialized through the CLI** (short transactions, timeouts/retries for SQLITE_BUSY)
- The CLI concentrates commits, enforces checkpoints
- Write becomes a shared resource ("serialization node" in the DAG) without killing overall parallelism

### 6.4 Explicit Versioning Against Semantic Corruption

Canonical knowledge items must be treated as **versioned**:

```
fields: valid_from, valid_to, supersedes_id, derived_from_episode_id, status
status: proposed | validated | deprecated
```

Versioning transforms conflicts into diagnosable data instead of silent losses. When agents disagree, the conflict becomes version coexistence -- not evidence destruction.

### 6.5 Corruption Prevention: Contracts at the Border

Design by Contract applied to memory means every CLI memory operation has:
- **Preconditions**: Mandatory schema, consistency checks
- **Postconditions**: Cross-validation with indexes/FTS5
- **Invariants**: Promotion rules (only QA promotes to operational truth)

Don't force early consensus. Store justifications and treat contradictions as data until final validation.

### 6.6 Wisdom Pipeline

Accumulated wisdom is a **curation pipeline**, not raw accumulation:

```
Episode -> Score (impact, risk, recurrence) -> Reflection/Consolidation -> Validated Promotion -> Relational Linking
```

Without this pipeline, the knowledge base grows in volume but not utility.

### 6.7 Search Strategy

Start with **FTS5 (lexical, deterministic, auditable)** and evolve progressively:

| Layer | Technology | When |
|-------|-----------|------|
| Lexical search | FTS5/BM25 | Default, always available |
| Relational structure | Entity/relation tables in SQLite | When knowledge has dependencies, conflicts, substitutions |
| Semantic search | Embeddings (sqlite-vec) | When conceptual discovery beyond keywords is needed |
| Hybrid ranking | RRF (Reciprocal Rank Fusion) | Combining lexical + semantic results |

For canonical memory, lexical search with explainable ranking and integrity checks may be more reliable than depending on embeddings alone.

---

## 7. Coding Agents

### 7.1 Agent-Computer Interface (ACI)

Agents communicate with the environment through a **structured, deterministic, LLM-optimized interface** (the CLI Trust Layer). This is the technical manifestation of the ACI concept from SWE-agent.

**Principles**:
- Constrained output (e.g., 100-line file viewer limits)
- Built-in syntax checker forces correction before proceeding
- Formatted feedback concise enough for LLM consumption
- Typed commands with validated inputs/outputs

Limiting agent operational freedom through structured interfaces improves success rates more than scaling model parameters.

### 7.2 Git Worktrees for Parallel Isolation

Each DAG node gets a **dedicated, isolated worktree** -- a physical copy of the codebase on a separate branch.

**Why not Jujutsu (jj)?**: Its working-copy-is-always-a-commit model causes concurrent sub-agents to absorb each other's changes into massive corrupted commits. Agents require the explicit file-selection boundaries (like `git add`) that humans find tedious.

**Key insight**: Systems designed to eliminate friction for human developers (automatic staging, implicit commit absorption) often create catastrophic friction for autonomous agents. Agents require explicit, deterministic boundaries.

### 7.3 Harness Engineering

The orchestrator implements "harness engineering" with explicit protocol boundaries between agent decision-making and tool execution. A bidirectional protocol decouples the harness from the agent loop.

### 7.4 Context Engineering > Prompt Engineering

Project brain files (CLAUDE.md, SKILL.md, reference templates) plus persistent SQLite memory with FTS5 indexing constitute the "project brain" that survives individual interactions. Configuration-driven context outperforms inference-only approaches for maintaining project consistency.

---

## 8. Resilience and Recovery

### 8.1 Durable Execution (Temporal Pattern)

Orchestration logic must be **strictly deterministic**. Non-deterministic behavior (LLM calls, tool executions) is isolated in "Activities." The Event History is complete, immutable, and append-only.

If an agent crashes, the orchestrator reconstructs the exact pre-crash state via Event History replay and resumes execution without path divergence.

**The orchestrator's primary function is state preservation, not intelligence.**

### 8.2 Coordinated Checkpointing

Each Oraculo phase (Discover/Plan/Execute/Validate/Deliver) materializes as a **reproducible checkpoint** in SQLite. With FTS5, checkpoints are searchable (by decision, risk, evidence).

On failure, rollback is partial -- to the last valid checkpoint of the sub-graph -- without destroying work from unaffected parallel branches.

### 8.3 Circuit Breakers

Monitor failure rates and processing latency at each agent handoff. When a Code Agent consistently fails Trust Layer validation:

1. Circuit breaker trips -- blocks further processing in that DAG branch
2. Corrupted data is prevented from flowing downstream
3. State is logged to SQLite
4. Debugger Agent is deployed to evaluate the snapshot
5. Recovery sequence begins with refined prompt

---

## 9. Identified Gaps (Cross-Cutting)

Research areas that require further investigation before finalizing this philosophy:

### 9.1 Economics
- No systematic analysis of operational costs (cost per DAG executed, cost per agent-hour)
- Anthropic reports 15x token consumption for parallel sub-agents (with 90.2% performance gain), but ROI analysis is missing
- Break-even points vs. human developers not quantified

### 9.2 Orchestrator Failure
- All research assumes the central orchestrator is reliable
- No coverage of what happens when Oraculo itself enters inconsistent state, exhausts context, or hallucinates during DAG planning
- Fallback strategies for orchestrator failure are undocumented

### 9.3 Knowledge Decay
- No concrete mechanisms for detecting when canonical knowledge becomes obsolete
- Need policies for "semantic TTL" or re-validation triggers
- When does a validated pattern become deprecated? (framework migration, large refactor)

### 9.4 Real-Time Divergence Detection
- Research describes how to RESOLVE divergences but not how to DETECT them in real-time during parallel execution
- No mechanism for continuous semantic coherence monitoring between parallel agents

### 9.5 Cross-Phase Context Propagation
- Well-covered within a phase, but how does accumulated context from Discover optimally propagate to Plan, and from Plan to Execute, especially over long time intervals?

### 9.6 Empirical Production Data
- Most evidence comes from academic benchmarks (SWE-bench) and conceptual frameworks
- Lack of detailed case studies from real production deployments with throughput, quality, cost, and team satisfaction metrics

### 9.7 Security
- Agent sandboxing and inter-agent security (preventing a compromised agent from affecting others) has minimal coverage
- Memory access control (which agents can read/write which domains) is not addressed

---

## 10. Summary: Design Decisions for Oraculo

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Orchestrator role | Planner only, never executes | Preserves context window, reduces hallucination |
| Execution model | DAG with dynamic topology routing | Maximizes parallelism while respecting dependencies |
| Parallel isolation | Git worktrees per DAG node | Physical isolation prevents cross-contamination |
| Throughput control | Drum-Buffer-Rope (QA as constraint) | Quality mathematically prioritized over speed |
| Communication substrate | SQLite blackboard + CLI gateway | Deterministic, auditable, governable |
| Context delivery | Minimal Viable Information via CLI | ~80% token reduction, more deterministic output |
| Handoff protocol | Structured JSON contracts validated by CLI | Machine-readable, inspectable, testable |
| QA architecture | Independent agent, clean context, no shared memory with generator | Breaks sycophancy cycle |
| TDD enforcement | Separate test-author and implementer agents | Prevents test hacking (94.3% vs 69.8%) |
| Trust Layer | Design by Contract via CLI (preconditions, postconditions, invariants) | Deterministic answer to "who watches the watchmen" |
| Validation depth | Critic + Adversarial + Mutation Testing | Cooperative QA insufficient for robust security |
| Model diversity | Heterogeneous models across roles | Prevents shared training biases from cascading |
| Memory architecture | 3-tier (working/episodic/semantic) in SQLite | Event sourcing + versioned canonical knowledge |
| Search | FTS5 lexical first, evolve to hybrid | Deterministic, auditable, integrity-checkable |
| Failure recovery | Durable Execution + Coordinated Checkpointing + Circuit Breakers | Partial rollback without destroying parallel work |
| Squad size | 4-5 agents per squad | Non-linear coordination overhead beyond this |
| Re-planning | Concurrent (PIE) while executing | No execution pause for re-optimization |
| Human role | On-the-Loop (40-60 rule) | Audit reasoning and architecture, not syntax |
