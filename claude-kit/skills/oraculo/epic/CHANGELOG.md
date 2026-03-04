# Epic Skill Changelog

## 2026-03-04

Improvements based on skill benchmark (iteration-1, 14/14 assertions passing).
Grader feedback identified 4 behavioral gaps not yet covered by assertions.

### Changed

- **Phase 00 — Epic naming enforcement**: Added `<critical>` block requiring exactly 3-4 name candidates presented via `AskUserQuestion`. No auto-picking or single-option presentations.
- **Phase 00 — Triage problem-solution separation**: Halt block now requires explicitly separating **Observation** (symptom) from **Assumed fix** (proposed solution) before redirecting to `/oraculo:story`.
- **Phase 01 — Problem-first input depth**: New anti-pattern "Deference to articulate input" and 4 targeted probes for well-framed inputs (challenge framing, root cause vs symptom, hidden assumptions, boundaries).
- **SKILL.md — One question per turn**: Added hard rule — each turn MUST contain exactly one `AskUserQuestion` call, no batching across all phases.
