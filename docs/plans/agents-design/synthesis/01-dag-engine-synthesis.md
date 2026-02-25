# Synthesis: DAG Engine and Orchestration Mechanics

## Sources

- **Research A**: "Oraculo DAG Engine and Orchestration Mechanics" -- survey-style analysis drawing on LangGraph, Temporal, Prefect, Airflow, Routine, and DBR/Kanban literature.
- **Research B**: "Architectural Blueprint: DAG Engine and Orchestration Mechanics for Oraculo" -- deep architectural narrative with implementation-level detail on each subsystem.
- **Prompt**: 01-dag-engine-and-orchestration.md (6 guiding questions)
- **Alignment docs**: docs/philosophy.md, docs/design.md

---

## 1. Graph Representation

**Guiding question:** How should the DAG be represented for an LLM orchestrator? What structure lets the LLM reason about the graph while the CLI validates it deterministically?

### Consensus

Both sources converge on a **dual-representation** model:

| Concern | LLM-facing | CLI / Storage |
|---|---|---|
| Format | JSON with explicit `nodes[]` and `edges[]` arrays | Normalized relational tables in SQLite (`dag_nodes`, `dag_edges`) |
| Purpose | Semantic reasoning, plan emission | Deterministic validation, fast querying |
| Authority | Proposes | Enforces |

Both agree the LLM must never directly mutate the persisted graph. It emits a JSON structure; the CLI validates and commits.

### Complementary insights

- **Research A** emphasizes the **node vs. run separation** (borrowed from Prefect/Airflow): `dag_nodes` are static definitions; `node_runs` are runtime executions with their own state machine. This prevents the dispatch engine from editing node definitions and keeps the schema clean for re-runs and retries.
- **Research B** adds two important node-level properties absent from Research A:
  - **Isolation parameters** (`isolation: worktree`) embedded in each node definition so the CLI knows how to spawn the sub-agent.
  - **Input/output schema validation**: edges should be rejected if the source node's output schema does not satisfy the target node's input requirements. This catches hallucinated dependencies at commit time.
- **Research A** references LangGraph's `Command` objects as a way to express runtime control flow (skip, redirect) without modifying the static graph, a useful distinction between *structural mutations* (new nodes/edges) and *runtime commands* (annotations on node runs).

### Conflicts

- **Coordinate-based IDs vs. opaque IDs**: Research B proposes hierarchical string coordinates (e.g., `1.0`, `1.1.a`) as primary keys to allow mid-sequence insertion without re-indexing. Research A instead recommends opaque numeric IDs in SQLite with human-readable coordinate *labels* for the LLM. The Research A approach is safer -- using semantic strings as primary keys couples the DB schema to the LLM's naming convention and makes foreign-key management fragile. The coordinate label can be a non-unique display field.

### Synthesized recommendation

Adopt the **dual-representation** with node/run separation. Each `dag_node` carries an opaque `node_id` (integer PK), a `human_label` string for LLM-facing coordinates, `type` (code, qa, research, documentation), `squad_id`, `phase`, and an `isolation` field. A strict JSON schema governs LLM I/O. The CLI validates acyclicity (Kahn's algorithm or DFS), edge-target existence, and input/output schema compatibility before committing.

---

## 2. Dispatch Algorithm

**Guiding question:** How to implement the "dispatch all nodes with in-degree zero" loop? How to handle simultaneous unblocking vs. squad size limits?

### Consensus

Both sources agree on a **frontier-based dispatch loop** owned entirely by the Go CLI:

1. Query `node_runs` in PENDING state whose upstream dependencies are all in terminal states (COMPLETED or PRUNED).
2. Filter by squad capacity and per-stage WIP limits.
3. Transition selected runs to RUNNING in a single transaction.
4. Spawn Claude Code sub-agents with worktree isolation.

Both sources agree the LLM orchestrator is never polled for task readiness -- the CLI handles the high-frequency loop autonomously.

### Complementary insights

- **Research B** introduces a **critical-path priority heuristic** when ready nodes exceed squad capacity: `pri(v) = (critical_path_length, fan_out)`. This ensures the CLI dispatches nodes that unblock the longest dependency chains first, rather than using naive FIFO. Research A mentions squad-as-resource-pool but does not specify a priority algorithm.
- **Research A** provides the SQL-transaction-level detail: the dispatch cycle recomputes per-squad and per-stage WIP, queries the frontier, applies constraints, and marks runs as RUNNING -- all within a single SQLite transaction for atomicity.
- **Research B** specifies the **state machine** explicitly: `PENDING -> RUNNING -> COMPLETED | FAILED`. A FAILED node does *not* decrement the in-degree of its successors, freezing that branch until the orchestrator intervenes. Research A mentions the same via Prefect's state model but adds CANCELLED and CRASHED as additional terminal states worth considering.

### Conflicts

None significant. The sources are complementary.

### Synthesized recommendation

The CLI dispatch loop runs autonomously on a polling interval. It queries the frontier, scores candidates using `(critical_path_length, fan_out)`, filters by squad capacity and WIP limits, and dispatches in a single transaction. The node-run state machine should include at minimum: `PENDING`, `RUNNING`, `COMPLETED`, `FAILED`, `PRUNED`, `CANCELLED`. Every state transition generates an append-only event.

---

## 3. Dynamic Mutation

**Guiding question:** How to implement branch pruning and node insertion at runtime without corrupting the graph? What invariants must be preserved?

### Consensus

Both sources agree on core invariants:

1. **Never mutate RUNNING or COMPLETED nodes.** Pruning can only target PENDING nodes.
2. **Pruned nodes are logically marked, not physically deleted.** This preserves auditability and prevents orphaned sub-agents.
3. **All mutations are wrapped in a single atomic SQLite transaction.** No partial graph rewrites.
4. **The LLM proposes mutations via a structured tool call** (Research B calls it `MutateGraph`); the CLI validates and applies.

### Complementary insights

- **Research B** provides the most detailed mutation mechanics: upon failure, the CLI computes the **transitive closure** of all downstream nodes that depend exclusively on the failed path, prunes them, then grafts new nodes. It also introduces the `MutateGraph` JSON command schema as a formal tool interface.
- **Research A** adds **branch-completion semantics at merge points**: when a branch is pruned, the merge node's dependency on that branch is considered satisfied by the PRUNED status, allowing downstream execution to continue without executing the pruned work.
- **Research A** also distinguishes **structural mutations** (new nodes/edges via plan versioning) from **runtime commands** (skip, redirect as events on node runs), borrowing from LangGraph's `Command` pattern. This distinction reduces the need for heavy graph mutations when the adjustment is purely a runtime routing decision.

### Conflicts

- **Coordinate-based insertion** (Research B) vs. **opaque-ID insertion with plan versioning** (Research A). Research B argues coordinates eliminate cascading re-indexing; Research A uses plan versions to isolate mutations. The plan-versioning approach from Research A is more robust because it preserves referential integrity and avoids coupling the database schema to a naming convention. Coordinate labels can still be used for LLM communication.

### Synthesized recommendation

Define a `MutateGraph` tool schema that accepts: nodes to prune (by ID), new nodes to insert (with edges), and the plan version being mutated. The CLI validates acyclicity, computes the transitive closure of pruned paths, marks affected PENDING nodes as PRUNED, grafts new nodes, updates merge-point dependencies, and commits in a single transaction. Runtime routing adjustments (skip, redirect) should use lightweight `Command` events on node runs rather than full graph mutations.

---

## 4. Throughput Control (Drum-Buffer-Rope)

**Guiding question:** How to implement the feedback loop where QA throughput throttles dispatch? What signals does the orchestrator monitor? How to detect bottleneck shifts?

### Consensus

Both sources map DBR to Oraculo identically:

| DBR concept | Oraculo mapping |
|---|---|
| Drum | QA process (bottleneck that sets system pace) |
| Buffer | Queue of completed code-generation tasks awaiting QA |
| Rope | Feedback mechanism that throttles upstream dispatch when buffer is full |

Both agree that when the QA buffer exceeds a configured threshold, the CLI stops dispatching new code-generation nodes even if they have in-degree zero.

### Complementary insights

- **Research B** adds significant operational detail:
  - **Buffer sizing** via Little's Law and queuing theory: buffer must absorb upstream variability without starving QA, but must not grow unbounded.
  - **Non-constraint exploitation**: when throttled, the CLI redirects idle agents to non-bottleneck tasks (documentation, research, dependency auditing) rather than leaving them idle.
  - **Shifting bottleneck detection**: rolling flow-time metrics per node type allow the CLI to detect when the bottleneck moves (e.g., from QA to a long-running research task) and dynamically reassign the Drum.
  - Introduces a `THROTTLED` state in the dispatcher that overrides standard readiness checks.
- **Research A** provides the configuration-data perspective: per-stage and per-squad WIP limits are stored in SQLite tables (`stage_limits`, `squads`), making them queryable and adjustable without code changes.

### Conflicts

None. Research B provides the "how" while Research A provides the "where to store it."

### Synthesized recommendation

Implement DBR as follows:

1. **WIP limits** stored in `stage_limits(stage_name, wip_limit)` and `squads(squad_id, max_running)` tables.
2. **Buffer metric**: count of nodes in COMPLETED state with an edge to a PENDING QA node. When this exceeds `stage_limits['awaiting_qa'].wip_limit`, the rope goes taut.
3. **Throttling**: the dispatch loop skips code-generation nodes when the rope is taut, but continues dispatching non-bottleneck tasks (documentation, research).
4. **Bottleneck detection**: the CLI computes rolling average completion times per `node.type`. If a non-QA type's average significantly exceeds QA's, the Drum shifts and the rope redirects accordingly.
5. **Node type tagging**: each `dag_node` must carry a `type` field (code, qa, research, documentation) so the dispatch loop can distinguish bottleneck-bound from free-flow work.

---

## 5. Persistence and Recovery

**Guiding question:** How to persist DAG state in SQLite so the orchestrator can reconstruct exact state and resume after a crash?

### Consensus

Both sources agree on **event sourcing with materialized views**:

- An append-only `events` table is the single source of truth.
- Materialized tables (`dag_nodes`, `dag_edges`, `node_runs`) are projections derived from the event stream for fast querying.
- Recovery replays events from the last snapshot to reconstruct exact state.

### Complementary insights

- **Research B** provides deeper implementation detail:
  - **CQRS pattern**: separate the write path (event appending) from the read path (materialized snapshot queries). The CLI generates consistent snapshots periodically; the dispatch loop queries snapshots for O(1) lookups.
  - **WAL mode** (`PRAGMA journal_mode = WAL`): essential for concurrent reads/writes when multiple sub-agents report status simultaneously.
  - **Monotonically increasing version integers** on `(aggregate_id, version)` with unique constraints to prevent out-of-order event application.
- **Research A** provides the concrete schema sketch with table definitions that Research B describes conceptually but does not tabulate:
  - `journeys`, `dag_nodes`, `dag_edges`, `node_runs`, `events`, `squads`, `stage_limits`.

### Conflicts

- **Full CQRS with periodic snapshots** (Research B) vs. **synchronous materialized tables** (Research A). Research B warns that replaying thousands of events per dispatch cycle is prohibitive and recommends periodic snapshots. Research A implies materialized tables are maintained synchronously alongside event insertion. The pragmatic choice is synchronous materialization: update the materialized tables within the same transaction that appends the event. This avoids snapshot staleness without the full complexity of periodic folding. True CQRS with asynchronous projection is unnecessary at Oraculo's scale (dozens to low hundreds of nodes per journey, not millions).

### Synthesized recommendation

**Schema** (combining both sources):

```sql
-- Core tables
journeys(journey_id, root_goal, plan_version_current, status, created_at)

dag_nodes(node_id INTEGER PK, journey_id, plan_version, human_label,
          type, squad_id, phase, isolation, input_schema_json,
          output_schema_json, is_terminal, created_at)

dag_edges(edge_id INTEGER PK, journey_id, plan_version,
          from_node_id, to_node_id, condition, active, created_at)

node_runs(node_run_id INTEGER PK, node_id, journey_id, status,
          result_ref, error_kind, retry_count,
          started_at, finished_at)

events(event_id INTEGER PK AUTOINCREMENT, journey_id,
       event_type TEXT, payload_json TEXT,
       node_id, node_run_id, plan_version,
       version INTEGER, created_at)

squads(squad_id, journey_id, max_running)
stage_limits(stage_name, wip_limit, journey_id)
```

- `PRAGMA journal_mode = WAL` enforced at initialization.
- Materialized tables updated synchronously within the same transaction as event insertion.
- `UNIQUE(journey_id, version)` on the events table to guarantee ordering.
- Recovery: on restart, the CLI reads the materialized tables directly. If corruption is suspected, it can rebuild by replaying the events table from the beginning.

---

## 6. Concurrent Re-Planning

**Guiding question:** How to re-plan downstream nodes while upstream nodes are still executing? How to avoid conflicts?

### Consensus

Both sources agree that re-planning must be non-blocking -- the orchestrator should not lock the graph while generating a new plan.

### Complementary insights

- **Research A** introduces **plan versioning**: re-planning creates a new `plan_version` with updated nodes and edges for downstream stages. Nodes already RUNNING or COMPLETED remain on their original version. Cutover happens at natural phase boundaries. It also identifies **re-planning triggers**: repeated failures, QA feedback requiring redesign, user-initiated scope changes.
- **Research B** introduces **Optimistic Concurrency Control (OCC)** with **ETAG validation**: when the orchestrator reads the graph to begin re-planning, the CLI provides a `version_id`. When the orchestrator submits its mutation, it includes this `version_id`. If the graph has changed in the interim, the CLI rejects the mutation and provides the updated state. This prevents stale-state corruption without pessimistic locking.
- **Research B** also adds **localized transaction boundaries**: version tracking is scoped to topological sub-graphs (e.g., per-phase or per-branch), so an unrelated change in a frontend branch does not invalidate a backend re-plan.

### Conflicts

- **Plan versioning** (Research A) vs. **OCC with ETAGs** (Research B). These are not mutually exclusive. Plan versioning provides the structural mechanism (new nodes/edges live under a new version number). OCC provides the concurrency-safety mechanism (reject stale mutations). They compose well: the `MutateGraph` command includes both the target `plan_version` and the `version_id` (ETAG). The CLI checks that the ETAG matches, then commits the new plan version.

### Synthesized recommendation

Combine both mechanisms:

1. **Plan versioning**: each re-plan creates a new `plan_version`. Downstream nodes under the new version coexist with upstream nodes under the old version. Cutover is at phase boundaries.
2. **OCC with ETAG**: the CLI provides a `version_id` (derived from the latest event sequence number) when the orchestrator reads the graph. The `MutateGraph` command must include this `version_id`. On mismatch, the CLI rejects and returns the current state.
3. **Localized scoping**: version tracking is per-branch or per-phase, not global. An unrelated sub-agent completion does not invalidate a concurrent re-plan on a different branch.
4. **Re-planning triggers** (stored as events in SQLite):
   - Repeated failures of a critical node (e.g., 2+ retries exhausted)
   - QA feedback requiring architectural changes
   - User-initiated scope changes
   - Discovery of new architectural context during execution

---

## Gaps

The following areas from the guiding questions were not adequately addressed by either source:

1. **Concrete Kahn's algorithm implementation**: Both sources mention acyclicity validation but neither provides pseudocode or specifies how the CLI should handle the validation in the context of incremental mutations (validating only the delta vs. re-validating the entire graph).

2. **Worktree merge protocol**: Research B mentions "tree-shaking and merge protocol" for synchronizing worktree changes back to main, but neither source details how merge conflicts between parallel agents are detected, reported, or resolved. This is a critical gap given that up to 5 agents may be mutating the same repository.

3. **Agent lifecycle management**: Neither source addresses how the CLI monitors spawned Claude Code sub-agents at the OS process level -- timeouts, heartbeats, zombie process detection, or what happens if an agent hangs indefinitely without reporting COMPLETED or FAILED.

4. **Schema migration**: Neither source discusses how the SQLite schema evolves as Oraculo matures. Event sourcing schemas are notoriously difficult to migrate. What happens when a new event type is introduced or an existing payload format changes?

5. **LLM context window management**: Neither source addresses how to present large graphs to the LLM orchestrator without exceeding context limits. Filtering, summarization, or progressive disclosure strategies for the JSON representation are not discussed.

6. **Concrete buffer sizing**: Research B mentions Little's Law for QA buffer sizing but does not provide a formula or heuristic. What is the initial default? How does it adapt?

---

## Key Design Decisions Summary

1. **Dual-representation model**: The LLM operates on a JSON schema with explicit `nodes[]` and `edges[]` arrays. The CLI persists and validates using normalized SQLite tables. The LLM proposes; the CLI enforces.

2. **Node/Run separation**: Static node definitions (`dag_nodes`) are distinct from runtime executions (`node_runs`). The dispatch engine operates on run state, never on node definitions. This enables retries, re-runs, and clean auditability.

3. **Opaque IDs with coordinate labels**: Primary keys are opaque integers. Human-readable coordinate labels (e.g., `1.2.b`) are stored as a display field for LLM communication. This preserves referential integrity while enabling intuitive branch/step references.

4. **CLI-owned dispatch loop**: The Go CLI autonomously computes the frontier, scores candidates by `(critical_path_length, fan_out)`, filters by squad capacity and WIP limits, and dispatches -- all without invoking the LLM. The orchestrator is only consulted for planning and re-planning.

5. **Structured mutation via `MutateGraph` tool**: All graph mutations from the LLM go through a typed tool schema. The CLI validates acyclicity, transitive-closure pruning of PENDING nodes, merge-point dependency satisfaction, and edge schema compatibility before committing in a single atomic transaction.

6. **Immutable pruning**: Pruned nodes are marked PRUNED, never deleted. Merge-point dependencies on pruned branches are treated as satisfied. RUNNING and COMPLETED nodes are immutable to pruning.

7. **DBR with stored WIP limits**: QA is the default Drum. Buffer size is tracked as count of COMPLETED nodes awaiting QA. WIP limits are stored in SQLite configuration tables. The dispatch loop throttles code-generation nodes when the buffer exceeds threshold, redirecting agents to non-bottleneck tasks.

8. **Shifting bottleneck detection**: Rolling average completion times per node type allow the CLI to detect when the constraint moves away from QA and dynamically reassign the Drum.

9. **Event-sourced persistence with synchronous materialization**: All state changes are appended to an immutable `events` table. Materialized tables (`dag_nodes`, `dag_edges`, `node_runs`) are updated within the same transaction. WAL mode is mandatory. Recovery reads materialized tables directly; full replay is the fallback for corruption.

10. **Plan versioning with OCC**: Re-planning creates new `plan_version` entries. The `MutateGraph` tool requires an ETAG (`version_id`). The CLI rejects stale mutations. Version tracking is scoped per-branch/phase to minimize false rejections.

11. **Lightweight runtime commands**: Not all control-flow adjustments require graph mutations. Skip, redirect, and short-circuit operations are expressed as `Command` events on `node_runs`, borrowing from LangGraph's pattern. This reduces mutation overhead for routine routing decisions.

12. **Node isolation declaration**: Each node definition includes an `isolation` field (e.g., `worktree`) so the CLI knows how to spawn and sandbox the corresponding sub-agent. This is part of the JSON schema, not an implicit convention.
