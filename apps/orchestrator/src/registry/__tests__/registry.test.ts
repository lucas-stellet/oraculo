import { describe, it, expect, beforeEach, afterEach } from "bun:test";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { register, unregister, list, defaultPath } from "../registry.ts";
import type { ServerEntry } from "../registry.ts";

const ts = "2026-01-01T00:00:00.000Z";

describe("registry", () => {
  let tempDir: string;
  let registryPath: string;

  beforeEach(async () => {
    tempDir = await mkdtemp(join(tmpdir(), "oraculo-registry-test-"));
    registryPath = join(tempDir, "servers.json");
  });

  afterEach(async () => {
    await rm(tempDir, { recursive: true, force: true });
  });

  describe("register", () => {
    it("creates registry file and adds entry", async () => {
      const entry: ServerEntry = {
        project: "test-project",
        port: 8080,
        pid: process.pid,
        path: "/tmp/test",
        started_at: ts,
      };

      await register(registryPath, entry);
      const entries = await list(registryPath);
      expect(entries).toHaveLength(1);
      expect(entries[0].project).toBe("test-project");
      expect(entries[0].port).toBe(8080);
    });

    it("updates existing entry with same path", async () => {
      const entry1: ServerEntry = {
        project: "proj",
        port: 8080,
        pid: process.pid,
        path: "/tmp/test",
        started_at: ts,
      };
      await register(registryPath, entry1);

      const entry2: ServerEntry = {
        project: "proj",
        port: 9090,
        pid: process.pid,
        path: "/tmp/test",
        started_at: ts,
      };
      await register(registryPath, entry2);

      const entries = await list(registryPath);
      expect(entries).toHaveLength(1);
      expect(entries[0].port).toBe(9090);
    });

    it("adds multiple entries with different paths", async () => {
      await register(registryPath, {
        project: "proj-a",
        port: 8080,
        pid: process.pid,
        path: "/tmp/a",
        started_at: ts,
      });
      await register(registryPath, {
        project: "proj-b",
        port: 8081,
        pid: process.pid,
        path: "/tmp/b",
        started_at: ts,
      });

      const entries = await list(registryPath);
      expect(entries).toHaveLength(2);
    });

    it("creates parent directories if needed", async () => {
      const nestedPath = join(tempDir, "nested", "dir", "servers.json");
      await register(nestedPath, {
        project: "proj",
        port: 8080,
        pid: process.pid,
        path: "/tmp/test",
        started_at: ts,
      });

      const entries = await list(nestedPath);
      expect(entries).toHaveLength(1);
    });
  });

  describe("unregister", () => {
    it("removes entry by path", async () => {
      await register(registryPath, {
        project: "proj",
        port: 8080,
        pid: process.pid,
        path: "/tmp/test",
        started_at: ts,
      });

      await unregister(registryPath, "/tmp/test");
      const entries = await list(registryPath);
      expect(entries).toHaveLength(0);
    });

    it("does nothing when path not found", async () => {
      await register(registryPath, {
        project: "proj",
        port: 8080,
        pid: process.pid,
        path: "/tmp/test",
        started_at: ts,
      });

      await unregister(registryPath, "/tmp/nonexistent");
      const entries = await list(registryPath);
      expect(entries).toHaveLength(1);
    });

    it("does nothing when registry file does not exist", async () => {
      await expect(unregister(registryPath, "/tmp/test")).resolves.toBeUndefined();
    });
  });

  describe("list", () => {
    it("returns empty array when file does not exist", async () => {
      const entries = await list(registryPath);
      expect(entries).toEqual([]);
    });

    it("filters out stale entries with non-running PIDs", async () => {
      // Write entries directly — one with current PID (alive), one with bogus PID (stale)
      const entries: ServerEntry[] = [
        { project: "alive", port: 8080, pid: process.pid, path: "/tmp/alive", started_at: ts },
        { project: "stale", port: 8081, pid: 999999, path: "/tmp/stale", started_at: ts },
      ];
      await Bun.write(registryPath, JSON.stringify(entries, null, 2) + "\n");

      const result = await list(registryPath);
      expect(result).toHaveLength(1);
      expect(result[0].project).toBe("alive");
    });
  });

  describe("defaultPath", () => {
    it("returns path under home directory", () => {
      const p = defaultPath();
      expect(p).toContain(".oraculo");
      expect(p).toEndWith("servers.json");
    });
  });
});
