import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../index.ts";
import { EpicStore } from "../epic-store.ts";
import { DomainError } from "../../../domain/index.ts";

describe("EpicStore", () => {
  let db: Database;
  let store: EpicStore;

  beforeEach(() => {
    db = openMemory();
    store = new EpicStore(db);
  });

  afterEach(() => {
    db.close();
  });

  // --- Create ---

  test("Create inserts a new epic and returns it", () => {
    const epic = store.create("my-epic", "A test epic");

    expect(epic.name).toBe("my-epic");
    expect(epic.description).toBe("A test epic");
    expect(epic.approval_status).toBe("none");
    expect(epic.id).toBeGreaterThan(0);
    expect(epic.created_at).toBeTruthy();
    expect(epic.updated_at).toBeTruthy();
  });

  test("Create throws ErrAlreadyExists on duplicate name", () => {
    store.create("dup-epic", "first");

    expect(() => store.create("dup-epic", "second")).toThrow(DomainError);
    try {
      store.create("dup-epic", "second");
    } catch (err) {
      expect(err).toBeInstanceOf(DomainError);
      expect((err as DomainError).code).toBe("already_exists");
    }
  });

  // --- GetByName ---

  test("GetByName returns the epic", () => {
    store.create("find-me", "desc");
    const epic = store.getByName("find-me");

    expect(epic.name).toBe("find-me");
    expect(epic.description).toBe("desc");
  });

  test("GetByName throws ErrNotFound for missing epic", () => {
    expect(() => store.getByName("nope")).toThrow(DomainError);
    try {
      store.getByName("nope");
    } catch (err) {
      expect(err).toBeInstanceOf(DomainError);
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- List ---

  test("List returns all epics ordered by creation time", () => {
    store.create("alpha", "a");
    store.create("beta", "b");
    store.create("gamma", "c");

    const epics = store.list();
    expect(epics).toHaveLength(3);
    expect(epics.map((e) => e.name)).toEqual(["alpha", "beta", "gamma"]);
  });

  test("List returns empty array when no epics exist", () => {
    const epics = store.list();
    expect(epics).toEqual([]);
  });

  // --- Update ---

  test("Update changes description and returns updated epic", () => {
    store.create("upd-epic", "old desc");
    const updated = store.update("upd-epic", "new desc");

    expect(updated.name).toBe("upd-epic");
    expect(updated.description).toBe("new desc");
  });

  test("Update throws ErrNotFound for missing epic", () => {
    expect(() => store.update("ghost", "desc")).toThrow(DomainError);
    try {
      store.update("ghost", "desc");
    } catch (err) {
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- Delete ---

  test("Delete removes an epic", () => {
    store.create("del-epic", "bye");
    store.delete("del-epic");

    expect(() => store.getByName("del-epic")).toThrow(DomainError);
  });

  test("Delete throws ErrNotFound for missing epic", () => {
    expect(() => store.delete("ghost")).toThrow(DomainError);
    try {
      store.delete("ghost");
    } catch (err) {
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- UpdateApprovalStatus ---

  test("UpdateApprovalStatus sets the status", () => {
    store.create("approval-epic", "desc");
    store.updateApprovalStatus("approval-epic", "approved");

    const epic = store.getByName("approval-epic");
    expect(epic.approval_status).toBe("approved");
  });

  test("UpdateApprovalStatus throws ErrNotFound for missing epic", () => {
    expect(() => store.updateApprovalStatus("ghost", "approved")).toThrow(DomainError);
    try {
      store.updateApprovalStatus("ghost", "approved");
    } catch (err) {
      expect((err as DomainError).code).toBe("not_found");
    }
  });

  // --- ListSummaries ---

  test("ListSummaries returns empty array when no epics", () => {
    const summaries = store.listSummaries();
    expect(summaries).toEqual([]);
  });

  test("ListSummaries returns epic with default phase and counts", () => {
    store.create("summary-epic", "desc");
    const summaries = store.listSummaries();

    expect(summaries).toHaveLength(1);
    expect(summaries[0].name).toBe("summary-epic");
    expect(summaries[0].phase).toBe("discover");
    expect(summaries[0].phase_status).toBe("pending");
    expect(summaries[0].story_count).toBe(0);
    expect(summaries[0].task_count).toBe(0);
    expect(summaries[0].completed_task_count).toBe(0);
  });

  test("ListSummaries aggregates story and task counts", () => {
    store.create("counted-epic", "desc");
    const epic = store.getByName("counted-epic");

    // Insert stories and tasks directly
    db.run("INSERT INTO stories (epic_id, name, description) VALUES (?, 'story-1', 'desc')", [epic.id]);
    db.run("INSERT INTO stories (epic_id, name, description) VALUES (?, 'story-2', 'desc')", [epic.id]);

    const storyRow = db.query("SELECT id FROM stories WHERE name = 'story-1'").get() as { id: number };
    db.run("INSERT INTO tasks (story_id, name, status) VALUES (?, 'task-1', 'completed')", [storyRow.id]);
    db.run("INSERT INTO tasks (story_id, name, status) VALUES (?, 'task-2', 'pending')", [storyRow.id]);

    const summaries = store.listSummaries();
    expect(summaries).toHaveLength(1);
    expect(summaries[0].story_count).toBe(2);
    expect(summaries[0].task_count).toBe(2);
    expect(summaries[0].completed_task_count).toBe(1);
  });

  test("ListSummaries reflects active session phase", () => {
    store.create("session-epic", "desc");
    const epic = store.getByName("session-epic");

    db.run(
      "INSERT INTO sessions (id, type, epic_id, status) VALUES ('s1', 'plan', ?, 'active')",
      [epic.id],
    );

    const summaries = store.listSummaries();
    expect(summaries).toHaveLength(1);
    expect(summaries[0].phase).toBe("plan");
    expect(summaries[0].phase_status).toBe("active");
  });

  test("ListSummaries reflects completed session phase", () => {
    store.create("completed-epic", "desc");
    const epic = store.getByName("completed-epic");

    db.run(
      "INSERT INTO sessions (id, type, epic_id, status, closed_at) VALUES ('s2', 'execute', ?, 'completed', datetime('now'))",
      [epic.id],
    );

    const summaries = store.listSummaries();
    expect(summaries).toHaveLength(1);
    expect(summaries[0].phase).toBe("execute");
    expect(summaries[0].phase_status).toBe("completed");
  });
});
