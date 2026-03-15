import { Command } from "commander";
import { VERSION } from "../version.ts";
import { startCmd, startHttpCmd, startMcpCmd } from "./lifecycle.ts";
import { killCmd } from "./lifecycle.ts";
import { restartCmd } from "./lifecycle.ts";
import { statusCmd } from "./lifecycle.ts";
import { logsCmd } from "./lifecycle.ts";
import { versionCmd } from "./lifecycle.ts";
import { installCmd } from "./install.ts";
import { setupCmd } from "./install.ts";
import { uninstallCmd } from "./install.ts";

export function createProgram(): Command {
  const program = new Command();

  program
    .name("oraculo")
    .description("Socratic guide for quality product development")
    .version(VERSION);

  // Lifecycle commands
  const start = startCmd();
  start.addCommand(startHttpCmd());
  start.addCommand(startMcpCmd());
  program.addCommand(start);
  program.addCommand(killCmd());
  program.addCommand(restartCmd());
  program.addCommand(statusCmd());
  program.addCommand(logsCmd());
  program.addCommand(versionCmd());

  // Install / setup / uninstall
  program.addCommand(installCmd());
  program.addCommand(setupCmd());
  program.addCommand(uninstallCmd());

  return program;
}
