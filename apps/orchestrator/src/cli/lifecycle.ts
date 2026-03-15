import { Command } from "commander";
import { basename } from "node:path";
import { readConfig } from "../config/config.ts";
import { VERSION } from "../version.ts";
import { openProjectDb } from "../db/index.ts";
import { createApp } from "../server/http.ts";
import * as registry from "../registry/registry.ts";
import { formatTable } from "../utils/output.ts";

const DEFAULT_PORT = 3100;

/**
 * Resolves the configured port, falling back to DEFAULT_PORT.
 */
async function resolvedPort(): Promise<number> {
  const cfg = await readConfig(process.cwd());
  return cfg.port || DEFAULT_PORT;
}

// ---------------------------------------------------------------------------
// start
// ---------------------------------------------------------------------------

export function startCmd(): Command {
  return new Command("start")
    .description("Start Oraculo services");
}

export function startHttpCmd(): Command {
  return new Command("http")
    .description("Start HTTP + WebSocket server as daemon")
    .action(async () => {
      await runStartHttp();
    });
}

export function startMcpCmd(): Command {
  return new Command("mcp")
    .description("Start MCP server on stdio (managed by Claude Code)")
    .action(async () => {
      console.log("MCP server not yet implemented in TypeScript orchestrator.");
      process.exit(1);
    });
}

/**
 * Starts the HTTP server in the foreground.
 * Registers in the global server registry and unregisters on shutdown.
 */
async function runStartHttp(): Promise<void> {
  const db = openProjectDb();
  const cfg = await readConfig(process.cwd());
  const port = cfg.port || DEFAULT_PORT;

  const { app, hub, broadcaster } = createApp(db, { version: VERSION });

  const wd = process.cwd();
  const projectName = cfg.name || basename(wd);
  const regPath = registry.defaultPath();

  await registry.register(regPath, {
    project: projectName,
    port,
    pid: process.pid,
    path: wd,
    started_at: new Date().toISOString(),
  });

  const cleanup = async () => {
    await registry.unregister(regPath, wd);
    db.close();
  };

  const server = Bun.serve({
    port,
    fetch: app.fetch,
  });

  console.log(`Oraculo HTTP server started on port ${server.port}`);

  process.on("SIGTERM", async () => {
    console.log("Received SIGTERM, shutting down...");
    server.stop();
    await cleanup();
    process.exit(0);
  });

  process.on("SIGINT", async () => {
    console.log("Received SIGINT, shutting down...");
    server.stop();
    await cleanup();
    process.exit(0);
  });
}

// ---------------------------------------------------------------------------
// kill
// ---------------------------------------------------------------------------

export function killCmd(): Command {
  return new Command("kill")
    .description("Stop the running Oraculo HTTP server")
    .action(async () => {
      const port = await resolvedPort();
      await killServer(port);
    });
}

/**
 * Finds processes listening on the given port via lsof,
 * sends SIGTERM, waits for port release, escalates to SIGKILL.
 */
export async function killServer(port: number): Promise<void> {
  const pids = await pidsOnPort(port);

  if (pids.length === 0) {
    console.log(`No process found on port ${port}.`);
    return;
  }

  for (const pid of pids) {
    if (!isProcessRunning(pid)) continue;
    console.log(`Stopping Oraculo (pid ${pid}, port ${port})...`);
    try {
      process.kill(pid, "SIGTERM");
    } catch {
      // process may have already exited
    }
  }

  // Wait up to 5s for port release
  for (let i = 0; i < 50; i++) {
    await sleep(100);
    if (!(await portInUse(port))) {
      console.log("Oraculo stopped.");
      return;
    }
  }

  // Force-kill survivors
  for (const pid of pids) {
    try {
      process.kill(pid, "SIGKILL");
    } catch {
      // ignore
    }
  }
  await sleep(300);
  console.log("Oraculo force-stopped.");
}

/**
 * Returns PIDs listening on the given TCP port using lsof.
 */
async function pidsOnPort(port: number): Promise<number[]> {
  try {
    const proc = Bun.spawn(["lsof", "-ti", `tcp:${port}`], {
      stdout: "pipe",
      stderr: "ignore",
    });
    const text = await new Response(proc.stdout).text();
    await proc.exited;

    return text
      .trim()
      .split("\n")
      .map((l) => l.trim())
      .filter((l) => l.length > 0)
      .map(Number)
      .filter((n) => !Number.isNaN(n));
  } catch {
    return [];
  }
}

/**
 * Checks if a port is currently in use by attempting to listen on it.
 */
async function portInUse(port: number): Promise<boolean> {
  try {
    const server = Bun.serve({ port, fetch: () => new Response("") });
    server.stop(true);
    return false;
  } catch {
    return true;
  }
}

function isProcessRunning(pid: number): boolean {
  try {
    process.kill(pid, 0);
    return true;
  } catch {
    return false;
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ---------------------------------------------------------------------------
// restart
// ---------------------------------------------------------------------------

export function restartCmd(): Command {
  return new Command("restart")
    .description("Restart the Oraculo HTTP server")
    .action(async () => {
      const port = await resolvedPort();

      // Kill existing instance
      await killServer(port);

      // Wait for port to be free
      for (let i = 0; i < 50; i++) {
        if (!(await portInUse(port))) break;
        await sleep(100);
      }
      if (await portInUse(port)) {
        console.error(`Port ${port} still in use after shutdown.`);
        process.exit(1);
      }

      // Spawn detached child: bun <this script> start http
      const scriptPath = process.argv[1]!;
      const child = Bun.spawn(["bun", "run", scriptPath, "start", "http"], {
        stdio: ["ignore", "ignore", "ignore"],
        env: process.env as Record<string, string>,
      });
      child.unref();

      console.log(`Oraculo started (port ${port}).`);
    });
}

// ---------------------------------------------------------------------------
// status
// ---------------------------------------------------------------------------

export function statusCmd(): Command {
  return new Command("status")
    .description("Show project dashboard")
    .action(async () => {
      const { existsSync } = await import("node:fs");

      if (!existsSync(".oraculo")) {
        console.log("No Oraculo project found. Run `oraculo install` to get started.");
        return;
      }

      const db = openProjectDb();
      try {
        console.log("=== Oraculo Status ===");

        // Epics
        const { EpicStore } = await import("../db/stores/epic-store.ts");
        const epics = new EpicStore(db).list();

        console.log();
        console.log("Epics");
        if (epics.length === 0) {
          console.log("(none)");
        } else {
          const headers = ["Name", "Status"];
          const rows = epics.map((e) => [e.name, e.approval_status]);
          console.log(formatTable(headers, rows));
        }

        // Task summary
        const stmt = db.query<{ status: string; count: number }, []>(
          "SELECT status, COUNT(*) as count FROM tasks GROUP BY status ORDER BY status",
        );
        const taskCounts = stmt.all();
        const counts: Record<string, number> = {
          pending: 0,
          in_progress: 0,
          completed: 0,
          failed: 0,
        };
        for (const row of taskCounts) {
          counts[row.status] = row.count;
        }

        console.log();
        console.log("Tasks");
        console.log(
          formatTable(
            ["Status", "Count"],
            [
              ["pending", String(counts["pending"])],
              ["in_progress", String(counts["in_progress"])],
              ["completed", String(counts["completed"])],
              ["failed", String(counts["failed"])],
            ],
          ),
        );

        // Pending approvals
        const { ApprovalStore } = await import("../db/stores/approval-store.ts");
        const pendingApprovals = new ApprovalStore(db).list(true);

        console.log();
        console.log(`Pending Approvals: ${pendingApprovals.length}`);
      } finally {
        db.close();
      }
    });
}

// ---------------------------------------------------------------------------
// logs
// ---------------------------------------------------------------------------

export function logsCmd(): Command {
  return new Command("logs")
    .description("Stream logs from the running Oraculo instance")
    .action(async () => {
      const port = await resolvedPort();
      const url = `http://localhost:${port}/logs`;

      let response: Response;
      try {
        response = await fetch(url, {
          headers: { Accept: "text/event-stream" },
        });
      } catch (err) {
        console.error(`Failed to connect to Oraculo instance at :${port}: ${err}`);
        process.exit(1);
        return; // unreachable, for type narrowing
      }

      if (!response.ok) {
        console.error(`Unexpected status ${response.status} from /logs`);
        process.exit(1);
      }

      if (!response.body) {
        console.error("No response body from /logs");
        process.exit(1);
        return;
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const text = decoder.decode(value, { stream: true });
        for (const line of text.split("\n")) {
          const out = parseSseLine(line);
          if (out) console.log(out);
        }
      }
    });
}

/**
 * Extracts and formats a log entry from an SSE "data:" line.
 * Returns empty string for non-data lines.
 */
export function parseSseLine(line: string): string {
  if (!line.startsWith("data: ")) return "";
  const raw = line.slice(6);
  try {
    const e = JSON.parse(raw) as { time?: string; level?: string; msg?: string };
    if (e.time && e.level && e.msg) {
      return `${e.time} ${e.level} ${e.msg}`;
    }
    return raw;
  } catch {
    return raw;
  }
}

// ---------------------------------------------------------------------------
// version
// ---------------------------------------------------------------------------

export function versionCmd(): Command {
  return new Command("version")
    .description("Print the CLI version")
    .action(() => {
      console.log(VERSION);
    });
}
