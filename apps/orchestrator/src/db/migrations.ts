import type { Database } from "bun:sqlite";

type Migration = (db: Database) => void;

const migrations: Migration[] = [
  // V1: Core tables
  (db) => {
    const stmts = [
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
        type           TEXT DEFAULT 'code'
                       CHECK (type IN ('code','research','qa')),
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
        findings    TEXT DEFAULT '',
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
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },

  // V2: Sessions tables
  (db) => {
    const stmts = [
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
        ended_at   TEXT,
        metadata   TEXT DEFAULT '{}'
      )`,
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },

  // V3: Agents and tool_events tables
  (db) => {
    const stmts = [
      `CREATE TABLE agents (
        id          INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id  TEXT NOT NULL,
        name        TEXT NOT NULL,
        type        TEXT NOT NULL DEFAULT 'unknown',
        status      TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'completed', 'failed')),
        started_at  TEXT NOT NULL,
        stopped_at  TEXT,
        task_id     INTEGER REFERENCES tasks(id)
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
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },

  // V4: No-op for fresh DBs (was rebuild of approvals CHECK constraint)
  (db) => {
    const stmts = [
      `CREATE TABLE approvals_new (
        id               TEXT PRIMARY KEY,
        type             TEXT NOT NULL CHECK (type IN ('epic-requirements','story-definition',
                                                       'qa-escalation','execution-plan','design')),
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
      `INSERT INTO approvals_new SELECT * FROM approvals`,
      `DROP TABLE approvals`,
      `ALTER TABLE approvals_new RENAME TO approvals`,
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },

  // V5: epic_versions, story_versions, reviews tables
  (db) => {
    const stmts = [
      `CREATE TABLE epic_versions (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        epic_id    INTEGER NOT NULL REFERENCES epics(id),
        number     INTEGER NOT NULL,
        content    TEXT NOT NULL,
        created_at TEXT DEFAULT (datetime('now')),
        UNIQUE(epic_id, number)
      )`,
      `CREATE TABLE story_versions (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        story_id   INTEGER NOT NULL REFERENCES stories(id),
        number     INTEGER NOT NULL,
        content    TEXT NOT NULL,
        created_at TEXT DEFAULT (datetime('now')),
        UNIQUE(story_id, number)
      )`,
      `CREATE TABLE reviews (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        version_id   INTEGER NOT NULL,
        version_type TEXT NOT NULL CHECK (version_type IN ('epic','story')),
        verdict      TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
        comment      TEXT DEFAULT '',
        created_at   TEXT DEFAULT (datetime('now'))
      )`,
      `CREATE INDEX idx_reviews_version ON reviews(version_id, version_type)`,
      `CREATE TABLE approvals_new (
        id              TEXT PRIMARY KEY,
        type            TEXT NOT NULL CHECK (type IN ('qa-escalation','execution-plan','design')),
        epic_id         INTEGER REFERENCES epics(id),
        story_id        INTEGER REFERENCES stories(id),
        content         TEXT NOT NULL,
        previous_version TEXT DEFAULT '',
        status          TEXT DEFAULT 'pending'
                        CHECK (status IN ('pending','approved','rejected','needs_revision')),
        verdict_comment TEXT DEFAULT '',
        requested_at    TEXT DEFAULT (datetime('now')),
        decided_at      TEXT
      )`,
      `INSERT INTO approvals_new SELECT * FROM approvals
        WHERE type IN ('qa-escalation','execution-plan','design')`,
      `DROP TABLE approvals`,
      `ALTER TABLE approvals_new RENAME TO approvals`,
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },

  // V6: session_events table (ended_at already in V2 for fresh DBs)
  (db) => {
    try {
      db.run(`ALTER TABLE claude_sessions ADD COLUMN ended_at TEXT`);
    } catch {
      // Column already exists for fresh DBs
    }
    const stmts = [
      `CREATE TABLE session_events (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id TEXT NOT NULL REFERENCES claude_sessions(id),
        event_type TEXT NOT NULL
                   CHECK (event_type IN ('task_completed','stop','teammate_idle','session_end')),
        payload    TEXT DEFAULT '{}',
        created_at TEXT DEFAULT (datetime('now'))
      )`,
      `CREATE INDEX idx_session_events_session ON session_events(session_id)`,
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },

  // V7: validations.findings column (already in V1 SQL for fresh DBs)
  (db) => {
    try {
      db.run(`ALTER TABLE validations ADD COLUMN findings TEXT DEFAULT ''`);
    } catch {
      // Column already exists for fresh DBs
    }
  },

  // V8: agents.task_id column (already in V3 for fresh DBs)
  (db) => {
    try {
      db.run(`ALTER TABLE agents ADD COLUMN task_id INTEGER REFERENCES tasks(id)`);
    } catch {
      // Column already exists for fresh DBs
    }
  },

  // V9: approval_comments table
  (db) => {
    const stmts = [
      `CREATE TABLE approval_comments (
        id            INTEGER PRIMARY KEY AUTOINCREMENT,
        approval_id   TEXT NOT NULL REFERENCES approvals(id),
        selected_text TEXT NOT NULL,
        comment       TEXT NOT NULL,
        created_at    TEXT DEFAULT (datetime('now'))
      )`,
      `CREATE INDEX idx_approval_comments_approval ON approval_comments(approval_id)`,
    ];
    for (const stmt of stmts) {
      db.run(stmt);
    }
  },
];

export function migrate(db: Database): void {
  const result = db.query("PRAGMA user_version").get() as { user_version: number };
  let current = result.user_version;

  for (let i = current; i < migrations.length; i++) {
    db.run("BEGIN");
    try {
      migrations[i](db);
      db.run("COMMIT");
    } catch (err) {
      db.run("ROLLBACK");
      throw new Error(`migration ${i + 1}: ${err}`);
    }
    db.run(`PRAGMA user_version = ${i + 1}`);
  }
}
