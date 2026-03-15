import { describe, it, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../db/database.ts";
import { executeStory } from "../execute.ts";
import { TaskStore } from "../../db/stores/task-store.ts";

let db: Database;
let taskStore: TaskStore;

beforeEach(() => {
  db = openMemory();
  db.run("INSERT INTO epics (name, description) VALUES ('epic', 'desc')");
  db.run("INSERT INTO stories (epic_id, name, description) VALUES (1, 'story', 'desc')");
  taskStore = new TaskStore(db);
});

describe("executeStory", () => {
  it("executes independent tasks concurrently", async () => {
    taskStore.create(1, "task-a", "First task", []);
    taskStore.create(1, "task-b", "Second task", []);

    const result = await executeStory({
      storyId: 1,
      sessionId: "sess-1",
      epicName: "epic",
      storyName: "story",
      db,
      runner: async () => "Done",
    });

    expect(result.success).toBe(true);
    expect(result.results).toHaveLength(2);
    expect(result.results.every((r) => r.success)).toBe(true);
  });

  it("executes dependent tasks in order", async () => {
    taskStore.create(1, "task-a", "First task", []);
    taskStore.create(1, "task-b", "Depends on A", ["task-a"]);

    const order: string[] = [];
    const result = await executeStory({
      storyId: 1,
      sessionId: "sess-1",
      epicName: "epic",
      storyName: "story",
      db,
      runner: async (opts) => {
        // Extract task name from prompt
        const match = opts.prompt.match(/# Task: (.+)/);
        if (match) order.push(match[1]);
        return "Done";
      },
    });

    expect(result.success).toBe(true);
    expect(order).toEqual(["task-a", "task-b"]);
  });

  it("triggers circuit breaker after max failures", async () => {
    taskStore.create(1, "task-a", "Will fail", []);
    taskStore.create(1, "task-b", "Will fail too", []);
    taskStore.create(1, "task-c", "Will fail three", []);

    const result = await executeStory({
      storyId: 1,
      sessionId: "sess-1",
      epicName: "epic",
      storyName: "story",
      db,
      maxFailures: 2,
      runner: async () => {
        throw new Error("boom");
      },
    });

    expect(result.success).toBe(false);
    // Should stop after 2 consecutive failures (circuit breaker)
    // All 3 tasks are independent, so they run in one wave.
    // After wave completes with 3 failures, consecutive count reaches 3 >= 2.
    expect(result.results.length).toBeGreaterThanOrEqual(2);
  });

  it("detects deadlock", async () => {
    // Create tasks with circular dependency manually
    db.run("INSERT INTO tasks (story_id, name, description) VALUES (1, 'task-x', 'x')");
    db.run("INSERT INTO tasks (story_id, name, description) VALUES (1, 'task-y', 'y')");
    // task-x depends on task-y and vice versa
    db.run("INSERT INTO task_dependencies (task_id, depends_on) VALUES (1, 2)");
    db.run("INSERT INTO task_dependencies (task_id, depends_on) VALUES (2, 1)");

    const result = await executeStory({
      storyId: 1,
      sessionId: "sess-1",
      epicName: "epic",
      storyName: "story",
      db,
      runner: async () => "Done",
    });

    expect(result.success).toBe(false);
    expect(result.results).toHaveLength(0);
  });

  it("succeeds with no tasks", async () => {
    const result = await executeStory({
      storyId: 1,
      sessionId: "sess-1",
      epicName: "epic",
      storyName: "story",
      db,
      runner: async () => "Done",
    });

    expect(result.success).toBe(true);
    expect(result.results).toHaveLength(0);
  });

  it("resets consecutive failures on success", async () => {
    // task-a fails, task-b succeeds, task-c (depends on b) fails
    // consecutive failures: 1 (a), 0 (b succeeds resets), 1 (c)
    taskStore.create(1, "task-a", "Will fail", []);
    taskStore.create(1, "task-b", "Will succeed", []);
    taskStore.create(1, "task-c", "Depends on b, will fail", ["task-b"]);

    let callCount = 0;
    const result = await executeStory({
      storyId: 1,
      sessionId: "sess-1",
      epicName: "epic",
      storyName: "story",
      db,
      maxFailures: 2,
      runner: async (opts) => {
        callCount++;
        if (opts.prompt.includes("task-b")) return "Done";
        throw new Error("fail");
      },
    });

    // Should not circuit-break because success resets counter
    // Wave 1: task-a (fail), task-b (success) => consecutive resets to 0 after b
    // Wave 2: task-c (fail) => consecutive = 1, below threshold of 2
    // No more pending tasks (a=failed, b=completed, c=failed) => loop exits
    expect(callCount).toBe(3);
    // All tasks have been processed (none pending), so the loop considers it done.
    // Individual failures are tracked in results.
    expect(result.success).toBe(true);
    expect(result.results.filter((r) => !r.success)).toHaveLength(2);
  });
});
