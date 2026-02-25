# Agents Design — Code Agent

## 1. Single Agent, Loaded Skills

Oraculo uses a single code agent type. There is no separate test-author agent or implementer agent — a single agent handles both testing and implementation, guided by the TDD skill.

The orchestrator spawns a code agent for each task and loads the appropriate skills based on the task's nature. The skill defines the agent's workflow, constraints, and quality gates. The agent follows the skill's instructions — it does not invent its own process.

This design is simpler than multi-agent TDD pipelines (separate test-author + implementer) while preserving the core discipline: tests are written before implementation, and the agent cannot skip steps.

## 2. TDD via Skill

The TDD skill enforces the red-green-refactor loop:

**Red:** Write a failing test that captures the requirement. Run it. Observe the failure. The test must fail for the right reason — not because of a syntax error or missing import, but because the feature does not exist yet.

**Green:** Write the minimum implementation to make the test pass. Run the test again. Observe the pass. No more code than necessary.

**Refactor:** Clean up the implementation without changing behavior. Run tests again to confirm nothing broke.

The skill instructions are explicit about what the agent must not do:
- Do not write implementation before the test fails
- Do not weaken assertions to make tests pass
- Do not skip the refactor phase
- Do not modify test assertions after writing implementation

These constraints are enforced by the skill's workflow — the agent receives step-by-step instructions that make skipping phases unnatural. The CLI can further validate the TDD trace by checking that test failures precede implementation changes in the commit history.

## 3. Context the Agent Receives

Each code agent receives a focused context bundle assembled by the orchestrator:

- **Task description:** What to implement, acceptance criteria, expected behavior
- **Relevant files:** Only the files the agent needs to read or modify — not the entire repository
- **Project conventions:** From CLAUDE.md — code style, patterns, architectural decisions
- **Test context:** Existing tests in the area, testing patterns used in the project
- **Story/epic context:** The broader requirements document for understanding intent

**Less is more.** Agents perform better with focused context than with a full repository dump. The orchestrator curates this context based on the task's scope and the files it will touch.

## 4. Scope Boundaries

The agent's scope is defined by the task description and skill instructions — not by filesystem ACLs or permission systems:

- The task description specifies which files to modify
- The skill instructions define the workflow (TDD, frontend-design, etc.)
- The orchestrator ensures no two concurrent agents touch the same files through DAG dependencies

Since all agents work on the same branch in the same directory, scope is a convention enforced by instructions rather than a physical isolation boundary. This is simpler and sufficient when the orchestrator correctly sequences tasks that share files.

## 5. On QA Rejection

When QA rejects a task, the orchestrator does **not** send the rejection back to the same agent. Instead:

1. The original agent's session is terminated
2. A **new** code agent is spawned with:
   - The original task description
   - QA's structured findings (what failed, what was wrong, what to fix)
   - Fresh context — no memory of the previous attempt
3. The new agent follows the same TDD skill from scratch

**Why a new agent?** Agents that receive rejection feedback on their own work tend to defend their previous decisions rather than reconsidering from first principles. A fresh agent with QA's findings treats the feedback as ground truth, not as a challenge to its prior reasoning.

This pattern continues until:
- The task passes QA, or
- The circuit breaker trips after N failures and escalates to the human

## 6. What the Code Agent Does Not Do

- **Does not dispatch other agents.** Only the orchestrator dispatches.
- **Does not write to SQLite.** Only the CLI writes operational state.
- **Does not choose its own skills.** The orchestrator assigns skills.
- **Does not communicate with other agents.** Each agent works independently. Coordination happens through the orchestrator and the DAG.
- **Does not merge code.** Integration is handled by the orchestrator and CLI.
