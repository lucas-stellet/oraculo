# Question Bank

All questions organized by phase and reasoning level. Referenced by the main epic skill.

---

## 0. Triage

### Epic scope confirmation

- "Does this involve more than one area or discipline?"
- "Are there parts that could be delivered independently?"
- "Are there significant uncertainties about how to solve this?"

If answers suggest small, focused, well-understood work → suggest `/oraculo:story` instead.

---

## 1. Reframing

### Deep

- "You described what you want to build. What problem does this solve for the user?"
- "If this feature didn't exist, what would the user do? What's the cost of that alternative?"
- "What is explicitly out of scope for this exploration?"

### Light

- "You received this task. What's the problem it solves, as described by whoever assigned it?"
- "How do you know you understood what was asked? What system behavior will change?"
- "What is explicitly out of scope?"

### Exit condition

- **Deep:** User has articulated a problem statement independent from any specific solution.
- **Light:** User can articulate the expected system behavior change independently from the implementation approach.

---

## 2. Divergence — JTBD

### Deep

- "In what specific situation does the user encounter this problem?"
- "What is the user trying to accomplish — functionally, emotionally, socially?"
- "What alternatives exist today? Why are they insufficient?"
- "What would make the user 'hire' a new solution for this job?"

### Light

- "What part of the code handles this case today? Why is it insufficient?"
- "What pre-conditions must exist for this behavior to trigger?"

---

## 3. Divergence — Four Forces

### Deep

- **Push:** "What's frustrating enough about the current state to motivate change?"
- **Pull:** "What makes this solution more attractive than the status quo?"
- **Anxiety:** "What might make users hesitate? Learning curve? Data loss? Dependencies?"
- **Habit:** "What does the user do automatically today that this would need to replace or integrate with?"

### Light

- **Push:** "What technical limitation is driving this task?"
- **Pull:** "What makes this approach better architecturally?"
- **Anxiety:** "What can break for other systems?"
- **Habit:** "What code patterns must be preserved?"

---

## 4. Divergence — Edge Cases

### Deep

- "What happens if the user does [unexpected behavior]?"
- "How does this behave with empty state, limit data, or concurrent users?"
- "Who could be negatively affected by this change even if they're not the target user?"

### Light

- "What happens with invalid inputs, concurrent state, or network errors?"
- "What other services, queues, or jobs depend on this component?"

### Exit condition (Divergence overall)

- **Deep:** At least three JTBD dimensions explored, all four forces mapped, one edge case the user hadn't considered surfaced.
- **Light:** Technical edge cases explored, system dependencies identified, one risk the user hadn't considered surfaced.

---

## 5. Convergence

### Deep

- "From everything we explored, what is the core pain — the root cause, not the symptom?"
- "If you had to write the problem in one sentence from the user's perspective, what would it be?"
- "What business metric does this problem directly affect?"
- "If you could solve only one aspect of this, which would deliver the most value?"

### Light

- "What is the core functional requirement — what must change in the system?"
- "Describe the expected system behavior in one sentence."
- "What observable behavior changes after deploy?"
- "If you could solve only one aspect, which delivers the most value?"

### Exit condition

- **Deep:** Clear problem statement (one sentence, user's perspective) + target outcome (measurable metric or user behavior change).
- **Light:** Clear system behavior statement + observable verification criteria.

---

## 6. Assumption Mapping

### Assumption categories

| Category | Deep | Light |
|----------|------|-------|
| Problem hypothesis | Fully explored and scored | Recorded as "external dependency" |
| Solution hypothesis | Fully explored and scored | Evaluated: "Does this implementation solve the use case?" |
| Value hypothesis | Fully explored and scored | Recorded as "external dependency" |
| Feasibility hypothesis | Explored alongside other categories | **Central focus.** Architecture fit, dependencies, regressions, performance |

### Prompts (both levels)

- "What are we assuming to be true that, if wrong, would break this idea?"
- "I'm recording this as an unvalidated assumption. Do you have any evidence that already confirms part of this?"

### Exit condition

All assumptions with risk score above threshold acknowledged by the user. Critical assumptions (high impact, low evidence) have either supporting evidence or are flagged as known risks for the Plan phase.

---

## 7. Stress Test

### Deep — PR/FAQ (Working Backwards)

- "If this feature launched tomorrow, how would the announcement read in two sentences?"
- "What would a skeptical customer's first question be?"
- "What would engineering's main objection be?"
- "How would you measure success three months after launch?"

### Light — Technical Viability Checklist

- "Does the chosen approach fit the current architecture without major refactoring?"
- "What tests must be written to cover the mapped edge cases?"
- "What would another dev's main objection be when reviewing this PR?"
- "How do you verify this works on deploy day? What are the technical acceptance criteria?"

### Exit condition

User answers all four questions without significant hesitation or contradiction. If the user struggles, identify which gap exists and loop back to the step that addresses it.

---

## 8. Exit Gate — Four Risks

### Deep

| Risk | Question | If gap |
|------|----------|--------|
| Value | "Is the user value clear and the need validated (or explicitly marked as assumption)?" | Return to Divergence |
| Usability | "Is the solution understandable and usable without excessive friction?" | Return to Divergence (edge cases) |
| Technical feasibility | "Has the codebase analysis confirmed viability (or flagged known risks)?" | Return to Codebase Analysis |
| Business viability | "Are success metrics defined and impact on the business understood?" | Return to Convergence |

### Light

| Risk | Question | If gap |
|------|----------|--------|
| Value | "Is the expected behavior described in the task clear enough to implement?" | **Escalate to PM** |
| Usability | "Is the solution consistent with codebase patterns and maintainable?" | Return to edge case exploration |
| Technical feasibility | "Has the codebase analysis confirmed viability (or flagged known risks)?" | Return to Codebase Analysis |
| Business viability | "Are the technical acceptance criteria defined and testable?" | **Escalate to PM** |

**Light escalation:** When the dev cannot answer a gate question because it requires product context, generate a "Questions for PM" artifact instead of looping the dev back.
