# Documentation Consistency — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Apply all 14 decisions from the design doc to align Oraculo's documentation, eliminating contradictions across 18 files.

**Architecture:** Pure documentation changes — editing markdown files to resolve terminology, path, and structural inconsistencies. One new file (`research-agent.md`). No code changes.

**Tech Stack:** Markdown only. No build, no tests.

---

### Task 1: Fix CLAUDE.md — asterisk, frameworks, agent references

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Fix the stray asterisk on line 20**

Replace:
```
- **Software Engineering*** (story only): Plan > Execute > Validate
```
With:
```
- **Software Engineering** (story only): Plan > Execute > Validate
```

**Step 2: Update framework list on line 26**

Replace:
```
Transforms raw ideas into validated problem definitions through Socratic exploration. Uses Double Diamond, JTBD, OST, and Assumption Mapping frameworks.
```
With:
```
Transforms raw ideas into validated problem definitions through Socratic exploration. Uses Double Diamond, JTBD (with Four Forces), TOC, and Assumption Mapping frameworks.
```

**Step 3: Update artifact-templates comment on line 63**

Replace:
```
        │       ├── artifact-templates.md      # Intermediate artifact formats (OST, Codebase Impact)
```
With:
```
        │       ├── artifact-templates.md      # Intermediate artifact formats (Codebase Impact)
```

**Step 4: Add agents philosophy link and fix design link**

Replace lines 42-46:
```
## Agents

The agent layer is Oraculo's execution workforce. The orchestrator decomposes work into a DAG and dispatches code agents (with TDD skill) and QA agents (with clean context). All agents work on the same branch — no worktrees. Ephemeral SQLite tracks state; committed markdowns capture outcomes.

Full document: [docs/agents/design.md](docs/agents/design.md)
```
With:
```
## Agents

The agent layer is Oraculo's execution workforce. The orchestrator decomposes work into a DAG and dispatches code agents (with TDD skill), research agents (for codebase analysis), and QA agents (with clean context). All agents work on the same branch — no worktrees. SQLite tracks operational state; the knowledge table persists lessons learned; committed markdowns capture outcomes.

Full document: [docs/agents/philosophy.md](docs/agents/philosophy.md)

Design index: [docs/agents/design.md](docs/agents/design.md)
```

**Step 5: Commit**

```bash
git add CLAUDE.md
git commit -m "fix: align CLAUDE.md with design decisions — frameworks, agent refs, asterisk"
```

---

### Task 2: Update docs/design.md — operating modes, SQLite, markdown principle, team terminology, standalone stories

**Files:**
- Modify: `docs/design.md`

**Step 1: Update operating mode terminology (lines 7-9)**

Replace:
```
**Full mode (epics):** Discover > Plan > Execute > Validate

**Reduced mode (stories without an epic):** Plan > Execute > Validate
```
With:
```
**Product Engineering (epics):** Discover > Plan > Execute > Validate

**Software Engineering (stories without an epic):** Plan > Execute > Validate
```

**Step 2: Rewrite SQLite section (lines 37-54)**

Replace:
```
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
```
With:
```
### SQLite — Operational State + Accumulated Knowledge

A single SQLite database (`.oraculo/oraculo.db`) within the project serves two purposes:

**Transient operational data** — Task status, dependencies, QA verdicts, and execution state. This data is essential during active development but can be cleaned after an epic completes.

**Persistent knowledge** — The `knowledge` table accumulates lessons learned, codebase patterns, and conventions discovered across all epics. This data survives epic lifecycle and grows richer over time.

On epic/story completion:
1. Generate a markdown overview (summary of what was implemented, key decisions, QA outcome)
2. Extract lessons learned into the `knowledge` table
3. Operational data for that epic may be cleaned (optional)

### Markdown as Phase Output

Each phase of the operating model produces its own markdown artifact:
- **Discover** outputs a requirements document (`requirements.md` at epic level)
- **Plan** consumes requirements and produces the DAG (tracked in SQLite)
- **Validate** triggers generation of an overview markdown summarizing the implementation

SQLite holds operational state and accumulated knowledge. Markdown files capture product definitions and implementation summaries. The project stays clean — one `.db` file, one `requirements.md` per epic/story, and one overview per validated implementation.
```

**Step 3: Update Teams terminology (lines 64-66)**

Replace:
```
### Teams (Agents)

The execution engine. Oraculo uses Claude Code's team functionality to assemble teams of specialized agents — code agents, QA agent, research agents. Each agent receives well-defined context and scope. Oraculo is the team leader, never an executing member.
```
With:
```
### Team of Agents

The execution engine. Oraculo uses Claude Code's team functionality to assemble a team of specialized agents — code agents, research agents, and QA agents. Each agent receives well-defined context and scope. Oraculo is the team leader, never an executing member.
```

**Step 4: Add standalone story note (after line 11)**

Replace:
```
Stories without an associated epic skip Discover — context comes directly from the user. The minimum path is always Plan > Execute > Validate.
```
With:
```
Stories without an associated epic skip Discover — context comes directly from the user. The minimum path is always Plan > Execute > Validate. Standalone stories use a lightweight epic created automatically by the CLI, preserving the hierarchical data model transparently.
```

**Step 5: Update typical flow labels (lines 90, 98)**

Replace:
```
**Epic flow (full mode):**
```
With:
```
**Epic flow (Product Engineering):**
```

Replace:
```
**Story flow (reduced mode):**
```
With:
```
**Story flow (Software Engineering):**
```

**Step 6: Commit**

```bash
git add docs/design.md
git commit -m "fix: align design.md — terminology, SQLite role, markdown principle, team naming"
```

---

### Task 3: Update docs/epic/philosophy.md — framework consolidation

**Files:**
- Modify: `docs/epic/philosophy.md`

**Step 1: Fold Four Forces into JTBD section (section 3.2)**

Replace section 3.2:
```
### 3.2 Jobs to Be Done (JTBD) — Uncover the Real Job

Users don't want features — they "hire" products to make progress in specific circumstances. JTBD forces the conversation beyond surface requests into the functional, emotional, and social dimensions of the user's need.

**Why it matters:** When someone asks for a solution, the real question is: what job are they trying to accomplish? In what situation? What do they use today? Why is it insufficient? The output is a Job Story, not a feature description:

> "When [specific situation], I want [motivation], so I can [expected outcome]."
```
With:
```
### 3.2 Jobs to Be Done (JTBD) — Uncover the Real Job

Users don't want features — they "hire" products to make progress in specific circumstances. JTBD forces the conversation beyond surface requests into the functional, emotional, and social dimensions of the user's need.

**Why it matters:** When someone asks for a solution, the real question is: what job are they trying to accomplish? In what situation? What do they use today? Why is it insufficient? The output is a Job Story, not a feature description:

> "When [specific situation], I want [motivation], so I can [expected outcome]."

**Four Forces of Progress.** JTBD includes a model of what drives and resists adoption: Push (frustration with current state), Pull (attraction of the new), Anxiety (fear of the unknown), and Habit (inertia of current behavior). These forces surface adoption edge cases that purely functional analysis misses — a solution can be technically perfect and still fail because users are anxious about change or attached to their current habits.
```

**Step 2: Replace OST section (3.3) with TOC**

Replace:
```
### 3.3 Opportunity Solution Tree (OST) — Structure the Exploration

Teresa Torres' framework provides the structural backbone: desired Outcome at the top, branching into Opportunities (real user needs), then potential Solutions, then Experiments to validate assumptions. This tree prevents solutions from appearing without evidence-based anchoring.

**Why it matters:** Without structure, exploration becomes a wandering conversation. The OST keeps every idea connected to a real user need and a measurable outcome. Solutions that can't trace back to an opportunity have no foundation.
```
With:
```
### 3.3 Theory of Constraints (TOC) — Focus on the Bottleneck

Every system has a single constraint that limits its throughput. TOC asks: what is the one bottleneck that, if removed, would unlock the most value? Applied to discovery, this forces prioritization — instead of exploring every opportunity equally, the session identifies which problem is the real constraint for the user or the system.

**Why it matters:** Without focus, exploration produces a broad but shallow map. TOC forces convergence toward the highest-impact problem. This same principle applies later in execution, where QA throughput governs the pace of code production.
```

**Step 3: Remove Working Backwards section (3.5)**

Delete the entire section:
```
### 3.5 Working Backwards (PR/FAQ) — Stress Test the Idea

Amazon's technique of writing a fictional press release before building forces clarity about who benefits, what changes, and why it matters. If the announcement doesn't sound compelling, the idea isn't ready.

**Why it matters:** Describing the end result as if it already exists forces the user to confront gaps in their thinking that abstract discussion misses. It answers: "Would anyone care about this?"
```

**Step 4: Remove Four Forces standalone section (3.6)**

Delete the entire section:
```
### 3.6 Four Forces of Progress — Map Behavioral Edge Cases

The JTBD Four Forces model maps what drives and resists adoption: Push (frustration with current state), Pull (attraction of the new), Anxiety (fear of the unknown), and Habit (inertia of current behavior). This surfaces edge cases that purely functional analysis misses.

**Why it matters:** A solution can be technically perfect and still fail because users are anxious about change or attached to their current habits. These forces reveal the human side of adoption that feature lists ignore.
```

**Step 5: Renumber section 3.4 to 3.4 (stays the same — Assumption Mapping)**

No change needed — Assumption Mapping was 3.4 and remains 3.4 after removing 3.3/3.5/3.6.

**Step 6: Commit**

```bash
git add docs/epic/philosophy.md
git commit -m "refactor: consolidate epic frameworks — keep DD, JTBD+4Forces, TOC, AM; remove OST, WB"
```

---

### Task 4: Update docs/epic/design.md — paths, phases, frameworks, 9 sections

**Files:**
- Modify: `docs/epic/design.md`

**Step 1: Update Divergence phase (section 1.2) — remove standalone Four Forces header**

Replace the section header `Oraculo applies JTBD and Four Forces question templates` with `Oraculo applies JTBD question templates (including Four Forces)`.

Keep all JTBD and Four Forces questions in place — they're now under JTBD.

**Step 2: Update Codebase Analysis phase (section 1.3) — context-dependent dispatch**

Replace:
```
While the dialogue progresses, Oraculo dispatches sub-agents to analyze the existing codebase. This runs in parallel — it does not block the conversation.
```
With:
```
When no prior context exists (no documentation, no defined sources), Oraculo dispatches research agents to analyze the existing codebase. This runs in parallel — it does not block the conversation. When prior context exists (defined sources, existing documentation), the orchestrator conducts a light analysis directly without dispatching agents.
```

**Step 3: Remove Stress Test phase (section 1.6)**

Delete the entire section 1.6 (Working Backwards / PR/FAQ stress test):
```
### 1.6 Stress Test — Working Backwards

Oraculo runs a simplified PR/FAQ to test whether the idea has sufficient clarity to proceed.

**Stress test questions:**
- "If this feature launched tomorrow, how would the announcement read in two sentences?"
- "What would a skeptical customer's first question be?"
- "What would engineering's main objection be?"
- "How would you measure success three months after launch?"

**Exit condition:** The user can answer all four questions without significant hesitation or contradiction with earlier answers. If the user struggles, Oraculo identifies which gap exists and loops back to the step that addresses it (typically 1.2 or 1.4).
```

Renumber subsequent sections: 1.7 becomes 1.6, 1.8 becomes 1.7.

**Step 4: Update artifact generation path (formerly 1.8, now 1.7)**

Replace:
```
Oraculo generates the Requirements Document and saves it to `.oraculo/projects/<project-name>/epic.md`.
```
With:
```
Oraculo generates the Requirements Document and saves it to `.oraculo/epics/<epic-name>/requirements.md`.
```

**Step 5: Remove OST artifact (section 2.2)**

Delete:
```
### 2.2 Opportunity Solution Tree (OST)

A structured tree:

\```
Outcome (business metric / user behavior)
  └── Opportunity (user need, written in first person)
        └── Solution (proposed approach)
              └── Assumption (hypothesis + validation criteria)
\```
```

Renumber: 2.3 → 2.2, 2.4 → 2.3, 2.5 → 2.4.

**Step 6: Update Requirements Document reference (formerly 2.5, now 2.4)**

Replace:
```
The primary output — a 9-section document saved to `.oraculo/projects/<project-name>/epic.md`. Each REC-N requirement is a candidate for decomposition into stories.
```
With:
```
The primary output — saved to `.oraculo/epics/<epic-name>/requirements.md`. The document has 9 sections:

1. Summary
2. Problem
3. Context & Motivation
4. Requirements
5. User Experience
6. Scope
7. Assumptions & Risks
8. Success Criteria
9. Open Questions

Each REC-N requirement is a candidate for decomposition into stories.
```

**Step 7: Update Deep Reasoning framework references (section 4.2)**

Replace:
```
- Uses full framework toolkit (JTBD, Four Forces, PR/FAQ, OST)
```
With:
```
- Uses full framework toolkit (JTBD with Four Forces, TOC, Assumption Mapping)
```

**Step 8: Update Standard Epic flow (section 3.1)**

Replace:
```
**Flow:** Full sequence — Session Setup → Reframing → Divergence → Codebase Analysis → Convergence → Assumption Mapping → Stress Test → Exit Gate → Artifact Generation.
```
With:
```
**Flow:** Full sequence — Session Setup → Reframing → Divergence → Codebase Analysis → Convergence → Assumption Mapping → Exit Gate → Artifact Generation.
```

**Step 9: Update Integration with Stories (section 5)**

Replace:
```
1. Invoke `/oraculo:story <project-name>`
2. The Story skill reads `.oraculo/projects/<project-name>/epic.md`
```
With:
```
1. Invoke `/oraculo:story <epic-name>`
2. The Story skill reads `.oraculo/epics/<epic-name>/requirements.md`
```

**Step 10: Update Output Path (section 6)**

Replace:
```
All epic artifacts are saved to `.oraculo/projects/<project-name>/epic.md` inside the target application. The project name is derived from the idea's domain or explicitly asked from the user during artifact generation.
```
With:
```
All epic artifacts are saved to `.oraculo/epics/<epic-name>/requirements.md` inside the target application. The epic name is derived from the idea's domain or explicitly asked from the user during artifact generation.
```

**Step 11: Commit**

```bash
git add docs/epic/design.md
git commit -m "refactor: update epic design — paths, remove WB/OST phases, enumerate 9 sections"
```

---

### Task 5: Update docs/story/design.md — path convention

**Files:**
- Modify: `docs/story/design.md`

**Step 1: Update parent epic lookup (section 1.0, line 13)**

Replace:
```
  - Search for `.oraculo/projects/<project-name>/epic.md`
```
With:
```
  - Search for `.oraculo/epics/<epic-name>/requirements.md`
```

**Step 2: Update output path (section 1.4, line 72)**

Replace:
```
**Output path:** `.oraculo/projects/<project-name>/story-<title-slug>.md`
```
With:
```
**Output path:** `.oraculo/epics/<epic-name>/stories/<story-name>/requirements.md`
```

**Step 3: Update parent epic workflow (section 4, line 130)**

Replace:
```
1. **Reading:** The Story skill reads `.oraculo/projects/<project-name>/epic.md`
```
With:
```
1. **Reading:** The Story skill reads `.oraculo/epics/<epic-name>/requirements.md`
```

**Step 4: Update Output Path section (section 7, line 161)**

Replace:
```
Story artifacts are saved to `.oraculo/projects/<project-name>/story-<title-slug>.md` inside the target application. The title slug is derived from the story title (lowercase, hyphenated). If no project name was provided, the user is asked during artifact generation.
```
With:
```
Story artifacts are saved to `.oraculo/epics/<epic-name>/stories/<story-name>/requirements.md` inside the target application. The story name is derived from the story title (lowercase, hyphenated). If no epic name was provided and the story is standalone, the CLI creates a lightweight epic automatically.
```

**Step 5: Commit**

```bash
git add docs/story/design.md
git commit -m "fix: update story design paths to .oraculo/epics/ convention"
```

---

### Task 6: Update docs/story/philosophy.md — lightweight epic note

**Files:**
- Modify: `docs/story/philosophy.md`

**Step 1: Add lightweight epic note to Relationship section (section 6)**

Replace:
```
- **Standalone:** Work that doesn't need an Epic. The problem is clear, the scope is small, and the Story skill provides enough structure to define it.
```
With:
```
- **Standalone:** Work that doesn't need an Epic. The problem is clear, the scope is small, and the Story skill provides enough structure to define it. Under the hood, the CLI creates a lightweight epic to preserve the hierarchical data model — this is transparent to the user.
```

**Step 2: Commit**

```bash
git add docs/story/philosophy.md
git commit -m "docs: add lightweight epic note to story philosophy"
```

---

### Task 7: Update docs/cli/philosophy.md — fix scope claim

**Files:**
- Modify: `docs/cli/philosophy.md`

**Step 1: Fix outdated scope claim (line 96)**

Replace:
```
Currently it covers the Discover phase; as Oraculo matures, it will cover all core phases: Discover, Plan, Execute, and Validate.
```
With:
```
It covers all core phases: Discover, Plan, Execute, and Validate.
```

**Step 2: Commit**

```bash
git add docs/cli/philosophy.md
git commit -m "fix: correct CLI philosophy scope — already covers all phases"
```

---

### Task 8: Update docs/cli/design.md — lightweight epic, validations schema

**Files:**
- Modify: `docs/cli/design.md`

**Step 1: Add lightweight epic documentation (after line 127)**

Replace:
```
- Stories always belong to an epic. CLI enforces this; standalone stories use a lightweight epic.
```
With:
```
- Stories always belong to an epic. CLI enforces this. When a story is created without an explicit epic, the CLI auto-creates a **lightweight epic** — a minimal epic with the name derived from the story and no requirements markdown. This preserves the hierarchical data model while keeping the standalone story UX seamless.
```

**Step 2: Add `task_id` to validations table (line 189-194)**

Replace:
```sql
CREATE TABLE validations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id    INTEGER NOT NULL REFERENCES stories(id),
    verdict     TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
    created_at  TEXT DEFAULT (datetime('now'))
);
```
With:
```sql
CREATE TABLE validations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id    INTEGER NOT NULL REFERENCES stories(id),
    task_id     INTEGER REFERENCES tasks(id),              -- NULL = story-level validation
    verdict     TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
    created_at  TEXT DEFAULT (datetime('now'))
);
```

**Step 3: Update validations design decision text (line 231)**

Replace:
```
- `validations` stores only the structural verdict. QA analysis lives in Markdown.
```
With:
```
- `validations` stores structural verdicts at two levels: per-task (`task_id` filled) during execution and per-story (`task_id` NULL) as the final gate. QA analysis lives in Markdown.
```

**Step 4: Add knowledge persistence note after the schema section**

After the design decisions list (line 233), add:
```

**Knowledge persistence:** Unlike operational tables (`tasks`, `task_dependencies`, `task_results`, `validations`), the `knowledge` table is persistent — it accumulates lessons learned across all epics and survives epic completion. Operational data can be cleaned after an epic completes; knowledge data is retained.
```

**Step 5: Commit**

```bash
git add docs/cli/design.md
git commit -m "fix: lightweight epic definition, two-level validations, knowledge persistence"
```

---

### Task 9: Update docs/agents/philosophy.md — SQLite and knowledge store

**Files:**
- Modify: `docs/agents/philosophy.md`

**Step 1: Rewrite memory section (section 6, lines 51-61)**

Replace:
```
**Ephemeral SQLite tracks execution state.** Task status, dependencies, QA verdicts, and operational data live in SQLite — per-epic, not committed, disposable after the epic completes. This is infrastructure, not knowledge.

**No complex memory system.** There is no three-tier architecture, no curation pipeline, no semantic knowledge store. The system keeps what it needs in the simplest form that works: markdown for humans, SQLite for machines, CLAUDE.md for agents.
```
With:
```
**SQLite tracks operational state and accumulated knowledge.** A single database per project (`.oraculo/oraculo.db`) holds two categories of data. Transient operational data — task status, dependencies, QA verdicts — is essential during execution and can be cleaned after epic completion. Persistent knowledge — lessons learned, codebase patterns, conventions — accumulates across all epics and survives their lifecycle. The database is not committed (`.gitignore`), but the knowledge it contains is the project's long-term memory.

**Simple knowledge store, not a complex memory system.** The `knowledge` table with full-text search provides a single, queryable store for codebase findings. There is no three-tier architecture, no curation pipeline, no promotion scoring. This is a practical knowledge store — not the deferred rich memory system described in future-work.md.
```

**Step 2: Commit**

```bash
git add docs/agents/philosophy.md
git commit -m "fix: redefine SQLite role — transient operations + persistent knowledge"
```

---

### Task 10: Update docs/agents/design.md — research agent, terminology

**Files:**
- Modify: `docs/agents/design.md`

**Step 1: Update runtime section (section 5, lines 52-54)**

Replace:
```
### Ephemeral SQLite

Task status, dependencies, QA verdicts, and operational data live in a SQLite database at `.oraculo/oraculo.db`. The database uses the schema from [`docs/cli/design.md`](../cli/design.md) §4.3 — `epics`, `stories`, `tasks`, `task_dependencies`, `task_results`, `validations`, and `knowledge` tables. It is listed in `.gitignore` and is not committed. The database is infrastructure — essential during execution, disposable after.
```
With:
```
### SQLite — Operational State + Knowledge

A single SQLite database at `.oraculo/oraculo.db` tracks two categories of data. Operational state (tasks, dependencies, QA verdicts) is transient — essential during execution, cleanable after epic completion. The `knowledge` table is persistent — it accumulates lessons learned across all epics. The database uses the schema from [`docs/cli/design.md`](../cli/design.md) §4.3. It is listed in `.gitignore` and not committed, but its knowledge data is the project's long-term memory.
```

**Step 2: Update Memory Model (section 5, lines 61-67)**

Replace:
```
- **Ephemeral SQLite** — Execution state during active development (transient, not committed)

No three-tier memory architecture, no curation pipeline, no semantic knowledge store. The simplest model that works.
```
With:
```
- **SQLite** — Operational state (transient) + accumulated knowledge (persistent), not committed

No three-tier memory architecture, no curation pipeline. A simple knowledge table with full-text search — not a promotion scoring system.
```

**Step 3: Add Research Agent section (after section 4, before section 5)**

After the QA Agent section (line 48), add:

```

## 4.5 Research Agent

Research agents investigate the codebase and external references to provide evidence for decision-making. They are dispatched during Discover (when no prior context exists) and during Plan (for technical feasibility analysis).

Each research agent receives a focused investigation scope: relevant directories, specific patterns to look for, or external references to analyze. They return structured findings — not opinions. The orchestrator integrates findings into the dialogue or the planning process.

Research agents do not write code, run tests, or modify files. They observe, analyze, and report.

Full details: [`design/research-agent.md`](design/research-agent.md)
```

**Step 4: Commit**

```bash
git add docs/agents/design.md
git commit -m "fix: SQLite role, add research agent section to agents index"
```

---

### Task 11: Create docs/agents/design/research-agent.md

**Files:**
- Create: `docs/agents/design/research-agent.md`

**Step 1: Write the research agent design document**

```markdown
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
```

**Step 2: Commit**

```bash
git add docs/agents/design/research-agent.md
git commit -m "docs: add research agent design document"
```

---

### Task 12: Update docs/agents/design/overview.md — Discover phase, research agent, QA levels

**Files:**
- Modify: `docs/agents/design/overview.md`

**Step 1: Update Discover phase (section 2.1)**

Replace:
```
### 2.1 Discover

**Actors:** Orchestrator + Human

The orchestrator asks questions, not agents. No code agents or QA agents are involved. The output is a validated requirements document stored as markdown in `.oraculo/epics/<name>/requirements.md`.
```
With:
```
### 2.1 Discover

**Actors:** Orchestrator + Human (+ Research Agents when no prior context exists)

The orchestrator asks questions, not code or QA agents. When no prior context exists (no documentation, no defined sources), research agents are dispatched in parallel to analyze the codebase and surface evidence. When prior context exists, the orchestrator conducts a light analysis directly. The output is a validated requirements document stored as markdown in `.oraculo/epics/<name>/requirements.md`.
```

**Step 2: Update Validation phase (section 2.4)**

Replace:
```
### 2.4 Validate

**Actors:** QA Agent

A dedicated QA agent reviews the implementation with a clean context window. It receives the diff, the specs, and test results — never the code agent's reasoning. QA never fixes; it reports structured findings. If QA rejects, the orchestrator spawns a new code agent with QA's feedback.
```
With:
```
### 2.4 Validate

**Actors:** QA Agent

Validation operates at two levels:

**Task-level validation:** After each code agent completes a task, the QA agent reviews the implementation with a clean context window. It receives the diff, the specs, and test results — never the code agent's reasoning. QA never fixes; it reports structured findings. If QA rejects, the orchestrator spawns a new code agent with QA's feedback.

**Story-level validation:** Once all tasks in a story pass individual QA, the QA agent performs an integrated review — cross-task coherence, story requirements as a whole, and overall quality. The story verdict is the final gate before markdown and knowledge generation.
```

**Step 3: Commit**

```bash
git add docs/agents/design/overview.md
git commit -m "fix: qualify Discover actors, add two-level QA validation"
```

---

### Task 13: Update docs/agents/design/qa-agent.md — two-level validation

**Files:**
- Modify: `docs/agents/design/qa-agent.md`

**Step 1: Add a section after section 1 (Role)**

After line 8, add:

```

## 1.1 Two Levels of Validation

QA operates at two distinct levels:

**Task-level:** Each completed task is reviewed independently. The QA agent receives the diff, specs, and test results for that specific task. Verdicts are recorded with `task_id` in the `validations` table.

**Story-level:** Once all tasks in a story pass individual QA, a final integrated review checks cross-task coherence, story requirements as a whole, and overall quality. This verdict is recorded with `task_id = NULL` in the `validations` table and serves as the final gate before generating the story's markdown summary and extracting lessons learned into the knowledge table.
```

**Step 2: Commit**

```bash
git add docs/agents/design/qa-agent.md
git commit -m "docs: add two-level validation section to QA agent design"
```

---

### Task 14: Update docs/agents/design/runtime.md — single DB, data categories

**Files:**
- Modify: `docs/agents/design/runtime.md`

**Step 1: Fix "each epic gets its own database" (line 5)**

Replace:
```
Each epic gets its own SQLite database at `.oraculo/oraculo.db`. This database is ephemeral — it is listed in `.gitignore` and is not committed to the repository. It serves as infrastructure for tracking execution state, not as a knowledge store.
```
With:
```
A single SQLite database at `.oraculo/oraculo.db` serves the entire project. It is listed in `.gitignore` and is not committed to the repository. The database holds two categories of data: transient operational state (tasks, dependencies, QA verdicts) that can be cleaned after epic completion, and persistent knowledge (lessons learned, codebase patterns) that accumulates across all epics.
```

**Step 2: Update validations row in table (line 14)**

Replace:
```
- **`validations`** — QA verdicts per story (approved/rejected)
```
With:
```
- **`validations`** — QA verdicts per task and per story (approved/rejected)
```

**Step 3: Update QA verdicts row in tracking table (line 24)**

Replace:
```
| QA verdicts | `validations` | Whether a story passed or failed validation |
```
With:
```
| QA verdicts | `validations` | Whether a task or story passed or failed validation |
```

**Step 4: Rewrite lifecycle section (lines 27-31)**

Replace:
```
### Lifecycle

The database is created automatically by the CLI on the first `oraculo tools` command. It lives for the duration of the epic's active development. Once the epic is complete and the final markdown artifact is generated, the database has served its purpose — it can be deleted without losing any committed knowledge.

**Why ephemeral?** The database contains operational state (task progress, agent logs, intermediate results) that is valuable during execution but not after. Committing it would pollute the repository with transient data. The meaningful output — requirements, decisions, and summaries — is captured in committed markdown files.
```
With:
```
### Lifecycle

The database is created automatically by the CLI on the first `oraculo tools` command. It persists for the life of the project.

**Transient data** (tasks, dependencies, QA verdicts, task results) is valuable during execution but not after. Once an epic completes and its markdown artifacts are generated, this operational data can be cleaned.

**Persistent data** (knowledge table) accumulates across all epics. When an epic/story completes, lessons learned are extracted into the knowledge table. This data is never cleaned — it is the project's long-term memory.

The database is not committed (`.gitignore`) because it contains operational state that would pollute the repository. The meaningful output — requirements, decisions, summaries, and knowledge — is captured in committed markdown files and the persistent knowledge table.
```

**Step 5: Update "What This Is Not" section (lines 88-96)**

Replace:
```
There is no three-tier memory architecture (working/episodic/semantic). There is no curation pipeline. There is no semantic knowledge store with scoring and promotion. These are powerful concepts but add complexity that is not justified at this stage. The simple model — CLAUDE.md + markdowns + ephemeral SQLite — covers the essential needs:
```
With:
```
There is no three-tier memory architecture (working/episodic/semantic). There is no curation pipeline or promotion scoring. The knowledge table with full-text search provides a simple, queryable store for codebase findings — not the deferred rich memory system described in future-work.md. The model — CLAUDE.md + markdowns + SQLite (operational state + knowledge) — covers the essential needs:
```

**Step 6: Commit**

```bash
git add docs/agents/design/runtime.md
git commit -m "fix: single DB per project, two data categories, two-level validations"
```

---

### Task 15: Update skill reference files — frameworks, artifacts, questions

**Files:**
- Modify: `claude-kit/skills/oraculo/epic/references/frameworks.md`
- Modify: `claude-kit/skills/oraculo/epic/references/artifact-templates.md`
- Modify: `claude-kit/skills/oraculo/epic/references/question-bank.md`

**Step 1: Rewrite frameworks.md**

Replace entire content with:

```markdown
# Frameworks Reference

Concise descriptions of each framework used in Epic exploration and when it applies.

---

## Double Diamond (Diverge/Converge)

The session follows two movements: first expand the problem space (diverge), then narrow to a clear definition (converge). This prevents the most common failure — jumping to solutions before understanding the problem. Both Light and Deep levels use this structure; Light diverges in the technical space, Deep in the product space.

**When:** Always. This is the macro-structure of every epic session.

---

## Jobs to Be Done (JTBD)

Users don't want features — they "hire" products to make progress in specific circumstances. JTBD explores the functional, emotional, and social dimensions of the user's need, producing Job Stories instead of feature descriptions.

Includes the **Four Forces of Progress** — Push (frustration with current state), Pull (attraction of the new), Anxiety (fear of the unknown), Habit (inertia of current behavior). These forces surface adoption edge cases that functional analysis misses.

**When:** Divergence phase. Full dimensions (functional, emotional, social) and all four forces in Deep only. Light uses a technical variant focused on system behavior, pre-conditions, and technical adoption forces.

---

## Theory of Constraints (TOC)

Every system has a single constraint that limits its throughput. TOC asks: what is the one bottleneck that, if removed, would unlock the most value? Applied to discovery, this forces prioritization toward the highest-impact problem.

**When:** Convergence phase, both levels. Deep identifies the user/business constraint. Light identifies the technical constraint. The same principle applies later in execution, where QA throughput governs the pace.

---

## Assumption Mapping

Makes hidden assumptions explicit and scores them by risk: Impact (if wrong, how bad?) x (6 - Evidence). The riskiest assumptions become priorities before any investment.

**When:** Assumption Mapping phase, both levels. Light focuses on Solution and Feasibility categories; Deep scores all four categories (Problem, Solution, Value, Feasibility).

---

## Technical Viability Checklist

Light-mode stress test. Tests whether the implementation approach fits the current architecture, what tests are needed, what code review objections might arise, and what technical acceptance criteria apply.

**When:** Exit Gate phase, Light only.
```

**Step 2: Update artifact-templates.md — remove OST artifact**

Delete the entire section "## 2. Opportunity Solution Tree (OST)" (lines 33-61). Renumber: section 3 → 2, section 4 → 3, section 5 → 4.

**Step 3: Update question-bank.md — remove standalone PR/FAQ, update Four Forces framing**

In section 3 header, replace:
```
## 3. Divergence — Four Forces
```
With:
```
## 3. Divergence — JTBD Four Forces
```

In section 7, replace:
```
## 7. Stress Test

### Deep — PR/FAQ (Working Backwards)

- "If this feature launched tomorrow, how would the announcement read in two sentences?"
- "What would a skeptical customer's first question be?"
- "What would engineering's main objection be?"
- "How would you measure success three months after launch?"

### Light — Technical Viability Checklist
```
With:
```
## 7. Stress Test

### Deep — Constraint-Focused Validation

- "From everything we explored, what is THE bottleneck — the single constraint that limits the most value?"
- "What would a skeptical customer's first question be?"
- "What would engineering's main objection be?"
- "How would you measure success three months after launch?"

### Light — Technical Viability Checklist
```

**Step 4: Commit**

```bash
git add claude-kit/skills/oraculo/epic/references/frameworks.md claude-kit/skills/oraculo/epic/references/artifact-templates.md claude-kit/skills/oraculo/epic/references/question-bank.md
git commit -m "refactor: align skill references with framework consolidation"
```

---

### Task 16: Final verification — grep for stale references

**Step 1: Search for remaining stale references**

```bash
# Search for old path convention
rg "\.oraculo/projects/" docs/ claude-kit/ CLAUDE.md

# Search for removed frameworks as standalone concepts
rg "Working Backwards" docs/ claude-kit/ CLAUDE.md
rg "PR/FAQ" docs/ claude-kit/ CLAUDE.md
rg "Opportunity Solution Tree" docs/ claude-kit/ CLAUDE.md

# Search for "Full mode" / "Reduced mode"
rg "Full mode" docs/ CLAUDE.md
rg "Reduced mode" docs/ CLAUDE.md

# Search for "ephemeral SQLite" (should not appear as the sole descriptor)
rg "ephemeral SQLite" docs/ CLAUDE.md

# Search for "each epic gets" (per-epic database claim)
rg "each epic gets" docs/
```

**Step 2: Fix any remaining stale references found**

**Step 3: Commit if changes were needed**

```bash
git add -A
git commit -m "fix: clean up remaining stale references"
```

---

## Execution Notes

- Tasks 1-15 are independent by file but should be executed in order for clean commit history.
- Task 16 is a verification sweep that catches anything missed.
- No code changes, no tests. All changes are markdown documentation.
- The `.pt-br.md` translation files are NOT updated in this plan — they should be updated in a separate pass after the English docs are finalized.
