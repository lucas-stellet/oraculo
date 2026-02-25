# Synthesis: Memory System and Knowledge Pipeline

## Source Material

| # | Source | Scope |
|---|--------|-------|
| R1 | _research/agents/design/Architecture and Design of the Oraculo Memory System and Knowledge Pipeline.md | Primary research -- covers all 6 guiding questions |

**Cross-references consulted:**
- Resilient Agent Orchestration (communication/handoff doc) -- confirms blackboard-as-shared-state model where agents read/write SQLite via CLI
- Execution Environments and Isolation Boundaries -- reinforces MVI context injection and structural topological mapping as the basis for working set assembly

---

## 1. Evaluation Against Guiding Questions

### GQ1: SQLite Schema Design for Three-Tier Memory

**Coverage: Strong**

R1 provides concrete, implementation-ready schema designs for all three tiers:

- **Semantic tier** -- `semantic_knowledge` table with typed prefixed IDs (`sem-x7k2m`), temporal versioning (`valid_from`, `valid_to`), status enum (`proposed`/`validated`/`deprecated`), provenance linking via `provenance_run_id`, and self-referencing `supersedes_id`. A companion `semantic_knowledge_history` table managed by SQLite triggers handles archival automatically.
- **Episodic tier** -- `episodes` table using ULIDs for lexicographic ordering, with `agent_id`, `run_id`, `phase`, `event_type`, JSON `payload`, and UTC `timestamp`.
- **Working tier** -- Not stored as a table; assembled dynamically by the CLI at query time from scored combinations of semantic facts and recent episodic entries.
- **Relational linking** -- `semantic_edges` table with triplet schema (`subject_id`, `predicate`, `object_id`, `confidence`) enabling graph-like queries without a graph database.

**Gap:** The research does not address schema migration strategy. As the system evolves, how do tables change without corrupting the append-only episodic log? No `schema_version` table or migration tooling is discussed.

**Gap:** No explicit definition of the `semantic_knowledge_history` trigger SQL. The concept is described but the implementation is left abstract.

### GQ2: Event Sourcing Implementation

**Coverage: Strong**

R1 commits to a clear event sourcing model:

- Append-only `episodes` table as the single source of truth
- Compensating events rather than mutations for error correction
- CQRS separation: writes go to the event log; reads derive state from snapshots
- ULID-based primary keys for chronological ordering without timestamp dependency
- `active_runs_snapshot` table maintained by SQLite triggers for O(1) current-state lookups

**Compaction strategy:** Hybrid snapshotting -- the full archive remains for auditing while a compacted snapshot table provides fast lookups via trigger-maintained upserts. This is practical and well-aligned with SQLite's strengths.

**Gap:** No discussion of event stream partitioning for very long-lived projects (years of accumulated events). While the snapshot table handles read performance, the raw `episodes` table will grow unbounded. No archival or cold-storage strategy is proposed.

**Gap:** No explicit event versioning scheme. If the `payload` JSON schema evolves over time, how are older events interpreted? No `event_version` field or payload migration strategy.

### GQ3: Working Set Builder

**Coverage: Moderate-to-Strong**

R1 proposes a well-reasoned approach with three components:

1. **Composite FTS5 scoring** -- BM25 base score modified by exponential temporal decay: `score = BM25 * e^(-0.03 * days_old)`. Implemented directly in SQL `ORDER BY`.
2. **Line-limit truncation** -- `LIMIT` and `SUBSTR` at the database layer to enforce the <100 lines (concepts) and <150 lines (guides) budgets.
3. **Three-part payload** -- Rolling task summary + last k conversational turns + ranked semantic facts.

**Evaluation protocol:** "Needle in a Haystack" testing -- synthetic facts injected into deep records, daily benchmarks verify retrieval accuracy within the top 150 lines.

**Gap:** The decay constant `0.03` is stated without empirical justification. No discussion of how to tune this parameter or make it configurable per-project or per-domain.

**Gap:** No algorithm for determining the `k` in "last k conversational turns." The balance between recency context and semantic facts is left undefined.

**Gap:** The relationship between the working set builder and the MVI context injection described in the Execution Environments research is not reconciled. MVI injects structural/topological context (file maps, dependency graphs); the working set builder injects semantic/episodic context. How these two streams are merged into a single prompt is unspecified.

### GQ4: Curation Pipeline

**Coverage: Strong**

R1 describes a complete pipeline: Episode -> Score -> Reflection -> Validated Promotion -> Relational Linking.

- **Scoring:** Impact score based on task success, error severity, and temporal recurrence. Only episodes crossing a severity threshold trigger the expensive LLM reflection cycle (token conservation).
- **Reflection:** Reflexion framework -- a dedicated curation agent evaluates episodic trajectories, extracting root causes and behavioral rules. Operates asynchronously, outside the critical execution path.
- **Promotion:** Multi-Agent Reflexion (MAR) -- a jury of distinct LLM personas cross-evaluates proposed insights. Mean consensus score required. CLI enforces the contract (consensus score + provenance links mandatory).
- **Conflict handling:** New insights contradicting existing knowledge are written as `proposed` with a `conflicts_with` edge. No forced consensus. Human or senior orchestrator resolves.

**Gap:** No definition of the severity threshold formula. What constitutes "high enough impact" to trigger reflection? This is a critical tuning parameter left undefined.

**Gap:** No specification of the jury composition. How many agents? Which model variations? How is tie-breaking handled?

**Gap:** No scheduling or triggering mechanism. The research says "a scheduled background process managed by the CLI" but does not specify whether this is cron-based, event-driven (triggered after run completion), or on-demand.

### GQ5: Concurrency and Write Serialization

**Coverage: Strong**

R1 provides a complete, implementation-ready concurrency model:

- **WAL mode:** `PRAGMA journal_mode=WAL` + `PRAGMA synchronous=NORMAL` for non-blocking reads.
- **Dual connection pool:** Read pool with high limits; write pool restricted to `db.SetMaxOpenConns(1)`.
- **Busy timeout:** `PRAGMA busy_timeout=5000` as a secondary safety net.
- **CLI as single write gateway:** All schema validation and provenance verification happens in-memory before the transaction reaches SQLite.

This aligns perfectly with the philosophy's "writes serialized through CLI; reads parallel" mandate.

**Gap:** No discussion of write queue observability. When multiple agents queue writes, how does the orchestrator know about backpressure? No metrics, logging, or alerting for queue depth.

**Gap:** No discussion of transaction size guidelines. The research says "keep write transactions short" but provides no concrete boundaries (max rows per transaction, max payload size).

### GQ6: Search Strategy Evolution

**Coverage: Strong**

R1 provides a clear three-phase evolution path:

1. **Phase 1 (MVI):** FTS5/BM25 with temporal decay -- the working set builder described in GQ3.
2. **Phase 2 (Hybrid):** Add `sqlite-vec` extension with embedding BLOB column on `semantic_knowledge`. Reciprocal Rank Fusion (RRF) via CTEs: `RRF_Score = 1/(k + Rank_FTS) + 1/(k + Rank_Vector)`, k=60.
3. **Phase 3 (Graph):** `WITH RECURSIVE` CTEs against `semantic_edges` for breadth-first dependency traversal. Depth limits mandatory to prevent runaway queries.

All three phases operate within pure SQLite, maintaining the single-storage-engine constraint.

**Gap:** No guidance on when to transition between phases. What volume of semantic knowledge or what retrieval failure rate triggers the move from Phase 1 to Phase 2?

**Gap:** No discussion of embedding generation. Which model produces the vectors? At what dimensionality? When are embeddings computed -- at write time (CLI responsibility) or as a background batch job?

**Gap:** The `sqlite-vec` extension requires compilation and distribution. No discussion of how this integrates with the Go CLI build pipeline.

---

## 2. Cross-Reference with Philosophy and Design

### Alignment Assessment

| Principle | Research Position | Alignment |
|-----------|-------------------|-----------|
| Three distinct tiers, mixing corrupts all three | Enforced at schema level: separate tables, separate ID prefixes, no shared columns | Full alignment |
| CLI validates every write | Dual connection pool, in-memory validation before transaction | Full alignment |
| Contradictions stored as versioned proposals | `proposed` status + `conflicts_with` edge, no forced consensus | Full alignment |
| Only QA-validated knowledge promoted | MAR consensus jury + mandatory contract for semantic writes | Full alignment |
| Curation pipeline: episode -> ... -> relational linking | Complete pipeline described with all intermediate steps | Full alignment |
| SQLite as storage engine (WAL mode) | WAL + synchronous=NORMAL, single .db file | Full alignment |
| Event sourcing: append-only, immutable | Compensating events, no UPDATE/DELETE on episodes | Full alignment |
| Agents access memory exclusively through CLI | CLI as single write gateway; dual pool architecture | Full alignment |
| Orchestrate, never execute | Curation agent is a separate concern from orchestrator | Full alignment |
| Maximize parallelism | N readers in parallel; writes serialized but non-blocking for reads | Full alignment |

### Tension Points

1. **Markdown at the end vs. semantic memory.** The design doc states "Markdown only at the end" for completed features. The memory research positions semantic memory as the persistent knowledge store. These are complementary but their relationship needs explicit definition: semantic memory is the machine-readable operational truth; markdown is the human-readable summary. The CLI should generate the final markdown from semantic queries.

2. **CLAUDE.md as persistent context vs. working set builder.** The design doc mentions "CLAUDE.md / Memory" as persistent context for agents. The working set builder assembles context dynamically from SQLite. These must be reconciled: CLAUDE.md provides static project-level context (conventions, patterns); the working set builder provides dynamic task-level context (relevant facts, recent episodes). The prompt structure should layer these explicitly.

3. **Documentation as Project Memory (design.md section 2) vs. three-tier architecture.** The design doc describes SQLite as storing "proposed ideas, decisions, requirements, execution plans, QA results, agent logs." This maps cleanly to the three tiers: ideas/decisions/requirements are semantic, plans/logs are episodic, and the active view is working memory. But the mapping is implicit and should be made explicit in the final design.

---

## 3. What Is Well-Covered vs. Gaps Requiring Investigation

### Well-Covered

- **Schema architecture** -- Production-ready table definitions for all persistent tiers
- **Event sourcing mechanics** -- Append-only discipline, ULID ordering, compensating events, snapshot tables
- **Write serialization** -- Complete Go-level implementation pattern with dual pools and busy timeout
- **Search evolution roadmap** -- Three-phase plan staying within SQLite, with RRF as the hybrid fusion method
- **Curation pipeline philosophy** -- Full Episode-to-Wisdom flow with reflection, scoring, and consensus validation
- **Conflict handling** -- Versioned proposals with relational edges, no premature consensus

### Gaps Requiring Additional Investigation

| Gap | Priority | Reason |
|-----|----------|--------|
| Schema migration strategy | High | Without this, the system cannot evolve safely over time |
| Event payload versioning | High | JSON schema drift in long-lived projects will cause deserialization failures |
| Working set + MVI context merge | High | Two independent context streams (semantic/episodic from memory, structural/topological from execution env) must be unified into a single prompt |
| Curation trigger mechanism | Medium | Scheduling and triggering curation is an operational concern that affects token spend |
| Severity threshold formula | Medium | Without a concrete formula, the curation pipeline cannot be implemented |
| Decay constant tuning | Medium | The 0.03 constant needs empirical validation or configurability |
| Embedding generation pipeline | Medium | Required for Phase 2 search but unspecified |
| sqlite-vec build integration | Medium | Distribution and compilation concern for the Go CLI |
| Write queue observability | Low | Important for production but not blocking for MVI |
| Event stream archival for very long projects | Low | Only relevant at significant scale |

---

## 4. Actionable Recommendations

### R1: Adopt the proposed schema with additions

Use the `semantic_knowledge`, `semantic_knowledge_history`, `episodes`, `semantic_edges`, and `active_runs_snapshot` tables as designed. Add the following:

```sql
-- Schema versioning
CREATE TABLE schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  DATETIME DEFAULT (datetime('now')),
    description TEXT NOT NULL
);

-- Event payload versioning
ALTER TABLE episodes ADD COLUMN event_version INTEGER DEFAULT 1;
```

### R2: Implement the composite scoring as a Go function, not raw SQL

The exponential decay calculation (`e^(-0.03 * days_old)`) should be implemented as a Go function that constructs the SQL query with parameterized decay constants. This allows per-project tuning:

```go
type ScoringConfig struct {
    DecayRate      float64 // default 0.03
    RecencyWeight  float64 // weight for time-decay component
    RelevanceWeight float64 // weight for BM25 component
}
```

### R3: Define explicit prompt structure for context assembly

The working set builder output should follow a deterministic structure:

```
[SYSTEM CONTEXT]        -- from CLAUDE.md (static, project-level)
[TASK CONTEXT]          -- rolling summary of current task
[STRUCTURAL CONTEXT]    -- from MVI/topological mapping (files, dependencies)
[SEMANTIC CONTEXT]      -- ranked facts from working set builder
[EPISODIC CONTEXT]      -- last k turns, recent relevant episodes
[INSTRUCTIONS]          -- the actual task directive
```

This merges the memory system's output with the execution environment's context injection.

### R4: Implement curation as event-driven, not scheduled

Rather than a cron-based background process, trigger curation when a `run_id` completes (a terminal event like `TaskCompleted` or `TaskFailed` is appended). The CLI detects this event and enqueues a curation job. This is more responsive and avoids polling.

### R5: Define the severity threshold as a configurable composite score

```go
type EpisodeScore struct {
    ImpactLevel    int     // 1-5: how much did this affect outcomes
    ErrorSeverity  int     // 0-5: 0=success, 5=critical failure
    RecurrenceCount int    // how many times this pattern appeared
    Composite      float64 // weighted sum, threshold = 3.0 by default
}
```

### R6: Start with 3-agent jury for MAR consensus

For the curation pipeline's multi-agent consensus:
- 3 evaluator agents with distinct system prompts (skeptic, integrator, domain expert)
- Majority vote (2/3) required for promotion to `validated`
- All scores and reasoning traces stored as episodic events for auditability

### R7: Defer sqlite-vec to Phase 2, design the column now

Add the embedding column to the schema from day one, but leave it nullable. This avoids a schema migration when Phase 2 begins:

```sql
ALTER TABLE semantic_knowledge ADD COLUMN embedding BLOB DEFAULT NULL;
```

Embedding generation can be a background CLI command (`oraculo memory embed --backfill`) that populates the column incrementally.

---

## 5. Key Design Decisions Summary

1. **Three-tier schema with typed ID prefixes** -- `sem-`, `edg-`, and ULID for episodes. Each tier has dedicated tables with no cross-tier mutations. Provenance links connect tiers via foreign keys.

2. **Append-only episodic log with compensating events** -- No UPDATE or DELETE on the `episodes` table. Errors are corrected by appending compensating events. State is derived from snapshots, not replays.

3. **Trigger-maintained `active_runs_snapshot`** -- SQLite triggers on the `episodes` table automatically upsert the latest state per `run_id`, providing O(1) current-state lookups without full event replay.

4. **Composite FTS5 scoring with configurable temporal decay** -- BM25 base score modified by `e^(-decay_rate * days_old)`, where `decay_rate` is configurable per project. Implemented as parameterized SQL constructed by the Go CLI.

5. **Strict line-limit truncation at the database layer** -- `LIMIT` and `SUBSTR` in SQL enforce the <100/<150 line budgets. No oversized payloads ever reach the application layer.

6. **Six-zone prompt structure for context assembly** -- System context, task context, structural context (MVI), semantic context (memory), episodic context (recent turns), and instructions. Deterministic ordering exploits the LLM's U-shaped attention curve.

7. **Event-driven curation with severity threshold gating** -- Curation triggers on run completion, not on a schedule. Only episodes exceeding a configurable composite severity score invoke the expensive LLM reflection cycle.

8. **Multi-Agent Reflexion (MAR) with 3-agent jury** -- Three evaluator personas cross-validate proposed knowledge. Majority consensus required. No promotion without provenance links and consensus score in the write contract.

9. **Versioned conflict resolution, never forced consensus** -- Contradictory knowledge enters as `proposed` with `conflicts_with` edges. Deprecation requires explicit human or orchestrator decision, updating `valid_to` on the old entry.

10. **Dual connection pool with single-writer serialization** -- Read pool with high concurrency; write pool restricted to `MaxOpenConns(1)`. WAL mode + `busy_timeout=5000ms`. All schema validation and provenance checks happen in-memory before the transaction.

11. **Three-phase search evolution within pure SQLite** -- Phase 1: FTS5/BM25 with decay. Phase 2: Add sqlite-vec embeddings + RRF fusion via CTEs. Phase 3: `WITH RECURSIVE` graph traversal on `semantic_edges`. Each phase is additive; no rip-and-replace.

12. **Schema migration table from day one** -- `schema_migrations` table tracks all structural changes. Event payloads carry an `event_version` field for forward-compatible deserialization.
