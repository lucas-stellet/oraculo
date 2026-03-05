# Story Phase — System Design

## 1. Session Structure

The Story phase follows a compact flow: four phases that move from problem understanding to artifact generation. Oraculo advances only when each phase's exit condition is satisfied.

### 1.0 Session Setup

The user arrives with a work item. Oraculo checks for a parent epic, captures the raw description, detects reasoning level, and confirms the work is story-sized.

**Parent epic workflow:**
- If project name was passed as argument (`/oraculo:story gastos-app`):
  - Search for `.oraculo/epics/<epic-name>/requirements.md`
  - If found: verify `approval_status = approved` — if not approved, block decomposition and inform the user that the parent epic is still pending human review; if approved, read the epic and ask which REC-N to work on (or if it's something new)
  - If not found: proceed as standalone
- If no argument: proceed as standalone

**Escalation check:** If the work is too large (multiple areas, high uncertainty, vague scope), suggest `/oraculo:epic` instead.

**Exit condition:** Work item captured, reasoning level determined internally, parent epic loaded and approval verified (if applicable), story-sized scope confirmed.

### 1.1 Reframing

Oraculo separates the problem from any proposed solution and defines scope boundaries.

**Deep questions:**
- "You described what you want. What problem does this solve for the user?"
- "What is explicitly out of scope for this story?"

**Light questions:**
- "You received this task. What's the problem it solves?"
- "What is explicitly out of scope?"

**Epic-derived questions:**
- "The epic describes [REC-N context]. What's the specific problem this story addresses?"
- "What parts of the requirement are NOT covered by this story?"

**Exit condition:** Problem clearly stated and separated from solution. Scope boundaries defined.

### 1.2 Assumption Check

Oraculo surfaces the critical assumptions behind the work.

**Deep (value focus):**
- "What are we assuming about the user that, if wrong, makes this story pointless?"
- "What's the biggest risk — building it wrong, or building the wrong thing?"

**Light (feasibility focus):**
- "What are we assuming about the current system that this depends on?"
- "What's the biggest technical risk?"

**Exit condition:** 2-3 critical assumptions identified and acknowledged. Each has an evidence level (high/medium/low).

### 1.3 Exit Gate

Final checkpoint — evaluate four risk dimensions before generating the artifact.

**Four Risks:**
1. **Value** — Is it clear why someone wants this?
2. **Usability** — Can it be used without friction?
3. **Technical feasibility** — Is it buildable within the current system?
4. **Business viability** — Does it align with goals / are criteria testable?

**Light escalation:** When the dev can't answer Value or Business Viability questions (requires product context), note the gap in the story artifact rather than looping back.

**Exit condition:** All four risks addressed (passed or explicitly accepted).

### 1.4 Artifact Generation

Oraculo generates the Story Document and saves it via CLI, then creates a version for human review. The workflow has four steps:

1. **Generate and save** — Oraculo generates the Story Document and saves it to `.oraculo/epics/<epic-name>/stories/<story-name>/requirements.md` via `oraculo story save`. If derived from an epic, the document includes parent references in header comments.
2. **Create version** — Oraculo calls `oraculo tools story version <story-name> --epic <epic-name>`, passing the markdown through stdin. This creates a versioned snapshot and displays it on the dashboard for human review.
3. **Await verdict** — The agent monitors reviews via `oraculo tools review list <version-id> --type story`. Workflow does not advance until a human issues a verdict from the dashboard.
4. **Handle verdict:**
   - `approved` — Story definition is finalized. The story is ready for execution.
   - `rejected` — Return to §1.1 (Reframing) to reopen the problem space based on reviewer feedback.

## 2. Saturation

The Story skill tracks 5 slots to determine when enough information has been gathered:

| Slot | Deep weight | Light weight |
|------|:-----------:|:------------:|
| Problem statement | Standard | **High** |
| Desired outcome | Standard | **High** |
| Acceptance criteria | Standard | Standard |
| Scope boundaries | Low | Standard |
| Key assumptions | Standard | Standard |

**Saturation rule:** All High slots complete + all Standard slots at least partial → present the draft.

**Weight definitions:**
- **High** — Must reach **complete** for saturation. Requires dedicated questions.
- **Standard** — Must reach at least **partial** for saturation.
- **Low** — Does not block saturation. A single indirect mention counts as partial.

## 3. Reasoning Levels

The Story phase operates at two reasoning levels that share the same Socratic discipline but differ in focus.

### 3.1 Light Reasoning (Executor/Developer)

- Received a task from someone else
- Technical expertise, limited product context
- Focus: understanding the task correctly, technical feasibility, edge cases
- Assumption check focuses on **feasibility** (architecture fit, dependencies)
- Escalates to PM on value/business gaps (doesn't loop back)

### 3.2 Deep Reasoning (Decision-Maker/PM)

- Exploring a change they identified as needed
- Has user context, business understanding
- Focus: user value, problem clarity, measurable outcome
- Assumption check focuses on **value** (is this the right problem?)
- Answers all gate questions autonomously

### 3.3 Selection Signals

| Signal | Light | Deep |
|--------|-------|------|
| User describes a task received from someone else | Yes | — |
| User describes an idea or problem they identified | — | Yes |
| User has business metrics or user data to share | — | Yes |
| User focuses on implementation approach | Yes | — |
| User's role is primarily technical | Yes | — |
| User's role involves product decisions | — | Yes |

## 4. Parent Epic Workflow

When a story is derived from an epic:

1. **Reading:** The Story skill reads `.oraculo/epics/<epic-name>/requirements.md`
2. **Selection:** Asks the user which REC-N to work on, or if it's something new
3. **Context inheritance:** The epic's problem statement, context, and scope inform the story session — the user doesn't repeat what was already captured
4. **Adapted questions:** Reframing questions reference the specific REC-N context
5. **Output reference:** The story document includes `parent_epic` and `parent_rec` metadata

## 5. Output: Story Template

The Story Document has 4 sections:

1. **Motivation** — Problem and context (from user's or system's perspective)
2. **Requirement** — Single user/system story with GIVEN/WHEN/THEN acceptance criteria
3. **Scope** — What's in, what's explicitly out
4. **Assumptions** — Key assumptions with evidence levels

The document captures WHAT and WHY — never HOW. No technical implementation details.

## 6. Escalation

If at any point during a Story session the work reveals itself to be epic-sized, Oraculo suggests escalating:

**Signals for escalation:**
- Multiple areas or disciplines involved
- High uncertainty about the approach
- Scope keeps expanding during reframing
- User can't articulate a focused problem statement

**Behavior:** Suggest `/oraculo:epic` with a clear explanation of why the full exploration would be more appropriate. Respect the user's decision if they want to continue as a story.

## 7. Output Path

Story artifacts are saved to `.oraculo/epics/<epic-name>/stories/<story-name>/requirements.md` inside the target application. The story name is derived from the story title (lowercase, hyphenated). If no epic name was provided and the story is standalone, the CLI creates a lightweight epic automatically.

The artifact is not considered final until a version review with verdict `approved` exists. A saved story document that has not yet received a human verdict (or that received `rejected`) must not be used as input for execution planning.
