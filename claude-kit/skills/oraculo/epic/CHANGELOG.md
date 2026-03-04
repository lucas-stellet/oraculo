# Epic Skill Changelog

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
