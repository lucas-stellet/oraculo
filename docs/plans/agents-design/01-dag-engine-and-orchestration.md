# Research Prompt 1: DAG Engine and Orchestration Mechanics

## Context: Oraculo Agent Philosophy

Oraculo's orchestrator plans and delegates but never executes. It decomposes work into a DAG (Directed Acyclic Graph), dispatches unblocked nodes in parallel, mutates the graph dynamically when agents encounter obstacles, and throttles dispatch to match QA throughput.

Key beliefs from the philosophy:
- The execution plan is a computable graph, not a text list
- The orchestrator evaluates graph state continuously, dispatching all nodes with in-degree zero
- When an agent hits an obstacle, the orchestrator prunes invalid branches and schedules new nodes without regenerating the entire plan
- Code generation is subordinated to QA throughput (Drum-Buffer-Rope)
- Squads are limited to 4-5 agents
- The orchestrator re-plans later stages concurrently while current stages execute

Architecture constraints:
- CLI (Go binary) is the Trust Layer -- deterministic, validated, contract-based
- SQLite stores the full journey (event sourcing, append-only)
- Agents are Claude Code sub-agents spawned via Task tool with worktree isolation
- The orchestrator itself is a Claude Code agent guided by skills

## Research Mission

Design the DAG engine that powers Oraculo's orchestration. I need to understand how to represent, evaluate, mutate, and persist execution graphs in a system where the orchestrator is an LLM agent coordinating other LLM agents via CLI.

## Guiding Questions

1. **Graph representation**: How should the DAG be represented for an LLM orchestrator? JSON schema? Adjacency list in SQLite? What structure lets the LLM reason about the graph while the CLI validates it deterministically? How do existing tools (LangGraph, Temporal, Prefect) represent their execution graphs?

2. **Dispatch algorithm**: How to implement the "dispatch all nodes with in-degree zero" loop? How does the orchestrator track node states (pending/running/completed/failed)? How to handle the case where multiple nodes become unblocked simultaneously but squad size limits apply?

3. **Dynamic mutation**: How to implement branch pruning and node insertion at runtime without corrupting the graph? What invariants must be preserved during mutation? How do systems like the Routine Framework handle coordinate-based step numbering for branches?

4. **Throughput control (Drum-Buffer-Rope)**: How to implement the feedback loop where QA throughput throttles dispatch? What signals does the orchestrator monitor? How to size the buffer? How to detect when the bottleneck shifts (e.g., from QA to a long-running research task)?

5. **Persistence and recovery**: How to persist DAG state in SQLite so that if the orchestrator crashes, it can reconstruct exact state and resume? Event sourcing for graph mutations? What's the schema for storing nodes, edges, state transitions, and checkpoints?

6. **Concurrent re-planning**: How to re-plan downstream nodes while upstream nodes are still executing? How to avoid conflicts between the executing plan and the re-planned version? What triggers re-planning?

## Expected Output

For each finding, provide:
- **Pattern/Concept**: Name
- **Source**: URL, paper, repository, or documentation
- **Summary**: 2-3 sentences describing the implementation approach
- **Applicability to Oraculo**: How it maps to an LLM orchestrator using CLI + SQLite + Claude Code sub-agents
- **Key design decision**: The one implementation choice this informs

Focus on implementations that work with LLM-based orchestrators (2024-2026). Include relevant patterns from workflow engines (Temporal, Prefect, Airflow) and agent frameworks (LangGraph, CrewAI) that translate to Oraculo's architecture.

## Reference Implementations to Study

Study these projects for concrete orchestration patterns:

- **Gastown** (https://github.com/steveyegge/gastown) -- Multi-agent workspace manager for Claude Code. Uses a "Mayor" orchestrator that coordinates worker agents with persistent identity but ephemeral sessions. Git-backed issue tracking with structured work state. Relevant for: dispatch patterns, agent lifecycle, workspace-level orchestration.

- **GSD / Get Shit Done** (https://github.com/gsd-build/get-shit-done) -- Spec-driven development system for Claude Code and OpenCode. Features meta-prompting, context engineering, and subagent orchestration with state management. `/gsd:map-codebase` spawns parallel agents for codebase analysis. Relevant for: spec-to-DAG decomposition, parallel agent dispatch, state management patterns.

- **BMAD-METHOD** (https://github.com/bmad-code-org/BMAD-METHOD) -- Breakthrough Method for Agile AI-Driven Development. Complete agile framework with agent roles, workflows invocable via slash commands, and structured development phases. Relevant for: role-based agent orchestration, workflow templates, phase management.
