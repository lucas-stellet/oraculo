# Oraculo CLI — Design

## 1. Command Architecture

The CLI has two command layers, separated by audience and purpose:

| Layer | Audience | Output | Purpose |
|-------|----------|--------|---------|
| **Root** (`oraculo <cmd>`) | Humans | Formatted tables/text | Lifecycle, dashboard |
| **Tools** (`oraculo tools <domain> <action>`) | Agents | JSON | Deterministic operations |

Both layers share internal Go packages but have independent command handlers and output formatting. Root commands prioritize readability. Tools commands prioritize parseability.

## 2. Root Commands (Human-facing)

```
oraculo install [--global|--local]    # Install skills, hooks, settings into Claude Code
oraculo update                        # Update the CLI binary to latest version
oraculo status                        # Dashboard: epics, stories, tasks, progress
oraculo version                       # CLI version
```

**`install`** — Configures Claude Code to work as Oraculo. Two modes:

- `--global`: copies skills to `~/.claude/skills/`, hooks to `~/.claude/hooks/`, modifies global settings. All projects gain access to Oraculo.
- `--local`: same but into `.claude/` of the current project. Only that project gets Oraculo.

**`update`** — Updates the Oraculo CLI binary itself to the latest version. Distinct from `install` — `install` manages Claude Code configuration (skills, hooks, settings), `update` manages the binary.

**`status`** — Human-readable dashboard showing epics, stories count, task progress (pending/in_progress/completed/failed), percentages.

**No `init` command.** Self-bootstrapping: the first `tools` command that touches storage creates `.oraculo/` and the database automatically.

## 3. Tools Commands (Agent-facing)

Convention: entity name is always the first positional argument. Parent references use flags (`--epic`, `--story`).

All `init` commands are idempotent — calling them twice returns `"created": false` instead of failing. Status transition commands (`start`, `complete`, `fail`) enforce the lifecycle and reject invalid transitions with a structured error.

### 3.1 Epic

```bash
oraculo tools epic init <name> [--description "..."]
oraculo tools epic save <name>                          # stdin: markdown
oraculo tools epic get <name>                           # stdout: raw markdown
oraculo tools epic list
oraculo tools epic update <name> [--description "..."]
oraculo tools epic delete <name>
```

### 3.2 Story

```bash
oraculo tools story init <name> --epic <epic>
oraculo tools story save <name> --epic <epic>            # stdin: markdown
oraculo tools story get <name> --epic <epic>             # stdout: raw markdown
oraculo tools story list --epic <epic>
oraculo tools story update <name> --epic <epic> [--description "..."]
oraculo tools story delete <name> --epic <epic>
```

### 3.3 Task

```bash
oraculo tools task init <name> --epic <epic> --story <story> [--description "..."] [--depends-on <task>]
oraculo tools task start <name> --epic <epic> --story <story>
oraculo tools task complete <name> --epic <epic> --story <story>    # stdin: JSON
oraculo tools task fail <name> --epic <epic> --story <story> --reason "..."
oraculo tools task get <name> --epic <epic> --story <story>
oraculo tools task list --epic <epic> --story <story>
oraculo tools task delete <name> --epic <epic> --story <story>
```

**Task lifecycle:** `pending → in_progress → completed | failed`

**`task init`** — Task IDs come from the story document. When the user approves a story, the agent creates tasks based on the IDs defined there. `--depends-on` builds the DAG.

**`task complete` stdin format:**

```json
{
  "summary": "Implemented login form with validation",
  "logs": "Created src/components/LoginForm.tsx...",
  "skills_used": ["test-driven-development", "frontend-design"],
  "files_modified": ["src/components/LoginForm.tsx", "src/tests/login.test.tsx"]
}
```

**`task fail`** — `--reason` is required. The contract refuses a failure without justification.

### 3.4 Memory

```bash
oraculo tools memory store --domain <d> --category <c> --finding "..." [--source "..."] [--confidence high|medium|low]
oraculo tools memory search <query> [--domain <d>] [--limit N]
oraculo tools memory domains
```

Categories: `pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`.

## 4. Data Model

### 4.1 Responsibility Split

| Data type | Where | Why |
|-----------|-------|-----|
| Structural (DAG, status, verdicts) | SQLite | Deterministic, queryable |
| Content (analysis, decisions, reasoning) | Markdown | Creative, narrative |
| Bridge | `save` + `memory` | CLI indexes, doesn't replace |

### 4.2 File System

```
.oraculo/
├── oraculo.db                              # SQLite — infrastructure, .gitignore
└── epics/
    └── gastos-control/
        ├── requirements.md                 # Epic requirements (output of /oraculo:epic)
        └── stories/
            └── login-flow/
                ├── requirements.md         # Story requirements (output of /oraculo:story)
                └── tasks/                  # Task results stored in DB, not files
```

Rules:

- Stories always belong to an epic. CLI enforces this. When a story is created without an explicit epic, the CLI auto-creates a **lightweight epic** — a minimal epic with the name derived from the story and no requirements markdown. This preserves the hierarchical data model while keeping the standalone story UX seamless.
- Each level's `requirements.md` is the product definition — WHAT and WHY, never HOW.
- The SQLite database is infrastructure — `.gitignore`.
- Markdown files are versionable in git.

### 4.3 SQLite Schema

```sql
-- Core entities
CREATE TABLE epics (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    created_at  TEXT DEFAULT (datetime('now')),
    updated_at  TEXT DEFAULT (datetime('now'))
);

CREATE TABLE stories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    epic_id     INTEGER NOT NULL REFERENCES epics(id),
    name        TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at  TEXT DEFAULT (datetime('now')),
    updated_at  TEXT DEFAULT (datetime('now')),
    UNIQUE(epic_id, name)
);

CREATE TABLE tasks (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id       INTEGER NOT NULL REFERENCES stories(id),
    name           TEXT NOT NULL,
    description    TEXT DEFAULT '',
    status         TEXT DEFAULT 'pending'
                   CHECK (status IN ('pending','in_progress','completed','failed')),
    failure_reason TEXT DEFAULT '',
    created_at     TEXT DEFAULT (datetime('now')),
    updated_at     TEXT DEFAULT (datetime('now')),
    started_at     TEXT,
    completed_at   TEXT,
    UNIQUE(story_id, name)
);

-- DAG: task dependencies
CREATE TABLE task_dependencies (
    task_id    INTEGER NOT NULL REFERENCES tasks(id),
    depends_on INTEGER NOT NULL REFERENCES tasks(id),
    PRIMARY KEY (task_id, depends_on),
    CHECK (task_id != depends_on)
);

-- Rich completion data (1:1 with task, only on completion)
CREATE TABLE task_results (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id        INTEGER UNIQUE NOT NULL REFERENCES tasks(id),
    summary        TEXT NOT NULL,
    logs           TEXT DEFAULT '',
    skills_used    TEXT DEFAULT '',     -- JSON array stored as text
    files_modified TEXT DEFAULT '',     -- JSON array stored as text
    created_at     TEXT DEFAULT (datetime('now'))
);

-- QA validation verdicts (structural part; analysis lives in markdown)
CREATE TABLE validations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id    INTEGER NOT NULL REFERENCES stories(id),
    task_id     INTEGER REFERENCES tasks(id),              -- NULL = story-level validation
    verdict     TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
    created_at  TEXT DEFAULT (datetime('now'))
);

-- Codebase knowledge
CREATE TABLE knowledge (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    domain       TEXT NOT NULL,
    category     TEXT NOT NULL
                 CHECK (category IN ('pattern','convention','constraint','dependency','test','architecture')),
    finding      TEXT NOT NULL,
    source_files TEXT DEFAULT '',
    confidence   TEXT DEFAULT 'medium'
                 CHECK (confidence IN ('high','medium','low')),
    created_at   TEXT DEFAULT (datetime('now'))
);

-- Full-text search index
CREATE VIRTUAL TABLE knowledge_fts USING fts5(
    domain, category, finding, source_files,
    content=knowledge, content_rowid=id
);

CREATE TRIGGER knowledge_ins AFTER INSERT ON knowledge BEGIN
    INSERT INTO knowledge_fts(rowid, domain, category, finding, source_files)
    VALUES (new.id, new.domain, new.category, new.finding, new.source_files);
END;

CREATE TRIGGER knowledge_del AFTER DELETE ON knowledge BEGIN
    INSERT INTO knowledge_fts(knowledge_fts, rowid, domain, category, finding, source_files)
    VALUES ('delete', old.id, old.domain, old.category, old.finding, old.source_files);
END;
```

Design decisions:

- `epics` and `stories` track metadata only. Content lives in Markdown files.
- `tasks` track status lifecycle and metadata. Rich completion data in `task_results`.
- `task_dependencies` models the DAG. CLI can validate no cycles exist.
- `validations` stores structural verdicts at two levels: per-task (`task_id` filled) during execution and per-story (`task_id` NULL) as the final gate. QA analysis lives in Markdown.
- `knowledge` uses FTS5 for full-text search. Sync triggers keep index consistent.
- JSON arrays (`skills_used`, `files_modified`) stored as TEXT — simple, queryable with `json_each()`.

**Knowledge persistence:** Unlike operational tables (`tasks`, `task_dependencies`, `task_results`, `validations`), the `knowledge` table is persistent — it accumulates lessons learned across all epics and survives epic completion. Operational data can be cleaned after an epic completes; knowledge data is retained.

## 5. Output Format

**Tools commands:** JSON on stdout, exit code 0 for success, exit code 1 for error.

Success example:

```json
{
  "name": "gastos-control",
  "path": ".oraculo/epics/gastos-control",
  "created": true
}
```

Error example:

```json
{
  "error": "epic_not_found",
  "message": "No epic found with name 'nonexistent'"
}
```

**Content retrieval** (`get` commands): raw Markdown on stdout, exit code 0. Errors still return JSON with exit code 1.

**Root commands:** human-readable formatted output (tables, progress bars, aligned text). Not meant for programmatic parsing.

## 6. Go Project Structure

```
cmd/oraculo/
├── main.go
├── cmd/
│   ├── root.go                 # Cobra root command
│   ├── install.go              # install --global|--local
│   ├── status.go               # Human dashboard
│   ├── version.go              # Version info
│   └── tools/
│       ├── tools.go            # tools parent command + auto-bootstrap middleware
│       ├── epic.go             # tools epic init|save|get|list|update|delete
│       ├── story.go            # tools story init|save|get|list|update|delete
│       ├── task.go             # tools task init|start|complete|fail|get|list|delete
│       └── memory.go           # tools memory store|search|domains
├── internal/
│   ├── db/
│   │   ├── db.go               # Open, auto-create .oraculo/, migrate
│   │   └── migrations.go       # Schema versioning
│   ├── epic/
│   │   └── epic.go             # Epic business logic
│   ├── story/
│   │   └── story.go            # Story business logic
│   ├── task/
│   │   └── task.go             # Task business logic + DAG validation
│   ├── memory/
│   │   └── memory.go           # Knowledge store/search
│   ├── installer/
│   │   └── installer.go        # Install logic (copy skills, hooks, settings)
│   └── output/
│       ├── json.go             # JSON response helpers (tools)
│       └── table.go            # Human-readable formatting (root)
├── go.mod
└── go.sum
```

**Dependencies:**

- `github.com/spf13/cobra` — CLI framework
- `modernc.org/sqlite` — Pure-Go SQLite (no CGO)

**Architecture notes:**

- `cmd/tools/tools.go` contains a Cobra `PersistentPreRun` that handles auto-bootstrapping (create `.oraculo/` and DB if missing) for all tools subcommands.
- `internal/` packages contain pure business logic, no CLI concerns. Both `tools` commands and `status` command use the same internal packages.
- Validation of preconditions (epic exists, story belongs to epic, valid status transition) lives in `internal/` packages, not in command handlers.
- Command handlers are thin: parse flags → call internal logic → format output.
