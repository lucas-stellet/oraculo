import { describe, it, expect } from "bun:test";
import { openMemory } from "../database.ts";
import { migrate } from "../migrations.ts";

describe("migrations", () => {
  it("runs all 9 migrations", () => {
    const db = openMemory();
    const version = db.query("PRAGMA user_version").get() as { user_version: number };
    expect(version.user_version).toBe(9);
    db.close();
  });

  it("creates epics table", () => {
    const db = openMemory();
    const result = db.query("SELECT name FROM sqlite_master WHERE type='table' AND name='epics'").get();
    expect(result).toBeTruthy();
    db.close();
  });

  it("creates tasks table with all columns", () => {
    const db = openMemory();
    const cols = db.query("PRAGMA table_info(tasks)").all() as { name: string }[];
    const colNames = cols.map((c) => c.name);
    expect(colNames).toContain("id");
    expect(colNames).toContain("story_id");
    expect(colNames).toContain("name");
    expect(colNames).toContain("description");
    expect(colNames).toContain("status");
    expect(colNames).toContain("type");
    db.close();
  });

  it("creates knowledge FTS5 virtual table", () => {
    const db = openMemory();
    const result = db.query("SELECT name FROM sqlite_master WHERE type='table' AND name='knowledge_fts'").get();
    expect(result).toBeTruthy();
    db.close();
  });

  it("creates approval_comments table (V9)", () => {
    const db = openMemory();
    const result = db.query("SELECT name FROM sqlite_master WHERE type='table' AND name='approval_comments'").get();
    expect(result).toBeTruthy();
    db.close();
  });

  it("creates agents table with task_id column (V8)", () => {
    const db = openMemory();
    const cols = db.query("PRAGMA table_info(agents)").all() as { name: string }[];
    expect(cols.map((c) => c.name)).toContain("task_id");
    db.close();
  });

  it("creates validations table with findings column (V7)", () => {
    const db = openMemory();
    const cols = db.query("PRAGMA table_info(validations)").all() as { name: string }[];
    expect(cols.map((c) => c.name)).toContain("findings");
    db.close();
  });

  it("is idempotent", () => {
    const db = openMemory();
    const version1 = db.query("PRAGMA user_version").get() as { user_version: number };
    migrate(db);
    const version2 = db.query("PRAGMA user_version").get() as { user_version: number };
    expect(version2.user_version).toBe(version1.user_version);
    db.close();
  });
});
