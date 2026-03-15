import { Database } from "bun:sqlite";
import { mkdirSync } from "node:fs";
import { join } from "node:path";
import { migrate } from "./migrations.ts";

export function openDatabase(path: string): Database {
  const db = new Database(path);
  db.run("PRAGMA journal_mode = WAL");
  db.run("PRAGMA foreign_keys = ON");
  db.run("PRAGMA busy_timeout = 5000");
  db.run("PRAGMA synchronous = NORMAL");
  migrate(db);
  return db;
}

export function openMemory(): Database {
  return openDatabase(":memory:");
}

export function openProjectDb(): Database {
  const dir = join(process.cwd(), ".oraculo");
  mkdirSync(dir, { recursive: true });
  return openDatabase(join(dir, "oraculo.db"));
}
