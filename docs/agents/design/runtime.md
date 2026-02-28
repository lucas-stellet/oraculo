# Agents Design — Runtime

## 1. SQLite — Operational State + Knowledge

A single SQLite database at `.oraculo/oraculo.db` serves the entire project. It is listed in `.gitignore` and is not committed to the repository. The database holds two categories of data: transient operational state (tasks, dependencies, QA verdicts) that can be cleaned after epic completion, and persistent knowledge (lessons learned, codebase patterns) that accumulates across all epics.

The database uses the schema defined in [`docs/cli/design.md`](../../cli/design.md) §4.3:

- **`epics`** — Epic metadata (name, description, timestamps)
- **`stories`** — Stories belonging to epics
- **`tasks`** — Task status lifecycle (`pending → in_progress → completed | failed`)
- **`task_dependencies`** — DAG edges (which task depends on which)
- **`task_results`** — Rich completion data (summary, logs, skills used, files modified)
- **`validations`** — QA verdicts per task and per story (approved/rejected)
- **`approvals`** — Approval gate records (type, status, artifact snapshot, verdict, comments) — see [`docs/cli/design.md`](../../cli/design.md) §4.3 for schema
- **`knowledge`** — Codebase knowledge with full-text search

### What Gets Tracked

| Data | Table | Purpose |
|------|-------|---------|
| Task status | `tasks` | The orchestrator queries this to determine what to dispatch next |
| Dependencies | `task_dependencies` | The DAG structure — which tasks block which |
| Completion data | `task_results` | Summary, logs, files modified — used for QA context and markdown generation |
| QA verdicts | `validations` | Whether a task or story passed or failed validation |
| Codebase knowledge | `knowledge` | Patterns, conventions, constraints discovered during execution |

### Lifecycle

The database is created automatically by the CLI on the first `oraculo tools` command. It persists for the life of the project.

**Transient data** (tasks, dependencies, QA verdicts, task results, approvals) is valuable during execution but not after. Once an epic completes and its markdown artifacts are generated, this operational data can be cleaned. Approval records follow the same lifecycle — they are operational state, not long-term history.

**Persistent data** (knowledge table) accumulates across all epics. When an epic/story completes, lessons learned are extracted into the knowledge table. This data is never cleaned — it is the project's long-term memory.

The database is not committed (`.gitignore`) because it contains operational state that would pollute the repository. The meaningful output — requirements, decisions, summaries, and knowledge — is captured in committed markdown files and the persistent knowledge table.

## 2. Single Markdown Artifact

When a story is completed and validated by QA, the system generates a single markdown file summarizing the implementation. This is the only committed artifact from the execution process.

The markdown contains:
- What was implemented (summary of the story)
- Key decisions made during implementation
- Files created or modified
- QA outcome

This file lives in the epic's directory structure:

```
.oraculo/
└── epics/
    └── <epic-name>/
        ├── requirements.md          # Epic requirements (output of /oraculo:epic)
        └── stories/
            └── <story-name>/
                └── requirements.md  # Story requirements (output of /oraculo:story)
```

**One file per story, not per task.** Individual task results are too granular for human reading. The story-level summary captures what matters: what was done, why, and what the QA agent verified.

## 3. Memory Model

Oraculo's memory is deliberately simple. Three sources, three purposes:

### CLAUDE.md — Project Context

CLAUDE.md is the persistent context for all agents. It contains:
- Project structure and conventions
- Architectural decisions
- Code style and patterns
- Testing strategies
- Dependencies and constraints

The orchestrator feeds CLAUDE.md content to every agent it spawns. This ensures all agents follow the same conventions without needing a complex knowledge retrieval system.

CLAUDE.md is updated by the development team (human or agent-assisted) when conventions change. It is the single source of truth for "how things are done in this project."

### Epic Markdowns — Feature History

Requirements documents, story definitions, and task summaries capture the accumulated intelligence of each feature:

- **Epic requirements** — The validated problem definition, acceptance criteria, edge cases
- **Story requirements** — Specific implementation scope and constraints
- **Completion summaries** — What was built, decisions made, QA outcomes

These files are committed to the repository. They serve as historical record and as context for future work in the same area.

### SQLite — Operational State + Knowledge

Transient data (task status, dependencies, QA verdicts) lives in SQLite for the duration of the epic — essential during execution, cleanable after. The knowledge table persists across all epics, accumulating lessons learned. This is the system's working memory plus its long-term memory.

### What This Is Not

There is no three-tier memory architecture (working/episodic/semantic). There is no curation pipeline or promotion scoring. The knowledge table with full-text search provides a simple, queryable store for codebase findings — not the deferred rich memory system described in future-work.md. The model — CLAUDE.md + markdowns + SQLite (operational state + knowledge) — covers the essential needs:

- Agents know the project's conventions (CLAUDE.md)
- Agents know the feature's requirements (epic markdowns)
- The system tracks what's done and what's pending (SQLite)

If experience reveals that a richer memory system is needed, it can be added later. See [`future-work.md`](future-work.md) for deferred memory capabilities.
