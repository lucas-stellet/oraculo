> **Internal reference only.** These questions guide the Socratic conversation. Methodology terminology MUST NOT appear in the generated artifact.

# Story Question Bank

## Setup and Triage

- "What work item are we actually trying to define?"
- "Is this standalone work or part of an approved epic?"
- "What makes this small enough to be one story?"
- "If this turns out broader, what would force us to escalate to `/epic`?"

## Reframing

- "What problem does this story solve?"
- "What would happen if this story did not exist?"
- "What is explicitly out of scope?"
- "What does success look like for this slice?"

## Epic-Derived Prompts

- "Which epic requirement does this story implement?"
- "What part of that requirement does this story cover?"
- "What does the epic already tell us that we should not rediscover here?"
- "What remains for future stories?"

## Assumptions

- "What are we assuming about the current system?"
- "What are we assuming about the user or operator?"
- "What evidence supports that assumption?"
- "If this assumption is wrong, how bad is the failure?"

## Exit Gate

- "Why does this story matter?"
- "What would make this story hard to use or validate?"
- "What is the biggest technical risk?"
- "What product or business context is still missing, if any?"

## Business Rules

- "What rules or constraints govern this behavior?"
- "What conditions must always hold true?"
- "Are there different behaviors depending on user type, state, or configuration?"
- "What happens when the input is invalid or unexpected?"

## Expected Behavior

- "What does the user see when this works correctly?"
- "What are the key interaction steps from the user's perspective?"
- "What does the system do in the backend when this is triggered?"
- "What feedback does the user receive (success, error, warning)?"

## Acceptance Criteria

- "How will we know this story is done?"
- "What are the pass/fail conditions?"
- "What edge cases must be explicitly tested?"
- "What should NOT change as a result of this work?"

## Dependencies

- "What other stories, systems, or APIs does this depend on?"
- "What must be ready before this can start?"
- "What can be worked on in parallel?"

## Question Quality Anti-Patterns

### Compound Questions (never do this)

- BAD: "What are the business rules? What is a valid value? What is the character limit?" → Three questions packed into one turn. Ask each in a separate turn.
- BAD: "Does the system support X? Or could Y happen?" → Two inquiries disguised as one. Split into separate turns.
- GOOD: "What business rules govern this behavior?" → Single atomic question. Follow up based on the answer.

### Passive Confirmation Options (never do this)

- BAD: "Does this capture what you had in mind?" with options ["Yes, looks good", "No, needs changes"] → Passive yes/no that closes inquiry.
- BAD: "Do you recognize these as the main risks?" with options ["Yes, those are the risks", "No"] → Agreement/disagreement, not exploration.
- GOOD: "Which of these risks concerns you most?" with options ["Data cascade scope", "Session invalidation", "Reversibility", "Other"] → Each option opens a different direction.
- GOOD: "How would you adjust this problem statement?" with options ["Narrow the scope to X", "Add constraint Y", "Reframe around Z", "Other"] → Competing interpretations, not confirmation.
