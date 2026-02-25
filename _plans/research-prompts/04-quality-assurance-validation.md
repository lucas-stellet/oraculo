# Research Prompt 4: Quality Assurance and Validation in Multi-Agent Systems

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

Oraculo uses an independent QA agent that reviews all implementation with "fresh eyes, no bias." I need to understand how multi-agent systems handle quality assurance, output validation, and review chains.

## Guiding Questions

1. What patterns exist for validating agent output? (independent review agent, consensus voting, self-critique, debate)
2. How do systems implement "review chains" where output passes through validation before being accepted?
3. What is the role of "critic" or "adversarial" agents? How effective are they at catching errors?
4. How do systems handle the QA agent's own failures? Who watches the watchmen?
5. What metrics or signals do systems use to determine if agent output meets quality standards?

## Expected Output

For each finding, provide:
- **Title**: Name of the QA pattern, validation framework, or concept
- **Source**: URL, paper, or repository
- **Summary**: 2-3 sentences
- **Relevance to Oraculo**: How it connects to independent QA validation, TDD-first execution, and the Validate phase
- **Key insight**: The one takeaway

Focus on implementations where quality validation is a separate, independent step — not self-review by the same agent. Prioritize patterns that align with Oraculo's "dedicated QA agent" model.
