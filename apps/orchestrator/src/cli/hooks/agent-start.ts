import { Command } from "commander";
import { readConfig } from "../../config/config.ts";

export function agentStartCommand(): Command {
  return new Command("agent-start")
    .description("Register an agent start with task association")
    .requiredOption("--agent-name <name>", "Agent name")
    .option("--agent-type <type>", "Agent type (code, research, qa)", "code")
    .requiredOption("--task-name <name>", "Task name")
    .requiredOption("--story-name <name>", "Story name")
    .requiredOption("--epic-name <name>", "Epic name")
    .option("--session-id <id>", "Claude Code session ID", "")
    .action(async (opts) => {
      try {
        await handleAgentStart(opts);
      } catch (err) {
        console.error(`warning: hook agent-start: ${err}`);
      }
    });
}

async function handleAgentStart(opts: {
  agentName: string;
  agentType: string;
  taskName: string;
  storyName: string;
  epicName: string;
  sessionId: string;
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
    session_id: opts.sessionId,
    agent_name: opts.agentName,
    agent_type: opts.agentType,
    task_name: opts.taskName,
    story_name: opts.storyName,
    epic_name: opts.epicName,
  });

  const resp = await fetch(`http://localhost:${port}/hooks/agent-start`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: payload,
    signal: AbortSignal.timeout(2000),
  });

  if (!resp.ok) {
    throw new Error(`agent-start returned ${resp.status}`);
  }
}
