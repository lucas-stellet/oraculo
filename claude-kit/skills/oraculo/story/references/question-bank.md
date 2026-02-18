# Question Bank

All questions organized by phase and reasoning level. Referenced by the story skill.

---

## 1. Reframing

### Deep

- "You described what you want. What problem does this solve for the user?"
- "What is explicitly out of scope for this story?"

### Light

- "You received this task. What's the problem it solves?"
- "What is explicitly out of scope?"

### Derived from epic

- "The epic describes [REC-N context]. In your words, what's the specific problem this story addresses within that requirement?"
- "What parts of the requirement are NOT covered by this story?"

### Exit condition

- **Deep:** User has articulated a problem statement independent from any specific solution. Scope boundaries are defined.
- **Light:** User can articulate the expected system behavior change independently from the implementation approach. Scope boundaries are defined.

---

## 2. Assumption Check

### Deep — Value focus

- "What are we assuming to be true about the user that, if wrong, would make this story pointless?"
- "What's the biggest risk — that we build it wrong, or that we build the wrong thing?"
- "Do you have any evidence that supports this assumption?"

### Light — Feasibility focus

- "What are we assuming about the current system that this depends on?"
- "What's the biggest technical risk here?"
- "Do you have any evidence that this approach will work?"

### Exit condition

2-3 critical assumptions identified and acknowledged by the user. Each has an evidence level assessment (high/medium/low).

---

## 3. Exit Gate — Four Risks

### Deep

| Risk | Question | If gap |
|------|----------|--------|
| Value | "Is it clear why a user would want this?" | Return to Reframing |
| Usability | "Can the user understand and use this without friction?" | Return to Reframing |
| Technical feasibility | "Is this technically feasible within the current system?" | Return to Assumption Check |
| Business viability | "Does this align with business goals?" | Return to Reframing |

### Light

| Risk | Question | If gap |
|------|----------|--------|
| Value | "Is the expected behavior clear enough to implement?" | **Escalate to PM** |
| Usability | "Is this consistent with existing patterns and maintainable?" | Return to Assumption Check |
| Technical feasibility | "Is this technically feasible within the current system?" | Return to Assumption Check |
| Business viability | "Are the acceptance criteria defined and testable?" | **Escalate to PM** |

**Light escalation:** When the dev cannot answer a gate question because it requires product context, note the gap in the story artifact under Assumptions with a flag: "Needs PM input before implementation."
