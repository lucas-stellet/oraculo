import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { ToolEventStore } from "../tool-event-store.ts";

describe("ToolEventStore", () => {
  let db: Database;
  let store: ToolEventStore;

  beforeEach(() => {
    db = openMemory();
    store = new ToolEventStore(db);
  });

  test("Create inserts and returns a tool event", () => {
    const e = store.Create("session-1", "Read", "/src/main.ts");
    expect(e.id).toBeGreaterThan(0);
    expect(e.session_id).toBe("session-1");
    expect(e.tool_name).toBe("Read");
    expect(e.file_path).toBe("/src/main.ts");
    expect(e.timestamp).toBeTruthy();
  });

  test("ListBySession returns events ordered by timestamp", () => {
    store.Create("session-1", "Read", "/a.ts");
    store.Create("session-1", "Edit", "/b.ts");
    store.Create("session-2", "Grep", "/c.ts");

    const events = store.ListBySession("session-1");
    expect(events).toHaveLength(2);
    expect(events[0].tool_name).toBe("Read");
    expect(events[1].tool_name).toBe("Edit");
  });

  test("ListBySession returns empty for unknown session", () => {
    expect(store.ListBySession("nonexistent")).toHaveLength(0);
  });
});
