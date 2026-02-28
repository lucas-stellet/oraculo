# Oraculo Commands — System Design

## 1. Anatomy of a Command

A command is composed of 3 layers of files:

```
skills/oraculo/<command>/
  SKILL.md                    # Entry point — frontmatter + persona + workflow overview
  phases/                     # Step-files loaded just-in-time
    00-setup.md
    01-phase-name.md
    ...
  references/                 # Supporting material loaded on demand
    question-bank.md
    frameworks.md
    artifact-templates.md
```

**SKILL.md** — The entry point. Contains:
- Enriched frontmatter (name, description, argument-hint, allowed-tools, disable-model-invocation)
- Oraculo's single functional persona
- Workflow overview (list of phases with one-line description each)
- Bootstrap instructions (check active session via CLI, load correct phase)

**phases/*.md** — One file per phase. Contains:
- XML guardrail tags relevant to that phase
- Phase-specific questions and logic
- Exit phase gate with exit criteria
- Anti-patterns and HALT conditions specific to that phase

**references/*.md** — Supporting material. Never loaded automatically. Read explicitly when a phase needs it (e.g., question-bank during Divergence, artifact-template during generation).

**Fundamental rule:** SKILL.md never exceeds ~100 lines. Complexity lives in phase files and references. The entry point is thin.

## 2. Frontmatter

The frontmatter defines metadata that Claude Code uses for registration and dispatch. The `description` describes WHEN to use the command, never summarizes the workflow.

```yaml
---
name: oraculo:epic
description: >
  Use when exploring a new feature idea that spans multiple areas,
  has significant uncertainty, or needs structured discovery before planning
argument-hint: <idea or problem description>
allowed-tools:
  - Read
  - Bash
  - Task
  - AskUserQuestion
disable-model-invocation: true
---
```

**Required fields:**

| Field | Purpose |
|-------|---------|
| `name` | Slash command identifier. Format: `oraculo:<phase>` |
| `description` | Trigger conditions — when to invoke. Never summarizes the workflow. |

**Optional fields:**

| Field | Purpose | When to use |
|-------|---------|-------------|
| `argument-hint` | Placeholder shown to user | When the command accepts input |
| `allowed-tools` | Tool whitelist | When the command should be restricted |
| `disable-model-invocation` | Prevents auto-invocation by the model | Always `true` for Oraculo commands — only the user invokes |

**What does NOT go in frontmatter:**
- Workflow logic (goes in SKILL.md body)
- File references (go in each phase's instructions)
- State or runtime variables (managed by CLI)

## 3. XML Vocabulary

10 tags divided into two categories. Each tag has a documented origin from research, justification from philosophy, and usage rules.

### Semantic Tags (organize prompt sections)

**`<persona>`** — Declares Oraculo's identity for the current phase.

```xml
<persona>
  You are Oraculo, a Socratic guide for quality product development.
  You ask before doing. You orchestrate, never execute.
  You prioritize understanding over speed.

  In this phase (Divergence), your focus is expanding the problem space —
  surfacing dimensions the user hasn't considered.
</persona>
```

Rule: Appears once in SKILL.md (base identity) and once in each phase file (phase focus). The second does not repeat the first — it adds phase context.

**`<workflow>`** — Declares the phase sequence in SKILL.md. Contains no logic — only the overview.

```xml
<workflow>
  00-setup:       Session triage and reasoning level detection
  01-reframing:   Separate problem from solution
  02-divergence:  Expand problem space (JTBD, Four Forces, edge cases)
  03-codebase:    Dispatch research agents (parallel)
  04-convergence: Narrow to validated problem definition
  05-assumptions: Map and score hidden assumptions
  06-exit-gate:   Validate four risks before proceeding
  07-artifact:    Generate requirements document via CLI
</workflow>
```

Rule: Appears only in SKILL.md. Never in phase files. Serves as a map, not an instruction.

**`<anti-patterns>`** — Named list of failure modes specific to the phase.

```xml
<anti-patterns>
  "Premature solutioning" — proposing HOW before validating WHAT/WHY
  "Assumption blindness" — accepting user hypotheses without questioning evidence
  "Shallow reframing" — accepting the first problem statement without pushing deeper
</anti-patterns>
```

Rule: Each anti-pattern has a quoted name followed by a description. Names must be memorable. Each phase file defines its own — there is no global list.

**`<context-boundaries>`** — Explicitly declares what the LLM can and cannot access in this phase.

```xml
<context-boundaries>
  Available: Raw idea from user, reasoning level from setup
  Not available: Codebase analysis (runs in parallel, results arrive later)
  Focus: Problem exploration, not solution design
  References: Load question-bank.md section "JTBD" when starting Four Forces
</context-boundaries>
```

Rule: Appears in each phase file. Prevents the LLM from inventing context it doesn't have or loading unnecessary references.

### Flow Tags (control execution at critical points)

**`<phase-gate>`** — Defines exit criteria and CLI persistence. The LLM does not advance without satisfying them.

```xml
<phase-gate phase="divergence">
  Exit conditions:
    - At least 3 JTBD dimensions explored (functional, emotional, social)
    - All four forces addressed (push, pull, anxiety, habit)
    - At least 1 edge case or risk the user hadn't previously considered

  Persist via CLI:
    - oraculo phase complete divergence --session=$SESSION_ID

  If CLI rejects: surface rejection to user. Do not retry with looser criteria.
  On success: Read phases/04-convergence.md
</phase-gate>
```

Rule: Exactly one per phase file, always at the end. Contains exit conditions, CLI command, and navigation instruction.

**`<halt>`** — Automatic stop condition when something is wrong.

```xml
<halt>
  - Zero assumptions identified after Divergence — something was ignored, re-analyze
  - User cannot articulate problem in one sentence after Convergence — return to Reframing
  - Codebase Analysis found zero related components — verify analysis was sufficient
</halt>
```

Rule: Each condition is a declarative sentence. The action is always stop and react, never continue silently.

**`<self-check>`** — Verification of claims before delivering an artifact or completing a phase.

```xml
<self-check phase="exit-gate">
  Before declaring all four risks addressed:
  1. Reread problem statement — still coherent with what was explored?
  2. Verify each critical assumption has evidence or explicit flag
  3. Confirm Codebase Analysis results were incorporated (not ignored)
  4. Check that scope boundaries defined in Reframing are still respected

  If any check fails — do not generate artifact. Return to appropriate phase.
</self-check>
```

Rule: Appears before irreversible transitions (artifact generation, handoff to next operating model phase). Checks are verifiable, not subjective.

**`<rationalization-table>`** — Maps LLM excuses to reality. Cognitive inoculation.

```xml
<rationalization-table>
  "The user already knows what they want"    → If they did, they wouldn't need Oraculo. Explore.
  "This idea is too simple for 8 phases"     → Simple ideas hide the most assumptions. Follow the process.
  "I already understand the problem"         → Understanding is not validating. The user must confirm.
  "We can skip Assumption Mapping, it's obvious" → Obvious assumptions are the most dangerous ones.
</rationalization-table>
```

Rule: Fixed format: quoted excuse followed by reality. Each entry is one sentence. Appears in phases where shortcut risk is highest.

**`<iteration-limit>`** — Defines maximum attempts before changing approach.

```xml
<iteration-limit phase="reframing" max="3">
  If after 3 attempts the user cannot separate problem from solution:
  1. Record the proposed solution as-is
  2. Ask: "What outcome do you expect this solution to produce for the user?"
  3. Use the answer as a proxy problem statement
  Do not insist indefinitely — adapt the approach.
</iteration-limit>
```

Rule: Always includes `max` and the escalation action. Escalation is never "give up" — it is "adapt."

**`<critical>`** — Inviolable rule repeated at the point of maximum drift risk.

```xml
<critical>
  The command does NOT generate requirements without user validation at each phase.
  Verbal confirmation is process, not approval.
  Approval is a validated artifact persisted through the CLI.
</critical>
```

Rule: Reserved for rules that, if violated, compromise system integrity. Maximum 2-3 per phase file. If everything is `<critical>`, nothing is.

### Vocabulary Summary

| Tag | Category | Where it appears | Frequency |
|-----|----------|-----------------|-----------|
| `<persona>` | Semantic | SKILL.md + each phase file | 1 per file |
| `<workflow>` | Semantic | SKILL.md only | 1 total |
| `<anti-patterns>` | Semantic | Each phase file | 1 per phase |
| `<context-boundaries>` | Semantic | Each phase file | 1 per phase |
| `<phase-gate>` | Flow | End of each phase file | 1 per phase |
| `<halt>` | Flow | Phase files where relevant | 0-1 per phase |
| `<self-check>` | Flow | Before irreversible transitions | 0-1 per phase |
| `<rationalization-table>` | Flow | Phases with highest shortcut risk | 0-1 per phase |
| `<iteration-limit>` | Flow | Phases with iterative interaction | 0-1 per phase |
| `<critical>` | Flow | Point of maximum drift risk | 0-3 per phase |

## 4. Phase File Structure

Each phase file follows a fixed structure. Section order is not arbitrary — it reflects the LLM's attention priority (instructions at the top have higher adherence).

Example: `phases/01-reframing.md`

```markdown
# Phase 01: Reframing

<persona>
  You are Oraculo in the Reframing phase.
  Your focus: separate the problem from any proposed solution.
</persona>

<context-boundaries>
  Available: Raw idea from user, reasoning level from 00-setup
  Not available: Codebase analysis (not yet started)
  Focus: Problem clarity, not solution exploration
  References: Load references/question-bank.md section "Reframing" if needed
</context-boundaries>

<critical>
  Do NOT accept a solution description as a problem statement.
  "I want to add a cache" is a solution. "Pages load too slowly" is a problem.
</critical>

<rationalization-table>
  "The user clearly described the problem" → Did they describe a problem or a solution they already chose?
  "Reframing will frustrate the user"     → Skipping reframing produces requirements that solve the wrong problem.
</rationalization-table>

<anti-patterns>
  "Premature solutioning" — accepting HOW as the starting point instead of WHY
  "Shallow reframing" — accepting the first restatement without probing deeper
  "Scope inflation" — allowing the problem statement to absorb adjacent problems
</anti-patterns>

## Questions

### Deep Reasoning
- "You described what you want to build. What problem does this solve for the user?"
- "If this feature didn't exist, what would the user do? What's the cost of that alternative?"
- "What is explicitly out of scope for this exploration?"

### Light Reasoning
- "You received this task. What's the problem it solves?"
- "What is explicitly out of scope?"

### Epic-Derived
- "The epic describes [REC-N context]. What's the specific problem this story addresses?"

## Execution

Ask one question at a time. Wait for the user's answer before proceeding.

After each answer, evaluate:
- Did the user articulate a PROBLEM (not a solution)?
- Is the problem statement independent from implementation?
- Are scope boundaries emerging?

When all three are satisfied, proceed to the phase gate.

<iteration-limit phase="reframing" max="3">
  If after 3 attempts the user cannot separate problem from solution:
  1. Record the proposed solution as-is
  2. Ask: "What outcome do you expect this solution to produce for the user?"
  3. Use the answer as a proxy problem statement
  Do not insist indefinitely — adapt the approach.
</iteration-limit>

<halt>
  - User explicitly refuses to engage with reframing — note the gap, do not force
  - Problem statement is identical to the raw idea after 3 iterations — flag as risk
</halt>

<phase-gate phase="reframing">
  Exit conditions:
    - Problem statement articulated, independent from any specific solution
    - Raw idea preserved as originally stated (separate from reframed problem)
    - Scope boundaries defined (at least one explicit exclusion)

  Persist via CLI:
    - oraculo phase complete reframing --session=$SESSION_ID

  If CLI rejects: surface rejection to user. Do not retry with looser criteria.
  On success: Read phases/02-divergence.md
</phase-gate>
```

**Fixed section order:**

| Order | Section | Required | Purpose |
|-------|---------|:--------:|---------|
| 1 | `<persona>` | Yes | Phase focus — read first |
| 2 | `<context-boundaries>` | Yes | What the LLM has/doesn't have available |
| 3 | `<critical>` | No | Inviolable rules — immediate reinforcement |
| 4 | `<rationalization-table>` | No | Inoculation before execution |
| 5 | `<anti-patterns>` | Yes | Named failure modes |
| 6 | Questions / Execution | Yes | The phase's actual work (free Markdown) |
| 7 | `<iteration-limit>` | No | Retry limits |
| 8 | `<halt>` | No | Stop conditions |
| 9 | `<phase-gate>` | Yes | Always last — exit criteria |

The semantic tags (persona, boundaries, critical) come first because the LLM needs them before acting. The flow tags (iteration-limit, halt, phase-gate) come after because they control when to stop, not how to start. The executive content (questions, execution logic) sits in the middle — the actual work, framed by guardrails on both sides.

## 5. Context Management

Three rules govern how context is used throughout a session.

### Rule 1: Only one phase file in focus at a time

SKILL.md loads the current phase's file via `Read`. When the phase completes and the `<phase-gate>` advances, the next phase file is loaded. The previous one naturally falls out of the LLM's focus as the conversation grows.

```
Session starts → Read SKILL.md → Read phases/00-setup.md
Setup completes → Read phases/01-reframing.md
Reframing completes → Read phases/02-divergence.md
...
```

The LLM never loads multiple phase files simultaneously. Never anticipates future phases. Never creates "mental todo lists" of the complete workflow — the `<workflow>` in SKILL.md is the map, not the instruction.

### Rule 2: References are loaded by explicit instruction

Each `<context-boundaries>` declares which references can be loaded and when. The phase file never says "read all of question-bank.md" — it says which section.

```xml
<context-boundaries>
  References: Load references/question-bank.md section "JTBD" when starting Four Forces
  References: Load references/frameworks.md section "Assumption Mapping" when scoring begins
</context-boundaries>
```

If a reference is not declared in the current phase's `<context-boundaries>`, the LLM does not load it. This prevents speculative loading that wastes context.

### Rule 3: State lives in the CLI, not in the conversation

Validated results are persisted immediately via CLI in the `<phase-gate>`. If the session drops, the CLI has everything that was validated. The conversation is ephemeral — the artifact is what survives.

Practical consequences:
- The command never accumulates state internally ("so far we've identified X assumptions...")
- If the LLM needs data from previous phases, it queries the CLI (`oraculo session state`)
- If context compresses (context window overflow), the current phase file + CLI state are sufficient to continue

### Session Resumption

SKILL.md includes a bootstrap instruction that checks for an active session before anything else:

```xml
<phase-gate phase="bootstrap">
  1. Call: oraculo session status --type=epic
  2. If active session found:
     - Read phase file for current phase
     - Display to user: "Active session found. Resuming from [phase name]."
  3. If no active session:
     - Read phases/00-setup.md
     - CLI creates session on first phase completion
</phase-gate>
```

The CLI is the truth. The command is stateless.

### Loading Order Enforcement

Loading order is enforced through three layers:

**Layer 1 — Explicit navigation (prompt):** Each `<phase-gate>` ends with the exact next file instruction (`On success: Read phases/02-divergence.md`). The LLM doesn't need to decide what comes next — it's written.

**Layer 2 — Contextual signal (structure):** If the LLM skips ahead, the target phase's `<context-boundaries>` describes inputs that don't exist yet, creating natural friction.

**Layer 3 — CLI rejects (deterministic):** The CLI validates the sequence. `oraculo phase complete divergence` fails if `reframing` wasn't completed first. The command guides behavior; the CLI guarantees correctness.

## 6. Oraculo Commands

Each command maps to a phase of the operating model defined in `docs/design.md`.

### Catalog

| Command | Operating Model Phase | Input | Output |
|---------|----------------------|-------|--------|
| `/oraculo:epic` | Discover | Raw idea or problem | Validated `requirements.md` via CLI |
| `/oraculo:story` | Discover (light) | Work item or REC-N from epic | Story `requirements.md` via CLI |
| `/oraculo:plan` | Plan | Validated requirements | Task DAG in SQLite via CLI |
| `/oraculo:execute` | Execute | Ready DAG | Agents execute, code committed |
| `/oraculo:validate` | Validate | Complete implementation | QA verdict via CLI |

### `/oraculo:epic`

The densest command. Guides the complete Socratic exploration of the Double Diamond.

```
skills/oraculo/epic/
  SKILL.md
  phases/
    00-setup.md              # Triage, reasoning level, session check
    01-reframing.md          # Separate problem from solution
    02-divergence.md         # JTBD, Four Forces, edge cases
    03-codebase-analysis.md  # Dispatch research agents
    04-convergence.md        # Narrow to validated problem
    05-assumptions.md        # Map and score assumptions
    06-exit-gate.md          # Four risks validation + self-check
    07-artifact.md           # Generate requirements via CLI
  references/
    question-bank.md         # Questions by phase and reasoning level
    frameworks.md            # JTBD, Four Forces, TOC, Assumption Mapping
    artifact-templates.md    # Requirements document template
```

8 phases. Long session. All XML tags are used.

### `/oraculo:story`

Compact version of epic. Same structural pattern, fewer phases.

```
skills/oraculo/story/
  SKILL.md
  phases/
    00-setup.md              # Triage, parent epic check, reasoning level
    01-reframing.md          # Focused problem/scope definition
    02-assumptions.md        # Quick assumption check (2-3 critical)
    03-exit-gate.md          # Four risks (light version)
    04-artifact.md           # Generate story document via CLI
  references/
    question-bank.md         # Story-specific questions
    artifact-templates.md    # Story document template
```

5 phases. Shorter session. Inherits epic context when derived. The `00-setup.md` reads the epic from CLI when passed as argument.

### `/oraculo:plan`

Consumes requirements and produces the execution DAG. Here Oraculo shifts mode — from Socratic guide to technical planner.

```
skills/oraculo/plan/
  SKILL.md
  phases/
    00-setup.md              # Load requirements, validate completeness
    01-decomposition.md      # Break requirements into tasks
    02-dependencies.md       # Model DAG, identify constraint (TOC)
    03-optimization.md       # Maximize parallelism, minimize critical path
    04-artifact.md           # Persist DAG via CLI
  references/
    decomposition-patterns.md  # How to break requirements into tasks
```

5 phases. Input: validated requirements. Output: DAG in SQLite. Oraculo may dispatch research agents to analyze the codebase during decomposition.

### `/oraculo:execute`

The most different command. Oraculo does not execute — it orchestrates. This command assembles the agent team and delegates.

```
skills/oraculo/execute/
  SKILL.md
  phases/
    00-setup.md              # Load DAG from CLI, verify readiness
    01-team-assembly.md      # Select agents, assign tasks per DAG
    02-monitoring.md         # Track progress, handle failures
    03-completion.md         # Verify all tasks done, prepare for QA
  references/
    agent-dispatch.md        # How to spawn and configure agents
```

4 phases. This command interacts more with the agent system (`docs/agents/design.md`) than with the user. The `<persona>` shifts: Oraculo here is orchestrator, not Socratic guide.

### `/oraculo:validate`

Dispatches the QA agent with clean context (fresh eyes). Receives the verdict and decides the next step.

```
skills/oraculo/validate/
  SKILL.md
  phases/
    00-setup.md              # Load implementation state, prepare QA context
    01-qa-dispatch.md        # Spawn QA agent with clean context
    02-verdict.md            # Process QA result, decide next step
  references/
    qa-criteria.md           # What the QA agent checks
```

3 phases. If QA rejects, the command identifies the correct phase for return and instructs the user.

### Common Pattern

All commands share:
- Same directory structure (SKILL.md + phases/ + references/)
- Same section order in phase files
- Same XML vocabulary (10 tags)
- Same phase-gate mechanics (CLI persists, CLI validates)
- Same base persona (adapted per phase)

What varies:
- Number of phases (3 to 8)
- Guardrail density (epic > story > plan > execute > validate)
- Persona tone (Socratic in epic/story, orchestrator in execute/validate)
- Available references (question banks in epic/story, patterns in plan, dispatch rules in execute)

## 7. Layer Interaction

The philosophy defines three layers with absolute responsibilities: Command, CLI, and Agents.

### The Contract

```
Command (interpretive)     CLI (deterministic)       Agents (execution)
──────────────────────     ─────────────────────     ──────────────────
Guides conversation        Validates and persists    Execute code
Decides WHEN to ask        Decides IF it accepts     Follow TDD
Produces artifacts         Guarantees schema/seq     Work on the DAG
Calls the CLI              Rejects or confirms       Report results
```

**Inviolable rule:** No layer does another layer's work.

- The command never writes directly to SQLite, constructs file paths, or validates schemas
- The CLI never asks the user questions, interprets intent, or adapts tone
- Agents never do discovery, modify requirements, or decide scope

### Handoffs Between Commands

Each command ends by suggesting the next, but never invokes it automatically. The user decides when to advance.

```xml
<!-- In epic's 07-artifact.md -->
<phase-gate phase="artifact">
  ...
  On success:
    Display: "Requirements document saved. To decompose into stories, run /oraculo:story <epic-name>"
    Do NOT invoke /oraculo:story automatically.
    The user decides when to proceed.
</phase-gate>
```

Each transition is a user decision point. The user may want to review requirements with the team, wait, or adjust. Oraculo suggests, never decides.

### Agent Interaction

The commands `plan`, `execute`, and `validate` interact with the agent system defined in `docs/agents/design.md`.

| Command | Agent Type | When | How |
|---------|-----------|------|-----|
| `epic` | Research agent | Phase 03 (Codebase Analysis) | Task tool, parallel to conversation |
| `story` | None | — | Purely Socratic session |
| `plan` | Research agent (optional) | Phase 01 (Decomposition) | Codebase analysis to inform tasks |
| `execute` | Code agents + Research agents | Phases 01-02 | Team assembly via TeamCreate |
| `validate` | QA agent | Phase 01 | Task tool, clean context |

The command never configures the agent directly. It declares intent and the dispatch follows the rules in `docs/agents/design.md`.

### Rejection and Return

When QA rejects in `validate`, the command identifies the return phase:

```xml
<phase-gate phase="verdict">
  If QA verdict is REJECTED:
    Analyze rejection reasons:
    - Implementation bug → return to /oraculo:execute (re-execute failed tasks)
    - Missing requirement → return to /oraculo:plan (re-decompose)
    - Wrong requirement → return to /oraculo:story or /oraculo:epic (re-discover)

    Display the verdict, the analysis, and the suggested return point.
    The user decides which command to re-invoke.
    Do NOT re-invoke automatically.
</phase-gate>
```

Oraculo never forces a path. It analyzes, recommends, and the user decides.

## 8. Evolution Criteria

The XML vocabulary has 10 tags today. The structure has 5 commands. This section defines how the system evolves without degrading.

### Adding a New XML Tag

A new tag must pass three criteria. If it fails any one, it does not enter the vocabulary.

| Criterion | Test |
|-----------|------|
| **Maps to philosophy** | Does the tag solve a problem described in `docs/commands/philosophy.md`? If no principle supports it, it is invention without foundation. |
| **Markdown doesn't solve it** | Can the same effect be achieved with headers, bold, tables, or code blocks? If yes, use Markdown — an extra tag is cost without benefit. |
| **Evidence of need** | Was there documented drift or failure that this tag would have prevented? Tags created preventively, without evidence of a real problem, dilute the vocabulary. |

**Process:** Document the justification in the design doc before implementing. The tag enters the vocabulary with its origin (which project inspired it or which failure motivated it) and a usage example.

**Soft limit:** If the vocabulary exceeds 15 tags, revisit existing ones. One is probably not justifying its cost.

### Adding a New Phase to a Command

| Criterion | Test |
|-----------|------|
| **Distinct exit condition** | Does the phase have an exit condition that doesn't overlap with adjacent phases? If the exit condition is the same as the previous or next phase, it's not a phase — it's part of another. |
| **Loading justifies separation** | Is the phase's content large enough to justify a separate file? If it fits in 20 lines, it's a section within another phase, not a new phase. |
| **CLI validates the transition** | Does the CLI need to validate something between this phase and the next? If there's nothing to persist or validate, the separation is cosmetic. |

### Adding a New Command

Commands map to operating model phases. A new command implies a new phase in the model — an architectural change.

| Criterion | Test |
|-----------|------|
| **Operating model phase** | Does the command map to a phase not covered by any existing command? If covered, it's an extension of an existing command, not a new one. |
| **Distinct artifact** | Does the command produce an artifact that no other command produces? If the output is the same type as another command, it's probably a variation, not a new command. |
| **Clear handoff** | Can you define what comes before and after this command in the flow? Commands without connection to the flow are loose features. |

### Removing Elements

Removal follows the inverse criterion: if a tag, phase, or command has not been used in the last N development cycles, or if the problem that motivated its creation no longer exists, it is a candidate for removal.

### What This Document Does NOT Govern

- Phase file content (questions, frameworks, templates) — that is the domain of the Epic and Story designs
- CLI behavior (validations, schemas, persistence) — that is the domain of `docs/cli/design.md`
- Agent behavior (TDD, dispatch, QA) — that is the domain of `docs/agents/design.md`

This document governs how commands are built — the anatomy, the vocabulary, the structure, and the rules of evolution. Content lives in the designs of each phase.
