import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { AgentStore } from "../agent-store.ts";

describe("AgentStore", () => {
  let db: Database;
  let store: AgentStore;

  beforeEach(() => {
    db = openMemory();
    store = new AgentStore(db);
  });

  test("Create creates an active agent", () => {
    const a = store.Create("session-1", "code-agent", "code");
    expect(a.id).toBeGreaterThan(0);
    expect(a.session_id).toBe("session-1");
    expect(a.name).toBe("code-agent");
    expect(a.type).toBe("code");
    expect(a.status).toBe("active");
    expect(a.started_at).toBeTruthy();
    expect(a.stopped_at).toBeNull();
    expect(a.task_id).toBeNull();
  });

  test("Create with taskId associates agent to task", () => {
    db.run("INSERT INTO epics (name) VALUES ('e')");
    db.run("INSERT INTO stories (epic_id, name) VALUES (1, 's')");
    db.run("INSERT INTO tasks (story_id, name) VALUES (1, 't')");
    const a = store.Create("session-1", "code-agent", "code", 1);
    expect(a.task_id).toBe(1);
  });

  test("Stop marks agent as completed", () => {
    const a = store.Create("session-1", "agent", "code");
    store.Stop(a.id);
    const agents = store.ListBySession("session-1");
    expect(agents[0].status).toBe("completed");
    expect(agents[0].stopped_at).toBeTruthy();
  });

  test("ListBySession returns agents for a session", () => {
    store.Create("session-1", "agent-1", "code");
    store.Create("session-1", "agent-2", "research");
    store.Create("session-2", "agent-3", "qa");

    const agents = store.ListBySession("session-1");
    expect(agents).toHaveLength(2);
    expect(agents[0].name).toBe("agent-1");
    expect(agents[1].name).toBe("agent-2");
  });

  test("Active returns only active agents", () => {
    const a1 = store.Create("session-1", "agent-1", "code");
    store.Create("session-1", "agent-2", "research");
    store.Stop(a1.id);

    const active = store.Active();
    expect(active).toHaveLength(1);
    expect(active[0].name).toBe("agent-2");
  });

  test("Active returns empty when no active agents", () => {
    const a = store.Create("session-1", "agent-1", "code");
    store.Stop(a.id);
    expect(store.Active()).toHaveLength(0);
  });
});
