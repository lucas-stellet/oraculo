# Research Prompt 2: Task Decomposition and Parallel Execution

## Context: The Oraculo System

Oraculo is a Socratic guide and team orchestrator for quality product development. It operates through a 5-phase model: Discover > Plan > Execute > Validate > Deliver.

Core principles:
1. Ask before doing — No action without deep understanding of the problem
2. Orchestrate, never execute — The main agent only delegates; it never writes code directly
3. Maximize parallelism — All independent tasks run simultaneously, modeled as a DAG (Directed Acyclic Graph)
4. Quality over speed — Every line of code follows TDD and project standards

Architecture:
- A CLI serves as the "Trust Layer" — deterministic, validated, contract-based (Design by Contract). Agents call the CLI and trust its results completely.
- Skills guide Socratic exploration (Epic for discovery, Story for refinement)
- Teams of specialized agents execute work: code agents, QA agents, research agents
- A dedicated QA agent independently validates all agent output
- SQLite stores the full journey; Markdown is generated only as final summaries

The main agent (Oraculo) acts as team lead: it receives project context, assembles agent teams, distributes tasks following the DAG, monitors progress, and coordinates validation. It never executes directly.

## Research Mission

Oraculo decomposes work into a DAG (Directed Acyclic Graph) where independent tasks run in parallel and dependent tasks respect ordering. I need to understand how modern agent systems handle task decomposition, dependency modeling, and parallel execution.

## Guiding Questions

1. How do multi-agent systems decompose high-level goals into executable subtasks? What decomposition strategies exist (hierarchical, recursive, LLM-planned)?
2. How is the dependency graph (DAG) constructed? Is it planned upfront, discovered dynamically, or a hybrid?
3. How do systems maximize parallelism while respecting dependencies? Scheduling algorithms, resource contention, bottleneck identification.
4. How does Theory of Constraints (TOC) or similar bottleneck-first thinking apply to agent task scheduling?
5. What happens when a parallel branch fails? Rollback strategies, partial completion handling, re-planning.

## Expected Output

For each finding, provide:
- **Title**: Name of the pattern, algorithm, or framework feature
- **Source**: URL, paper, or repository
- **Summary**: 2-3 sentences
- **Relevance to Oraculo**: How it connects to DAG-based execution, Theory of Constraints, and parallel agent teams
- **Key insight**: The one takeaway

Focus on LLM-based agent systems (2025-2026). Also include any relevant patterns from workflow orchestration (Temporal, Prefect, Airflow) that transfer to agent orchestration.
