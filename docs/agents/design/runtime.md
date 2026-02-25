# Agents Design — Runtime

## 1. Ephemeral SQLite

Each epic gets its own SQLite database at `.oraculo/oraculo.db`. This database is ephemeral — it is listed in `.gitignore` and is not committed to the repository. It serves as infrastructure for tracking execution state, not as a knowledge store.

The database uses the schema defined in [`docs/cli/design.md`](../../cli/design.md) §4.3:

- **`epics`** — Epic metadata (name, description, timestamps)
- **`stories`** — Stories belonging to epics
- **`tasks`** — Task status lifecycle (`pending → in_progress → completed | failed`)
- **`task_dependencies`** — DAG edges (which task depends on which)
- **`task_results`** — Rich completion data (summary, logs, skills used, files modified)
- **`validations`** — QA verdicts per story (approved/rejected)
- **`knowledge`** — Codebase knowledge with full-text search

### What Gets Tracked

| Data | Table | Purpose |
|------|-------|---------|
| Task status | `tasks` | The orchestrator queries this to determine what to dispatch next |
| Dependencies | `task_dependencies` | The DAG structure — which tasks block which |
| Completion data | `task_results` | Summary, logs, files modified — used for QA context and markdown generation |
| QA verdicts | `validations` | Whether a story passed or failed validation |
| Codebase knowledge | `knowledge` | Patterns, conventions, constraints discovered during execution |

### Lifecycle

The database is created automatically by the CLI on the first `oraculo tools` command. It lives for the duration of the epic's active development. Once the epic is complete and the final markdown artifact is generated, the database has served its purpose — it can be deleted without losing any committed knowledge.

**Why ephemeral?** The database contains operational state (task progress, agent logs, intermediate results) that is valuable during execution but not after. Committing it would pollute the repository with transient data. The meaningful output — requirements, decisions, and summaries — is captured in committed markdown files.

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

### Ephemeral SQLite — Execution State

Task status, dependencies, QA verdicts, and operational data live in SQLite for the duration of the epic. This is the working memory of the system — essential during execution, disposable after.

### What This Is Not

There is no three-tier memory architecture (working/episodic/semantic). There is no curation pipeline. There is no semantic knowledge store with scoring and promotion. These are powerful concepts but add complexity that is not justified at this stage. The simple model — CLAUDE.md + markdowns + ephemeral SQLite — covers the essential needs:

- Agents know the project's conventions (CLAUDE.md)
- Agents know the feature's requirements (epic markdowns)
- The system tracks what's done and what's pending (SQLite)

If experience reveals that a richer memory system is needed, it can be added later. See [`future-work.md`](future-work.md) for deferred memory capabilities.
