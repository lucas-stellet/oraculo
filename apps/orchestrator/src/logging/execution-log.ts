import {
  appendFileSync,
  mkdirSync,
  existsSync,
  readdirSync,
  rmSync,
  statSync,
} from "node:fs";
import { join } from "node:path";

export interface ExecutionLog {
  write(entry: Record<string, unknown>): void;
  path: string;
}

export function createExecutionLog(
  runId: string,
  taskId: string,
  baseDir?: string,
): ExecutionLog {
  const dir =
    baseDir ?? join(process.env.HOME ?? "~", ".oraculo", "runs", runId);
  mkdirSync(dir, { recursive: true });
  const path = join(dir, `agent-${taskId}.jsonl`);

  return {
    path,
    write(entry: Record<string, unknown>) {
      try {
        const line = JSON.stringify({ ...entry, ts: new Date().toISOString() });
        appendFileSync(path, line + "\n");
      } catch {
        // Never crash the agent due to logging
      }
    },
  };
}

export function pruneOldRuns(baseDir?: string): void {
  const dir =
    baseDir ?? join(process.env.HOME ?? "~", ".oraculo", "runs");
  if (!existsSync(dir)) return;

  const threshold = Date.now() - 3 * 24 * 60 * 60 * 1000; // 3 days
  try {
    for (const entry of readdirSync(dir)) {
      const fullPath = join(dir, entry);
      const stat = statSync(fullPath);
      if (stat.isDirectory() && stat.mtimeMs < threshold) {
        rmSync(fullPath, { recursive: true, force: true });
      }
    }
  } catch {
    // Best effort
  }
}
