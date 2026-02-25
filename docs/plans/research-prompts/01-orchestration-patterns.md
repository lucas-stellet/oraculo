# Research Prompt 1: Multi-Agent Orchestration Patterns

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

I'm building the agent orchestration layer for Oraculo. I need to understand how modern multi-agent frameworks handle the "orchestrator" pattern — where one agent coordinates a team of specialized agents without executing work itself.

## Guiding Questions

1. What are the established patterns for a central orchestrator delegating to specialized agents? (e.g., hierarchical, flat, hub-and-spoke, blackboard)
2. How do frameworks like CrewAI, AutoGen, LangGraph, OpenAI Swarm, and Claude's multi-agent patterns handle delegation and coordination?
3. What are the failure modes of multi-agent orchestration? How do systems handle agent failure, timeout, or unexpected output?
4. How does the orchestrator decide which agent gets which task? Role assignment, capability matching, dynamic vs. static teams.
5. What papers, blog posts, or repositories document real-world orchestration patterns for LLM-based agents?

## Expected Output

For each finding, provide:
- **Title**: Name of the framework, pattern, or concept
- **Source**: URL, paper title, or repository link
- **Summary**: 2-3 sentences describing what it is
- **Relevance to Oraculo**: How it connects to Oraculo's orchestration model (delegate-only, DAG-based, quality-first)
- **Key insight**: The one thing worth adopting or studying further

Focus on practical, modern implementations (2025-2026). Prioritize patterns that align with Oraculo's "orchestrate, never execute" principle.
