import type { Database } from "bun:sqlite";
import type { Approval, ApprovalComment } from "../../domain/entities.ts";
import type { ApprovalType, Verdict } from "../../domain/enums.ts";
import { ErrNotFound, ErrApprovalDecided } from "../../domain/errors.ts";

export class ApprovalStore {
  constructor(private db: Database) {}

  Create(
    type_: ApprovalType,
    epicId: number | null,
    storyId: number | null,
    content: string,
    previousVersion: string = "",
  ): Approval {
    const id = crypto.randomUUID();
    this.db.run(
      `INSERT INTO approvals (id, type, epic_id, story_id, content, previous_version, status)
       VALUES (?, ?, ?, ?, ?, ?, 'pending')`,
      [id, type_, epicId, storyId, content, previousVersion],
    );
    return this.Get(id);
  }

  Get(id: string): Approval {
    const row = this.db
      .query<Approval, [string]>(
        `SELECT id, type, epic_id, story_id, content, previous_version, status,
                verdict_comment, requested_at, decided_at
         FROM approvals WHERE id = ?`,
      )
      .get(id);
    if (!row) throw ErrNotFound("approval", id);
    return row;
  }

  List(epicId?: number): Approval[] {
    if (epicId !== undefined) {
      return this.db
        .query<Approval, [number]>(
          `SELECT id, type, epic_id, story_id, content, previous_version, status,
                  verdict_comment, requested_at, decided_at
           FROM approvals WHERE epic_id = ? ORDER BY requested_at`,
        )
        .all(epicId);
    }
    return this.db
      .query<Approval, []>(
        `SELECT id, type, epic_id, story_id, content, previous_version, status,
                verdict_comment, requested_at, decided_at
         FROM approvals ORDER BY requested_at`,
      )
      .all();
  }

  ListPending(): Approval[] {
    return this.db
      .query<Approval, []>(
        `SELECT id, type, epic_id, story_id, content, previous_version, status,
                verdict_comment, requested_at, decided_at
         FROM approvals WHERE status = 'pending' ORDER BY requested_at`,
      )
      .all();
  }

  SetVerdict(id: string, verdict: Verdict, comment: string = ""): void {
    const current = this.Get(id);
    if (current.status !== "pending") {
      throw ErrApprovalDecided(id);
    }

    let previousVersion = current.previous_version;
    if (verdict === "needs_revision") {
      previousVersion = current.content;
    }

    this.db.run(
      `UPDATE approvals
       SET status = ?, verdict_comment = ?, previous_version = ?, decided_at = datetime('now')
       WHERE id = ?`,
      [verdict, comment, previousVersion, id],
    );

    if (verdict === "approved") {
      this.DeleteCommentsByApproval(id);
    }
  }

  AddComment(
    approvalId: string,
    selectedText: string,
    comment: string,
  ): ApprovalComment {
    this.db.run(
      "INSERT INTO approval_comments (approval_id, selected_text, comment) VALUES (?, ?, ?)",
      [approvalId, selectedText, comment],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;
    return this.db
      .query<ApprovalComment, [number]>(
        "SELECT id, selected_text, comment, created_at FROM approval_comments WHERE id = ?",
      )
      .get(id)!;
  }

  ListComments(approvalId: string): ApprovalComment[] {
    return this.db
      .query<ApprovalComment, [string]>(
        `SELECT id, selected_text, comment, created_at
         FROM approval_comments WHERE approval_id = ? ORDER BY created_at`,
      )
      .all(approvalId);
  }

  DeleteCommentsByApproval(approvalId: string): void {
    this.db.run("DELETE FROM approval_comments WHERE approval_id = ?", [
      approvalId,
    ]);
  }
}
