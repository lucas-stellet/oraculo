import { Command } from "commander";
import { withDb } from "../context.ts";
import { epicCommands } from "./epic.ts";
import { storyCommands } from "./story.ts";
import { taskCommands } from "./task.ts";
import { memoryCommands } from "./memory.ts";
import { approvalCommands } from "./approval.ts";
import { sessionCommands } from "./session.ts";
import { phaseCommands } from "./phase.ts";
import { designCommands } from "./design.ts";
import { reviewCommands } from "./review.ts";
import { validationCommands } from "./validation.ts";

export function toolCommands(): Command {
  const cmd = new Command("tools").description(
    "Agent-facing commands (JSON output)",
  );

  withDb(cmd);

  cmd.addCommand(epicCommands());
  cmd.addCommand(storyCommands());
  cmd.addCommand(taskCommands());
  cmd.addCommand(memoryCommands());
  cmd.addCommand(approvalCommands());
  cmd.addCommand(sessionCommands());
  cmd.addCommand(phaseCommands());
  cmd.addCommand(designCommands());
  cmd.addCommand(reviewCommands());
  cmd.addCommand(validationCommands());

  return cmd;
}
