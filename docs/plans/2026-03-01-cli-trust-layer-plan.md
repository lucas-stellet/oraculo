# CLI Trust Layer Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the complete Oraculo CLI skeleton — a single Go binary with all 28 commands, SQLite persistence with the full schema, and tests at every layer.

**Architecture:** Bottom-up: pure domain types → SQLite stores → output formatting → Cobra commands → main entrypoint. Each layer depends only on the one below it. Interfaces defined at point of use (consumer-side), not in the domain.

**Tech Stack:** Go, Cobra (CLI framework), modernc.org/sqlite (pure-Go SQLite, no CGO)

**Design doc:** `docs/plans/2026-03-01-cli-trust-layer-design.md`
**CLI spec:** `docs/cli/design.md`

---

### Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `cmd/oraculo/main.go`
- Create: `src/domain/.gitkeep` (placeholder)

**Step 1: Initialize Go module**

```bash
cd /Users/lucas/dev/projects/oraculo
go mod init github.com/lucas/oraculo
```

**Step 2: Create directory structure**

```bash
mkdir -p cmd/oraculo
mkdir -p src/{domain,db,cli/tools,output,installer}
```

**Step 3: Add dependencies**

```bash
go get github.com/spf13/cobra@latest
go get modernc.org/sqlite@latest
```

**Step 4: Create minimal main.go**

```go
// cmd/oraculo/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "oraculo: not yet wired")
	os.Exit(1)
}
```

**Step 5: Verify it compiles**

Run: `go build -o /dev/null ./cmd/oraculo`
Expected: exits 0 (compiles clean)

**Step 6: Commit**

```bash
git add go.mod go.sum cmd/ src/
git commit -m "feat: initialize Go project with module and directory structure"
```

---

### Task 2: Domain Types — Epic, Story, Shared Types

**Files:**
- Create: `src/domain/epic.go`
- Create: `src/domain/story.go`
- Create: `src/domain/errors.go`
- Create: `src/domain/status.go`
- Create: `src/domain/status_test.go`

**Step 1: Write the test for ApprovalStatus validation**

```go
// src/domain/status_test.go
package domain

import "testing"

func TestApprovalStatus_Valid(t *testing.T) {
	tests := []struct {
		s    ApprovalStatus
		want bool
	}{
		{ApprovalNone, true},
		{ApprovalPending, true},
		{ApprovalApproved, true},
		{ApprovalRejected, true},
		{ApprovalNeedsRevision, true},
		{ApprovalStatus("invalid"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("ApprovalStatus(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/domain/ -run TestApprovalStatus -v`
Expected: FAIL — types not defined

**Step 3: Write status.go with shared status types**

```go
// src/domain/status.go
package domain

// ApprovalStatus represents the approval state of an epic or story.
type ApprovalStatus string

const (
	ApprovalNone          ApprovalStatus = "none"
	ApprovalPending       ApprovalStatus = "pending"
	ApprovalApproved      ApprovalStatus = "approved"
	ApprovalRejected      ApprovalStatus = "rejected"
	ApprovalNeedsRevision ApprovalStatus = "needs_revision"
)

var validApprovalStatuses = map[ApprovalStatus]bool{
	ApprovalNone:          true,
	ApprovalPending:       true,
	ApprovalApproved:      true,
	ApprovalRejected:      true,
	ApprovalNeedsRevision: true,
}

// Valid reports whether s is a recognized approval status.
func (s ApprovalStatus) Valid() bool {
	return validApprovalStatuses[s]
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/domain/ -run TestApprovalStatus -v`
Expected: PASS

**Step 5: Write errors.go**

```go
// src/domain/errors.go
package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrCyclicDependency  = errors.New("cyclic dependency in task graph")
	ErrMissingRequired   = errors.New("missing required field")
)
```

**Step 6: Write epic.go**

```go
// src/domain/epic.go
package domain

import "time"

// Epic represents a product engineering initiative.
type Epic struct {
	ID             int
	Name           string
	Description    string
	ApprovalStatus ApprovalStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
```

**Step 7: Write story.go**

```go
// src/domain/story.go
package domain

import "time"

// Story represents a work item scoped to an epic.
type Story struct {
	ID             int
	EpicID         int
	Name           string
	Description    string
	ApprovalStatus ApprovalStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
```

**Step 8: Run all domain tests**

Run: `go test ./src/domain/ -v`
Expected: PASS

**Step 9: Commit**

```bash
git add src/domain/
git commit -m "feat(domain): add Epic, Story types and shared status/errors"
```

---

### Task 3: Domain Types — Task with Lifecycle Validation

**Files:**
- Create: `src/domain/task.go`
- Create: `src/domain/task_test.go`

**Step 1: Write the failing test for task status transitions**

```go
// src/domain/task_test.go
package domain

import "testing"

func TestTaskStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from TaskStatus
		to   TaskStatus
		want bool
	}{
		{TaskPending, TaskInProgress, true},
		{TaskPending, TaskCompleted, false},
		{TaskPending, TaskFailed, false},
		{TaskPending, TaskPending, false},
		{TaskInProgress, TaskCompleted, true},
		{TaskInProgress, TaskFailed, true},
		{TaskInProgress, TaskPending, false},
		{TaskInProgress, TaskInProgress, false},
		{TaskCompleted, TaskPending, false},
		{TaskCompleted, TaskInProgress, false},
		{TaskCompleted, TaskFailed, false},
		{TaskFailed, TaskPending, false},
		{TaskFailed, TaskInProgress, false},
		{TaskFailed, TaskCompleted, false},
	}
	for _, tt := range tests {
		got := tt.from.CanTransitionTo(tt.to)
		if got != tt.want {
			t.Errorf("%s → %s: got %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestTaskStatus_Valid(t *testing.T) {
	tests := []struct {
		s    TaskStatus
		want bool
	}{
		{TaskPending, true},
		{TaskInProgress, true},
		{TaskCompleted, true},
		{TaskFailed, true},
		{TaskStatus("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("TaskStatus(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/domain/ -run TestTaskStatus -v`
Expected: FAIL — TaskStatus not defined

**Step 3: Write task.go**

```go
// src/domain/task.go
package domain

import "time"

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
)

var validTaskStatuses = map[TaskStatus]bool{
	TaskPending:    true,
	TaskInProgress: true,
	TaskCompleted:  true,
	TaskFailed:     true,
}

// Valid reports whether s is a recognized task status.
func (s TaskStatus) Valid() bool {
	return validTaskStatuses[s]
}

var validTransitions = map[TaskStatus][]TaskStatus{
	TaskPending:    {TaskInProgress},
	TaskInProgress: {TaskCompleted, TaskFailed},
}

// CanTransitionTo reports whether a transition from s to target is valid.
// Valid transitions: pending → in_progress → completed | failed.
func (s TaskStatus) CanTransitionTo(target TaskStatus) bool {
	for _, allowed := range validTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// Task represents an executable unit of work within a story.
type Task struct {
	ID            int
	StoryID       int
	Name          string
	Description   string
	Status        TaskStatus
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

// TaskResult holds rich completion data for a finished task.
type TaskResult struct {
	ID            int
	TaskID        int
	Summary       string
	Logs          string
	SkillsUsed    []string
	FilesModified []string
	CreatedAt     time.Time
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/domain/ -run TestTaskStatus -v`
Expected: PASS

**Step 5: Commit**

```bash
git add src/domain/task.go src/domain/task_test.go
git commit -m "feat(domain): add Task types with lifecycle validation"
```

---

### Task 4: Domain Types — Knowledge and Approval

**Files:**
- Create: `src/domain/memory.go`
- Create: `src/domain/memory_test.go`
- Create: `src/domain/approval.go`
- Create: `src/domain/approval_test.go`

**Step 1: Write the failing test for knowledge category validation**

```go
// src/domain/memory_test.go
package domain

import "testing"

func TestCategory_Valid(t *testing.T) {
	tests := []struct {
		c    Category
		want bool
	}{
		{CategoryPattern, true},
		{CategoryConvention, true},
		{CategoryConstraint, true},
		{CategoryDependency, true},
		{CategoryTest, true},
		{CategoryArchitecture, true},
		{Category("bogus"), false},
	}
	for _, tt := range tests {
		if got := tt.c.Valid(); got != tt.want {
			t.Errorf("Category(%q).Valid() = %v, want %v", tt.c, got, tt.want)
		}
	}
}

func TestConfidence_Valid(t *testing.T) {
	tests := []struct {
		c    Confidence
		want bool
	}{
		{ConfidenceHigh, true},
		{ConfidenceMedium, true},
		{ConfidenceLow, true},
		{Confidence("none"), false},
	}
	for _, tt := range tests {
		if got := tt.c.Valid(); got != tt.want {
			t.Errorf("Confidence(%q).Valid() = %v, want %v", tt.c, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/domain/ -run "TestCategory|TestConfidence" -v`
Expected: FAIL

**Step 3: Write memory.go**

```go
// src/domain/memory.go
package domain

import "time"

// Category classifies a knowledge finding.
type Category string

const (
	CategoryPattern      Category = "pattern"
	CategoryConvention   Category = "convention"
	CategoryConstraint   Category = "constraint"
	CategoryDependency   Category = "dependency"
	CategoryTest         Category = "test"
	CategoryArchitecture Category = "architecture"
)

var validCategories = map[Category]bool{
	CategoryPattern:      true,
	CategoryConvention:   true,
	CategoryConstraint:   true,
	CategoryDependency:   true,
	CategoryTest:         true,
	CategoryArchitecture: true,
}

// Valid reports whether c is a recognized category.
func (c Category) Valid() bool {
	return validCategories[c]
}

// Confidence indicates how certain a knowledge finding is.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

var validConfidences = map[Confidence]bool{
	ConfidenceHigh:   true,
	ConfidenceMedium: true,
	ConfidenceLow:    true,
}

// Valid reports whether c is a recognized confidence level.
func (c Confidence) Valid() bool {
	return validConfidences[c]
}

// Knowledge represents a codebase finding persisted in the knowledge table.
type Knowledge struct {
	ID          int
	Domain      string
	Category    Category
	Finding     string
	SourceFiles string
	Confidence  Confidence
	CreatedAt   time.Time
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/domain/ -run "TestCategory|TestConfidence" -v`
Expected: PASS

**Step 5: Write the failing test for approval types and verdicts**

```go
// src/domain/approval_test.go
package domain

import "testing"

func TestApprovalType_Valid(t *testing.T) {
	tests := []struct {
		a    ApprovalType
		want bool
	}{
		{ApprovalEpicRequirements, true},
		{ApprovalStoryDefinition, true},
		{ApprovalQAEscalation, true},
		{ApprovalExecutionPlan, true},
		{ApprovalType("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.a.Valid(); got != tt.want {
			t.Errorf("ApprovalType(%q).Valid() = %v, want %v", tt.a, got, tt.want)
		}
	}
}

func TestVerdict_Valid(t *testing.T) {
	tests := []struct {
		v    Verdict
		want bool
	}{
		{VerdictApproved, true},
		{VerdictRejected, true},
		{VerdictNeedsRevision, true},
		{Verdict("maybe"), false},
	}
	for _, tt := range tests {
		if got := tt.v.Valid(); got != tt.want {
			t.Errorf("Verdict(%q).Valid() = %v, want %v", tt.v, got, tt.want)
		}
	}
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./src/domain/ -run "TestApprovalType|TestVerdict" -v`
Expected: FAIL

**Step 7: Write approval.go**

```go
// src/domain/approval.go
package domain

import "time"

// ApprovalType identifies the kind of approval gate.
type ApprovalType string

const (
	ApprovalEpicRequirements ApprovalType = "epic-requirements"
	ApprovalStoryDefinition  ApprovalType = "story-definition"
	ApprovalQAEscalation     ApprovalType = "qa-escalation"
	ApprovalExecutionPlan    ApprovalType = "execution-plan"
)

var validApprovalTypes = map[ApprovalType]bool{
	ApprovalEpicRequirements: true,
	ApprovalStoryDefinition:  true,
	ApprovalQAEscalation:     true,
	ApprovalExecutionPlan:    true,
}

// Valid reports whether a is a recognized approval type.
func (a ApprovalType) Valid() bool {
	return validApprovalTypes[a]
}

// Verdict represents a human decision on an approval request.
type Verdict string

const (
	VerdictApproved      Verdict = "approved"
	VerdictRejected      Verdict = "rejected"
	VerdictNeedsRevision Verdict = "needs_revision"
)

var validVerdicts = map[Verdict]bool{
	VerdictApproved:      true,
	VerdictRejected:      true,
	VerdictNeedsRevision: true,
}

// Valid reports whether v is a recognized verdict.
func (v Verdict) Valid() bool {
	return validVerdicts[v]
}

// Approval represents a human-in-the-loop approval gate request.
type Approval struct {
	ID              string
	Type            ApprovalType
	EpicID          *int
	StoryID         *int
	Content         string
	PreviousVersion string
	Status          ApprovalStatus
	VerdictComment  string
	RequestedAt     time.Time
	DecidedAt       *time.Time
}
```

**Step 8: Run test to verify it passes**

Run: `go test ./src/domain/ -run "TestApprovalType|TestVerdict" -v`
Expected: PASS

**Step 9: Run all domain tests**

Run: `go test ./src/domain/ -v`
Expected: ALL PASS

**Step 10: Commit**

```bash
git add src/domain/memory.go src/domain/memory_test.go src/domain/approval.go src/domain/approval_test.go
git commit -m "feat(domain): add Knowledge and Approval types with validation"
```

---

### Task 5: DB Layer — Connection, Bootstrap, and Migrations

**Files:**
- Create: `src/db/db.go`
- Create: `src/db/migrations.go`
- Create: `src/db/testutil_test.go`
- Create: `src/db/db_test.go`

**Step 1: Write the failing test for DB open and migration**

```go
// src/db/db_test.go
package db

import "testing"

func TestOpen_CreatesSchemaOnFreshDB(t *testing.T) {
	database := testDB(t)

	// Verify the schema was created by querying for tables
	tables := []string{"epics", "stories", "tasks", "task_dependencies",
		"task_results", "validations", "knowledge", "approvals"}
	for _, table := range tables {
		var name string
		err := database.conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestOpen_MigrationIsIdempotent(t *testing.T) {
	database := testDB(t)

	// Running migrate again should not error
	if err := database.migrate(); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}

func TestOpen_TracksSchemaVersion(t *testing.T) {
	database := testDB(t)

	var version int
	err := database.conn.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		t.Fatalf("failed to read user_version: %v", err)
	}
	if version != len(migrations) {
		t.Errorf("user_version = %d, want %d", version, len(migrations))
	}
}
```

**Step 2: Write the test helper**

```go
// src/db/testutil_test.go
package db

import "testing"

// testDB opens an in-memory SQLite database with all migrations applied.
// The database is closed automatically when the test ends.
func testDB(t *testing.T) *DB {
	t.Helper()
	database, err := openPath(":memory:")
	if err != nil {
		t.Fatalf("testDB: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./src/db/ -run TestOpen -v`
Expected: FAIL — package doesn't exist

**Step 4: Write db.go — connection and bootstrap**

```go
// src/db/db.go
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps an SQLite connection with auto-bootstrap behavior.
type DB struct {
	conn *sql.DB
}

// Open opens the database at .oraculo/oraculo.db relative to the current directory.
// It creates .oraculo/ if missing and runs pending migrations.
func Open() (*DB, error) {
	dir := ".oraculo"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", dir, err)
	}
	return openPath(filepath.Join(dir, "oraculo.db"))
}

// openPath opens (or creates) a database at the given path and runs migrations.
func openPath(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set journal mode: %w", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

// Conn returns the underlying *sql.DB for use by stores.
func (d *DB) Conn() *sql.DB {
	return d.conn
}
```

**Step 5: Write migrations.go — full schema**

```go
// src/db/migrations.go
package db

import (
	"database/sql"
	"fmt"
)

var migrations = []func(*sql.Tx) error{
	migrateV1,
}

// migrateV1 creates all core tables, knowledge with FTS5, approvals, and validations.
func migrateV1(tx *sql.Tx) error {
	stmts := []string{
		// Core entities
		`CREATE TABLE epics (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			name            TEXT UNIQUE NOT NULL,
			description     TEXT DEFAULT '',
			approval_status TEXT DEFAULT 'none'
			                CHECK (approval_status IN ('none','pending','approved','rejected','needs_revision')),
			created_at      TEXT DEFAULT (datetime('now')),
			updated_at      TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE stories (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			epic_id         INTEGER NOT NULL REFERENCES epics(id),
			name            TEXT NOT NULL,
			description     TEXT DEFAULT '',
			approval_status TEXT DEFAULT 'none'
			                CHECK (approval_status IN ('none','pending','approved','rejected','needs_revision')),
			created_at      TEXT DEFAULT (datetime('now')),
			updated_at      TEXT DEFAULT (datetime('now')),
			UNIQUE(epic_id, name)
		)`,
		`CREATE TABLE tasks (
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
		)`,
		`CREATE TABLE task_dependencies (
			task_id    INTEGER NOT NULL REFERENCES tasks(id),
			depends_on INTEGER NOT NULL REFERENCES tasks(id),
			PRIMARY KEY (task_id, depends_on),
			CHECK (task_id != depends_on)
		)`,
		`CREATE TABLE task_results (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id        INTEGER UNIQUE NOT NULL REFERENCES tasks(id),
			summary        TEXT NOT NULL,
			logs           TEXT DEFAULT '',
			skills_used    TEXT DEFAULT '',
			files_modified TEXT DEFAULT '',
			created_at     TEXT DEFAULT (datetime('now'))
		)`,
		// QA validations
		`CREATE TABLE validations (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			story_id    INTEGER NOT NULL REFERENCES stories(id),
			task_id     INTEGER REFERENCES tasks(id),
			verdict     TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
			created_at  TEXT DEFAULT (datetime('now'))
		)`,
		// Codebase knowledge
		`CREATE TABLE knowledge (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			domain       TEXT NOT NULL,
			category     TEXT NOT NULL
			             CHECK (category IN ('pattern','convention','constraint','dependency','test','architecture')),
			finding      TEXT NOT NULL,
			source_files TEXT DEFAULT '',
			confidence   TEXT DEFAULT 'medium'
			             CHECK (confidence IN ('high','medium','low')),
			created_at   TEXT DEFAULT (datetime('now'))
		)`,
		// FTS5 index for knowledge
		`CREATE VIRTUAL TABLE knowledge_fts USING fts5(
			domain, category, finding, source_files,
			content=knowledge, content_rowid=id
		)`,
		`CREATE TRIGGER knowledge_ins AFTER INSERT ON knowledge BEGIN
			INSERT INTO knowledge_fts(rowid, domain, category, finding, source_files)
			VALUES (new.id, new.domain, new.category, new.finding, new.source_files);
		END`,
		`CREATE TRIGGER knowledge_del AFTER DELETE ON knowledge BEGIN
			INSERT INTO knowledge_fts(knowledge_fts, rowid, domain, category, finding, source_files)
			VALUES ('delete', old.id, old.domain, old.category, old.finding, old.source_files);
		END`,
		// HITL approval gates
		`CREATE TABLE approvals (
			id               TEXT PRIMARY KEY,
			type             TEXT NOT NULL CHECK (type IN ('epic-requirements','story-definition',
			                                               'qa-escalation','execution-plan')),
			epic_id          INTEGER REFERENCES epics(id),
			story_id         INTEGER REFERENCES stories(id),
			content          TEXT NOT NULL,
			previous_version TEXT DEFAULT '',
			status           TEXT DEFAULT 'pending'
			                 CHECK (status IN ('pending','approved','rejected','needs_revision')),
			verdict_comment  TEXT DEFAULT '',
			requested_at     TEXT DEFAULT (datetime('now')),
			decided_at       TEXT
		)`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration v1: %w\nSQL: %s", err, stmt)
		}
	}
	return nil
}

// migrate runs all pending migrations inside transactions.
// Schema version is tracked via PRAGMA user_version.
func (d *DB) migrate() error {
	var current int
	if err := d.conn.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	for i := current; i < len(migrations); i++ {
		tx, err := d.conn.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", i+1, err)
		}
		if err := migrations[i](tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		// PRAGMA cannot run inside a transaction, so we set it after commit
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", i+1, err)
		}
		if _, err := d.conn.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			return fmt.Errorf("set schema version %d: %w", i+1, err)
		}
	}
	return nil
}
```

**Step 6: Run tests to verify they pass**

Run: `go test ./src/db/ -run TestOpen -v`
Expected: ALL PASS

**Step 7: Commit**

```bash
git add src/db/db.go src/db/migrations.go src/db/db_test.go src/db/testutil_test.go
git commit -m "feat(db): add SQLite connection, bootstrap, and full schema migration"
```

---

### Task 6: DB Layer — EpicStore

**Files:**
- Create: `src/db/epic_store.go`
- Create: `src/db/epic_store_test.go`

**Step 1: Write the failing tests**

```go
// src/db/epic_store_test.go
package db

import (
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func TestEpicStore_Create(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	epic, created, err := store.Create("my-epic", "a description")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if epic.Name != "my-epic" {
		t.Errorf("Name = %q, want %q", epic.Name, "my-epic")
	}
	if epic.Description != "a description" {
		t.Errorf("Description = %q, want %q", epic.Description, "a description")
	}
}

func TestEpicStore_CreateIdempotent(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	e1, _, _ := store.Create("my-epic", "desc")
	e2, created, err := store.Create("my-epic", "desc")
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if e2.ID != e1.ID {
		t.Errorf("IDs differ: %d vs %d", e2.ID, e1.ID)
	}
}

func TestEpicStore_GetByName(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("test-epic", "desc")
	epic, err := store.GetByName("test-epic")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if epic.Name != "test-epic" {
		t.Errorf("Name = %q, want %q", epic.Name, "test-epic")
	}
}

func TestEpicStore_GetByName_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	_, err := store.GetByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent epic")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_List(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("alpha", "")
	store.Create("beta", "")

	epics, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(epics) != 2 {
		t.Fatalf("len = %d, want 2", len(epics))
	}
}

func TestEpicStore_Update(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("my-epic", "old")
	epic, err := store.Update("my-epic", "new description")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if epic.Description != "new description" {
		t.Errorf("Description = %q, want %q", epic.Description, "new description")
	}
}

func TestEpicStore_Delete(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("doomed", "")
	if err := store.Delete("doomed"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.GetByName("doomed")
	if !isNotFound(err) {
		t.Error("expected ErrNotFound after delete")
	}
}

// isNotFound checks if err wraps domain.ErrNotFound.
func isNotFound(err error) bool {
	return err != nil && errors.Is(err, domain.ErrNotFound)
}
```

Note: add `"errors"` to imports.

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestEpicStore -v`
Expected: FAIL — EpicStore not defined

**Step 3: Write epic_store.go**

```go
// src/db/epic_store.go
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas/oraculo/src/domain"
)

// EpicStore handles epic persistence.
type EpicStore struct {
	db *DB
}

// NewEpicStore returns an EpicStore backed by the given database.
func NewEpicStore(db *DB) *EpicStore {
	return &EpicStore{db: db}
}

// Create inserts a new epic. If one with the same name exists, returns
// it with created=false.
func (s *EpicStore) Create(name, description string) (*domain.Epic, bool, error) {
	res, err := s.db.conn.Exec(
		"INSERT OR IGNORE INTO epics (name, description) VALUES (?, ?)",
		name, description,
	)
	if err != nil {
		return nil, false, fmt.Errorf("create epic %q: %w", name, err)
	}
	rows, _ := res.RowsAffected()
	epic, err := s.GetByName(name)
	if err != nil {
		return nil, false, err
	}
	return epic, rows > 0, nil
}

// GetByName returns an epic by name.
func (s *EpicStore) GetByName(name string) (*domain.Epic, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, name, description, approval_status, created_at, updated_at FROM epics WHERE name = ?",
		name,
	)
	return scanEpic(row, name)
}

// List returns all epics ordered by creation time.
func (s *EpicStore) List() ([]domain.Epic, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, name, description, approval_status, created_at, updated_at FROM epics ORDER BY created_at",
	)
	if err != nil {
		return nil, fmt.Errorf("list epics: %w", err)
	}
	defer rows.Close()
	var epics []domain.Epic
	for rows.Next() {
		var e domain.Epic
		var createdAt, updatedAt string
		var status string
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &status, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan epic: %w", err)
		}
		e.ApprovalStatus = domain.ApprovalStatus(status)
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		e.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		epics = append(epics, e)
	}
	return epics, rows.Err()
}

// Update changes an epic's description.
func (s *EpicStore) Update(name, description string) (*domain.Epic, error) {
	res, err := s.db.conn.Exec(
		"UPDATE epics SET description = ?, updated_at = datetime('now') WHERE name = ?",
		description, name,
	)
	if err != nil {
		return nil, fmt.Errorf("update epic %q: %w", name, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	return s.GetByName(name)
}

// Delete removes an epic by name.
func (s *EpicStore) Delete(name string) error {
	res, err := s.db.conn.Exec("DELETE FROM epics WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete epic %q: %w", name, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	return nil
}

// UpdateApprovalStatus sets the approval_status of an epic.
func (s *EpicStore) UpdateApprovalStatus(name string, status domain.ApprovalStatus) error {
	res, err := s.db.conn.Exec(
		"UPDATE epics SET approval_status = ?, updated_at = datetime('now') WHERE name = ?",
		string(status), name,
	)
	if err != nil {
		return fmt.Errorf("update approval status for epic %q: %w", name, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	return nil
}

func scanEpic(row *sql.Row, name string) (*domain.Epic, error) {
	var e domain.Epic
	var createdAt, updatedAt string
	var status string
	err := row.Scan(&e.ID, &e.Name, &e.Description, &status, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("scan epic %q: %w", name, err)
	}
	e.ApprovalStatus = domain.ApprovalStatus(status)
	e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	e.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &e, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./src/db/ -run TestEpicStore -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add src/db/epic_store.go src/db/epic_store_test.go
git commit -m "feat(db): add EpicStore with CRUD and idempotent create"
```

---

### Task 7: DB Layer — StoryStore

**Files:**
- Create: `src/db/story_store.go`
- Create: `src/db/story_store_test.go`

**Step 1: Write the failing tests**

Tests should cover: Create (scoped to epic), idempotent create, GetByName (requires epic), List (scoped to epic), Update, Delete, not found errors. Follow the same pattern as EpicStore tests but using `--epic` scoping. Each story is identified by (epicID, name).

Key test cases:
- `Create` with valid epicID succeeds
- `Create` same name under same epic returns `created=false`
- `Create` same name under different epic succeeds (names are unique per epic, not globally)
- `GetByName` with wrong epicID returns not found
- `List` only returns stories for the given epicID
- `Delete` cascades correctly (or refuses if tasks exist)

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestStoryStore -v`
Expected: FAIL

**Step 3: Write story_store.go**

Same structure as EpicStore but all queries are scoped by `epic_id`. Methods:
- `Create(epicID int, name, description string) (*domain.Story, bool, error)`
- `GetByName(epicID int, name string) (*domain.Story, error)`
- `List(epicID int) ([]domain.Story, error)`
- `Update(epicID int, name, description string) (*domain.Story, error)`
- `Delete(epicID int, name string) error`
- `UpdateApprovalStatus(epicID int, name string, status domain.ApprovalStatus) error`

**Step 4: Run tests to verify they pass**

Run: `go test ./src/db/ -run TestStoryStore -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add src/db/story_store.go src/db/story_store_test.go
git commit -m "feat(db): add StoryStore with epic-scoped CRUD"
```

---

### Task 8: DB Layer — TaskStore with DAG Validation

**Files:**
- Create: `src/db/task_store.go`
- Create: `src/db/task_store_test.go`

This is the most complex store. It handles:
- CRUD scoped to storyID
- Status lifecycle transitions (using `domain.TaskStatus.CanTransitionTo`)
- DAG dependency creation with cycle detection
- Task completion with TaskResult insertion

**Step 1: Write the failing tests**

Key test cases:
- `Create` inserts a task with status "pending"
- `Create` idempotent (same storyID + name)
- `Create` with `dependsOn` names that resolve to task IDs in the same story
- `Create` with cyclic dependency returns `domain.ErrCyclicDependency`
- `Start` transitions pending → in_progress, sets `started_at`
- `Start` on non-pending task returns `domain.ErrInvalidTransition`
- `Complete` transitions in_progress → completed, inserts TaskResult, sets `completed_at`
- `Fail` transitions in_progress → failed with reason
- `GetByName` returns task with current status
- `List` returns all tasks for a story with their statuses

For cycle detection, test: A depends on B, B depends on C, C depends on A → error on C.

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestTaskStore -v`
Expected: FAIL

**Step 3: Write task_store.go**

Methods:
- `Create(storyID int, name, description string, dependsOn []string) (*domain.Task, bool, error)` — resolves dependsOn names to IDs within the same story, checks for cycles via DFS, inserts task and dependencies
- `Start(storyID int, name string) (*domain.Task, error)` — validates transition, updates status and started_at
- `Complete(storyID int, name string, result domain.TaskResult) (*domain.Task, error)` — validates transition, updates status and completed_at, inserts task_results row
- `Fail(storyID int, name, reason string) (*domain.Task, error)` — validates transition, updates status and failure_reason
- `GetByName(storyID int, name string) (*domain.Task, error)`
- `List(storyID int) ([]domain.Task, error)`
- `Delete(storyID int, name string) error`

Cycle detection: BFS/DFS from the new dependency backward through existing edges to check if the target task is reachable from the depends_on task.

**Step 4: Run tests to verify they pass**

Run: `go test ./src/db/ -run TestTaskStore -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add src/db/task_store.go src/db/task_store_test.go
git commit -m "feat(db): add TaskStore with lifecycle transitions and DAG cycle detection"
```

---

### Task 9: DB Layer — MemoryStore with FTS5

**Files:**
- Create: `src/db/memory_store.go`
- Create: `src/db/memory_store_test.go`

**Step 1: Write the failing tests**

Key test cases:
- `Store` inserts a knowledge entry and it's retrievable
- `Search` finds entries by keyword using FTS5
- `Search` with domain filter limits results
- `Search` with limit caps results
- `Search` returns empty for no match
- `Domains` returns distinct domain values

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestMemoryStore -v`
Expected: FAIL

**Step 3: Write memory_store.go**

Methods:
- `Store(domain string, category domain.Category, finding, sourceFiles string, confidence domain.Confidence) (*domain.Knowledge, error)`
- `Search(query, domainFilter string, limit int) ([]domain.Knowledge, error)` — uses `SELECT ... FROM knowledge WHERE id IN (SELECT rowid FROM knowledge_fts WHERE knowledge_fts MATCH ?)`
- `Domains() ([]string, error)` — `SELECT DISTINCT domain FROM knowledge ORDER BY domain`

**Step 4: Run tests to verify they pass**

Run: `go test ./src/db/ -run TestMemoryStore -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add src/db/memory_store.go src/db/memory_store_test.go
git commit -m "feat(db): add MemoryStore with FTS5 search"
```

---

### Task 10: DB Layer — ApprovalStore

**Files:**
- Create: `src/db/approval_store.go`
- Create: `src/db/approval_store_test.go`

**Step 1: Write the failing tests**

Key test cases:
- `Request` creates an approval with UUID, status "pending"
- `GetByID` returns the approval
- `GetByID` not found returns error
- `List` returns all approvals
- `List` with pendingOnly=true only returns pending
- `Verdict` updates status, sets verdict_comment and decided_at
- `Verdict` on non-pending returns error
- `Verdict` with "needs_revision" stores previous_version

**Step 2: Run test to verify it fails**

Run: `go test ./src/db/ -run TestApprovalStore -v`
Expected: FAIL

**Step 3: Write approval_store.go**

Methods:
- `Request(approvalType domain.ApprovalType, epicID, storyID *int, content string) (*domain.Approval, error)` — generates UUID, inserts with status pending
- `GetByID(id string) (*domain.Approval, error)`
- `List(pendingOnly bool) ([]domain.Approval, error)`
- `Verdict(id string, verdict domain.Verdict, comment string) (*domain.Approval, error)` — validates current status is pending, updates status/comment/decided_at. If verdict is needs_revision, copies current content to previous_version.

Use `crypto/rand` or `github.com/google/uuid` for UUID generation. Prefer a simple implementation with `crypto/rand` + hex encoding to avoid a new dependency:

```go
func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./src/db/ -run TestApprovalStore -v`
Expected: ALL PASS

**Step 5: Run all DB tests**

Run: `go test ./src/db/ -v`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add src/db/approval_store.go src/db/approval_store_test.go
git commit -m "feat(db): add ApprovalStore with request/verdict workflow"
```

---

### Task 11: Output Layer

**Files:**
- Create: `src/output/json.go`
- Create: `src/output/table.go`
- Create: `src/output/json_test.go`
- Create: `src/output/table_test.go`

**Step 1: Write the failing tests for JSON output**

```go
// src/output/json_test.go
package output

import (
	"bytes"
	"errors"
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	err := WriteJSON(&buf, map[string]any{"name": "test", "created": true})
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	got := buf.String()
	if got == "" {
		t.Error("empty output")
	}
	// Should be valid JSON with newline
	if got[len(got)-1] != '\n' {
		t.Error("missing trailing newline")
	}
}

func TestWriteError_NotFound(t *testing.T) {
	var buf bytes.Buffer
	err := fmt.Errorf("epic %q: %w", "test", domain.ErrNotFound)
	WriteError(&buf, err)
	got := buf.String()
	if !strings.Contains(got, `"error"`) {
		t.Errorf("missing error key in: %s", got)
	}
	if !strings.Contains(got, "not_found") {
		t.Errorf("expected not_found code in: %s", got)
	}
}
```

Note: add `"fmt"` and `"strings"` to imports.

**Step 2: Run test to verify it fails**

Run: `go test ./src/output/ -v`
Expected: FAIL

**Step 3: Write json.go**

```go
// src/output/json.go
package output

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/lucas/oraculo/src/domain"
)

// WriteJSON marshals v to indented JSON and writes it to w.
func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteError maps a domain error to the documented JSON error format and writes to w.
func WriteError(w io.Writer, err error) {
	code := "unknown_error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		code = "not_found"
	case errors.Is(err, domain.ErrAlreadyExists):
		code = "already_exists"
	case errors.Is(err, domain.ErrInvalidTransition):
		code = "invalid_transition"
	case errors.Is(err, domain.ErrCyclicDependency):
		code = "cyclic_dependency"
	case errors.Is(err, domain.ErrMissingRequired):
		code = "missing_required"
	}
	WriteJSON(w, map[string]string{
		"error":   code,
		"message": err.Error(),
	})
}
```

**Step 4: Write the failing test for table output**

```go
// src/output/table_test.go
package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteTable(t *testing.T) {
	var buf bytes.Buffer
	err := WriteTable(&buf, []string{"Name", "Status"}, [][]string{
		{"alpha", "pending"},
		{"beta", "completed"},
	})
	if err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "Name") {
		t.Error("missing header")
	}
	if !strings.Contains(got, "alpha") {
		t.Error("missing row data")
	}
}
```

**Step 5: Write table.go**

```go
// src/output/table.go
package output

import (
	"fmt"
	"io"
	"strings"
)

// WriteTable renders a human-readable aligned table.
func WriteTable(w io.Writer, headers []string, rows [][]string) error {
	if len(headers) == 0 {
		return nil
	}
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	// Print header
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%-*s", widths[i], h)
	}
	fmt.Fprintln(w)
	// Print separator
	for i, width := range widths {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprint(w, strings.Repeat("─", width))
	}
	fmt.Fprintln(w)
	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			if i < len(widths) {
				fmt.Fprintf(w, "%-*s", widths[i], cell)
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}
```

**Step 6: Run all output tests**

Run: `go test ./src/output/ -v`
Expected: ALL PASS

**Step 7: Commit**

```bash
git add src/output/
git commit -m "feat(output): add JSON and table output formatters"
```

---

### Task 12: CLI Layer — Root, Version, Install Stub, Tools Bootstrap

**Files:**
- Create: `src/cli/root.go`
- Create: `src/cli/version.go`
- Create: `src/cli/install.go`
- Create: `src/cli/tools/tools.go`
- Create: `src/cli/tools/context.go`
- Update: `cmd/oraculo/main.go`

**Step 1: Write root.go — root command assembly**

```go
// src/cli/root.go
package cli

import (
	"github.com/lucas/oraculo/src/cli/tools"
	"github.com/spf13/cobra"
)

// NewRoot returns the root cobra.Command for oraculo.
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "oraculo",
		Short: "Socratic guide for quality product development",
		Long:  "Oraculo is a Socratic guide and team orchestrator for quality product development.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(
		newVersionCmd(version),
		newInstallCmd(),
		tools.NewToolsCmd(),
	)
	return root
}
```

**Step 2: Write version.go**

```go
// src/cli/version.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}
```

**Step 3: Write install.go (stub)**

```go
// src/cli/install.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Oraculo skills and hooks into Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			global, _ := cmd.Flags().GetBool("global")
			scope := "local (.claude/)"
			if global {
				scope = "global (~/.claude/)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "oraculo install: %s (not yet implemented)\n", scope)
			return nil
		},
	}
	cmd.Flags().Bool("global", false, "Install globally for all projects")
	cmd.Flags().Bool("local", true, "Install locally for current project")
	return cmd
}
```

**Step 4: Write tools/context.go — DB context helpers**

```go
// src/cli/tools/context.go
package tools

import (
	"context"

	"github.com/lucas/oraculo/src/db"
)

type contextKey string

const dbKey contextKey = "db"

func withDB(ctx context.Context, database *db.DB) context.Context {
	return context.WithValue(ctx, dbKey, database)
}

func dbFromContext(ctx context.Context) *db.DB {
	return ctx.Value(dbKey).(*db.DB)
}
```

**Step 5: Write tools/tools.go — parent command with bootstrap**

```go
// src/cli/tools/tools.go
package tools

import (
	"github.com/lucas/oraculo/src/db"
	"github.com/spf13/cobra"
)

// NewToolsCmd returns the "tools" parent command.
// PersistentPreRunE handles auto-bootstrap: calls db.Open() and stores
// the DB in the command context for all subcommands.
func NewToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Agent-facing commands (JSON output)",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			database, err := db.Open()
			if err != nil {
				return err
			}
			cmd.SetContext(withDB(cmd.Context(), database))
			return nil
		},
	}
	return cmd
}
```

**Step 6: Update main.go to wire the root command**

```go
// cmd/oraculo/main.go
package main

import (
	"fmt"
	"os"

	"github.com/lucas/oraculo/src/cli"
)

var version = "dev"

func main() {
	cmd := cli.NewRoot(version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 7: Verify it compiles and version works**

Run: `go build -o /tmp/oraculo ./cmd/oraculo && /tmp/oraculo version`
Expected: prints "dev"

Run: `/tmp/oraculo --help`
Expected: shows help with install, version, tools subcommands

**Step 8: Commit**

```bash
git add cmd/oraculo/main.go src/cli/
git commit -m "feat(cli): add root command, version, install stub, and tools bootstrap"
```

---

### Task 13: CLI Tools — Epic Commands + E2E Tests

**Files:**
- Create: `src/cli/tools/epic.go`
- Create: `src/cli/tools/epic_test.go`

**Step 1: Write the failing E2E tests**

Tests execute the full Cobra command tree against a temp directory. Each test:
1. Creates a temp dir, chdirs into it (or configures the DB path)
2. Builds a root command
3. Executes with `SetArgs()`
4. Asserts JSON stdout and filesystem effects

Key test cases:
- `tools epic init my-epic --description "test"` → JSON with `created: true`
- `tools epic init my-epic` again → JSON with `created: false`
- `tools epic get my-epic` after save → raw markdown stdout
- `tools epic list` → JSON array
- `tools epic update my-epic --description "new"` → JSON with updated description
- `tools epic delete my-epic` → JSON success
- `tools epic get nonexistent` → JSON error with exit code

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/tools/ -run TestEpic -v`
Expected: FAIL

**Step 3: Write epic.go — all 6 epic subcommands**

Implement `init`, `save`, `get`, `list`, `update`, `delete`. Each:
1. Gets DB from context
2. Creates EpicStore
3. Calls the appropriate method
4. Formats output via `output.WriteJSON` or `output.WriteError`

`save` reads stdin until EOF, creates `.oraculo/epics/<name>/` directory, writes `requirements.md`.
`get` reads `.oraculo/epics/<name>/requirements.md` and writes raw content to stdout.

**Step 4: Run tests to verify they pass**

Run: `go test ./src/cli/tools/ -run TestEpic -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add src/cli/tools/epic.go src/cli/tools/epic_test.go
git commit -m "feat(cli): add tools epic commands (init/save/get/list/update/delete)"
```

---

### Task 14: CLI Tools — Story Commands + E2E Tests

**Files:**
- Create: `src/cli/tools/story.go`
- Create: `src/cli/tools/story_test.go`

Same structure as Task 13 but for story commands. All commands require `--epic` flag. Tests must create an epic first, then operate on stories within it.

Key difference: `story init` must resolve the epic name to an epic ID before calling StoryStore.

**Step 1: Write the failing E2E tests**

Key test cases:
- `tools story init login --epic my-epic` → JSON with `created: true`
- `tools story init login --epic nonexistent` → JSON error (epic not found)
- `tools story list --epic my-epic` → JSON array of stories
- `tools story save login --epic my-epic` with stdin → creates markdown file
- `tools story get login --epic my-epic` → raw markdown

**Step 2-5: Same pattern as Task 13**

Commit: `git commit -m "feat(cli): add tools story commands (init/save/get/list/update/delete)"`

---

### Task 15: CLI Tools — Task Commands + E2E Tests

**Files:**
- Create: `src/cli/tools/task.go`
- Create: `src/cli/tools/task_test.go`

Most complex CLI task. All commands require `--epic` and `--story` flags. Must resolve epic name → epic ID → story name → story ID before operating.

Commands: `init`, `start`, `complete`, `fail`, `get`, `list`, `delete`.

- `init` supports `--depends-on` flag (can be specified multiple times)
- `complete` reads JSON from stdin for TaskResult
- `fail` requires `--reason` flag

**Step 1: Write the failing E2E tests**

Key test cases:
- Full lifecycle: init → start → complete with result JSON on stdin
- Full lifecycle: init → start → fail with --reason
- init with --depends-on creates DAG edge
- init with circular dependency returns error
- start on non-pending returns error
- complete on non-in_progress returns error
- list shows all tasks with statuses

**Step 2-5: Same pattern as Task 13**

Commit: `git commit -m "feat(cli): add tools task commands with lifecycle and DAG support"`

---

### Task 16: CLI Tools — Memory Commands + E2E Tests

**Files:**
- Create: `src/cli/tools/memory.go`
- Create: `src/cli/tools/memory_test.go`

Commands: `store`, `search`, `domains`.

- `store` requires `--domain`, `--category`, `--finding`. Optional: `--source`, `--confidence`
- `search` takes a positional query arg. Optional: `--domain`, `--limit`
- `domains` returns distinct domain values

**Step 1: Write the failing E2E tests**

Key test cases:
- `tools memory store --domain payments --category pattern --finding "Uses repository pattern"` → JSON
- `tools memory search "repository"` → JSON array containing the stored finding
- `tools memory search "repository" --domain payments` → filtered results
- `tools memory domains` → JSON array with "payments"
- `tools memory store` with missing required flags → error

**Step 2-5: Same pattern as Task 13**

Commit: `git commit -m "feat(cli): add tools memory commands (store/search/domains)"`

---

### Task 17: CLI Tools — Approval Commands + E2E Tests

**Files:**
- Create: `src/cli/tools/approval.go`
- Create: `src/cli/tools/approval_test.go`

Commands: `request`, `status`, `list`, `verdict`.

- `request` requires `--type`, `--epic`. Optional: `--story`. Reads content from stdin.
- `status` takes approval ID as positional arg
- `list` has optional `--pending` flag
- `verdict` takes ID as positional arg, requires `--verdict`. Optional: `--comment`

**Step 1: Write the failing E2E tests**

Key test cases:
- Full workflow: request → status (pending) → verdict (approved) → status (approved)
- request with all approval types
- verdict with needs_revision stores previous version
- list --pending filters correctly
- verdict on non-pending returns error

**Step 2-5: Same pattern as Task 13**

Commit: `git commit -m "feat(cli): add tools approval commands (request/status/list/verdict)"`

---

### Task 18: CLI — Status Command

**Files:**
- Create: `src/cli/status.go`
- Create: `src/cli/status_test.go`

The `status` command opens the DB and renders a human-readable dashboard showing:
- Epic count with approval statuses
- Story count per epic
- Task progress (pending/in_progress/completed/failed counts and percentages)
- Pending approval count

Unlike tools commands, this outputs formatted text using `output.WriteTable`.

Note: the `status` command also needs to call `db.Open()`, but it's not under `tools/`. Add its own `PreRunE` that opens the DB, or share a helper.

**Step 1: Write the failing test**

Test creates epics, stories, tasks in various states, then runs `oraculo status` and checks the formatted output contains expected counts.

**Step 2-5: Same pattern**

Commit: `git commit -m "feat(cli): add status command with human-readable dashboard"`

---

### Task 19: Register All Subcommands

**Files:**
- Modify: `src/cli/tools/tools.go` — register epic, story, task, memory, approval subcommands
- Modify: `src/cli/root.go` — register status command

**Step 1: Update tools.go to add all subcommands**

```go
cmd.AddCommand(
    newEpicCmd(),
    newStoryCmd(),
    newTaskCmd(),
    newMemoryCmd(),
    newApprovalCmd(),
)
```

**Step 2: Update root.go to add status**

```go
root.AddCommand(
    newVersionCmd(version),
    newInstallCmd(),
    newStatusCmd(),
    tools.NewToolsCmd(),
)
```

**Step 3: Build and verify help output**

Run: `go build -o /tmp/oraculo ./cmd/oraculo && /tmp/oraculo tools --help`
Expected: shows epic, story, task, memory, approval subcommands

Run: `/tmp/oraculo tools epic --help`
Expected: shows init, save, get, list, update, delete subcommands

**Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add src/cli/
git commit -m "feat(cli): register all subcommands and verify full command tree"
```

---

### Task 20: Final Integration and Cleanup

**Files:**
- Verify: `cmd/oraculo/main.go`
- Run: full test suite

**Step 1: Run the full test suite**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 2: Run go vet**

Run: `go vet ./...`
Expected: no issues

**Step 3: Build the binary**

Run: `go build -o /tmp/oraculo ./cmd/oraculo`
Expected: clean build, single binary

**Step 4: Smoke test the binary**

```bash
cd $(mktemp -d)
/tmp/oraculo version
/tmp/oraculo tools epic init test-epic --description "smoke test"
/tmp/oraculo tools epic list
/tmp/oraculo tools epic get test-epic
echo "# Requirements" | /tmp/oraculo tools epic save test-epic
/tmp/oraculo tools epic get test-epic
/tmp/oraculo status
```

Verify: each command produces expected output, `.oraculo/` directory was created, `oraculo.db` exists.

**Step 5: Commit any fixes**

```bash
git add -A
git commit -m "feat: complete CLI Trust Layer skeleton with all 28 commands"
```
