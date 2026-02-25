# Oraculo Agents Design

> **Status:** DRAFT
> **Audience:** Implementers and future contributors
> **Final location:** `docs/agents/design.md`

---

## Table of Contents

1. [Architectural Overview](#1-architectural-overview)
2. [The DAG Engine](#2-the-dag-engine)
3. [Agent Isolation and Execution Environment](#3-agent-isolation-and-execution-environment)
4. [TDD Pipeline](#4-tdd-pipeline)
5. [QA Architecture](#5-qa-architecture)
6. [Memory System](#6-memory-system)
7. [Communication, Handoff, and Failure Recovery](#7-communication-handoff-and-failure-recovery)
8. [Cross-Cutting Concerns](#8-cross-cutting-concerns)
9. [MVP Definition](#9-mvp-definition)
10. [Future Work](#10-future-work)

---

## 1. Architectural Overview

### 1.1 Operating Modes

Oraculo supports two operating modes. Every task goes through the appropriate mode -- the golden rule is that no task bypasses its mode's phases.

**Full mode** (epics): 5 phases -- Discover, Plan, Execute, Validate, Deliver (future).

**Reduced mode** (stories without an epic): 3 phases -- Plan, Execute, Validate. Discover is skipped because context comes directly from the user. Deliver is skipped because integration into mainline is future work. The minimum path is Plan, Execute, Validate.

| Phase | Primary Actor | Input | Output | Key Constraint | Mode |
|-------|--------------|-------|--------|---------------|------|
| **Discover** | Orchestrator + Human | Raw idea or problem statement | Validated requirements document | Socratic exploration; no action without understanding | Full only |
| **Plan** | Orchestrator | Requirements document (full) or user-provided context (reduced) | DAG of tasks with dependencies, squads, and resource budgets | TOC-informed decomposition; critical path identified | Both |
| **Execute** | Specialized agents (code, test-author, research) | Individual DAG node specs + MVI context bundles | Code changes in isolated worktrees | TDD discipline; agents never see full repo | Both |
| **Validate** | QA agents (4 personas) + CLI | Git diffs + specs + test logs | Accept/reject verdict per node | Heterogeneous models; no partial approval; CLI is final arbiter | Both |
| **Deliver** (future) | CLI + Merge Agent (if conflicts) | Validated branches | Merged mainline + markdown summary | Serial integration; pre-merge validation | Full only (future) |

The orchestrator spans all active phases but **never executes** -- it delegates every concrete action to a specialized agent or to deterministic CLI logic. This is the "orchestrate, never execute" principle made concrete.

### 1.2 End-to-End Lifecycle of a Task

A task's lifecycle follows a deterministic path through the system:

**1. Birth (DAG Node Creation)**
During the Plan phase, the orchestrator emits a JSON plan with `nodes[]` and `edges[]`. The CLI validates this structure (acyclicity, edge-target existence, input/output schema compatibility) and persists it as normalized rows in SQLite tables (`dag_nodes`, `dag_edges`). Each node receives an opaque integer primary key and a human-readable coordinate label for LLM communication.

**2. Dispatch (Frontier Computation)**
The CLI autonomously runs a dispatch loop. It queries `node_runs` for nodes in PENDING state whose upstream dependencies are all terminal (COMPLETED or PRUNED). It scores candidates by `(critical_path_length, fan_out)`, filters by squad capacity and per-stage WIP limits, and transitions selected runs to RUNNING in a single SQLite transaction. No LLM is involved in dispatch -- the CLI handles it deterministically.

**3. Execution (Isolated Worktree)**
The CLI creates an ephemeral Git worktree at `.oraculo/worktrees/<dag_run_id>/<node_id>/` from a pinned `base_sha`. It builds the Minimal Viable Information (MVI) context bundle -- the agent's working set -- using a radius-based dependency graph walk augmented by in-degree ranking, fitted to a token budget. An `AGENT_<node_id>.md` contract file is generated, defining the agent's goal, editable files, read-only files, forbidden paths, and anti-goals. The agent is spawned as a Claude Code sub-agent with PreToolUse hooks enforcing file-level ACLs.

**4. Validation (QA Pipeline)**
When the agent completes, the CLI initiates a four-gate validation sub-DAG:
- **Gate 0:** TDD trace verification (queries SQLite event store for Red-Green proof)
- **Gate 1:** Parallel QA agent provisioning (Functional Reviewer, Adversarial Auditor, Style Checker, each in its own worktree with bounded context payload)
- **Gate 2:** Deterministic test execution by the CLI (all QA-generated test suites)
- **Gate 3:** Mutation score threshold enforcement

The final accept/reject decision is a deterministic SQL query against the append-only validation log. If rejected, the code agent receives targeted feedback and the full validation pipeline re-runs from scratch. A circuit breaker limits this to 3 cycles before escalation.

**5. Integration (Merge) -- Future Work**
The Deliver phase (merging validated agent work back to mainline) is deferred to future work. See [Section 10](#10-future-work) for details on what it will include.

### 1.3 Data Flow Diagram (Textual Description)

The system's data flow can be visualized as three parallel streams converging at decision points:

```
HUMAN INPUT                 CLI (TRUST LAYER)              AGENTS
    |                            |                            |
    | idea/problem               |                            |
    +--------------------------->|                            |
    |                            | spawn orchestrator         |
    |                            +--------------------------->|
    |      Socratic questions    |<------ questions -----------|
    |<---------------------------+                            |
    | answers                    |                            |
    +--------------------------->| validate requirements      |
    |                            |                            |
    |                            | emit DAG (JSON)            |
    |                            |<--- plan proposal ---------|
    |                            |                            |
    |                            | validate DAG (acyclicity,  |
    |                            |   schemas, edges)          |
    |                            | persist to SQLite          |
    |                            |                            |
    |                            | DISPATCH LOOP              |
    |                            | (autonomous, no LLM)       |
    |                            |   query frontier           |
    |                            |   score by critical path   |
    |                            |   filter by WIP limits     |
    |                            |   create worktrees         |
    |                            |   build MVI bundles        |
    |                            |   spawn agents ----------->|
    |                            |                    execute  |
    |                            |                    (TDD)    |
    |                            |<--- result branch ---------|
    |                            |                            |
    |                            | VALIDATION SUB-DAG         |
    |                            |   Gate 0: TDD trace check  |
    |                            |   spawn QA agents -------->|
    |                            |                    review   |
    |                            |<--- test suites -----------|
    |                            |   Gate 2: execute tests    |
    |                            |   Gate 3: mutation score   |
    |                            |   SQL verdict query        |
    |                            |                            |
    |                            | (Deliver phase: future)    |
    |                            |                            |
```

The key architectural invariant visible in this flow: **the CLI sits between every agent and every persistent state change**. Agents never write to SQLite directly. Agents never merge to main. Agents never dispatch other agents. The CLI is the sole authority.

### 1.4 Key Architectural Invariants

These invariants hold across all phases and all subsystems:

1. **CLI as Trust Layer.** Every write to SQLite, every accept/reject verdict, and every resource limit enforcement passes through the CLI. When the Deliver phase is implemented (future), every merge to mainline will also pass through the CLI. The CLI is deterministic -- its decisions are based on data and exit codes, never on LLM reasoning.

2. **Append-only state.** All state changes are appended as immutable events. Nodes are never deleted -- they are marked PRUNED. Artifacts are never overwritten -- they are marked deprecated with a compensating event. This enables full auditability and safe rollback.

3. **Isolation by default.** Every agent operates in its own Git worktree. No agent has access to another agent's in-progress work. Context is explicitly constructed and injected by the CLI, never discovered by the agent.

4. **LLM proposes, CLI disposes.** The LLM orchestrator proposes plans and mutations. The CLI validates structure (acyclicity, schemas, permissions) and commits or rejects. The LLM never directly mutates persisted state.

5. **Every task goes through the appropriate mode.** Epics use the full mode (Discover, Plan, Execute, Validate). Stories without an epic use the reduced mode (Plan, Execute, Validate). The Deliver phase is future work for both modes. The depth of each phase scales with task complexity, but within a mode the sequence is invariant.

6. **Heterogeneous models for orthogonal validation.** QA agents must use different model families from code agents. This prevents shared training biases from creating false consensus.

7. **Parallel execution.** Agents execute in parallel (limited by squad capacity and WIP limits). When the Deliver phase is implemented (future), merges to mainline will be serial and always validated.

---

## 2. The DAG Engine

The DAG engine is the computational core of Oraculo's orchestration. It models tasks as a directed acyclic graph, manages dispatch of work to agents, controls throughput via Theory of Constraints principles, and persists all state via event sourcing in SQLite.

### 2.1 Graph Representation (Dual Model)

The DAG exists in two synchronized representations:

| Concern | LLM-Facing | CLI / Storage |
|---------|-----------|---------------|
| Format | JSON with explicit `nodes[]` and `edges[]` arrays | Normalized relational tables in SQLite |
| Purpose | Semantic reasoning, plan emission, mutation proposals | Deterministic validation, fast querying, dispatch |
| Authority | Proposes | Enforces |

The LLM never directly mutates the persisted graph. It emits a JSON structure; the CLI validates and commits. This separation is fundamental to the Trust Layer architecture.

**Rationale:** LLMs reason well over explicit, self-contained JSON structures. Relational tables provide referential integrity, fast frontier queries, and transactional atomicity. The dual model gives each concern its optimal representation.

### 2.2 Node and Edge Schemas

#### Node Schema

Each DAG node carries the following fields:

```sql
CREATE TABLE dag_nodes (
    node_id           INTEGER PRIMARY KEY,
    journey_id        TEXT    NOT NULL,
    plan_version      INTEGER NOT NULL,
    human_label       TEXT    NOT NULL,     -- e.g., "1.2.b" for LLM communication
    type              TEXT    NOT NULL      -- code, qa, research, documentation, merge, test-author, refactor
                      CHECK (type IN ('code','qa','research','documentation',
                                      'merge','test-author','refactor')),
    squad_id          TEXT,
    phase             TEXT    NOT NULL      -- discover, plan, execute, validate (deliver: future)
                      CHECK (phase IN ('discover','plan','execute','validate','deliver')),
    isolation         TEXT    NOT NULL DEFAULT 'worktree'
                      CHECK (isolation IN ('worktree','inline')),
    input_schema_json TEXT,                -- JSON Schema for expected inputs
    output_schema_json TEXT,               -- JSON Schema for expected outputs
    is_terminal       INTEGER NOT NULL DEFAULT 0,
    created_at        TEXT    NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);
```

**Design decisions:**

- **Opaque integer IDs as primary keys.** Human-readable coordinate labels (`human_label`) are stored as a display field for LLM communication. This preserves referential integrity while enabling intuitive references like "1.2.b". Using semantic strings as PKs would couple the DB schema to the LLM's naming convention and make foreign-key management fragile.

- **Node/Run separation.** The `dag_nodes` table contains static definitions. Runtime execution state lives in a separate `node_runs` table (see below). The dispatch engine operates on run state, never on node definitions. This enables retries, re-runs, and clean auditability.

- **Isolation field.** Each node declares how the CLI should spawn its agent. The default is `worktree` (ephemeral Git worktree). `inline` is reserved for lightweight tasks (summarization, research) that do not modify the codebase.

- **Input/output schema validation.** Edges are rejected at commit time if the source node's output schema does not satisfy the target node's input requirements. This catches hallucinated dependencies before execution begins.

#### Edge Schema

```sql
CREATE TABLE dag_edges (
    edge_id       INTEGER PRIMARY KEY,
    journey_id    TEXT    NOT NULL,
    plan_version  INTEGER NOT NULL,
    from_node_id  INTEGER NOT NULL,
    to_node_id    INTEGER NOT NULL,
    condition     TEXT,              -- optional conditional expression
    active        INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (from_node_id) REFERENCES dag_nodes(node_id),
    FOREIGN KEY (to_node_id) REFERENCES dag_nodes(node_id)
);
```

#### Node Run Schema

```sql
CREATE TABLE node_runs (
    node_run_id   INTEGER PRIMARY KEY,
    node_id       INTEGER NOT NULL,
    journey_id    TEXT    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','running','completed',
                                    'failed','pruned','cancelled')),
    result_ref    TEXT,              -- pointer to artifact or branch
    error_kind    TEXT,
    retry_count   INTEGER NOT NULL DEFAULT 0,
    started_at    TEXT,
    finished_at   TEXT,
    FOREIGN KEY (node_id) REFERENCES dag_nodes(node_id)
);
```

**State machine for node runs:**

```
PENDING --> RUNNING --> COMPLETED
                   \-> FAILED
                   \-> CANCELLED

PENDING --> PRUNED  (when upstream branch is pruned)
```

A FAILED node does **not** decrement the in-degree of its successors, freezing that branch until the orchestrator intervenes (retry, re-plan, or escalate).

#### Journey Schema

```sql
CREATE TABLE journeys (
    journey_id           TEXT PRIMARY KEY,
    root_goal            TEXT NOT NULL,
    plan_version_current INTEGER NOT NULL DEFAULT 1,
    status               TEXT NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','completed','failed','paused')),
    created_at           TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### 2.3 Dispatch Algorithm

The dispatch loop is entirely CLI-owned. No LLM is involved. It runs autonomously on a polling interval and executes the following algorithm:

```
DISPATCH_LOOP:
  1. BEGIN TRANSACTION

  2. frontier = SELECT nr.node_run_id, nr.node_id, dn.squad_id, dn.type, dn.phase
                FROM node_runs nr
                JOIN dag_nodes dn ON nr.node_id = dn.node_id
                WHERE nr.status = 'pending'
                AND NOT EXISTS (
                    SELECT 1 FROM dag_edges de
                    JOIN node_runs upstream ON upstream.node_id = de.from_node_id
                    WHERE de.to_node_id = nr.node_id
                    AND de.active = 1
                    AND upstream.status NOT IN ('completed', 'pruned')
                )

  3. -- Filter by squad capacity
     FOR EACH candidate IN frontier:
       running_in_squad = COUNT(*) FROM node_runs
                          WHERE squad_id = candidate.squad_id
                          AND status = 'running'
       IF running_in_squad >= squads.max_running:
         REMOVE candidate FROM frontier

  4. -- Filter by per-stage WIP limits
     FOR EACH candidate IN frontier:
       running_in_stage = COUNT(*) FROM node_runs nr
                          JOIN dag_nodes dn ON nr.node_id = dn.node_id
                          WHERE dn.phase = candidate.phase
                          AND nr.status = 'running'
       IF running_in_stage >= stage_limits.wip_limit:
         REMOVE candidate FROM frontier

  5. -- Apply DBR throttle (see section 2.5)
     IF qa_buffer_full:
       REMOVE candidates WHERE type = 'code' FROM frontier

  6. -- Score remaining candidates by priority
     FOR EACH candidate IN frontier:
       candidate.priority = (critical_path_length(candidate), fan_out(candidate))
     SORT frontier BY priority DESC

  7. -- Dispatch top N candidates
     FOR EACH candidate IN frontier[0..max_dispatch]:
       UPDATE node_runs SET status = 'running', started_at = now()
         WHERE node_run_id = candidate.node_run_id
       APPEND event('node_dispatched', candidate)

  8. COMMIT TRANSACTION

  9. -- Outside transaction: spawn agents
     FOR EACH dispatched candidate:
       create_worktree(candidate)
       build_mvi_bundle(candidate)
       spawn_agent(candidate)
```

**Priority scoring heuristic:** When ready nodes exceed squad capacity, the dispatch loop prioritizes nodes on the critical path. `critical_path_length` is the length of the longest path from the candidate to a terminal node. `fan_out` is the number of immediate successors. This ensures the CLI dispatches nodes that unblock the longest dependency chains first.

**Rationale:** A CLI-owned dispatch loop avoids the latency and non-determinism of invoking the LLM for scheduling decisions. The LLM is consulted only for planning and re-planning -- high-level semantic decisions -- not for mechanical frontier computation.

### 2.4 Dynamic Mutation

The LLM orchestrator proposes graph mutations via a structured `MutateGraph` tool call. The CLI validates and applies mutations atomically.

#### MutateGraph Tool Schema

```json
{
  "tool": "MutateGraph",
  "params": {
    "journey_id": "string",
    "plan_version": "integer (current version being mutated)",
    "version_id": "integer (ETAG for optimistic concurrency control)",
    "prune_nodes": ["node_id_1", "node_id_2"],
    "insert_nodes": [
      {
        "human_label": "1.3.c",
        "type": "code",
        "squad_id": "backend",
        "phase": "execute",
        "isolation": "worktree",
        "input_schema_json": "...",
        "output_schema_json": "..."
      }
    ],
    "insert_edges": [
      { "from_label": "1.2.b", "to_label": "1.3.c" }
    ]
  }
}
```

#### Mutation Invariants

1. **Never mutate RUNNING or COMPLETED nodes.** Pruning can only target PENDING nodes.
2. **Pruned nodes are logically marked, never physically deleted.** This preserves auditability.
3. **All mutations are wrapped in a single atomic SQLite transaction.** No partial graph rewrites.
4. **Acyclicity is validated after mutation.** The CLI runs Kahn's algorithm on the proposed new graph state.

#### Mutation Algorithm

```
MUTATE_GRAPH(mutation):
  1. Validate ETAG: if mutation.version_id != latest_event_sequence_number,
     REJECT with current state (stale mutation)

  2. BEGIN TRANSACTION

  3. FOR EACH node_id IN mutation.prune_nodes:
     a. Verify status is PENDING; reject if RUNNING or COMPLETED
     b. Compute transitive closure of downstream nodes that depend
        exclusively on the pruned path
     c. Mark all nodes in closure as PRUNED
     d. Update merge-point dependencies: if a merge node depended on
        a pruned branch, treat that dependency as satisfied

  4. FOR EACH node IN mutation.insert_nodes:
     a. Assign opaque integer PK
     b. Insert into dag_nodes with new plan_version
     c. Insert corresponding dag_edges

  5. Validate acyclicity of resulting graph (Kahn's algorithm)

  6. Validate edge schema compatibility (output schema of source
     satisfies input schema of target)

  7. IF validation fails: ROLLBACK, return structured error

  8. Increment plan_version_current on journey

  9. APPEND event('graph_mutated', mutation_details)

  10. COMMIT TRANSACTION
```

#### Lightweight Runtime Commands

Not all control-flow adjustments require graph mutations. Skip, redirect, and short-circuit operations are expressed as `Command` events on `node_runs`, borrowing from LangGraph's `Command` pattern. This reduces mutation overhead for routine routing decisions.

**Example:** If a test-author's output shows no new tests are needed (the existing suite already covers the acceptance criteria), the orchestrator can issue a `skip` command on the test-author node rather than pruning and re-grafting the graph.

### 2.5 Throughput Control (Drum-Buffer-Rope)

Oraculo applies the Theory of Constraints (TOC) Drum-Buffer-Rope (DBR) model to manage throughput:

| DBR Concept | Oraculo Mapping |
|-------------|----------------|
| **Drum** | QA process (bottleneck that sets system pace) |
| **Buffer** | Queue of completed code-generation tasks awaiting QA |
| **Rope** | Feedback mechanism that throttles upstream dispatch when buffer is full |

**Rationale:** QA is the natural bottleneck because it requires multiple agents (four personas), mutation testing, and full re-validation on rejection. If the dispatch loop sends code tasks faster than QA can absorb them, completed branches pile up, consuming disk space and complicating merge ordering.

#### Configuration Tables

```sql
CREATE TABLE squads (
    squad_id    TEXT PRIMARY KEY,
    journey_id  TEXT NOT NULL,
    max_running INTEGER NOT NULL DEFAULT 3,
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);

CREATE TABLE stage_limits (
    stage_name TEXT PRIMARY KEY,
    wip_limit  INTEGER NOT NULL,
    journey_id TEXT NOT NULL,
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);
```

#### Throttling Logic

```
qa_buffer_size = COUNT(*) FROM node_runs nr
                 JOIN dag_nodes dn ON nr.node_id = dn.node_id
                 WHERE nr.status = 'completed'
                 AND dn.type = 'code'
                 AND EXISTS (
                     SELECT 1 FROM dag_edges de
                     JOIN dag_nodes downstream ON de.to_node_id = downstream.node_id
                     WHERE de.from_node_id = nr.node_id
                     AND downstream.type = 'qa'
                     AND downstream.node_id IN (
                         SELECT node_id FROM node_runs WHERE status = 'pending'
                     )
                 )

IF qa_buffer_size >= stage_limits['awaiting_qa'].wip_limit:
    rope_taut = true
    -- dispatch loop skips code-generation nodes
    -- but continues dispatching documentation, research, and other non-bottleneck tasks
```

#### Shifting Bottleneck Detection

The CLI computes rolling average completion times per `node.type`. If a non-QA type's average significantly exceeds QA's, the Drum shifts and the throttling logic redirects accordingly. This prevents rigid assumptions about which stage is the bottleneck.

**MVP simplification:** For the initial implementation, the Drum is fixed at QA. Shifting bottleneck detection is deferred. WIP limits are set to conservative defaults (e.g., 3 concurrent code agents, 2 concurrent QA agents). Limits are stored in SQLite and adjustable without code changes.

### 2.6 Persistence (Event Sourcing + Materialized Tables)

All state changes are appended to an immutable `events` table. Materialized tables (`dag_nodes`, `dag_edges`, `node_runs`) are projections maintained for fast querying.

#### Events Table

```sql
CREATE TABLE events (
    event_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    journey_id   TEXT    NOT NULL,
    event_type   TEXT    NOT NULL,
    payload_json TEXT    NOT NULL,
    node_id      INTEGER,
    node_run_id  INTEGER,
    plan_version INTEGER,
    version      INTEGER NOT NULL,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(journey_id, version)
);
```

**Design decisions:**

- **WAL mode mandatory.** `PRAGMA journal_mode = WAL` is enforced at database initialization. This enables concurrent reads (multiple agents querying status) without blocking writes (the dispatch loop updating state).

- **Synchronous materialization.** Materialized tables are updated within the same transaction that appends the event. This avoids snapshot staleness without the complexity of periodic folding. At Oraculo's scale (dozens to low hundreds of nodes per journey, not millions), synchronous materialization is practical.

- **Recovery.** On restart, the CLI reads the materialized tables directly. If corruption is suspected, it can rebuild by replaying the events table from the beginning.

- **Monotonically increasing version.** The `UNIQUE(journey_id, version)` constraint guarantees event ordering and prevents out-of-order application.

### 2.7 Re-Planning Mechanics

Re-planning creates new plan versions without blocking in-progress execution.

**Triggers:**
- Repeated failures of a critical node (2+ retries exhausted)
- QA feedback requiring architectural changes
- User-initiated scope changes
- Discovery of new architectural context during execution

**Mechanism: Plan Versioning + Optimistic Concurrency Control**

1. The orchestrator reads the current graph state. The CLI provides a `version_id` (ETAG derived from the latest event sequence number).
2. The orchestrator emits a `MutateGraph` call with the new plan and the ETAG.
3. If the graph has changed in the interim (another agent completed, another mutation landed), the CLI rejects with the current state. The orchestrator retries with fresh data.
4. New nodes under the new `plan_version` coexist with in-progress nodes under the old version. Cutover happens at natural phase boundaries.
5. **Localized scoping:** Version tracking is per-branch or per-phase, not global. An unrelated sub-agent completion does not invalidate a concurrent re-plan on a different branch.

**Rationale:** Combining plan versioning (structural mechanism) with OCC (concurrency-safety mechanism) prevents both stale-state corruption and pessimistic locking that would block the dispatch loop.

### 2.8 MVP Simplifications vs. Full Vision

| Aspect | MVP | Full Vision |
|--------|-----|-------------|
| Dispatch priority | FIFO within squad capacity | Critical-path scoring with `(critical_path_length, fan_out)` |
| DBR Drum | Fixed at QA | Dynamic with shifting bottleneck detection |
| Re-planning | Manual trigger by user or orchestrator | Automatic triggers on failure patterns |
| Plan versioning | Single active version | Multiple coexisting versions with localized scoping |
| OCC/ETAG | Simplified: reject all concurrent mutations | Full localized ETAG with per-branch scoping |
| Graph validation | Full acyclicity + edge existence | Full acyclicity + edge existence + I/O schema compatibility |
| Runtime commands | Not implemented; all routing via graph mutations | `Command` events for skip/redirect/short-circuit |

---

## 3. Agent Isolation and Execution Environment

Every code agent operates in an ephemeral, isolated Git worktree. This section defines the worktree lifecycle, context injection, permission enforcement, and resource limits.

### 3.1 Worktree Lifecycle

#### Creation

One worktree per DAG node. The CLI is the sole creator and destroyer. Worktrees are ephemeral -- they exist only for the duration of a node's execution.

```
Path:   .oraculo/worktrees/<dag_run_id>/<node_id>/
Branch: oraculo/<dag_run_id>/<node_id>
```

**Rationale for naming convention:**
- Scoped under the Oraculo namespace, avoiding collision with user-created worktrees.
- Includes `dag_run_id`, enabling multiple DAG runs to coexist (important for retry scenarios and parallel feature tracks).
- Creates a natural filesystem hierarchy for monitoring and cleanup.

#### Base Commit Pinning

Every worktree is created from a known, deterministic commit (`base_sha`) recorded in SQLite. Without pinning, retries or late-scheduled nodes could silently pick up unrelated changes from main.

#### Retry Semantics

On retry, the old worktree is removed before the new one is created. The attempt number is tracked in SQLite but does not alter the path or branch name (the old branch is deleted first). This enforces the "pristine environment" principle -- no state leaks between attempts.

#### Cleanup

After a node completes (regardless of outcome), the CLI runs `git worktree remove` followed by `git worktree prune`. Failed worktrees are retained for a configurable diagnostic window before cleanup.

#### SQLite Schema

```sql
CREATE TABLE worktrees (
    dag_run_id    TEXT    NOT NULL,
    node_id       TEXT    NOT NULL,
    attempt       INTEGER NOT NULL DEFAULT 1,
    base_sha      TEXT    NOT NULL,
    result_sha    TEXT,
    worktree_path TEXT    NOT NULL,
    branch_name   TEXT    NOT NULL,
    status        TEXT    NOT NULL
                  CHECK (status IN (
                      'provisioning','active','succeeded',
                      'failed','cancelled','expired','cleaning'
                  )),
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    cleaned_at    TEXT,
    PRIMARY KEY (dag_run_id, node_id, attempt)
);
```

### 3.2 Context Injection (MVI Working Set Builder)

The agent must not discover its own context. The CLI builds the working set deterministically before the agent starts. This is the **Minimal Viable Information (MVI)** principle.

#### Algorithm

```
BUILD_MVI_BUNDLE(node):
  1. ANCHOR IDENTIFICATION
     The DAG node's task spec names target files/symbols.

  2. RADIUS WALK (depth configurable, default 2)
     Walk the dependency graph from the anchor, collecting direct
     dependencies and dependents. The dependency graph is built via
     Tree-sitter AST parsing and maintained as a persistent
     repository map in SQLite.

  3. IMPORTANCE FILTER
     Within the walked set, prioritize symbols with higher in-degree
     (number of dependents). This is a local approximation of PageRank
     without the global computation cost. Core types and interfaces
     surface naturally.

  4. TOKEN BUDGET TRUNCATION
     Use binary search to fit the selected symbols within the node's
     token budget, dropping lowest-importance symbols first.

  5. AGENT SPEC FILE
     Generate AGENT_<node_id>.md containing:
       - Task goal and acceptance criteria
       - Editable files (write whitelist)
       - Read-only files (dependency neighborhood)
       - Forbidden paths (explicit blacklist)
       - Anti-goals (what the agent must NOT do)
       - Run command for tests
       - Relevant conventions from CLAUDE.md
```

**Rationale:** Starting with a radius-based walk (rather than full PageRank) is simpler, faster, and more deterministic. The in-degree ranking provides sufficient importance signaling without global graph materialization. Full PageRank is deferred to a later optimization phase.

#### Six-Zone Prompt Structure

The MVI bundle is assembled into a deterministic six-zone prompt structure that exploits the LLM's U-shaped attention curve (stronger attention at the beginning and end of the context window):

```
[SYSTEM CONTEXT]        -- from CLAUDE.md (static, project-level conventions)
[TASK CONTEXT]          -- rolling summary of current task
[STRUCTURAL CONTEXT]    -- from MVI/topological mapping (files, dependencies, repo map slice)
[SEMANTIC CONTEXT]      -- ranked facts from memory working set builder (see section 6)
[EPISODIC CONTEXT]      -- last k turns, recent relevant episodes
[INSTRUCTIONS]          -- the actual task directive (AGENT_<node_id>.md)
```

### 3.3 Permission Boundaries

Permission enforcement uses **defense in depth** -- two complementary layers:

#### Layer 1: PreToolUse Hooks (Primary Enforcement)

Claude Code lifecycle hooks intercept every `Edit` or `Write` tool call before the filesystem is mutated. The hook validates the `file_path` against an ACL stored in SQLite:

- **Allowed write globs:** Files the agent is authorized to modify
- **Forbidden globs:** Files the agent must never touch (other agents' targets, config files, test files for implementers)
- **Exit code 2:** Block the action and feed a rejection message back to the LLM

This provides **fail-fast** behavior -- the agent redirects early, saving tokens on unauthorized work.

#### Layer 2: Diff Validation (Secondary Enforcement)

Before the CLI commits the agent's changes, it validates the full diff against the SQLite ACL:

- Files created via shell commands that bypass Edit/Write tools are caught
- Aggregate constraints are enforced: max diff lines, max files changed
- Secret-like literals are detected and rejected

#### PostToolUse Hooks for Quality Gates

After successful writes, trigger linting and formatting via deterministic tools. This shifts syntactic validation from the LLM's stochastic reasoning to binary exit codes.

#### ACL Schema

```sql
CREATE TABLE node_acls (
    dag_run_id          TEXT    NOT NULL,
    node_id             TEXT    NOT NULL,
    allowed_write_globs TEXT    NOT NULL,  -- JSON array of glob patterns
    allowed_read_globs  TEXT    NOT NULL,  -- JSON array of glob patterns
    forbidden_globs     TEXT    NOT NULL,  -- JSON array of glob patterns
    max_diff_lines      INTEGER,
    max_files_changed   INTEGER,
    FOREIGN KEY (dag_run_id, node_id)
        REFERENCES worktrees(dag_run_id, node_id)
);
```

### 3.4 Merge Strategy (Future)

> **Note:** This section describes the future Deliver phase design. It is not part of the current implementation scope. See [Section 10](#10-future-work) for details.

The CLI will be the **sole writer to main** (single-writer principle).

#### Clean Merge Path (CLI-Only)

The CLI will attempt to fast-forward or cherry-pick `result_sha` onto current main. If no conflicts, the CLI commits directly. No agent needed.

#### Conflict Path (Merge Agent)

If conflicts are detected, the CLI will create a new DAG node of type `merge-resolution`. A specialized Merge Agent will be spawned with:
- The conflicting files and their diff hunks
- A repo map slice covering only the conflicting files' dependency neighborhood
- The commit messages and task specs from the conflicting nodes (for semantic intent)

The Merge Agent will receive **only** what it needs -- never the full repository context.

#### Serial Integration Order

Even when multiple nodes complete in parallel, merges will be applied one at a time in topological order. This prevents compound conflicts and keeps the mainline in a known-good state after each merge.

#### Pre-Merge Validation

After every merge (clean or agent-resolved), the CLI will run the validation pipeline (lint, tests, static analysis) before advancing main HEAD. A failed validation rejects the merge and marks the node for investigation.

#### Base SHA Divergence Handling

When a merge succeeds and advances main, any still-running nodes based on an older `base_sha` will **not** be automatically rebased. When those nodes complete, the CLI will detect the divergence and perform the merge/conflict-check at that point. This avoids disrupting in-progress agents.

### 3.5 Resource Limits (Three-Tier Model)

#### Tier 1: Per-Node Hard Limits

| Limit | Description | Default |
|-------|-------------|---------|
| `max_tokens` | Model input + output combined | Configurable per node type |
| `max_steps` | Maximum tool call count (step budget) | Configurable per node type |
| `timeout_seconds` | Wall-clock timeout | Configurable per node type |

On breach: terminate agent, mark node as `circuit_breaker_tripped`, trigger worktree cleanup.

#### Tier 2: Behavioral Detection

- Track consecutive identical tool failures (e.g., same file, same error 3+ times)
- Track token velocity (tokens per meaningful code change) to detect spinning
- On detection: trip the circuit breaker before hitting the hard limit

#### Tier 3: Global Type-Level Throttling

- Track failure rates per node type across the DAG run
- If a node type exceeds a configurable error threshold (e.g., 3 consecutive failures):
  - Reduce parallelism for that type
  - Pause scheduling and escalate to the orchestrator for human review

#### Resource Limits Schema

```sql
CREATE TABLE node_resource_limits (
    dag_run_id      TEXT    NOT NULL,
    node_id         TEXT    NOT NULL,
    max_tokens      INTEGER NOT NULL,
    max_steps       INTEGER NOT NULL,
    timeout_seconds INTEGER NOT NULL,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    steps_used      INTEGER NOT NULL DEFAULT 0,
    started_at      TEXT,
    status          TEXT    NOT NULL
                    CHECK (status IN (
                        'pending','running','succeeded','failed',
                        'circuit_breaker_tripped','cancelled','expired'
                    )),
    failure_reason  TEXT,
    failure_logs    TEXT,   -- JSON, injected into retry MVI bundle
    PRIMARY KEY (dag_run_id, node_id)
);
```

#### Retry Enrichment

When a failed node is retried, the CLI injects a structured failure summary into the new agent's MVI bundle. This summary includes the previous failure reason, the files that were problematic, and error logs -- preventing the new agent from repeating the same mistakes.

### 3.6 Environment Setup

- **Shared toolchain, per-worktree code.** The host machine provides the toolchain. Worktrees contain only source code. Dependencies are installed once and shared via hard links (pnpm for Node.js; equivalent for other ecosystems).
- **Dynamic environment injection.** For nodes that run tests, the CLI generates isolated environment configurations (unique ports, ephemeral database names) to prevent collisions between parallel agents. Injected as environment variables, never written to tracked files.
- **Read-only dependency directories.** PreToolUse hooks prevent agents from modifying `node_modules/`, `vendor/`, or equivalent directories.
- **Secrets never in worktrees.** Secrets are injected via environment variables by the CLI. Environment profiles in SQLite define which secrets and services each node type needs.

### 3.7 MVP Simplifications vs. Full Vision

| Aspect | MVP | Full Vision |
|--------|-----|-------------|
| Context injection | Radius walk (depth 1), no in-degree ranking | Radius walk (depth 2+) with in-degree ranking and token budget binary search |
| Repo map | Static file list from initial scan | Tree-sitter AST-based persistent dependency graph with incremental updates |
| Permission hooks | PreToolUse hooks only | PreToolUse + diff validation + PostToolUse linting |
| Merge strategy (future) | Deferred -- validated branches remain in worktrees | CLI fast-forward merge; Merge Agent for automatic conflict resolution |
| Resource limits | Per-node hard limits only (tokens, steps, timeout) | Three-tier model with behavioral detection and global throttling |
| Environment | Shared host toolchain | Environment profiles in SQLite; potential containerization |

---

## 4. TDD Pipeline

The TDD pipeline enforces test-driven development discipline across all code-producing agents. Test-authors and implementers are distinct agents operating in isolated worktrees with strict permission boundaries and event-sourced trace verification.

### 4.1 Test-Author / Implementer Separation

The test-author and implementer are **distinct Claude Code sub-agents** operating in isolated Git worktrees. This separation is architectural, not advisory.

**Why separate agents?** When a single context holds both test and implementation concerns, the agent optimizes tests for the solution it imagines, producing weak oracles. By shielding the test-author from implementation details, tests validate observable behavior and contracts, not implementation internals.

#### Test-Author Context

The test-author receives **only:**
- Immutable specification documents (requirements, acceptance criteria)
- Technical design documents (for mock schema generation)
- Domain glossary (ubiquitous language from DDD)
- **No access** to existing implementation code or planned architecture

#### Test-Author Prompt Structure

```
ROLE: You are a test-author. Your job is to translate specifications
into executable test cases.

CONTEXT: [Approved specification document, acceptance criteria,
domain glossary]

CONSTRAINTS:
- Do NOT infer or assume implementation details
- Write tests as a consumer of the API/interface
- Every acceptance criterion must map to at least one test case
- Include: happy path, edge cases, boundary conditions, error cases
- Use the domain's ubiquitous language in test names and assertions

OUTPUT FORMAT: [Structured JSON schema with test blocks containing:
description, mocked dependencies, execution call, assertion logic]
```

#### Test Granularity by Change Type

| Change Type | Primary Test Type | Rationale |
|-------------|------------------|-----------|
| New feature | Acceptance/integration tests | Validate user-visible behavior |
| Bug fix | Fail-to-pass reproduction test (unit or integration) | Prove the bug exists, then prove it is fixed |
| Refactoring | No new tests; existing suite is the contract | Existing tests validate behavioral equivalence |

E2E tests are excluded from the TDD loop due to non-determinism and latency.

#### Test-Author Definition of Done

1. All acceptance criteria mapped to test cases
2. Tests pass structural validation (JSON schema)
3. Tests fail deterministically in the "Before" state (verified by N>=2 runs)
4. Tests are syntactically valid and executable

Flakiness is treated as a test-author bug, not a pipeline issue. Flaky tests block the handoff.

### 4.2 Handoff Mechanism ("Before" Branch)

The handoff from test-author to implementer follows a precise branch-based protocol:

```
main
  |
  +-- feature-tests (test-author worktree)
  |     |-- writes tests, commits
  |     |-- orchestrator verifies deterministic failure (N>=2 runs)
  |
  +-- feature-impl (implementer worktree, branched FROM feature-tests)
        |-- tests already present (read-only)
        |-- writes implementation only
```

The orchestrator creates a "Before" branch containing only the new tests, verifies deterministic failure, then derives the implementer's branch from it. The implementer's first `git diff` never includes test changes.

#### Handoff Metadata Bundle (tdd_bundle.json)

```json
{
  "bundle_id": "uuid",
  "story_id": "STORY-123",
  "base_commit": "abc123",
  "test_branch": "feature-tests",
  "test_files": [
    {
      "path": "tests/test_feature.py",
      "test_ids": ["test_happy_path", "test_edge_case", "test_error"],
      "acceptance_criteria_refs": ["AC-1", "AC-2"]
    }
  ],
  "run_command": "pytest tests/test_feature.py --junitxml=report.xml",
  "expected_failures": ["test_happy_path", "test_edge_case", "test_error"],
  "timeout_seconds": 120,
  "determinism_verified": true,
  "determinism_runs": 2,
  "created_at": "ISO-8601"
}
```

#### Permission Enforcement During TDD

| Agent | Write Access | Read Access | Enforcement |
|-------|-------------|-------------|-------------|
| Test-author | `tests/**` | Specifications, domain glossary | PreToolUse hooks block `src/**` writes |
| Implementer | `src/**` | Tests (read-only), specifications | PreToolUse hooks block `tests/**` writes; FS-level read-only as defense-in-depth |

Two-layer enforcement: Claude Code PreToolUse hooks provide fast feedback (blocks before action). Filesystem-level read-only permissions provide defense-in-depth (blocks even if hooks are bypassed).

### 4.3 Red-Green-Refactor Trace Enforcement

Every test run, file write, and commit attempt is recorded as an immutable, append-only event in SQLite. The CLI queries this event stream to prove the TDD sequence was followed.

#### TDD Events Schema

```sql
CREATE TABLE tdd_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    stream_id   TEXT NOT NULL,     -- feature branch / task ID
    session_id  TEXT NOT NULL,     -- agent session
    timestamp   TEXT NOT NULL,     -- ISO-8601 microsecond precision
    event_type  TEXT NOT NULL,     -- SessionStart, TestRunStart, TestRunResult,
                                  -- FileWriteAttempt, FileModified, CommitRequested,
                                  -- CycleMarker, LintRunResult, PolicyViolation
    agent_role  TEXT NOT NULL,     -- test_author, implementer, refactor
    data        TEXT NOT NULL      -- JSON payload
);

CREATE INDEX idx_tdd_events_stream_time ON tdd_events(stream_id, timestamp);
CREATE INDEX idx_tdd_events_stream_type ON tdd_events(stream_id, event_type);
```

#### Key Event Types and Payloads

| Event Type | Payload Fields |
|-----------|----------------|
| `SessionStart` | `worktree_id`, `base_commit`, `story_id`, `tdd_bundle_hash` |
| `TestRunResult` | `exit_code`, `total`, `failed`, `skipped`, `duration_ms`, `report_format`, `report_digest`, `failed_test_ids[]` |
| `FileWriteAttempt` | `path`, `category` (test/prod/config), `allowed`, `blocked_reason`, `before_hash`, `after_hash` |
| `CommitRequested` | `commit_hash`, `files_changed[]`, `validation_result` |
| `CycleMarker` | `phase` (red/green/refactor), `notes` |
| `PolicyViolation` | `violation_type`, `details`, `action_taken` |

#### Validation Algorithm

The CLI runs this validation before accepting any commit from an implementer agent:

```
Given stream_id S and tdd_bundle B:

1. Find T_impl_first = first FileWriteAttempt where category=prod
   AND allowed=true after the handoff event

2. REQUIRE: there exists a TestRunResult R_red where:
   - R_red.timestamp < T_impl_first.timestamp (causal ordering)
   - R_red.exit_code != 0
   - R_red.failed > 0
   - R_red.failed_test_ids INTERSECT B.expected_failures is non-empty

3. Find T_impl_last = last FileWriteAttempt where category=prod

4. REQUIRE: there exists a TestRunResult R_green where:
   - R_green.timestamp > T_impl_last.timestamp
   - R_green.exit_code == 0
   - R_green.failed == 0

5. REQUIRE: no FileWriteAttempt where category=test AND allowed=true
   exists in the implementer's session

If any requirement fails: REJECT commit, log PolicyViolation.
```

**Rationale:** The orchestrator gate is the authoritative enforcement point, not local hooks alone. Local hooks provide fast feedback. The orchestrator gate provides trust.

### 4.4 Test Hacking Defense Layers

Permission boundaries are necessary but insufficient. Defense-in-depth is required:

| Layer | Mechanism | What It Catches |
|-------|-----------|----------------|
| 1. Permission boundary | Test files read-only for implementer | Direct test modification |
| 2. Trace validation | Red-Green sequence proven in SQLite | Skipped Red phase, backdated writes |
| 3. Mutation testing | Detect weak oracles / trivial assertions | Tests that pass on any implementation |
| 4. Test determinism | Verify consistent failure (N>=2 runs) | Flaky tests that produce false Green |
| 5. Behavioral anomaly detection | Detect brute-forcing / excessive test runs | Side-channel test inference |

Behavioral anomaly detection monitors the SQLite event stream for suspicious patterns:
- Excessive test executions without code modifications (brute-forcing)
- Rapid-fire test runs (inferring test logic through side channels)

The CLI flags these patterns for human review. This is a lower priority than mutation testing but provides an additional safety net.

### 4.5 Mutation Testing Integration

Mutation testing validates test quality, not TDD sequence. It answers "are the tests strong enough?" while trace validation answers "was the process followed?" Both are required.

#### Execution Flow

```
All tests pass (Green state confirmed)
  --> CLI runs mutation framework, scoped to the diff
  --> Framework mutates implementation code (e.g., a + b --> a - b)
  --> CLI re-runs test suites against mutated code
  --> Calculate mutation score:
      Mutation Score = Killed Mutants / Total Mutants * 100
  --> Score >= threshold?  --> APPROVED
  --> Score < threshold?   --> REJECTED
      --> CLI parses surviving mutant report
      --> Extracts exact lines and mutation operators that survived
      --> Feeds surviving mutants back to test-author with targeted prompt
      --> Test-author iterates until threshold met or retry limit reached
```

#### Thresholds

- **Standard threshold:** >= 80%
- **High-risk threshold:** >= 90% (for security-sensitive or core-logic changes)

**Key insight:** Mutation testing validates the test-author, not the implementer. A low mutation score is a failure of the test suite. Surviving mutant reports route to QA agents (who wrote the tests), not to the code agent.

#### Tool Mapping

| Language | Tool | JUnit XML |
|----------|------|-----------|
| JavaScript/TypeScript | Stryker Mutator | Native |
| Python | Mutmut | Via adapter |
| Java/Kotlin | PIT (PITest) | Native |
| Go | go-mutesting | Via converter |

#### Execution Policy

| Policy | Scope | Frequency |
|--------|-------|-----------|
| Incremental | Modules touched by current story | Every story completion |
| Full | Entire codebase | Nightly/weekly |

### 4.6 Conditional Refactor Phase

The refactor phase is conditionally triggered by static analysis metrics, not always invoked:

```
After Green phase:
  1. Run lint/format (deterministic, mandatory)
     - If violations: auto-fix or reject
     - Log LintRunResult event

  2. Calculate cyclomatic complexity of new/modified functions
     - If complexity < threshold (e.g., M < 10): skip refactor
     - If complexity >= threshold: invoke Refactor Agent

  3. If Refactor Agent invoked:
     - Separate agent with distinct prompt (structural improvement only)
     - Same permission model: read-only on tests, write on src
     - After each modification: auto-run tests
     - If tests fail: rollback to last green state, retry
     - Log all events to SQLite
```

**Rationale:** Avoiding token spend on refactoring clean code. The cyclomatic complexity trigger ensures the Refactor Agent is invoked only when the code demonstrably benefits from restructuring. Lint/format is always mandatory as a deterministic quality floor.

### 4.7 Multi-Language Test Result Support

All test runners must produce structured output. JUnit XML is the lingua franca:

| Priority | Strategy | When |
|----------|----------|------|
| 1 | Native structured output (JSON/XML) from test runner | Always preferred |
| 2 | JUnit XML via reporter plugin | When native output is not structured |
| 3 | TAP (Test Anything Protocol) parsing | Fallback for constrained environments |
| 4 | Exit code only | Last resort (pass/fail but no test-level detail) |

The parser layer is versioned and tested with synthetic failure injection. Parser errors are treated as trust violations (default to Red).

### 4.8 MVP Simplifications vs. Full Vision

| Aspect | MVP | Full Vision |
|--------|-----|-------------|
| Test-author isolation | Separate worktree, specification-only context | Full GSD/spec-workflow-mcp pattern with structured output validation |
| Handoff verification | Single failure run (N=1) | N>=2 deterministic failure runs |
| Trace validation | Red-before-Green check only | Full causal ordering with FileWriteAttempt tracking |
| Permission enforcement | PreToolUse hooks only | PreToolUse + FS-level read-only + diff validation |
| Mutation testing | Deferred (manual trigger only) | Automatic incremental per-story with threshold gating |
| Refactor phase | Manual (user-triggered) | Conditional on cyclomatic complexity with dedicated agent |
| Anomaly detection | Not implemented | SQLite event stream monitoring for brute-force patterns |

---

## 5. QA Architecture

The QA architecture implements Oraculo's validation pipeline -- the "immune system" that ensures no code is approved without rigorous, independent verification.

### 5.1 Four-Persona Model

Four specialized QA personas validate each code agent's output:

| Persona | Phase | Responsibility | Output | Activation |
|---------|-------|---------------|--------|-----------|
| **Test Architect** | Pre-execution | Risk-based test strategy from PRD/architecture; boundary conditions matrix; adversarial focus scoping | Test Strategy Document | Conditional (high-risk tasks) |
| **Functional QA Reviewer** | Post-execution | Verify code against PRD and acceptance criteria; generate unit/integration tests | Functional Test Suite | Always |
| **Adversarial Security Auditor** | Post-execution | Actively attempt to break code: race conditions, edge cases, security flaws | Adversarial Test Suite (executable scripts only) | Conditional (high-risk tasks) |
| **Style/Convention Checker** | Post-execution | Enforce project coding standards, documentation, dependency hygiene | Linting Report | Always |

**Additional QA functions (not separate personas):**

| Function | Type | Responsibility |
|----------|------|---------------|
| Refactor Agent | Agent (conditional) | Non-behavioral code improvements after Green state (see section 4.6) |
| Behavioral Anomaly Detector | CLI-based (automated) | Monitor SQLite event stream for suspicious patterns (see section 4.4) |

**Conditional activation:** The Test Architect and Adversarial Auditor are conditionally activated based on a risk score assigned during the Plan phase. The Functional Reviewer and Style Checker always run. This avoids burning tokens on adversarial fuzzing for trivial changes.

The risk score is computed from task metadata:
- Files changed and modules affected
- Whether security-sensitive code is touched
- Complexity of the change
- Whether the change touches public API surfaces

### 5.2 Review Topology: Hybrid Parallel-Convergent

```
Phase 1 (Sequential):  Test Architect --> Test Strategy + Adversarial Focus Matrix
Phase 2 (Sequential):  Code Agent --> Implementation + Basic Tests (TDD)
Phase 3 (Parallel):    Functional Reviewer | Adversarial Auditor | Style Checker
Phase 4 (Convergent):  CLI Trust Layer --> Unified Execution --> Mutation Gate --> Verdict
Phase 5 (Conditional): Refactor Agent (if complexity threshold breached)
```

The three post-execution QA agents run in parallel within isolated worktrees, generating independent test suites. Filesystem isolation prevents collisions. The CLI then converges their outputs by executing all test suites deterministically.

### 5.3 Clean Context Enforcement (Bounded Context Payload)

Each QA persona is spawned as a fresh Claude Code sub-agent in a dedicated ephemeral worktree. The agent has **zero access** to:
- The code agent's conversational history
- The code agent's scratchpad or intermediate reasoning
- Any other QA agent's worktree

Each QA agent receives **exactly three inputs:**

1. **Immutable Specification** -- Relevant PRD/story sections defining what the code should do (not what it does). Retrieved from the SQLite blackboard, linked via task ID.

2. **Raw Git Diff** -- Output of `git diff main...feature/<branch>`, forcing attention on changed lines only.

3. **Test Runner Logs** -- Deterministic CLI output from existing test execution in JUnit XML format (pass/fail traces with assertion messages).

No implementation rationale. No code agent reasoning. No shared context windows. If broader context is needed, the agent has read-only access to the local worktree, but its initial context remains aggressively minimized to maximize token efficiency and analytical sharpness.

**Rationale:** "Independence is architectural: clean context, no access to code agent's reasoning." Worktree isolation is necessary but not sufficient -- the bounded payload is equally important.

### 5.4 Executable Proof Pattern (Adversarial Validation)

This is the single most important design decision for adversarial testing. The adversarial agent's textual claims, vulnerability reports, and subjective critiques are **never forwarded to the orchestrator**. Its sole output mechanism is an executable test script.

```
Adversarial agent claims vulnerability
  --> Agent produces executable test script (pytest, Playwright, etc.)
    --> CLI attempts to compile and run the test
      --> Test fails to compile?     --> Claim discarded as hallucination
      --> Test compiles and passes?  --> Code is secure against this vector
      --> Test compiles and FAILS?   --> VULNERABILITY CONFIRMED (valid finding)
```

Only the exit code of the CLI-executed test script determines whether a vulnerability exists. This filters out AI hallucinations entirely and relies on deterministic reality. The orchestrator never reads the adversarial agent's reasoning -- only the test results matter.

**Rationale:** "The CLI verifies, not the QA agent -- deterministic reality over probabilistic opinion." The executable proof pattern makes false positives architecturally impossible.

#### Adversarial Prompt Patterns

The adversarial agent is guided by four categories of attack:

1. **State Manipulation / Environmental Chaos** -- Evaluate code under extreme, out-of-distribution conditions. Simulate network timeouts, disk full, database connection drops during transactions.

2. **Authority Spoofing / Trust Boundary Testing** -- Search for blind trust in the code. Inject mock payloads simulating elevated privileges, attempt to bypass access controls through superficial markers.

3. **Round-Trip Review Protocol (RTRP) / Semantic Drift Detection** -- Give the adversarial agent the implementation without the PRD and ask it to reverse-engineer the specification. If the reverse-engineered spec diverges from the actual PRD, flag a conceptual vulnerability.

4. **Concurrency and Fuzzing Injections** -- Bombard target functions with null bytes, integer overflows, maximum-length strings, and race conditions.

#### Adversarial Scope Control

The Test Architect's pre-execution output includes an adversarial focus matrix that bounds the Adversarial Auditor's attack surface:

```yaml
adversarial_focus:
  priority_vectors:
    - input_validation    # change touches user input parsing
    - error_handling      # change modifies error paths
  excluded_vectors:
    - network_fuzzing     # no network code changed
    - auth_bypass         # auth layer untouched
  max_test_count: 15      # cap to prevent token burn
```

### 5.5 Four-Gate Validation Sub-DAG

The validation pipeline within each Validate node is a sub-DAG with four sequential gates, ordered by cost (cheapest first):

```
Code Agent Completes (commits to feature branch)
  |
  v
Gate 0: TDD Trace Verification (cheapest, runs first)
  - CLI queries SQLite event store for Red -> Code -> Green sequence
  - If trace not proven --> IMMEDIATE REJECTION
  |
  v
Gate 1: Parallel QA Provisioning + Test Generation
  - CLI creates isolated Git worktrees
  - Provisions Functional QA (Model A), Adversarial QA (Model B), Style Checker (Model C)
  - All three consume PRD + Git diff in parallel
  |
  v
Gate 2: Deterministic Test Execution
  - CLI executes ALL test suites in isolated worktree
  - ANY test failure --> REJECTION
  - Failed logs routed to Code Agent for revision
  |
  v
Gate 3: Mutation Validation + Threshold Enforcement
  - CLI runs mutation framework on implementation code, scoped to diff
  - Re-runs test suites against mutated code
  - Mutation score < threshold --> REJECTED (surviving mutants fed back to QA agents)
  - Mutation score >= threshold --> APPROVED
  |
  v
Final Approval: CLI transitions node to "Approved"
  - DAG node marked as approved (or conditional Refactor phase)
```

### 5.6 No-Partial-Approval Policy

If functional tests pass but the adversarial agent uncovers a valid edge case (proven by executable test failure), the **entire validation node fails**. The code agent must fix the edge case, and then the complete validation sub-DAG re-runs from scratch. This prevents regression masking.

**Revision cycles within the DAG:** When QA rejects, the orchestrator does not traverse backward along existing edges (which would violate DAG acyclicity). Instead, it dynamically appends new `Revision --> Review` nodes to the graph, preserving the complete, append-only history of the iteration.

**Circuit breaker:** A maximum of 3 full validation cycles is enforced per task. After 3 rejections, the circuit breaker trips: the DAG branch is paused, a diagnostic summary is logged, and the system escalates.

**Feedback routing on rejection:**

| Gate Failure | Route To | Content |
|-------------|----------|---------|
| Gate 0 (TDD trace) | Code Agent | Protocol violation notice |
| Gate 2 (test execution) | Code Agent | Failed test logs, exact error messages |
| Gate 3 (mutation score) | QA agents | Surviving mutant report with targeted prompt |

### 5.7 Heterogeneous Model Selection

Different LLM providers are assigned to different roles to prevent shared training biases:

| Role | Recommended Model Family | Rationale |
|------|------------------------|-----------|
| Code Generator | Claude (Anthropic) | Strong coding benchmarks, architectural fidelity |
| Functional QA Reviewer | Gemini (Google) | Large context window, strong cross-referencing |
| Adversarial Auditor | DeepSeek R1 / open-weight model | Different training data, structurally distinct |
| Style/Convention Checker | Claude Haiku (Anthropic) | Lightweight, cost-effective for formatting tasks |
| Orchestrator | Claude (Anthropic) | Reasoning and planning capability |

**Hard constraint:** QA agents must use a different model family from the code agents they validate. This is non-negotiable and enforced by the orchestrator configuration.

**Disagreement resolution:** When heterogeneous models disagree, the system does **not** attempt LLM consensus through debate. Instead:
1. The dissenting model must produce an executable test case
2. The CLI runs the test
3. Test fails --> dissenting model wins
4. Test passes --> approving model's verdict stands

This reuses the Executable Proof pattern, creating architectural consistency.

#### Model Configuration Schema

```yaml
model_assignments:
  code_generator:
    primary: claude-sonnet-4-20250514
    fallback: claude-haiku-4-20250514
    max_tokens_per_task: 100000
  functional_qa:
    primary: gemini-2.5-pro
    fallback: claude-sonnet-4-20250514
    max_tokens_per_task: 50000
  adversarial_qa:
    primary: deepseek-r1
    fallback: claude-sonnet-4-20250514
    max_tokens_per_task: 50000
  style_checker:
    primary: claude-haiku-4-20250514
    max_tokens_per_task: 20000
```

**Final authorization is a deterministic SQL query:**

```sql
SELECT
  CASE WHEN
    COUNT(*) = COUNT(CASE WHEN status = 'SUCCESS' THEN 1 END)
    AND MIN(mutation_score) >= 80
  THEN 'APPROVED'
  ELSE 'REJECTED'
  END AS verdict
FROM validation_events
WHERE task_id = ? AND phase = 'validate';
```

The LLM is completely removed from the final authorization decision.

### 5.8 MVP Simplifications vs. Full Vision

| Aspect | MVP | Full Vision |
|--------|-----|-------------|
| QA personas | Functional Reviewer + Style Checker only | All four personas with conditional activation |
| Adversarial testing | Not implemented | Executable Proof pattern with four attack categories |
| Test Architect | Not implemented | Pre-execution risk analysis and adversarial focus matrix |
| Mutation testing | Deferred (manual) | Automatic per-story with threshold gating and feedback loops |
| Model heterogeneity | Single model (Claude) for all roles | Different model families per role |
| Disagreement resolution | Not applicable (single model) | Executable Proof-based arbitration |
| Validation gates | Gate 0 (TDD trace) + Gate 2 (test execution) | All four gates (0, 1, 2, 3) |
| Circuit breaker | Fixed 3-retry limit | Configurable per-node with Debugger agent |

---

## 6. Memory System

Oraculo's memory system implements a three-tier architecture: working memory (ephemeral, assembled at query time), episodic memory (append-only event log), and semantic memory (validated knowledge with temporal versioning). All tiers live in SQLite.

### 6.1 Three-Tier Architecture

| Tier | Persistence | Content | Access Pattern |
|------|------------|---------|---------------|
| **Working** | None (assembled dynamically) | Scored combination of semantic facts + recent episodic entries, fitted to token budget | Assembled by CLI at task dispatch; consumed by agent |
| **Episodic** | Append-only `episodes` table | Every event: test runs, file writes, decisions, failures, completions | Written by agents via CLI; queried for curation and debugging |
| **Semantic** | `semantic_knowledge` table with temporal versioning | Validated facts, patterns, rules, decisions with status lifecycle | Written only after MAR consensus; queried for context assembly |

**Rationale:** Mixing tiers corrupts all three. Raw events (episodic) are too noisy for agent context. Validated knowledge (semantic) is too compressed for debugging. Working memory is ephemeral by design -- assembling it fresh each time prevents context rot.

### 6.2 SQLite Schema

#### Semantic Knowledge

```sql
CREATE TABLE semantic_knowledge (
    knowledge_id   TEXT PRIMARY KEY,     -- typed prefix: "sem-x7k2m"
    type           TEXT NOT NULL,         -- fact, pattern, rule, decision
    content        TEXT NOT NULL,
    confidence     REAL NOT NULL DEFAULT 0.5,
    status         TEXT NOT NULL DEFAULT 'proposed'
                   CHECK (status IN ('proposed','validated','deprecated')),
    provenance_run_id TEXT,              -- link to originating episode
    supersedes_id  TEXT,                  -- self-referencing for versioning
    valid_from     TEXT NOT NULL DEFAULT (datetime('now')),
    valid_to       TEXT,                  -- NULL = currently valid
    embedding      BLOB DEFAULT NULL,    -- for Phase 2 search (nullable from day one)
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (supersedes_id) REFERENCES semantic_knowledge(knowledge_id)
);

CREATE TABLE semantic_knowledge_history (
    history_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    knowledge_id   TEXT NOT NULL,
    previous_status TEXT,
    new_status     TEXT,
    changed_at     TEXT NOT NULL DEFAULT (datetime('now')),
    reason         TEXT
);
-- Trigger: on UPDATE of semantic_knowledge.status, insert into history
```

#### Semantic Edges (Relational Linking)

```sql
CREATE TABLE semantic_edges (
    edge_id      TEXT PRIMARY KEY,       -- typed prefix: "edg-a3b4c"
    subject_id   TEXT NOT NULL,
    predicate    TEXT NOT NULL,           -- e.g., "depends_on", "conflicts_with", "derived_from"
    object_id    TEXT NOT NULL,
    confidence   REAL NOT NULL DEFAULT 1.0,
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (subject_id) REFERENCES semantic_knowledge(knowledge_id),
    FOREIGN KEY (object_id) REFERENCES semantic_knowledge(knowledge_id)
);
```

This triplet schema enables graph-like queries without a graph database. For example, finding all knowledge that conflicts with a given fact:

```sql
SELECT sk.* FROM semantic_knowledge sk
JOIN semantic_edges se ON sk.knowledge_id = se.object_id
WHERE se.subject_id = ? AND se.predicate = 'conflicts_with';
```

#### Episodic Events

```sql
CREATE TABLE episodes (
    id             TEXT PRIMARY KEY,     -- ULID for lexicographic ordering
    agent_id       TEXT NOT NULL,
    run_id         TEXT NOT NULL,
    phase          TEXT NOT NULL,
    event_type     TEXT NOT NULL,
    event_version  INTEGER DEFAULT 1,    -- payload schema version
    payload        TEXT NOT NULL,         -- JSON
    timestamp      TEXT NOT NULL          -- UTC ISO-8601
);

CREATE INDEX idx_episodes_run ON episodes(run_id, timestamp);
CREATE INDEX idx_episodes_type ON episodes(event_type, timestamp);
```

**Design decisions:**

- **ULID primary keys** for lexicographic ordering without timestamp dependency. ULIDs encode time in their structure, enabling chronological queries via `ORDER BY id`.

- **`event_version` field** for forward-compatible deserialization. As payload schemas evolve, older events carry their version number, enabling the CLI to apply the correct parser.

- **Compensating events** rather than mutations. Errors are corrected by appending a compensating event (e.g., `CorrectionEvent` that references the original `id`). No UPDATE or DELETE on the `episodes` table.

#### Active Runs Snapshot

```sql
CREATE TABLE active_runs_snapshot (
    run_id         TEXT PRIMARY KEY,
    latest_status  TEXT NOT NULL,
    latest_event_id TEXT NOT NULL,
    updated_at     TEXT NOT NULL
);
-- Maintained by SQLite triggers on the episodes table
-- Provides O(1) current-state lookups without full event replay
```

#### Schema Migrations

```sql
CREATE TABLE schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  DATETIME DEFAULT (datetime('now')),
    description TEXT NOT NULL
);
```

### 6.3 Event Sourcing Implementation

All state changes are appended to the `episodes` table. Materialized views (`active_runs_snapshot`) are maintained by SQLite triggers for fast lookups.

**Write path:** Agent produces event --> CLI validates structure and provenance --> CLI appends to `episodes` within a transaction --> SQLite trigger updates `active_runs_snapshot`.

**Read path:** CLI queries `active_runs_snapshot` for current state or `episodes` for historical analysis.

**Compaction:** The full archive remains for auditing. The `active_runs_snapshot` table provides the compacted current-state view. This is a hybrid approach: full history for compliance, fast snapshots for operations.

**Recovery:** On restart, the CLI reads the snapshot tables directly. If corruption is suspected, it rebuilds by replaying the episodes table.

### 6.4 Concurrency Model

```
PRAGMA journal_mode = WAL;         -- Non-blocking reads
PRAGMA synchronous = NORMAL;       -- Performance with acceptable safety
PRAGMA busy_timeout = 5000;        -- Wait 5s on contention

-- In Go CLI:
readPool.SetMaxOpenConns(10)       -- High concurrency for reads
writePool.SetMaxOpenConns(1)       -- Serialized writes
```

**Rationale:** WAL mode enables multiple agents to read concurrently without blocking writes. The single-writer pool serializes writes through the CLI, ensuring no concurrent write conflicts. All schema validation and provenance verification happens in-memory before the transaction reaches SQLite.

### 6.5 Curation Pipeline (Episode to Validated Knowledge)

The curation pipeline transforms raw episodic events into validated semantic knowledge:

```
Episode --> Score --> Reflection --> Validated Promotion --> Relational Linking
```

#### Step 1: Scoring

Impact score based on task success, error severity, and temporal recurrence. The composite score determines whether an episode warrants the expensive LLM reflection cycle:

```go
type EpisodeScore struct {
    ImpactLevel     int     // 1-5: how much did this affect outcomes
    ErrorSeverity   int     // 0-5: 0=success, 5=critical failure
    RecurrenceCount int     // how many times this pattern appeared
    Composite       float64 // weighted sum, threshold = 3.0 by default
}
```

Only episodes crossing the severity threshold trigger reflection. This conserves tokens.

#### Step 2: Reflection (Asynchronous)

A dedicated curation agent evaluates episodic trajectories, extracting root causes and behavioral rules. Operates asynchronously, outside the critical execution path. Uses the Reflexion framework.

**Trigger:** Event-driven, not scheduled. Curation fires when a `run_id` completes (terminal event like `TaskCompleted` or `TaskFailed`). The CLI detects this event and enqueues a curation job.

#### Step 3: Promotion (MAR Consensus)

Multi-Agent Reflexion (MAR) -- a jury of distinct LLM personas cross-evaluates proposed insights:

- 3 evaluator agents with distinct system prompts: **skeptic**, **integrator**, **domain expert**
- Majority vote (2/3) required for promotion to `validated` status
- All scores and reasoning traces stored as episodic events for auditability
- CLI enforces the contract: consensus score + provenance links mandatory for any semantic write

#### Step 4: Conflict Handling

New insights contradicting existing knowledge are written as `proposed` with a `conflicts_with` edge. No forced consensus. Human or senior orchestrator resolves. Deprecation requires explicit decision, updating `valid_to` on the old entry.

### 6.6 Context Assembly (Working Memory)

The working set builder assembles context dynamically at dispatch time:

#### Composite FTS5 Scoring

```sql
SELECT knowledge_id, content,
       bm25(semantic_knowledge_fts) * exp(-0.03 * julianday('now') - julianday(valid_from)) AS score
FROM semantic_knowledge_fts
WHERE semantic_knowledge_fts MATCH ?
AND status = 'validated'
ORDER BY score DESC
LIMIT 100;
```

The decay constant (`0.03`) is configurable per project. Recent knowledge scores higher than stale knowledge, but high-relevance old knowledge can still surface.

#### Token Budget Enforcement

Line-limit truncation at the database layer: `LIMIT` and `SUBSTR` enforce the <100 lines (concepts) and <150 lines (guides) budgets. No oversized payloads ever reach the application layer.

### 6.7 Search Evolution Roadmap

All three phases operate within pure SQLite, maintaining the single-storage-engine constraint:

| Phase | Mechanism | Capability |
|-------|-----------|-----------|
| **Phase 1 (MVP)** | FTS5/BM25 with temporal decay | Keyword search with recency bias |
| **Phase 2** | Add `sqlite-vec` embeddings + Reciprocal Rank Fusion via CTEs | Semantic similarity search combined with keyword search |
| **Phase 3** | `WITH RECURSIVE` CTEs against `semantic_edges` | Graph traversal for dependency and relationship queries |

#### Phase 2: Hybrid Search (RRF)

```sql
-- Reciprocal Rank Fusion
WITH fts_results AS (
    SELECT knowledge_id, ROW_NUMBER() OVER (ORDER BY bm25(fts) DESC) AS rank_fts
    FROM semantic_knowledge_fts WHERE fts MATCH ?
),
vec_results AS (
    SELECT knowledge_id, ROW_NUMBER() OVER (ORDER BY vec_distance ASC) AS rank_vec
    FROM semantic_knowledge WHERE embedding IS NOT NULL
    ORDER BY vec_distance_cosine(embedding, ?) ASC
)
SELECT COALESCE(f.knowledge_id, v.knowledge_id) AS knowledge_id,
       1.0/(60 + COALESCE(f.rank_fts, 1000)) + 1.0/(60 + COALESCE(v.rank_vec, 1000)) AS rrf_score
FROM fts_results f
FULL OUTER JOIN vec_results v ON f.knowledge_id = v.knowledge_id
ORDER BY rrf_score DESC
LIMIT 50;
```

#### Phase 3: Graph Traversal

```sql
WITH RECURSIVE reachable(knowledge_id, depth) AS (
    SELECT object_id, 1 FROM semantic_edges
    WHERE subject_id = ? AND predicate = 'depends_on'
    UNION ALL
    SELECT se.object_id, r.depth + 1
    FROM semantic_edges se
    JOIN reachable r ON se.subject_id = r.knowledge_id
    WHERE r.depth < 3  -- depth limit mandatory
)
SELECT sk.* FROM semantic_knowledge sk
JOIN reachable r ON sk.knowledge_id = r.knowledge_id;
```

### 6.8 MVP Simplifications vs. Full Vision

| Aspect | MVP | Full Vision |
|--------|-----|-------------|
| Search | FTS5/BM25 with temporal decay only | Three-phase: FTS5 -> hybrid (sqlite-vec + RRF) -> graph traversal |
| Embeddings | Not generated | Background CLI command for incremental embedding |
| Curation | Manual promotion by orchestrator | Automatic event-driven curation with MAR 3-agent jury |
| Conflict resolution | Manual (orchestrator flags conflicts) | Versioned proposals with `conflicts_with` edges |
| Context assembly | Simple: last k episodes + task spec | Six-zone prompt structure with scored semantic facts |
| Graph queries | Not implemented | `WITH RECURSIVE` on semantic_edges |

---

## 7. Communication, Handoff, and Failure Recovery

This section defines how agents communicate via the SQLite blackboard, how work is handed off between phases, and how the system recovers from failures.

### 7.1 Handoff Contract Schema

Every handoff between agents is a structured JSON document written to a `handoff_contracts` table via the CLI:

```json
{
  "trace_id": "uuid-for-end-to-end-tracing",
  "task_id": "DAG-node-id",
  "source_agent": "test-author-agent-id",
  "target_agent": "implementer-agent-id",
  "payload": {
    "summary": "...",
    "instructions": "...",
    "file_pointers": ["tests/test_feature.py"],
    "verbatim_artifacts": { "tdd_bundle": { "..." } },
    "summarized_history": "..."
  },
  "expected_output_schema": { "$ref": "#/definitions/implementation_result" },
  "constraints": {
    "timeout_seconds": 600,
    "max_retries": 3,
    "token_budget": 100000,
    "negative_constraints": ["Do not modify test files"]
  },
  "idempotency_key": "unique-key-for-dedup",
  "schema_version": "1.0"
}
```

#### Mandatory Fields

| Field | Type | Purpose |
|-------|------|---------|
| `trace_id` | UUID | Cross-hop lineage tracking (W3C Trace Context compatible) |
| `task_id` | string | Binds the handoff to a specific DAG node |
| `source_agent` | string | Provenance and auditability |
| `target_agent` | string | Routing |
| `payload` | object | Operational data: summarized context, instructions, file pointers |
| `expected_output_schema` | JSON Schema | Exact structure the target agent must produce |
| `constraints` | object | Runtime bounds: timeout, max retries, token budget, negative constraints |
| `idempotency_key` | string | Prevents duplicate work during retries (UNIQUE constraint in SQLite) |
| `schema_version` | string | Compatibility management |

#### CLI Validation

Before writing to SQLite, the CLI validates the JSON payload against the registered schema. On validation failure, the CLI returns a structured error message to the source agent, triggering an auto-repair loop. The source agent corrects its output without the malformed payload ever reaching the blackboard.

#### Schema Registry

```sql
CREATE TABLE schema_registry (
    schema_name    TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    schema_json    TEXT NOT NULL,     -- immutable JSON Schema definition
    created_at     TEXT NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (schema_name, schema_version)
);
```

The CLI implements upcasting logic to translate older payloads into current formats where backward-compatible, or explicitly rejects payloads using unsupported versions.

### 7.2 Blackboard Governance (RBAC Model)

Since SQLite lacks native RBAC, governance is implemented through three reinforcing layers:

#### Layer 1: Policy Storage

```sql
CREATE TABLE security_policies (
    role           TEXT NOT NULL,
    table_name     TEXT NOT NULL,
    operation      TEXT NOT NULL CHECK (operation IN ('read','write','delete')),
    allowed        INTEGER NOT NULL DEFAULT 0,
    phase_scope    TEXT,              -- optional: restrict to specific phases
    PRIMARY KEY (role, table_name, operation)
);
```

| Role | Write Access | Read Access |
|------|-------------|-------------|
| Orchestrator | `task_assignments`, `dag_nodes`, `workflow_state`, `checkpoints` | All tables |
| Implementation Agent | Own phase rows in `artifacts`, `implementation_results` | Requirements, architecture docs, own task assignments |
| QA Agent | `validation_verdicts`, `approval_gates` | All implementation artifacts, test results |
| Summarizer Agent | `context_summaries` | Upstream phase artifacts and logs |

#### Layer 2: Database Enforcement

SQLite `BEFORE INSERT`, `BEFORE UPDATE`, and `BEFORE DELETE` triggers cross-reference the session identifier against `security_policies`. Unauthorized writes trigger `RAISE(FAIL, 'insufficient privileges')`, aborting the transaction deterministically.

#### Layer 3: CLI Middleware

The CLI authenticates the agent's persistent identifier upon connection, injects it into a session context, and restricts:
- Which tables/views the agent can query (read-optimized views filter unauthorized data)
- Which tools the agent can invoke (tool permission isolation by role)

A Discover-phase agent must not access build tools. An Implementer must not access external search tools.

### 7.3 SOP Templates as YAML-Driven DAG Templates

SOPs (Standard Operating Procedures) are encoded as YAML-driven DAGs that define reusable workflow patterns:

```yaml
# Standard Development Pipeline SOP
name: standard-dev-pipeline
description: "Full development lifecycle from epic to validated code"
parameters:
  - name: repository_path
    required: true
  - name: testing_framework
    required: true
  - name: language
    required: true

nodes:
  - id: epic-analysis
    type: research
    role: analyst
    phase: discover
    tools: [search, read]
    gate: approval_required

  - id: prd
    type: documentation
    role: product_manager
    phase: discover
    depends_on: [epic-analysis]
    gate: approval_required

  - id: architecture
    type: research
    role: architect
    phase: plan
    depends_on: [prd]
    gate: approval_required

  - id: test-authoring
    type: test-author
    role: test_author
    phase: execute
    depends_on: [architecture]
    sub_dag: tdd-red-phase

  - id: implementation
    type: code
    role: implementer
    phase: execute
    depends_on: [test-authoring]
    sub_dag: tdd-green-phase

  - id: qa-validation
    type: qa
    role: qa_team
    phase: validate
    depends_on: [implementation]
    sub_dag: four-gate-validation

  # merge-deliver node deferred to future Deliver phase
  # - id: merge-deliver
  #   type: merge
  #   role: merge_agent
  #   phase: deliver
  #   depends_on: [qa-validation]
  #   gate: approval_required

failure_hooks:
  default:
    max_retries: 3
    circuit_breaker_threshold: 3
    escalation: human_review
```

**Key properties:**
- The orchestrator is a **generic DAG traversal engine**, not a hardcoded pipeline
- SOPs are the single source of truth for workflow definitions
- **Sub-DAG invocation**: Complex nodes can trigger nested SOP sub-graphs (e.g., a TDD red-green-refactor loop)
- **Revision cycles append new nodes** rather than traversing backward, preserving acyclic properties
- **Parameterization**: The CLI substitutes project-specific variables before populating `dag_nodes` and `dag_edges` tables

### 7.4 Circuit Breaker Implementation (Three-State Machine)

```
CLOSED (normal) --> OPEN (tripped) --> HALF-OPEN (testing) --> CLOSED or OPEN
```

#### State Transitions

| State | Behavior | Transition Trigger |
|-------|----------|-------------------|
| **CLOSED** | All tasks flow normally; metrics tracked | Failure threshold breached --> OPEN |
| **OPEN** | No new tasks dispatched to failing agent/branch; recovery sequence initiates | Cooldown period elapsed --> HALF-OPEN |
| **HALF-OPEN** | Single controlled retry with Debugger agent's refined prompt | Retry succeeds --> CLOSED; Retry fails --> OPEN (with exponential backoff) |

#### Thresholds (Configurable per DAG Node in YAML)

| Trigger | Default Threshold |
|---------|------------------|
| Consecutive validation failures | 3 |
| Error rate over rolling window | 15% within 10-minute epoch |
| Tool execution latency | > 120 seconds |

#### Recovery Sequence

1. **Log failure event:** Exact state, memory context, failure sequence to `failure_diagnostics` table
2. **Deploy Debugger agent:** A specialized diagnostic supervisor that analyzes violated constraints, recent telemetry, and stack traces
3. **Debugger synthesizes:** A targeted remediation prompt for the worker agent
4. **After cooldown:** Attempt single retry with refined prompt (HALF-OPEN state)
5. **If retry succeeds** and passes validation gates: reset to CLOSED
6. **If retry fails:** Revert to OPEN, apply exponential backoff: `T_wait = T_base * 2^failures`
7. **After max retries exceeded** (defined in YAML template): escalate to fallback chain or human intervention

**Deterministic exit conditions:** Maximum retry limits prevent infinite retry loops and force escalation. These limits are defined per DAG node in the YAML template.

### 7.5 Checkpointing and Rollback (Saga Pattern)

#### What a Checkpoint Captures

| Component | Storage | Format |
|-----------|---------|--------|
| DAG node states | `checkpoints` table | Compressed JSON |
| Output artifact references | `checkpoints` table | Pointers to artifact rows |
| Accumulated context/memory | `checkpoints` table | Summarized JSON |
| Schema versions in use | `checkpoints` table | Version identifiers |
| Trace lineage | `checkpoints` table | `trace_id` + `node_id` scope |

#### When Checkpoints Occur

At the completion of every major phase (Discover, Plan, Execute, Validate). Phase boundaries act as global synchronization barriers. When the Deliver phase is implemented (future), it will also trigger checkpoints.

#### Rollback Mechanism (Compensating Transactions)

The blackboard operates as an append-only ledger. Rollback does **NOT** use SQLite ROLLBACK or DELETE. Instead:

1. Orchestrator detects rollback need (circuit breaker exhaustion, QA fundamental rejection)
2. Orchestrator identifies the most recent valid checkpoint for the failing branch (scoped by `trace_id` and `node_id`)
3. Orchestrator writes a **compensating transaction**: a new event indicating the state transition, updates DAG pointers to reference the checkpoint's context
4. Corrupted downstream artifacts are marked as `deprecated` (not deleted)
5. Parallel branches remain entirely unaffected
6. Failed branch restarts from the pristine checkpointed state

**Rationale:** Application-level logical checkpoints (not database-level SAVEPOINTs). Database-level rollback would destroy unrelated parallel branch progress. The Saga pattern preserves the append-only audit trail while enabling branch-scoped recovery.

#### Rollback vs. Retry Decision

| Condition | Action |
|-----------|--------|
| Transient failure (network timeout, rate limit, minor prompt issue) | Retry |
| Circuit breaker exhausted all retries | Rollback |
| QA fundamentally rejects the approach (not just the implementation) | Rollback |
| DAG path is mathematically unviable given accumulated constraints | Rollback + re-plan |

### 7.6 Context Propagation Across Phases

Context propagation follows a strict two-category model:

#### Category 1: Verbatim Artifacts (MUST NOT be summarized)

- Technical specifications
- Required function signatures
- Code diffs and patches
- API schemas (OpenAPI, JSON Schema)
- Explicit user constraints
- Test results and exact error messages

#### Category 2: Summarized Procedural History (CAN be lossily compressed)

- How the upstream agent arrived at conclusions
- Dead-ends explored
- Conversational history with the user
- Intermediate reasoning steps

**Summarization responsibility:** A dedicated Summarizer agent, using a fast/cheap model (e.g., Claude Haiku), applies map-reduce summarization over upstream logs before writing the summary to the blackboard. The orchestrator delegates this -- it does not summarize itself.

**Session termination:** When an agent completes a phase, it writes its structured output artifact to the blackboard. Its session is then completely terminated. The downstream agent starts with a fresh context window. No raw chat history crosses phase boundaries. This prevents context rot.

### 7.7 Traceability (trace_id)

Every handoff contract carries a `trace_id` that propagates through the entire lifecycle of a task. This enables:

- **End-to-end distributed tracing:** Follow a task from Discover through Validate (and through Deliver when implemented)
- **Failure forensics:** When a circuit breaker trips, trace all upstream events
- **Audit compliance:** Prove that every validation gate was passed

The `trace_id` is compatible with W3C Trace Context standards (`traceparent`, `tracestate`), enabling integration with observability tools.

### 7.8 MVP Simplifications vs. Full Vision

| Aspect | MVP | Full Vision |
|--------|-----|-------------|
| Handoff contracts | Simplified schema: task_id, payload, constraints | Full schema with trace_id, idempotency_key, schema_version, auto-repair loops |
| Governance | CLI middleware only (no SQLite triggers) | Three-layer RBAC: policy table + triggers + CLI middleware |
| SOP templates | Hardcoded pipeline in orchestrator prompt | YAML-driven DAG templates with parameterization and sub-DAG invocation |
| Circuit breaker | Fixed 3-retry limit, no Debugger agent | Three-state machine with Debugger agent, exponential backoff, configurable thresholds |
| Checkpointing | Phase-boundary checkpoints only | Branch-scoped checkpoints with compensating transactions |
| Context propagation | Verbatim artifacts only (no summarization) | Verbatim + summarized history via dedicated Summarizer agent |
| Traceability | Simple `task_id` threading | Full `trace_id` with W3C Trace Context compatibility |
| Schema registry | Not implemented; schemas embedded in code | Immutable `schema_registry` table with upcasting |

---

## 8. Cross-Cutting Concerns

### 8.1 Agent Roles Catalog

| Role | Type | Model Tier | Phase | Description |
|------|------|-----------|-------|-------------|
| **Orchestrator** | Coordinator | High-capability (Claude Opus/Sonnet) | All | Plans, delegates, re-plans. Never executes. Proposes DAG mutations; CLI validates. |
| **Code Agent** | Executor | High-capability (Claude Sonnet) | Execute | Writes implementation code following TDD. Operates in isolated worktree with AGENT spec contract. |
| **Test-Author** | Executor | High-capability (Claude Sonnet) | Execute | Writes test cases from specifications only. No access to implementation code or architecture. |
| **Functional QA Reviewer** | Validator | Cross-family (Gemini) | Validate | Verifies code against PRD and acceptance criteria. Generates functional test suites. |
| **Adversarial Security Auditor** | Validator | Cross-family (DeepSeek R1) | Validate | Actively attempts to break code via executable proof pattern. |
| **Style/Convention Checker** | Validator | Low-cost (Claude Haiku) | Validate | Enforces coding standards, documentation, dependency hygiene. |
| **Test Architect** | Planner | High-capability | Plan/Validate | Risk-based test strategy; adversarial focus matrix. Conditional activation. |
| **Merge Agent** (future) | Executor | High-capability | Deliver (future) | Resolves merge conflicts. Receives only conflicting files and their dependency neighborhood. |
| **Debugger Agent** | Diagnostician | High-capability | Recovery | Analyzes circuit breaker failures. Synthesizes remediation prompts for retried agents. |
| **Summarizer Agent** | Transformer | Low-cost (Claude Haiku) | Transitions | Map-reduce summarization of procedural history at phase boundaries. |
| **Refactor Agent** | Executor | High-capability | Execute (conditional) | Structural code improvements triggered by cyclomatic complexity. Same test-contract constraints. |
| **Curation Agent** | Evaluator | Varied (3-agent jury) | Background | Reflexion-based evaluation of episodic events for promotion to semantic knowledge. |

### 8.2 Model Selection Strategy per Role

The model selection strategy follows three principles:

1. **Heterogeneity for validation:** QA agents must use different model families from code agents. This is enforced by configuration, not convention.

2. **Cost proportionality:** Lightweight tasks (summarization, style checking) use small, fast models. Complex tasks (code generation, adversarial testing, debugging) use high-capability models.

3. **Static by default:** Model assignments are configured in the orchestrator manifest (YAML). Dynamic selection based on task complexity is deferred to post-MVP.

| Capability Need | Model Tier | Example Roles |
|----------------|-----------|---------------|
| Complex reasoning + code generation | Tier 1 (highest capability) | Orchestrator, Code Agent, Debugger |
| Analytical review + cross-referencing | Tier 1 (cross-family) | Functional QA, Adversarial Auditor |
| Pattern matching + formatting | Tier 3 (fast, cheap) | Style Checker, Summarizer |

Fallback chains are configured per role. If the primary model's API is unavailable, the CLI falls back to the configured alternative.

### 8.3 Error Handling Philosophy

Error handling in Oraculo follows a hierarchy:

1. **Prevent errors structurally.** PreToolUse hooks, ACLs, schema validation, and bounded context payloads prevent most errors from occurring. The "Pit of Success" principle: make the correct path the easiest path.

2. **Detect errors deterministically.** The CLI detects errors through exit codes, diff validation, trace validation, and mutation scores. No error detection relies on LLM judgment.

3. **Recover without data loss.** All state is append-only. Rollback uses compensating transactions on the Saga pattern. Failed artifacts are deprecated, not deleted. Parallel branches are never affected by a single branch's failure.

4. **Escalate with context.** When automated recovery fails (circuit breaker exhausted), the system provides the human with a complete diagnostic: failure logs, telemetry, agent traces, and the specific checkpoint to resume from.

5. **Learn from errors.** The curation pipeline extracts patterns from failures and promotes them to semantic knowledge, preventing the same class of error from recurring.

### 8.4 Security Boundaries

| Boundary | Enforcement |
|----------|------------|
| Agents cannot write to unauthorized files | PreToolUse hooks + diff validation (defense in depth) |
| Agents cannot read other agents' in-progress work | Worktree isolation (separate filesystem trees) |
| Agents cannot modify dependency directories | PreToolUse hooks block writes to node_modules/, vendor/, etc. |
| Agents cannot access secrets directly | Secrets injected via environment variables by CLI, never stored in worktree |
| Agents cannot merge to main | CLI is sole writer to mainline branch (enforced when Deliver phase is implemented) |
| Agents cannot write to SQLite directly | CLI is sole validation gateway for all writes |
| Agents cannot invoke unauthorized tools | Tool permission isolation by role in CLI middleware |
| QA agents cannot share context with code agents | Separate worktrees + bounded context payloads |
| Malformed payloads cannot enter the blackboard | CLI validates against JSON Schema before writing |

---

## 9. MVP Definition

### 9.1 What Is Included: Simplified Complete Pipeline

The MVP implements the Discover --> Plan --> Execute --> Validate pipeline with all subsystems present in simplified form. No active subsystem is omitted; each is implemented at its minimum viable level. The Deliver phase is explicitly deferred.

#### Phase Implementation

| Phase | MVP Scope |
|-------|-----------|
| **Discover** | Orchestrator conducts Socratic exploration, produces requirements document. Existing skill (`/oraculo:epic`, `/oraculo:story`). |
| **Plan** | Orchestrator decomposes requirements into a DAG. CLI validates and persists. FIFO dispatch (no critical-path scoring). Fixed WIP limits. |
| **Execute** | Code Agent + Test-Author in isolated worktrees. PreToolUse hooks for permission enforcement. Simplified MVI context (radius walk depth 1, no in-degree ranking). TDD trace validation (Red-before-Green check). |
| **Validate** | Functional QA Reviewer + Style Checker. Gate 0 (TDD trace) + Gate 2 (test execution). Single model family (Claude). No mutation testing. No adversarial testing. |
| **Deliver** | Deferred. See [Section 10](#10-future-work). |

#### Subsystem Simplifications

| Subsystem | MVP | Deferred |
|-----------|-----|---------|
| **DAG Engine** | Dual representation (JSON + SQLite). Node/Run separation. FIFO dispatch. Fixed WIP limits. Manual re-planning. | Critical-path scoring, shifting bottleneck detection, automatic re-planning, OCC with localized ETAGs, runtime commands |
| **Agent Isolation** | One worktree per node. Base SHA pinning. AGENT spec files. PreToolUse hooks. | Diff validation, PostToolUse linting, Merge Agent (future, Deliver phase), behavioral detection, global throttling |
| **TDD Pipeline** | Separate test-author/implementer. "Before" branch handoff. Red-before-Green validation. Single failure run for determinism. | N>=2 determinism runs, mutation testing, refactor phase, behavioral anomaly detection |
| **QA Architecture** | Functional Reviewer + Style Checker. Test execution gate. No-partial-approval. 3-retry circuit breaker. | Adversarial Auditor, Test Architect, mutation gate, heterogeneous models, Debugger agent |
| **Memory System** | Episodic events in SQLite. FTS5/BM25 search. Simple context assembly (task spec + recent episodes). | Semantic knowledge tier, curation pipeline, MAR consensus, RRF hybrid search, graph traversal |
| **Communication** | Simplified handoff contracts. CLI middleware for governance. Fixed retry limits. Phase-boundary checkpoints. | Full RBAC with triggers, YAML SOP templates, Debugger agent, Summarizer agent, trace_id, schema registry |

### 9.2 What Is Explicitly Deferred

The following capabilities are **not** part of the MVP:

1. **Dynamic re-planning** -- No automatic triggers for re-planning. The user or orchestrator manually initiates re-plans.
2. **Semantic memory tier** -- No validated knowledge with temporal versioning. Only episodic events and simple search.
3. **Adversarial QA** -- No Adversarial Security Auditor, no Executable Proof pattern. Security testing is manual.
4. **Mutation testing** -- No automatic mutation score gating. Can be triggered manually.
5. **Heterogeneous models** -- Single model family (Claude) for all roles. Cross-family validation is deferred.
6. **YAML SOP templates** -- Pipeline logic is embedded in the orchestrator prompt, not in parameterized YAML files.
7. **Debugger Agent** -- Failed tasks are retried with enriched context (failure logs injected) but without specialized diagnostic analysis.
8. **Summarizer Agent** -- Context is propagated as verbatim artifacts only. No lossy compression at phase boundaries.
9. **Merge Agent** -- Merge conflicts pause the pipeline for human intervention. Part of the deferred Deliver phase.
10. **Deliver phase (Integration)** -- Merging validated agent work back to mainline, serial integration, pre-merge validation, and markdown summary generation are all deferred. See [Section 10](#10-future-work).
11. **Curation pipeline** -- No automatic promotion of episodic knowledge to semantic knowledge.
12. **Shifting bottleneck detection** -- QA is the fixed Drum in DBR.
13. **Schema registry** -- Handoff schemas are embedded in code, not stored as versioned records.

### 9.3 Success Criteria for MVP

The MVP is successful when:

1. **End-to-end flow works.** A user can go from idea to validated code through the appropriate mode's phases (full or reduced) with agent assistance.
2. **TDD is enforced.** The CLI rejects code that was not developed test-first (Red-before-Green proof in SQLite).
3. **Isolation is real.** Agents cannot modify files outside their ACL. Worktrees are created and destroyed cleanly.
4. **QA is independent.** The Functional Reviewer operates in a separate worktree with no access to the code agent's reasoning.
5. **The DAG drives execution.** Tasks are dispatched based on dependency resolution, not manual sequencing. At least 2 agents can run in parallel.
6. **State survives crashes.** The CLI can resume a journey from SQLite state after an unexpected termination.
7. **The pipeline rejects bad code.** QA rejection triggers rework, not approval. The no-partial-approval policy holds.

---

## 10. Future Work

### 10.1 Prioritized Gap List

Gaps are categorized by priority relative to the MVP:

#### High Priority (Post-MVP, First Iteration)

| Gap | Source | Rationale |
|-----|--------|-----------|
| Mutation testing integration | Synthesis 03, 04 | Core quality mechanism; tests without mutation validation have unknown strength |
| Heterogeneous model selection | Synthesis 04 | Prevents shared training bias in QA; current single-model setup risks false consensus |
| Adversarial QA with Executable Proof | Synthesis 04 | Catches security and edge-case issues that functional testing misses |
| YAML SOP templates | Synthesis 06 | Decouples pipeline logic from orchestrator prompt; enables workflow versioning |
| Deliver phase (Integration) | Design v1 | The full Deliver phase: CLI fast-forward merge of validated branches onto mainline, serial integration order, pre-merge validation (lint + tests + static analysis), markdown summary generation, and base SHA divergence handling. Includes the Merge Agent for automatic conflict resolution when fast-forward fails, receiving only conflicting files and their dependency neighborhood. Without this phase, validated branches remain in worktrees and require manual integration. |
| Semantic memory tier | Synthesis 05 | Without validated knowledge, every journey starts from scratch -- no organizational learning |
| Critical-path dispatch scoring | Synthesis 01 | FIFO dispatch wastes time on non-critical nodes while the critical path starves |

#### Medium Priority (Second Iteration)

| Gap | Source | Rationale |
|-----|--------|-----------|
| Debugger Agent | Synthesis 06 | Retries without diagnosis repeat the same mistakes |
| Summarizer Agent | Synthesis 06 | Verbatim-only propagation wastes tokens on procedural history |
| Curation pipeline (episode to knowledge) | Synthesis 05 | Without curation, episodic memory grows unbounded without insight extraction |
| Full RBAC with SQLite triggers | Synthesis 06 | CLI middleware is a soft boundary; triggers provide deterministic enforcement |
| Schema registry with upcasting | Synthesis 06 | Schema evolution without a registry leads to silent deserialization failures |
| Diff validation (Layer 2 permission enforcement) | Synthesis 02 | Catches writes that bypass PreToolUse hooks (shell commands) |
| PostToolUse linting hooks | Synthesis 02 | Shifts syntactic validation from LLM to deterministic tools |
| N>=2 determinism runs for test handoff | Synthesis 03 | Single-run flakiness detection is insufficient |
| Behavioral anomaly detection | Synthesis 03 | Secondary defense against test hacking patterns |
| Conditional refactor phase | Synthesis 03 | Avoids token spend on clean code; targets complex implementations |
| Risk classification algorithm for QA activation | Synthesis 04 | Currently no mechanism to scale QA depth based on task risk |

#### Low Priority (Third Iteration and Beyond)

| Gap | Source | Rationale |
|-----|--------|-----------|
| Dynamic re-planning with automatic triggers | Synthesis 01 | Manual re-planning is adequate until scale justifies automation |
| Shifting bottleneck detection | Synthesis 01 | Fixed QA Drum is acceptable at small scale |
| Localized OCC with per-branch ETAGs | Synthesis 01 | Simplified OCC is sufficient for low-concurrency scenarios |
| Runtime commands (skip/redirect) | Synthesis 01 | Graph mutations handle all routing in MVP |
| sqlite-vec hybrid search (Phase 2) | Synthesis 05 | FTS5 is sufficient until semantic knowledge volume grows significantly |
| Graph traversal search (Phase 3) | Synthesis 05 | Requires sufficient semantic_edges density to be useful |
| Embedding generation pipeline | Synthesis 05 | Blocked by sqlite-vec integration |
| Dynamic model selection per task | Synthesis 04 | Static assignment is sufficient; dynamic adds complexity without clear ROI |
| Containerization per worktree | Synthesis 02 | Only justified for incompatible toolchains or untrusted code |
| Multi-language repo map (polyglot AST) | Synthesis 02 | Needed for polyglot projects but not blocking for single-language repos |
| Checkpoint storage pruning and archival | Synthesis 06 | Only relevant for very long-lived projects |
| Event stream partitioning for long-lived projects | Synthesis 05 | Snapshot tables handle read performance at current scale |
| Write queue observability and metrics | Synthesis 05 | Important for production but not functionally blocking |
| Cross-platform worktree behavior (Windows) | Synthesis 02 | All current development assumes Unix-like systems |
| Human-in-the-loop escalation UX | Synthesis 04, 06 | Notification channels, diagnostic summary format, re-entry flow |
| Mutation tool output standardization | Synthesis 04 | CLI adapter layer needed for cross-tool report parsing |
| Adversarial prompt library versioning | Synthesis 04 | Four patterns are a starting point; needs extensible library |
| Concrete Kahn's algorithm for incremental validation | Synthesis 01 | Full-graph re-validation is acceptable at current scale |

### 10.2 Unresolved Design Questions

These questions emerged across multiple syntheses and require deliberate design decisions:

1. **Incremental TDD within a single story.** When a story decomposes into multiple functions, should the test-author write all tests upfront or should there be multiple Red-Green cycles within a single story? Neither research stream addressed the granularity of cycles within a task.

2. **Agent-to-agent communication.** The current model assumes total isolation between agents. Are there edge cases where a lightweight read-only channel is needed (e.g., checking if a dependency interface has been finalized by a parallel agent)?

3. **Merge ordering heuristics (future, Deliver phase).** Both syntheses agree on serial merging but do not specify the ordering strategy when multiple nodes complete simultaneously. Options: topological order (dependencies first), smallest-diff-first, or longest-running-first.

4. **Working set + MVI context merge.** Two independent context streams -- semantic/episodic from the memory system and structural/topological from the execution environment -- must be unified into a single prompt. The six-zone structure is proposed but the merging algorithm is unspecified.

5. **Buffer sizing for DBR.** Little's Law is referenced but no concrete formula or default value is provided. What is the initial QA buffer limit?

6. **Schema migration for event sourcing.** Event sourcing schemas are notoriously difficult to migrate. When a new event type is introduced or an existing payload format changes, what is the migration strategy?

7. **LLM context window management for large graphs.** How to present large DAGs to the LLM orchestrator without exceeding context limits? Filtering, summarization, or progressive disclosure strategies are not specified.

---

## Appendix A: Complete SQLite Schema (MVP)

This appendix consolidates all SQLite tables referenced throughout the document into a single, ready-to-implement schema for the MVP:

```sql
-- ============================================================
-- Oraculo Agent System - MVP SQLite Schema
-- ============================================================

PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA busy_timeout = 5000;
PRAGMA foreign_keys = ON;

-- ============================================================
-- A. DAG Engine
-- ============================================================

CREATE TABLE journeys (
    journey_id           TEXT    PRIMARY KEY,
    root_goal            TEXT    NOT NULL,
    plan_version_current INTEGER NOT NULL DEFAULT 1,
    status               TEXT    NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','completed','failed','paused')),
    created_at           TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE dag_nodes (
    node_id           INTEGER PRIMARY KEY,
    journey_id        TEXT    NOT NULL,
    plan_version      INTEGER NOT NULL,
    human_label       TEXT    NOT NULL,
    type              TEXT    NOT NULL
                      CHECK (type IN ('code','qa','research','documentation',
                                      'merge','test-author','refactor')),
    squad_id          TEXT,
    phase             TEXT    NOT NULL      -- deliver value retained for future use
                      CHECK (phase IN ('discover','plan','execute','validate','deliver')),
    isolation         TEXT    NOT NULL DEFAULT 'worktree'
                      CHECK (isolation IN ('worktree','inline')),
    input_schema_json TEXT,
    output_schema_json TEXT,
    is_terminal       INTEGER NOT NULL DEFAULT 0,
    created_at        TEXT    NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);

CREATE TABLE dag_edges (
    edge_id       INTEGER PRIMARY KEY,
    journey_id    TEXT    NOT NULL,
    plan_version  INTEGER NOT NULL,
    from_node_id  INTEGER NOT NULL,
    to_node_id    INTEGER NOT NULL,
    condition     TEXT,
    active        INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (from_node_id) REFERENCES dag_nodes(node_id),
    FOREIGN KEY (to_node_id) REFERENCES dag_nodes(node_id)
);

CREATE TABLE node_runs (
    node_run_id   INTEGER PRIMARY KEY,
    node_id       INTEGER NOT NULL,
    journey_id    TEXT    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending','running','completed',
                                    'failed','pruned','cancelled')),
    result_ref    TEXT,
    error_kind    TEXT,
    retry_count   INTEGER NOT NULL DEFAULT 0,
    started_at    TEXT,
    finished_at   TEXT,
    FOREIGN KEY (node_id) REFERENCES dag_nodes(node_id)
);

CREATE TABLE squads (
    squad_id    TEXT    PRIMARY KEY,
    journey_id  TEXT    NOT NULL,
    max_running INTEGER NOT NULL DEFAULT 3,
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);

CREATE TABLE stage_limits (
    stage_name TEXT    PRIMARY KEY,
    wip_limit  INTEGER NOT NULL,
    journey_id TEXT    NOT NULL,
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);

-- ============================================================
-- B. Events (Event Sourcing)
-- ============================================================

CREATE TABLE events (
    event_id     INTEGER PRIMARY KEY AUTOINCREMENT,
    journey_id   TEXT    NOT NULL,
    event_type   TEXT    NOT NULL,
    payload_json TEXT    NOT NULL,
    node_id      INTEGER,
    node_run_id  INTEGER,
    plan_version INTEGER,
    version      INTEGER NOT NULL,
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(journey_id, version)
);

-- ============================================================
-- C. Agent Isolation
-- ============================================================

CREATE TABLE worktrees (
    dag_run_id    TEXT    NOT NULL,
    node_id       TEXT    NOT NULL,
    attempt       INTEGER NOT NULL DEFAULT 1,
    base_sha      TEXT    NOT NULL,
    result_sha    TEXT,
    worktree_path TEXT    NOT NULL,
    branch_name   TEXT    NOT NULL,
    status        TEXT    NOT NULL
                  CHECK (status IN (
                      'provisioning','active','succeeded',
                      'failed','cancelled','expired','cleaning'
                  )),
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    cleaned_at    TEXT,
    PRIMARY KEY (dag_run_id, node_id, attempt)
);

CREATE TABLE node_acls (
    dag_run_id          TEXT    NOT NULL,
    node_id             TEXT    NOT NULL,
    allowed_write_globs TEXT    NOT NULL,
    allowed_read_globs  TEXT    NOT NULL,
    forbidden_globs     TEXT    NOT NULL,
    max_diff_lines      INTEGER,
    max_files_changed   INTEGER,
    PRIMARY KEY (dag_run_id, node_id)
);

CREATE TABLE node_resource_limits (
    dag_run_id      TEXT    NOT NULL,
    node_id         TEXT    NOT NULL,
    max_tokens      INTEGER NOT NULL,
    max_steps       INTEGER NOT NULL,
    timeout_seconds INTEGER NOT NULL,
    tokens_used     INTEGER NOT NULL DEFAULT 0,
    steps_used      INTEGER NOT NULL DEFAULT 0,
    started_at      TEXT,
    status          TEXT    NOT NULL DEFAULT 'pending'
                    CHECK (status IN (
                        'pending','running','succeeded','failed',
                        'circuit_breaker_tripped','cancelled','expired'
                    )),
    failure_reason  TEXT,
    failure_logs    TEXT,
    PRIMARY KEY (dag_run_id, node_id)
);

-- ============================================================
-- D. TDD Pipeline
-- ============================================================

CREATE TABLE tdd_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    stream_id   TEXT    NOT NULL,
    session_id  TEXT    NOT NULL,
    timestamp   TEXT    NOT NULL,
    event_type  TEXT    NOT NULL,
    agent_role  TEXT    NOT NULL,
    data        TEXT    NOT NULL
);

CREATE INDEX idx_tdd_events_stream_time ON tdd_events(stream_id, timestamp);
CREATE INDEX idx_tdd_events_stream_type ON tdd_events(stream_id, event_type);

-- ============================================================
-- E. Memory System
-- ============================================================

CREATE TABLE episodes (
    id             TEXT    PRIMARY KEY,
    agent_id       TEXT    NOT NULL,
    run_id         TEXT    NOT NULL,
    phase          TEXT    NOT NULL,
    event_type     TEXT    NOT NULL,
    event_version  INTEGER DEFAULT 1,
    payload        TEXT    NOT NULL,
    timestamp      TEXT    NOT NULL
);

CREATE INDEX idx_episodes_run ON episodes(run_id, timestamp);
CREATE INDEX idx_episodes_type ON episodes(event_type, timestamp);

CREATE TABLE active_runs_snapshot (
    run_id          TEXT PRIMARY KEY,
    latest_status   TEXT NOT NULL,
    latest_event_id TEXT NOT NULL,
    updated_at      TEXT NOT NULL
);

-- Semantic knowledge (schema prepared, curation deferred to post-MVP)
CREATE TABLE semantic_knowledge (
    knowledge_id      TEXT    PRIMARY KEY,
    type              TEXT    NOT NULL,
    content           TEXT    NOT NULL,
    confidence        REAL    NOT NULL DEFAULT 0.5,
    status            TEXT    NOT NULL DEFAULT 'proposed'
                      CHECK (status IN ('proposed','validated','deprecated')),
    provenance_run_id TEXT,
    supersedes_id     TEXT,
    valid_from        TEXT    NOT NULL DEFAULT (datetime('now')),
    valid_to          TEXT,
    embedding         BLOB    DEFAULT NULL,
    created_at        TEXT    NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (supersedes_id) REFERENCES semantic_knowledge(knowledge_id)
);

CREATE TABLE semantic_edges (
    edge_id    TEXT PRIMARY KEY,
    subject_id TEXT NOT NULL,
    predicate  TEXT NOT NULL,
    object_id  TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 1.0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (subject_id) REFERENCES semantic_knowledge(knowledge_id),
    FOREIGN KEY (object_id) REFERENCES semantic_knowledge(knowledge_id)
);

-- ============================================================
-- F. Communication and Handoff
-- ============================================================

CREATE TABLE handoff_contracts (
    contract_id     TEXT    PRIMARY KEY,
    trace_id        TEXT    NOT NULL,
    task_id         TEXT    NOT NULL,
    source_agent    TEXT    NOT NULL,
    target_agent    TEXT    NOT NULL,
    payload_json    TEXT    NOT NULL,
    output_schema   TEXT    NOT NULL,
    constraints_json TEXT   NOT NULL,
    idempotency_key TEXT    UNIQUE,
    schema_version  TEXT    NOT NULL DEFAULT '1.0',
    status          TEXT    NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','accepted','completed','failed','cancelled')),
    created_at      TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE checkpoints (
    checkpoint_id   TEXT    PRIMARY KEY,
    journey_id      TEXT    NOT NULL,
    trace_id        TEXT    NOT NULL,
    node_id         TEXT,
    phase           TEXT    NOT NULL,
    state_json      TEXT    NOT NULL,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (journey_id) REFERENCES journeys(journey_id)
);

-- ============================================================
-- G. Schema Management
-- ============================================================

CREATE TABLE schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  DATETIME DEFAULT (datetime('now')),
    description TEXT NOT NULL
);
```

---

## Appendix B: Glossary

| Term | Definition |
|------|-----------|
| **ACL** | Access Control List -- glob patterns defining which files an agent can read/write |
| **Base SHA** | The Git commit hash from which a worktree is created; pinned to prevent drift |
| **Bounded Context Payload** | The minimal, curated set of inputs a QA agent receives (spec + diff + test logs) |
| **Circuit Breaker** | Three-state machine (CLOSED/OPEN/HALF-OPEN) that prevents infinite retry loops |
| **Compensating Transaction** | An append-only event that logically reverses a previous state change without deleting data |
| **DAG** | Directed Acyclic Graph -- the task dependency model |
| **DBR** | Drum-Buffer-Rope -- Theory of Constraints throughput control mechanism |
| **Dispatch Loop** | The CLI-owned autonomous loop that computes the frontier and spawns agents |
| **ETAG** | Version identifier used for Optimistic Concurrency Control on graph mutations |
| **Executable Proof** | Pattern where adversarial claims are only accepted as executable test scripts that the CLI runs |
| **Frontier** | The set of DAG nodes with all upstream dependencies satisfied (in-degree zero in active graph) |
| **Journey** | A complete lifecycle from idea to validated code; the top-level container for a DAG |
| **MAR** | Multi-Agent Reflexion -- a jury of distinct LLM personas that validate knowledge promotion |
| **MVI** | Minimal Viable Information -- the principle that agents receive only what they need |
| **Node Run** | A runtime execution instance of a static DAG node (supports retries) |
| **OCC** | Optimistic Concurrency Control -- reject-on-conflict rather than lock-before-read |
| **Plan Version** | Monotonically increasing integer tracking DAG structure evolution |
| **Saga Pattern** | Distributed transaction pattern using compensating actions for rollback |
| **SOP** | Standard Operating Procedure -- a YAML-driven DAG template for reusable workflows |
| **Trust Layer** | The CLI -- the deterministic, validated core that all agents depend on |
| **WAL** | Write-Ahead Logging -- SQLite mode enabling concurrent reads with serialized writes |
| **Working Set** | The dynamically assembled context bundle injected into an agent at dispatch time |
