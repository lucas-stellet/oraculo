import { Command } from "commander";
import { readConfig } from "../../config/config.ts";
import { openProjectDb } from "../../db/index.ts";

export function sessionStartCommand(): Command {
  return new Command("session-start")
    .description("Handle session start hook")
    .action(async () => {
      try {
        await handleSessionStart();
      } catch (err) {
        // Never block Claude Code session
        console.error(`warning: hook session-start: ${err}`);
      }
    });
}

async function handleSessionStart(): Promise<void> {
  const input = JSON.parse(await Bun.stdin.text());
  const sessionId = input.session_id;
  if (!sessionId) {
    throw new Error("missing session_id in stdin");
  }

  // Register session in DB
  const db = openProjectDb();
  try {
    const wd = process.cwd();
    const branch = await gitBranch();
    const metadata = JSON.stringify({
      working_dir: wd,
      git_branch: branch,
      source: input.source ?? "",
      updated_at: new Date().toISOString(),
    });

    db.run(
      "INSERT OR IGNORE INTO claude_sessions (id, metadata) VALUES (?, ?)",
      [sessionId, metadata],
    );
    db.run("UPDATE claude_sessions SET metadata = ? WHERE id = ?", [
      metadata,
      sessionId,
    ]);
  } finally {
    db.close();
  }

  // Health check and auto-start
  const config = await readConfig(process.cwd());
  const port = config.port;
  if (port === 0) return;

  const healthUrl = `http://localhost:${port}/health`;
  let online = await isHealthy(healthUrl);

  if (!online) {
    console.error(
      `warning: Oraculo HTTP server offline — auto-starting on port ${port}`,
    );
    try {
      spawnDaemon();
    } catch (err) {
      const msg = `warning: failed to auto-start Oraculo server: ${err}`;
      console.error(msg);
      console.log(msg);
      return;
    }
    online = await pollHealth(healthUrl, 500, 10_000);
    if (!online) {
      const msg =
        "warning: Oraculo server started but not responding. Telemetry unavailable.";
      console.error(msg);
      console.log(msg);
      return;
    }
  }

  // POST session-start
  const config2 = await readConfig(process.cwd());
  const wd = process.cwd();
  const branch = await gitBranch();
  const payload = JSON.stringify({
    working_dir: wd,
    git_branch: branch,
    source: input.source ?? "",
    updated_at: new Date().toISOString(),
  });

  await fetch(`http://localhost:${port}/hooks/session-start`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
    signal: AbortSignal.timeout(2000),
  }).catch(() => {});
}

async function isHealthy(url: string): Promise<boolean> {
  try {
    const resp = await fetch(url, { signal: AbortSignal.timeout(2000) });
    return resp.ok;
  } catch {
    return false;
  }
}

async function pollHealth(
  url: string,
  intervalMs: number,
  timeoutMs: number,
): Promise<boolean> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    await Bun.sleep(intervalMs);
    if (await isHealthy(url)) return true;
  }
  return false;
}

/** Exported so tests can replace it. */
export let spawnDaemon = (): void => {
  const entrypoint = import.meta.dir + "/../../index.ts";
  Bun.spawn(["bun", "run", entrypoint, "start", "http"], {
    stdio: ["ignore", "ignore", "ignore"],
  });
};

async function gitBranch(): Promise<string> {
  try {
    const proc = Bun.spawn(["git", "rev-parse", "--abbrev-ref", "HEAD"], {
      stdout: "pipe",
      stderr: "ignore",
    });
    const text = await new Response(proc.stdout).text();
    return text.trim();
  } catch {
    return "";
  }
}
