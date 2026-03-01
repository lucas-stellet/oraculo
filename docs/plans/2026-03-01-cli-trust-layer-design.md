# CLI Trust Layer — Skeleton Design

## Summary

Build the complete skeleton of the Oraculo CLI Trust Layer: a single Go binary with all 28 documented commands, SQLite persistence with the full schema, and TDD from the start. This is the foundational infrastructure that skills, agents, and the UI depend on.

## Approach

**Bottom-Up:** `domain/` → `db/` → `cli/` → `main.go`. Each layer is testable in isolation before the next is built on top.

## Project Layout

```
oraculo/
├── go.mod                          # module github.com/lucas/oraculo
├── go.sum
├── cmd/
│   └── oraculo/
│       └── main.go                 # Entrypoint: wiring + cobra.Execute()
├── src/
│   ├── domain/                     # Pure types, validation, zero external deps
│   │   ├── epic.go
│   │   ├── story.go
│   │   ├── task.go
│   │   ├── memory.go
│   │   ├── approval.go
│   │   └── errors.go
│   ├── db/                         # SQLite: connection, migrations, stores
│   │   ├── db.go                   # Open(), auto-bootstrap .oraculo/
│   │   ├── migrations.go           # Schema versioning via PRAGMA user_version
│   │   ├── epic_store.go
│   │   ├── story_store.go
│   │   ├── task_store.go
│   │   ├── memory_store.go
│   │   └── approval_store.go
│   ├── cli/                        # Cobra commands, thin handlers
│   │   ├── root.go                 # Root command + subcommand registration
│   │   ├── install.go
│   │   ├── status.go
│   │   ├── version.go
│   │   └── tools/
│   │       ├── tools.go            # "tools" parent + PersistentPreRunE bootstrap
│   │       ├── epic.go
│   │       ├── story.go
│   │       ├── task.go
│   │       ├── memory.go
│   │       └── approval.go
│   ├── output/                     # Output formatting
│   │   ├── json.go                 # JSON helpers for tools commands
│   │   └── table.go                # Human-readable for root commands
│   └── installer/                  # Install logic (stub in this phase)
│       └── installer.go
├── docs/
├── claude-kit/
└── ...
```

Key decision: `src/` instead of `internal/`. The repo will grow to include `src/web/` (HTTP + WebSocket), `src/mcp/` (MCP server), and potentially frontend assets. Using `src/` keeps all source code organized under one scalable directory.

## Domain Layer (`src/domain/`)

Pure types with zero external imports. Validation logic lives on the types themselves.

### Types

Each entity is a struct with typed constants for state fields:

- `Epic` — ID, Name, Description, ApprovalStatus, timestamps
- `Story` — ID, EpicID, Name, Description, ApprovalStatus, timestamps
- `Task` — ID, StoryID, Name, Description, Status, FailureReason, timestamps
- `TaskResult` — Summary, Logs, SkillsUsed ([]string), FilesModified ([]string)
- `Knowledge` — ID, Domain, Category, Finding, SourceFiles, Confidence, timestamp
- `Approval` — ID (UUID string), Type, EpicID, StoryID, Content, PreviousVersion, Status, VerdictComment, timestamps

### State validation

```go
var validTransitions = map[TaskStatus][]TaskStatus{
    TaskPending:    {TaskInProgress},
    TaskInProgress: {TaskCompleted, TaskFailed},
}

func (s TaskStatus) CanTransitionTo(target TaskStatus) bool { ... }
```

### Errors

Sentinel errors with wrapping — no custom error type:

```go
var (
    ErrNotFound          = errors.New("not found")
    ErrAlreadyExists     = errors.New("already exists")
    ErrInvalidTransition = errors.New("invalid status transition")
    ErrCyclicDependency  = errors.New("cyclic dependency in task graph")
    ErrMissingRequired   = errors.New("missing required field")
)
```

The `db/` layer wraps with context: `fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)`. The `output/` layer maps sentinel errors to the documented JSON error format using `errors.Is()`.

### No interfaces in domain

Following idiomatic Go: interfaces are defined by consumers (`cli/` handlers), not producers. Each CLI handler defines a minimal interface (1-3 methods) with only what it needs. The `db/` layer exports concrete store types.

## Data Layer (`src/db/`)

### Connection and bootstrap

`db.Open()` is the single entry point. It:
1. Creates `.oraculo/` directory if missing
2. Creates or opens `.oraculo/oraculo.db`
3. Runs all pending migrations
4. Returns a `*DB` ready for use

No separate init step. The first tools command that touches storage creates everything.

### Migrations

Versioned using `PRAGMA user_version`. Each migration is a function that runs inside a transaction:

```go
var migrations = []func(*sql.Tx) error{
    v1CreateCoreTables,    // epics, stories, tasks, task_dependencies, task_results
    v1CreateKnowledge,     // knowledge + FTS5 + triggers
    v1CreateApprovals,     // approvals
    v1CreateValidations,   // validations
}
```

### Schema

The full SQLite schema from `docs/cli/design.md` §4.3, including:
- Core tables: `epics`, `stories`, `tasks`, `task_dependencies`, `task_results`
- Knowledge with FTS5: `knowledge`, `knowledge_fts`, sync triggers
- Approvals: `approvals` with UUID primary key
- Validations: `validations` with optional task-level granularity

### Stores

One concrete store per entity domain. No interfaces — consumers define their own at point of use.

- `EpicStore` — CRUD + approval status
- `StoryStore` — CRUD scoped to epic + approval status
- `TaskStore` — CRUD + lifecycle transitions + DAG validation (cycle detection)
- `MemoryStore` — store/search via FTS5 + domain listing
- `ApprovalStore` — request/verdict/query

Idempotency: all `Create` methods return `(entity, created bool, err)`. Calling Create twice returns the existing entity with `created=false`.

## CLI Layer (`src/cli/`)

### Architecture

- `NewRoot()` assembles the full command tree
- `tools/tools.go` has a `PersistentPreRunE` that calls `db.Open()` and stores `*DB` in context
- Each handler: parse flags → get store from context → call store method → format output
- Handlers define minimal interfaces for testability

### Commands (28 total)

**Root commands (human-facing, formatted text):**
- `oraculo install [--global|--local]` — stub in this phase
- `oraculo status` — dashboard showing epics, stories, tasks, progress
- `oraculo version` — version string

**Tools commands (agent-facing, JSON output):**

| Domain | Commands |
|--------|----------|
| Epic | `init`, `save`, `get`, `list`, `update`, `delete` |
| Story | `init`, `save`, `get`, `list`, `update`, `delete` |
| Task | `init`, `start`, `complete`, `fail`, `get`, `list`, `delete` |
| Memory | `store`, `search`, `domains` |
| Approval | `request`, `status`, `list`, `verdict` |

### Output format

- Tools commands: JSON on stdout, exit 0 for success, exit 1 for error
- Content commands (`get`): raw markdown on stdout
- Root commands: human-readable tables and text
- Error JSON: `{"error": "error_code", "message": "Human-readable message"}`

## Output Layer (`src/output/`)

- `WriteJSON(w, v)` — marshals any value to JSON
- `WriteError(w, err)` — maps `domain.Err*` sentinels to documented error JSON format
- `WriteTable(w, headers, rows)` — renders aligned text tables for root commands

## Installer (`src/installer/`)

Stub in this phase. Will eventually:
1. Copy skills from `claude-kit/skills/` to target Claude Code directory
2. Configure HTTP hooks in `settings.json`
3. Report what was installed

## File System

```
.oraculo/
├── oraculo.db                              # SQLite — auto-added to .gitignore
└── epics/
    └── <epic-name>/
        ├── requirements.md                 # Output of /oraculo:epic
        └── stories/
            └── <story-name>/
                └── requirements.md         # Output of /oraculo:story
```

- `db.Open()` creates `.oraculo/` and the database
- `epic save` / `story save` create directory structure and write markdown
- Lightweight epics (for standalone stories) are created automatically with empty description
- First bootstrap adds `.oraculo/oraculo.db` to `.gitignore`

## Testing Strategy

Three levels, TDD from the start:

### Level 1: Domain (pure unit tests)
- Table-driven tests for state transitions
- Validation logic tests
- Zero external dependencies

### Level 2: DB (integration with SQLite in-memory)
- `testDB(t)` helper: opens `:memory:`, runs migrations, returns clean `*DB`
- Tests idempotency, constraints, FTS5 search, DAG cycle detection
- Each test is isolated

### Level 3: CLI (end-to-end via Cobra)
- Execute full Cobra commands against a temp directory
- Verify JSON stdout and filesystem side effects
- Tests the complete flow: command → store → output

All tests run with `go test ./...`. No build tags or test categories.

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `modernc.org/sqlite` — Pure-Go SQLite (no CGO)

## Build Order

1. `src/domain/` — types, constants, validation, errors + tests
2. `src/db/` — Open, migrations, full schema, stores + integration tests
3. `src/output/` — JSON and table formatters + tests
4. `src/cli/` — all 28 commands wired to stores + E2E tests
5. `src/installer/` — stub
6. `cmd/oraculo/main.go` — wiring
