import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { ReviewStore } from "../review-store.ts";
import { VersionStore } from "../version-store.ts";
import { DomainError } from "../../../domain/errors.ts";

describe("ReviewStore", () => {
  let db: Database;
  let store: ReviewStore;
  let versionStore: VersionStore;

  beforeEach(() => {
    db = openMemory();
    store = new ReviewStore(db);
    versionStore = new VersionStore(db);
    db.run("INSERT INTO epics (name, description) VALUES ('test-epic', 'desc')");
    db.run("INSERT INTO stories (epic_id, name, description) VALUES (1, 'test-story', 'desc')");
  });

  test("Create inserts a review and returns it", () => {
    const v = versionStore.CreateEpicVersion(1, "content");
    const r = store.Create(v.id, "epic", "approved", "looks good");
    expect(r.id).toBeGreaterThan(0);
    expect(r.version_id).toBe(v.id);
    expect(r.version_type).toBe("epic");
    expect(r.verdict).toBe("approved");
    expect(r.comment).toBe("looks good");
    expect(r.created_at).toBeTruthy();
  });

  test("Create propagates approved verdict to epic approval_status", () => {
    const v = versionStore.CreateEpicVersion(1, "content");
    store.Create(v.id, "epic", "approved", "ok");
    const epic = db
      .query<{ approval_status: string }, [number]>(
        "SELECT approval_status FROM epics WHERE id = ?",
      )
      .get(1)!;
    expect(epic.approval_status).toBe("approved");
  });

  test("Create propagates rejected verdict to epic approval_status", () => {
    const v = versionStore.CreateEpicVersion(1, "content");
    store.Create(v.id, "epic", "rejected", "needs work");
    const epic = db
      .query<{ approval_status: string }, [number]>(
        "SELECT approval_status FROM epics WHERE id = ?",
      )
      .get(1)!;
    expect(epic.approval_status).toBe("rejected");
  });

  test("Create propagates verdict to story approval_status", () => {
    const v = versionStore.CreateStoryVersion(1, "content");
    store.Create(v.id, "story", "approved", "ok");
    const story = db
      .query<{ approval_status: string }, [number]>(
        "SELECT approval_status FROM stories WHERE id = ?",
      )
      .get(1)!;
    expect(story.approval_status).toBe("approved");
  });

  test("Get retrieves a review by id", () => {
    const v = versionStore.CreateEpicVersion(1, "content");
    const created = store.Create(v.id, "epic", "approved", "ok");
    const fetched = store.Get(created.id);
    expect(fetched.id).toBe(created.id);
    expect(fetched.verdict).toBe("approved");
  });

  test("Get throws not_found for missing id", () => {
    expect(() => store.Get(999)).toThrow(DomainError);
  });

  test("ListByEpic returns reviews through epic versions", () => {
    const v1 = versionStore.CreateEpicVersion(1, "v1");
    const v2 = versionStore.CreateEpicVersion(1, "v2");
    store.Create(v1.id, "epic", "rejected", "fix");
    store.Create(v2.id, "epic", "approved", "ok");

    const reviews = store.ListByEpic(1);
    expect(reviews).toHaveLength(2);
    expect(reviews[0].verdict).toBe("rejected");
    expect(reviews[1].verdict).toBe("approved");
  });

  test("ListByEpic returns empty for no reviews", () => {
    expect(store.ListByEpic(1)).toHaveLength(0);
  });

  test("ListByStory returns reviews through story versions", () => {
    const v = versionStore.CreateStoryVersion(1, "content");
    store.Create(v.id, "story", "approved", "ok");

    const reviews = store.ListByStory(1);
    expect(reviews).toHaveLength(1);
    expect(reviews[0].verdict).toBe("approved");
  });

  test("ListByStory returns empty for no reviews", () => {
    expect(store.ListByStory(1)).toHaveLength(0);
  });
});
