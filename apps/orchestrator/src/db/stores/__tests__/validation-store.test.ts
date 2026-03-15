import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { ValidationStore } from "../validation-store.ts";
import { DomainError } from "../../../domain/errors.ts";

describe("ValidationStore", () => {
  let db: Database;
  let store: ValidationStore;

  beforeEach(() => {
    db = openMemory();
    store = new ValidationStore(db);
    db.run("INSERT INTO epics (name, description) VALUES ('test-epic', 'desc')");
    db.run("INSERT INTO stories (epic_id, name, description) VALUES (1, 'test-story', 'desc')");
  });

  test("Create inserts and returns a validation", () => {
    const v = store.Create(1, "approved", "all tests pass");
    expect(v.id).toBeGreaterThan(0);
    expect(v.story_id).toBe(1);
    expect(v.verdict).toBe("approved");
    expect(v.findings).toBe("all tests pass");
    expect(v.created_at).toBeTruthy();
  });

  test("Get retrieves a validation by id", () => {
    const created = store.Create(1, "approved", "ok");
    const fetched = store.Get(created.id);
    expect(fetched.id).toBe(created.id);
    expect(fetched.verdict).toBe("approved");
  });

  test("Get throws not_found for missing id", () => {
    expect(() => store.Get(999)).toThrow(DomainError);
  });

  test("ListByStory returns validations ordered by created_at", () => {
    store.Create(1, "rejected", "test failures");
    store.Create(1, "approved", "all pass");

    const list = store.ListByStory(1);
    expect(list).toHaveLength(2);
    expect(list[0].verdict).toBe("rejected");
    expect(list[1].verdict).toBe("approved");
  });

  test("ListByStory returns empty when no validations exist", () => {
    expect(store.ListByStory(1)).toHaveLength(0);
  });
});
