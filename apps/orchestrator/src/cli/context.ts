import type { Command } from "commander";
import type { Database } from "bun:sqlite";
import { openProjectDb } from "../db/index.ts";

/**
 * Attaches a preAction hook that opens the project database and stores it
 * on the command so subcommands can retrieve it via `dbFromCmd`.
 */
export function withDb(cmd: Command): Command {
  cmd.hook("preAction", (thisCmd) => {
    thisCmd.setOptionValue("_db", openProjectDb());
  });
  return cmd;
}

/**
 * Retrieves the Database instance previously attached by `withDb`.
 * Walks up the parent chain to find the option.
 */
export function dbFromCmd(cmd: Command): Database {
  let current: Command | null = cmd;
  while (current) {
    const db = current.getOptionValue("_db") as Database | undefined;
    if (db) return db;
    current = current.parent;
  }
  throw new Error("database not initialized — is withDb() wired?");
}
