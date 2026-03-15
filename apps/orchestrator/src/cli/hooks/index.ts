import { Command } from "commander";
import { sessionStartCommand } from "./session-start.ts";
import { sessionEndCommand } from "./session-end.ts";
import { agentStartCommand } from "./agent-start.ts";
import { taskStartedCommand } from "./task-started.ts";

export function hookCommands(): Command {
  const cmd = new Command("hook").description(
    "Commands triggered by Claude Code hooks",
  );

  cmd.addCommand(sessionStartCommand());
  cmd.addCommand(sessionEndCommand());
  cmd.addCommand(agentStartCommand());
  cmd.addCommand(taskStartedCommand());

  return cmd;
}
