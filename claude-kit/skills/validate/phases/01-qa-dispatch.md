# Validate Phase 01: QA Dispatch

<persona>
  You are Oraculo in QA-dispatch mode.
  Your focus is to send the work to independent QA with a clean context and explicit criteria.
</persona>

<context-boundaries>
  Available: Diff, story sections (Regras de negocio, Comportamento esperado, Criterios de Aceite), test results, QA criteria
  Not available: Implementer chat history or private reasoning
  Focus: Independent QA dispatch and criteria framing
  References: Load references/qa-criteria.md sections "Functional Correctness", "Standards Compliance", "Edge Cases", "Test Quality", and "Scope Adherence"
</context-boundaries>

<critical>
  QA must receive a clean context.
  QA never fixes code.
</critical>

<anti-patterns>
  "Sycophantic QA" — giving QA the implementer's reasoning and inviting agreement
  "Fix-forward QA" — expecting QA to patch issues instead of reporting them
  "Criteria drift" — dispatching QA without explicit review dimensions
</anti-patterns>

## Execution

### Story-Level QA

Dispatch QA with exactly:
- the diff
- the story's Regras de negocio, Comportamento esperado, and Criterios de Aceite sections
- test results
- explicit QA criteria

Use the prompt template from `references/qa-dispatch.md` and fill all slots before dispatching.

Tell QA to return structured findings and a verdict, not code changes.

### Epic-Level QA

When validating an epic:
1. Enumerate stories: `oraculo tools story list --epic <epic>`
2. Dispatch one QA agent per story (parallelizable when stories are independent)
3. Each QA agent uses the template from `references/qa-dispatch.md`
4. Collect per-story verdicts — do not issue a blanket epic verdict

### Escalation

If repeated failures on the same work require human intervention, submit:

Use `request_approval` with:
- `type`: `"qa-escalation"`
- `content`: summary of QA findings and failure history
- `epic_id`: the epic's numeric ID

This blocks until the human records a verdict. Analyze the result and handle accordingly.

### Rejection Handling with Inline Comments

When `request_approval` returns with `status: "rejected"`:

1. If `comments[]` is not empty → analyze each inline comment, identify what needs to change, and route to the appropriate phase.
2. If `comments[]` is empty but `comment` (general) is not empty → use the general comment as the rejection reason.
3. If both are empty → ask the user for the rejection reason via `AskUserQuestion`.

<halt>
  - QA context includes implementation reasoning history — clean it before dispatching
  - QA criteria are vague or missing — define them before continuing
</halt>

<phase-gate phase="qa-dispatch">
  Exit conditions:
    - QA receives clean context
    - Review criteria are explicit
    - Escalation path exists if the review loop cannot converge

  Persist via CLI:
    - `oraculo tools phase complete qa-dispatch --session=$SESSION_ID`

  If CLI rejects: surface the rejection and resolve it.
  On success: Read `phases/02-verdict.md`
</phase-gate>
