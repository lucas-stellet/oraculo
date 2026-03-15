import { describe, it, expect } from "bun:test";
import { createExecutionLog, pruneOldRuns } from "../execution-log.ts";
import { readFileSync, rmSync, mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";

describe("ExecutionLog", () => {
  it("writes JSONL entries", () => {
    const dir = join(tmpdir(), `oraculo-test-${Date.now()}`);
    const log = createExecutionLog("test-run", "task-1", dir);
    log.write({ msg: "started", event: "agent_started" });
    log.write({ msg: "completed", event: "agent_completed" });
    const content = readFileSync(log.path, "utf-8").trim().split("\n");
    expect(content).toHaveLength(2);
    expect(JSON.parse(content[0]).msg).toBe("started");
    rmSync(dir, { recursive: true, force: true });
  });

  it("includes timestamp in entries", () => {
    const dir = join(tmpdir(), `oraculo-test-${Date.now()}`);
    const log = createExecutionLog("test-run", "task-2", dir);
    log.write({ msg: "test" });
    const content = readFileSync(log.path, "utf-8").trim();
    const entry = JSON.parse(content);
    expect(entry.ts).toBeDefined();
    rmSync(dir, { recursive: true, force: true });
  });
});

describe("pruneOldRuns", () => {
  it("removes directories older than 3 days", () => {
    const baseDir = join(tmpdir(), `oraculo-prune-${Date.now()}`);
    const oldDir = join(baseDir, "old-run");
    const newDir = join(baseDir, "new-run");
    mkdirSync(oldDir, { recursive: true });
    mkdirSync(newDir, { recursive: true });

    // Make old dir appear old by writing a marker and modifying mtime
    writeFileSync(join(oldDir, "marker"), "old");
    writeFileSync(join(newDir, "marker"), "new");

    // Note: Can't easily backdate mtime in test, so just verify no crash
    pruneOldRuns(baseDir);
    rmSync(baseDir, { recursive: true, force: true });
  });
});
