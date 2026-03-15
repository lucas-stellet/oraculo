import { Command } from "commander";
import { readConfig } from "../../config/config.ts";
import { openProjectDb } from "../../db/index.ts";

export function sessionEndCommand(): Command {
  return new Command("session-end")
    .description("Handle session end hook")
    .action(async () => {
      try {
        await handleSessionEnd();
      } catch (err) {
        console.error(`warning: hook session-end: ${err}`);
      }
    });
}

async function handleSessionEnd(): Promise<void> {
  const input = JSON.parse(await Bun.stdin.text());
  const sessionId = input.session_id;
  if (!sessionId) {
    throw new Error("missing session_id in stdin");
  }

  const db = openProjectDb();
  try {
    db.run(
      "UPDATE claude_sessions SET ended_at = datetime('now') WHERE id = ?",
      [sessionId],
    );

    const payload = JSON.stringify({ reason: input.reason ?? "" });
    db.run(
      "INSERT INTO session_events (session_id, event_type, payload) VALUES (?, 'session_end', ?)",
      [sessionId, payload],
    );
  } finally {
    db.close();
  }

  // Notify HTTP server if online
  const config = await readConfig(process.cwd());
  const port = config.port;
  if (port === 0) return;

  try {
    const healthResp = await fetch(`http://localhost:${port}/health`, {
      signal: AbortSignal.timeout(2000),
    });
    if (!healthResp.ok) return;
  } catch {
    return;
  }

  await fetch(`http://localhost:${port}/hooks/session-end`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      session_id: sessionId,
      reason: input.reason ?? "",
    }),
    signal: AbortSignal.timeout(2000),
  }).catch(() => {});
}
