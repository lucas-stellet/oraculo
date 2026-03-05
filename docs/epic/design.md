# Epic Phase — System Design

## 1. Session Structure

The Epic phase follows a structured flow with two macro-movements: **Diverge** (expand the problem space) then **Converge** (narrow to a validated definition). Oraculo progresses through eight phases, advancing only when each phase's exit condition is satisfied.

### 1.0 Entry — Session Setup & Triage

The user arrives with an idea. Oraculo captures the raw idea, detects reasoning level, and confirms this is epic-sized work.

**Triage signals for epic scope:**
- Involves more than one area or discipline
- Has parts that could be delivered independently
- Significant uncertainties about how to solve it

If the work appears small and focused, Oraculo suggests `/oraculo:story` instead.

**Exit condition:** Raw idea captured, reasoning level determined internally, epic-sized scope confirmed.

### 1.1 Reframing

Oraculo separates the problem from any proposed solution before proceeding.

**Trigger questions:**
- "You described what you want to build. What problem does this solve for the user?"
- "If this feature didn't exist, what would the user do? What's the cost of that alternative?"
- "What is explicitly out of scope for this exploration?"

**Exit condition:** The user has articulated a problem statement that is independent from any specific solution. Oraculo records the raw idea as originally stated and the reframed problem as separate artifacts.

### 1.2 Divergence — Expanding the Problem Space

Oraculo applies JTBD question templates (including Four Forces) to explore dimensions the user hasn't considered.

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

When no prior context exists (no documentation, no defined sources), Oraculo dispatches research agents to analyze the existing codebase. This runs in parallel — it does not block the conversation. When prior context exists (defined sources, existing documentation), the orchestrator conducts a light analysis directly without dispatching agents.

**What sub-agents investigate:**
- Existing implementations that overlap with the proposed idea
- Architectural patterns the new feature must respect
- Tests that cover related behavior (potential breakage points)
- Technical constraints that limit or enable specific approaches

**How results enter the conversation:**
- "The analysis found that [component X] already handles a similar case. Do you want to build on that or create something independent? Why?"
- "There's an edge case in [area Y] that the current code explicitly handles. How does your idea behave in that scenario?"

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

**Exit condition:** All assumptions with risk score above threshold are acknowledged by the user. Each critical assumption (high impact, low evidence) has either supporting evidence or is flagged as a known risk for the Plan phase.

### 1.6 Exit Gate

The final checkpoint before advancing to the Approval Gate (§1.7). Oraculo evaluates four risk dimensions:

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

Only when all four risks are addressed does the session produce its final artifacts and hand off to the Approval Gate (§1.7).

### 1.7 Artifact Generation

Oraculo generates the Requirements Document and saves it via CLI, then creates a version for human review. The workflow has four steps:

1. **Generate and save** — Oraculo generates the Requirements Document and saves it to `.oraculo/epics/<epic-name>/requirements.md` via `oraculo epic save`.
2. **Create version** — Oraculo calls `oraculo tools epic version <epic-name>`, passing the markdown through stdin. This creates a versioned snapshot and displays it on the dashboard for human review.
3. **Await verdict** — The agent monitors reviews via `oraculo tools review list <version-id> --type epic`. Workflow does not advance until a human issues a verdict from the dashboard.
4. **Handle verdict:**
   - `approved` — Requirements are finalized. Oraculo suggests decomposing the epic into stories with `/oraculo:story`.
   - `rejected` — Return to §1.2 (Divergence) to reopen the problem space based on reviewer feedback.

## 2. Output Artifacts

The Epic phase produces structured artifacts that feed directly into stories and the Plan phase.

### 2.1 Job Stories

One or more Job Stories in the format:

> "When [specific situation/context], I want [motivation/job to be done], so I can [expected measurable outcome]."

Job Stories become the input for story decomposition. Each story must trace back to a Job Story or REC-N.

### 2.2 Assumption Register

A prioritized list of assumptions with:
- Description of the assumption
- Category (problem / solution / value / feasibility)
- Risk score (impact × lack of evidence)
- Status: unvalidated / validated / refuted
- Associated evidence (if any)

Critical assumptions (high risk) become explicit inputs for the Plan phase.

### 2.3 Codebase Impact Summary

Results from sub-agent analysis, recorded as:
- Existing components that relate to the idea
- Architectural constraints identified
- Edge cases from current code that the new feature must handle

### 2.4 Requirements Document

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

## 3. Scaling with Complexity

Oraculo selects the exploration depth based on the idea's characteristics at entry. The Epic phase uses only two complexity levels:

### 3.1 Standard Epic

**When:** The user describes a new feature or a significant change to existing functionality. The problem needs exploration but the domain is understood.

**Flow:** Full sequence — Session Setup → Reframing → Divergence → Codebase Analysis → Convergence → Assumption Mapping → Exit Gate → Artifact Generation → Approval Gate.

### 3.2 Deep Epic

**When:** The idea is vague, ambitious, cross-cutting, or the user cannot clearly articulate the problem even after initial reframing.

**Flow:** Full sequence with parallel sub-agents exploring multiple angles simultaneously:
- Agent analyzing codebase impact across multiple domains
- Agent exploring technical feasibility of different approaches
- Agent reviewing related past patterns and conventions

Results are consolidated and fed back into the dialogue, enabling richer questioning grounded in evidence.

**Flow:** Full sequence with parallel sub-agents — Session Setup → Reframing → Divergence (+ parallel agents) → Codebase Analysis → Convergence → Assumption Mapping → Exit Gate → Artifact Generation → Approval Gate.

## 4. Reasoning Levels

The Epic phase operates at two reasoning levels that share the same Socratic discipline but differ in domain of inquiry.

### 4.1 Light Reasoning (Executor/Developer)

- Received a task from someone else
- Technical expertise, limited product context
- Focuses on: understanding task, technical risks, edge cases
- Codebase Analysis is the **center of gravity**
- Escalates to PM on product context gaps

### 4.2 Deep Reasoning (Decision-Maker/PM)

- Exploring idea they identified
- Has user data, business metrics, strategic vision
- Focuses on: user behavior, business impact, technical feasibility
- Uses full framework toolkit (JTBD with Four Forces, TOC, Assumption Mapping)
- Answers all gate questions autonomously

### 4.3 Selection Signals

| Signal | Light | Deep |
|--------|-------|------|
| User describes a task received from someone else | Yes | — |
| User describes an idea or problem they identified | — | Yes |
| User has business metrics or user data to share | — | Yes |
| User focuses on implementation approach | Yes | — |
| User's role is primarily technical | Yes | — |
| User's role involves product decisions | — | Yes |

### 4.4 Complexity × Reasoning Matrix

|  | Standard | Deep Complexity |
|--|----------|-----------------|
| **Light** | Full Light flow with all phases. Codebase Analysis is center of gravity. No business questions. | Full Light flow with parallel sub-agents analyzing codebase across multiple domains. Long session, no business questions. |
| **Deep** | Full flow — all frameworks, all questions, all gates. | Full flow with parallel sub-agents. Maximum duration and depth across all dimensions. |

## 5. Integration with Stories

The Epic phase produces a Requirements Document with REC-N items. To implement:

**Precondition:** Before decomposing into stories, verify that the epic's `approval_status = approved`. An epic whose requirements have not been approved by a human reviewer must not be decomposed — the artifact may still change based on the pending verdict.

1. Invoke `/oraculo:story <epic-name>`
2. The Story skill reads `.oraculo/epics/<epic-name>/requirements.md`
3. The user selects which REC-N to work on
4. The Story skill runs a focused session to produce an executable story definition

Each REC-N can generate one or more stories depending on its scope.

## 6. Output Path

All epic artifacts are saved to `.oraculo/epics/<epic-name>/requirements.md` inside the target application. The epic name is derived from the idea's domain or explicitly asked from the user during artifact generation.
