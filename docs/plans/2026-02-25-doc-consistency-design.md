# Documentation Consistency Alignment

## Context

A cross-reference analysis of all Oraculo documentation (`docs/`) revealed 20 inconsistencies across philosophy, design, CLI, story, epic, and agents documents. This design captures the decisions made to resolve each one.

---

## Decisions

### 1. Directory Structure & File Paths

**Inconsistencies:** #2, #3, #6, #17

**Decision:** Epic is the root organizational entity. Hierarchical layout.

```
.oraculo/
├── oraculo.db
└── epics/
    └── <epic-name>/
        ├── requirements.md
        └── stories/
            └── <story-name>/
                └── requirements.md
```

**Changes required:**
- `epic/design.md`: replace all `.oraculo/projects/<project-name>/epic.md` with `.oraculo/epics/<epic-name>/requirements.md`
- `story/design.md`: replace all `.oraculo/projects/<project-name>/story-<slug>.md` with `.oraculo/epics/<epic-name>/stories/<story-name>/requirements.md`
- `story/design.md` and `epic/design.md`: update story-to-epic lookup paths

---

### 2. Standalone Stories via Lightweight Epic

**Inconsistency:** #5

**Decision:** CLI auto-creates a lightweight epic when a story has no epic. The user experience of standalone stories is preserved; the database always has an epic parent.

**Changes required:**
- `cli/design.md`: define "lightweight epic" concept — auto-created, name derived from story, minimal metadata
- `design.md`: document that standalone stories use a lightweight epic under the hood
- `story/philosophy.md`: keep standalone as valid UX concept, add note that CLI handles the epic creation transparently

---

### 3. SQLite: Transient Operations + Persistent Knowledge

**Inconsistency:** #8, #9

**Decision:** Single database per project (`.oraculo/oraculo.db`). Two categories of data:

- **Transient:** tasks, status, dependencies, QA verdicts — can be cleaned after epic completion
- **Persistent:** `knowledge` table — accumulates lessons learned across all epics, survives epic lifecycle

On epic/story completion:
1. Generate markdown overview (existing behavior)
2. Extract lessons learned into `knowledge` table (new)
3. Operational data for that epic may be cleaned (optional)

**Changes required:**
- `design.md`: rewrite SQLite section — not "Oraculo's memory" but "operational state + accumulated knowledge"; remove "entire journey lives there" language
- `agents/philosophy.md`: replace "ephemeral SQLite" with "SQLite with transient operational data and persistent knowledge"
- `agents/design/runtime.md`: fix "each epic gets its own database" to "single database per project"; document two data categories
- `cli/design.md`: no structural changes needed (already has knowledge table), but add documentation about knowledge persistence after epic completion

---

### 4. QA Validation at Two Levels

**Inconsistency:** #19

**Decision:** QA validates at both task level (fine-grained during execution) and story level (final integrated gate).

Flow:
1. Code agent completes task → QA validates the task (diff + specs + tests)
2. If rejected → new code agent (existing flow)
3. All tasks pass → QA performs story-level validation (integrated view, cross-task coherence, story requirements as a whole)
4. Story verdict is the final gate before markdown + knowledge generation

**Changes required:**
- `cli/design.md`: add `task_id INTEGER REFERENCES tasks(id)` (nullable) to `validations` table — when filled, it's a task validation; when null, it's a story validation
- `agents/design/qa-agent.md`: document two-level validation
- `agents/design/overview.md`: update task lifecycle to reflect both levels

---

### 5. Operating Mode Terminology

**Inconsistency:** #1

**Decision:** Standardize on "Product Engineering" / "Software Engineering" everywhere. Fix stray asterisk in CLAUDE.md.

**Changes required:**
- `design.md`: replace "Full mode (epics)" with "Product Engineering (epics)" and "Reduced mode (stories)" with "Software Engineering (stories)"
- `CLAUDE.md`: fix `Software Engineering***` → `Software Engineering`

---

### 6. Epic Frameworks Consolidation

**Inconsistency:** #7, plus framework conflict analysis

**Decision:** Reduce from 6 to 4 frameworks. Remove those that conflict with or are subtypes of others.

**Keep:**
| Framework | Role |
|-----------|------|
| Double Diamond | How the process moves (diverge → converge) |
| JTBD (absorbing Four Forces) | What to explore (job, situation, adoption forces) |
| Theory of Constraints (TOC) | Where to focus (identify the real bottleneck) |
| Assumption Mapping | What to validate (risk = impact x lack of evidence) |

**Remove:**
| Framework | Reason |
|-----------|--------|
| OST (Opportunity Solution Tree) | Conflicts with JTBD (top-down business metric vs bottom-up user job); conflicts with Double Diamond (forces classification during divergence) |
| Working Backwards (PR/FAQ) | Conflicts with Double Diamond (premature convergence); conflicts with Four Forces (optimistic narrative vs resistance mapping) |
| Four Forces (as separate item) | Subtype of JTBD (Moesta/Switch school); absorbed into JTBD |

**Changes required:**
- `CLAUDE.md`: update framework list to Double Diamond, JTBD, TOC, Assumption Mapping
- `epic/philosophy.md`: remove OST, Working Backwards, and Four Forces sections; fold Four Forces into JTBD section; add TOC section
- `epic/design.md`: remove phases/questions referencing OST, PR/FAQ stress test, and standalone Four Forces; add TOC application during convergence; update artifact templates
- `claude-kit/skills/oraculo/epic/references/frameworks.md`: align with new framework set
- `claude-kit/skills/oraculo/epic/references/artifact-templates.md`: remove OST artifact template
- `claude-kit/skills/oraculo/epic/references/question-bank.md`: update question sets

---

### 7. Markdown Generation Principle

**Inconsistency:** #10

**Decision:** Replace "Markdown Only at the End" with "each phase produces its markdown artifact; SQLite holds operational state."

**Changes required:**
- `design.md`: rewrite section 2.0 subtitle and content — markdowns are phase outputs (requirements during Discover/Plan, overview after Validate), not end-only artifacts

---

### 8. Agent Terminology

**Inconsistency:** #11

**Decision:** Replace "Teams (Agents)" with "Team of Agents."

**Changes required:**
- `design.md`: update section 3.2 header and references

---

### 9. Research Agent Definition

**Inconsistency:** #18

**Decision:** Define research agent as a real agent type with its own documentation.

**Changes required:**
- Create `agents/design/research-agent.md`: role (codebase analysis, lib research, pattern discovery), context received, skills, what it does and does not do, relationship to Discover phase
- `agents/design.md` (index): add research agent section
- `agents/design/overview.md`: include research agent in agent types

---

### 10. CLAUDE.md Agent References

**Inconsistency:** #4, #13

**Decision:** Add link to `agents/philosophy.md` and clarify that `agents/design.md` is an index.

**Changes required:**
- `CLAUDE.md`: add `Full document: [docs/agents/philosophy.md](docs/agents/philosophy.md)` and adjust design link text

---

### 11. Sub-Agents During Discover

**Inconsistency:** #15

**Decision:** Depends on prior context availability:
- **With prior context** (existing docs, defined sources): orchestrator conducts light analysis directly, no sub-agents
- **Without prior context** (raw idea, no documentation): research agents dispatched in parallel for deep codebase and external analysis

**Changes required:**
- `agents/design/overview.md`: qualify "no agents during Discover" — add condition for context-dependent research agent dispatch
- `epic/design.md`: link sub-agent dispatch to absence of prior context, not just complexity level

---

### 12. CLI Philosophy Scope

**Inconsistency:** #20

**Decision:** Fix outdated claim. CLI already covers all phases.

**Changes required:**
- `cli/philosophy.md`: replace "Currently it covers the Discover phase" with accurate statement reflecting full phase coverage

---

### 13. Epic Document Sections

**Inconsistency:** #16

**Decision:** List all 9 sections in `epic/design.md` so the design is self-contained.

The 9 sections are:
1. Summary
2. Problem
3. Context & Motivation
4. Requirements
5. User Experience
6. Scope
7. Assumptions & Risks
8. Success Criteria
9. Open Questions

**Changes required:**
- `epic/design.md`: enumerate the 9 sections where the "9-section document" is referenced

---

### 14. "No Worktrees" Statement

**Inconsistency:** #14

**Decision:** Keep as absolute statement in CLAUDE.md. No change needed.

---

## Summary of All Documents Affected

| Document | Changes |
|----------|---------|
| `CLAUDE.md` | Fix asterisk, update frameworks, add agents philosophy link, adjust agents design link text |
| `docs/design.md` | Operating mode terminology, markdown principle, SQLite role, team terminology, standalone stories |
| `docs/philosophy.md` | No changes |
| `docs/epic/philosophy.md` | Framework consolidation (remove OST, WB, standalone Four Forces; add TOC; fold Four Forces into JTBD) |
| `docs/epic/design.md` | Path convention, framework-related phases/questions, enumerate 9 sections, sub-agent dispatch context |
| `docs/story/philosophy.md` | Note on lightweight epic for standalone stories |
| `docs/story/design.md` | Path convention |
| `docs/cli/philosophy.md` | Fix scope claim |
| `docs/cli/design.md` | Lightweight epic definition, `task_id` in validations schema |
| `docs/agents/philosophy.md` | SQLite data categories, knowledge store acknowledgment |
| `docs/agents/design.md` | Add research agent to index |
| `docs/agents/design/overview.md` | Research agent type, Discover phase qualification, two-level QA, task lifecycle |
| `docs/agents/design/qa-agent.md` | Two-level validation documentation |
| `docs/agents/design/runtime.md` | Single database per project, data categories |
| `docs/agents/design/research-agent.md` | **New file** — research agent definition |
| `claude-kit/skills/oraculo/epic/references/frameworks.md` | Align with new framework set |
| `claude-kit/skills/oraculo/epic/references/artifact-templates.md` | Remove OST artifact |
| `claude-kit/skills/oraculo/epic/references/question-bank.md` | Update question sets |
