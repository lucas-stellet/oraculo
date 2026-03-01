# Session & Phase Commands — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add session lifecycle management and phase sequence enforcement to the CLI Trust Layer — the foundation skills depend on.

**Architecture:** New `sessions` and `session_phases` tables in SQLite, a `SessionStore` with sequence validation, five `tools` subcommands (session init/status/state/close, phase complete), and a `hook session-start` command. Follows the existing bottom-up pattern: domain → db → cli.

**Tech Stack:** Go 1.24, cobra, modernc/sqlite (pure Go), github.com/google/uuid

**Design doc:** `docs/plans/2026-03-01-session-phase-commands-design.md`

---

### Task 1: Domain types — SessionType and SessionStatus

**Files:**
- Create: `src/domain/session.go`

**Step 1: Write the failing test**

Create `src/domain/session_test.go`:

```go
package domain

import "testing"

func TestSessionType_Valid(t *testing.T) {
	tests := []struct {
		s    SessionType
		want bool
	}{
		{SessionEpic, true},
		{SessionStory, true},
		{SessionPlan, true},
		{SessionExecute, true},
		{SessionValidate, true},
		{SessionType("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("SessionType(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestSessionStatus_Valid(t *testing.T) {
	tests := []struct {
		s    SessionStatus
		want bool
	}{
		{SessionActive, true},
		{SessionCompleted, true},
		{SessionAbandoned, true},
		{SessionStatus("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("SessionStatus(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/domain/ -run TestSession -v`
Expected: FAIL — `SessionType` undefined

**Step 3: Write minimal implementation**

Create `src/domain/session.go`:

```go
// src/domain/session.go
package domain

import "time"

// SessionType identifies which command owns a session.
type SessionType string

const (
	SessionEpic     SessionType = "epic"
	SessionStory    SessionType = "story"
	SessionPlan     SessionType = "plan"
	SessionExecute  SessionType = "execute"
	SessionValidate SessionType = "validate"
)

var validSessionTypes = map[SessionType]bool{
	SessionEpic:     true,
	SessionStory:    true,
	SessionPlan:     true,
	SessionExecute:  true,
	SessionValidate: true,
}

// Valid reports whether s is a recognized session type.
func (s SessionType) Valid() bool {
	return validSessionTypes[s]
}

// SessionStatus represents the lifecycle state of a session.
type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
	SessionAbandoned SessionStatus = "abandoned"
)

var validSessionStatuses = map[SessionStatus]bool{
	SessionActive:    true,
	SessionCompleted: true,
	SessionAbandoned: true,
}

// Valid reports whether s is a recognized session status.
func (s SessionStatus) Valid() bool {
	return validSessionStatuses[s]
}

// Session tracks a skill invocation's progress through its phases.
type Session struct {
	ID        string
	Type      SessionType
	EpicID    *int
	Status    SessionStatus
	CreatedAt time.Time
	ClosedAt  *time.Time
}

// Phase records one completed phase within a session.
type Phase struct {
	SessionID   string
	Name        string
	Data        string // JSON blob, skill-defined
	CompletedAt time.Time
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/domain/ -run TestSession -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/domain/session.go src/domain/session_test.go
git commit -m "feat(domain): add Session and Phase types with validation"
```

---

### Task 2: Domain — Phase ordering and PhaseIndex

**Files:**
- Create: `src/domain/phases.go`
- Create: `src/domain/phases_test.go`

**Step 1: Write the failing test**

Create `src/domain/phases_test.go`:

```go
package domain

import "testing"

func TestPhaseIndex(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		phase       string
		want        int
	}{
		{SessionEpic, "setup", 0},
		{SessionEpic, "reframing", 1},
		{SessionEpic, "artifact", 8},
		{SessionStory, "setup", 0},
		{SessionStory, "artifact", 4},
		{SessionPlan, "decomposition", 1},
		{SessionExecute, "team-assembly", 1},
		{SessionValidate, "verdict", 2},
	}
	for _, tt := range tests {
		got := PhaseIndex(tt.sessionType, tt.phase)
		if got != tt.want {
			t.Errorf("PhaseIndex(%q, %q) = %d, want %d", tt.sessionType, tt.phase, got, tt.want)
		}
	}
}

func TestPhaseIndex_Unknown(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		phase       string
	}{
		{SessionEpic, "nonexistent"},
		{SessionType("bad"), "setup"},
		{SessionStory, "divergence"}, // valid for epic, not story
	}
	for _, tt := range tests {
		got := PhaseIndex(tt.sessionType, tt.phase)
		if got != -1 {
			t.Errorf("PhaseIndex(%q, %q) = %d, want -1", tt.sessionType, tt.phase, got)
		}
	}
}

func TestPhasesCount(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		want        int
	}{
		{SessionEpic, 9},
		{SessionStory, 5},
		{SessionPlan, 5},
		{SessionExecute, 4},
		{SessionValidate, 3},
	}
	for _, tt := range tests {
		got := len(Phases[tt.sessionType])
		if got != tt.want {
			t.Errorf("len(Phases[%q]) = %d, want %d", tt.sessionType, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/domain/ -run TestPhase -v`
Expected: FAIL — `Phases` undefined

**Step 3: Write minimal implementation**

Create `src/domain/phases.go`:

```go
// src/domain/phases.go
package domain

// Phases defines the ordered sequence for each session type.
// The CLI enforces that phase N completes only if phase N-1 exists.
var Phases = map[SessionType][]string{
	SessionEpic:     {"setup", "reframing", "divergence", "codebase", "convergence", "assumptions", "stress-test", "exit-gate", "artifact"},
	SessionStory:    {"setup", "reframing", "assumptions", "exit-gate", "artifact"},
	SessionPlan:     {"setup", "decomposition", "dependencies", "optimization", "artifact"},
	SessionExecute:  {"setup", "team-assembly", "monitoring", "completion"},
	SessionValidate: {"setup", "qa-dispatch", "verdict"},
}

// PhaseIndex returns the position of phase in the sequence for the given
// session type. It returns -1 if the session type or phase is not recognized.
func PhaseIndex(sessionType SessionType, phase string) int {
	phases, ok := Phases[sessionType]
	if !ok {
		return -1
	}
	for i, p := range phases {
		if p == phase {
			return i
		}
	}
	return -1
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/domain/ -run TestPhase -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/domain/phases.go src/domain/phases_test.go
git commit -m "feat(domain): add phase ordering with sequence lookup"
```

---

### Task 3: Domain — ErrInvalidPhase + update WriteError

**Files:**
- Modify: `src/domain/errors.go`
- Modify: `src/output/json.go`

**Step 1: Add ErrInvalidPhase to errors.go**

Add to `src/domain/errors.go`:

```go
ErrInvalidPhase = errors.New("invalid phase")
```

**Step 2: Add error mapping to output/json.go**

Add a case to the `WriteError` switch in `src/output/json.go`:

```go
case errors.Is(err, domain.ErrInvalidPhase):
	code = "invalid_phase"
```

**Step 3: Run existing tests to verify nothing broke**

Run: `go test ./src/domain/ ./src/output/ -v`
Expected: PASS

**Step 4: Commit**

```bash
git add src/domain/errors.go src/output/json.go
git commit -m "feat(domain): add ErrInvalidPhase sentinel error"
```

---

### Task 4: Migration v2 — sessions, session_phases, claude_sessions

**Files:**
- Modify: `src/db/migrations.go`

**Step 1: Write the migration**

Add `migrateV2` to `src/db/migrations.go`. Register it in the `migrations` slice.

```go
var migrations = []func(*sql.Tx) error{
	migrateV1,
	migrateV2,
}

func migrateV2(tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE sessions (
			id         TEXT PRIMARY KEY,
			type       TEXT NOT NULL
			           CHECK (type IN ('epic','story','plan','execute','validate')),
			epic_id    INTEGER REFERENCES epics(id),
			status     TEXT DEFAULT 'active'
			           CHECK (status IN ('active','completed','abandoned')),
			created_at TEXT DEFAULT (datetime('now')),
			closed_at  TEXT
		)`,
		`CREATE UNIQUE INDEX idx_sessions_active
			ON sessions(epic_id, type) WHERE status = 'active'`,
		`CREATE TABLE session_phases (
			session_id   TEXT NOT NULL REFERENCES sessions(id),
			phase        TEXT NOT NULL,
			data         TEXT DEFAULT '{}',
			completed_at TEXT DEFAULT (datetime('now')),
			PRIMARY KEY (session_id, phase)
		)`,
		`CREATE TABLE claude_sessions (
			id         TEXT PRIMARY KEY,
			started_at TEXT DEFAULT (datetime('now')),
			metadata   TEXT DEFAULT '{}'
		)`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration v2: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}
```

**Step 2: Run existing tests to verify migration runs cleanly**

Run: `go test ./src/db/ -v`
Expected: PASS — all existing tests pass, `:memory:` DBs now run both migrations.

**Step 3: Commit**

```bash
git add src/db/migrations.go
git commit -m "feat(db): add migration v2 — sessions, session_phases, claude_sessions"
```

---

### Task 5: SessionStore — Create and Get

**Files:**
- Create: `src/db/session_store.go`
- Create: `src/db/session_store_test.go`

**Step 1: Write the failing tests**

Create `src/db/session_store_test.go`:

```go
package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func createEpicForTest(t *testing.T, database *DB, name string) int {
	t.Helper()
	store := NewEpicStore(database)
	epic, _, err := store.Create(name, "")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	return epic.ID
}

func TestSessionStore_Create(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, created, err := store.Create(domain.SessionEpic, &epicID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if session.Type != domain.SessionEpic {
		t.Errorf("Type = %q, want %q", session.Type, domain.SessionEpic)
	}
	if session.Status != domain.SessionActive {
		t.Errorf("Status = %q, want %q", session.Status, domain.SessionActive)
	}
	if session.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestSessionStore_Create_Idempotent(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	s1, _, _ := store.Create(domain.SessionEpic, &epicID)
	s2, created, err := store.Create(domain.SessionEpic, &epicID)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if s2.ID != s1.ID {
		t.Errorf("IDs differ: %q vs %q", s2.ID, s1.ID)
	}
}

func TestSessionStore_Get(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	created, _, _ := store.Create(domain.SessionEpic, &epicID)
	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewSessionStore(database)

	_, err := store.Get("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSessionStore_ActiveByEpic(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	created, _, _ := store.Create(domain.SessionEpic, &epicID)
	got, err := store.ActiveByEpic(epicID, domain.SessionEpic)
	if err != nil {
		t.Fatalf("ActiveByEpic: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestSessionStore_ActiveByEpic_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	_, err := store.ActiveByEpic(epicID, domain.SessionEpic)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestSessionStore -v`
Expected: FAIL — `NewSessionStore` undefined

**Step 3: Write minimal implementation**

Create `src/db/session_store.go`:

```go
// src/db/session_store.go
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/oraculo/src/domain"
)

// SessionStore provides operations for sessions and phases backed by SQLite.
type SessionStore struct {
	db *DB
}

// NewSessionStore returns a SessionStore that uses the given database connection.
func NewSessionStore(db *DB) *SessionStore {
	return &SessionStore{db: db}
}

// Create starts a new session. If an active session of the same type already
// exists for the given epic, it returns that session with created=false.
func (s *SessionStore) Create(sessionType domain.SessionType, epicID *int) (*domain.Session, bool, error) {
	// Check for existing active session
	if epicID != nil {
		existing, err := s.ActiveByEpic(*epicID, sessionType)
		if err == nil {
			return existing, false, nil
		}
	}
	id := uuid.New().String()
	_, err := s.db.conn.Exec(
		"INSERT INTO sessions (id, type, epic_id) VALUES (?, ?, ?)",
		id, string(sessionType), epicID,
	)
	if err != nil {
		return nil, false, fmt.Errorf("create session: %w", err)
	}
	session, err := s.Get(id)
	if err != nil {
		return nil, false, fmt.Errorf("fetch after create: %w", err)
	}
	return session, true, nil
}

func scanSession(row interface{ Scan(...any) error }) (*domain.Session, error) {
	var (
		sess              domain.Session
		epicID            sql.NullInt64
		closedAt          sql.NullString
		createdAt         string
	)
	if err := row.Scan(
		&sess.ID, &sess.Type, &epicID, &sess.Status,
		&createdAt, &closedAt,
	); err != nil {
		return nil, err
	}
	var err error
	if sess.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	if epicID.Valid {
		v := int(epicID.Int64)
		sess.EpicID = &v
	}
	if closedAt.Valid {
		t, err := time.Parse(sqliteTimeFmt, closedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse closed_at: %w", err)
		}
		sess.ClosedAt = &t
	}
	return &sess, nil
}

// Get retrieves a session by ID. Returns ErrNotFound if absent.
func (s *SessionStore) Get(id string) (*domain.Session, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, type, epic_id, status, created_at, closed_at FROM sessions WHERE id = ?",
		id,
	)
	sess, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session %q: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

// ActiveByEpic returns the active session of the given type for an epic.
// Returns ErrNotFound if no active session exists.
func (s *SessionStore) ActiveByEpic(epicID int, sessionType domain.SessionType) (*domain.Session, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, type, epic_id, status, created_at, closed_at FROM sessions WHERE epic_id = ? AND type = ? AND status = 'active'",
		epicID, string(sessionType),
	)
	sess, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("active session for epic %d type %q: %w", epicID, sessionType, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("active by epic: %w", err)
	}
	return sess, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/db/ -run TestSessionStore -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/db/session_store.go src/db/session_store_test.go
git commit -m "feat(db): add SessionStore with Create, Get, ActiveByEpic"
```

---

### Task 6: SessionStore — Close

**Files:**
- Modify: `src/db/session_store.go`
- Modify: `src/db/session_store_test.go`

**Step 1: Write the failing tests**

Append to `src/db/session_store_test.go`:

```go
func TestSessionStore_Close(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	if err := store.Close(session.ID, domain.SessionCompleted); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, err := store.Get(session.ID)
	if err != nil {
		t.Fatalf("Get after close: %v", err)
	}
	if got.Status != domain.SessionCompleted {
		t.Errorf("Status = %q, want %q", got.Status, domain.SessionCompleted)
	}
	if got.ClosedAt == nil {
		t.Error("expected non-nil ClosedAt")
	}
}

func TestSessionStore_Close_Abandoned(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	if err := store.Close(session.ID, domain.SessionAbandoned); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, _ := store.Get(session.ID)
	if got.Status != domain.SessionAbandoned {
		t.Errorf("Status = %q, want %q", got.Status, domain.SessionAbandoned)
	}
}

func TestSessionStore_Close_AlreadyClosed(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	store.Close(session.ID, domain.SessionCompleted)

	err := store.Close(session.ID, domain.SessionCompleted)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestSessionStore_Close_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewSessionStore(database)

	err := store.Close("nonexistent", domain.SessionCompleted)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestSessionStore_Close -v`
Expected: FAIL — `Close` method undefined

**Step 3: Write minimal implementation**

Add to `src/db/session_store.go`:

```go
// Close marks a session as completed or abandoned.
// Returns ErrNotFound if absent, ErrInvalidTransition if already closed.
func (s *SessionStore) Close(id string, status domain.SessionStatus) error {
	res, err := s.db.conn.Exec(
		"UPDATE sessions SET status = ?, closed_at = datetime('now'), updated_at = datetime('now') WHERE id = ? AND status = 'active'",
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("close session: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		// Distinguish not found from already closed
		_, getErr := s.Get(id)
		if getErr != nil {
			return fmt.Errorf("session %q: %w", id, domain.ErrNotFound)
		}
		return fmt.Errorf("session %q already closed: %w", id, domain.ErrInvalidTransition)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/db/ -run TestSessionStore_Close -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/db/session_store.go src/db/session_store_test.go
git commit -m "feat(db): add SessionStore.Close with transition validation"
```

---

### Task 7: SessionStore — CompletePhase

**Files:**
- Modify: `src/db/session_store.go`
- Modify: `src/db/session_store_test.go`

**Step 1: Write the failing tests**

Append to `src/db/session_store_test.go`:

```go
func TestSessionStore_CompletePhase(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	err := store.CompletePhase(session.ID, "setup", `{"reasoning_level":"deep"}`)
	if err != nil {
		t.Fatalf("CompletePhase: %v", err)
	}
}

func TestSessionStore_CompletePhase_OutOfOrder(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	// Skip setup, try to complete reframing
	err := store.CompletePhase(session.ID, "reframing", "{}")
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestSessionStore_CompletePhase_Duplicate(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	store.CompletePhase(session.ID, "setup", "{}")
	err := store.CompletePhase(session.ID, "setup", "{}")
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got: %v", err)
	}
}

func TestSessionStore_CompletePhase_UnknownPhase(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	err := store.CompletePhase(session.ID, "nonexistent", "{}")
	if !errors.Is(err, domain.ErrInvalidPhase) {
		t.Errorf("expected ErrInvalidPhase, got: %v", err)
	}
}

func TestSessionStore_CompletePhase_AutoClose(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionValidate, &epicID) // 3 phases: setup, qa-dispatch, verdict
	store.CompletePhase(session.ID, "setup", "{}")
	store.CompletePhase(session.ID, "qa-dispatch", "{}")
	err := store.CompletePhase(session.ID, "verdict", "{}")
	if err != nil {
		t.Fatalf("CompletePhase last: %v", err)
	}
	got, _ := store.Get(session.ID)
	if got.Status != domain.SessionCompleted {
		t.Errorf("Status = %q, want %q after last phase", got.Status, domain.SessionCompleted)
	}
}

func TestSessionStore_CompletePhase_SessionNotFound(t *testing.T) {
	database := testDB(t)
	store := NewSessionStore(database)

	err := store.CompletePhase("nonexistent", "setup", "{}")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestSessionStore_CompletePhase -v`
Expected: FAIL — `CompletePhase` method undefined

**Step 3: Write minimal implementation**

Add to `src/db/session_store.go`:

```go
// CompletePhase records a phase completion with its data.
// Validates that the previous phase in the sequence has been completed.
// Auto-closes the session when the last phase completes.
func (s *SessionStore) CompletePhase(sessionID, phase, data string) error {
	session, err := s.Get(sessionID)
	if err != nil {
		return err
	}
	if session.Status != domain.SessionActive {
		return fmt.Errorf("session %q is %s: %w", sessionID, session.Status, domain.ErrInvalidTransition)
	}

	idx := domain.PhaseIndex(session.Type, phase)
	if idx < 0 {
		return fmt.Errorf("phase %q is not valid for session type %q: %w", phase, session.Type, domain.ErrInvalidPhase)
	}

	// Check predecessor
	if idx > 0 {
		predecessor := domain.Phases[session.Type][idx-1]
		var count int
		err := s.db.conn.QueryRow(
			"SELECT COUNT(*) FROM session_phases WHERE session_id = ? AND phase = ?",
			sessionID, predecessor,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("check predecessor: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("phase %q must be completed before %q: %w", predecessor, phase, domain.ErrInvalidTransition)
		}
	}

	_, err = s.db.conn.Exec(
		"INSERT INTO session_phases (session_id, phase, data) VALUES (?, ?, ?)",
		sessionID, phase, data,
	)
	if err != nil {
		// PRIMARY KEY violation = duplicate
		return fmt.Errorf("phase %q already completed: %w", phase, domain.ErrAlreadyExists)
	}

	// Auto-close if last phase
	phases := domain.Phases[session.Type]
	if idx == len(phases)-1 {
		return s.Close(sessionID, domain.SessionCompleted)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/db/ -run TestSessionStore_CompletePhase -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/db/session_store.go src/db/session_store_test.go
git commit -m "feat(db): add SessionStore.CompletePhase with sequence validation"
```

---

### Task 8: SessionStore — Phases and CurrentPhase

**Files:**
- Modify: `src/db/session_store.go`
- Modify: `src/db/session_store_test.go`

**Step 1: Write the failing tests**

Append to `src/db/session_store_test.go`:

```go
func TestSessionStore_Phases(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	store.CompletePhase(session.ID, "setup", `{"level":"deep"}`)
	store.CompletePhase(session.ID, "reframing", `{"problem":"test"}`)

	phases, err := store.Phases(session.ID)
	if err != nil {
		t.Fatalf("Phases: %v", err)
	}
	if len(phases) != 2 {
		t.Fatalf("len = %d, want 2", len(phases))
	}
	if phases[0].Name != "setup" {
		t.Errorf("phases[0].Name = %q, want %q", phases[0].Name, "setup")
	}
	if phases[1].Name != "reframing" {
		t.Errorf("phases[1].Name = %q, want %q", phases[1].Name, "reframing")
	}
	if phases[0].Data != `{"level":"deep"}` {
		t.Errorf("phases[0].Data = %q, want %q", phases[0].Data, `{"level":"deep"}`)
	}
}

func TestSessionStore_Phases_Empty(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	phases, err := store.Phases(session.ID)
	if err != nil {
		t.Fatalf("Phases: %v", err)
	}
	if len(phases) != 0 {
		t.Errorf("len = %d, want 0", len(phases))
	}
}

func TestSessionStore_CurrentPhase(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)

	// No phases completed — current is setup
	current, err := store.CurrentPhase(session.ID)
	if err != nil {
		t.Fatalf("CurrentPhase: %v", err)
	}
	if current != "setup" {
		t.Errorf("current = %q, want %q", current, "setup")
	}

	// Complete setup — current is reframing
	store.CompletePhase(session.ID, "setup", "{}")
	current, _ = store.CurrentPhase(session.ID)
	if current != "reframing" {
		t.Errorf("current = %q, want %q", current, "reframing")
	}
}

func TestSessionStore_CurrentPhase_AllDone(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionValidate, &epicID) // 3 phases
	store.CompletePhase(session.ID, "setup", "{}")
	store.CompletePhase(session.ID, "qa-dispatch", "{}")
	store.CompletePhase(session.ID, "verdict", "{}")

	current, err := store.CurrentPhase(session.ID)
	if err != nil {
		t.Fatalf("CurrentPhase: %v", err)
	}
	if current != "" {
		t.Errorf("current = %q, want empty", current)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run "TestSessionStore_Phases|TestSessionStore_CurrentPhase" -v`
Expected: FAIL — `Phases` and `CurrentPhase` methods undefined

**Step 3: Write minimal implementation**

Add to `src/db/session_store.go`:

```go
// Phases returns all completed phases for a session, ordered by completion time.
func (s *SessionStore) Phases(sessionID string) ([]domain.Phase, error) {
	rows, err := s.db.conn.Query(
		"SELECT session_id, phase, data, completed_at FROM session_phases WHERE session_id = ? ORDER BY completed_at",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}
	defer rows.Close()

	var phases []domain.Phase
	for rows.Next() {
		var p domain.Phase
		var completedAt string
		if err := rows.Scan(&p.SessionID, &p.Name, &p.Data, &completedAt); err != nil {
			return nil, fmt.Errorf("scan phase: %w", err)
		}
		if p.CompletedAt, err = time.Parse(sqliteTimeFmt, completedAt); err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		phases = append(phases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate phases: %w", err)
	}
	return phases, nil
}

// CurrentPhase returns the name of the next pending phase.
// If all phases are complete, returns empty string and nil error.
func (s *SessionStore) CurrentPhase(sessionID string) (string, error) {
	session, err := s.Get(sessionID)
	if err != nil {
		return "", err
	}
	completed, err := s.Phases(sessionID)
	if err != nil {
		return "", err
	}
	done := make(map[string]bool, len(completed))
	for _, p := range completed {
		done[p.Name] = true
	}
	for _, phase := range domain.Phases[session.Type] {
		if !done[phase] {
			return phase, nil
		}
	}
	return "", nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/db/ -run "TestSessionStore_Phases|TestSessionStore_CurrentPhase" -v`
Expected: PASS

**Step 5: Run all DB tests**

Run: `go test ./src/db/ -v`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add src/db/session_store.go src/db/session_store_test.go
git commit -m "feat(db): add SessionStore.Phases and CurrentPhase"
```

---

### Task 9: CLI — session init and session status commands

**Files:**
- Create: `src/cli/tools/session.go`
- Modify: `src/cli/tools/tools.go` (register session command)

**Step 1: Write the failing tests**

Create `src/cli/tools/session_test.go`:

```go
package tools_test

import (
	"encoding/json"
	"testing"
)

func TestSessionInit(t *testing.T) {
	setupTestDir(t)

	// Create the epic first
	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["type"] != "epic" {
		t.Errorf("type = %v, want %q", result["type"], "epic")
	}
	if result["epic"] != "my-epic" {
		t.Errorf("epic = %v, want %q", result["epic"], "my-epic")
	}
	if result["status"] != "active" {
		t.Errorf("status = %v, want %q", result["status"], "active")
	}
	if result["created"] != true {
		t.Errorf("created = %v, want true", result["created"])
	}
	if result["id"] == nil || result["id"] == "" {
		t.Error("expected non-empty id")
	}
}

func TestSessionInit_Idempotent(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["created"] != false {
		t.Errorf("created = %v, want false", result["created"])
	}
}

func TestSessionStatus_Active(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	_, _ = executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	out, err := executeCmd(t, "tools", "session", "status", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}
	if result["current_phase"] != "setup" {
		t.Errorf("current_phase = %v, want %q", result["current_phase"], "setup")
	}
}

func TestSessionStatus_NoSession(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")

	out, err := executeCmd(t, "tools", "session", "status", "--type", "epic", "--epic", "my-epic")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["active"] != false {
		t.Errorf("active = %v, want false", result["active"])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/tools/ -run "TestSession" -v`
Expected: FAIL — "session" command not registered

**Step 3: Write minimal implementation**

Create `src/cli/tools/session.go`:

```go
// src/cli/tools/session.go
package tools

import (
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/output"
	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage skill sessions",
	}
	cmd.AddCommand(
		newSessionInitCmd(),
		newSessionStatusCmd(),
		newSessionStateCmd(),
		newSessionCloseCmd(),
	)
	return cmd
}

func newSessionInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Start a new session",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionType, _ := cmd.Flags().GetString("type")
			epicName, _ := cmd.Flags().GetString("epic")

			epicStore := db.NewEpicStore(database)
			epic, _, err := epicStore.Create(epicName, "")
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			sessionStore := db.NewSessionStore(database)
			session, created, err := sessionStore.Create(domain.SessionType(sessionType), &epic.ID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			return output.WriteJSON(w, map[string]any{
				"id":      session.ID,
				"type":    session.Type,
				"epic":    epicName,
				"status":  session.Status,
				"created": created,
			})
		},
	}
	cmd.Flags().String("type", "", "Session type (epic, story, plan, execute, validate)")
	cmd.Flags().String("epic", "", "Epic name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("epic")
	return cmd
}

func newSessionStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check for an active session",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionType, _ := cmd.Flags().GetString("type")
			epicName, _ := cmd.Flags().GetString("epic")

			epicStore := db.NewEpicStore(database)
			epic, err := epicStore.GetByName(epicName)
			if err != nil {
				return output.WriteJSON(w, map[string]any{"active": false})
			}

			sessionStore := db.NewSessionStore(database)
			session, err := sessionStore.ActiveByEpic(epic.ID, domain.SessionType(sessionType))
			if err != nil {
				return output.WriteJSON(w, map[string]any{"active": false})
			}

			currentPhase, _ := sessionStore.CurrentPhase(session.ID)
			phases, _ := sessionStore.Phases(session.ID)
			completedNames := make([]string, len(phases))
			for i, p := range phases {
				completedNames[i] = p.Name
			}

			return output.WriteJSON(w, map[string]any{
				"active":           true,
				"id":               session.ID,
				"type":             session.Type,
				"epic":             epicName,
				"current_phase":    currentPhase,
				"completed_phases": completedNames,
			})
		},
	}
	cmd.Flags().String("type", "", "Session type")
	cmd.Flags().String("epic", "", "Epic name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("epic")
	return cmd
}

func newSessionStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Get full session state with phase data",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionID, _ := cmd.Flags().GetString("session")
			sessionStore := db.NewSessionStore(database)

			session, err := sessionStore.Get(sessionID)
			if err != nil {
				output.WriteError(w, err)
				return err
			}

			currentPhase, _ := sessionStore.CurrentPhase(session.ID)
			phases, _ := sessionStore.Phases(session.ID)

			phaseData := make(map[string]any, len(phases))
			for _, p := range phases {
				var parsed any
				if err := json.Unmarshal([]byte(p.Data), &parsed); err != nil {
					phaseData[p.Name] = p.Data // raw string fallback
				} else {
					phaseData[p.Name] = parsed
				}
			}

			return output.WriteJSON(w, map[string]any{
				"id":            session.ID,
				"type":          session.Type,
				"status":        session.Status,
				"current_phase": currentPhase,
				"phases":        phaseData,
			})
		},
	}
	cmd.Flags().String("session", "", "Session ID")
	cmd.MarkFlagRequired("session")
	return cmd
}

func newSessionCloseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())

			sessionID, _ := cmd.Flags().GetString("session")
			reason, _ := cmd.Flags().GetString("reason")

			status := domain.SessionCompleted
			if reason == "abandoned" {
				status = domain.SessionAbandoned
			}

			sessionStore := db.NewSessionStore(database)
			if err := sessionStore.Close(sessionID, status); err != nil {
				output.WriteError(w, err)
				return err
			}
			return output.WriteJSON(w, map[string]any{
				"id":     sessionID,
				"status": status,
				"closed": true,
			})
		},
	}
	cmd.Flags().String("session", "", "Session ID")
	cmd.Flags().String("reason", "", "Close reason (abandoned)")
	cmd.MarkFlagRequired("session")
	return cmd
}
```

Note: `newSessionStateCmd` uses `encoding/json` — add it to the import block.

Register in `src/cli/tools/tools.go`:

```go
cmd.AddCommand(newEpicCmd(), newStoryCmd(), newTaskCmd(), newMemoryCmd(), newApprovalCmd(), newSessionCmd())
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/cli/tools/ -run "TestSession" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/cli/tools/session.go src/cli/tools/session_test.go src/cli/tools/tools.go
git commit -m "feat(cli): add session init, status, state, close commands"
```

---

### Task 10: CLI — session state and close E2E tests

**Files:**
- Modify: `src/cli/tools/session_test.go`

**Step 1: Write the tests**

Append to `src/cli/tools/session_test.go`:

```go
func TestSessionState(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	initOut, _ := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	var initResult map[string]any
	json.Unmarshal([]byte(initOut), &initResult)
	sessionID := initResult["id"].(string)

	// Complete a phase via phase complete (tested in Task 11)
	// For now, test state with no phases
	out, err := executeCmd(t, "tools", "session", "state", "--session", sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["id"] != sessionID {
		t.Errorf("id = %v, want %q", result["id"], sessionID)
	}
	if result["type"] != "epic" {
		t.Errorf("type = %v, want %q", result["type"], "epic")
	}
	if result["current_phase"] != "setup" {
		t.Errorf("current_phase = %v, want %q", result["current_phase"], "setup")
	}
}

func TestSessionClose(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	initOut, _ := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	var initResult map[string]any
	json.Unmarshal([]byte(initOut), &initResult)
	sessionID := initResult["id"].(string)

	out, err := executeCmd(t, "tools", "session", "close", "--session", sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["closed"] != true {
		t.Errorf("closed = %v, want true", result["closed"])
	}
	if result["status"] != "completed" {
		t.Errorf("status = %v, want %q", result["status"], "completed")
	}
}

func TestSessionClose_Abandoned(t *testing.T) {
	setupTestDir(t)

	_, _ = executeCmd(t, "tools", "epic", "init", "my-epic")
	initOut, _ := executeCmd(t, "tools", "session", "init", "--type", "epic", "--epic", "my-epic")

	var initResult map[string]any
	json.Unmarshal([]byte(initOut), &initResult)
	sessionID := initResult["id"].(string)

	out, err := executeCmd(t, "tools", "session", "close", "--session", sessionID, "--reason", "abandoned")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["status"] != "abandoned" {
		t.Errorf("status = %v, want %q", result["status"], "abandoned")
	}
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./src/cli/tools/ -run "TestSessionState|TestSessionClose" -v`
Expected: PASS (implementation already exists from Task 9)

**Step 3: Commit**

```bash
git add src/cli/tools/session_test.go
git commit -m "test(cli): add E2E tests for session state and close"
```

---

### Task 11: CLI — phase complete command

**Files:**
- Create: `src/cli/tools/phase.go`
- Create: `src/cli/tools/phase_test.go`

**Step 1: Write the failing tests**

Create `src/cli/tools/phase_test.go`:

```go
package tools_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

func initSessionForTest(t *testing.T, epicName, sessionType string) string {
	t.Helper()
	_, _ = executeCmd(t, "tools", "epic", "init", epicName)
	out, err := executeCmd(t, "tools", "session", "init", "--type", sessionType, "--epic", epicName)
	if err != nil {
		t.Fatalf("init session: %v\noutput: %s", err, out)
	}
	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	return result["id"].(string)
}

func executePhaseComplete(t *testing.T, sessionID, phase, data string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(data))
	cmd.SetArgs([]string{"tools", "phase", "complete", phase, "--session", sessionID})
	err := cmd.Execute()
	return buf.String(), err
}

func TestPhaseComplete(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "validate") // 3 phases: setup, qa-dispatch, verdict

	out, err := executePhaseComplete(t, sessionID, "setup", `{"level":"deep"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["phase"] != "setup" {
		t.Errorf("phase = %v, want %q", result["phase"], "setup")
	}
	if result["completed"] != true {
		t.Errorf("completed = %v, want true", result["completed"])
	}
	if result["next"] != "qa-dispatch" {
		t.Errorf("next = %v, want %q", result["next"], "qa-dispatch")
	}
}

func TestPhaseComplete_OutOfOrder(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "epic")

	out, err := executePhaseComplete(t, sessionID, "reframing", "{}")
	if err == nil {
		t.Fatal("expected error for out-of-order phase")
	}

	var result map[string]string
	json.Unmarshal([]byte(out), &result)
	if result["error"] != "invalid_transition" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_transition")
	}
}

func TestPhaseComplete_UnknownPhase(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "epic")

	out, err := executePhaseComplete(t, sessionID, "nonexistent", "{}")
	if err == nil {
		t.Fatal("expected error for unknown phase")
	}

	var result map[string]string
	json.Unmarshal([]byte(out), &result)
	if result["error"] != "invalid_phase" {
		t.Errorf("error = %q, want %q", result["error"], "invalid_phase")
	}
}

func TestPhaseComplete_LastPhase(t *testing.T) {
	setupTestDir(t)
	sessionID := initSessionForTest(t, "my-epic", "validate") // 3 phases

	executePhaseComplete(t, sessionID, "setup", "{}")
	executePhaseComplete(t, sessionID, "qa-dispatch", "{}")
	out, err := executePhaseComplete(t, sessionID, "verdict", "{}")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	json.Unmarshal([]byte(out), &result)
	if result["next"] != "" {
		t.Errorf("next = %v, want empty", result["next"])
	}
	if result["session_closed"] != true {
		t.Errorf("session_closed = %v, want true", result["session_closed"])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/tools/ -run "TestPhaseComplete" -v`
Expected: FAIL — "phase" command not registered

**Step 3: Write minimal implementation**

Create `src/cli/tools/phase.go`:

```go
// src/cli/tools/phase.go
package tools

import (
	"io"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/output"
	"github.com/spf13/cobra"
)

func newPhaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage session phases",
	}
	cmd.AddCommand(newPhaseCompleteCmd())
	return cmd
}

func newPhaseCompleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete <phase>",
		Short: "Complete a phase (reads JSON data from stdin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			phase := args[0]
			w := cmd.OutOrStdout()
			database := dbFromContext(cmd.Context())
			sessionID, _ := cmd.Flags().GetString("session")

			data, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				output.WriteError(w, err)
				return err
			}
			dataStr := string(data)
			if dataStr == "" {
				dataStr = "{}"
			}

			sessionStore := db.NewSessionStore(database)
			if err := sessionStore.CompletePhase(sessionID, phase, dataStr); err != nil {
				output.WriteError(w, err)
				return err
			}

			// Determine next phase
			session, _ := sessionStore.Get(sessionID)
			next, _ := sessionStore.CurrentPhase(sessionID)
			sessionClosed := session.Status == domain.SessionCompleted

			result := map[string]any{
				"phase":     phase,
				"completed": true,
				"next":      next,
			}
			if sessionClosed {
				result["session_closed"] = true
			}
			return output.WriteJSON(w, result)
		},
	}
	cmd.Flags().String("session", "", "Session ID")
	cmd.MarkFlagRequired("session")
	return cmd
}
```

Register in `src/cli/tools/tools.go`:

```go
cmd.AddCommand(newEpicCmd(), newStoryCmd(), newTaskCmd(), newMemoryCmd(), newApprovalCmd(), newSessionCmd(), newPhaseCmd())
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/cli/tools/ -run "TestPhaseComplete" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/cli/tools/phase.go src/cli/tools/phase_test.go src/cli/tools/tools.go
git commit -m "feat(cli): add phase complete command with sequence enforcement"
```

---

### Task 12: CLI — hook session-start

**Files:**
- Create: `src/cli/hook.go`
- Create: `src/cli/hook_session.go`
- Create: `src/cli/hook_session_test.go`
- Modify: `src/cli/root.go` (register hook command)

**Step 1: Write the failing test**

Create `src/cli/hook_session_test.go`:

```go
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

func TestHookSessionStart_NoConfig(t *testing.T) {
	// Setup temp dir
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	// Must always succeed (exit 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v\noutput: %s", err, buf.String())
	}

	// Should have created .oraculo/ and registered session
	dbPath := filepath.Join(".oraculo", "oraculo.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected .oraculo/oraculo.db to be created")
	}
}

func TestHookSessionStart_AlwaysExitsZero(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	t.Cleanup(func() { os.Chdir(orig) })

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook must always exit 0, got: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run "TestHookSession" -v`
Expected: FAIL — "hook" command not registered

**Step 3: Write minimal implementation**

Create `src/cli/hook.go`:

```go
// src/cli/hook.go
package cli

import "github.com/spf13/cobra"

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Commands triggered by Claude Code hooks",
	}
}
```

Create `src/cli/hook_session.go`:

```go
// src/cli/hook_session.go
package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/oraculo/src/db"
	"github.com/spf13/cobra"
)

func newHookSessionStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "session-start",
		Short: "Register a Claude Code session start",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Always succeed — never block Claude Code session
			if err := hookSessionStart(cmd); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "⚠ hook session-start: %v\n", err)
			}
			return nil
		},
	}
}

func hookSessionStart(cmd *cobra.Command) error {
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Collect metadata
	id := uuid.New().String()
	wd, _ := os.Getwd()
	branch := gitBranch()
	metadata := map[string]string{
		"session_id":  id,
		"working_dir": wd,
		"git_branch":  branch,
		"started_at":  time.Now().UTC().Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Register in SQLite
	_, err = database.Conn().Exec(
		"INSERT INTO claude_sessions (id, metadata) VALUES (?, ?)",
		id, string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("register session: %w", err)
	}

	// Health check if config exists
	port := readConfigPort()
	if port > 0 {
		healthURL := fmt.Sprintf("http://localhost:%d/health", port)
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(healthURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Fprintf(cmd.ErrOrStderr(), "⚠ Oraculo dashboard is offline. Run 'oraculo server' to start it.\n")
			return nil
		}
		// POST session-start
		postURL := fmt.Sprintf("http://localhost:%d/hooks/session-start", port)
		client.Post(postURL, "application/json", strings.NewReader(string(metadataJSON)))
	}
	return nil
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

type configFile struct {
	Port int `json:"port"`
}

func readConfigPort() int {
	data, err := os.ReadFile(".oraculo/config.json")
	if err != nil {
		return 0
	}
	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return 0
	}
	return cfg.Port
}
```

Register in `src/cli/root.go`:

```go
func NewRoot(version string) *cobra.Command {
	// ... existing code ...
	hookCmd := newHookCmd()
	hookCmd.AddCommand(newHookSessionStartCmd())
	root.AddCommand(
		newVersionCmd(version),
		newInstallCmd(),
		newStatusCmd(),
		tools.NewToolsCmd(),
		hookCmd,
	)
	return root
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/cli/ -run "TestHookSession" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/cli/hook.go src/cli/hook_session.go src/cli/hook_session_test.go src/cli/root.go
git commit -m "feat(cli): add hook session-start with SQLite registration and health check"
```

---

### Task 13: Full integration — run all tests

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 2: Run vet**

Run: `go vet ./...`
Expected: No issues

**Step 3: Build**

Run: `make build`
Expected: Binary compiles successfully

**Step 4: Manual smoke test**

```bash
./oraculo tools session init --type epic --epic smoke-test
./oraculo tools session status --type epic --epic smoke-test
echo '{"level":"deep"}' | ./oraculo tools phase complete setup --session <id-from-init>
./oraculo tools session state --session <id-from-init>
./oraculo tools session close --session <id-from-init> --reason abandoned
./oraculo hook session-start
```

**Step 5: Commit (if any fixes needed)**

```bash
git add -A
git commit -m "fix: address integration issues from full test run"
```

---

### Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Domain: SessionType, SessionStatus, Session, Phase | `src/domain/session.go`, `session_test.go` |
| 2 | Domain: Phases map + PhaseIndex | `src/domain/phases.go`, `phases_test.go` |
| 3 | Domain: ErrInvalidPhase + WriteError mapping | `src/domain/errors.go`, `src/output/json.go` |
| 4 | Migration v2: 3 new tables | `src/db/migrations.go` |
| 5 | SessionStore: Create, Get, ActiveByEpic | `src/db/session_store.go`, `session_store_test.go` |
| 6 | SessionStore: Close | `src/db/session_store.go`, `session_store_test.go` |
| 7 | SessionStore: CompletePhase | `src/db/session_store.go`, `session_store_test.go` |
| 8 | SessionStore: Phases, CurrentPhase | `src/db/session_store.go`, `session_store_test.go` |
| 9 | CLI: session init, status, state, close | `src/cli/tools/session.go`, `tools.go` |
| 10 | CLI: session state/close E2E tests | `src/cli/tools/session_test.go` |
| 11 | CLI: phase complete | `src/cli/tools/phase.go`, `phase_test.go`, `tools.go` |
| 12 | CLI: hook session-start | `src/cli/hook.go`, `hook_session.go`, `root.go` |
| 13 | Full integration: all tests + build + smoke | — |
