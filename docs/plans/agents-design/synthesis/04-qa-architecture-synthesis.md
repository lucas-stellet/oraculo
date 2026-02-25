# Synthesis 04: QA Architecture and Validation Pipeline

## Overview

This synthesis evaluates the research document "QA Architecture and Validation Pipeline: Designing the Immune System for Autonomous Engineering" against the six guiding questions from [Research Prompt 04](../04-qa-architecture-and-validation-pipeline.md). It cross-references findings from complementary research on TDD pipelines, execution environments, and resilient orchestration to produce actionable design decisions for Oraculo.

**Primary research:** "QA Architecture and Validation Pipeline: Designing the Immune System for Autonomous Engineering" -- a comprehensive, dedicated document addressing all six guiding questions.

**Complementary research:**
- "Architecting the Agentic TDD Pipeline" -- adds depth on test-author/implementer separation, mutation testing feedback loops, and TDD trace verification
- "Execution Environments and Isolation Boundaries" -- reinforces worktree isolation, MVI principles, and permission enforcement
- "Resilient Agent Orchestration" -- provides circuit breaker patterns, checkpointing, and failure recovery mechanisms relevant to validation pipeline resilience

**Research quality assessment:** The primary research document is comprehensive and strongly aligned with Oraculo's philosophy. It addresses all six guiding questions with concrete patterns, cites relevant reference implementations (BMAD-METHOD, spec-workflow-mcp), and consistently anchors recommendations in the CLI Trust Layer as final arbiter. Its most valuable contribution is the "Executable Proof" pattern for adversarial testing, which eliminates false positives architecturally.

---

## Guiding Question 1: QA Agent Types and Responsibilities

### Guiding Question

What distinct QA roles are needed? How many QA agents per code agent? What is the optimal review chain topology -- sequential, parallel, or hybrid?

### Evaluation

**Well-covered.** The primary research provides a clear four-role decomposition and a well-reasoned hybrid parallel-convergent topology. The BMAD-METHOD reference validates the persona-based approach with production evidence. The TDD pipeline research adds the Refactor Agent and Behavioral Anomaly Detector as additional QA functions.

### Synthesized Answer

**Four specialized QA personas per code agent:**

| Persona | Phase | Responsibility | Output |
|---|---|---|---|
| Test Architect | Pre-execution | Risk-based test strategy from PRD/architecture; boundary conditions matrix; adversarial focus scoping | Test Strategy Document |
| Functional QA Reviewer | Post-execution | Verify code against PRD and acceptance criteria; generate unit/integration tests | Functional Test Suite |
| Adversarial Security Auditor | Post-execution | Actively attempt to break code: race conditions, edge cases, security flaws | Adversarial Test Suite (executable scripts only) |
| Style/Convention Checker | Post-execution | Enforce project coding standards, documentation, dependency hygiene | Linting Report |

**Two additional QA functions from complementary research:**

| Function | Type | Responsibility |
|---|---|---|
| Refactor Agent | Agent (conditional) | Non-behavioral code improvements after Green state, triggered by cyclomatic complexity > threshold |
| Behavioral Anomaly Detector | CLI-based (automated) | Monitor SQLite event stream for suspicious patterns (brute-force test execution, prompt injection attempts) |

**Topology: Hybrid Parallel-Convergent**

```
Phase 1 (Sequential):  Test Architect --> Test Strategy + Adversarial Focus Matrix
Phase 2 (Sequential):  Code Agent --> Implementation + Basic Tests
Phase 3 (Parallel):    Functional Reviewer | Adversarial Auditor | Style Checker
Phase 4 (Convergent):  CLI Trust Layer --> Unified Execution --> Mutation Gate --> Verdict
Phase 5 (Conditional): Refactor Agent (if complexity threshold breached)
```

The three post-execution QA agents run in parallel within isolated worktrees, generating independent test suites. Filesystem isolation prevents collisions. The CLI Trust Layer then converges their outputs by executing all test suites deterministically.

### Gaps

- **Conditional activation criteria.** The research does not discuss scaling QA depth based on task risk. A simple style-only change may not justify a full adversarial audit. The TDD pipeline research suggests using cyclomatic complexity as a trigger for optional phases -- this same principle should apply to QA persona activation.
- **Optimal agent ratio.** The research states "four distinct QA roles per code generation agent" but does not analyze whether this 4:1 ratio is practical or whether certain ratios produce better cost/quality tradeoffs.

### Recommendation

Adopt the four-persona model but make the Adversarial Auditor and Test Architect **conditionally activated** based on a risk score assigned during the Plan phase. The Functional Reviewer and Style Checker always run. This avoids burning tokens on adversarial fuzzing for trivial changes. The risk score should be computed from task metadata: files changed, modules affected, whether security-sensitive code is touched, and complexity of the change.

---

## Guiding Question 2: Clean Context Enforcement

### Guiding Question

How to guarantee the QA agent has no context from the generation process? Is a separate worktree necessary but sufficient? What context produces the best review quality?

### Evaluation

**Well-covered.** The primary research provides strong architectural reasoning for worktree isolation and defines a precise "Bounded Context Payload" strategy. The execution environments research reinforces this with MVI (Minimal Viable Information) principles, the wipe-on-fail pattern, and AST-based context construction. The communication/handoff research adds the critical distinction between verbatim artifacts and summarized procedural history.

### Synthesized Answer

**Isolation mechanism: Ephemeral Git worktrees**

Each QA persona is spawned as a fresh Claude Code sub-agent (`isolation: "worktree"`) in a dedicated directory (e.g., `.claude/worktrees/qa-functional-<task-id>/`). The agent has zero access to:
- The code agent's conversational history
- The code agent's scratchpad or intermediate reasoning
- Any other QA agent's worktree

**Bounded Context Payload (exactly three components):**

1. **Immutable Specification** -- Relevant PRD/story sections defining what the code should do (not what it does). Retrieved from the SQLite blackboard, linked via task ID.
2. **Raw Git Diff** -- Output of `git diff main...feature/<branch>`, forcing attention on changed lines only.
3. **Test Runner Logs** -- Deterministic CLI output from existing test execution in JUnit XML format (pass/fail traces with assertion messages).

The agent receives no explanation of implementation choices. If broader context is needed, it has read-only access to the local worktree, but its initial context remains aggressively minimized to maximize token efficiency and analytical sharpness.

**Worktree lifecycle:** Provisioned at validation start, destroyed immediately after the QA agent returns its structured report. On failure-and-retry, the worktree is completely wiped and rebuilt from scratch -- never reused. This mirrors CI/CD ephemeral runner patterns and prevents state corruption.

**Dual-layer permission enforcement:** (a) PreToolUse hooks intercept all filesystem modifications and validate target paths against a whitelist, returning Exit 2 to block unauthorized access; (b) filesystem-level read-only permissions on protected paths as a defense-in-depth measure.

### Cross-reference with Philosophy

This directly implements: "Independence is architectural: clean context, no access to code agent's reasoning." The research correctly identifies that worktree isolation is necessary but not sufficient -- the bounded payload is equally important.

### Gaps

- **CLI implementation of context assembly.** The research does not specify the exact CLI command or algorithm for constructing the bounded payload. The execution environments research fills part of this gap with AST-based topological graphing (Tree-sitter + PageRank), but that mechanism is designed for code agents, not QA reviewers. QA agents may need different context composition (e.g., more weight on spec sections, less on implementation details).

### Recommendation

The CLI must implement a `build-qa-context` command that:
1. Extracts the diff from the feature branch
2. Resolves the relevant spec sections from the SQLite blackboard (linked via task ID in the DAG)
3. Runs existing tests and captures JUnit XML output
4. Assembles these three components into the QA agent's initial prompt -- nothing more

This command is deterministic and produces identical output for identical inputs, ensuring reproducible validation runs.

---

## Guiding Question 3: Adversarial Testing Implementation

### Guiding Question

How to build an agent that actively tries to break code? What prompt patterns produce effective adversarial behavior? How to prevent false positives?

### Evaluation

**Well-covered with a critical architectural insight.** This was previously the largest gap in the research corpus. The primary research document provides four concrete adversarial prompt patterns and, most importantly, the "Executable Proof" pattern that solves the false positive problem architecturally. The TDD pipeline research complements this with the test-author's implicit adversarial role and behavioral anomaly detection.

### Synthesized Answer

**Adversarial prompt patterns (four categories):**

1. **State Manipulation / Environmental Chaos** -- Evaluate code under extreme, out-of-distribution conditions. Prompt template: "Assume the database connection drops halfway through this transaction. Write a test that simulates a network timeout during the execution of lines 45-50. Does the system handle the rollback gracefully, or does it leave orphaned records?"

2. **Authority Spoofing / Trust Boundary Testing** -- Search for blind trust within the code. Inject mock payloads simulating elevated privileges, attempt to add comments like `// Audited by AppSec` or mock tokens, and verify that access controls enforce regardless of superficial markers.

3. **Round-Trip Review Protocol (RTRP) / Semantic Drift Detection** -- Give the adversarial agent the implementation without the PRD and ask it to reverse-engineer the specification. If the reverse-engineered spec diverges from the actual PRD (measured via similarity scoring), flag a conceptual vulnerability where the code subtly deviates from intent.

4. **Concurrency and Fuzzing Injections** -- Write test harnesses that intentionally bombard target functions with null bytes, integer overflows, maximum-length strings, and asynchronous race conditions. Force the system to handle concurrent thread overlapping and deadlock scenarios.

**False positive elimination: Executable Proof Only**

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

**Complementary adversarial mechanisms from TDD research:**

- **Test-author as implicit adversary:** The test-author agent's prompt includes a blocking instruction: "Do not infer implementation; describe only observable contracts and cases, as if you were a consumer of the API." This makes the test-author inherently adversarial toward implementation shortcuts.
- **LLM-guided mutation testing:** For periodic builds, use LLMs to generate semantically complex mutants that mimic realistic developer errors, beyond simple AST operator swaps. These serve as an advanced adversarial layer.
- **Behavioral anomaly detection:** The CLI monitors the SQLite event stream for suspicious patterns (e.g., hundreds of test executions per minute without code modifications), identifying brute-force attempts or side-channel test inference.

### Cross-reference with Philosophy

Perfectly aligned with: "The CLI verifies, not the QA agent -- deterministic reality over probabilistic opinion." The executable proof pattern is the concrete, operational implementation of this principle. It also implements "effective validation needs orthogonal expertise" by separating the agent's analytical capability from the CLI's verification authority.

### Gaps

- **Scoping the adversarial surface.** Without boundaries, the adversarial agent could attempt to test every conceivable attack vector, consuming massive tokens for diminishing returns. The agent needs guidance on which attack categories to prioritize based on the nature of the change.
- **Adversarial prompt library versioning.** The four patterns described are a starting point, but the library needs to evolve as new attack vectors emerge. No mechanism for versioning and extending the prompt library is described.

### Recommendation

The Test Architect's pre-execution output should include an **adversarial focus matrix** that scopes the Adversarial Auditor's effort:

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

The prompt pattern library should be stored as a versioned reference document within the CLI configuration, allowing teams to extend it with domain-specific patterns.

---

## Guiding Question 4: Mutation Testing Integration

### Guiding Question

How to automate mutation testing? Which tools for which languages? How to integrate mutation results into accept/reject? What threshold?

### Evaluation

**Well-covered across both the primary research and the TDD pipeline research.** The two documents converge on mutation testing as the "validator of validators" and provide concrete tool recommendations, threshold enforcement, and feedback loop mechanics. The TDD pipeline research adds practical considerations on incremental mode and performance.

### Synthesized Answer

**Tool mapping by language:**

| Language | Tool | Integration |
|---|---|---|
| JavaScript/TypeScript | Stryker Mutator | CLI invokes `npx stryker run` post-test-pass |
| Python | Mutmut | CLI invokes `mutmut run` post-test-pass |
| Java/Kotlin | PIT (PITest) | CLI invokes via Maven/Gradle plugin |
| .NET | Stryker.NET | Via dotnet CLI |
| Go | go-mutesting | Via go CLI |

**Mutation score formula and threshold:**

$$Mutation\ Score = \frac{Killed\ Mutants}{Total\ Mutants} \times 100$$

- **Standard threshold:** >= 80% (philosophy mandate)
- **High-risk threshold:** >= 90% (for security-sensitive or core-logic changes, determined by risk classification)

**Integration into the validation pipeline (Gate 2 and Gate 3 of the sub-DAG):**

```
All QA-generated tests pass (Gate 1)
  --> CLI runs mutation framework, scoped to the diff
  --> Framework mutates implementation code (e.g., `a + b` --> `a - b`)
  --> CLI re-runs QA test suites against mutated code
  --> Calculate mutation score
  --> Score >= threshold? --> APPROVED
  --> Score < threshold?  --> REJECTED
    --> CLI parses surviving mutant report
    --> Extracts exact lines and mutation operators that survived
    --> Feeds surviving mutants back to QA agents with targeted prompt:
        "Your tests passed, but when line 45 changed from
        `if (x && y)` to `if (x || y)`, your tests still passed.
        Write new assertions to kill this surviving mutant."
    --> QA agent iterates until threshold met or retry limit reached
```

**Key insight from TDD research:** Mutation testing validates the test-author, not the implementer. A low mutation score is treated as a failure of the test suite. The feedback loop routes surviving mutant reports to the QA agents (who wrote the tests), not to the code agent.

**Performance optimization:** Run mutation testing in incremental mode, scoped to modules touched by the change (`--mutate` flag in Stryker, `--paths-to-mutate` in Mutmut). Full mutation runs across the entire codebase are reserved for periodic nightly/weekly builds.

**Advanced: LLM-guided mutants.** For nightly builds or high-risk architectural validations, use LLMs to generate semantically complex mutants that mimic realistic developer errors, beyond simple AST operator swaps. Standard deterministic tools remain the primary gatekeepers for fast per-story loops.

### Cross-reference with Philosophy

The mutation score threshold is enforced mechanically by the CLI, not by agent judgment. The append-only SQLite log records every mutation score for auditability. This directly implements "deterministic reality over probabilistic opinion" -- the CLI calculates the score and makes the pass/reject decision with zero LLM involvement.

### Gaps

- **Mutation tool output standardization.** JUnit XML is standardized for test results, but mutation tool output formats vary across tools (HTML reports, JSON, custom formats). The CLI needs an adapter layer or a universal format for parsing mutation results.
- **Threshold calibration.** The 80% default and 90% high-risk thresholds lack empirical validation against agent-generated code. Real-world benchmarking is needed.

### Recommendation

1. Configure mutation tools to operate only on the diff for per-story runs.
2. Store mutation scores per-task in SQLite for trend analysis and threshold calibration over time.
3. Use a two-tier threshold: 80% standard, 90% high-risk (determined by Plan phase risk classification).
4. Build a mutation report parser in the CLI that normalizes output from Stryker, PIT, and Mutmut into a common JSON format for consistent processing.

---

## Guiding Question 5: Heterogeneous Model Selection

### Guiding Question

How to choose which model for which role? Static or dynamic selection? What combinations reduce shared bias?

### Evaluation

**Adequately covered with a strong theoretical foundation.** The primary research provides a concrete model-to-role mapping, cites empirical evidence (X-MAS-Bench) that heterogeneous configurations outperform homogeneous ones, and proposes an elegant disagreement resolution mechanism. However, implementation specifics (configuration schema, cost management, provider fallback chains) remain underspecified. The communication/handoff research adds the three-tier model assignment pattern.

### Synthesized Answer

**Static role-to-model mapping (configured in orchestrator manifest):**

| Role | Recommended Model Family | Rationale |
|---|---|---|
| Code Generator | Claude (Anthropic) | Strong coding benchmarks, architectural fidelity |
| Functional QA Reviewer | Gemini (Google) | Large context window, strong cross-referencing |
| Adversarial Auditor | DeepSeek R1 / open-weight model | Different training data, different alignment guardrails, structurally distinct |
| Style/Convention Checker | Claude Haiku (Anthropic) | Lightweight, cost-effective for formatting tasks |
| Orchestrator | Claude (Anthropic) | Reasoning and planning capability |
| Summarizer Agent | Claude Haiku (Anthropic) | Fast, cost-effective for compression tasks |
| Debugger Agent | High-capability model (varies) | Needs advanced reasoning for root-cause analysis |

**Empirical basis:** The X-MAS-Bench study confirms that heterogeneous multi-agent configurations consistently outperform homogeneous ones by exploiting complementary strengths across domains. Models exhibit "selection bias," "token bias," and "conformity bias" -- using models from the same lineage for both generation and review creates false consensus.

**Disagreement resolution: CLI as final arbiter (Consensus via Cryptographic Reality)**

When heterogeneous models disagree (e.g., Gemini approves but DeepSeek flags a race condition), the system does **not** attempt LLM consensus through debate, as this leads to the stronger model bullying the weaker one into conformity. Instead:

1. The dissenting model must produce an executable test case
2. The CLI runs the test
3. If the test induces a failure: the dissenting model wins, regardless of the approving model's verdict
4. If the test passes: the approving model's verdict stands

This reuses the same "Executable Proof" pattern from adversarial testing, creating architectural consistency across the system.

**Hard architectural constraint:** QA agents must use a different model family from the code agents they validate. This is non-negotiable and enforced by the orchestrator configuration.

### Cross-reference with Philosophy

Directly implements: "Heterogeneous models across roles prevent shared training biases." The CLI arbitration pattern extends the Trust Layer philosophy to inter-model disputes, maintaining the principle that deterministic test execution supersedes probabilistic opinion in all cases.

### Gaps

- **Configuration schema.** No concrete YAML/JSON schema for specifying model-to-role mappings with fallback chains is defined.
- **Dynamic model selection.** The research mentions "dynamic fallback based on task complexity" but provides no algorithm or criteria for when to upgrade or downgrade model capability.
- **Cost management.** Different models have vastly different token costs. No guidance on per-role budgets or cost optimization strategies.
- **Provider availability.** No discussion of fallback chains when a provider API is unavailable or rate-limited.

### Recommendation

Define a model configuration schema in the orchestrator manifest:

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

Model selection should be **static by default** with provider-availability fallback. Dynamic selection based on task complexity is deferred to a future iteration -- it adds system complexity without clear evidence of ROI at the current stage of maturity.

---

## Guiding Question 6: Validation Pipeline Orchestration

### Guiding Question

How does the validation pipeline fit into the DAG? Is it a sub-DAG within each Validate node? What is the flow? How to handle partial approval?

### Evaluation

**Well-covered with strong structural clarity.** The primary research defines a precise sub-DAG state machine with three explicit gates, a no-partial-approval policy, and immutable SQLite logging. The spec-workflow-mcp reference validates the approval gate pattern. The resilient orchestration research adds critical depth on circuit breakers, checkpointing, and revision loop management. The TDD pipeline research contributes the Red-Green trace verification as a prerequisite gate.

### Synthesized Answer

**Validation sub-DAG (within each Validate node of the main DAG):**

```
Code Agent Completes (commits to feature branch)
  |
  v
Gate 0: TDD Trace Verification (cheapest, runs first)
  - CLI queries SQLite event store for T1 (Red) -> T2 (Code) -> T3 (Green) sequence
  - If trace not proven --> IMMEDIATE REJECTION (protocol violation, not a quality issue)
  |
  v
Gate 1: Parallel QA Provisioning + Test Generation
  - CLI creates isolated Git worktrees
  - Provisions Functional QA (Model A), Adversarial QA (Model B), Style Checker (Model C)
  - All three consume PRD + Git diff in parallel
  - Functional agent --> Functional.test
  - Adversarial agent --> Adversarial.test (executable scripts only)
  - Style checker --> Linting report
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
  - DAG progresses to Deliver (or conditional Refactor phase if complexity threshold breached)
```

**No partial approval.** The primary research is explicit: "The Oraculo architecture strictly prohibits 'partial approval.'" If functional tests pass but the adversarial agent uncovers a valid edge case (proven by executable test failure), the **entire validation node fails**. The code agent must fix the edge case, and then the complete validation sub-DAG re-runs from scratch to catch regressions introduced by the fix.

**Revision cycles within the DAG.** From the communication/handoff research: "Instead of moving backward along a directed edge -- which violates the mathematical acyclic properties of a DAG -- the orchestrator dynamically appends a new sequence of 'Revision -> Review' nodes to the graph, preserving the complete, append-only history of the iteration."

**Immutable logging.** Every validation event is logged to SQLite (append-only):
- Worktree provisioning events (directory path, branch, timestamp)
- Model selection per persona (model ID, provider)
- Generated test file paths
- CLI exit codes for each test execution
- Mutation score percentages
- Final accept/reject verdict with justification

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

The LLM is completely removed from the final authorization decision. The CLI queries the SQLite ledger and makes the determination based purely on recorded deterministic outcomes.

**Circuit breaker integration (from resilient orchestration research):** A maximum of 3 full validation cycles is enforced per task. After 3 rejections, the circuit breaker trips: the DAG branch is paused, a diagnostic summary is logged to SQLite, and the system escalates (either to a human operator or to a re-planning phase where the task is decomposed differently).

### Cross-reference with Philosophy

This is the concrete implementation of: "No DAG edge connects Execute to Deliver without passing through Validate." The no-partial-approval policy, SQL-based verdict, and append-only logging directly express the Trust Layer philosophy. The revision-as-append pattern preserves the DAG's mathematical acyclic properties while supporting iterative improvement.

### Gaps

- **Re-validation scope optimization.** The mandate for full re-validation after any fix is correct for safety but expensive. The research does not analyze whether certain fix categories (e.g., purely additive changes that cannot cause regressions) could safely use incremental re-validation.
- **Human escalation protocol.** When the circuit breaker trips, the concrete protocol (notification channels, diagnostic summary format, re-entry flow) needs definition.
- **Feedback routing on rejection.** When QA rejects, the system needs logic to determine whether the failure is a code agent problem (implementation bug), a test-author problem (weak tests with low mutation score), or both. This routing logic is not specified in the research.

### Recommendation

1. Always re-run the full validation sub-DAG after any fix. Incremental re-validation introduces risk of regression masking. The cost is justified by the safety guarantee.
2. Set a maximum of 3 full validation cycles per task (circuit breaker limit).
3. Define feedback routing rules:
   - Gate 2 failure (test execution fails) --> Route to Code Agent for implementation fix
   - Gate 3 failure (mutation score too low) --> Route to QA agents for test strengthening
   - Gate 0 failure (TDD trace violation) --> Route to Code Agent with protocol violation notice
4. Define escalation protocol: after 3 failures, pause DAG branch, log diagnostic summary (all failure logs, mutation reports, agent traces) to SQLite, notify human via configured channel.

---

## Cross-Reference with Oraculo Philosophy

| Principle | Alignment Assessment |
|-----------|---------------------|
| **Ask before doing** | The Validate phase is structurally mandatory -- no DAG edge connects Execute to Deliver without passing through Validate. QA cannot be skipped. The Test Architect operates pre-execution, asking questions before testing begins. |
| **Orchestrate, never execute** | QA agents are sub-agents delegated by the orchestrator. The orchestrator writes approval gates but never validates code itself. The CLI executes tests, not the QA agents. |
| **Maximize parallelism** | Directly addressed. The hybrid parallel-convergent topology runs Functional Reviewer, Adversarial Auditor, and Style Checker simultaneously in isolated worktrees. |
| **Quality over speed** | Mutation testing as a mandatory gate, no-partial-approval policy, and full re-validation after fixes all prioritize quality over velocity. The two-tier threshold (80%/90%) calibrates rigor to risk. |
| **Adversarial validation** | Strongly addressed. Dedicated Adversarial Auditor with four prompt patterns, executable proof requirement, and RTRP semantic drift detection. |
| **Heterogeneous models** | Addressed with concrete model-to-role mapping and empirical justification. Hard constraint: QA agents must use different model families from code agents. |
| **Clean context** | Thoroughly addressed. Worktree isolation, bounded context payload (spec + diff + test logs only), dual-layer permission enforcement, and wipe-on-fail lifecycle. |
| **CLI as Trust Layer** | The strongest alignment point. The CLI verifies TDD traces, executes test suites, calculates mutation scores, resolves inter-model disagreements via executable tests, and makes the final accept/reject decision via SQL query. No LLM has authorization authority. |
| **SQLite as append-only ledger** | Every validation event is immutably logged. Final authorization is a SQL query against the ledger. Revision cycles append new nodes rather than modifying existing ones. |

---

## Key Design Decisions Summary

1. **Four-persona QA model with conditional activation.** Always run Functional Reviewer and Style Checker. Activate Test Architect and Adversarial Auditor based on risk classification from the Plan phase. Each persona operates in its own ephemeral Git worktree.

2. **Bounded Context Payload for all QA agents.** Each QA agent receives exactly three inputs: the immutable specification (PRD/story), the raw git diff, and test runner logs in JUnit XML. No implementation rationale, no code agent reasoning, no shared context windows.

3. **Executable Proof as the sole adversarial validation mechanism.** The adversarial agent's textual claims are architecturally ignored. Only executable test scripts that compile and successfully induce a failure in the CLI runner are accepted as evidence. This eliminates false positives by design and prevents infinite revision loops.

4. **Mutation testing as the mandatory validator of validators.** The CLI enforces a mutation score threshold of >= 80% (configurable to 90% for high-risk changes). Surviving mutants are fed back to QA agents with targeted prompts in a self-healing loop. Mutation runs are scoped to the diff for performance.

5. **Heterogeneous model assignment with static configuration.** Different LLM providers are mapped to different QA roles (e.g., Claude for generation, Gemini for functional review, DeepSeek for adversarial testing). Hard constraint: QA and code agents must use different model families. Disagreements resolved by CLI-executed test cases.

6. **No partial approval in the validation sub-DAG.** Any single gate failure rejects the entire validation node. The code agent must fix the issue and the complete validation pipeline re-runs from scratch. Circuit breaker limits this to 3 cycles before escalation.

7. **SQL-based final authorization.** The CLI's accept/reject decision is a deterministic SQL query against the append-only SQLite validation log. The LLM has zero authority over the final verdict. Every validation event is immutably logged.

8. **Four-gate validation sub-DAG.** Gate 0: TDD trace verification. Gate 1: Parallel QA test generation. Gate 2: Deterministic test execution. Gate 3: Mutation score threshold. Gates are ordered by cost (cheapest first) and no gate can be skipped.

9. **Worktree lifecycle tied to validation lifecycle.** Provisioned at validation start, destroyed after the QA agent returns its report. On failure-and-retry, worktrees are completely wiped and rebuilt -- never reused.

10. **Adversarial scope controlled by Test Architect.** The Test Architect's pre-execution output includes an adversarial focus matrix that bounds the Adversarial Auditor's attack surface, preventing unbounded token consumption on low-risk changes.

11. **Feedback routing on rejection.** Gate 2 failures (test execution) route to Code Agent. Gate 3 failures (mutation score) route to QA agents. Gate 0 failures (TDD trace) route to Code Agent with protocol violation notice.

12. **Inter-model disagreements resolved by deterministic test execution.** When heterogeneous models disagree, the dissenting model must produce an executable test. The CLI runs it. The test result -- not the model's reasoning -- determines the outcome.

---

## Areas Requiring Additional Investigation

| Area | Priority | Rationale |
|---|---|---|
| Risk classification algorithm | High | Conditional QA activation depends on a risk score, but no algorithm is defined for computing it from task metadata (files changed, modules affected, security sensitivity) |
| Mutation tool output standardization | High | CLI needs a universal adapter for parsing Stryker, PIT, and Mutmut reports into a common JSON format for consistent processing |
| Human escalation protocol | Medium | Circuit breaker behavior after max retries needs concrete definition: notification channels, diagnostic summary format, re-entry flow |
| Incremental mutation testing benchmarks | Medium | Need empirical data on diff-scoped mutation performance with Stryker/Mutmut to validate that per-story runs are practical |
| Adversarial prompt library versioning | Medium | The four prompt patterns need expansion into a versioned, extensible library stored in CLI configuration |
| Dynamic model selection criteria | Low | Static assignment is sufficient for v1; criteria for when to upgrade/downgrade model capability per role can be explored in later iterations |
| Mutation score threshold calibration | Low | The 80%/90% thresholds need validation against real agent-generated code across different languages and project types |
