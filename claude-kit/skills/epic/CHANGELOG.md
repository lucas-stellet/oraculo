# Epic Skill Changelog

## 2026-03-04 (iteration-5)

Fixes based on iteration-4 benchmark (51/55 assertions, 92.7%). Four failures across three evals: compound question evasion via dashes, scope exclusion inflation, JTBD social dimension synthesis, and setup phase file loading inconsistency.

### Changed

- **SKILL.md — Compound questions broadened**: Rule now covers dash-separated sub-questions ("is it X, Y, or Z?"), semicolons, and "or"-chained alternatives — not just "and"-joined. Added new BAD examples for each pattern and a GOOD example showing how to put alternatives in the `options` array instead of the question stem.
- **SKILL.md — Bootstrap loading reinforced**: Wrapped `00-setup.md` loading instruction in `<critical>` tag. Added "visibly loaded" and "implicit loading is not sufficient" language to ensure the loading appears in transcript actions.
- **Phase 01 — Scope exclusion self-check**: Added `<critical>` block inside phase-gate (before persist step) requiring every exclusion to map to a specific user statement. Strengthened existing checklist item: "not said or agreed = not confirmed".
- **Phase 02 — JTBD dimension checklist**: Added `<critical>` block requiring a dedicated question for each of functional, emotional, and social dimensions before completing divergence. Updated phase-gate exit condition to explicitly prohibit synthesizing missing dimensions from other answers.

---

## 2026-03-04 (iteration-4)

Fixes based on iteration-3 benchmark (45/47 assertions, 95.7%). Two failures: compound question slipping through despite the rule, and 00-setup.md loaded one turn late.

### Changed

- **SKILL.md — Compound questions**: Added explicit violation/correction examples to the atomic question rule showing the exact "and"-joined pattern and how to split it.
- **SKILL.md — Bootstrap timing**: Step 6 now explicitly says to read `phases/00-setup.md` immediately in the same turn as bootstrap checks, not a subsequent turn.

### Evals

- **reads-setup-phase-first**: Relaxed from "loaded in Turn 1" to "loaded before any setup-phase question is asked" to account for bootstrap overhead.
- **New assertions**: `phase-file-loading-protocol` (all evals), `scope-exclusions-in-phase-gate` (vague-feature, solution-first), `jtbd-depth-not-just-coverage` (solution-first), `explores-jtbd-dimensions` (solution-first), `session-init-solution-neutral-name` (solution-first).

---

## 2026-03-04 (iteration-3)

Improvements based on iteration-2 benchmark (35/35 assertions, 100%) and grader-suggested next steps.
Four new behavioral areas: compound questions, passive confirmation options, phase gate data quality, JTBD depth, and user pushback handling.

### Changed

- **SKILL.md — Atomic questions**: Strengthened one-question-per-turn rule to explicitly prohibit compound questions joined by "and" or packed into one block.
- **SKILL.md — No passive confirmation options**: New rule prohibiting options like "Yes, that captures it well" or "Agreed, let's continue". Every option must propose a direction to investigate. Confirmation should be framed as competing interpretations.
- **SKILL.md — Phase gate payload quality**: New rule requiring substantive, conversation-specific content in phase gate JSON payloads — no generic summaries or placeholder text.
- **Phase 01 — User pushback handling**: New section guiding how to handle resistance to reframing without caving in or becoming combative. Record unexamined areas as risks after two pushback attempts.
- **Phase 02 — Depth over coverage**: New section requiring genuine insight extraction from JTBD/Four Forces, not just checkbox coverage. Push deeper on generic answers.
- **Phase 02 — User pushback handling**: New section for handling resistance to exploration. Acknowledge, explain value, offer concrete examples, record gaps as risks.
- **Phase 02 — Phase gate specificity**: Exit conditions now require specific concrete insights (not generic statements) and specific behaviors/costs/outcomes for each force.

---

## 2026-03-04 (iteration-2)

Fixes based on skill benchmark iteration-1 results (24/25 assertions, 95.8%).
One assertion failure (plain-text question leak) and two grader-flagged concerns (technology anchoring, leading options).

### Changed

- **SKILL.md — AskUserQuestion coverage**: Expanded rule to explicitly list transition prompts ("Ready to continue?"), confirmation prompts, phase-gate prompts, and any sentence ending with "?" as requiring `AskUserQuestion`. Plain-text questions are now called out as a rule violation.
- **SKILL.md — Non-leading options**: New rule requiring question options to be directional and exploratory, never funneling toward a predetermined answer.
- **Phase 01 — Technology anchoring**: New anti-pattern prohibiting naming specific technologies, architectures, or products (e.g., "Postgres read replicas", "CQRS", "Redis") during problem exploration, even when framed as evidence.
- **Phase 01 — Solution-term self-check**: Execution checklist now verifies the reframed problem avoids solution-specific terms ("dashboard", "API", "screen", "migration").
- **Phase 01 — Scope exclusion confirmation**: Execution checklist now verifies scope exclusions were confirmed with the user, not decided unilaterally.
- **Phase 02 — Technology anchoring**: Same anti-pattern added to divergence phase.
- **Phase 02 — Force question framing**: Four Forces questions must be framed in terms of user behaviors, costs, and outcomes — never referencing specific technologies by name.

---

## 2026-03-04 (iteration-1)

Improvements based on skill benchmark (iteration-1, 14/14 assertions passing).
Grader feedback identified 4 behavioral gaps not yet covered by assertions.

### Changed

- **Phase 00 — Epic naming enforcement**: Added `<critical>` block requiring exactly 3-4 name candidates presented via `AskUserQuestion`. No auto-picking or single-option presentations.
- **Phase 00 — Triage problem-solution separation**: Halt block now requires explicitly separating **Observation** (symptom) from **Assumed fix** (proposed solution) before redirecting to `/oraculo:story`.
- **Phase 01 — Problem-first input depth**: New anti-pattern "Deference to articulate input" and 4 targeted probes for well-framed inputs (challenge framing, root cause vs symptom, hidden assumptions, boundaries).
- **SKILL.md — One question per turn**: Added hard rule — each turn MUST contain exactly one `AskUserQuestion` call, no batching across all phases.
