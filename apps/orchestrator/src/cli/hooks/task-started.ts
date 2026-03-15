import { Command } from "commander";
import { readConfig } from "../../config/config.ts";

export function taskStartedCommand(): Command {
  return new Command("task-started")
    .description("Broadcast a task-started event")
    .requiredOption("--task-name <name>", "Task name")
    .requiredOption("--story-name <name>", "Story name")
    .requiredOption("--epic-name <name>", "Epic name")
    .action(async (opts) => {
      try {
        await handleTaskStarted(opts);
      } catch (err) {
        console.error(`warning: hook task-started: ${err}`);
      }
    });
}

async function handleTaskStarted(opts: {
  taskName: string;
  storyName: string;
  epicName: string;
}): Promise<void> {
  const config = await readConfig(process.cwd());
  const port = config.port;
  if (port === 0) {
    throw new Error("server port not configured");
  }

  const healthResp = await fetch(`http://localhost:${port}/health`, {
    signal: AbortSignal.timeout(2000),
  });
  if (!healthResp.ok) {
    throw new Error(`server not reachable on port ${port}`);
  }

  const payload = JSON.stringify({
    task_name: opts.taskName,
    story_name: opts.storyName,
    epic_name: opts.epicName,
  });

  const resp = await fetch(`http://localhost:${port}/hooks/task-started`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
    signal: AbortSignal.timeout(2000),
  });

  if (!resp.ok) {
    throw new Error(`task-started returned ${resp.status}`);
  }
}
