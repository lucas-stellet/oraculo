import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { SessionEventStore } from "../session-event-store.ts";

describe("SessionEventStore", () => {
  let db: Database;
  let store: SessionEventStore;

  beforeEach(() => {
    db = openMemory();
    store = new SessionEventStore(db);
  });

  test("Create inserts and returns a session event", () => {
    const e = store.Create("cs-1", "task_completed", '{"task_id":1}');
    expect(e.id).toBeGreaterThan(0);
    expect(e.session_id).toBe("cs-1");
    expect(e.event_type).toBe("task_completed");
    expect(e.payload).toBe('{"task_id":1}');
    expect(e.created_at).toBeTruthy();
  });

  test("Create auto-creates claude_session if missing", () => {
    store.Create("cs-new", "stop", "{}");
    const row = db
      .query<{ id: string }, [string]>("SELECT id FROM claude_sessions WHERE id = ?")
      .get("cs-new");
    expect(row).not.toBeNull();
  });

  test("ListBySession returns events ordered by created_at", () => {
    store.Create("cs-1", "task_completed", '{"task_id":1}');
    store.Create("cs-1", "stop", "{}");
    store.Create("cs-2", "session_end", "{}");

    const events = store.ListBySession("cs-1");
    expect(events).toHaveLength(2);
    expect(events[0].event_type).toBe("task_completed");
    expect(events[1].event_type).toBe("stop");
  });

  test("ListBySession returns empty for unknown session", () => {
    expect(store.ListBySession("nonexistent")).toHaveLength(0);
  });
});
