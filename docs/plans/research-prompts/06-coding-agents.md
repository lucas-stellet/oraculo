# Research Prompt 6: Coding Agents — Patterns, Quality, and Orchestration

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

Oraculo orchestrates coding agents that write implementation code following TDD (test-first, red-green-refactor). I need to understand the state of the art in coding agents: how they operate, what makes them effective, and how they're orchestrated for quality output.

## Guiding Questions

1. How do modern coding agents work? Architecture and patterns behind Claude Code, OpenAI Codex (and its CLI), OpenCode, Aider, SWE-agent, Devin, Cline, Windsurf, Cursor Agent. What makes each approach distinct?
2. What patterns exist for TDD with coding agents? How do agents write tests first, then implementation? What's the success rate vs. implementation-first?
3. How do coding agents handle project context? Understanding existing codebases, respecting conventions, following architectural patterns.
4. How are multiple coding agents coordinated on the same codebase? Merge conflicts, file locking, code consistency across agents.
5. What quality signals distinguish good coding agent output from bad? Metrics, validation techniques, human review patterns.

## Expected Output

For each finding, provide:
- **Title**: Name of the coding agent, pattern, or technique
- **Source**: URL, paper, or repository
- **Summary**: 2-3 sentences
- **Relevance to Oraculo**: How it connects to Oraculo's TDD-first execution, agent orchestration, CLI trust layer, and DAG-based parallel coding
- **Key insight**: The one takeaway

Focus on the latest generation of coding agents (2024-2025). Include both commercial tools and open-source implementations. Prioritize findings about quality and orchestration over raw capability benchmarks.
