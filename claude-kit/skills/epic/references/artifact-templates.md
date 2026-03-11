# Epic Artifact Templates

## Epic Requirements Document

Use this structure:

1. Contexto — Narrativa que situa o leitor no dominio
2. Problema — Consequencias concretas da situacao atual
3. Objetivos — Tres sub-secoes: "O que queremos alcancar", "Como vamos medir sucesso", "O que queremos aprender"
4. Publico-alvo — Quem e afetado e suas necessidades
5. Historias de Usuario — Formato natural, sem jargao de framework
6. Escopo — Dois sub-itens: "Dentro" e "Fora"
7. Armadilhas conhecidas — (opcional) Onde o time pode se perder dentro do escopo
8. Dependencias — Componentes, servicos, sistemas
9. Questoes em aberto — Incertezas a resolver

### Rules

- The final document MUST NOT contain methodology jargon (JTBD, Four Forces, Double Diamond, Assumption Mapping, INVEST, Job Stories). These are internal conversation tools only.
- User stories use natural domain language, not the "As a [role]" or "When [situation]" formulas.
- "Objetivos" must have all three sub-sections, not just a label.
- "Escopo" must explicitly separate what is inside from what is outside.
- "Armadilhas conhecidas" is optional — include only when there are real traps. Each entry explains why it is dangerous and what to do about it.
- Codebase findings (components, constraints, reuse opportunities) go into "Dependencias".
- Assumptions with high risk and low evidence go into "Armadilhas conhecidas" or "Questoes em aberto" depending on whether they are traps or open decisions.

### What feeds each section

| Section | Source phases |
|---------|-------------|
| Contexto | reframing (raw idea + reframed problem) |
| Problema | reframing + convergence (consequences) |
| Objetivos | convergence (target outcome) + assumptions (what to learn) |
| Publico-alvo | divergence (who is affected) |
| Historias de Usuario | divergence (JTBD dimensions, rewritten in natural language) |
| Escopo | reframing (scope exclusions) + divergence (edge cases) |
| Armadilhas conhecidas | assumptions (high-risk items) + codebase-analysis (technical risks) |
| Dependencias | codebase-analysis (components, constraints, reuse) |
| Questoes em aberto | assumptions (unresolved) + exit-gate (gaps) |

## Approval Handoff Notes

When submitting for approval, make clear:
- what artifact was saved
- what decision the reviewer is making
- which phase to return to if the verdict is not approved
