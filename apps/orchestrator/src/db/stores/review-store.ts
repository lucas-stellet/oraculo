import type { Database } from "bun:sqlite";
import type { Review } from "../../domain/entities.ts";
import type { ReviewVerdict, VersionType, ApprovalStatus } from "../../domain/enums.ts";
import { ErrNotFound } from "../../domain/errors.ts";

export class ReviewStore {
  constructor(private db: Database) {}

  Create(
    versionId: number,
    type_: VersionType,
    verdict: ReviewVerdict,
    comment: string,
  ): Review {
    this.db.run(
      "INSERT INTO reviews (version_id, version_type, verdict, comment) VALUES (?, ?, ?, ?)",
      [versionId, type_, verdict, comment],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;

    // Map verdict to approval status
    const status: ApprovalStatus = verdict === "approved" ? "approved" : "rejected";

    // Propagate to parent entity
    if (type_ === "epic") {
      const row = this.db
        .query<{ epic_id: number }, [number]>(
          "SELECT epic_id FROM epic_versions WHERE id = ?",
        )
        .get(versionId);
      if (row) {
        this.db.run(
          "UPDATE epics SET approval_status = ?, updated_at = datetime('now') WHERE id = ?",
          [status, row.epic_id],
        );
      }
    } else {
      const row = this.db
        .query<{ story_id: number }, [number]>(
          "SELECT story_id FROM story_versions WHERE id = ?",
        )
        .get(versionId);
      if (row) {
        this.db.run(
          "UPDATE stories SET approval_status = ?, updated_at = datetime('now') WHERE id = ?",
          [status, row.story_id],
        );
      }
    }

    return this.Get(id);
  }

  Get(id: number): Review {
    const row = this.db
      .query<Review, [number]>(
        "SELECT id, version_id, version_type, verdict, comment, created_at FROM reviews WHERE id = ?",
      )
      .get(id);
    if (!row) throw ErrNotFound("review", id);
    return row;
  }

  ListByEpic(epicId: number): Review[] {
    return this.db
      .query<Review, [number]>(
        `SELECT r.id, r.version_id, r.version_type, r.verdict, r.comment, r.created_at
         FROM reviews r
         JOIN epic_versions ev ON r.version_id = ev.id AND r.version_type = 'epic'
         WHERE ev.epic_id = ?
         ORDER BY r.created_at`,
      )
      .all(epicId);
  }

  ListByStory(storyId: number): Review[] {
    return this.db
      .query<Review, [number]>(
        `SELECT r.id, r.version_id, r.version_type, r.verdict, r.comment, r.created_at
         FROM reviews r
         JOIN story_versions sv ON r.version_id = sv.id AND r.version_type = 'story'
         WHERE sv.story_id = ?
         ORDER BY r.created_at`,
      )
      .all(storyId);
  }
}
