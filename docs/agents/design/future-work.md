# Agents Design — Future Work

This document lists capabilities that were explored during research but deferred from the current design. Each item has a clear rationale for deferral and a summary of the research findings that will inform future implementation.

## 1. Deliver/Merge Phase

**What:** A fifth phase in the full operating mode — after Validate, the system merges validated work into mainline, resolves conflicts, and produces a final summary.

**Why deferred:** The current system validates code but relies on the human to integrate it. Automated merging introduces risks (conflict resolution, CI pipeline interaction, branch protection rules) that require careful design.

**Research findings:** The draft design outlined serial integration with pre-merge validation and a dedicated Merge Agent for conflict resolution. The CLI would enforce that only validated code reaches mainline.

## 2. Worktree Isolation

**What:** Each code agent operates in a dedicated Git worktree on a separate branch, with physical filesystem isolation.

**Why deferred:** Worktrees add significant operational complexity (worktree lifecycle management, branch proliferation, merge conflict resolution between parallel agents). The simpler same-branch model with DAG-based file coordination is sufficient for the current scope.

**Research findings:** Synthesis 02 explored worktree provisioning, the `.oraculo/worktrees/` directory convention, ephemeral worktree lifecycle (create on dispatch, destroy on completion), and pinned `base_sha` for deterministic execution environments.

## 3. Separate Test-Author/Implementer Agents

**What:** Two distinct agent types for TDD — one writes tests from specifications (no implementation knowledge), another writes implementation to make those tests pass (no test modification access).

**Why deferred:** The two-agent model provides stronger guarantees against test hacking (the implementer literally cannot modify tests), but doubles agent overhead and introduces handoff complexity. The single-agent-with-TDD-skill model preserves TDD discipline through skill instructions rather than physical separation.

**Research findings:** Synthesis 03 provided detailed designs for test-author context isolation (specification as sole input, anti-implementation prompt blocking), the `tdd_bundle.json` handoff format, two-layer permission enforcement (filesystem ACLs + Claude Code hooks), and the "Before" branch mechanism where the implementer's branch is derived from the test branch.

## 4. Adversarial QA (Executable Proof Pattern)

**What:** A specialized adversarial QA agent that actively tries to break code through injection, fuzzing, race condition testing, and semantic drift detection. Its claims are validated through the "Executable Proof" pattern — only executable test scripts that actually induce failures are accepted as evidence.

**Why deferred:** The current QA agent handles functional correctness, standards, and edge cases. Adversarial testing adds significant token cost and requires scoping mechanisms (adversarial focus matrix) to prevent unbounded exploration.

**Research findings:** Synthesis 04 defined four adversarial prompt patterns (state manipulation, authority spoofing, round-trip review protocol, concurrency/fuzzing). The Executable Proof pattern eliminates false positives by ignoring textual claims — only CLI-executed test failures constitute valid findings.

## 5. Mutation Testing

**What:** Automated mutation of implementation code (e.g., changing `&&` to `||`, `+` to `-`) followed by test suite re-execution to validate test quality. Tests that still pass after mutations are "surviving mutants" indicating weak assertions.

**Why deferred:** Mutation testing validates the test suite, not the implementation. It is a powerful second-order quality gate but requires tool integration per language (Stryker for JS/TS, mutmut for Python, PIT for JVM) and adds significant execution time.

**Research findings:** Synthesis 03 and 04 recommended incremental mutation testing per-story (scoped to the diff) with a configurable threshold (80% standard, 90% for high-risk changes). Surviving mutant reports would route back to the test-author for assertion strengthening.

## 6. Heterogeneous Model Selection

**What:** Assigning different LLM model families to different agent roles — e.g., Claude for code generation, Gemini for functional review, DeepSeek for adversarial testing. Different training data and alignment approaches catch different blind spots.

**Why deferred:** Requires multi-provider infrastructure, cost management per role, and fallback chains for provider unavailability. The current system works with a single model family across all agents.

**Research findings:** Synthesis 04 cited the X-MAS-Bench study confirming that heterogeneous multi-agent configurations outperform homogeneous ones. The recommended hard constraint: QA agents must use a different model family from code agents they validate.

## 7. Rich Memory System

**What:** A three-tier memory architecture — working memory (assembled on demand), episodic memory (immutable event log), and semantic memory (validated knowledge). Includes a curation pipeline (episode scoring, LLM reflection, multi-agent jury for promotion) and relational linking between knowledge entries.

**Why deferred:** The current simple model (CLAUDE.md + markdowns + ephemeral SQLite) covers the essential needs. The three-tier system adds significant complexity: schema design, curation triggers, severity thresholds, embedding generation, and search evolution (FTS5 → sqlite-vec → graph queries).

**Research findings:** Synthesis 05 provided production-ready schemas for all three tiers, composite FTS5 scoring with temporal decay, a complete curation pipeline (Episode → Score → Reflection → MAR Consensus → Promotion), and a three-phase search evolution roadmap within pure SQLite.

## 8. Advanced Dispatch

**What:** Sophisticated dispatch algorithms including critical-path priority scoring (`(critical_path_length, fan_out)`), shifting bottleneck detection (rolling average completion times per node type), and WIP limits stored in SQLite configuration tables.

**Why deferred:** The current dispatch model (the orchestrator evaluates the DAG and dispatches ready tasks) is straightforward and sufficient. Advanced dispatch optimizations matter when running many parallel agents with tight resource constraints — a scale the system has not reached.

**Research findings:** Synthesis 01 defined the frontier-based dispatch loop, the critical-path priority heuristic, and Drum-Buffer-Rope mapping (QA as Drum, completed tasks awaiting QA as Buffer, dispatch throttling as Rope). The dispatch loop would run autonomously in the CLI without LLM involvement.

## 9. Event Sourcing

**What:** An append-only event log as the single source of truth for all state changes. Every task status change, QA verdict, and DAG mutation is recorded as an immutable event. Materialized tables are projections derived from the event stream.

**Why deferred:** The current SQLite schema uses direct status updates (simpler to implement and query). Event sourcing provides superior auditability and crash recovery but adds complexity in schema design, event versioning, and snapshot management.

**Research findings:** Synthesis 01 and 05 both recommended event sourcing with synchronous materialization — updating materialized tables within the same transaction as event insertion. WAL mode (`PRAGMA journal_mode = WAL`) was recommended for concurrent read/write access.

## 10. RBAC and Governance

**What:** Role-based access control for the SQLite blackboard — restricting which agents can read/write which tables, enforced through CLI middleware and SQLite triggers.

**Why deferred:** The current system relies on the orchestrator to assign appropriate tasks and context to agents. Formal RBAC adds database-level enforcement but is unnecessary when the orchestrator is the sole dispatcher and the CLI is the sole write gateway.

**Research findings:** Synthesis 06 designed a three-layer governance model: policy storage in SQLite (`security_policies` table), database enforcement via `BEFORE INSERT/UPDATE/DELETE` triggers with `RAISE(FAIL)`, and CLI middleware for tool permission isolation per role.
