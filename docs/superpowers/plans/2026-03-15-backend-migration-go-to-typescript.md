# Backend Migration: Go → TypeScript/Bun Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Go backend with a TypeScript/Bun application that replicates all functionality and adds Claude Agent SDK orchestration.

**Architecture:** Single TypeScript package (`apps/orchestrator/`) using bun:sqlite for data, Hono on Bun.serve() for HTTP/WS/SSE, commander for CLI, and @anthropic-ai/claude-agent-sdk for agent dispatch. Compiles to standalone binary via `bun build --compile`.

**Tech Stack:** Bun, TypeScript, bun:sqlite, Hono, commander, pino, zod, @modelcontextprotocol/sdk, @anthropic-ai/claude-agent-sdk

**Spec:** `docs/superpowers/specs/2026-03-15-backend-migration-go-to-typescript-design.md`

---

## Chunk 1: Foundation — Project Setup, Domain, Database

### Task 1: Project Setup and Dependencies

**Files:**
- Modify: `apps/orchestrator/package.json`
- Modify: `apps/orchestrator/tsconfig.json`
- Create: `apps/orchestrator/src/index.ts`

- [ ] **Step 1: Install all dependencies**

```bash
cd apps/orchestrator
bun add hono commander @commander-js/extra-typings @modelcontextprotocol/sdk zod pino pino-roll @anthropic-ai/claude-agent-sdk
bun add -d pino-pretty bun-plugin-pino @types/bun
```

- [ ] **Step 2: Configure tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "types": ["bun-types"]
  },
  "include": ["src/**/*.ts"],
  "exclude": ["node_modules", "dist"]
}
```

- [ ] **Step 3: Create version constant**

```typescript
// src/version.ts
export const VERSION = "0.1.0";
```

- [ ] **Step 4: Create entry point stub**

```typescript
// src/index.ts
#!/usr/bin/env bun
console.log("oraculo");
```

- [ ] **Step 4: Verify setup**

Run: `cd apps/orchestrator && bun run src/index.ts`
Expected: prints "oraculo"

- [ ] **Step 5: Commit**

```bash
git add apps/orchestrator/package.json apps/orchestrator/tsconfig.json apps/orchestrator/src/index.ts apps/orchestrator/bun.lock
git commit -m "chore(orchestrator): initialize project with dependencies"
```

---

### Task 2: Domain — Enums and Type Definitions

**Files:**
- Create: `apps/orchestrator/src/domain/enums.ts`
- Create: `apps/orchestrator/src/domain/entities.ts`
- Test: `apps/orchestrator/src/domain/__tests__/enums.test.ts`

Reference: `apps/backend/src/domain/epic.go`, `story.go`, `task.go`, `session.go`, `phases.go`, `approval.go`, `knowledge.go`, `validation.go`, `review.go`

- [ ] **Step 1: Write the failing test for enums**

```typescript
// src/domain/__tests__/enums.test.ts
import { describe, it, expect } from "bun:test";
import {
  type TaskStatus,
  type ApprovalStatus,
  type ApprovalType,
  type Verdict,
  type SessionType,
  type SessionStatus,
  type Category,
  type Confidence,
  type ReviewVerdict,
  type VersionType,
  TASK_STATUSES,
  APPROVAL_STATUSES,
  APPROVAL_TYPES,
  VERDICTS,
  SESSION_TYPES,
  SESSION_STATUSES,
  CATEGORIES,
  CONFIDENCES,
  SESSION_TYPE_TO_PHASE,
  PHASES,
  phaseIndex,
} from "../enums";

describe("enums", () => {
  it("defines all task statuses", () => {
    expect(TASK_STATUSES).toEqual(["pending", "in_progress", "completed", "failed"]);
  });

  it("defines all approval types", () => {
    expect(APPROVAL_TYPES).toEqual(["qa-escalation", "execution-plan", "design"]);
  });

  it("defines all session types", () => {
    expect(SESSION_TYPES).toEqual(["epic", "story", "plan", "execute", "validate"]);
  });

  it("maps session types to dashboard phases", () => {
    expect(SESSION_TYPE_TO_PHASE.epic).toBe("discover");
    expect(SESSION_TYPE_TO_PHASE.story).toBe("define");
    expect(SESSION_TYPE_TO_PHASE.plan).toBe("plan");
    expect(SESSION_TYPE_TO_PHASE.execute).toBe("execute");
    expect(SESSION_TYPE_TO_PHASE.validate).toBe("validate");
  });

  it("returns correct phase index", () => {
    expect(phaseIndex("epic", "discover")).toBe(0);
    expect(phaseIndex("epic", "define")).toBe(1);
    expect(phaseIndex("execute", "plan")).toBe(0);
    expect(phaseIndex("execute", "execute")).toBe(1);
  });

  it("returns -1 for unknown phase", () => {
    expect(phaseIndex("epic", "nonexistent")).toBe(-1);
  });

  it("defines phases per session type", () => {
    expect(PHASES.epic).toEqual(["discover", "define", "plan", "execute", "validate"]);
    expect(PHASES.story).toEqual(["define", "plan", "execute", "validate"]);
    expect(PHASES.plan).toEqual(["plan", "execute", "validate"]);
    expect(PHASES.execute).toEqual(["execute", "validate"]);
    expect(PHASES.validate).toEqual(["validate"]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/orchestrator && bun test src/domain/__tests__/enums.test.ts`
Expected: FAIL — module not found

- [ ] **Step 3: Implement enums**

```typescript
// src/domain/enums.ts

// --- Task ---
export type TaskStatus = "pending" | "in_progress" | "completed" | "failed";
export const TASK_STATUSES: TaskStatus[] = ["pending", "in_progress", "completed", "failed"];

// --- Approval ---
export type ApprovalStatus = "pending" | "approved" | "rejected" | "needs_revision";
export const APPROVAL_STATUSES: ApprovalStatus[] = ["pending", "approved", "rejected", "needs_revision"];

export type ApprovalType = "qa-escalation" | "execution-plan" | "design";
export const APPROVAL_TYPES: ApprovalType[] = ["qa-escalation", "execution-plan", "design"];

export type Verdict = "approved" | "rejected" | "needs_revision";
export const VERDICTS: Verdict[] = ["approved", "rejected", "needs_revision"];

// --- Session ---
export type SessionType = "epic" | "story" | "plan" | "execute" | "validate";
export const SESSION_TYPES: SessionType[] = ["epic", "story", "plan", "execute", "validate"];

export type SessionStatus = "active" | "completed" | "abandoned";
export const SESSION_STATUSES: SessionStatus[] = ["active", "completed", "abandoned"];

// --- Knowledge ---
export type Category = "pattern" | "convention" | "constraint" | "dependency" | "test" | "architecture";
export const CATEGORIES: Category[] = ["pattern", "convention", "constraint", "dependency", "test", "architecture"];

export type Confidence = "high" | "medium" | "low";
export const CONFIDENCES: Confidence[] = ["high", "medium", "low"];

// --- Review ---
export type ReviewVerdict = "approved" | "rejected";
export const REVIEW_VERDICTS: ReviewVerdict[] = ["approved", "rejected"];

// --- Version ---
export type VersionType = "epic" | "story";
export const VERSION_TYPES: VersionType[] = ["epic", "story"];

// --- Phase mapping ---
export const SESSION_TYPE_TO_PHASE: Record<SessionType, string> = {
  epic: "discover",
  story: "define",
  plan: "plan",
  execute: "execute",
  validate: "validate",
};

export const PHASES: Record<SessionType, string[]> = {
  epic: ["discover", "define", "plan", "execute", "validate"],
  story: ["define", "plan", "execute", "validate"],
  plan: ["plan", "execute", "validate"],
  execute: ["execute", "validate"],
  validate: ["validate"],
};

export function phaseIndex(sessionType: SessionType, phase: string): number {
  return PHASES[sessionType]?.indexOf(phase) ?? -1;
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/orchestrator && bun test src/domain/__tests__/enums.test.ts`
Expected: PASS

- [ ] **Step 5: Create entity interfaces**

Reference the Go structs in `apps/backend/src/domain/` for exact field names and types. Every field in every Go struct must have a corresponding TypeScript field.

```typescript
// src/domain/entities.ts
import type {
  TaskStatus, ApprovalStatus, ApprovalType, Verdict,
  SessionType, SessionStatus, Category, Confidence,
  ReviewVerdict, VersionType,
} from "./enums";

// --- Epic ---
export interface Epic {
  id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
}

export interface EpicSummary extends Epic {
  phase: string | null;
  phase_status: string | null;
  story_count: number;
  task_count: number;
  completed_task_count: number;
}

// --- Story ---
export interface Story {
  id: number;
  epic_id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
}

export interface StorySummary extends Story {
  task_count: number;
  completed_task_count: number;
  failed_task_count: number;
}

// --- Task ---
export interface Task {
  id: number;
  story_id: number;
  name: string;
  description: string;
  status: TaskStatus;
  type: string;
  created_at: string;
  updated_at: string;
}

export interface TaskResult {
  id: number;
  task_id: number;
  summary: string;
  log: string;
  skills_used: string;
  files_modified: string;
  created_at: string;
}

export interface TaskEnriched extends Task {
  depends_on: number[];
  result: TaskResult | null;
}

// --- Knowledge ---
export interface Knowledge {
  id: number;
  domain: string;
  category: Category;
  finding: string;
  source_files: string;
  confidence: Confidence;
  created_at: string;
}

// --- Approval ---
export interface Approval {
  id: string;
  type: ApprovalType;
  epic_id: number | null;
  story_id: number | null;
  task_id: number | null;
  version_id: number | null;
  status: ApprovalStatus;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface ApprovalComment {
  id: string;
  approval_id: string;
  content: string;
  created_at: string;
}

// --- Session ---
export interface Session {
  id: number;
  type: SessionType;
  epic_id: number;
  description: string;
  status: SessionStatus;
  created_at: string;
  updated_at: string;
}

export interface SessionPhase {
  id: number;
  session_id: number;
  name: string;
  data: string;
  completed_at: string;
}

export interface ClaudeSession {
  id: number;
  session_id: string;
  working_directory: string;
  git_branch: string;
  source: string;
  started_at: string;
  ended_at: string | null;
}

// --- Agent ---
export interface Agent {
  id: number;
  session_id: string;
  name: string;
  type: string;
  task_id: number | null;
  started_at: string;
  stopped_at: string | null;
}

// --- Tool Event ---
export interface ToolEvent {
  id: number;
  session_id: string;
  tool_name: string;
  input: string;
  output: string;
  created_at: string;
}

// --- Session Event ---
export interface SessionEvent {
  id: number;
  session_id: string;
  type: string;
  data: string;
  created_at: string;
}

// --- Version ---
export interface EpicVersion {
  id: number;
  epic_id: number;
  version: number;
  content: string;
  created_at: string;
}

export interface StoryVersion {
  id: number;
  story_id: number;
  version: number;
  content: string;
  created_at: string;
}

// --- Review ---
export interface Review {
  id: number;
  version_id: number;
  type: VersionType;
  verdict: ReviewVerdict;
  comment: string;
  created_at: string;
}

// --- Validation ---
export interface Validation {
  id: number;
  story_id: number;
  verdict: string;
  findings: string;
  created_at: string;
}
```

- [ ] **Step 6: Commit**

```bash
git add apps/orchestrator/src/domain/
git commit -m "feat(orchestrator): add domain enums and entity types"
```

---

### Task 3: Domain — Errors and State Machine

**Files:**
- Create: `apps/orchestrator/src/domain/errors.ts`
- Create: `apps/orchestrator/src/domain/state.ts`
- Create: `apps/orchestrator/src/domain/index.ts`
- Test: `apps/orchestrator/src/domain/__tests__/state.test.ts`
- Test: `apps/orchestrator/src/domain/__tests__/errors.test.ts`

- [ ] **Step 1: Write failing test for state machine**

```typescript
// src/domain/__tests__/state.test.ts
import { describe, it, expect } from "bun:test";
import { canTransition } from "../state";

describe("canTransition", () => {
  it("allows pending → in_progress", () => {
    expect(canTransition("pending", "in_progress")).toBe(true);
  });

  it("allows in_progress → completed", () => {
    expect(canTransition("in_progress", "completed")).toBe(true);
  });

  it("allows in_progress → failed", () => {
    expect(canTransition("in_progress", "failed")).toBe(true);
  });

  it("allows failed → pending (retry)", () => {
    expect(canTransition("failed", "pending")).toBe(true);
  });

  it("rejects pending → completed (skip)", () => {
    expect(canTransition("pending", "completed")).toBe(false);
  });

  it("rejects completed → anything", () => {
    expect(canTransition("completed", "pending")).toBe(false);
    expect(canTransition("completed", "in_progress")).toBe(false);
    expect(canTransition("completed", "failed")).toBe(false);
  });

  it("rejects same-state transitions", () => {
    expect(canTransition("pending", "pending")).toBe(false);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/orchestrator && bun test src/domain/__tests__/state.test.ts`
Expected: FAIL

- [ ] **Step 3: Implement state machine**

```typescript
// src/domain/state.ts
import type { TaskStatus } from "./enums";

const VALID_TRANSITIONS: Record<TaskStatus, TaskStatus[]> = {
  pending: ["in_progress"],
  in_progress: ["completed", "failed"],
  completed: [],
  failed: ["pending"],
};

export function canTransition(from: TaskStatus, to: TaskStatus): boolean {
  return VALID_TRANSITIONS[from]?.includes(to) ?? false;
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/orchestrator && bun test src/domain/__tests__/state.test.ts`
Expected: PASS

- [ ] **Step 5: Write failing test for errors**

```typescript
// src/domain/__tests__/errors.test.ts
import { describe, it, expect } from "bun:test";
import { DomainError, ErrNotFound, ErrCyclicDependency, ErrInvalidTransition } from "../errors";

describe("DomainError", () => {
  it("creates error with code", () => {
    const err = ErrNotFound("task", "abc");
    expect(err).toBeInstanceOf(DomainError);
    expect(err.code).toBe("not_found");
    expect(err.message).toBe("task 'abc' not found");
  });

  it("creates cyclic dependency error", () => {
    const err = ErrCyclicDependency("a", "b");
    expect(err.code).toBe("cyclic_dependency");
  });

  it("creates invalid transition error", () => {
    const err = ErrInvalidTransition("pending", "completed");
    expect(err.code).toBe("invalid_transition");
  });
});
```

- [ ] **Step 6: Run test to verify it fails**

Run: `cd apps/orchestrator && bun test src/domain/__tests__/errors.test.ts`
Expected: FAIL

- [ ] **Step 7: Implement errors**

```typescript
// src/domain/errors.ts
export type DomainErrorCode =
  | "not_found"
  | "already_exists"
  | "invalid_transition"
  | "cyclic_dependency"
  | "missing_required"
  | "invalid_phase"
  | "approval_decided";

export class DomainError extends Error {
  constructor(
    public readonly code: DomainErrorCode,
    message: string,
    public readonly context?: Record<string, unknown>,
  ) {
    super(message);
    this.name = "DomainError";
  }
}

export const ErrNotFound = (entity: string, id: string | number) =>
  new DomainError("not_found", `${entity} '${id}' not found`);

export const ErrAlreadyExists = (entity: string, id: string | number) =>
  new DomainError("already_exists", `${entity} '${id}' already exists`);

export const ErrInvalidTransition = (from: string, to: string) =>
  new DomainError("invalid_transition", `cannot transition from '${from}' to '${to}'`);

export const ErrCyclicDependency = (from: string | number, to: string | number) =>
  new DomainError("cyclic_dependency", `adding ${from} → ${to} creates a cycle`);

export const ErrMissingRequired = (field: string) =>
  new DomainError("missing_required", `missing required field: ${field}`);

export const ErrInvalidPhase = (phase: string) =>
  new DomainError("invalid_phase", `invalid phase: ${phase}`);

export const ErrApprovalDecided = (id: string) =>
  new DomainError("approval_decided", `approval '${id}' already decided`);
```

- [ ] **Step 8: Run test to verify it passes**

Run: `cd apps/orchestrator && bun test src/domain/__tests__/errors.test.ts`
Expected: PASS

- [ ] **Step 9: Create barrel export**

```typescript
// src/domain/index.ts
export * from "./enums";
export * from "./entities";
export * from "./errors";
export * from "./state";
```

- [ ] **Step 10: Run all domain tests**

Run: `cd apps/orchestrator && bun test src/domain/`
Expected: ALL PASS

- [ ] **Step 11: Commit**

```bash
git add apps/orchestrator/src/domain/
git commit -m "feat(orchestrator): add domain errors, state machine, and barrel export"
```

---

### Task 4: Database — Setup and Connection

**Files:**
- Create: `apps/orchestrator/src/db/database.ts`
- Test: `apps/orchestrator/src/db/__tests__/database.test.ts`

- [ ] **Step 1: Write failing test for database**

```typescript
// src/db/__tests__/database.test.ts
import { describe, it, expect } from "bun:test";
import { openDatabase, openMemory } from "../database";

describe("openDatabase", () => {
  it("opens an in-memory database", () => {
    const db = openMemory();
    expect(db).toBeDefined();
    db.close();
  });

  it("sets WAL journal mode", () => {
    const db = openMemory();
    const result = db.query("PRAGMA journal_mode").get() as { journal_mode: string };
    // In-memory DB uses "memory" mode, but WAL is set for file DBs
    expect(result.journal_mode).toBe("memory");
    db.close();
  });

  it("enables foreign keys", () => {
    const db = openMemory();
    const result = db.query("PRAGMA foreign_keys").get() as { foreign_keys: number };
    expect(result.foreign_keys).toBe(1);
    db.close();
  });

  it("sets busy timeout", () => {
    const db = openMemory();
    const result = db.query("PRAGMA busy_timeout").get() as { busy_timeout: number };
    expect(result.busy_timeout).toBe(5000);
    db.close();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/orchestrator && bun test src/db/__tests__/database.test.ts`
Expected: FAIL

- [ ] **Step 3: Implement database**

```typescript
// src/db/database.ts
import { Database } from "bun:sqlite";
import { mkdirSync } from "node:fs";
import { join } from "node:path";
import { migrate } from "./migrations";

export function openDatabase(path: string): Database {
  const db = new Database(path);
  db.run("PRAGMA journal_mode = WAL");
  db.run("PRAGMA foreign_keys = ON");
  db.run("PRAGMA busy_timeout = 5000");
  db.run("PRAGMA synchronous = NORMAL");
  migrate(db);
  return db;
}

export function openMemory(): Database {
  return openDatabase(":memory:");
}

export function openProjectDb(): Database {
  const dir = join(process.cwd(), ".oraculo");
  mkdirSync(dir, { recursive: true });
  return openDatabase(join(dir, "oraculo.db"));
}
```

- [ ] **Step 4: Create migrations stub** (so imports resolve)

```typescript
// src/db/migrations.ts
import type { Database } from "bun:sqlite";

export function migrate(_db: Database): void {
  // Will be implemented in Task 5
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd apps/orchestrator && bun test src/db/__tests__/database.test.ts`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add apps/orchestrator/src/db/
git commit -m "feat(orchestrator): add database setup with pragmas"
```

---

### Task 5: Database — Migrations

**Files:**
- Modify: `apps/orchestrator/src/db/migrations.ts`
- Test: `apps/orchestrator/src/db/__tests__/migrations.test.ts`

Reference: `apps/backend/src/db/migrations.go` — replicate ALL 9 migrations exactly.

- [ ] **Step 1: Write failing test for migrations**

```typescript
// src/db/__tests__/migrations.test.ts
import { describe, it, expect } from "bun:test";
import { openMemory } from "../database";

describe("migrations", () => {
  it("runs all 9 migrations", () => {
    const db = openMemory();
    const version = db.query("PRAGMA user_version").get() as { user_version: number };
    expect(version.user_version).toBe(9);
    db.close();
  });

  it("creates epics table", () => {
    const db = openMemory();
    const result = db.query("SELECT name FROM sqlite_master WHERE type='table' AND name='epics'").get();
    expect(result).toBeTruthy();
    db.close();
  });

  it("creates tasks table with all columns", () => {
    const db = openMemory();
    const cols = db.query("PRAGMA table_info(tasks)").all() as { name: string }[];
    const colNames = cols.map((c) => c.name);
    expect(colNames).toContain("id");
    expect(colNames).toContain("story_id");
    expect(colNames).toContain("name");
    expect(colNames).toContain("description");
    expect(colNames).toContain("status");
    expect(colNames).toContain("type");
    db.close();
  });

  it("creates knowledge FTS5 virtual table", () => {
    const db = openMemory();
    const result = db.query("SELECT name FROM sqlite_master WHERE type='table' AND name='knowledge_fts'").get();
    expect(result).toBeTruthy();
    db.close();
  });

  it("creates approval_comments table (V9)", () => {
    const db = openMemory();
    const result = db.query("SELECT name FROM sqlite_master WHERE type='table' AND name='approval_comments'").get();
    expect(result).toBeTruthy();
    db.close();
  });

  it("creates agents table with task_id column (V8)", () => {
    const db = openMemory();
    const cols = db.query("PRAGMA table_info(agents)").all() as { name: string }[];
    expect(cols.map((c) => c.name)).toContain("task_id");
    db.close();
  });

  it("creates validations table with findings column (V7)", () => {
    const db = openMemory();
    const cols = db.query("PRAGMA table_info(validations)").all() as { name: string }[];
    expect(cols.map((c) => c.name)).toContain("findings");
    db.close();
  });

  it("is idempotent", () => {
    const db = openMemory();
    // openMemory already ran migrate. Close and reopen should not fail.
    const version1 = db.query("PRAGMA user_version").get() as { user_version: number };
    // Manually call migrate again (dynamic import to avoid hoisting)
    const { migrate } = await import("../migrations");
    migrate(db);
    const version2 = db.query("PRAGMA user_version").get() as { user_version: number };
    expect(version2.user_version).toBe(version1.user_version);
    db.close();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/orchestrator && bun test src/db/__tests__/migrations.test.ts`
Expected: FAIL (user_version = 0, tables don't exist)

- [ ] **Step 3: Implement all 9 migrations**

Read `apps/backend/src/db/migrations.go` for exact SQL. Replicate every CREATE TABLE, every ALTER TABLE, every trigger, every FTS5 virtual table. Each migration must run inside `db.transaction()`.

```typescript
// src/db/migrations.ts
import type { Database } from "bun:sqlite";

type Migration = (db: Database) => void;

const migrations: Migration[] = [
  // V1: Core tables
  (db) => {
    db.run(`CREATE TABLE IF NOT EXISTS epics (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT NOT NULL DEFAULT '',
      approval_status TEXT NOT NULL DEFAULT 'pending'
        CHECK(approval_status IN ('pending','approved','rejected','needs_revision')),
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS stories (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      epic_id INTEGER NOT NULL REFERENCES epics(id),
      name TEXT NOT NULL,
      description TEXT NOT NULL DEFAULT '',
      approval_status TEXT NOT NULL DEFAULT 'pending'
        CHECK(approval_status IN ('pending','approved','rejected','needs_revision')),
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now')),
      UNIQUE(epic_id, name)
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS tasks (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      story_id INTEGER NOT NULL REFERENCES stories(id),
      name TEXT NOT NULL,
      description TEXT NOT NULL DEFAULT '',
      status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending','in_progress','completed','failed')),
      type TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now')),
      UNIQUE(story_id, name)
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS task_dependencies (
      task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
      depends_on INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
      PRIMARY KEY(task_id, depends_on)
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS task_results (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
      summary TEXT NOT NULL DEFAULT '',
      log TEXT NOT NULL DEFAULT '',
      skills_used TEXT NOT NULL DEFAULT '',
      files_modified TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS validations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      story_id INTEGER NOT NULL REFERENCES stories(id),
      verdict TEXT NOT NULL,
      findings TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS knowledge (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      domain TEXT NOT NULL,
      category TEXT NOT NULL
        CHECK(category IN ('pattern','convention','constraint','dependency','test','architecture')),
      finding TEXT NOT NULL,
      source_files TEXT NOT NULL DEFAULT '',
      confidence TEXT NOT NULL DEFAULT 'medium'
        CHECK(confidence IN ('high','medium','low')),
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);

    db.run(`CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_fts USING fts5(
      domain, category, finding, source_files, content=knowledge, content_rowid=id
    )`);

    db.run(`CREATE TRIGGER IF NOT EXISTS knowledge_ai AFTER INSERT ON knowledge BEGIN
      INSERT INTO knowledge_fts(rowid, domain, category, finding, source_files)
      VALUES (new.id, new.domain, new.category, new.finding, new.source_files);
    END`);

    db.run(`CREATE TRIGGER IF NOT EXISTS knowledge_ad AFTER DELETE ON knowledge BEGIN
      INSERT INTO knowledge_fts(knowledge_fts, rowid, domain, category, finding, source_files)
      VALUES ('delete', old.id, old.domain, old.category, old.finding, old.source_files);
    END`);

    db.run(`CREATE TABLE IF NOT EXISTS approvals (
      id TEXT PRIMARY KEY,
      type TEXT NOT NULL CHECK(type IN ('qa-escalation','execution-plan','design')),
      epic_id INTEGER REFERENCES epics(id),
      story_id INTEGER REFERENCES stories(id),
      task_id INTEGER REFERENCES tasks(id),
      version_id INTEGER,
      status TEXT NOT NULL DEFAULT 'pending'
        CHECK(status IN ('pending','approved','rejected','needs_revision')),
      description TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);
  },

  // V2: Sessions
  (db) => {
    db.run(`CREATE TABLE IF NOT EXISTS sessions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      type TEXT NOT NULL CHECK(type IN ('epic','story','plan','execute','validate')),
      epic_id INTEGER NOT NULL REFERENCES epics(id),
      description TEXT NOT NULL DEFAULT '',
      status TEXT NOT NULL DEFAULT 'active'
        CHECK(status IN ('active','completed','abandoned')),
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      updated_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS session_phases (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id INTEGER NOT NULL REFERENCES sessions(id),
      name TEXT NOT NULL,
      data TEXT NOT NULL DEFAULT '{}',
      completed_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS claude_sessions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL UNIQUE,
      working_directory TEXT NOT NULL DEFAULT '',
      git_branch TEXT NOT NULL DEFAULT '',
      source TEXT NOT NULL DEFAULT '',
      started_at TEXT NOT NULL DEFAULT (datetime('now')),
      ended_at TEXT
    )`);
  },

  // V3: Agents and tool events
  (db) => {
    db.run(`CREATE TABLE IF NOT EXISTS agents (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL,
      name TEXT NOT NULL DEFAULT '',
      type TEXT NOT NULL DEFAULT '',
      started_at TEXT NOT NULL DEFAULT (datetime('now')),
      stopped_at TEXT
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS tool_events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL,
      tool_name TEXT NOT NULL,
      input TEXT NOT NULL DEFAULT '',
      output TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);
  },

  // V4: Rebuild approvals (add 'design' type) — already included in V1 above
  // In Go this was needed to alter the CHECK constraint. Since we create with all types in V1, this is a no-op.
  (_db) => {},

  // V5: Versions and reviews
  (db) => {
    db.run(`CREATE TABLE IF NOT EXISTS epic_versions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      epic_id INTEGER NOT NULL REFERENCES epics(id),
      version INTEGER NOT NULL,
      content TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      UNIQUE(epic_id, version)
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS story_versions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      story_id INTEGER NOT NULL REFERENCES stories(id),
      version INTEGER NOT NULL,
      content TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now')),
      UNIQUE(story_id, version)
    )`);

    db.run(`CREATE TABLE IF NOT EXISTS reviews (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      version_id INTEGER NOT NULL,
      type TEXT NOT NULL CHECK(type IN ('epic','story')),
      verdict TEXT NOT NULL CHECK(verdict IN ('approved','rejected')),
      comment TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);
  },

  // V6: claude_sessions.ended_at + session_events
  (db) => {
    // ended_at already in V2 creation above for fresh DBs
    // ALTER needed for DBs migrating from old schema
    try {
      db.run("ALTER TABLE claude_sessions ADD COLUMN ended_at TEXT");
    } catch {
      // Column already exists (fresh DB)
    }
    db.run(`CREATE TABLE IF NOT EXISTS session_events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL,
      type TEXT NOT NULL,
      data TEXT NOT NULL DEFAULT '{}',
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);
  },

  // V7: validations.findings
  (db) => {
    // Already in V1 creation? No — V1 doesn't include findings. Need to add it.
    // Use ALTER TABLE for DBs migrating from V1. For fresh DBs, IF NOT EXISTS handles it.
    try {
      db.run("ALTER TABLE validations ADD COLUMN findings TEXT NOT NULL DEFAULT ''");
    } catch {
      // Column already exists (fresh DB created with V1 that already has it? No, V1 doesn't have it)
      // This catch handles the case where the migration runs on a DB that already has the column
    }
  },

  // V8: agents.task_id
  (db) => {
    try {
      db.run("ALTER TABLE agents ADD COLUMN task_id INTEGER REFERENCES tasks(id)");
    } catch {
      // Column already exists
    }
  },

  // V9: approval_comments
  (db) => {
    db.run(`CREATE TABLE IF NOT EXISTS approval_comments (
      id TEXT PRIMARY KEY,
      approval_id TEXT NOT NULL REFERENCES approvals(id) ON DELETE CASCADE,
      content TEXT NOT NULL,
      created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`);
  },
];

export function migrate(db: Database): void {
  const { user_version } = db.query("PRAGMA user_version").get() as { user_version: number };

  for (let i = user_version; i < migrations.length; i++) {
    db.transaction(() => {
      migrations[i](db);
      db.run(`PRAGMA user_version = ${i + 1}`);
    })();
  }
}
```

**Important**: Read `apps/backend/src/db/migrations.go` carefully. The SQL above is a best approximation — verify every column name, type, constraint, and default against the Go source. Pay special attention to:
- V4 rebuilds the approvals table (different CHECK constraint) — since we define V1 with the final constraint, V4 can be a no-op
- V5 rebuilds approvals again (removes old types) — same logic, no-op for fresh DBs
- V6 adds `ended_at` to claude_sessions — since we define V2 with it, this is just session_events
- V7 and V8 use ALTER TABLE — wrap in try/catch for idempotency

- [ ] **Step 4: Run test to verify it passes**

Run: `cd apps/orchestrator && bun test src/db/__tests__/migrations.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add apps/orchestrator/src/db/migrations.ts apps/orchestrator/src/db/__tests__/
git commit -m "feat(orchestrator): add all 9 database migrations"
```

---

### Task 6: Utilities — Output and Env

**Files:**
- Create: `apps/orchestrator/src/utils/output.ts`
- Create: `apps/orchestrator/src/utils/env.ts`
- Create: `apps/orchestrator/src/utils/uuid.ts`
- Create: `apps/orchestrator/src/db/index.ts`
- Test: `apps/orchestrator/src/utils/__tests__/output.test.ts`
- Test: `apps/orchestrator/src/utils/__tests__/env.test.ts`

Reference: `apps/backend/src/output/output.go` for writeJSON, writeError, writeTable formats.

- [ ] **Step 1: Write failing test for output utils**

```typescript
// src/utils/__tests__/output.test.ts
import { describe, it, expect } from "bun:test";
import { formatJSON, formatError, formatTable } from "../output";
import { DomainError } from "../../domain/errors";

describe("formatJSON", () => {
  it("returns indented JSON", () => {
    const result = formatJSON({ name: "test" });
    expect(result).toBe('{\n  "name": "test"\n}');
  });
});

describe("formatError", () => {
  it("maps domain error to code", () => {
    const err = new DomainError("not_found", "task 'x' not found");
    const result = JSON.parse(formatError(err));
    expect(result.error).toBe("not_found");
    expect(result.message).toBe("task 'x' not found");
  });

  it("maps generic error", () => {
    const err = new Error("something broke");
    const result = JSON.parse(formatError(err));
    expect(result.error).toBe("internal");
    expect(result.message).toBe("something broke");
  });
});

describe("formatTable", () => {
  it("renders aligned table", () => {
    const result = formatTable(
      ["Name", "Status"],
      [
        ["task-1", "pending"],
        ["task-2", "completed"],
      ],
    );
    expect(result).toContain("Name");
    expect(result).toContain("Status");
    expect(result).toContain("task-1");
    expect(result).toContain("pending");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd apps/orchestrator && bun test src/utils/__tests__/output.test.ts`
Expected: FAIL

- [ ] **Step 3: Implement output utils**

```typescript
// src/utils/output.ts
import { DomainError } from "../domain/errors";

export function writeJSON(data: unknown): void {
  console.log(formatJSON(data));
}

export function formatJSON(data: unknown): string {
  return JSON.stringify(data, null, 2);
}

export function writeError(err: unknown): void {
  console.error(formatError(err instanceof Error ? err : new Error(String(err))));
  process.exit(1);
}

export function formatError(err: Error): string {
  const code = err instanceof DomainError ? err.code : "internal";
  return JSON.stringify({ error: code, message: err.message }, null, 2);
}

export function writeTable(headers: string[], rows: string[][]): void {
  console.log(formatTable(headers, rows));
}

export function formatTable(headers: string[], rows: string[][]): string {
  const widths = headers.map((h, i) =>
    Math.max(h.length, ...rows.map((r) => (r[i] ?? "").length)),
  );

  const line = widths.map((w) => "─".repeat(w + 2)).join("┼");
  const header = headers.map((h, i) => ` ${h.padEnd(widths[i])} `).join("│");
  const body = rows
    .map((row) => row.map((cell, i) => ` ${(cell ?? "").padEnd(widths[i])} `).join("│"))
    .join("\n");

  return `${header}\n${line}\n${body}`;
}
```

- [ ] **Step 4: Implement env and uuid utils**

```typescript
// src/utils/env.ts
export function cleanEnv(): Record<string, string> {
  const env: Record<string, string> = {};
  for (const [key, value] of Object.entries(process.env)) {
    if (key !== "CLAUDECODE" && value !== undefined) {
      env[key] = value;
    }
  }
  return env;
}
```

```typescript
// src/utils/uuid.ts
export function generateUUID(): string {
  return crypto.randomUUID();
}
```

- [ ] **Step 5: Write failing test for env utils**

```typescript
// src/utils/__tests__/env.test.ts
import { describe, it, expect } from "bun:test";
import { cleanEnv } from "../env";

describe("cleanEnv", () => {
  it("strips CLAUDECODE from env", () => {
    const original = process.env.CLAUDECODE;
    process.env.CLAUDECODE = "1";
    const env = cleanEnv();
    expect(env.CLAUDECODE).toBeUndefined();
    if (original) process.env.CLAUDECODE = original;
    else delete process.env.CLAUDECODE;
  });

  it("preserves other env vars", () => {
    const env = cleanEnv();
    expect(env.PATH).toBeDefined();
  });
});
```

- [ ] **Step 6: Run all utils tests**

Run: `cd apps/orchestrator && bun test src/utils/`
Expected: ALL PASS

- [ ] **Step 7: Create barrel exports**

```typescript
// src/db/index.ts
export { openDatabase, openMemory, openProjectDb } from "./database";
export { migrate } from "./migrations";
```

```typescript
// src/utils/index.ts
export * from "./output";
export * from "./env";
export * from "./uuid";
```

- [ ] **Step 8: Commit**

```bash
git add apps/orchestrator/src/utils/ apps/orchestrator/src/db/index.ts
git commit -m "feat(orchestrator): add output, env, uuid utilities"
```

---

## Chunk 2: Data Layer — Stores

### Task 7: EpicStore

**Files:**
- Create: `apps/orchestrator/src/db/stores/epic-store.ts`
- Test: `apps/orchestrator/src/db/stores/__tests__/epic-store.test.ts`

Reference: `apps/backend/src/db/epic_store.go` — replicate ALL methods including ListSummaries with its complex JOIN.

- [ ] **Step 1: Write failing tests**

Test every method: Create, GetByName, List, ListSummaries, Update, Delete, UpdateApprovalStatus. Include edge cases: duplicate names (ErrAlreadyExists), not found (ErrNotFound).

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd apps/orchestrator && bun test src/db/stores/__tests__/epic-store.test.ts`

- [ ] **Step 3: Implement EpicStore**

Read `apps/backend/src/db/epic_store.go` for exact SQL queries. Pay special attention to `ListSummaries` which JOINs sessions, stories, and tasks to compute phase/status/counts.

- [ ] **Step 4: Run tests to verify they pass**

- [ ] **Step 5: Commit**

```bash
git add apps/orchestrator/src/db/stores/epic-store.ts apps/orchestrator/src/db/stores/__tests__/epic-store.test.ts
git commit -m "feat(orchestrator): add EpicStore with ListSummaries"
```

---

### Task 8: StoryStore

**Files:**
- Create: `apps/orchestrator/src/db/stores/story-store.ts`
- Test: `apps/orchestrator/src/db/stores/__tests__/story-store.test.ts`

Reference: `apps/backend/src/db/story_store.go`

- [ ] **Step 1-5**: Same TDD pattern as Task 7. Include ListSummaries (JOIN with task counts), UpdateApprovalStatus, and update-status.

---

### Task 9: TaskStore (most complex store)

**Files:**
- Create: `apps/orchestrator/src/db/stores/task-store.ts`
- Test: `apps/orchestrator/src/db/stores/__tests__/task-store.test.ts`

Reference: `apps/backend/src/db/task_store.go` — this is the most complex store with transactions, dependency management, and cycle detection.

- [ ] **Step 1: Write failing tests**

Cover: Create (with deps), GetByName, List, ListEnriched (batch deps + results), Start, Complete (with transaction + result insert), Fail, Delete, addDependencies, detectCycle (BFS). Test cycle detection with A→B→C→A.

- [ ] **Step 2-4: TDD cycle**

Pay special attention to:
- `Create` uses a transaction to insert the task AND its dependencies
- `Complete` uses a transaction to update status AND insert TaskResult
- `ListEnriched` does batch queries for deps and results per task
- `detectCycle` uses BFS traversal
- `addDependencies` validates that dependent tasks belong to the same story

- [ ] **Step 5: Commit**

---

### Task 10: Remaining Stores (batch)

**Files:**
- Create: `apps/orchestrator/src/db/stores/memory-store.ts` (FTS5)
- Create: `apps/orchestrator/src/db/stores/approval-store.ts` (incl. ListByEpic, DeleteCommentsByApproval)
- Create: `apps/orchestrator/src/db/stores/session-store.ts` (incl. ActiveByEpic)
- Create: `apps/orchestrator/src/db/stores/version-store.ts` (incl. LatestEpicVersion, LatestStoryVersion)
- Create: `apps/orchestrator/src/db/stores/review-store.ts` (incl. ListByStory, propagation)
- Create: `apps/orchestrator/src/db/stores/validation-store.ts`
- Create: `apps/orchestrator/src/db/stores/agent-store.ts`
- Create: `apps/orchestrator/src/db/stores/tool-event-store.ts`
- Create: `apps/orchestrator/src/db/stores/session-event-store.ts`
- Create: `apps/orchestrator/src/db/stores/index.ts`
- Tests: One test file per store in `__tests__/`

Reference: corresponding `_store.go` files in `apps/backend/src/db/`

For each store:
- [ ] **Step 1**: Write failing tests covering ALL methods listed in the spec's "Complete Store Method Reference"
- [ ] **Step 2**: Run tests to verify failure
- [ ] **Step 3**: Implement (read the Go source for exact SQL)
- [ ] **Step 4**: Run tests to verify pass
- [ ] **Step 5**: Commit

Special attention:
- **MemoryStore**: FTS5 search query syntax (`MATCH`), domain filter
- **ApprovalStore**: UUID generation for IDs, ListByEpic filter, DeleteCommentsByApproval on approve
- **SessionStore**: CompletePhase auto-closes session when last phase completes (check `phases.go`)
- **VersionStore**: Auto-increment version number per entity, LatestEpicVersion/LatestStoryVersion queries
- **ReviewStore**: `Create` propagates verdict to parent entity's `approval_status` (calls UpdateApprovalStatus on epic/story store)

Create barrel export:

```typescript
// src/db/stores/index.ts
export { EpicStore } from "./epic-store";
export { StoryStore } from "./story-store";
export { TaskStore } from "./task-store";
export { MemoryStore } from "./memory-store";
export { ApprovalStore } from "./approval-store";
export { SessionStore } from "./session-store";
export { VersionStore } from "./version-store";
export { ReviewStore } from "./review-store";
export { ValidationStore } from "./validation-store";
export { AgentStore } from "./agent-store";
export { ToolEventStore } from "./tool-event-store";
export { SessionEventStore } from "./session-event-store";
```

Final commit for all remaining stores:

```bash
git add apps/orchestrator/src/db/stores/
git commit -m "feat(orchestrator): add all 12 data stores with tests"
```

---

## Chunk 3: Server and CLI

### Task 11: WebSocket Hub

**Files:**
- Create: `apps/orchestrator/src/server/ws.ts`
- Test: `apps/orchestrator/src/server/__tests__/ws.test.ts`

Reference: `apps/backend/src/ws/hub.go`

- [ ] **Step 1: Write failing test for hub broadcast**

```typescript
// src/server/__tests__/ws.test.ts
import { describe, it, expect } from "bun:test";
import { WebSocketHub } from "../ws";

describe("WebSocketHub", () => {
  it("broadcasts to connected clients", () => {
    const hub = new WebSocketHub();
    const messages: string[] = [];
    const mockWs = { send: (msg: string) => messages.push(msg), readyState: 1 };
    hub.add(mockWs as any);
    hub.broadcast({ type: "test", data: "hello" });
    expect(messages).toHaveLength(1);
    expect(JSON.parse(messages[0])).toEqual({ type: "test", data: "hello" });
  });

  it("removes disconnected clients", () => {
    const hub = new WebSocketHub();
    const mockWs = { send: () => {}, readyState: 1 };
    hub.add(mockWs as any);
    hub.remove(mockWs as any);
    // Should not throw when broadcasting to empty hub
    hub.broadcast({ type: "test" });
  });
});
```

- [ ] **Step 2-4: TDD cycle**
- [ ] **Step 5: Commit**

---

### Task 12: SSE Broadcaster

**Files:**
- Create: `apps/orchestrator/src/server/broadcaster.ts`
- Test: `apps/orchestrator/src/server/__tests__/broadcaster.test.ts`

Reference: `apps/backend/src/applog/broadcaster.go` — ring buffer of 200 entries, fan-out to SSE subscribers.

- [ ] **Step 1-5: TDD cycle**

---

### Task 13: HTTP Server with Hono

**Files:**
- Create: `apps/orchestrator/src/server/http.ts`
- Create: `apps/orchestrator/src/server/middleware.ts`
- Create: `apps/orchestrator/src/server/routes.ts`
- Create: `apps/orchestrator/src/server/hooks.ts`
- Create: `apps/orchestrator/src/server/index.ts`
- Test: `apps/orchestrator/src/server/__tests__/http.test.ts`

Reference: `apps/backend/src/server/server.go`, `api.go`, `hooks.go`, `cors.go`

- [ ] **Step 1: Write failing test for health endpoint**

```typescript
// src/server/__tests__/http.test.ts
import { describe, it, expect, beforeAll, afterAll } from "bun:test";
import { createApp } from "../http";
import { openMemory } from "../../db";

describe("HTTP Server", () => {
  let server: ReturnType<typeof Bun.serve>;
  let port: number;

  beforeAll(() => {
    const db = openMemory();
    const app = createApp(db);
    server = Bun.serve({ port: 0, fetch: app.fetch });
    port = server.port;
  });

  afterAll(() => {
    server.stop();
  });

  it("returns health status", async () => {
    const res = await fetch(`http://localhost:${port}/health`);
    expect(res.status).toBe(200);
    const body = await res.json();
    expect(body.status).toBe("ok");
  });
});
```

- [ ] **Step 2-3: Implement server with Hono**

Implement ALL endpoints listed in the spec's "Complete HTTP Endpoint Reference":
- Health, system status
- Epic CRUD + summaries
- Story CRUD + summaries + versions + reviews + validations
- Task list (enriched)
- Approval CRUD + verdict + comments
- All hook endpoints (agent-start, agent-stop, tool-used, task-completed, task-started, stop, teammate-idle, session-start, session-end)
- WebSocket upgrade
- SSE logs

- [ ] **Step 4: Run tests**
- [ ] **Step 5: Commit per logical group (API, hooks, WS/SSE)**

---

### Task 14: Config and Registry

**Files:**
- Create: `apps/orchestrator/src/config/config.ts`
- Create: `apps/orchestrator/src/registry/registry.ts`
- Tests for each

Reference: `apps/backend/src/config/config.go`, `apps/backend/src/registry/registry.go`

- [ ] **Step 1-5: TDD cycle for config** (read/write `.oraculo/config.json`, atomic write via temp+rename, ProjectName fallback)
- [ ] **Step 6-10: TDD cycle for registry** (`~/.oraculo/servers.json`, mkdirSync lock, register/unregister/list, stale PID validation)
- [ ] **Step 11: Commit**

---

### Task 15: CLI — Lifecycle Commands

**Files:**
- Create: `apps/orchestrator/src/cli/root.ts`
- Create: `apps/orchestrator/src/cli/lifecycle.ts`
- Modify: `apps/orchestrator/src/index.ts`

Reference: `apps/backend/src/cli/start.go`, `kill.go`, `restart.go`, `status.go`, `version.go`

- [ ] **Step 1: Implement root command tree**

```typescript
// src/cli/root.ts
import { Command } from "commander";
import { VERSION } from "../version";

export function createProgram(): Command {
  const program = new Command();
  program.name("oraculo").version(VERSION);

  // Add all subcommands
  addLifecycleCommands(program);
  addInstallCommands(program);
  addHookCommands(program);
  addToolCommands(program);

  return program;
}
```

- [ ] **Step 2: Implement lifecycle commands**

`start` (HTTP+MCP via Promise.all), `start http` (daemon), `start mcp` (stdio), `kill` (lsof→SIGTERM→SIGKILL→poll), `restart`, `status` (dashboard table), `logs` (SSE client), `version`.

- [ ] **Step 3: Wire entry point**

```typescript
// src/index.ts
#!/usr/bin/env bun
import { createProgram } from "./cli/root";
createProgram().parse();
```

- [ ] **Step 4: Test CLI commands**

Run: `cd apps/orchestrator && bun run src/index.ts version`
Expected: prints version

- [ ] **Step 5: Commit**

---

### Task 16: CLI — Install, Setup, Uninstall

**Files:**
- Create: `apps/orchestrator/src/cli/install.ts`

Reference: `apps/backend/src/cli/install.go`, `setup.go`, `uninstall.go`

- [ ] **Step 1-5: TDD cycle**

`install` (--global/--local/--lang): creates .oraculo/, DB, finds port, writes config, writes .claude/settings.json
`setup` (--lang): plugin-mode, hooks-only, writes .mcp.json
`uninstall` (--purge): removes from settings.json, optional purge .oraculo/

---

### Task 17: CLI — Hook Commands

**Files:**
- Create: `apps/orchestrator/src/cli/hooks/session-start.ts`
- Create: `apps/orchestrator/src/cli/hooks/session-end.ts`
- Create: `apps/orchestrator/src/cli/hooks/agent-start.ts`
- Create: `apps/orchestrator/src/cli/hooks/task-started.ts`
- Create: `apps/orchestrator/src/cli/hooks/index.ts`

Reference: `apps/backend/src/cli/hooks.go` — read stdin JSON, proxy to HTTP server, auto-start daemon.

- [ ] **Step 1: Implement session-start hook**

Critical: this hook auto-starts the HTTP daemon if offline (health check → spawn daemon → poll health → POST).

- [ ] **Step 2-5: Implement remaining hooks + commit**

---

### Task 18: CLI — Tool Commands (all 10)

**Files:**
- Create: `apps/orchestrator/src/cli/tools/epic.ts`
- Create: `apps/orchestrator/src/cli/tools/story.ts`
- Create: `apps/orchestrator/src/cli/tools/task.ts`
- Create: `apps/orchestrator/src/cli/tools/memory.ts`
- Create: `apps/orchestrator/src/cli/tools/approval.ts`
- Create: `apps/orchestrator/src/cli/tools/session.ts`
- Create: `apps/orchestrator/src/cli/tools/phase.ts`
- Create: `apps/orchestrator/src/cli/tools/design.ts`
- Create: `apps/orchestrator/src/cli/tools/review.ts`
- Create: `apps/orchestrator/src/cli/tools/validation.ts`
- Create: `apps/orchestrator/src/cli/tools/index.ts`
- Create: `apps/orchestrator/src/cli/context.ts`

Reference: `apps/backend/src/cli/tools/` — ALL subcommands, ALL flags, ALL stdin parsing, ALL filesystem operations.

- [ ] **Step 1: Implement DB context injection**

```typescript
// src/cli/context.ts
import { Command } from "commander";
import { openProjectDb } from "../db";

export function withDb(cmd: Command): Command {
  cmd.hook("preAction", (thisCmd) => {
    thisCmd.setOptionValue("_db", openProjectDb());
  });
  return cmd;
}
```

- [ ] **Step 2: Implement epic commands**

All subcommands from spec: init, save (stdin→.oraculo/epics/{name}/requirements.md), get (read md), list, update, delete, version (stdin→creates version + design approval), versions.

- [ ] **Step 3: Implement story commands** (same pattern + update-status)
- [ ] **Step 4: Implement task commands** (init with --depends-on, start, complete with stdin JSON, fail, get, list, delete)
- [ ] **Step 5: Implement memory commands** (store, search with FTS5, domains)
- [ ] **Step 6: Implement approval commands** (request with stdin, status, list --pending, verdict)
- [ ] **Step 7: Implement session commands** (init, status, state, close)
- [ ] **Step 8: Implement phase commands** (complete with stdin, auto-close on last phase)
- [ ] **Step 9: Implement design commands** (save stdin→md, get→read md)
- [ ] **Step 10: Implement review commands** (create with propagation, get, list)
- [ ] **Step 11: Implement validation commands** (save with verdict + findings)
- [ ] **Step 12: Commit**

```bash
git add apps/orchestrator/src/cli/
git commit -m "feat(orchestrator): add all CLI tool commands"
```

---

### Task 19: MCP Server

**Files:**
- Create: `apps/orchestrator/src/mcp/server.ts`

Reference: `apps/backend/src/mcp/server.go`

- [ ] **Step 1: Implement MCP server with two tools**

`request_approval` (blocking — uses approval bridge Promise), `approval_status` (non-blocking poll). Zod schemas for input validation. StdioServerTransport.

- [ ] **Step 2: Commit**

---

## Chunk 4: Logging, Orchestrator, Build, Desktop

### Task 20: Logging System

**Files:**
- Create: `apps/orchestrator/src/logging/logger.ts`
- Create: `apps/orchestrator/src/logging/execution-log.ts`
- Create: `apps/orchestrator/src/logging/index.ts`
- Test: `apps/orchestrator/src/logging/__tests__/execution-log.test.ts`

- [ ] **Step 1: Implement operational logger with pino**

Setup pino with:
- Dev: pino-pretty to stdout
- Prod: pino-roll (daily + 50MB, 7 files, symlink)
- Error: always to stderr
- Sync fallback if pino.transport() fails

- [ ] **Step 2: Implement execution log (JSONL per run)**

Per-run directory in `~/.oraculo/runs/<runId>/`. Per-agent file `agent-<taskId>.jsonl`. Pruning of runs > 3 days on startup. All writes in try/catch.

- [ ] **Step 3: Test execution log**

```typescript
// src/logging/__tests__/execution-log.test.ts
import { describe, it, expect } from "bun:test";
import { createExecutionLog } from "../execution-log";
import { readFileSync, rmSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";

describe("ExecutionLog", () => {
  it("writes JSONL entries", () => {
    const dir = join(tmpdir(), `oraculo-test-${Date.now()}`);
    const log = createExecutionLog("test-run", "task-1", dir);
    log.write({ msg: "started", event: "agent_started" });
    log.write({ msg: "completed", event: "agent_completed" });
    const content = readFileSync(log.path, "utf-8").trim().split("\n");
    expect(content).toHaveLength(2);
    expect(JSON.parse(content[0]).msg).toBe("started");
    rmSync(dir, { recursive: true, force: true });
  });
});
```

- [ ] **Step 4: Commit**

---

### Task 21: Orchestrator — DAG and Dispatch

**Files:**
- Create: `apps/orchestrator/src/orchestrator/dag.ts`
- Create: `apps/orchestrator/src/orchestrator/agents.ts`
- Create: `apps/orchestrator/src/orchestrator/prompts.ts`
- Create: `apps/orchestrator/src/orchestrator/dispatch.ts`
- Create: `apps/orchestrator/src/orchestrator/execute.ts`
- Create: `apps/orchestrator/src/orchestrator/approval-bridge.ts`
- Create: `apps/orchestrator/src/orchestrator/index.ts`
- Test: `apps/orchestrator/src/orchestrator/__tests__/dag.test.ts`
- Test: `apps/orchestrator/src/orchestrator/__tests__/approval-bridge.test.ts`

- [ ] **Step 1: Write failing test for DAG evaluation**

```typescript
// src/orchestrator/__tests__/dag.test.ts
import { describe, it, expect } from "bun:test";
import { nextWave } from "../dag";
import type { TaskEnriched } from "../../domain";

describe("nextWave", () => {
  it("returns tasks with no dependencies", () => {
    const tasks: TaskEnriched[] = [
      { id: 1, story_id: 1, name: "t1", description: "", status: "pending", type: "code", depends_on: [], result: null, created_at: "", updated_at: "" },
      { id: 2, story_id: 1, name: "t2", description: "", status: "pending", type: "code", depends_on: [1], result: null, created_at: "", updated_at: "" },
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks.map((t) => t.id)).toEqual([1]);
  });

  it("returns tasks whose deps are completed", () => {
    const tasks: TaskEnriched[] = [
      { id: 1, story_id: 1, name: "t1", description: "", status: "completed", type: "code", depends_on: [], result: null, created_at: "", updated_at: "" },
      { id: 2, story_id: 1, name: "t2", description: "", status: "pending", type: "code", depends_on: [1], result: null, created_at: "", updated_at: "" },
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks.map((t) => t.id)).toEqual([2]);
  });

  it("returns empty wave when deadlocked", () => {
    const tasks: TaskEnriched[] = [
      { id: 1, story_id: 1, name: "t1", description: "", status: "pending", type: "code", depends_on: [2], result: null, created_at: "", updated_at: "" },
      { id: 2, story_id: 1, name: "t2", description: "", status: "pending", type: "code", depends_on: [1], result: null, created_at: "", updated_at: "" },
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks).toHaveLength(0);
  });
});
```

- [ ] **Step 2: Implement DAG**

```typescript
// src/orchestrator/dag.ts
import type { TaskEnriched } from "../domain";

export interface Wave {
  tasks: TaskEnriched[];
}

export function nextWave(tasks: TaskEnriched[]): Wave {
  const ready = tasks.filter((t) =>
    t.status === "pending" &&
    t.depends_on.every((depId) => {
      const dep = tasks.find((d) => d.id === depId);
      return dep?.status === "completed";
    }),
  );
  return { tasks: ready };
}
```

- [ ] **Step 3: Run test to verify it passes**
- [ ] **Step 4: Implement agent definitions**

```typescript
// src/orchestrator/agents.ts — agent type definitions with model, tools, system prompt
```

- [ ] **Step 5: Implement prompt builders**

```typescript
// src/orchestrator/prompts.ts — buildTaskPrompt, buildQAPrompt, buildResearchPrompt
```

- [ ] **Step 6: Implement approval bridge with tests**

Test: request blocks until decide is called. Test: multiple concurrent approvals.

- [ ] **Step 7: Implement dispatch** (Agent SDK query() per task, streaming events, WebSocket broadcast)
- [ ] **Step 8: Implement execute** (wave loop, circuit breaker, approval gates)
- [ ] **Step 9: Add `oraculo execute --story <id>` command to CLI**
- [ ] **Step 10: Commit**

```bash
git add apps/orchestrator/src/orchestrator/
git commit -m "feat(orchestrator): add Agent SDK orchestration with DAG dispatch"
```

---

### Task 22: Build Script and Cross-Compilation

**Files:**
- Create: `apps/orchestrator/build.ts`

- [ ] **Step 1: Implement build script**

```typescript
// build.ts
// Note: bun build --compile must be invoked via CLI, not Bun.build() API
const targets = [
  "bun-darwin-arm64",
  "bun-darwin-x64",
  "bun-linux-x64",
  "bun-linux-arm64",
] as const;

for (const target of targets) {
  console.log(`Building for ${target}...`);
  const proc = Bun.spawn([
    "bun", "build", "--compile",
    `--target=${target}`,
    "--outfile", `./dist/${target}/oraculo`,
    "./src/index.ts",
  ], { stdout: "inherit", stderr: "inherit" });
  const exitCode = await proc.exited;
  if (exitCode !== 0) {
    console.error(`Build failed for ${target}`);
    process.exit(1);
  }
}

console.log("All builds complete.");
```

- [ ] **Step 2: Test local build**

Run: `cd apps/orchestrator && bun run build.ts`
Expected: binary at `dist/bun-darwin-arm64/oraculo` (or current platform)

- [ ] **Step 3: Test compiled binary**

Run: `./dist/bun-darwin-arm64/oraculo version`
Expected: prints version

Run: `./dist/bun-darwin-arm64/oraculo tools task list --epic test --story test`
Expected: valid JSON output (empty array or error)

- [ ] **Step 3.5: Test MCP SDK bundling**

If the compiled binary fails with MCP-related errors, update build.ts to mark `@modelcontextprotocol/sdk` as external and bundle it separately. Test again.

- [ ] **Step 4: Commit**

```bash
git add apps/orchestrator/build.ts
git commit -m "feat(orchestrator): add build script with bun-plugin-pino"
```

---

### Task 23: Desktop App — Inline Shared Code

**Files:**
- Modify: `apps/desktop/spa.go`
- Modify: `apps/desktop/launcher_service.go`
- Modify: `apps/desktop/go.mod`

- [ ] **Step 1: Inline spa functions**

Copy `WithPlaceholders()` (~40 LOC) and `Shell()` (~30 LOC) from `apps/backend/src/spa/spa.go` directly into `apps/desktop/spa.go`. Remove the import of `github.com/lucas/oraculo/apps/backend/src/spa`.

- [ ] **Step 2: Inline registry functions**

Copy `DefaultPath()`, `List()`, `WriteAll()` (~110 LOC) from `apps/backend/src/registry/registry.go` into a new file `apps/desktop/registry.go` or inline into `launcher_service.go`. Remove the import of `github.com/lucas/oraculo/apps/backend/src/registry`.

- [ ] **Step 3: Update go.mod**

Remove the `replace` directive for the backend module. Run `go mod tidy`.

- [ ] **Step 4: Test desktop build**

Run: `cd apps/desktop && go build .`
Expected: compiles without errors

- [ ] **Step 5: Commit**

```bash
git add apps/desktop/
git commit -m "refactor(desktop): inline spa and registry from backend"
```

---

### Task 24: npm Package Updates

**Files:**
- Modify: `npm/cli/package.json` (if binary name changes)
- Modify: `npm/cli-darwin-arm64/package.json` etc.

- [ ] **Step 1: Update platform packages to ship Bun binary instead of Go binary**

The `bin/oraculo` in each platform package now points to the compiled Bun binary. No other changes needed — the launcher `cli.js` resolves the binary by platform as before.

- [ ] **Step 2: Commit**

---

### Task 25: Cleanup — Remove Go Backend

**Files:**
- Delete: `apps/backend/` (entire directory)
- Delete: `claude-kit/embed.go`
- Modify: `claude-kit/` as needed

- [ ] **Step 1: Verify no remaining imports of backend**

Run: `grep -r "apps/backend" --include="*.go" --include="*.ts" --include="*.json" .`
Expected: no results (except maybe docs/specs)

- [ ] **Step 2: Delete Go backend**

```bash
rm -rf apps/backend/
```

- [ ] **Step 3: Delete embed.go**

```bash
rm claude-kit/embed.go
```

- [ ] **Step 4: Run full test suite**

```bash
cd apps/orchestrator && bun test
```
Expected: ALL PASS

- [ ] **Step 5: Test compiled binary end-to-end**

```bash
cd apps/orchestrator && bun run build.ts
./dist/bun-darwin-arm64/index install
./dist/bun-darwin-arm64/index start http
curl http://localhost:3100/health
./dist/bun-darwin-arm64/index kill
```

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "chore: remove Go backend, migration complete"
```

---

## Task Dependency Graph

```
Task 1 (setup)
  └── Task 2 (enums) + Task 3 (errors/state)
       └── Task 4 (DB setup) + Task 5 (migrations) + Task 6 (utils)
            └── Task 7-10 (stores)
                 └── Task 11-13 (server)
                 └── Task 14 (config/registry)
                      └── Task 15-18 (CLI)
                      └── Task 19 (MCP)
                 └── Task 20 (logging)
                      └── Task 21 (orchestrator)
                           └── Task 22 (build)
                                └── Task 23 (desktop)
                                └── Task 24 (npm)
                                     └── Task 25 (cleanup)
```

**Parallelizable groups:**
- Tasks 2+3 (domain enums and errors)
- Tasks 4+5+6 (DB, migrations, utils)
- Tasks 7-10 (all stores — independent of each other, depend on DB)
- Tasks 11+12+13 (server components)
- Tasks 15+16+17+18 (CLI commands — can be parallelized after server)
- Tasks 23+24 (desktop + npm — independent)
