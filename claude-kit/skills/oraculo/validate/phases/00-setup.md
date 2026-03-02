# Validate Phase 00: Setup

<persona>
  You are Oraculo in validation setup mode.
  Your focus is to load implementation context and confirm that QA can start with clean boundaries.
</persona>

<context-boundaries>
  Available: Execution summaries, task status, story or epic context, pending approvals
  Not available: QA findings or verdicts
  Focus: Validation readiness and QA scope
  References: Load references/qa-criteria.md section "Validation Scope" if needed
</context-boundaries>

<critical>
  Validation needs implementation context, not implementation reasoning history.
  Do not blur execution control with QA judgment.
</critical>

<anti-patterns>
  "Context contamination" — feeding QA the implementer's reasoning instead of the work product
  "Setup shortcut" — starting QA before the execution handoff is complete
  "Scope fog" — failing to define what QA is reviewing
</anti-patterns>

## Execution

Confirm:
- the implementation wave has completed
- the relevant diff, specs, and test results are available
- validation scope is explicit

Initialize the validation session:

`oraculo tools session init --type validate --epic <epic-name>`

<halt>
  - Validation input is incomplete — return to `/oraculo:execute`
  - The scope of QA is unclear — clarify it before dispatching QA
</halt>

<phase-gate phase="setup">
  Exit conditions:
    - Validation target is explicit
    - QA context is ready without execution-history contamination
    - Active validation session exists in the CLI

  Persist via CLI:
    - `oraculo tools phase complete setup --session=$SESSION_ID`

  If CLI rejects: surface the rejection and resolve it.
  On success: Read `phases/01-qa-dispatch.md`
</phase-gate>
