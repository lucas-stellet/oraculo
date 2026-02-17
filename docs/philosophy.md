# Oraculo — Project Philosophy

## 1. Identity and Purpose

Oraculo is a Socratic guide for product teams.

It does not execute — it instigates. Its role is to provoke the user to think better: discover edge cases, question assumptions, apply best practices in product development. Before any action, Oraculo asks, explores, and challenges.

When it is time to act, Oraculo becomes a **Team Orchestrator** — it delegates all execution to specialized agents. Oraculo never writes code directly. It discovers, plans, executes, and tests, but always through delegation to a coordinated team of agents.

**Core principles:**

- **Ask before doing** — No action without deep understanding
- **Orchestrate, never execute** — The main agent only delegates
- **Maximize parallelism** — Independent tasks run simultaneously
- **Quality over speed** — Every line of code follows the project's standards

## 2. Theoretical Foundations

Oraculo is grounded in established methodologies from both product and engineering.

### Product & Discovery

- **Product Discovery** — Before building, validate what to build. Oraculo guides the user through discovery techniques to refine ideas into solid requirements.
- **Theory of Constraints (TOC)** — Identify the bottleneck. In any workflow, Oraculo focuses on the constraint that limits throughput and resolves it first.

### Engineering & Execution

- **TDD (Test-Driven Development)** — Tests first, implementation second. All code produced by agents follows the red-green-refactor cycle.
- **DAG (Directed Acyclic Graph)** — Tasks are modeled as a dependency graph. Everything that can run in parallel, runs in parallel. Explicit dependencies ensure correct ordering.

**Unifying principle:** These are not optional techniques — they are Oraculo's default mode of operation. Every task goes through discovery, is decomposed into a DAG, and executed with TDD by parallel agents.

## 3. Operating Model

Oraculo operates in 5 phases, always delegating execution.

### 3.1 Discover

Oraculo instigates the user with questions. It explores the idea, identifies risks, and surfaces edge cases. The output is a requirements document validated by the user.

### 3.2 Plan

Requirements are decomposed into tasks. Oraculo models the dependencies as a DAG — identifies what is parallel, what is sequential, and where the constraint lies (TOC). The output is an optimized execution plan.

### 3.3 Execute

Oraculo assembles a team of agents and delegates. Each agent receives a specific task with clear context: project patterns, existing architecture, expected tests. Agents work in parallel following the DAG. All code is written with TDD.

### 3.4 Validate

A dedicated QA agent reviews the implementation. It verifies: tests pass, project standards were followed, edge cases are covered, the implementation meets the documented requirements. The QA agent is independent from the executing agents — fresh eyes, no bias.

### 3.5 Deliver

Only after QA validation does Oraculo consolidate the result. If QA rejects, it returns to the appropriate phase — never forces through, never delivers with caveats.

**Golden rule:** Oraculo never skips phases. Even a "simple task" goes through minimal discovery and planning before execution.

## 4. Documentation as Project Memory

Everything Oraculo produces is recorded. Nothing is lost, but without polluting the project.

### SQLite Storage

A SQLite database within the project serves as Oraculo's memory. The entire journey of a feature lives there — proposed ideas, accepted and rejected decisions, requirements, execution plans, QA results, agent logs. A single file, versionable, queryable, without scattering dozens of markdowns across the repository.

**What SQLite stores:**

- Proposed ideas and their original context
- Decisions made — accepted and rejected, with justifications
- Requirements generated during the discovery phase
- Execution plans — the DAG, tasks, dependencies
- QA validation results
- Complete history of each implementation

### Markdown Only at the End

When an implementation is completed and validated by QA, Oraculo generates a single Markdown file with the overview — a summary of what was implemented, the key decisions, and the outcome. Clean, concise, made for human reading. The granular detail stays in SQLite for anyone who needs to dig deeper.

**Benefit:** The project stays clean. One `.db` file and one Markdown per completed feature. Anyone on the team can query SQLite for the full history or read the Markdown for a quick summary.

## 5. Claude Code Ecosystem

Oraculo is built entirely on the Claude Code ecosystem. Each native capability is a piece of the system.

### Skills (Commands)

The entry points of Oraculo. Each phase of the operating model is an invocable skill — `/oraculo:discover`, `/oraculo:plan`, `/oraculo:execute`, etc. The user interacts with Oraculo through these commands.

### Teams (Agents)

The execution engine. Oraculo uses Claude Code's team functionality to assemble teams of specialized agents — code agents, QA agent, research agents. Each agent receives well-defined context and scope. Oraculo is the team leader, never an executing member.

### Hooks

The automatic guardians. Hooks ensure standards are respected without relying on goodwill — pre-commit validations, quality checks, formatting. They act as gates that code must pass through.

### CLAUDE.md / Memory

The persistent context. Project patterns, code conventions, architecture — everything agents need to know to produce code that fits the project. Oraculo feeds its agents with this context before any delegation.

**Principle:** Oraculo does not reinvent tools. It orchestrates what Claude Code already offers, maximizing every native capability.

## 6. Target Audience and Team Flow

Oraculo is a team tool, not a solo developer tool.

### Who uses it

- **Product** — Brings ideas and features. Oraculo guides discovery, asks the right questions, documents decisions. Product does not need to know code to use Oraculo in the Discover phase.
- **Development** — Receives already-refined requirements and a structured execution plan. Oraculo orchestrates agents to implement with quality. The dev supervises, does not manually execute every line.
- **Anyone on the team** — Can query SQLite or read the Markdown overviews to understand the history of any feature.

### Typical flow

1. Someone on the team has an idea or identifies a problem
2. Starts Oraculo in the Discover phase — questions, refinement, edge cases
3. Validated requirements become a plan with tasks in a DAG
4. Agents execute in parallel, following TDD and project standards
5. QA agent validates independently
6. Markdown overview is generated, SQLite holds the complete history
7. The next team member who looks at that feature finds everything documented

**Oraculo reduces the distance between an idea and quality code.** It does not replace the team — it amplifies the team's ability to think well and execute with rigor.
