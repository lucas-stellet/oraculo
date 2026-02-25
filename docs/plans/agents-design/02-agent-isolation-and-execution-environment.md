# Research Prompt 2: Agent Isolation and Execution Environment

## Context: Oraculo Agent Philosophy

Every code agent works in physical isolation -- a dedicated git worktree on a separate branch. No agent sees another agent's in-progress work. Code agents receive minimal viable information for their task, not the entire repository. The CLI builds the working set.

Key beliefs from the philosophy:
- Each agent operates in a dedicated git worktree on a separate branch
- Less context produces better code -- minimum viable information per task
- The CLI builds the working set for each agent
- Agents require explicit, deterministic boundaries (not implicit abstractions)
- Systems designed to eliminate friction for humans often create catastrophic friction for agents

Architecture constraints:
- Agents are Claude Code sub-agents spawned via `Task` tool with `isolation: "worktree"` parameter
- The CLI (Go binary) manages context injection and validation
- SQLite is the shared state (many readers, writes serialized through CLI)
- Git is the version control system (not Jujutsu)

## Research Mission

Design the execution environment for Oraculo's code agents. I need to understand how to create, manage, and tear down isolated agent environments, how to inject minimal context, and how to merge results back safely.

## Guiding Questions

1. **Worktree lifecycle**: How to automate git worktree creation per DAG node? Naming conventions? Branch strategies? How to clean up worktrees after completion? What happens to a worktree when its DAG node fails and needs retry?

2. **Context injection (working set builder)**: How to build the minimal context package for each agent? What should the CLI's "build working set" command return? How to determine which files, interfaces, and patterns are relevant to a specific task? How do tools like Aider (topological mapping) and OpenAgentsControl (MVI) implement this?

3. **Permission boundaries**: How to enforce that a code agent cannot write to test files? File-level permissions in the worktree? CLI-level validation on commit? Git hooks? What's the lightest mechanism that's also deterministic?

4. **Merge strategy**: How to merge a completed agent's worktree back to the main branch? How to detect and handle merge conflicts between parallel agents? Should the orchestrator merge, or should there be a dedicated merge agent? What pre-merge validation is needed?

5. **Environment setup**: How to ensure each worktree has the right dependencies, environment variables, and tooling? How to handle agents that need to run tests (which may need database connections, API keys, etc.)? Containerization vs. shared host environment?

6. **Resource limits**: How to prevent a runaway agent from consuming unlimited tokens or time? Per-agent token budgets? Timeout policies? How do these limits interact with the circuit breaker pattern?

## Expected Output

For each finding, provide:
- **Pattern/Concept**: Name
- **Source**: URL, paper, repository, or documentation
- **Summary**: 2-3 sentences describing the implementation approach
- **Applicability to Oraculo**: How it maps to Claude Code sub-agents with worktree isolation
- **Key design decision**: The one implementation choice this informs

Focus on practical implementations for multi-agent coding systems (2024-2026). Include patterns from CI/CD systems (parallel test runners, build matrices) that solve similar isolation problems.

## Reference Implementations to Study

Study these projects for concrete isolation and environment patterns:

- **Gastown** (https://github.com/steveyegge/gastown) -- Multi-agent workspace manager for Claude Code. Worker agents have persistent identity but ephemeral sessions, with git-backed state persistence. Relevant for: worktree lifecycle management, agent identity/session separation, workspace isolation patterns.

- **GSD / Get Shit Done** (https://github.com/gsd-build/get-shit-done) -- Spec-driven development system with context engineering layer. `/gsd:map-codebase` spawns parallel agents to analyze stack, architecture, conventions, and concerns -- then planning automatically loads project patterns. Relevant for: context injection, working set construction, codebase-aware agent bootstrapping.
