import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { MemoryStore } from "../memory-store.ts";

describe("MemoryStore", () => {
  let db: Database;
  let store: MemoryStore;

  beforeEach(() => {
    db = openMemory();
    store = new MemoryStore(db);
  });

  test("Store creates and returns a knowledge entry", () => {
    const k = store.Store("backend", "pattern", "Use repository pattern", "src/db/", "high");
    expect(k.id).toBeGreaterThan(0);
    expect(k.domain).toBe("backend");
    expect(k.category).toBe("pattern");
    expect(k.finding).toBe("Use repository pattern");
    expect(k.source_files).toBe("src/db/");
    expect(k.confidence).toBe("high");
    expect(k.created_at).toBeTruthy();
  });

  test("Search returns matching entries via FTS5", () => {
    store.Store("backend", "pattern", "Use repository pattern for data access", "src/db/", "high");
    store.Store("frontend", "convention", "Use React hooks for state", "src/ui/", "medium");

    const results = store.Search("repository");
    expect(results).toHaveLength(1);
    expect(results[0].domain).toBe("backend");
  });

  test("Search filters by domain when provided", () => {
    store.Store("backend", "pattern", "Use repository pattern", "src/db/", "high");
    store.Store("frontend", "pattern", "Use repository hook pattern", "src/ui/", "medium");

    const all = store.Search("repository");
    expect(all).toHaveLength(2);

    const filtered = store.Search("repository", "backend");
    expect(filtered).toHaveLength(1);
    expect(filtered[0].domain).toBe("backend");
  });

  test("Search returns empty array when no matches", () => {
    store.Store("backend", "pattern", "Use repository pattern", "src/db/", "high");
    const results = store.Search("nonexistent");
    expect(results).toHaveLength(0);
  });

  test("Domains returns distinct domains sorted alphabetically", () => {
    store.Store("frontend", "pattern", "Hooks", "", "medium");
    store.Store("backend", "convention", "REST", "", "high");
    store.Store("backend", "pattern", "Repository", "", "high");

    const domains = store.Domains();
    expect(domains).toEqual(["backend", "frontend"]);
  });

  test("Domains returns empty when no knowledge exists", () => {
    expect(store.Domains()).toEqual([]);
  });
});
