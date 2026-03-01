# Session & Phase Commands — Design

## 1. Purpose

The CLI Trust Layer lacks session and phase tracking. Skills reference `oraculo tools session status`, `oraculo tools session state`, and `oraculo tools phase complete` — none of which exist. Without these, skills cannot persist progress, enforce phase ordering, or resume after a session drop.

This design adds session lifecycle management, phase sequence enforcement, and a Claude Code hook for telemetry.

## 2. Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Phase data storage | JSON blob per phase | Skills define content freely. CLI stores without validating structure. Skills evolve without migrations. |
| Session concurrency | One active session per epic per type | Each epic is an independent context. Multiple epics can have concurrent sessions. |
| Phase ordering | Hardcoded in CLI per session type | CLI is the enforcement layer. Phase sequences are part of the contract. |
| Hook behavior | Register in SQLite + health check HTTP | Captures telemetry even without server. Notifies dashboard when available. |
| Session lifecycle | Auto-close on last phase + explicit close | Covers both normal completion and user abandonment. |

## 3. Data Model

### Schema (migration v2)

```sql
CREATE TABLE sessions (
    id         TEXT PRIMARY KEY,
    type       TEXT NOT NULL
               CHECK (type IN ('epic','story','plan','execute','validate')),
    epic_id    INTEGER REFERENCES epics(id),
    status     TEXT DEFAULT 'active'
               CHECK (status IN ('active','completed','abandoned')),
    created_at TEXT DEFAULT (datetime('now')),
    closed_at  TEXT
);

CREATE UNIQUE INDEX idx_sessions_active
    ON sessions(epic_id, type) WHERE status = 'active';

CREATE TABLE session_phases (
    session_id   TEXT NOT NULL REFERENCES sessions(id),
    phase        TEXT NOT NULL,
    data         TEXT DEFAULT '{}',
    completed_at TEXT DEFAULT (datetime('now')),
    PRIMARY KEY (session_id, phase)
);

CREATE TABLE claude_sessions (
    id         TEXT PRIMARY KEY,
    started_at TEXT DEFAULT (datetime('now')),
    metadata   TEXT DEFAULT '{}'
);
```

### Domain types (`src/domain/session.go`)

```go
type SessionType string

const (
    SessionEpic     SessionType = "epic"
    SessionStory    SessionType = "story"
    SessionPlan     SessionType = "plan"
    SessionExecute  SessionType = "execute"
    SessionValidate SessionType = "validate"
)

type SessionStatus string

const (
    SessionActive    SessionStatus = "active"
    SessionCompleted SessionStatus = "completed"
    SessionAbandoned SessionStatus = "abandoned"
)

type Session struct {
    ID        string
    Type      SessionType
    EpicID    *int
    Status    SessionStatus
    CreatedAt time.Time
    ClosedAt  *time.Time
}

type Phase struct {
    SessionID   string
    Name        string
    Data        string
    CompletedAt time.Time
}
```

### Phase ordering (`src/domain/phases.go`)

```go
var Phases = map[SessionType][]string{
    SessionEpic:     {"setup", "reframing", "divergence", "codebase", "convergence", "assumptions", "stress-test", "exit-gate", "artifact"},
    SessionStory:    {"setup", "reframing", "assumptions", "exit-gate", "artifact"},
    SessionPlan:     {"setup", "decomposition", "dependencies", "optimization", "artifact"},
    SessionExecute:  {"setup", "team-assembly", "monitoring", "completion"},
    SessionValidate: {"setup", "qa-dispatch", "verdict"},
}
```

Sequence validation: phase at index N completes only if phase at index N-1 exists in `session_phases`. Phase at index 0 (`setup`) always completes.

## 4. Store Layer (`src/db/session_store.go`)

```go
type SessionStore struct {
    db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore
```

### Methods

| Method | Purpose |
|--------|---------|
| `Create(session Session) error` | Start a new session. `ErrAlreadyExists` if active session exists for epic+type. |
| `Get(id string) (Session, error)` | Retrieve by ID. `ErrNotFound` if absent. |
| `ActiveByEpic(epicID int, sessionType SessionType) (Session, error)` | Find active session. `ErrNotFound` if none. |
| `Close(id string, status SessionStatus) error` | Mark completed or abandoned. `ErrInvalidTransition` if already closed. |
| `CompletePhase(sessionID, phase, data string) error` | Record phase completion. Validates sequence, rejects unknown phases, auto-closes on last phase. |
| `Phases(sessionID string) ([]Phase, error)` | All completed phases, ordered by completion time. |
| `CurrentPhase(sessionID string) (string, error)` | Next pending phase name. Empty string if all complete. |

### CompletePhase validation logic

1. Get session to determine type
2. Look up phase index in `domain.Phases[session.Type]` — unknown phase returns `ErrInvalidPhase`
3. If index > 0, check `session_phases` for `phase[index-1]` — missing returns `ErrInvalidTransition`
4. INSERT into `session_phases` — duplicate returns `ErrAlreadyExists`
5. If last phase in sequence, auto-close session as `completed`

## 5. CLI Commands

### Tools commands (agent-facing, JSON output)

**`oraculo tools session init --type <type> --epic <epic>`**

Creates an active session. Generates UUID. Resolves `--epic` to `epic_id` (creates epic if absent, following existing idempotent pattern). Returns `created: false` if active session already exists.

```json
{"id": "a1b2c3d4", "type": "epic", "epic": "gastos-app", "status": "active", "created": true}
```

**`oraculo tools session status --type <type> --epic <epic>`**

Returns active session state for bootstrap. Used at the start of every skill invocation.

```json
{"active": true, "id": "a1b2c3d4", "type": "epic", "epic": "gastos-app", "current_phase": "divergence", "completed_phases": ["setup", "reframing"]}
```

No active session:

```json
{"active": false}
```

**`oraculo tools session state --session <id>`**

Returns all completed phase data. Used for mid-session context reconstruction.

```json
{
  "id": "a1b2c3d4",
  "type": "epic",
  "status": "active",
  "current_phase": "divergence",
  "phases": {
    "setup": {"reasoning_level": "deep", "epic": "gastos-app"},
    "reframing": {"problem_statement": "...", "raw_idea": "...", "scope_out": ["..."]}
  }
}
```

**`oraculo tools session close --session <id> [--reason abandoned]`**

Closes session. Without `--reason`, closes as `completed`. With `--reason abandoned`, marks as abandoned.

```json
{"id": "a1b2c3d4", "status": "abandoned", "closed": true}
```

**`oraculo tools phase complete <phase> --session <id>`** (stdin: JSON)

Validates sequence, persists phase data, advances session. Stdin receives the JSON blob with validated phase outputs.

Success:
```json
{"phase": "reframing", "completed": true, "next": "divergence"}
```

Out of sequence:
```json
{"error": "invalid_transition", "message": "phase 'reframing' must be completed before 'divergence'"}
```

Unknown phase:
```json
{"error": "invalid_phase", "message": "phase 'unknown' is not valid for session type 'epic'"}
```

Last phase (auto-close):
```json
{"phase": "artifact", "completed": true, "next": "", "session_closed": true}
```

### Hook command (automatic, no stdout)

**`oraculo hook session-start`**

Runs automatically when a Claude Code session starts. No stdout. Warning on stderr if dashboard is offline. Always exits 0.

Behavior:
1. Read port from `.oraculo/config.json` (if exists)
2. Register in SQLite: INSERT into `claude_sessions`
3. If port configured: GET `http://localhost:<port>/health`
   - Online: POST `/hooks/session-start` with metadata
   - Offline: stderr warning
4. Exit 0

Metadata collected: `session_id` (UUID), `working_dir` (os.Getwd), `git_branch` (git rev-parse), `started_at`.

### File structure

```
src/domain/
├── session.go          # Session, Phase, SessionType, SessionStatus
├── session_test.go     # Unit tests
├── phases.go           # Phases map
├── phases_test.go      # Unit tests
└── errors.go           # + ErrInvalidPhase

src/db/
├── migrations.go       # + migrateV2
├── session_store.go    # SessionStore
└── session_store_test.go

src/cli/
├── root.go             # + register hook command
├── hook.go             # "hook" parent command
├── hook_session.go     # "hook session-start"
├── hook_session_test.go
└── tools/
    ├── tools.go        # + register session, phase commands
    ├── session.go      # session init|status|state|close
    ├── session_test.go
    ├── phase.go        # phase complete
    └── phase_test.go
```

## 6. Tests

### Domain unit tests (`src/domain/session_test.go`, `phases_test.go`)

- `SessionType.Valid` / `SessionStatus.Valid` for valid and invalid values
- `PhaseIndex` lookup returns correct position
- `PhaseIndex` for unknown phase returns -1

### DB integration tests (`src/db/session_store_test.go`)

Against `:memory:` SQLite:

- Create: happy path, idempotent (created=false), duplicate active (ErrAlreadyExists)
- Get: found, not found (ErrNotFound)
- ActiveByEpic: found, not found
- Close: happy path, already closed (ErrInvalidTransition)
- CompletePhase: happy path, out of order (ErrInvalidTransition), duplicate (ErrAlreadyExists), unknown phase (ErrInvalidPhase), auto-close on last phase
- Phases: returns ordered list
- CurrentPhase: returns next pending phase

### CLI E2E tests (`src/cli/tools/session_test.go`, `phase_test.go`)

Via `cobra.Command.Execute()`:

- session init: creates session, JSON output, idempotent
- session status: active session, no session
- session state: returns phase data
- session close: completed, abandoned
- phase complete: happy path, out of order, last phase auto-close

### Hook test (`src/cli/hook_session_test.go`)

- No config.json: skips health check, registers in SQLite
- Always exits zero

## 7. Build Order

```
1. domain types + phases + error     → unit tests
2. migration v2                      → verify schema
3. session_store                     → integration tests
4. CLI session + phase commands      → E2E tests
5. CLI hook session-start            → E2E test
```

## 8. What This Design Does NOT Cover

- Skills (epic, story, plan, execute, validate) — next design
- HTTP server, WebSocket, MCP server
- Changes to existing tables
- `oraculo install` implementation
- `oraculo update` or `oraculo uninstall`
