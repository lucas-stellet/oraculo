# Synthesis: Agent Isolation and Execution Environment

This document synthesizes two independent research results on the execution environment for Oraculo's code agents, cross-referencing them against the six guiding questions from the research prompt and the project's core philosophy.

**Research sources:**
- **Source A:** "Execution Environments and Isolation Boundaries for Multi-Agent Coding Systems" -- broad architectural analysis with emphasis on deterministic infrastructure patterns.
- **Source B:** "Oraculo Code Agent Execution Environment" -- concrete specification oriented toward CLI responsibilities and SQLite-backed state management.

---

## 1. Worktree Lifecycle

### Consensus

Both sources fully agree on the foundational model:

| Aspect | Source A | Source B | Verdict |
|---|---|---|---|
| One worktree per DAG node | Yes | Yes | **Settled** |
| CLI is sole creator/destroyer | Yes | Yes | **Settled** |
| Worktrees are ephemeral | Yes | Yes | **Settled** |
| Wipe-and-rebuild on retry | Yes (explicit "pristine environment forcing") | Yes ("never reuses an existing worktree") | **Settled** |
| Garbage collection via `git worktree remove` + `prune` | Yes | Yes | **Settled** |

### Complementary details

**Naming convention.** Source A proposes `.trees/task-<id>` with branches named `feature/task-<id>`. Source B proposes `.oraculo/worktrees/<dag_run_id>/<node_id>/` with branches named `oraculo/<dag_run_id>/<node_id>`. Source B's scheme is superior because it:
- Scopes worktrees under the Oraculo namespace, avoiding collision with user-created worktrees.
- Includes the `dag_run_id`, enabling multiple DAG runs to coexist (important for retry scenarios and parallel feature tracks).
- Creates a natural filesystem hierarchy for monitoring and cleanup.

**Base commit pinning.** Source B introduces the concept of recording `base_sha` in SQLite so every worktree is created from a known, deterministic commit. Source A does not address this explicitly. This is a critical addition -- without a pinned base SHA, retries or late-scheduled nodes could silently pick up unrelated changes from main.

**State tracking.** Source B records the full mapping `(dag_run_id, node_id) -> {base_sha, worktree_path, branch_name, result_sha}` in SQLite. Source A mentions this implicitly but does not formalize the schema. Source B's approach is more actionable.

### Synthesized recommendation

Adopt Source B's naming and state model with one refinement from Source A -- also track the `attempt_number` so retried nodes have distinguishable paths:

```
Path:   .oraculo/worktrees/<dag_run_id>/<node_id>/
Branch: oraculo/<dag_run_id>/<node_id>
```

On retry, the old worktree is removed before the new one is created. The attempt number is tracked in SQLite but does not alter the path or branch name (the old branch is deleted/orphaned first).

SQLite schema excerpt:

```sql
CREATE TABLE worktrees (
    dag_run_id   TEXT NOT NULL,
    node_id      TEXT NOT NULL,
    attempt      INTEGER NOT NULL DEFAULT 1,
    base_sha     TEXT NOT NULL,
    result_sha   TEXT,
    worktree_path TEXT NOT NULL,
    branch_name  TEXT NOT NULL,
    status       TEXT NOT NULL CHECK (status IN (
        'provisioning', 'active', 'succeeded',
        'failed', 'cancelled', 'expired', 'cleaning'
    )),
    created_at   TEXT NOT NULL DEFAULT (datetime('now')),
    cleaned_at   TEXT,
    PRIMARY KEY (dag_run_id, node_id, attempt)
);
```

---

## 2. Context Injection (Working Set Builder)

### Consensus

Both sources strongly agree on the principle of Minimal Viable Information (MVI): the agent must not discover its own context. The CLI builds the working set deterministically before the agent starts.

### Complementary details

**Source A** provides deep algorithmic detail:
- Tree-sitter-based AST parsing to build a structural dependency graph.
- PageRank with a personalized vector (10x-50x boost on task-relevant files) to rank symbol importance.
- Binary search optimization to fit the highest-ranked symbols within a token budget.
- Claims approximately 80% token reduction vs. unrestricted context.

**Source B** provides operational detail:
- The CLI maintains a persistent repository map in SQLite (files, symbols, signatures, dependency edges).
- The MVI bundle includes: task spec, scope (write/read files), repo map slice, conventions.
- Topological selection uses a configurable "radius" around the anchor file/symbol (direct dependencies + direct dependents + optionally tests).
- An `AGENT_<node_id>.md` file is generated inside the worktree as the agent's contract.

### Conflict: PageRank vs. radius-based graph walk

Source A proposes PageRank, a global algorithm that ranks every symbol in the codebase relative to the task. Source B proposes a simpler radius-based walk (depth-1 or depth-2 from the anchor). These are not mutually exclusive, but they represent different tradeoffs:

| Approach | Strength | Weakness |
|---|---|---|
| PageRank | Surfaces globally important symbols the agent should be aware of (e.g., core types used everywhere) | Computationally expensive; requires full graph materialization; may surface symbols that are important globally but irrelevant to the task |
| Radius walk | Fast, predictable, deterministic; directly aligned with the task's neighborhood | May miss important symbols outside the immediate radius (e.g., a utility function 3 hops away that the agent needs) |

### Synthesized recommendation

Start with the radius-based walk (Source B) as the primary algorithm because it is simpler, faster, and more deterministic. Augment it with a lightweight importance signal:

1. **Anchor identification** -- The DAG node's task spec names target files/symbols.
2. **Radius walk (depth configurable, default 2)** -- Walk the dependency graph from the anchor, collecting direct deps and dependents.
3. **Importance filter** -- Within the walked set, prioritize symbols with higher in-degree (number of dependents) to ensure core types/interfaces are included. This is a local approximation of PageRank without the global computation.
4. **Token budget truncation** -- Use binary search (from Source A) to fit the selected symbols within the node's token budget, dropping lowest-importance symbols first.
5. **AGENT spec file** -- Generate the `AGENT_<node_id>.md` contract (from Source B) summarizing scope, rules, and anti-goals.

Defer full PageRank implementation to a later optimization phase. The radius walk with in-degree ranking is sufficient for the initial system and avoids premature complexity.

---

## 3. Permission Boundaries

### Divergence

This is the area of greatest divergence between the two sources. They agree on the goal (agents cannot write to unauthorized files) but propose fundamentally different enforcement mechanisms.

**Source A: PreToolUse lifecycle hooks (pre-execution blocking)**
- Intercepts every `Edit` or `Write` tool call before the filesystem is mutated.
- Validates the `file_path` against a whitelist.
- Returns exit code 2 to block unauthorized access, feeding a rejection message back to the LLM.
- Also proposes PostToolUse hooks for automatic linting/formatting after writes.

**Source B: Centralized diff validation (post-execution blocking)**
- Agents are allowed to modify files freely during their session.
- The CLI inspects the resulting diff/patch before creating any Git commit.
- Rejects the entire changeset if it touches forbidden paths, exceeds diff size limits, or contains secret-like literals.
- ACLs stored in SQLite: `allowed_write_globs`, `allowed_read_globs`, `forbidden_globs`.

### Analysis

| Criterion | PreToolUse Hooks (A) | Diff Validation (B) |
|---|---|---|
| Fail-fast | Yes -- blocks immediately | No -- wasted tokens on unauthorized changes |
| Token efficiency | Higher -- agent redirects early | Lower -- agent may spend many tokens on work that gets rejected |
| Implementation complexity | Higher -- requires hook infrastructure per tool call | Lower -- single validation pass at commit time |
| False positives | Risk of blocking legitimate intermediate states | No risk -- validates final output only |
| Determinism | Fully deterministic (exit codes) | Fully deterministic (diff parsing) |
| Alignment with Oraculo philosophy | Strong ("explicit, deterministic barriers") | Strong ("CLI validates before commit") |

### Synthesized recommendation

Use **both mechanisms as complementary layers** (defense in depth):

1. **PreToolUse hooks (primary enforcement)** -- Block writes to clearly forbidden files (test files, config files, other agents' targets) before the filesystem is touched. This is the fast, cheap guard. Implement via Claude Code lifecycle hooks calling the CLI binary for validation.

2. **Diff validation (secondary enforcement)** -- Before the CLI commits the agent's changes, validate the full diff against the SQLite ACL. This catches anything the hooks missed (e.g., files created via shell commands that bypass the Edit/Write tools) and enforces aggregate constraints (diff size, file count, secret detection).

3. **PostToolUse hooks for quality gates** -- After successful writes, trigger linting/formatting (from Source A). This shifts syntactic validation from the LLM to deterministic tools.

SQLite ACL schema (from Source B, extended):

```sql
CREATE TABLE node_acls (
    dag_run_id         TEXT NOT NULL,
    node_id            TEXT NOT NULL,
    allowed_write_globs TEXT NOT NULL,  -- JSON array of glob patterns
    allowed_read_globs  TEXT NOT NULL,  -- JSON array of glob patterns
    forbidden_globs     TEXT NOT NULL,  -- JSON array of glob patterns
    max_diff_lines     INTEGER,
    max_files_changed  INTEGER,
    FOREIGN KEY (dag_run_id, node_id) REFERENCES worktrees(dag_run_id, node_id)
);
```

---

## 4. Merge Strategy

### Consensus

Both sources agree that:
- Only the CLI writes to the mainline branch (single-writer principle).
- Agents produce changes on their own branches; they never merge themselves.
- Merge conflicts should be handled by a dedicated mechanism, not by the original generating agent.

### Complementary details

**Source A** emphasizes:
- Standard `git merge` as the first attempt; fallback to a specialized Merge Agent.
- The Merge Agent receives only conflicting files, their dependents, and divergent commit histories.
- Pre-merge validation (test suite + AST linting) before finalizing.

**Source B** emphasizes:
- Merge is modeled as its own DAG node (not an ad-hoc side process).
- Merges are applied serially into main, even when nodes completed in parallel.
- Uses `base_sha` / `result_sha` pairs as changesets; attempts fast-forward merge or cherry-pick.
- The merge agent is optional -- the CLI handles clean merges; the agent handles conflicts.

### Synthesized recommendation

Adopt Source B's model of merge-as-DAG-node with Source A's conflict resolution pattern:

1. **Clean merge path (CLI-only):** The CLI attempts to fast-forward or cherry-pick `result_sha` onto current main. If no conflicts, the CLI commits directly. No agent needed.

2. **Conflict path (Merge Agent):** If conflicts are detected, the CLI creates a new DAG node of type `merge-resolution`. A specialized Merge Agent is spawned with:
   - The conflicting files and their diff hunks.
   - A repo map slice covering only the conflicting files' dependency neighborhood.
   - The commit messages and task specs from the conflicting nodes (for semantic intent).

3. **Serial integration order:** Even when multiple nodes complete in parallel, merges are applied one at a time in topological order. This prevents compound conflicts and keeps the mainline in a known-good state after each merge.

4. **Pre-merge validation:** After every merge (clean or agent-resolved), the CLI runs the validation pipeline (lint, tests, static analysis) before advancing the main HEAD. A failed validation rejects the merge and marks the node for investigation.

5. **Conflict prevention via base SHA management:** When a merge succeeds and advances main, any still-running nodes that were based on an older `base_sha` are not automatically rebased. Instead, when those nodes complete, the CLI detects the divergence and performs the merge/conflict-check at that point. This avoids disrupting in-progress agents.

---

## 5. Environment Setup

### Consensus

Both sources agree that worktrees should differ only in code, not in tooling. Dependencies and execution environments should be shared across worktrees.

### Complementary details

**Source A** provides specific tooling recommendations:
- Content-addressable package managers (pnpm) with hard links to avoid duplicate dependency installations.
- Copy-on-Write filesystem awareness (APFS on macOS, Btrfs on Linux) to prevent corruption of shared stores.
- Dynamic `.env` generation per worktree with unique ports/keys to prevent test collisions.

**Source B** provides architectural framing:
- Defines three environment models: Docker image, devcontainer, or host toolchain.
- Nodes that can run tests are explicitly marked in SQLite (`can_run_tests = true`).
- Secrets never live inside the worktree -- they are injected as environment variables by the runner.
- "Environment profiles" define which secrets and services a node needs.

### Synthesized recommendation

1. **Shared toolchain, per-worktree code:** The host machine (or container) provides the toolchain. Worktrees contain only the source code. Dependencies are installed once and shared via hard links (pnpm for Node.js; equivalent strategies for other ecosystems).

2. **Environment profiles in SQLite:** Each node type has an associated environment profile specifying required secrets, services, and test capabilities. The CLI resolves the profile before spawning the agent.

3. **Dynamic environment injection:** For nodes that run tests, the CLI generates isolated environment configurations (unique ports, ephemeral database names) to prevent collisions between parallel agents. These are injected as environment variables, never written to tracked files.

4. **Read-only dependency directories:** Enforce via PreToolUse hooks that agents cannot modify `node_modules/`, `vendor/`, or equivalent dependency directories. This prevents corruption of the shared dependency store.

5. **Defer containerization:** Start with the shared host toolchain model. Containerization per worktree adds significant orchestration overhead and is only justified when agents need incompatible toolchain versions or untrusted code execution. Revisit when running untrusted or user-submitted code.

---

## 6. Resource Limits and Circuit Breakers

### Consensus

Both sources agree on the need for hard resource limits enforced at the orchestrator level, not within the agent's own reasoning.

### Complementary details

**Source A** provides implementation patterns:
- Global token budgeting across the entire agent lifecycle (not per-API-call).
- `go-circuitbreaker` library for monitoring execution patterns.
- Circuit breaker state machine: CLOSED -> OPEN on breach; exponential backoff before retry.
- Chain-level containment: shared counter across nested tool calls to prevent exponential token burn.
- Injection of failure logs into the MVI context on retry to prevent historical repetition.

**Source B** provides operational patterns:
- Three budget dimensions: max tokens, max tool calls/steps, wall-clock timeout.
- Node status `circuit_breaker_tripped` as a terminal state.
- Minimum Viable Agent (MVA) principle: smaller nodes have more predictable budgets.
- Global circuit breaker by node type: if a category of nodes fails repeatedly, reduce parallelism or pause scheduling for that type.

### Synthesized recommendation

Adopt a three-tier resource control model:

**Tier 1: Per-node hard limits**
- Maximum tokens (model input + output combined).
- Maximum tool call count (step budget).
- Wall-clock timeout.
- On breach: terminate agent, mark node as `circuit_breaker_tripped`, trigger worktree cleanup.

**Tier 2: Behavioral detection**
- Track consecutive identical tool failures (e.g., same file, same error 3+ times).
- Track token velocity (tokens per meaningful code change) to detect spinning.
- On detection: trip the circuit breaker before hitting the hard limit.

**Tier 3: Global type-level throttling**
- Track failure rates per node type across the DAG run.
- If a node type exceeds a configurable error threshold (e.g., 3 consecutive failures):
  - Reduce parallelism for that type.
  - Pause scheduling and escalate to the orchestrator for human review.

**Retry enrichment:** When a failed node is retried, inject a structured failure summary into the new agent's MVI bundle. This summary includes the previous failure reason, the files that were problematic, and any error logs -- preventing the new agent from repeating the same mistakes.

SQLite schema excerpt:

```sql
CREATE TABLE node_resource_limits (
    dag_run_id       TEXT NOT NULL,
    node_id          TEXT NOT NULL,
    max_tokens       INTEGER NOT NULL,
    max_steps        INTEGER NOT NULL,
    timeout_seconds  INTEGER NOT NULL,
    tokens_used      INTEGER NOT NULL DEFAULT 0,
    steps_used       INTEGER NOT NULL DEFAULT 0,
    started_at       TEXT,
    status           TEXT NOT NULL CHECK (status IN (
        'pending', 'running', 'succeeded', 'failed',
        'circuit_breaker_tripped', 'cancelled', 'expired'
    )),
    failure_reason   TEXT,
    failure_logs     TEXT,  -- JSON, injected into retry MVI
    PRIMARY KEY (dag_run_id, node_id)
);
```

---

## Gaps and Open Questions

### Addressed inadequately by both sources

1. **Multi-language repository maps.** Both sources discuss Tree-sitter and AST parsing but neither addresses how the repo map handles polyglot repositories (e.g., a project with Go, TypeScript, and SQL). The CLI will need language-specific parsers and a unified symbol graph format.

2. **Worktree disk pressure at scale.** Source A mentions 172 GB in large deployments but neither source proposes a concrete disk quota or monitoring strategy. For Oraculo running on developer machines (not cloud VMs), disk pressure could become a real issue.

3. **Agent-to-agent communication.** Neither source addresses whether agents within the same DAG run should ever be able to read each other's in-progress state (e.g., checking if a dependency interface has been finalized by a parallel agent). The current model assumes total isolation, but there may be edge cases where a lightweight read-only channel is needed.

4. **Hot-reload of the repo map.** Source B states the repo map is "updated out-of-band" but does not specify when. If one agent's changes alter the dependency graph, subsequent agents in the DAG may need an updated map. The question of when/how to refresh the repo map between nodes is unresolved.

5. **Merge ordering heuristics.** Both sources agree on serial merging but neither specifies the ordering strategy when multiple nodes complete simultaneously. Options include topological order (dependencies first), smallest-diff-first, or longest-running-first. This impacts conflict probability.

6. **Windows/cross-platform worktree behavior.** All examples assume Unix-like systems. Git worktree behavior on Windows (path lengths, symlinks, file locking) may require additional handling.

---

## Key Design Decisions Summary

1. **Worktree model:** One ephemeral Git worktree per DAG node, created from a pinned `base_sha`, scoped under `.oraculo/worktrees/<dag_run_id>/<node_id>/`. Wipe and rebuild on every retry.

2. **State tracking:** SQLite is the source of truth for all worktree metadata, ACLs, resource limits, and DAG state. Git tracks only codebase evolution.

3. **Context injection:** The CLI builds the MVI bundle using a radius-based dependency graph walk with in-degree ranking, fitted to a token budget via binary search. The agent receives an `AGENT_<node_id>.md` contract file. Agents never search the repository themselves.

4. **Permission enforcement:** Defense in depth -- PreToolUse hooks block unauthorized file mutations in real-time; diff validation at commit time catches anything hooks missed. ACLs are stored in SQLite as glob patterns per node.

5. **Merge strategy:** The CLI is the sole writer to main. Clean merges are handled by the CLI directly. Conflicts spawn a dedicated Merge Agent as a new DAG node. All merges are serial and followed by validation (lint + tests).

6. **Environment setup:** Shared host toolchain with hard-linked dependencies (pnpm or equivalent). Agents cannot modify dependency directories. Secrets are injected via environment variables, never stored in the worktree. Containerization is deferred.

7. **Resource limits:** Three-tier model -- per-node hard limits (tokens, steps, timeout), behavioral detection (repeated failures, token velocity), and global type-level throttling. Failed nodes receive enriched retry context.

8. **Lifecycle hooks as infrastructure:** PreToolUse, PostToolUse, and related hooks form the deterministic control plane. They shift compliance from the LLM's stochastic reasoning to the CLI's binary exit codes.

9. **Merge Agent scope:** The Merge Agent receives only conflicting files, their dependency neighborhood, and the semantic intent from both sides. It does not receive the full repository context.

10. **AGENT spec as contract:** Every agent starts with a generated `AGENT_<node_id>.md` that defines its goal, editable files, read-only files, forbidden paths, and anti-goals. This file is the single point of reference for the agent's boundaries.
