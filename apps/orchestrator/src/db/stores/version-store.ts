import type { Database } from "bun:sqlite";
import type { EpicVersion, StoryVersion } from "../../domain/entities.ts";
import { ErrNotFound } from "../../domain/errors.ts";

export class VersionStore {
  constructor(private db: Database) {}

  CreateEpicVersion(epicId: number, content: string): EpicVersion {
    const nextNum = this.db
      .query<{ n: number }, [number]>(
        "SELECT COALESCE(MAX(number), 0) + 1 as n FROM epic_versions WHERE epic_id = ?",
      )
      .get(epicId)!.n;

    this.db.run(
      "INSERT INTO epic_versions (epic_id, number, content) VALUES (?, ?, ?)",
      [epicId, nextNum, content],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;

    this.db.run(
      "UPDATE epics SET approval_status = 'pending', updated_at = datetime('now') WHERE id = ?",
      [epicId],
    );

    return this.GetEpicVersion(id);
  }

  CreateStoryVersion(storyId: number, content: string): StoryVersion {
    const nextNum = this.db
      .query<{ n: number }, [number]>(
        "SELECT COALESCE(MAX(number), 0) + 1 as n FROM story_versions WHERE story_id = ?",
      )
      .get(storyId)!.n;

    this.db.run(
      "INSERT INTO story_versions (story_id, number, content) VALUES (?, ?, ?)",
      [storyId, nextNum, content],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;

    this.db.run(
      "UPDATE stories SET approval_status = 'pending', updated_at = datetime('now') WHERE id = ?",
      [storyId],
    );

    return this.GetStoryVersion(id);
  }

  GetEpicVersion(id: number): EpicVersion {
    const row = this.db
      .query<EpicVersion, [number]>(
        "SELECT id, epic_id, number, content, created_at FROM epic_versions WHERE id = ?",
      )
      .get(id);
    if (!row) throw ErrNotFound("epic_version", id);
    return row;
  }

  GetStoryVersion(id: number): StoryVersion {
    const row = this.db
      .query<StoryVersion, [number]>(
        "SELECT id, story_id, number, content, created_at FROM story_versions WHERE id = ?",
      )
      .get(id);
    if (!row) throw ErrNotFound("story_version", id);
    return row;
  }

  ListEpicVersions(epicId: number): EpicVersion[] {
    return this.db
      .query<EpicVersion, [number]>(
        "SELECT id, epic_id, number, content, created_at FROM epic_versions WHERE epic_id = ? ORDER BY number",
      )
      .all(epicId);
  }

  ListStoryVersions(storyId: number): StoryVersion[] {
    return this.db
      .query<StoryVersion, [number]>(
        "SELECT id, story_id, number, content, created_at FROM story_versions WHERE story_id = ? ORDER BY number",
      )
      .all(storyId);
  }

  LatestEpicVersion(epicId: number): EpicVersion | null {
    return (
      this.db
        .query<EpicVersion, [number]>(
          "SELECT id, epic_id, number, content, created_at FROM epic_versions WHERE epic_id = ? ORDER BY number DESC LIMIT 1",
        )
        .get(epicId) ?? null
    );
  }

  LatestStoryVersion(storyId: number): StoryVersion | null {
    return (
      this.db
        .query<StoryVersion, [number]>(
          "SELECT id, story_id, number, content, created_at FROM story_versions WHERE story_id = ? ORDER BY number DESC LIMIT 1",
        )
        .get(storyId) ?? null
    );
  }
}
