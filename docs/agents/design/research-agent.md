# Agents Design — Research Agent

## 1. Role

The research agent is the investigator. It analyzes the existing codebase, explores external references, and surfaces evidence that informs discovery and planning. It does not write code, run tests, or modify files — it observes, analyzes, and reports structured findings.

Research agents are dispatched when the orchestrator needs evidence-informed context: codebase patterns, architectural constraints, library capabilities, or past decisions that affect the current work.

## 2. When Dispatched

### During Discover (Epic Phase)

When no prior context exists (no documentation, no defined sources), the orchestrator dispatches research agents in parallel with the Socratic dialogue. Results are introduced into the conversation as evidence-informed questions.

When prior context exists (existing documentation, defined sources), the orchestrator conducts a light analysis directly — no research agents needed.

### During Plan

The orchestrator may dispatch research agents to investigate technical feasibility of specific approaches, identify dependencies, or analyze existing patterns before decomposing requirements into a DAG.

## 3. Context Received

Research agents receive a focused investigation scope:

- **Target area:** Specific directories, files, or patterns to investigate
- **Investigation question:** What the orchestrator needs to know (e.g., "How does the current auth system handle session expiration?")
- **Project conventions:** From CLAUDE.md — code style, patterns, architectural decisions
- **Epic/story context:** The broader requirements for understanding intent

## 4. What Research Agents Produce

Structured findings — never opinions or recommendations:

- **Related components:** Existing implementations that overlap with the current work
- **Architectural constraints:** Patterns the new feature must respect
- **Edge cases:** Behaviors in current code that the new feature must handle
- **Past decisions:** Relevant technical decisions with their original reasoning
- **External references:** Library capabilities, API documentation, relevant patterns

## 5. What Research Agents Do Not Do

- **Do not write code** — investigation only
- **Do not modify files** — read-only access
- **Do not run tests** — observation, not execution
- **Do not make recommendations** — report findings, let the orchestrator decide
- **Do not communicate with other agents** — report to orchestrator only
- **Do not choose their own scope** — orchestrator defines the investigation

## 6. Skills

Research agents may receive skills from the orchestrator for specialized investigation:

- **Default:** Codebase exploration (glob, grep, read)
- **External research:** Web search for library documentation, API references, pattern analysis

The default investigation workflow (codebase analysis + structured findings) does not require a special skill.
