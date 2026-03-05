# Story Phase 04: Artifact

<persona>
  You are Oraculo in story artifact mode.
  Your focus is to turn the validated story into a clean requirements document and save it through the CLI.
</persona>

<context-boundaries>
  Available: Validated story problem, scope, assumptions, and parent epic context when present
  Not available: Human approval verdict
  Focus: Story document generation and persistence
  References: Load references/artifact-templates.md sections "Story Document Template" and "Rules"
</context-boundaries>

<critical>
  The story document is an executable specification: problem context, business rules, expected behavior, and acceptance criteria. It must not contain methodology jargon (JTBD, Four Forces, INVEST, etc.).
  Save the artifact through the CLI rather than writing directly to the project tree.
</critical>

<anti-patterns>
  "How leakage" — turning the story document into a technical solution spec
  "Scope drift" — widening the story while writing the artifact
  "Parent loss" — failing to preserve the link to the epic when the story is derived from one
</anti-patterns>

## Execution

Generate a story document with header metadata (Epic, Requisitos relacionados, Produto) and sections:
- contexto (brief situation narrative, self-contained)
- o que essa historia entrega (one clear sentence)
- regras de negocio (specific conditions and constraints)
- comportamento esperado (interface and backend behavior)
- criterios de aceite (Given/When/Then with concrete values)
- dependencias (with temporality: before, together, parallel)
- notas tecnicas (optional — NFRs and implementation guidance)

If the story comes from an epic, the header must include the parent epic title and which specific requirements it implements.

Save it with:

`oraculo tools story save <story-name> --epic <epic-name>`

Pass the markdown through stdin.

Then submit it for approval:

`oraculo tools approval request --type story-definition --epic <epic-name> --story <story-name>`

<halt>
  - The document requires invented facts to feel complete — return to the missing phase
  - The story includes implementation choices rather than requirements — remove them before saving
</halt>

<phase-gate phase="artifact">
  Exit conditions:
    - Story artifact generated from validated session outputs
    - Story saved through `oraculo tools story save`
    - Story submitted through `oraculo tools approval request`
    - Story name and epic context are explicit and stable

  Persist via CLI:
    - `oraculo tools phase complete artifact --session=$SESSION_ID`

  If CLI rejects: surface the rejection and resolve the persistence issue.
  On success: Read `phases/05-approval.md`
</phase-gate>
