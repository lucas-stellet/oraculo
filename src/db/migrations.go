package db

import (
	"database/sql"
	"fmt"
)

var migrations = []func(*sql.Tx) error{
	migrateV1,
	migrateV2,
	migrateV3,
}

// migrateV1 creates all core tables, knowledge with FTS5, approvals, and validations.
func migrateV1(tx *sql.Tx) error {
	stmts := []string{
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
		`CREATE TABLE validations (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			story_id    INTEGER NOT NULL REFERENCES stories(id),
			task_id     INTEGER REFERENCES tasks(id),
			verdict     TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
			created_at  TEXT DEFAULT (datetime('now'))
		)`,
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

func migrateV3(tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE agents (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id  TEXT NOT NULL,
			name        TEXT NOT NULL,
			type        TEXT NOT NULL DEFAULT 'unknown',
			status      TEXT NOT NULL DEFAULT 'active'
			            CHECK (status IN ('active', 'completed', 'failed')),
			started_at  TEXT NOT NULL,
			stopped_at  TEXT
		)`,
		`CREATE INDEX idx_agents_session_id ON agents(session_id)`,
		`CREATE TABLE tool_events (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id  TEXT NOT NULL,
			tool_name   TEXT NOT NULL,
			file_path   TEXT,
			timestamp   TEXT NOT NULL
		)`,
		`CREATE INDEX idx_tool_events_session_id ON tool_events(session_id)`,
		`CREATE INDEX idx_tool_events_timestamp ON tool_events(timestamp)`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migration v3: %w\nSQL: %s", err, stmt)
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
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", i+1, err)
		}
		if _, err := d.conn.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			return fmt.Errorf("set schema version %d: %w", i+1, err)
		}
	}
	return nil
}
