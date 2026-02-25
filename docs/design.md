# Oraculo — System Design

## 1. Operating Model

Oraculo operates in two modes depending on the scope of the work.

**Full mode (epics):** Discover > Plan > Execute > Validate

**Reduced mode (stories without an epic):** Plan > Execute > Validate

Stories without an associated epic skip Discover — context comes directly from the user. The minimum path is always Plan > Execute > Validate.

### 1.1 Discover

Oraculo instigates the user with questions. It explores the idea, identifies risks, and surfaces edge cases. The output is a requirements document validated by the user.

_Mandatory for epics. Skipped when a story is submitted directly with sufficient context._

### 1.2 Plan

Requirements are decomposed into tasks. Oraculo models the dependencies as a DAG — identifies what is parallel, what is sequential, and where the constraint lies (TOC). The output is an optimized execution plan.

### 1.3 Execute

Oraculo assembles a team of agents and delegates. Each agent receives a specific task with clear context: project patterns, existing architecture, expected tests. Agents work in parallel following the DAG. All code is written with TDD.

### 1.4 Validate

A dedicated QA agent reviews the implementation. It verifies: tests pass, project standards were followed, edge cases are covered, the implementation meets the documented requirements. The QA agent is independent from the executing agents — fresh eyes, no bias. If QA rejects, Oraculo returns to the appropriate phase — never forces through, never accepts with caveats.

**Golden rule:** Oraculo never skips the core phases (Plan, Execute, Validate). For epics, Discover is mandatory.

## 2. Documentation as Project Memory

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

**Benefit:** The project stays clean. One `.db` file and one Markdown per validated feature. Anyone on the team can query SQLite for the full history or read the Markdown for a quick summary.

## 3. Claude Code Ecosystem

Oraculo is built entirely on the Claude Code ecosystem. Each native capability is a piece of the system.

### Skills (Commands)

The entry points of Oraculo. Each phase of the operating model is an invocable skill — `/oraculo:epic`, `/oraculo:story`, and so on. The user interacts with Oraculo through these commands.

### Teams (Agents)

The execution engine. Oraculo uses Claude Code's team functionality to assemble teams of specialized agents — code agents, QA agent, research agents. Each agent receives well-defined context and scope. Oraculo is the team leader, never an executing member.

### Hooks

The automatic guardians. Hooks ensure standards are respected without relying on goodwill — pre-commit validations, quality checks, formatting. They act as gates that code must pass through.

### CLAUDE.md / Memory

The persistent context. Project patterns, code conventions, architecture — everything agents need to know to produce code that fits the project. Oraculo feeds its agents with this context before any delegation.

**Principle:** Oraculo does not reinvent tools. It orchestrates what Claude Code already offers, maximizing every native capability.

## 4. Target Audience and Team Flow

Oraculo is a team tool, not a solo developer tool.

### Who uses it

- **Product** — Brings ideas and features. Oraculo guides discovery, asks the right questions, documents decisions. Product does not need to know code to use Oraculo in the Discover phase.
- **Development** — Receives already-refined requirements and a structured execution plan. Oraculo orchestrates agents to implement with quality. The dev supervises, does not manually execute every line.
- **Anyone on the team** — Can query SQLite or read the Markdown overviews to understand the history of any feature.

### Typical flow

**Epic flow (full mode):**

1. Someone on the team has an idea or identifies a problem
2. Starts Oraculo with `/oraculo:epic` — questions, refinement, edge cases
3. Validated requirements become a plan with tasks in a DAG
4. Agents execute in parallel, following TDD and project standards
5. QA agent validates independently — if rejected, returns to the appropriate phase

**Story flow (reduced mode):**

1. A work item is already defined or the user supplies direct context
2. Starts Oraculo with `/oraculo:story` — skips Discover, goes straight to Plan
3. Agents execute in parallel, following TDD and project standards
4. QA agent validates independently — if rejected, returns to the appropriate phase

**Oraculo reduces the distance between an idea and quality code.** It does not replace the team — it amplifies the team's ability to think well and execute with rigor.
