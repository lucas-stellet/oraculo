import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { TaskStore } from "../task-store.ts";
import { DomainError } from "../../../domain/errors.ts";

function expectDomainError(fn: () => unknown, code: string): void {
  try {
    fn();
    throw new Error("Expected DomainError but no error was thrown");
  } catch (err) {
    if (!(err instanceof DomainError)) {
      throw new Error(
        `Expected DomainError with code '${code}' but got: ${err}`,
      );
    }
    expect(err.code).toBe(code);
  }
}

let db: Database;
let store: TaskStore;
const STORY_ID = 1;
const STORY_ID_2 = 2;

beforeEach(() => {
  db = openMemory();
  // Seed an epic and two stories for tests
  db.run("INSERT INTO epics (name, description) VALUES ('test-epic', 'desc')");
  db.run(
    "INSERT INTO stories (epic_id, name, description) VALUES (1, 'test-story', 'desc')",
  );
  db.run(
    "INSERT INTO stories (epic_id, name, description) VALUES (1, 'other-story', 'desc')",
  );
  store = new TaskStore(db);
});

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------
describe("Create", () => {
  test("creates a task with no dependencies", () => {
    const [task, created] = store.create(STORY_ID, "task-a", "do A", []);
    expect(created).toBe(true);
    expect(task.id).toBeGreaterThan(0);
    expect(task.name).toBe("task-a");
    expect(task.description).toBe("do A");
    expect(task.story_id).toBe(STORY_ID);
    expect(task.status).toBe("pending");
    expect(task.failure_reason).toBe("");
    expect(task.started_at).toBeUndefined();
    expect(task.completed_at).toBeUndefined();
  });

  test("returns existing task when name already exists (created=false)", () => {
    const [first, c1] = store.create(STORY_ID, "task-a", "do A", []);
    const [second, c2] = store.create(STORY_ID, "task-a", "do A v2", []);
    expect(c1).toBe(true);
    expect(c2).toBe(false);
    expect(second.id).toBe(first.id);
  });

  test("creates a task with valid dependencies", () => {
    store.create(STORY_ID, "dep-1", "dependency 1", []);
    store.create(STORY_ID, "dep-2", "dependency 2", []);
    const [task] = store.create(
      STORY_ID,
      "task-b",
      "depends on dep-1 and dep-2",
      ["dep-1", "dep-2"],
    );
    expect(task.id).toBeGreaterThan(0);
  });

  test("fails when dependency does not exist", () => {
    expectDomainError(
      () => store.create(STORY_ID, "task-x", "desc", ["nonexistent"]),
      "not_found",
    );
  });

  test("fails on self-dependency", () => {
    expectDomainError(
      () => store.create(STORY_ID, "self-ref", "desc", ["self-ref"]),
      "cyclic_dependency",
    );
  });

  test("detects cycle A->B->C->A via re-create with new deps", () => {
    // Create chain: ca (no deps), cb depends on ca, cc depends on cb
    store.create(STORY_ID, "ca", "A", []);
    store.create(STORY_ID, "cb", "B", ["ca"]);
    store.create(STORY_ID, "cc", "C", ["cb"]);
    // Re-"create" ca with deps=["cc"]. Since ca already exists (INSERT OR IGNORE),
    // it will try to add ca->cc. But cc->cb->ca means ca is reachable from cc. CYCLE!
    expectDomainError(
      () => store.create(STORY_ID, "ca", "A again", ["cc"]),
      "cyclic_dependency",
    );
  });

  test("detects direct cycle A->B->A", () => {
    store.create(STORY_ID, "x", "X", []);
    store.create(STORY_ID, "y", "Y", ["x"]);
    // Re-"create" x with dep on y: x->y->x = cycle
    expectDomainError(
      () => store.create(STORY_ID, "x", "X", ["y"]),
      "cyclic_dependency",
    );
  });

  test("no false positive: valid diamond dependency", () => {
    // a -> b, a -> c, b -> d, c -> d (diamond, no cycle)
    store.create(STORY_ID, "d", "D", []);
    store.create(STORY_ID, "b", "B", ["d"]);
    store.create(STORY_ID, "c", "C", ["d"]);
    const [task] = store.create(STORY_ID, "a", "A", ["b", "c"]);
    expect(task.id).toBeGreaterThan(0);
  });
});

// ---------------------------------------------------------------------------
// GetByName
// ---------------------------------------------------------------------------
describe("GetByName", () => {
  test("returns existing task", () => {
    store.create(STORY_ID, "findme", "desc", []);
    const task = store.getByName(STORY_ID, "findme");
    expect(task.name).toBe("findme");
    expect(task.story_id).toBe(STORY_ID);
  });

  test("throws ErrNotFound for missing task", () => {
    expectDomainError(
      () => store.getByName(STORY_ID, "nope"),
      "not_found",
    );
  });
});

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------
describe("List", () => {
  test("returns empty array when no tasks", () => {
    const tasks = store.list(STORY_ID);
    expect(tasks).toEqual([]);
  });

  test("returns tasks ordered by created_at", () => {
    store.create(STORY_ID, "first", "1", []);
    store.create(STORY_ID, "second", "2", []);
    store.create(STORY_ID, "third", "3", []);
    const tasks = store.list(STORY_ID);
    expect(tasks).toHaveLength(3);
    expect(tasks.map((t) => t.name)).toEqual(["first", "second", "third"]);
  });

  test("only returns tasks for the given story", () => {
    store.create(STORY_ID, "s1-task", "desc", []);
    store.create(STORY_ID_2, "s2-task", "desc", []);
    const tasks = store.list(STORY_ID);
    expect(tasks).toHaveLength(1);
    expect(tasks[0].name).toBe("s1-task");
  });
});

// ---------------------------------------------------------------------------
// ListEnriched
// ---------------------------------------------------------------------------
describe("ListEnriched", () => {
  test("returns empty array when no tasks", () => {
    const enriched = store.listEnriched(STORY_ID);
    expect(enriched).toEqual([]);
  });

  test("includes dependencies", () => {
    store.create(STORY_ID, "dep", "dep", []);
    store.create(STORY_ID, "main", "main", ["dep"]);
    const enriched = store.listEnriched(STORY_ID);
    expect(enriched).toHaveLength(2);
    const main = enriched.find((t) => t.name === "main")!;
    const dep = enriched.find((t) => t.name === "dep")!;
    expect(main.depends_on).toContain(dep.id);
    expect(dep.depends_on).toEqual([]);
  });

  test("includes task result after completion", () => {
    store.create(STORY_ID, "task-r", "desc", []);
    store.start(STORY_ID, "task-r");
    store.complete(STORY_ID, "task-r", {
      summary: "done",
      logs: "log output",
      skills_used: ["tdd", "refactor"],
      files_modified: ["a.ts", "b.ts"],
    });
    const enriched = store.listEnriched(STORY_ID);
    expect(enriched).toHaveLength(1);
    const t = enriched[0];
    expect(t.result).toBeDefined();
    expect(t.result!.summary).toBe("done");
    expect(t.result!.logs).toBe("log output");
    expect(t.result!.skills_used).toEqual(["tdd", "refactor"]);
    expect(t.result!.files_modified).toEqual(["a.ts", "b.ts"]);
  });
});

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------
describe("Start", () => {
  test("transitions pending -> in_progress", () => {
    store.create(STORY_ID, "to-start", "desc", []);
    const task = store.start(STORY_ID, "to-start");
    expect(task.status).toBe("in_progress");
    expect(task.started_at).toBeDefined();
  });

  test("throws ErrInvalidTransition for non-pending task", () => {
    store.create(STORY_ID, "started", "desc", []);
    store.start(STORY_ID, "started");
    expectDomainError(
      () => store.start(STORY_ID, "started"),
      "invalid_transition",
    );
  });

  test("throws ErrNotFound for missing task", () => {
    expectDomainError(
      () => store.start(STORY_ID, "nope"),
      "not_found",
    );
  });
});

// ---------------------------------------------------------------------------
// Complete
// ---------------------------------------------------------------------------
describe("Complete", () => {
  test("transitions in_progress -> completed with result", () => {
    store.create(STORY_ID, "to-complete", "desc", []);
    store.start(STORY_ID, "to-complete");
    const task = store.complete(STORY_ID, "to-complete", {
      summary: "all done",
      logs: "build passed",
      skills_used: ["tdd"],
      files_modified: ["src/main.ts"],
    });
    expect(task.status).toBe("completed");
    expect(task.completed_at).toBeDefined();

    // Verify result was inserted
    const enriched = store.listEnriched(STORY_ID);
    const t = enriched.find((e) => e.name === "to-complete")!;
    expect(t.result).toBeDefined();
    expect(t.result!.summary).toBe("all done");
  });

  test("throws ErrInvalidTransition for pending task", () => {
    store.create(STORY_ID, "pending-task", "desc", []);
    expectDomainError(
      () =>
        store.complete(STORY_ID, "pending-task", {
          summary: "s",
          logs: "",
          skills_used: [],
          files_modified: [],
        }),
      "invalid_transition",
    );
  });

  test("handles empty skills_used and files_modified", () => {
    store.create(STORY_ID, "empty-result", "desc", []);
    store.start(STORY_ID, "empty-result");
    store.complete(STORY_ID, "empty-result", {
      summary: "done",
      logs: "",
      skills_used: [],
      files_modified: [],
    });
    const enriched = store.listEnriched(STORY_ID);
    const t = enriched.find((e) => e.name === "empty-result")!;
    expect(t.result!.skills_used).toEqual([]);
    expect(t.result!.files_modified).toEqual([]);
  });
});

// ---------------------------------------------------------------------------
// Fail
// ---------------------------------------------------------------------------
describe("Fail", () => {
  test("transitions in_progress -> failed", () => {
    store.create(STORY_ID, "to-fail", "desc", []);
    store.start(STORY_ID, "to-fail");
    const task = store.fail(STORY_ID, "to-fail", "compilation error");
    expect(task.status).toBe("failed");
    expect(task.failure_reason).toBe("compilation error");
  });

  test("throws ErrInvalidTransition for pending task", () => {
    store.create(STORY_ID, "pending-fail", "desc", []);
    expectDomainError(
      () => store.fail(STORY_ID, "pending-fail", "reason"),
      "invalid_transition",
    );
  });
});

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------
describe("Delete", () => {
  test("deletes an existing task", () => {
    store.create(STORY_ID, "to-delete", "desc", []);
    store.delete(STORY_ID, "to-delete");
    expectDomainError(
      () => store.getByName(STORY_ID, "to-delete"),
      "not_found",
    );
  });

  test("throws ErrNotFound for missing task", () => {
    expectDomainError(
      () => store.delete(STORY_ID, "nope"),
      "not_found",
    );
  });
});
