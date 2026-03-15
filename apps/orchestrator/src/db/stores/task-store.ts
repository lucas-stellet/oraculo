import type { Database } from "bun:sqlite";
import type { Task, TaskEnriched, TaskResult } from "../../domain/entities.ts";
import type { TaskStatus } from "../../domain/enums.ts";
import {
  ErrNotFound,
  ErrCyclicDependency,
  ErrInvalidTransition,
} from "../../domain/errors.ts";
import { canTransition } from "../../domain/state.ts";

const TASK_COLUMNS =
  "id, story_id, name, description, status, failure_reason, created_at, updated_at, started_at, completed_at";

interface TaskRow {
  id: number;
  story_id: number;
  name: string;
  description: string;
  status: TaskStatus;
  failure_reason: string;
  created_at: string;
  updated_at: string;
  started_at: string | null;
  completed_at: string | null;
}

function rowToTask(row: TaskRow): Task {
  const task: Task = {
    id: row.id,
    story_id: row.story_id,
    name: row.name,
    description: row.description,
    status: row.status,
    failure_reason: row.failure_reason,
    created_at: row.created_at,
    updated_at: row.updated_at,
  };
  if (row.started_at != null) task.started_at = row.started_at;
  if (row.completed_at != null) task.completed_at = row.completed_at;
  return task;
}

export interface CompleteInput {
  summary: string;
  logs: string;
  skills_used: string[];
  files_modified: string[];
}

export class TaskStore {
  constructor(private db: Database) {}

  /**
   * Inserts a new task or returns the existing one if (storyId, name) already exists.
   * Returns [task, created] where created is true when a new row was inserted.
   * If dependsOn is non-empty, dependency edges are created atomically.
   */
  create(
    storyId: number,
    name: string,
    description: string,
    dependsOn: string[],
  ): [Task, boolean] {
    // Use a transaction so insert + dependency edges are atomic
    const tx = this.db.transaction(() => {
      const res = this.db.run(
        "INSERT OR IGNORE INTO tasks (story_id, name, description) VALUES (?, ?, ?)",
        [storyId, name, description],
      );
      const affected = res.changes;

      const row = this.db
        .query(`SELECT ${TASK_COLUMNS} FROM tasks WHERE story_id = ? AND name = ?`)
        .get(storyId, name) as TaskRow;

      if (dependsOn.length > 0) {
        this.addDependencies(storyId, row.id, name, dependsOn);
      }

      return [rowToTask(row), affected > 0] as [Task, boolean];
    });

    return tx();
  }

  /**
   * Retrieves a task by its unique (storyId, name) pair.
   * Throws ErrNotFound when no matching task exists.
   */
  getByName(storyId: number, name: string): Task {
    const row = this.db
      .query(`SELECT ${TASK_COLUMNS} FROM tasks WHERE story_id = ? AND name = ?`)
      .get(storyId, name) as TaskRow | null;

    if (!row) {
      throw ErrNotFound("task", name);
    }
    return rowToTask(row);
  }

  /**
   * Returns all tasks for the given storyId ordered by creation time.
   */
  list(storyId: number): Task[] {
    const rows = this.db
      .query(
        `SELECT ${TASK_COLUMNS} FROM tasks WHERE story_id = ? ORDER BY created_at`,
      )
      .all(storyId) as TaskRow[];

    return rows.map(rowToTask);
  }

  /**
   * Returns all tasks for the given storyId with dependencies and results hydrated.
   */
  listEnriched(storyId: number): TaskEnriched[] {
    const tasks = this.list(storyId);
    if (tasks.length === 0) return [];

    const ids = tasks.map((t) => t.id);
    const idxById = new Map<number, number>();
    tasks.forEach((t, i) => idxById.set(t.id, i));

    const enriched: TaskEnriched[] = tasks.map((t) => ({
      ...t,
      depends_on: [],
    }));

    // Batch fetch dependencies
    const placeholders = ids.map(() => "?").join(",");
    const depRows = this.db
      .query(
        `SELECT task_id, depends_on FROM task_dependencies WHERE task_id IN (${placeholders})`,
      )
      .all(...ids) as { task_id: number; depends_on: number }[];

    for (const row of depRows) {
      const idx = idxById.get(row.task_id);
      if (idx !== undefined) {
        enriched[idx].depends_on.push(row.depends_on);
      }
    }

    // Batch fetch results
    const resRows = this.db
      .query(
        `SELECT id, task_id, summary, logs, skills_used, files_modified, created_at FROM task_results WHERE task_id IN (${placeholders})`,
      )
      .all(...ids) as {
      id: number;
      task_id: number;
      summary: string;
      logs: string;
      skills_used: string;
      files_modified: string;
      created_at: string;
    }[];

    for (const row of resRows) {
      const idx = idxById.get(row.task_id);
      if (idx !== undefined) {
        enriched[idx].result = {
          id: row.id,
          task_id: row.task_id,
          summary: row.summary,
          logs: row.logs,
          skills_used: row.skills_used ? row.skills_used.split(",") : [],
          files_modified: row.files_modified
            ? row.files_modified.split(",")
            : [],
          created_at: row.created_at,
        };
      }
    }

    return enriched;
  }

  /**
   * Transitions a task from pending to in_progress, setting started_at.
   */
  start(storyId: number, name: string): Task {
    const task = this.getByName(storyId, name);

    if (!canTransition(task.status, "in_progress")) {
      throw ErrInvalidTransition(task.status, "in_progress");
    }

    this.db.run(
      "UPDATE tasks SET status = ?, started_at = datetime('now'), updated_at = datetime('now') WHERE story_id = ? AND name = ?",
      ["in_progress", storyId, name],
    );

    return this.getByName(storyId, name);
  }

  /**
   * Transitions a task from in_progress to completed and inserts a TaskResult.
   */
  complete(storyId: number, name: string, input: CompleteInput): Task {
    const task = this.getByName(storyId, name);

    if (!canTransition(task.status, "completed")) {
      throw ErrInvalidTransition(task.status, "completed");
    }

    const tx = this.db.transaction(() => {
      this.db.run(
        "UPDATE tasks SET status = ?, completed_at = datetime('now'), updated_at = datetime('now') WHERE story_id = ? AND name = ?",
        ["completed", storyId, name],
      );

      const skillsUsed = input.skills_used.join(",");
      const filesModified = input.files_modified.join(",");

      this.db.run(
        "INSERT INTO task_results (task_id, summary, logs, skills_used, files_modified) VALUES (?, ?, ?, ?, ?)",
        [task.id, input.summary, input.logs, skillsUsed, filesModified],
      );
    });

    tx();
    return this.getByName(storyId, name);
  }

  /**
   * Transitions a task from in_progress to failed, setting failure_reason.
   */
  fail(storyId: number, name: string, reason: string): Task {
    const task = this.getByName(storyId, name);

    if (!canTransition(task.status, "failed")) {
      throw ErrInvalidTransition(task.status, "failed");
    }

    this.db.run(
      "UPDATE tasks SET status = ?, failure_reason = ?, updated_at = datetime('now') WHERE story_id = ? AND name = ?",
      ["failed", reason, storyId, name],
    );

    return this.getByName(storyId, name);
  }

  /**
   * Deletes a task by (storyId, name).
   * Throws ErrNotFound when no matching task exists.
   */
  delete(storyId: number, name: string): void {
    const res = this.db.run(
      "DELETE FROM tasks WHERE story_id = ? AND name = ?",
      [storyId, name],
    );
    if (res.changes === 0) {
      throw ErrNotFound("task", name);
    }
  }

  // --- Private helpers ---

  /**
   * Resolves dependency names to task IDs and inserts edges,
   * performing cycle detection before each insertion.
   */
  private addDependencies(
    storyId: number,
    taskId: number,
    taskName: string,
    dependsOn: string[],
  ): void {
    for (const depName of dependsOn) {
      // Self-dependency check
      if (depName === taskName) {
        throw ErrCyclicDependency(taskName, depName);
      }

      // Resolve dependency name to ID within the same story
      const depRow = this.db
        .query("SELECT id FROM tasks WHERE story_id = ? AND name = ?")
        .get(storyId, depName) as { id: number } | null;

      if (!depRow) {
        throw ErrNotFound("task dependency", depName);
      }

      // Cycle detection
      this.detectCycle(taskId, depRow.id);

      // Insert the dependency edge (idempotent)
      this.db.run(
        "INSERT OR IGNORE INTO task_dependencies (task_id, depends_on) VALUES (?, ?)",
        [taskId, depRow.id],
      );
    }
  }

  /**
   * BFS cycle detection: checks whether adding an edge (taskId depends on depId)
   * would create a cycle. A cycle exists if taskId is reachable from depId
   * by following "depends on" edges.
   */
  private detectCycle(taskId: number, depId: number): void {
    const visited = new Set<number>();
    const queue: number[] = [depId];

    while (queue.length > 0) {
      const current = queue.shift()!;

      if (current === taskId) {
        throw ErrCyclicDependency(taskId, depId);
      }

      if (visited.has(current)) continue;
      visited.add(current);

      const rows = this.db
        .query("SELECT depends_on FROM task_dependencies WHERE task_id = ?")
        .all(current) as { depends_on: number }[];

      for (const row of rows) {
        if (!visited.has(row.depends_on)) {
          queue.push(row.depends_on);
        }
      }
    }
  }
}
