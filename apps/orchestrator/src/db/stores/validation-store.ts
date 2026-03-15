import type { Database } from "bun:sqlite";
import type { Validation } from "../../domain/entities.ts";
import { ErrNotFound } from "../../domain/errors.ts";

export class ValidationStore {
  constructor(private db: Database) {}

  Create(storyId: number, verdict: string, findings: string): Validation {
    this.db.run(
      "INSERT INTO validations (story_id, verdict, findings) VALUES (?, ?, ?)",
      [storyId, verdict, findings],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;
    return this.db
      .query<Validation, [number]>(
        "SELECT id, story_id, task_id, verdict, findings, created_at FROM validations WHERE id = ?",
      )
      .get(id)!;
  }

  Get(id: number): Validation {
    const row = this.db
      .query<Validation, [number]>(
        "SELECT id, story_id, task_id, verdict, findings, created_at FROM validations WHERE id = ?",
      )
      .get(id);
    if (!row) throw ErrNotFound("validation", id);
    return row;
  }

  ListByStory(storyId: number): Validation[] {
    return this.db
      .query<Validation, [number]>(
        "SELECT id, story_id, task_id, verdict, findings, created_at FROM validations WHERE story_id = ? ORDER BY created_at",
      )
      .all(storyId);
  }
}
