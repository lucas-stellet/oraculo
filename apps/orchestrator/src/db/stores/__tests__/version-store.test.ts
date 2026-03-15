import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { VersionStore } from "../version-store.ts";
import { DomainError } from "../../../domain/errors.ts";

describe("VersionStore", () => {
  let db: Database;
  let store: VersionStore;

  beforeEach(() => {
    db = openMemory();
    store = new VersionStore(db);
    db.run("INSERT INTO epics (name, description) VALUES ('test-epic', 'desc')");
    db.run("INSERT INTO stories (epic_id, name, description) VALUES (1, 'test-story', 'desc')");
  });

  describe("EpicVersions", () => {
    test("CreateEpicVersion auto-increments version number", () => {
      const v1 = store.CreateEpicVersion(1, "v1 content");
      expect(v1.number).toBe(1);
      expect(v1.epic_id).toBe(1);
      expect(v1.content).toBe("v1 content");

      const v2 = store.CreateEpicVersion(1, "v2 content");
      expect(v2.number).toBe(2);
    });

    test("CreateEpicVersion sets epic approval_status to pending", () => {
      store.CreateEpicVersion(1, "content");
      const epic = db
        .query<{ approval_status: string }, [number]>(
          "SELECT approval_status FROM epics WHERE id = ?",
        )
        .get(1)!;
      expect(epic.approval_status).toBe("pending");
    });

    test("GetEpicVersion retrieves by id", () => {
      const created = store.CreateEpicVersion(1, "content");
      const fetched = store.GetEpicVersion(created.id);
      expect(fetched.id).toBe(created.id);
      expect(fetched.content).toBe("content");
    });

    test("GetEpicVersion throws not_found for missing id", () => {
      expect(() => store.GetEpicVersion(999)).toThrow(DomainError);
    });

    test("ListEpicVersions returns versions ordered by number", () => {
      store.CreateEpicVersion(1, "v1");
      store.CreateEpicVersion(1, "v2");
      store.CreateEpicVersion(1, "v3");

      const versions = store.ListEpicVersions(1);
      expect(versions).toHaveLength(3);
      expect(versions[0].number).toBe(1);
      expect(versions[1].number).toBe(2);
      expect(versions[2].number).toBe(3);
    });

    test("ListEpicVersions returns empty for no versions", () => {
      expect(store.ListEpicVersions(1)).toHaveLength(0);
    });

    test("LatestEpicVersion returns highest version", () => {
      store.CreateEpicVersion(1, "v1");
      store.CreateEpicVersion(1, "v2");
      const latest = store.LatestEpicVersion(1);
      expect(latest).not.toBeNull();
      expect(latest!.number).toBe(2);
      expect(latest!.content).toBe("v2");
    });

    test("LatestEpicVersion returns null when no versions", () => {
      expect(store.LatestEpicVersion(1)).toBeNull();
    });
  });

  describe("StoryVersions", () => {
    test("CreateStoryVersion auto-increments version number", () => {
      const v1 = store.CreateStoryVersion(1, "v1 content");
      expect(v1.number).toBe(1);
      expect(v1.story_id).toBe(1);

      const v2 = store.CreateStoryVersion(1, "v2 content");
      expect(v2.number).toBe(2);
    });

    test("CreateStoryVersion sets story approval_status to pending", () => {
      store.CreateStoryVersion(1, "content");
      const story = db
        .query<{ approval_status: string }, [number]>(
          "SELECT approval_status FROM stories WHERE id = ?",
        )
        .get(1)!;
      expect(story.approval_status).toBe("pending");
    });

    test("GetStoryVersion retrieves by id", () => {
      const created = store.CreateStoryVersion(1, "content");
      const fetched = store.GetStoryVersion(created.id);
      expect(fetched.id).toBe(created.id);
    });

    test("GetStoryVersion throws not_found for missing id", () => {
      expect(() => store.GetStoryVersion(999)).toThrow(DomainError);
    });

    test("ListStoryVersions returns versions ordered by number", () => {
      store.CreateStoryVersion(1, "v1");
      store.CreateStoryVersion(1, "v2");
      const versions = store.ListStoryVersions(1);
      expect(versions).toHaveLength(2);
      expect(versions[0].number).toBe(1);
      expect(versions[1].number).toBe(2);
    });

    test("LatestStoryVersion returns highest version", () => {
      store.CreateStoryVersion(1, "v1");
      store.CreateStoryVersion(1, "v2");
      const latest = store.LatestStoryVersion(1);
      expect(latest).not.toBeNull();
      expect(latest!.number).toBe(2);
    });

    test("LatestStoryVersion returns null when no versions", () => {
      expect(store.LatestStoryVersion(1)).toBeNull();
    });
  });
});
