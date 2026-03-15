import type { Database } from "bun:sqlite";
import type { Epic, EpicSummary } from "../../domain/entities.ts";
import type { ApprovalStatus, SessionType } from "../../domain/enums.ts";
import { SESSION_TYPE_TO_PHASE } from "../../domain/enums.ts";
import { ErrNotFound, ErrAlreadyExists } from "../../domain/errors.ts";

export class EpicStore {
  constructor(private db: Database) {}

  create(name: string, description: string): Epic {
    const stmt = this.db.prepare(
      "INSERT OR IGNORE INTO epics (name, description) VALUES (?, ?)",
    );
    const result = stmt.run(name, description);

    if (result.changes === 0) {
      throw ErrAlreadyExists("epic", name);
    }

    return this.getByName(name);
  }

  getByName(name: string): Epic {
    const row = this.db
      .query(
        "SELECT id, name, description, approval_status, created_at, updated_at FROM epics WHERE name = ?",
      )
      .get(name) as Epic | null;

    if (!row) {
      throw ErrNotFound("epic", name);
    }

    return row;
  }

  list(): Epic[] {
    return this.db
      .query(
        "SELECT id, name, description, approval_status, created_at, updated_at FROM epics ORDER BY created_at",
      )
      .all() as Epic[];
  }

  update(name: string, description: string): Epic {
    const stmt = this.db.prepare(
      "UPDATE epics SET description = ?, updated_at = datetime('now') WHERE name = ?",
    );
    const result = stmt.run(description, name);

    if (result.changes === 0) {
      throw ErrNotFound("epic", name);
    }

    return this.getByName(name);
  }

  delete(name: string): void {
    const stmt = this.db.prepare("DELETE FROM epics WHERE name = ?");
    const result = stmt.run(name);

    if (result.changes === 0) {
      throw ErrNotFound("epic", name);
    }
  }

  updateApprovalStatus(name: string, status: ApprovalStatus | "none"): void {
    const stmt = this.db.prepare(
      "UPDATE epics SET approval_status = ?, updated_at = datetime('now') WHERE name = ?",
    );
    const result = stmt.run(status, name);

    if (result.changes === 0) {
      throw ErrNotFound("epic", name);
    }
  }

  listSummaries(): EpicSummary[] {
    const query = `
      SELECT
        e.id, e.name, e.description, e.approval_status, e.created_at, e.updated_at,
        COALESCE(active_s.type, last_s.type, '') AS session_type,
        CASE
          WHEN active_s.id IS NOT NULL THEN 'active'
          WHEN last_s.id IS NOT NULL THEN 'completed'
          ELSE 'pending'
        END AS phase_status,
        COALESCE(story_counts.cnt, 0) AS story_count,
        COALESCE(task_counts.total, 0) AS task_count,
        COALESCE(task_counts.completed, 0) AS completed_task_count
      FROM epics e
      LEFT JOIN sessions active_s
        ON active_s.epic_id = e.id AND active_s.status = 'active'
      LEFT JOIN (
        SELECT epic_id, id, type
        FROM sessions
        WHERE status = 'completed'
        ORDER BY closed_at DESC
      ) last_s ON last_s.epic_id = e.id AND active_s.id IS NULL
      LEFT JOIN (
        SELECT epic_id, COUNT(*) AS cnt
        FROM stories
        GROUP BY epic_id
      ) story_counts ON story_counts.epic_id = e.id
      LEFT JOIN (
        SELECT s.epic_id,
          COUNT(t.id) AS total,
          COUNT(CASE WHEN t.status = 'completed' THEN 1 END) AS completed
        FROM tasks t
        JOIN stories s ON t.story_id = s.id
        GROUP BY s.epic_id
      ) task_counts ON task_counts.epic_id = e.id
      ORDER BY e.created_at
    `;

    const rows = this.db.query(query).all() as Array<
      Epic & {
        session_type: string;
        phase_status: string;
        story_count: number;
        task_count: number;
        completed_task_count: number;
      }
    >;

    return rows.map((row) => {
      const { session_type, ...rest } = row;
      let phase: string;
      if (session_type) {
        phase = SESSION_TYPE_TO_PHASE[session_type as SessionType] ?? "discover";
      } else {
        phase = "discover";
      }
      return { ...rest, phase };
    });
  }
}
