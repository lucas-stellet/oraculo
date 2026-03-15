import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../index.ts";
import { EpicStore } from "../epic-store.ts";
import { StoryStore } from "../story-store.ts";
import { DomainError } from "../../../domain/index.ts";

describe("StoryStore", () => {
  let db: Database;
  let epicStore: EpicStore;
  let store: StoryStore;
  let epicId: number;

  beforeEach(() => {
    db = openMemory();
    epicStore = new EpicStore(db);
    store = new StoryStore(db);
    // Create a parent epic for all story tests
    const epic = epicStore.create("test-epic", "parent epic");
    epicId = epic.id;
  });

  afterEach(() => {
    db.close();
  });

  // --- Create ---

  test("Create inserts a new story and returns it", () => {
    const story = store.create(epicId, "my-story", "A test story");

    expect(story.name).toBe("my-story");
    expect(story.description).toBe("A test story");
    expect(story.epic_id).toBe(epicId);
    expect(story.approval_status).toBe("none");
    expect(story.id).toBeGreaterThan(0);
    expect(story.created_at).toBeTruthy();
    expect(story.updated_at).toBeTruthy();
  });

  test("Create throws ErrAlreadyExists on duplicate (epicId, name)", () => {
    store.create(epicId, "dup-story", "first");

    expect(() => store.create(epicId, "dup-story", "second")).toThrow(DomainError);
    try {
      store.create(epicId, "dup-story", "second");
    } catch (err) {
      expect(err).toBeInstanceOf(DomainError);
      expect((err as DomainError).code).toBe("already_exists");
    }
  });

  test("Create allows same name under different epics", () => {
    const epic2 = epicStore.create("other-epic", "another");
    store.create(epicId, "shared-name", "desc1");
    const story2 = store.create(epic2.id, "shared-name", "desc2");

    expect(story2.name).toBe("shared-name");
    expect(story2.epic_id).toBe(epic2.id);
  });

  // --- GetByName ---

  test("GetByName returns the story", () => {
    store.create(epicId, "find-me", "desc");
    const story = store.getByName(epicId, "find-me");

    expect(story.name).toBe("find-me");
    expect(story.description).toBe("desc");
    expect(story.epic_id).toBe(epicId);
  });

  test("GetByName throws ErrNotFound for missing story", () => {
    expect(() => store.getByName(epicId, "nope")).toThrow(DomainError);
    try {
      store.getByName(epicId, "nope");
    } catch (err) {
      expect(err).toBeInstanceOf(DomainError);
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- List ---

  test("List returns all stories for an epic ordered by creation time", () => {
    store.create(epicId, "alpha", "a");
    store.create(epicId, "beta", "b");
    store.create(epicId, "gamma", "c");

    const stories = store.list(epicId);
    expect(stories).toHaveLength(3);
    expect(stories.map((s) => s.name)).toEqual(["alpha", "beta", "gamma"]);
  });

  test("List returns empty array when no stories exist for epic", () => {
    const stories = store.list(epicId);
    expect(stories).toEqual([]);
  });

  test("List only returns stories for the given epic", () => {
    const epic2 = epicStore.create("other-epic", "other");
    store.create(epicId, "story-a", "a");
    store.create(epic2.id, "story-b", "b");

    const stories = store.list(epicId);
    expect(stories).toHaveLength(1);
    expect(stories[0].name).toBe("story-a");
  });

  // --- Update ---

  test("Update changes description and returns updated story", () => {
    store.create(epicId, "upd-story", "old desc");
    const updated = store.update(epicId, "upd-story", "new desc");

    expect(updated.name).toBe("upd-story");
    expect(updated.description).toBe("new desc");
  });

  test("Update throws ErrNotFound for missing story", () => {
    expect(() => store.update(epicId, "ghost", "desc")).toThrow(DomainError);
    try {
      store.update(epicId, "ghost", "desc");
    } catch (err) {
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- Delete ---

  test("Delete removes a story", () => {
    store.create(epicId, "del-story", "bye");
    store.delete(epicId, "del-story");

    expect(() => store.getByName(epicId, "del-story")).toThrow(DomainError);
  });

  test("Delete throws ErrNotFound for missing story", () => {
    expect(() => store.delete(epicId, "ghost")).toThrow(DomainError);
    try {
      store.delete(epicId, "ghost");
    } catch (err) {
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- UpdateApprovalStatus ---

  test("UpdateApprovalStatus sets the status", () => {
    store.create(epicId, "approval-story", "desc");
    store.updateApprovalStatus(epicId, "approval-story", "approved");

    const story = store.getByName(epicId, "approval-story");
    expect(story.approval_status).toBe("approved");
  });

  test("UpdateApprovalStatus throws ErrNotFound for missing story", () => {
    expect(() => store.updateApprovalStatus(epicId, "ghost", "approved")).toThrow(DomainError);
    try {
      store.updateApprovalStatus(epicId, "ghost", "approved");
    } catch (err) {
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- ListSummaries ---

  test("ListSummaries returns empty array when no stories", () => {
    const summaries = store.listSummaries(epicId);
    expect(summaries).toEqual([]);
  });

  test("ListSummaries returns story with default task counts", () => {
    store.create(epicId, "summary-story", "desc");
    const summaries = store.listSummaries(epicId);

    expect(summaries).toHaveLength(1);
    expect(summaries[0].name).toBe("summary-story");
    expect(summaries[0].task_count).toBe(0);
    expect(summaries[0].completed_task_count).toBe(0);
    expect(summaries[0].failed_task_count).toBe(0);
  });

  test("ListSummaries aggregates task counts", () => {
    store.create(epicId, "counted-story", "desc");
    const story = store.getByName(epicId, "counted-story");

    db.run("INSERT INTO tasks (story_id, name, status) VALUES (?, 'task-1', 'completed')", [story.id]);
    db.run("INSERT INTO tasks (story_id, name, status) VALUES (?, 'task-2', 'failed')", [story.id]);
    db.run("INSERT INTO tasks (story_id, name, status) VALUES (?, 'task-3', 'pending')", [story.id]);

    const summaries = store.listSummaries(epicId);
    expect(summaries).toHaveLength(1);
    expect(summaries[0].task_count).toBe(3);
    expect(summaries[0].completed_task_count).toBe(1);
    expect(summaries[0].failed_task_count).toBe(1);
  });

  test("ListSummaries only returns stories for the given epic", () => {
    const epic2 = epicStore.create("other-epic", "other");
    store.create(epicId, "story-a", "a");
    store.create(epic2.id, "story-b", "b");

    const summaries = store.listSummaries(epicId);
    expect(summaries).toHaveLength(1);
    expect(summaries[0].name).toBe("story-a");
  });
});
