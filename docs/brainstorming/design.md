# Brainstorming Phase — System Design

## 1. Session Structure

The Brainstorming phase follows a structured flow with two macro-movements: **Diverge** (expand the problem space) then **Converge** (narrow to a validated definition). Oraculo progresses through seven steps, advancing only when each step's exit condition is satisfied.

### 1.1 Entry — Reframing

The user arrives with an idea. Oraculo separates the problem from any proposed solution before proceeding.

**Trigger questions:**
- "You described what you want to build. What problem does this solve for the user?"
- "If this feature didn't exist, what would the user do? What's the cost of that alternative?"
- "What is explicitly out of scope for this exploration?"

**Exit condition:** The user has articulated a problem statement that is independent from any specific solution. Oraculo records the raw idea as originally stated and the reframed problem as separate artifacts.

### 1.2 Divergence — Expanding the Problem Space

Oraculo applies JTBD and Four Forces question templates to explore dimensions the user hasn't considered.

**JTBD questions:**
- "In what specific situation does the user encounter this problem?"
- "What is the user trying to accomplish — functionally, emotionally, socially?"
- "What alternatives exist today? Why are they insufficient?"
- "What would make the user 'hire' a new solution for this job?"

**Four Forces questions:**
- **Push:** "What's frustrating enough about the current state to motivate change?"
- **Pull:** "What makes this solution more attractive than the status quo?"
- **Anxiety:** "What might make users hesitate? Learning curve? Data loss? Dependencies?"
- **Habit:** "What does the user do automatically today that this would need to replace or integrate with?"

**Edge case questions:**
- "What happens if the user does [unexpected behavior]?"
- "How does this behave with empty state, limit data, or concurrent users?"
- "Who could be negatively affected by this change even if they're not the target user?"

**Exit condition:** Oraculo has explored at least the three JTBD dimensions (functional, emotional, social) and all four forces. The user has identified at least one edge case or risk they hadn't previously considered.

### 1.3 Codebase Analysis (Parallel)

While the dialogue progresses, Oraculo dispatches sub-agents to analyze the existing codebase. This runs in parallel — it does not block the conversation.

**What sub-agents investigate:**
- Existing implementations that overlap with the proposed idea
- Architectural patterns the new feature must respect
- Tests that cover related behavior (potential breakage points)
- Technical constraints that limit or enable specific approaches
- Past decisions recorded in SQLite that relate to this domain

**How results enter the conversation:**
- "The analysis found that [component X] already handles a similar case. Do you want to build on that or create something independent? Why?"
- "There's an edge case in [area Y] that the current code explicitly handles. How does your idea behave in that scenario?"
- "A previous decision rejected [approach Z] for [reason]. Does that reasoning still apply?"

**Timing:** Codebase analysis is triggered at the start of Divergence (step 1.2). Results are introduced into the dialogue as they become available — during Divergence if ready early, or during Convergence if analysis takes longer.

### 1.4 Convergence — Defining the Problem

Oraculo shifts from expanding to narrowing. This transition happens after the Divergence exit condition is met.

**Convergence questions:**
- "From everything we explored, what is the core pain — the root cause, not the symptom?"
- "If you had to write the problem in one sentence from the user's perspective, what would it be?"
- "What business metric does this problem directly affect?"
- "If you could solve only one aspect of this, which would deliver the most value?"

**Exit condition:** The user has produced a clear problem statement (one sentence, from the user's perspective) and identified a target outcome (measurable metric or observable user behavior change).

### 1.5 Assumption Mapping

Oraculo surfaces hidden assumptions and scores them by risk.

**Assumption categories:**
- **Problem hypothesis:** "User X suffers from Y in situation Z."
- **Solution hypothesis:** "If we build A, it will resolve Y."
- **Value hypothesis:** "The user will perceive this as valuable enough to change behavior."
- **Feasibility hypothesis:** "This can be built within the current architecture without major refactoring."

**Risk scoring mechanism:**
Each assumption is evaluated on two axes:
- **Impact** (1-5): if wrong, how badly does it break the idea?
- **Evidence** (1-5): how much data supports this? (1 = pure intuition, 5 = validated data)
- **Risk score** = Impact × (6 - Evidence). Higher score = higher priority.

Assumptions are recorded in SQLite with status "unvalidated." Oraculo explicitly states: "I'm recording this as an unvalidated assumption. Do you have any evidence that already confirms part of this?"

**Exit condition:** All assumptions with risk score above threshold are acknowledged by the user. Each critical assumption (high impact, low evidence) has either supporting evidence or is flagged as a known risk for the Plan phase.

### 1.6 Stress Test — Working Backwards

Oraculo runs a simplified PR/FAQ to test whether the idea has sufficient clarity to proceed.

**Stress test questions:**
- "If this feature launched tomorrow, how would the announcement read in two sentences?"
- "What would a skeptical customer's first question be?"
- "What would engineering's main objection be?"
- "How would you measure success three months after launch?"

**Exit condition:** The user can answer all four questions without significant hesitation or contradiction with earlier answers. If the user struggles, Oraculo identifies which gap exists and loops back to the step that addresses it (typically 1.2 or 1.4).

### 1.7 Exit Gate

The final checkpoint before advancing to Plan. Oraculo evaluates four risk dimensions:

**Exit checklist (Four Risks):**
1. **Value** — Is the user value clear and the need validated (or explicitly marked as assumption)?
2. **Usability** — Is the solution understandable and usable without excessive friction?
3. **Technical feasibility** — Has the codebase analysis confirmed viability (or flagged known risks)?
4. **Business viability** — Are success metrics defined and impact on the business understood?

**Gate behavior:** Oraculo evaluates each risk and presents the assessment to the user. If any risk is unanswered, Oraculo identifies the specific gap and routes back to the appropriate step:
- Value gap → return to 1.2 (Divergence)
- Usability gap → return to 1.2 (Edge case exploration)
- Technical feasibility gap → return to 1.3 (Codebase Analysis)
- Business viability gap → return to 1.4 (Convergence)

Only when all four risks are addressed does the session produce its final artifacts and hand off to Plan.

## 2. Output Artifacts

The Brainstorming phase produces four structured artifacts that feed directly into the Plan phase.

### 2.1 Job Stories

One or more Job Stories in the format:

> "When [specific situation/context], I want [motivation/job to be done], so I can [expected measurable outcome]."

Job Stories become the input for task decomposition in the Plan phase. Each task in the DAG must trace back to a Job Story.

### 2.2 Opportunity Solution Tree (OST)

A structured tree persisted in SQLite:

```
Outcome (business metric / user behavior)
  └── Opportunity (user need, written in first person)
        └── Solution (proposed approach)
              └── Assumption (hypothesis + validation criteria)
```

The tree grows across sessions. When the user returns with a new idea, Oraculo queries existing opportunities: "In a previous session, we identified opportunity X linked to outcome Y. Does this new idea address the same opportunity or a new one?"

### 2.3 Assumption Register

A prioritized list of assumptions with:
- Description of the assumption
- Category (problem / solution / value / feasibility)
- Risk score (impact × lack of evidence)
- Status: unvalidated / validated / refuted
- Associated evidence (if any)

Critical assumptions (high risk) become explicit inputs for the Plan phase — they may generate validation tasks before any implementation begins.

### 2.4 Codebase Impact Summary

Results from sub-agent analysis, recorded as:
- Existing components that relate to the idea
- Architectural constraints identified
- Edge cases from current code that the new feature must handle
- Past decisions from SQLite that are relevant

## 3. Scaling with Complexity

Oraculo selects the brainstorming depth based on the idea's characteristics at entry.

### 3.1 Minimal Brainstorming

**When:** The user describes a small, well-scoped change to existing behavior. The problem is clear and the solution space is narrow.

**Flow:** Reframing (1.1) → basic assumption check (1.5) → Exit Gate (1.7). Typically 3-5 questions. The exit gate still applies — all four risks must be addressed even for small changes.

### 3.2 Standard Brainstorming

**When:** The user describes a new feature or a significant change to existing functionality. The problem needs exploration but the domain is understood.

**Flow:** Full sequence — Reframing → Divergence → Codebase Analysis → Convergence → Assumption Mapping → Stress Test → Exit Gate.

### 3.3 Deep Brainstorming

**When:** The idea is vague, ambitious, cross-cutting, or the user cannot clearly articulate the problem even after initial reframing.

**Flow:** Full sequence with parallel sub-agents exploring multiple angles simultaneously:
- Agent analyzing codebase impact across multiple domains
- Agent exploring technical feasibility of different approaches
- Agent reviewing past decisions and related features in SQLite

Results are consolidated and fed back into the dialogue, enabling richer questioning grounded in evidence.

### 3.4 Complexity Selection Criteria

Oraculo evaluates at entry:

| Signal | Minimal | Standard | Deep |
|--------|---------|----------|------|
| User can state the problem clearly | Yes | Partially | No |
| Solution space is narrow | Yes | Moderate | Wide or unknown |
| Affected codebase area | Single component | Multiple components | Cross-cutting |
| Past decisions exist in SQLite | None relevant | Some relevant | Many relevant or contradictory |

## 4. Integration with Oraculo's Operating Model

### 4.1 Relationship with Plan Phase

The Brainstorming phase produces the requirements document referenced in the main design:
- Job Stories define **what** needs to happen
- The OST provides **why** (connected to outcomes)
- The Assumption Register defines **what must be validated first**
- The Codebase Impact Summary defines **constraints** for the DAG

The Plan phase receives these artifacts and decomposes them into a task DAG. No task enters the DAG without traceability to a Job Story and an Outcome.

### 4.2 Entry Point

The Brainstorming phase is invoked via `/oraculo:brainstorm`. Any team member can start it — Product, Development, or anyone with an idea. Oraculo adapts the depth and language to the user's context.

### 4.3 Persistence in SQLite

Everything from the Brainstorming phase is recorded:
- The raw idea as originally stated
- The reframed problem
- All questions asked and answers given
- The OST built during the session
- Assumptions with their risk scores and status
- Codebase analysis results
- The exit gate evaluation
- The final Job Stories

This enables continuity across sessions and team members. Any team member can query SQLite to understand the full reasoning behind the requirements — not just what was decided, but why, and what was explicitly rejected.
