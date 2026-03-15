import { describe, it, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../db/database.ts";
import { dispatchAgent } from "../dispatch.ts";
import { CODE_AGENT } from "../agents.ts";
import { TaskStore } from "../../db/stores/task-store.ts";
import type { TaskEnriched } from "../../domain/index.ts";

let db: Database;
let taskStore: TaskStore;

beforeEach(() => {
  db = openMemory();
  db.run("INSERT INTO epics (name, description) VALUES ('epic', 'desc')");
  db.run("INSERT INTO stories (epic_id, name, description) VALUES (1, 'story', 'desc')");
  taskStore = new TaskStore(db);
});

function createTestTask(): TaskEnriched {
  const [task] = taskStore.create(1, "test-task", "A test task", []);
  return {
    ...task,
    depends_on: [],
  };
}

describe("dispatchAgent", () => {
  it("completes task on successful run", async () => {
    const task = createTestTask();

    const result = await dispatchAgent({
      task,
      agentDef: CODE_AGENT,
      prompt: "Do something",
      sessionId: "sess-1",
      db,
      runner: async () => "Done successfully",
    });

    expect(result.success).toBe(true);
    expect(result.summary).toBe("Done successfully");
    expect(result.taskId).toBe(task.id);

    // Verify task status in DB
    const updated = taskStore.getByName(1, "test-task");
    expect(updated.status).toBe("completed");
  });

  it("fails task on runner error", async () => {
    const task = createTestTask();

    const result = await dispatchAgent({
      task,
      agentDef: CODE_AGENT,
      prompt: "Do something",
      sessionId: "sess-1",
      db,
      runner: async () => {
        throw new Error("Agent crashed");
      },
    });

    expect(result.success).toBe(false);
    expect(result.summary).toBe("Agent crashed");

    // Verify task status in DB
    const updated = taskStore.getByName(1, "test-task");
    expect(updated.status).toBe("failed");
  });

  it("creates agent record", async () => {
    const task = createTestTask();

    await dispatchAgent({
      task,
      agentDef: CODE_AGENT,
      prompt: "Do something",
      sessionId: "sess-1",
      db,
      runner: async () => "Done",
    });

    const agents = db
      .query("SELECT * FROM agents WHERE session_id = ?")
      .all("sess-1") as Array<{ name: string; type: string; status: string }>;

    expect(agents).toHaveLength(1);
    expect(agents[0].name).toBe("code-agent");
    expect(agents[0].type).toBe("code");
    expect(agents[0].status).toBe("completed");
  });

  it("calls onEvent callback", async () => {
    const task = createTestTask();
    const events: unknown[] = [];

    await dispatchAgent({
      task,
      agentDef: CODE_AGENT,
      prompt: "Do something",
      sessionId: "sess-1",
      db,
      onEvent: (e) => events.push(e),
      runner: async (opts) => {
        opts.onEvent?.("test-event");
        return "Done";
      },
    });

    expect(events).toContain("test-event");
  });
});
