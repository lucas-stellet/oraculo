import { join, dirname } from "node:path";
import { homedir } from "node:os";
import { mkdir } from "node:fs/promises";
import { existsSync } from "node:fs";

export interface ServerEntry {
  project: string;
  port: number;
  pid: number;
  directory: string;
}

/**
 * Checks if a process with the given PID is still running.
 */
function isProcessRunning(pid: number): boolean {
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

/**
 * Reads all entries from the registry file, filtering out stale ones (dead PIDs).
 */
export async function list(registryPath: string): Promise<ServerEntry[]> {
  if (!existsSync(registryPath)) {
    return [];
  }

  const file = Bun.file(registryPath);
  const entries: ServerEntry[] = await file.json();

  return entries.filter((e) => isProcessRunning(e.pid));
}

/**
 * Registers a server entry. Updates existing entry if directory matches.
 * Creates parent directories if needed.
 */
export async function register(registryPath: string, entry: ServerEntry): Promise<void> {
  await mkdir(dirname(registryPath), { recursive: true });

  const entries = await readRaw(registryPath);

  const idx = entries.findIndex((e) => e.directory === entry.directory);
  if (idx >= 0) {
    entries[idx] = entry;
  } else {
    entries.push(entry);
  }

  await writeEntries(registryPath, entries);
}

/**
 * Removes the entry matching the given directory.
 */
export async function unregister(registryPath: string, directory: string): Promise<void> {
  const entries = await readRaw(registryPath);
  const filtered = entries.filter((e) => e.directory !== directory);
  await writeEntries(registryPath, filtered);
}

/**
 * Returns the default registry path: ~/.oraculo/servers.json
 */
export function defaultPath(): string {
  return join(homedir(), ".oraculo", "servers.json");
}

/**
 * Reads entries without filtering stale PIDs (for internal mutation).
 */
async function readRaw(registryPath: string): Promise<ServerEntry[]> {
  if (!existsSync(registryPath)) {
    return [];
  }
  const file = Bun.file(registryPath);
  return await file.json();
}

/**
 * Writes entries to the registry file.
 */
async function writeEntries(registryPath: string, entries: ServerEntry[]): Promise<void> {
  await mkdir(dirname(registryPath), { recursive: true });
  const json = JSON.stringify(entries, null, 2) + "\n";
  await Bun.write(registryPath, json);
}
