import { describe, it, expect } from "bun:test";
import { openMemory } from "../database.ts";

describe("openDatabase", () => {
  it("opens an in-memory database", () => {
    const db = openMemory();
    expect(db).toBeDefined();
    db.close();
  });

  it("sets WAL journal mode", () => {
    const db = openMemory();
    const result = db.query("PRAGMA journal_mode").get() as { journal_mode: string };
    expect(result.journal_mode).toBe("memory");
    db.close();
  });

  it("enables foreign keys", () => {
    const db = openMemory();
    const result = db.query("PRAGMA foreign_keys").get() as { foreign_keys: number };
    expect(result.foreign_keys).toBe(1);
    db.close();
  });

  it("sets busy timeout", () => {
    const db = openMemory();
    const result = db.query("PRAGMA busy_timeout").get() as { timeout: number };
    expect(result.timeout).toBe(5000);
    db.close();
  });
});
