import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { SessionStore } from "../session-store.ts";
import { DomainError } from "../../../domain/errors.ts";

describe("SessionStore", () => {
  let db: Database;
  let store: SessionStore;

  beforeEach(() => {
    db = openMemory();
    store = new SessionStore(db);
    db.run("INSERT INTO epics (name, description) VALUES ('test-epic', 'desc')");
  });

  test("Create creates an active session", () => {
    const s = store.Create("epic", 1);
    expect(s.id).toBeTruthy();
    expect(s.type).toBe("epic");
    expect(s.epic_id).toBe(1);
    expect(s.status).toBe("active");
    expect(s.created_at).toBeTruthy();
    expect(s.closed_at).toBeNull();
  });

  test("Create works without epicId", () => {
    const s = store.Create("validate", null);
    expect(s.epic_id).toBeNull();
    expect(s.status).toBe("active");
  });

  test("Get retrieves a session by id", () => {
    const created = store.Create("epic", 1);
    const fetched = store.Get(created.id);
    expect(fetched.id).toBe(created.id);
    expect(fetched.type).toBe("epic");
  });

  test("Get throws not_found for missing id", () => {
    expect(() => store.Get("nonexistent")).toThrow(DomainError);
  });

  test("ActiveByEpic returns active session for epic", () => {
    store.Create("epic", 1);
    const active = store.ActiveByEpic(1);
    expect(active).not.toBeNull();
    expect(active!.status).toBe("active");
  });

  test("ActiveByEpic returns null when no active session", () => {
    expect(store.ActiveByEpic(1)).toBeNull();
  });

  test("Complete marks session as completed", () => {
    const s = store.Create("epic", 1);
    store.Complete(s.id);
    const updated = store.Get(s.id);
    expect(updated.status).toBe("completed");
    expect(updated.closed_at).toBeTruthy();
  });

  test("Complete throws when session not active", () => {
    const s = store.Create("epic", 1);
    store.Complete(s.id);
    expect(() => store.Complete(s.id)).toThrow(DomainError);
  });

  test("Abandon marks session as abandoned", () => {
    const s = store.Create("epic", 1);
    store.Abandon(s.id);
    const updated = store.Get(s.id);
    expect(updated.status).toBe("abandoned");
    expect(updated.closed_at).toBeTruthy();
  });

  test("CompletePhase records a phase and returns it", () => {
    const s = store.Create("validate", 1);
    const phase = store.CompletePhase(s.id, "validate", '{"ok":true}');
    expect(phase.session_id).toBe(s.id);
    expect(phase.name).toBe("validate");
    expect(phase.data).toBe('{"ok":true}');
    expect(phase.completed_at).toBeTruthy();
  });

  test("CompletePhase auto-closes session on last phase", () => {
    const s = store.Create("validate", 1);
    store.CompletePhase(s.id, "validate", "{}");
    const updated = store.Get(s.id);
    expect(updated.status).toBe("completed");
  });

  test("CompletePhase requires predecessor for non-first phase", () => {
    const s = store.Create("execute", 1);
    // "execute" type has phases: ["execute", "validate"]
    // trying to complete "validate" without "execute" first
    expect(() => store.CompletePhase(s.id, "validate", "{}")).toThrow(DomainError);
  });

  test("CompletePhase succeeds when predecessor is complete", () => {
    const s = store.Create("execute", 1);
    store.CompletePhase(s.id, "execute", "{}");
    const phase = store.CompletePhase(s.id, "validate", "{}");
    expect(phase.name).toBe("validate");
  });

  test("CompletePhase throws for invalid phase name", () => {
    const s = store.Create("validate", 1);
    expect(() => store.CompletePhase(s.id, "nonexistent", "{}")).toThrow(DomainError);
  });

  test("CompletePhase throws for duplicate phase", () => {
    const s = store.Create("validate", 1);
    // First call auto-closes, so create a multi-phase session
    const s2 = store.Create("execute", 1);
    store.CompletePhase(s2.id, "execute", "{}");
    expect(() => store.CompletePhase(s2.id, "execute", "{}")).toThrow(DomainError);
  });

  test("CompletePhase throws when session not active", () => {
    const s = store.Create("epic", 1);
    store.Abandon(s.id);
    expect(() => store.CompletePhase(s.id, "discover", "{}")).toThrow(DomainError);
  });

  test("multi-phase session completes all phases in order", () => {
    const s = store.Create("story", 1);
    // story phases: ["define", "plan", "execute", "validate"]
    store.CompletePhase(s.id, "define", "{}");
    expect(store.Get(s.id).status).toBe("active");

    store.CompletePhase(s.id, "plan", "{}");
    expect(store.Get(s.id).status).toBe("active");

    store.CompletePhase(s.id, "execute", "{}");
    expect(store.Get(s.id).status).toBe("active");

    store.CompletePhase(s.id, "validate", "{}");
    expect(store.Get(s.id).status).toBe("completed");
  });
});
