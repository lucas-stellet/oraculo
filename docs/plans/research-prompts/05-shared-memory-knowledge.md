# Research Prompt 5: Shared Memory and Knowledge Accumulation

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

Oraculo stores knowledge in SQLite (patterns, conventions, constraints, dependencies, architecture findings) and accumulates wisdom across sessions. I need to understand how multi-agent systems handle shared memory, knowledge persistence, and cross-session learning.

## Guiding Questions

1. What memory architectures exist for multi-agent systems? (shared database, vector stores, knowledge graphs, episodic memory)
2. How do agents write and read from shared memory without conflicts? Concurrency patterns, memory consistency.
3. How do systems handle "accumulated wisdom" — knowledge that grows over time across sessions? What persists, what decays?
4. What's the role of structured vs. unstructured memory? When is a database better than embeddings?
5. How do systems prevent memory corruption — agents writing incorrect or contradictory knowledge?

## Expected Output

For each finding, provide:
- **Title**: Name of the memory pattern, architecture, or system
- **Source**: URL, paper, or repository
- **Summary**: 2-3 sentences
- **Relevance to Oraculo**: How it connects to SQLite-based memory, FTS5 search, domain-categorized knowledge, and the Trust Layer (CLI) as memory gateway
- **Key insight**: The one takeaway

Include both LLM-agent memory systems and relevant database/knowledge management patterns that transfer to agent architectures.
