import { Command } from "commander";
import { join, basename } from "node:path";
import { mkdir } from "node:fs/promises";
import { existsSync } from "node:fs";
import { readConfig, writeConfig } from "../config/config.ts";
import type { Config } from "../config/config.ts";
import { openProjectDb } from "../db/index.ts";

// ---------------------------------------------------------------------------
// Types for Claude Code settings
// ---------------------------------------------------------------------------

interface HookDef {
  type: string;
  command?: string;
  url?: string;
  timeout?: number;
}

interface HookGroup {
  matcher?: string;
  hooks: HookDef[];
}

interface ClaudeSettings {
  hooks: Record<string, HookGroup[]>;
  mcpServers?: Record<string, { command: string; args: string[] }>;
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

function httpHookGroup(url: string): HookGroup[] {
  return [{ hooks: [{ type: "http", url, timeout: 5 }] }];
}

function buildHooks(baseURL: string): Record<string, HookGroup[]> {
  return {
    SessionStart: [
      { hooks: [{ type: "command", command: "oraculo hook session-start" }] },
    ],
    SubagentStart: httpHookGroup(`${baseURL}/hooks/agent-start`),
    SubagentStop: httpHookGroup(`${baseURL}/hooks/agent-stop`),
    TaskCompleted: httpHookGroup(`${baseURL}/hooks/task-completed`),
    Stop: httpHookGroup(`${baseURL}/hooks/stop`),
    TeammateIdle: httpHookGroup(`${baseURL}/hooks/teammate-idle`),
    SessionEnd: [
      { hooks: [{ type: "command", command: "oraculo hook session-end" }] },
    ],
    PostToolUse: [
      {
        matcher: "Bash|Edit|Write|NotebookEdit",
        hooks: [{ type: "http", url: `${baseURL}/hooks/tool-used`, timeout: 5 }],
      },
    ],
  };
}

/**
 * Scans ports in [start, end] range and returns the first available one.
 */
async function findPort(start: number, end: number): Promise<number> {
  for (let port = start; port <= end; port++) {
    try {
      const server = Bun.serve({ port, fetch: () => new Response("") });
      server.stop(true);
      return port;
    } catch {
      // port in use, try next
    }
  }
  throw new Error(`No available port in range ${start}-${end}`);
}

// ---------------------------------------------------------------------------
// install
// ---------------------------------------------------------------------------

export function installCmd(): Command {
  return new Command("install")
    .description("Install Oraculo skills and hooks into Claude Code")
    .option("--global", "Install globally for all projects")
    .option("--local", "Install locally for current project")
    .option("--lang <lang>", "Preferred language (BCP 47, e.g. pt-BR, en-US)")
    .action(async (opts) => {
      await runInstall(opts);
    });
}

async function runInstall(opts: { global?: boolean; local?: boolean; lang?: string }): Promise<void> {
  // Step 1: Create .oraculo/ directory
  await mkdir(".oraculo", { recursive: true });
  console.log("created .oraculo/");

  // Step 2: Open and close DB to trigger migration
  const db = openProjectDb();
  db.close();
  console.log("created .oraculo/oraculo.db");

  // Step 3: Find an available port
  const port = await findPort(3100, 3199);

  // Step 4: Write config
  const lang = opts.lang || "pt-BR";
  const cfg: Config = {
    port,
    name: "",
    preferred_language: lang,
    skills: { design_agent: [], code_agent: [] },
  };
  await writeConfig(process.cwd(), cfg);
  console.log(`created .oraculo/config.json (port ${port})`);

  // Step 5: Create .claude/ directory
  await mkdir(".claude", { recursive: true });
  console.log("created .claude/");

  // Step 6: Write .claude/settings.json with hooks and MCP config
  const baseURL = `http://localhost:${port}`;
  const settings: ClaudeSettings = {
    hooks: buildHooks(baseURL),
    mcpServers: {
      oraculo: {
        command: "oraculo",
        args: ["start"],
      },
    },
  };
  const settingsData = JSON.stringify(settings, null, 2) + "\n";
  await Bun.write(join(".claude", "settings.json"), settingsData);
  console.log("created .claude/settings.json");

  console.log("Oraculo installed successfully.");
}

// ---------------------------------------------------------------------------
// setup
// ---------------------------------------------------------------------------

export function setupCmd(): Command {
  return new Command("setup")
    .description("Initialize Oraculo infrastructure (for plugin users)")
    .option("--lang <lang>", "Preferred language (BCP 47, e.g. pt-BR, en-US)")
    .action(async (opts) => {
      await runSetup(opts);
    });
}

async function runSetup(opts: { lang?: string }): Promise<void> {
  // Step 1: Create .oraculo/
  await mkdir(".oraculo", { recursive: true });
  console.log("created .oraculo/");

  // Step 2: Open DB to trigger migration
  const db = openProjectDb();
  db.close();
  console.log("created .oraculo/oraculo.db");

  // Step 3: Read existing config or find port
  const existing = await readConfig(process.cwd()).catch(() => null);
  let port = existing?.port || 0;
  if (port === 0) {
    port = await findPort(3100, 3199);
  }

  // Step 4: Write config
  let lang = opts.lang || "";
  if (!lang) {
    lang = existing?.preferred_language || "pt-BR";
  }
  const cfg: Config = {
    port,
    name: existing?.name || "",
    preferred_language: lang,
    skills: existing?.skills || { design_agent: [], code_agent: [] },
  };
  await writeConfig(process.cwd(), cfg);
  console.log(`created .oraculo/config.json (port ${port})`);

  // Step 5: Create .claude/ and write settings.json (hooks only)
  await mkdir(".claude", { recursive: true });
  const baseURL = `http://localhost:${port}`;
  const settings = { hooks: buildHooks(baseURL) };
  const settingsData = JSON.stringify(settings, null, 2) + "\n";
  await Bun.write(join(".claude", "settings.json"), settingsData);
  console.log("created .claude/settings.json (hooks only)");

  // Step 6: Write .mcp.json
  const binaryPath = process.argv[0]!;
  const mcpConfig = {
    mcpServers: {
      oraculo: {
        command: binaryPath,
        args: ["start"],
      },
    },
  };
  const mcpData = JSON.stringify(mcpConfig, null, 2) + "\n";
  await Bun.write(".mcp.json", mcpData);
  console.log(`created .mcp.json (command: ${binaryPath} start)`);

  console.log("Oraculo setup complete. Install the plugin for skills and MCP.");
}

// ---------------------------------------------------------------------------
// uninstall
// ---------------------------------------------------------------------------

export function uninstallCmd(): Command {
  return new Command("uninstall")
    .description("Remove Oraculo from the current project")
    .option("--purge", "Also remove .oraculo/ directory and database")
    .action(async (opts) => {
      await runUninstall(opts);
    });
}

async function runUninstall(opts: { purge?: boolean }): Promise<void> {
  const { readFile, writeFile, rm } = await import("node:fs/promises");

  // Step 1: Clean .claude/settings.json
  const settingsPath = join(".claude", "settings.json");
  if (existsSync(settingsPath)) {
    try {
      const data = await readFile(settingsPath, "utf-8");
      const settings = JSON.parse(data) as Record<string, unknown>;

      // Remove "oraculo" from mcpServers
      if (settings["mcpServers"] && typeof settings["mcpServers"] === "object") {
        const mcp = settings["mcpServers"] as Record<string, unknown>;
        delete mcp["oraculo"];
      }

      // Remove oraculo hook entries
      if (settings["hooks"] && typeof settings["hooks"] === "object") {
        const hooks = settings["hooks"] as Record<string, unknown[]>;
        for (const [event, entries] of Object.entries(hooks)) {
          if (!Array.isArray(entries)) continue;
          hooks[event] = entries.filter((e) => {
            if (typeof e === "string") return !e.startsWith("oraculo");
            return true;
          });
        }
      }

      const out = JSON.stringify(settings, null, 2) + "\n";
      await writeFile(settingsPath, out);
    } catch {
      // ignore parse errors
    }
  }

  // Step 2: Remove oraculo skill directories
  const skillsDir = join(".claude", "skills");
  if (existsSync(skillsDir)) {
    const { readdirSync } = await import("node:fs");
    try {
      const entries = readdirSync(skillsDir);
      for (const entry of entries) {
        if (entry.startsWith("oraculo-") || entry === "oraculo") {
          await rm(join(skillsDir, entry), { recursive: true, force: true });
        }
      }
    } catch {
      // ignore
    }
  }

  // Step 3: Optionally purge .oraculo/
  if (opts.purge) {
    await rm(".oraculo", { recursive: true, force: true });
    console.log("removed .oraculo/");
  }

  console.log("Oraculo uninstalled.");
}
