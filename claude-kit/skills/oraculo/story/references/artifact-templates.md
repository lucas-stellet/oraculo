# Story Artifact Templates

## Story Document Template

### Header metadata

Every story document starts with:
- **Epic:** [parent epic title]
- **Requisitos relacionados:** [which user stories or requirements from the epic this story implements]
- **Produto:** [product name]
- **Ultima atualizacao:** [date]

For standalone stories (no parent epic), omit the "Requisitos relacionados" field.

### Sections

1. Contexto — Brief, self-contained situation narrative
2. O que essa historia entrega — One clear sentence
3. Regras de negocio — Specific conditions and constraints
4. Comportamento esperado — Interface and backend behavior
5. Criterios de Aceite — Given/When/Then format
6. Dependencias — With temporality: "before", "together with", "in parallel"
7. Notas tecnicas — (optional) NFRs + implementation guidance

### Rules

- The final document MUST NOT contain methodology jargon (JTBD, Four Forces, Double Diamond, Assumption Mapping, INVEST). These are internal conversation tools only.
- The story must be self-contained — a developer reads only the story and knows what to build.
- "Criterios de Aceite" use Given/When/Then format with concrete values, not abstract descriptions.
- "Notas tecnicas" is optional — include only when there are real NFRs (performance, accessibility, security) or implementation constraints the developer must know.
- Section "Escopo" (inside/outside) is optional — include only when there is genuine ambiguity about what this story covers vs. what it does not.

### What feeds each section

| Section | Source phases |
|---------|-------------|
| Contexto | reframing (problem statement + parent epic context) |
| O que essa historia entrega | reframing (desired outcome) |
| Regras de negocio | reframing + assumptions (business constraints) |
| Comportamento esperado | reframing + exit-gate (UI/backend behavior) |
| Criterios de Aceite | exit-gate (four risks translated to testable criteria) |
| Dependencias | assumptions (system dependencies, other stories) |
| Notas tecnicas | assumptions (technical constraints, NFRs) |

## Approval Handoff Notes

When submitting for approval, make clear:
- which story was saved
- what slice of work it covers
- which phase to return to if revisions are requested
