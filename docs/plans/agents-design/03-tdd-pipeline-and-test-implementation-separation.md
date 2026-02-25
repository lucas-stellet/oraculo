# Research Prompt 3: TDD Pipeline and Test/Implementation Separation

## Context: Oraculo Agent Philosophy

TDD is the anti-hallucination mechanism. A separate agent authors tests from Epic/Story specifications. The code agent receives those tests and writes implementation to make them pass. The code agent has no write permission to test files. The CLI rejects any commit where the execution trace does not show a failing test immediately before implementation.

Key beliefs from the philosophy:
- Code agents do not write their own tests (prevents "test hacking")
- The red-green-refactor loop is enforced by the CLI, not by agent discipline
- The CLI rejects commits without proven failing-test-before-implementation trace
- TDD anchors probabilistic models to deterministic reality
- Empirical evidence: 94.3% success with ground-truth tests vs. 69.8% with self-generated

Architecture constraints:
- Test-author agent and implementation agent are separate Claude Code sub-agents
- Each operates in its own worktree
- The CLI validates the TDD trace in SQLite before accepting commits
- SQLite stores execution events (event sourcing, append-only)

## Research Mission

Design the TDD pipeline that enforces test-first development across separate agents. I need to understand how to implement the test-author/implementer separation, how to enforce the red-green-refactor trace, and how to detect test quality.

## Guiding Questions

1. **Test-author agent design**: What context does the test-author agent need? How does it translate Epic/Story specifications into test cases? What's the prompt structure? How to ensure tests are meaningful (not `assert True`)? What test granularity is optimal -- unit, integration, or acceptance?

2. **Handoff: tests to implementer**: How does the test file get from the test-author's worktree to the implementer's worktree? Does the orchestrator merge the test branch first? Does the implementer's worktree start from a branch that already contains the tests? What metadata accompanies the handoff?

3. **Red-green-refactor enforcement**: How does the CLI verify the TDD trace? What events must be recorded in SQLite? (e.g., "test run at T1: 3 failures", "implementation written at T2", "test run at T3: 0 failures"). What's the schema for these events? How to detect if an agent skipped the red phase?

4. **Test hacking detection**: Beyond permission boundaries (no write access to test files), how to detect subtler forms of test gaming? Mutation testing as automated validation? What mutation testing tools work well in CI pipelines? How to integrate mutation testing into the CLI Trust Layer?

5. **Refactor phase**: How to handle the refactor step? Does the same agent refactor, or is there a dedicated refactor pass? How to ensure refactoring doesn't break tests? Is the refactor step optional for agents (green is sufficient)?

6. **Multi-language support**: TDD enforcement must work across different programming languages and test frameworks. How to make the CLI's test-trace validation language-agnostic? Parse test runner output? Use exit codes? Structured test result formats (JUnit XML, TAP)?

## Expected Output

For each finding, provide:
- **Pattern/Concept**: Name
- **Source**: URL, paper, repository, or documentation
- **Summary**: 2-3 sentences describing the implementation approach
- **Applicability to Oraculo**: How it maps to separate test-author/implementer agents with CLI enforcement
- **Key design decision**: The one implementation choice this informs

Focus on TDD in automated/agent contexts (2024-2026). Include the TDFlow framework findings, mutation testing tools, and any CI/CD patterns that enforce test-first workflows programmatically.

## Reference Implementations to Study

Study these projects for concrete TDD and spec-driven patterns:

- **GSD / Get Shit Done** (https://github.com/gsd-build/get-shit-done) -- Spec-driven development system for Claude Code. Uses structured spec extraction before any code is written. The entire workflow is: describe idea -> extract specs -> plan -> implement. Relevant for: spec-to-test translation, enforcing spec-first discipline in agent workflows.

- **spec-workflow-mcp** (https://github.com/Pimzino/spec-workflow-mcp) -- MCP server providing structured spec-driven development workflow (Requirements -> Design -> Tasks -> Implementation). Features approval workflow with complete revision process, task progress tracking, and implementation logs with code statistics. Relevant for: structured spec-to-implementation pipeline, approval gates between phases, tracking implementation against specs.
