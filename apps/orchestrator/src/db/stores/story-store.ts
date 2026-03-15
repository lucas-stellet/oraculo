import type { Database } from "bun:sqlite";
import type { Story, StorySummary } from "../../domain/entities.ts";
import type { ApprovalStatus } from "../../domain/enums.ts";
import { ErrNotFound, ErrAlreadyExists } from "../../domain/errors.ts";

export class StoryStore {
  constructor(private db: Database) {}

  create(epicId: number, name: string, description: string): Story {
    const stmt = this.db.prepare(
      "INSERT OR IGNORE INTO stories (epic_id, name, description) VALUES (?, ?, ?)",
    );
    const result = stmt.run(epicId, name, description);

    if (result.changes === 0) {
      throw ErrAlreadyExists("story", name);
    }

    return this.getByName(epicId, name);
  }

  getByName(epicId: number, name: string): Story {
    const row = this.db
      .query(
        "SELECT id, epic_id, name, description, approval_status, created_at, updated_at FROM stories WHERE epic_id = ? AND name = ?",
      )
      .get(epicId, name) as Story | null;

    if (!row) {
      throw ErrNotFound("story", name);
    }

    return row;
  }

  list(epicId: number): Story[] {
    return this.db
      .query(
        "SELECT id, epic_id, name, description, approval_status, created_at, updated_at FROM stories WHERE epic_id = ? ORDER BY created_at",
      )
      .all(epicId) as Story[];
  }

  update(epicId: number, name: string, description: string): Story {
    const stmt = this.db.prepare(
      "UPDATE stories SET description = ?, updated_at = datetime('now') WHERE epic_id = ? AND name = ?",
    );
    const result = stmt.run(description, epicId, name);

    if (result.changes === 0) {
      throw ErrNotFound("story", name);
    }

    return this.getByName(epicId, name);
  }

  delete(epicId: number, name: string): void {
    const stmt = this.db.prepare(
      "DELETE FROM stories WHERE epic_id = ? AND name = ?",
    );
    const result = stmt.run(epicId, name);

    if (result.changes === 0) {
      throw ErrNotFound("story", name);
    }
  }

  updateApprovalStatus(epicId: number, name: string, status: ApprovalStatus | "none"): void {
    const stmt = this.db.prepare(
      "UPDATE stories SET approval_status = ?, updated_at = datetime('now') WHERE epic_id = ? AND name = ?",
    );
    const result = stmt.run(status, epicId, name);

    if (result.changes === 0) {
      throw ErrNotFound("story", name);
    }
  }

  listSummaries(epicId: number): StorySummary[] {
    const query = `
      SELECT
        s.id, s.epic_id, s.name, s.description, s.approval_status, s.created_at, s.updated_at,
        COALESCE(tc.total, 0) AS task_count,
        COALESCE(tc.completed, 0) AS completed_task_count,
        COALESCE(tc.failed, 0) AS failed_task_count
      FROM stories s
      LEFT JOIN (
        SELECT story_id,
          COUNT(*) AS total,
          COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed,
          COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed
        FROM tasks GROUP BY story_id
      ) tc ON tc.story_id = s.id
      WHERE s.epic_id = ?
      ORDER BY s.created_at
    `;

    return this.db.query(query).all(epicId) as StorySummary[];
  }
}
