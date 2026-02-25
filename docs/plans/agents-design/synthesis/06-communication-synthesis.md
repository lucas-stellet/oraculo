# Synthesis: Communication, Handoff, and Failure Recovery

## Sources

- **Research A**: "Resilient Communication and Failure Recovery Design for Oraculo's Deterministic Blackboard Architecture" (shorter, pattern-focused)
- **Research B**: "Resilient Agent Orchestration: Communication, Handoff, and Failure Recovery in the Oraculo Architecture" (longer, architecturally detailed)

---

## Cross-Source Agreement and Divergence

### Strong Consensus

Both sources agree on the following foundational points:

1. **CLI as validation gateway** -- All writes to the SQLite blackboard pass through CLI enforcement. The CLI rejects malformed payloads before they touch the database.
2. **JSON Schema for handoff contracts** -- Handoffs are structured JSON documents with mandatory fields (task_id, source_agent, target_agent, payload, expected_output_schema, constraints).
3. **RBAC enforced through CLI + SQLite** -- Since SQLite lacks native RBAC, governance is implemented via application-level logic (CLI middleware, SQLite triggers, policy tables).
4. **Three-state circuit breaker** -- Both adopt the CLOSED/OPEN/HALF-OPEN model from distributed systems (Hystrix, Resilience4j).
5. **Application-level checkpointing over raw SQLite rollback** -- Both explicitly reject relying solely on SQLite SAVEPOINT/ROLLBACK for recovery in parallel DAGs, since database-level rollback would destroy unrelated parallel branch progress.
6. **Context boundaries at phase transitions** -- Both advocate severing the context stream between phases, passing only structured artifacts and summaries rather than raw conversation histories.

### Complementary Insights

| Aspect | Research A contributes | Research B contributes |
|--------|----------------------|----------------------|
| Handoff schema | Comparison table of validation approaches (static, dynamic, embedded, hybrid) | A2A protocol ontology, Beads pattern, idempotency keys, auto-repair loops |
| Governance | ABAC as advanced option, OpenFGA mention | SQLite trigger implementation details (RAISE(FAIL)), read-optimized views, tool permission isolation per role |
| SOPs | Parameterized templates stored as JSON/SQL, Jinja rendering | YAML-driven DAGs, sub-DAG invocation, revision cycles as appended nodes (preserving acyclic properties) |
| Circuit breakers | Comparison table of strategies (time-based, count-based, manual, automatic) | Debugger agent pattern, exponential backoff formula, deterministic exit conditions, fallback chains |
| Checkpointing | SQLite WAL as checkpointing mechanism, incremental vs. full snapshots | Compensating transactions (Sagas pattern), append-only ledger, branch-scoped checkpoints |
| Context propagation | W3C Trace Context, OpenTelemetry baggage | Verbatim vs. summarized distinction, dedicated Summarizer agent using a lightweight model, map-reduce summarization |

### Conflicts

No direct contradictions exist. The differences are in depth and emphasis:

- Research A suggests SQLite WAL as a checkpointing mechanism; Research B treats WAL as a concurrency enabler (for read parallelism) but explicitly warns against relying on database-level rollback, preferring application-level logical checkpoints. **Resolution**: Both are correct at different layers. WAL enables concurrent reads; application-level checkpoints enable branch-isolated rollback.
- Research A proposes the orchestrator, a dedicated summarizer, or the CLI as possible summarization owners; Research B firmly assigns this to a dedicated Summarizer agent using a fast, cheap model. **Resolution**: Research B's approach is more concrete and aligned with Oraculo's "orchestrate, never execute" philosophy -- the orchestrator delegates summarization to a specialized agent.

---

## Guiding Question 1: Handoff Contract Schema

### Synthesized Answer

The handoff contract is a JSON document written to a `handoff_contracts` table in SQLite via the CLI. The schema must enforce the following mandatory fields:

| Field | Type | Purpose |
|-------|------|---------|
| `trace_id` | UUID | Cross-hop lineage tracking (from Research B's A2A protocol adaptation) |
| `task_id` | string | Binds the handoff to a specific DAG node |
| `source_agent` | string | Provenance and auditability |
| `target_agent` | string | Routing |
| `payload` | object | Operational data: summarized context, instructions, file pointers |
| `expected_output_schema` | JSON Schema | Exact structure the target agent must produce |
| `constraints` | object | Runtime bounds: timeout, max retries, token budget, negative constraints |
| `idempotency_key` | string | Prevents duplicate work during retries (enforced via SQLite UNIQUE constraint) |
| `schema_version` | string | Compatibility management |

Optional fields: `priority`, `dependencies`, `metadata`.

**CLI validation**: Before writing to SQLite, the CLI validates the JSON payload against the registered schema. On validation failure, the CLI returns a structured error message to the source agent, triggering an auto-repair loop (Research B). The source agent corrects its output without the malformed payload ever reaching the blackboard.

**Schema versioning**: Schema definitions are stored as immutable records in a `schema_registry` table. The CLI implements upcasting logic to translate older payloads into current formats where backward-compatible, or explicitly rejects payloads using unsupported versions. This isolates evolution logic within the CLI.

### Recommendation

Adopt the hybrid validation approach (from Research A's comparison table): CLI pre-write validation using static JSON Schema for fast rejection, combined with an embedded `$schema` field in payloads for runtime self-description as the system evolves. The `schema_registry` table (from Research B) provides the versioning backbone.

---

## Guiding Question 2: Blackboard Governance Model

### Synthesized Answer

Governance is implemented through three reinforcing layers:

**Layer 1 -- Policy Storage**: A `security_policies` table maps agent roles to permitted operations on specific tables and phases. Role hierarchy:

| Role | Write Access | Read Access |
|------|-------------|-------------|
| Orchestrator | `task_assignments`, `dag_nodes`, `workflow_state`, `checkpoints` | All tables |
| Implementation Agent | Own phase rows in `artifacts`, `implementation_results` | Requirements, architecture docs, own task assignments |
| QA Agent | `validation_verdicts`, `approval_gates` | All implementation artifacts, test results |
| Summarizer Agent | `context_summaries` | Upstream phase artifacts and logs |

**Layer 2 -- Database Enforcement**: SQLite `BEFORE INSERT`, `BEFORE UPDATE`, and `BEFORE DELETE` triggers cross-reference the session identifier against `security_policies`. Unauthorized writes trigger `RAISE(FAIL, 'insufficient privileges')`, aborting the transaction deterministically.

**Layer 3 -- CLI Middleware**: The CLI authenticates the agent's persistent identifier upon connection, injects it into a session context, and restricts:
- Which tables/views the agent can query (read-optimized views filter unauthorized data)
- Which tools the agent can invoke (tool permission isolation by role)

**Read parallelism**: SQLite in WAL mode allows multiple agents to read concurrently without blocking writes. Agents query role-scoped views, not raw tables.

### Recommendation

Start with RBAC via CLI middleware + SQLite triggers (both sources agree this is the right balance). Defer ABAC/OpenFGA complexity until the system requires fine-grained attribute-based policies. Implement tool permission isolation from day one -- a Discover-phase agent must not access build tools, and an Implementer must not access external search tools.

---

## Guiding Question 3: SOP Templates for Common Workflows

### Synthesized Answer

SOPs are encoded as YAML-driven DAGs (Research B's recommendation, more concrete than Research A's JSON/SQL approach). Each template defines:

- **Nodes**: Phases with required agent roles, system prompts, and tool access lists
- **Edges**: Directed dependencies with handoff schema requirements
- **Gates**: Boolean approval checkpoints where QA or the orchestrator must assert passage
- **Failure hooks**: Circuit breaker thresholds and fallback chains per node

**Standard Development Pipeline** (synthesized from both sources):

```
Epic (Analyst) --> [Approval Gate] --> PRD (Product Manager) --> [Approval Gate] -->
Architecture (Architect) --> [Approval Gate] -->
  |-- Test Suite (Test-Author, TDD) --|
  |-- Implementation (Implementer)  --|
  --> [Convergence Gate] --> QA Validation --> [Approval Gate] --> Merge/Deliver
```

**Revision cycles**: When QA rejects, the orchestrator does not traverse backward along existing edges (which would violate DAG acyclicity). Instead, it appends new `Revision --> Review` nodes to the graph, preserving the complete append-only iteration history.

**Parameterization**: The CLI parses YAML templates and substitutes project-specific variables (repository paths, testing frameworks, language requirements, domain constraints) before populating `dag_nodes` and `dag_edges` tables.

**Sub-DAG invocation**: Complex nodes (e.g., "Implement") can trigger nested SOP sub-graphs (e.g., a TDD red-green-refactor loop), enabling composition of tested primitives.

### Recommendation

Define SOPs in YAML stored within the project. The orchestrator is a generic DAG traversal engine, not a hardcoded pipeline. Treat the YAML templates as the single source of truth for workflow definitions. Support sub-DAG invocation for composability.

---

## Guiding Question 4: Circuit Breaker Implementation

### Synthesized Answer

**Monitoring**: The CLI tracks per-agent metrics in a `telemetry_metrics` table:
- Schema validation pass/fail
- Tool execution success/failure
- API response latency
- Token consumption per interaction

**Thresholds** (synthesized from both sources):

| Trigger | Threshold | Source |
|---------|-----------|--------|
| Consecutive validation failures | 5 | Research B |
| Error rate over rolling window | 15% within 10-minute epoch | Research B |
| Tool execution latency | > 120 seconds | Research B |
| Consecutive failures (simpler) | 3 | Research A |

**State machine**: CLOSED --> OPEN --> HALF-OPEN --> CLOSED/OPEN

1. **CLOSED** (normal): Metrics tracked, all tasks flow normally
2. **OPEN** (tripped): Orchestrator stops assigning tasks to the failing agent/branch. Recovery sequence initiates.
3. **HALF-OPEN** (testing): After cooldown, one controlled retry with the Debugger agent's refined prompt. Success resets to CLOSED; failure reverts to OPEN with exponential backoff.

**Recovery sequence** (merged from both sources):

1. Log failure event: exact state, memory context, failure sequence to `failure_diagnostics` table
2. Deploy Debugger agent: a specialized diagnostic supervisor that analyzes violated constraints, recent telemetry, and stack traces
3. Debugger synthesizes a targeted remediation prompt for the worker agent
4. After cooldown, attempt single retry with refined prompt (HALF-OPEN)
5. If retry succeeds and passes validation gates: reset to CLOSED
6. If retry fails: revert to OPEN, apply exponential backoff ($T_{wait} = T_{base} \times 2^{failures}$)
7. After max retries exceeded (hardcoded in YAML template): escalate to fallback chain or human intervention

**Deterministic exit conditions**: Maximum retry limits are defined in the SOP YAML template per node. This prevents infinite retry loops and forces escalation.

### Recommendation

Implement the three-state machine natively in SQLite with time-based queries for OPEN-to-HALF-OPEN transitions. Use the Debugger agent pattern (Research B) rather than simple retry-with-same-prompt. Hardcode max retries per DAG node in the YAML template. Abstract all retry and backoff logic into the CLI layer -- agents operate under the illusion of uninterrupted execution.

---

## Guiding Question 5: Coordinated Checkpointing

### Synthesized Answer

**What a checkpoint captures** (merged):

| Component | Storage | Format |
|-----------|---------|--------|
| DAG node states | `checkpoints` table | Compressed JSON |
| Output artifact references | `checkpoints` table | Pointers to artifact rows |
| Accumulated context/memory | `checkpoints` table | Summarized JSON |
| Schema versions in use | `checkpoints` table | Version identifiers |
| Trace lineage | `checkpoints` table | `trace_id` + `node_id` scope |

**When checkpoints occur**: At the completion of every major phase (Discover, Plan, Execute, Validate, Deliver). Phase boundaries act as global synchronization barriers -- a subsequent phase does not initiate until all prerequisite parallel branches have checkpointed.

**Rollback mechanism**: The blackboard operates as an append-only ledger. Rollback does NOT use SQLite ROLLBACK or DELETE. Instead:

1. Orchestrator detects rollback need (circuit breaker exhaustion, QA fundamental rejection, mathematical unviability of current DAG path)
2. Orchestrator identifies the most recent valid checkpoint for the specific failing branch (scoped by `trace_id` and `node_id`)
3. Orchestrator writes a compensating transaction: a new event indicating the state transition, updates DAG pointers to reference the checkpoint's context
4. Corrupted downstream artifacts are marked as `deprecated` (not deleted)
5. Parallel branches remain entirely unaffected
6. Failed branch restarts from the pristine checkpointed state

**Rollback vs. retry decision**: A retry is appropriate when the failure is transient (network timeout, rate limit, minor prompt issue). A rollback is needed when:
- The circuit breaker has exhausted all retries
- QA fundamentally rejects the approach (not just the implementation)
- The DAG path is mathematically unviable given accumulated constraints

### Recommendation

Implement application-level logical checkpoints (not database-level SAVEPOINTs). Use the append-only ledger pattern with compensating transactions (Sagas) for rollback. Scope checkpoints strictly to `trace_id` + `node_id` for branch isolation. Leverage SQLite WAL for read concurrency, but do not conflate WAL checkpointing with application-level workflow checkpointing.

---

## Guiding Question 6: Cross-Phase Context Propagation

### Synthesized Answer

Context propagation follows a strict two-category model:

**Category 1 -- Verbatim artifacts** (MUST NOT be summarized):
- Technical specifications
- Required function signatures
- Code diffs and patches
- API schemas (OpenAPI, JSON Schema)
- Explicit user constraints
- Test results and exact error messages

**Category 2 -- Summarized procedural history** (CAN be lossily compressed):
- How the upstream agent arrived at conclusions
- Dead-ends explored
- Conversational history with the user
- Intermediate reasoning steps

**Summarization responsibility**: A dedicated Summarizer agent, using a fast/cheap model (e.g., Claude Haiku), applies map-reduce summarization over upstream logs before writing the summary to the blackboard. The orchestrator delegates this -- it does not summarize itself (consistent with "orchestrate, never execute").

**Context assembly for downstream agents**: The orchestrator constructs the initialization prompt by:
1. Loading the agent's role-specific system prompt
2. Injecting verbatim artifacts from prerequisite DAG nodes (loaded from blackboard via template variables)
3. Injecting the dense context summary from the Summarizer
4. Providing relevant codebase context from CLAUDE.md / memory

**Session termination**: When an agent completes a phase, it writes its structured output artifact to the blackboard. Its session is then completely terminated. The downstream agent starts with a fresh context window. No raw chat history crosses phase boundaries.

**Traceability**: Adopt W3C Trace Context headers (`traceparent`, `tracestate`) embedded in handoff contracts for end-to-end distributed tracing and observability.

### Recommendation

Implement the verbatim-vs-summarized distinction as a first-class concept in the handoff schema. Deploy a Summarizer agent for procedural history compression at phase transitions. Terminate agent sessions completely between phases -- context-engineered boundaries prevent context rot. Use the `trace_id` from handoff contracts for end-to-end traceability.

---

## Gaps and Open Questions

### Adequately Addressed
All six guiding questions received substantive treatment from at least one source. No question was left without actionable guidance.

### Gaps Requiring Further Work

1. **Concrete YAML SOP schema**: Both sources describe the concept of YAML-driven DAGs, but neither provides a complete, validated YAML schema definition. A concrete schema with field types, required fields, and examples is needed before implementation.

2. **Schema registry migration tooling**: Both sources mention schema versioning and upcasting, but neither specifies how to handle breaking schema changes that cannot be upcast. A migration strategy (similar to database migrations) should be designed.

3. **Debugger agent prompt engineering**: Research B introduces the Debugger agent as a diagnostic supervisor, but does not detail how to construct the diagnostic prompt or what specific telemetry data it should analyze. The Debugger agent's SOP/persona needs its own design.

4. **Checkpoint storage size management**: Neither source addresses how to manage the growth of the `checkpoints` table over time, especially for long-running projects. Retention policies, archival strategies, or checkpoint pruning need design.

5. **Human-in-the-loop escalation protocol**: Both sources mention escalation to human intervention as a fallback, but neither defines the UX for this -- how the human is notified, what information they receive, and how they re-inject decisions into the blackboard.

6. **Network partition / crash recovery**: The sources assume a single-machine SQLite setup. If Oraculo ever needs to support distributed execution (multiple machines), the checkpoint and governance models would need significant rework. This is not an immediate concern but should be flagged.

7. **Metrics retention and dashboarding**: Research B mentions a `telemetry_metrics` table but does not address how to expose these metrics for human monitoring. The spec-workflow-mcp reference project's dashboard concept was not deeply explored.

8. **Conflict resolution for concurrent writes**: While WAL enables concurrent reads, the governance model assumes serialized writes through the CLI. Neither source addresses what happens if two agents attempt to write to the same logical resource simultaneously (e.g., both claim to have completed the same task).

---

## Key Design Decisions Summary

1. **Handoff contracts are JSON documents with mandatory fields** (`trace_id`, `task_id`, `source_agent`, `target_agent`, `payload`, `expected_output_schema`, `constraints`, `idempotency_key`, `schema_version`) written to a `handoff_contracts` table and validated by the CLI before insertion.

2. **The CLI is the sole validation gateway** -- it rejects malformed payloads with structured error messages, triggering auto-repair loops in source agents. No invalid data enters the blackboard.

3. **Schema versioning uses an immutable `schema_registry` table** -- the CLI implements upcasting for backward-compatible evolution and hard rejection for unsupported versions.

4. **Governance is RBAC via three layers**: policy table in SQLite, BEFORE triggers with RAISE(FAIL) for database-level enforcement, and CLI middleware for application-level enforcement including tool permission isolation.

5. **SQLite operates in WAL mode** for concurrent read parallelism. Agents query role-scoped views, never raw tables.

6. **SOPs are YAML-driven DAG templates** parameterized per project and stored as configuration. The orchestrator is a generic DAG traversal engine, not a hardcoded pipeline.

7. **Revision cycles append new nodes** to the DAG rather than traversing backward, preserving acyclic properties and full iteration history.

8. **Circuit breakers use a three-state machine** (CLOSED/OPEN/HALF-OPEN) tracked in SQLite, with configurable thresholds per DAG node defined in the YAML template.

9. **The Debugger agent is a specialized diagnostic supervisor** deployed on circuit breaker trip to analyze failures and generate remediation prompts. It uses a separate, highly capable model for root-cause analysis.

10. **Deterministic exit conditions** (max retries per node, hardcoded in YAML) prevent infinite retry loops and force escalation to fallback chains or human intervention.

11. **Checkpoints are application-level logical snapshots** stored as compressed JSON in a `checkpoints` table, scoped to `trace_id` + `node_id`. They are NOT SQLite SAVEPOINTs.

12. **Rollback uses compensating transactions** (Saga pattern) on an append-only ledger. Failed artifacts are marked deprecated, not deleted. Parallel branches are never affected.

13. **Context propagation distinguishes verbatim artifacts from summarized history**. Technical specifications pass through untouched; procedural history is compressed by a dedicated Summarizer agent using a fast model.

14. **Agent sessions terminate completely at phase boundaries**. Downstream agents start with fresh context windows assembled from blackboard artifacts and summaries. No raw chat history crosses phases.

15. **End-to-end traceability uses `trace_id`** propagated through handoff contracts, compatible with W3C Trace Context standards.
