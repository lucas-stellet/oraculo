# Synthesis: TDD Pipeline and Test/Implementation Separation

## Overview

This document synthesizes two independent research results on TDD pipeline design for Oraculo's multi-agent system:

- **Source A**: "Pipeline de TDD com Separacao de Agentes e Validacao de Traco" (Portuguese-language research)
- **Source B**: "Architecting the Agentic TDD Pipeline: Enforcing Test-First Discipline in Multi-Agent Systems" (English-language research)

Both sources converge strongly on the core architecture. Where they diverge, one typically fills a gap in the other. The synthesis below maps findings to the six guiding questions from the research prompt and distills actionable design decisions.

---

## Cross-Reference Summary

| Dimension | Source A | Source B | Agreement |
|---|---|---|---|
| Test-author context isolation | Anti-implementation prompt blocking | Spec-driven context locking (GSD/spec-workflow-mcp) | Strong convergence |
| Test granularity | Feature -> acceptance; bug -> unit | Unit + integration primary; E2E avoided | Complementary |
| Handoff mechanism | Branch "Before" + `tdd_bundle.json` metadata | Worktree provisioning + Pydantic JSON payload | Strong convergence |
| Permission enforcement | Two layers: FS/ACL + CLI hooks | Two layers: OS permissions + pre-commit hooks | Identical |
| Red-green trace validation | Event sourcing in SQLite with causal ordering | Event sourcing in SQLite with topological verification | Strong convergence |
| Validation algorithm | First prod write must follow a failing test run | T1(Red) -> T2(Code) -> T3(Green) timestamps | Identical logic, different notation |
| Test hacking defense | Defense in depth: permissions + trace + mutation | Mutation testing + behavioral anomaly detection | Complementary |
| Mutation testing | Incremental per-story + nightly full | Threshold-based gating (>85%) + LLM-guided mutants | Complementary |
| Refactor phase | Optional quality gate; lint/format enforcement | Dedicated Refactor Agent triggered by cyclomatic complexity | Complementary |
| Multi-language support | JUnit XML primary, TAP fallback, dual ingest | JUnit XML standardized via junitparser, TAP secondary | Strong convergence |
| Enforcement architecture | Local fast-fail + remote/orchestrator gate | Pre-commit hook + CLI rejection | Source A more thorough |

---

## Guiding Question 1: Test-Author Agent Design

### Consensus

Both sources agree on a fundamental principle: **the test-author agent must be shielded from implementation details**. Source A calls this "anti-implementation prompt blocking" and Source B calls it "spec-driven context locking." The reasoning is identical -- when a single context holds both test and implementation concerns, the agent optimizes tests for the solution it imagines, producing weak oracles.

### Context the Test-Author Needs

From Source B (GSD/spec-workflow-mcp patterns):
- **Immutable specification documents** (Requirements.md, acceptance criteria) as sole ground truth
- **Technical design documents** for mock schema generation
- **Ubiquitous language** from the domain (DDD principles)
- **No access** to existing implementation code or planned architecture

From Source A:
- Epic/Story specifications with acceptance criteria
- Explicit prompt instruction: "Do not infer implementation; describe only observable contracts and edge cases, as if you were a consumer of the API"
- Focus on behavior, contracts, and boundary conditions

### Prompt Structure

Combined recommendation:

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

### Test Granularity

The sources offer complementary guidance:

| Change Type | Source A Recommendation | Source B Recommendation | Synthesis |
|---|---|---|---|
| New feature | Acceptance/integration tests | Unit + integration tests | Start with acceptance-level tests for behavior, add unit tests for algorithmic logic |
| Bug fix | Minimal reproduction (unit/integration) with fail-to-pass | Unit tests for discrete logic | Fail-to-pass reproduction test (unit or integration) |
| Refactoring | Existing tests suffice | Existing tests suffice | No new tests needed; existing suite is the contract |

**Synthesized decision**: The dominant test type depends on the change type. Features get acceptance/integration tests that validate user-visible behavior. Bugs get minimal reproduction tests. Unit tests serve algorithmic components. E2E tests are excluded from the TDD loop due to non-determinism and latency.

### Test Quality Assurance at Authoring Time

Source A uniquely raises **test determinism** as a handoff prerequisite: the test-author must run tests twice in the "Before" state and confirm consistent failure. Flaky tests destroy the Red-Green proof because Red could be accidental and Green could be luck.

Source B adds **structured output validation**: the test-author agent should return tests in a strict JSON schema (or Pydantic model) that the CLI validates before writing to disk. This prevents malformed tests from entering the pipeline.

**Synthesized decision**: The test-author's "Definition of Done" is:
1. All acceptance criteria mapped to test cases
2. Tests pass structural validation (JSON schema)
3. Tests fail deterministically in the Before state (verified by N>=2 runs)
4. Tests are syntactically valid and executable

---

## Guiding Question 2: Handoff -- Tests to Implementer

### Consensus

Both sources converge on the same architecture:

1. **The orchestrator creates a "Before" branch** containing only the new tests (no implementation)
2. **The implementer's worktree starts from this branch**, so tests already exist when work begins
3. **Metadata accompanies the handoff** as a structured payload

### Branch and Worktree Flow

```
main
  |
  +-- feature-tests (test-author worktree)
  |     |-- writes tests, commits
  |     |-- orchestrator verifies deterministic failure
  |
  +-- feature-impl (implementer worktree, branched FROM feature-tests)
        |-- tests already present (read-only)
        |-- writes implementation only
```

Source A specifies: "The orchestrator must create a 'Before' branch (tests only) and derive the implementer's branch from it, ensuring the first `git diff` of the implementer never includes test changes."

Source B adds the `.trees/` directory convention from the Agent Worktree (AWT) pattern, with explicit provisioning and teardown scripts.

### Handoff Metadata Bundle

Source A proposes a `tdd_bundle.json`; Source B proposes a Pydantic-validated JSON payload. These are the same concept. Combined schema:

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

### Permission Enforcement

Both sources independently arrive at **two-layer enforcement**:

| Layer | Source A | Source B |
|---|---|---|
| OS/filesystem | FS/ACL read-only on test paths | OS-level read-only permissions on test directories |
| CLI/hooks | Claude Code `PreToolUse` hooks block writes to `tests/**` | Pre-commit hooks reject commits with test file modifications |

**Synthesized decision**: Implement both layers. The Claude Code hook provides fast feedback (blocks the action before it happens). The filesystem permission provides defense-in-depth (blocks even if hooks are bypassed). The test-author's worktree similarly blocks writes to `src/**` when appropriate.

---

## Guiding Question 3: Red-Green-Refactor Enforcement

### Consensus

Both sources agree on event sourcing in SQLite as the mechanism. The validation logic is identical in substance, differing only in notation.

### SQLite Event Schema

Source A provides a detailed, practical schema. Source B describes the same structure at a higher level. Combined and refined:

**Events table:**

```sql
CREATE TABLE events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    stream_id   TEXT NOT NULL,           -- feature branch / task ID
    session_id  TEXT NOT NULL,           -- agent session
    timestamp   TEXT NOT NULL,           -- ISO-8601 microsecond precision
    event_type  TEXT NOT NULL,           -- enum below
    agent_role  TEXT NOT NULL,           -- test_author | implementer | refactor
    data        TEXT NOT NULL            -- JSON payload
);

-- Index for trace validation queries
CREATE INDEX idx_events_stream_time ON events(stream_id, timestamp);
CREATE INDEX idx_events_stream_type ON events(stream_id, event_type);
```

**Event types and their payloads:**

| Event Type | Payload Fields | Source |
|---|---|---|
| `SessionStart` | `worktree_id`, `base_commit`, `story_id`, `tdd_bundle_hash` | A |
| `TestRunStart` | `test_run_id`, `command`, `cwd`, `env_fingerprint`, `requested_scope` | A |
| `TestRunResult` | `exit_code`, `total`, `failed`, `skipped`, `duration_ms`, `report_format`, `report_digest`, `failed_test_ids[]` | A+B |
| `FileWriteAttempt` | `path`, `category` (test/prod/config), `allowed`, `blocked_reason`, `before_hash`, `after_hash` | A |
| `FileModified` | `path`, `before_hash`, `after_hash` | B |
| `CommitRequested` | `commit_hash`, `files_changed[]`, `validation_result` | A+B |
| `CycleMarker` | `phase` (red/green/refactor), `notes` | A |
| `LintRunResult` | `tool`, `exit_code`, `issues_count` | A |
| `PolicyViolation` | `violation_type`, `details`, `action_taken` | B |

### Validation Algorithm

Both sources describe the same logic. Source A uses causal ordering; Source B uses timestamp-based topological verification. Synthesized:

```
Given stream_id S and tdd_bundle B:

1. Find T_impl_first = first FileWriteAttempt where category=prod
   AND allowed=true after the handoff event

2. REQUIRE: there exists a TestRunResult R_red where:
   - R_red.timestamp < T_impl_first.timestamp (causal ordering)
   - R_red.exit_code != 0
   - R_red.failed > 0
   - R_red.failed_test_ids INTERSECT B.expected_failures is non-empty
     (the failing tests belong to this story's bundle)

3. Find T_impl_last = last FileWriteAttempt where category=prod

4. REQUIRE: there exists a TestRunResult R_green where:
   - R_green.timestamp > T_impl_last.timestamp
   - R_green.exit_code == 0
   - R_green.failed == 0

5. REQUIRE: no FileWriteAttempt where category=test AND allowed=true
   exists in the implementer's session

If any requirement fails: REJECT commit, log PolicyViolation.
```

### Detecting Skipped Red Phase

Source A: If the first prod write precedes any test run, the Red phase was skipped.
Source B: If a `FileModified` event precedes the initial `TestRun`, the TDD protocol was violated.

Both lead to the same check: **the earliest production file modification must be strictly preceded by a failing test run targeting the story's tests**.

### Enforcement Architecture

Source A uniquely and importantly raises the point that **local hooks alone are insufficient** because they can be bypassed with `--no-verify`. The synthesized architecture:

| Layer | Purpose | Mechanism |
|---|---|---|
| Local (Claude Code hooks) | Fast feedback, prevent wasted tokens | `PreToolUse` hooks block writes before Red; `PostToolUse` hooks log events |
| Orchestrator gate | Authoritative validation | CLI queries SQLite trace before accepting merge |
| Remote (optional) | Defense-in-depth | `pre-receive` hook validates trace artifact/signature |

**Synthesized decision**: The orchestrator gate is the authoritative enforcement point. Local hooks optimize developer experience. Remote hooks add defense-in-depth but are not strictly required if the orchestrator controls all merges.

---

## Guiding Question 4: Test Hacking Detection

### Consensus

Both sources agree that permission boundaries are necessary but insufficient. Defense-in-depth is required.

### Defense Layers

| Layer | Mechanism | Source |
|---|---|---|
| 1. Permission boundary | Test files are read-only for implementer | A+B |
| 2. Trace validation | Red-Green sequence proven in SQLite | A+B |
| 3. Mutation testing | Detect weak oracles / trivial assertions | A+B |
| 4. Test determinism | Verify consistent failure (not flaky) | A |
| 5. Behavioral anomaly detection | Detect brute-forcing / excessive test runs | B |

### Mutation Testing Integration

Both sources strongly recommend mutation testing. Source A provides more practical integration guidance; Source B provides more detail on gating thresholds.

**Tools by ecosystem:**

| Language | Tool | Source |
|---|---|---|
| JavaScript/TypeScript | Stryker | A+B |
| JVM (Java/Kotlin) | PIT | A+B |
| Python | mutmut | A |
| .NET | Stryker.NET | A |
| PHP | Infection | A |
| Ruby | Mutant | A |
| Go | go-mutesting | A |

**Execution policies (Source A):**

| Policy | Scope | Frequency |
|---|---|---|
| Incremental mutation | Modules touched by current story | Every story completion |
| Full mutation | Entire codebase | Nightly/weekly |

**Gating threshold (Source B):** Mutation score must exceed a configured minimum (e.g., 85%) for the story's affected modules. Below threshold, the implementation is rejected and the **surviving mutant report is routed back to the test-author** for assertion strengthening.

**Synthesized decision**: Mutation testing validates test quality, not TDD sequence. It answers "are the tests strong enough?" while the trace validation answers "was the process followed?" Both are required. Incremental mutation runs per-story with a configurable threshold; full runs happen on a schedule.

### Behavioral Anomaly Detection (Source B only)

Source B uniquely proposes monitoring the SQLite event stream for suspicious patterns:
- Excessive test executions without code modifications (brute-forcing)
- Rapid-fire test runs (inferring test logic through side channels)

**Synthesized decision**: Implement anomaly detection as a secondary defense. The CLI monitors for patterns like >N test runs without intervening code changes and flags these for human review. This is a lower priority than mutation testing but provides an additional safety net.

### Coverage vs. Mutation Score

Source A cites research showing test suites can achieve 100% line coverage but only 4% mutation score. This makes a compelling case that **coverage alone is not a valid quality metric**.

**Synthesized decision**: Coverage is a hygiene metric (set a floor, e.g., >80%), but mutation score is the quality metric for TDD gating decisions.

---

## Guiding Question 5: Refactor Phase

### Divergence

This is where the sources diverge most significantly:

| Aspect | Source A | Source B |
|---|---|---|
| Who refactors? | Same implementer agent (small codebases); optional dedicated agent (large/legacy) | Dedicated Refactor Agent (always) |
| When to trigger? | After Green; optional quality gate | Conditionally, based on cyclomatic complexity threshold |
| What enforces quality? | Lint/format as deterministic gate | Continuous test re-execution after every refactor modification |
| Is it mandatory? | No -- Green is sufficient for merge; lint/format may be required | No -- triggered only when complexity exceeds threshold |

### Synthesized Approach

The cyclomatic complexity trigger from Source B is an elegant optimization: avoid spending tokens on refactoring when the code is already clean. Combined with Source A's lint/format gate:

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

**Synthesized decision**: The refactor phase is conditionally triggered by static analysis metrics. When triggered, a dedicated Refactor Agent operates under the same test-contract constraints. Lint/format is always mandatory as a deterministic quality floor.

---

## Guiding Question 6: Multi-Language Support

### Strong Consensus

Both sources arrive at nearly identical conclusions on this topic.

### Test Result Contract

**Primary format**: JUnit XML -- both sources agree this is the "lingua franca" of test results. Virtually all modern test frameworks can produce it natively or via plugins.

**Secondary format**: TAP (Test Anything Protocol) -- both sources agree this serves as a fallback for constrained environments.

**Language-specific structured output**: Source A notes Go's `test2json` as an example of native structured output. The pipeline should prefer these when available and convert to JUnit XML for uniform processing.

### CLI Integration Pattern

Source B provides a specific implementation pattern using `junitparser`:

```
1. Agent executes: pytest tests/ --junitxml=report.xml
2. CLI intercepts and injects format flags if missing
3. CLI parses: junitparser verify report.xml
4. CLI extracts: failure_count, test_ids, duration
5. CLI writes: TestRunResult event to SQLite
6. Trace validation queries operate on uniform data
```

### Parser Architecture

Source A raises an important reliability concern: **parser errors invalidate the TDD proof**. If the parser misreads a failure as a pass (or vice versa), the entire trace validation becomes meaningless.

**Synthesized decision**: Build a versioned parser/adapter layer with:

| Priority | Strategy | When |
|---|---|---|
| 1 | Native structured output (JSON/XML) from test runner | Always preferred |
| 2 | JUnit XML via reporter plugin | When native output is not structured |
| 3 | TAP parsing | Fallback for constrained environments |
| 4 | Exit code only | Last resort (provides pass/fail but no test-level detail) |

The parser layer itself must be tested with synthetic failure injection (Source A: "inject known failures and verify the parser detects them"). This prevents silent parser regressions from undermining the trust layer.

### Supported Frameworks (Non-Exhaustive)

| Language | Test Runner | JUnit XML Support |
|---|---|---|
| Python | pytest | `--junitxml=report.xml` |
| JavaScript | Jest | `jest-junit` reporter |
| TypeScript | Vitest | `vitest-junit-reporter` |
| Java | JUnit/Gradle/Maven | Native |
| Go | go test | `go-junit-report` converter |
| Rust | cargo-nextest | Native JUnit/XUnit XML |
| .NET | xUnit/NUnit | Native |
| Ruby | RSpec | `rspec_junit_formatter` |

---

## Gaps and Open Questions

### Adequately Addressed
All six guiding questions received thorough treatment across the two sources. No major gaps exist in coverage.

### Partially Addressed -- Needs Further Work

1. **Structured output schema for test-author agent**: Source B proposes JSON schema / Pydantic validation but does not provide a concrete schema. The `tdd_bundle.json` from Source A covers handoff metadata but not the test code structure itself. **Action needed**: Define the exact schema for test-author output validation.

2. **Mutation testing threshold calibration**: Source B suggests ">85%" but acknowledges this is a starting point. **Action needed**: Run experiments on real codebases to determine appropriate per-language thresholds. Consider making the threshold configurable per project.

3. **Behavioral anomaly detection thresholds**: Source B proposes monitoring for excessive test runs but does not specify what "excessive" means quantitatively. **Action needed**: Define heuristics based on observed agent behavior patterns.

4. **Refactor agent prompt design**: Both sources discuss when to trigger refactoring but neither provides a detailed prompt template for the Refactor Agent. **Action needed**: Design the Refactor Agent prompt with the same rigor as the test-author prompt.

### Not Addressed

5. **Incremental TDD within a single story**: When a story decomposes into multiple functions, should the test-author write all tests upfront, or should there be multiple Red-Green cycles within a single story? Neither source addresses the granularity of cycles within a task.

6. **Test-author and implementer context window management**: For large stories with many test files, how to manage the context window limits of the agents? Neither source addresses token budget constraints.

7. **Rollback semantics on repeated failure**: If the implementer fails to achieve Green after N attempts, what happens? Neither source defines retry limits or escalation paths (though Oraculo's design.md implies return to earlier phases).

---

## Key Design Decisions Summary

1. **Separate agents, separate worktrees**: The test-author and implementer are distinct Claude Code sub-agents operating in isolated Git worktrees. The test-author's prompt explicitly forbids implementation inference. The implementer's worktree has read-only access to test files.

2. **Specification as sole test-author input**: The test-author agent receives only approved specification documents (Requirements, acceptance criteria, domain glossary). No implementation code, no architecture diagrams, no existing source files beyond public interfaces.

3. **"Before" branch as handoff mechanism**: The orchestrator creates a branch containing only the new tests, verifies deterministic failure (N>=2 runs), then derives the implementer's branch from it. A `tdd_bundle.json` accompanies the handoff with test file paths, run commands, expected failures, and story references.

4. **Two-layer permission enforcement**: File system permissions (read-only on protected paths) combined with Claude Code `PreToolUse` hooks that block write attempts to forbidden file categories. Both layers are required.

5. **Event sourcing in SQLite as the system of record**: Every test run, file write, and commit attempt is recorded as an immutable, append-only event. The event schema captures `TestRunResult`, `FileWriteAttempt`, `SessionStart`, `CommitRequested`, and `CycleMarker` events with structured JSON payloads.

6. **Causal trace validation before commit acceptance**: The CLI queries the SQLite event stream to prove: (a) a failing test run targeting the story's tests preceded the first production file write, (b) a passing test run followed the last production file write, (c) no test files were modified by the implementer.

7. **Orchestrator gate as authoritative enforcement**: The orchestrator (not local hooks alone) is the authoritative validation point. Local hooks provide fast feedback; the orchestrator gate provides trust. Remote `pre-receive` hooks add optional defense-in-depth.

8. **Defense-in-depth against test hacking**: Four layers -- permission boundaries, trace validation, mutation testing (incremental per-story, full on schedule), and behavioral anomaly detection.

9. **Mutation score as test quality metric**: Coverage is a floor metric; mutation score is the quality gate. Incremental mutation testing runs per-story on affected modules with a configurable threshold (starting point: 85% killed mutants). Surviving mutant reports route back to the test-author for assertion strengthening.

10. **Conditional refactor phase triggered by static analysis**: Lint/format is always mandatory. The dedicated Refactor Agent is invoked only when cyclomatic complexity of new code exceeds a configurable threshold. The Refactor Agent operates under the same test-contract constraints with automatic rollback on regression.

11. **JUnit XML as the universal test result format**: All test runners must produce JUnit XML (or equivalent structured output). TAP is the fallback. The CLI parser layer is versioned, tested with synthetic failure injection, and treats parser errors as trust violations (default to Red).

12. **Flakiness as a handoff blocker**: The test-author must demonstrate deterministic test failure before the handoff is accepted. Flaky tests are treated as test-author bugs, not pipeline issues. Minimum two identical failure runs required.
