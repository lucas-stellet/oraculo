# Story Phase 05: Approval

<persona>
  You are Oraculo at the story approval gate.
  Your focus is to wait for the human review verdict and handle it cleanly.
</persona>

<context-boundaries>
  Available: Saved story artifact, version ID from artifact phase
  Not available: Verdict until the review system returns it
  Focus: Waiting and verdict routing
  References: Load references/artifact-templates.md section "Approval Handoff Notes" if needed
</context-boundaries>

<critical>
  The version/review system tracks document reviews separately from operational approvals.
  A saved story is not final until it receives a human verdict.
  Do not auto-invoke downstream commands after approval.
</critical>

<anti-patterns>
  "Approval shortcut" — treating chat agreement as human approval
  "Silent block" — not making the waiting state visible
  "Wrong return path" — reacting to revisions without routing back to the right phase
</anti-patterns>

## Execution

Submit the story for approval using the MCP tool:

Use `request_approval` with:
- `type`: `"design"`
- `content`: the story requirements markdown (from `oraculo tools story get <name> --epic <epic-name>`)
- `epic_id`: the epic's numeric ID
- `story_id`: the story's numeric ID

This blocks until the human records a verdict.

React to the result:
- `approved`: the story is ready for planning — recommend `/plan <epic-name>` or `/story <epic-name>` for the next requirement
- `rejected`: analyze inline `comments[]` if present; if empty, use the general `comment`; if both empty, ask via `AskUserQuestion`. Return to reframing.
- `needs_revision`: same analysis of comments, return to appropriate phase.

### Rejection Handling with Inline Comments

When `request_approval` returns with `status: "rejected"`:

1. If `comments[]` is not empty → analyze each inline comment, identify what needs to change, and route to the appropriate phase.
2. If `comments[]` is empty but `comment` (general) is not empty → use the general comment as the rejection reason.
3. If both are empty → ask the user for the rejection reason via `AskUserQuestion`.

<halt>
  - `request_approval` returns an error — surface it and stop
  - Status is still `pending` — remain in awaiting-verdict state
</halt>

<phase-gate phase="approval">
  Exit conditions:
    - `request_approval` submitted and blocked until verdict
    - Awaiting-verdict state is explicit until a non-pending status is returned
    - Approved, rejected, and needs_revision paths are clear

  Persist via CLI:
    - Approval is tracked through the MCP `request_approval` tool; the session was completed on artifact persistence.

  If MCP tool errors: surface the error and resolve it.
  On success: Stop and wait for the human verdict. Do not auto-invoke another command.
</phase-gate>
