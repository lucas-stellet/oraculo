# Research Prompt 3: Inter-Agent Communication and Context Sharing

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

In Oraculo, each agent receives "clear context" — project patterns, existing architecture, expected tests. I need to understand how modern multi-agent systems handle communication between agents: context propagation, message passing, shared state, and handoff protocols.

## Guiding Questions

1. What communication patterns exist between agents? (message passing, shared blackboard, event-driven, direct handoff)
2. How do systems propagate context without overwhelming agents? Context windowing, summarization, selective context injection.
3. How do agents hand off work to each other? What information must be preserved during handoff?
4. How do systems handle context divergence — when parallel agents develop contradictory understanding of the same codebase?
5. What are the costs and trade-offs of different communication patterns? (latency, token cost, accuracy)

## Expected Output

For each finding, provide:
- **Title**: Name of the communication pattern or protocol
- **Source**: URL, paper, or repository
- **Summary**: 2-3 sentences
- **Relevance to Oraculo**: How it connects to Oraculo's context-sharing model (CLI as single source of truth, structured JSON contracts, project context injection)
- **Key insight**: The one takeaway

Focus on practical implementations. Include both LLM-agent-specific patterns and relevant patterns from distributed systems that transfer well.
