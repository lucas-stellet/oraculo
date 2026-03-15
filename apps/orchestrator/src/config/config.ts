import { join, basename, dirname } from "node:path";
import { mkdir } from "node:fs/promises";
import { existsSync } from "node:fs";

const CONFIG_REL_PATH = ".oraculo/config.json";

export interface AgentSkills {
  design_agent: string[];
  code_agent: string[];
}

export interface Config {
  port: number;
  name: string;
  preferred_language: string;
  skills: AgentSkills;
}

function defaults(): Config {
  return {
    port: 0,
    name: "",
    preferred_language: "",
    skills: { design_agent: [], code_agent: [] },
  };
}

/**
 * Reads .oraculo/config.json from the given root directory.
 * Returns a Config with defaults when the file does not exist.
 */
export async function readConfig(rootDir: string): Promise<Config> {
  const configPath = join(rootDir, CONFIG_REL_PATH);
  const file = Bun.file(configPath);

  if (!existsSync(configPath)) {
    return defaults();
  }

  const data = await file.json();
  return { ...defaults(), ...data };
}

/**
 * Atomically writes config to .oraculo/config.json via temp file + rename.
 */
export async function writeConfig(rootDir: string, config: Config): Promise<void> {
  const configPath = join(rootDir, CONFIG_REL_PATH);
  const dir = dirname(configPath);

  await mkdir(dir, { recursive: true });

  const json = JSON.stringify(config, null, 2) + "\n";
  const tmpPath = join(dir, `config-${Date.now()}-${Math.random().toString(36).slice(2)}.json`);

  await Bun.write(tmpPath, json);

  const { rename, unlink } = await import("node:fs/promises");
  try {
    await rename(tmpPath, configPath);
  } catch (err) {
    await unlink(tmpPath).catch(() => {});
    throw err;
  }
}

/**
 * Returns the project name from config, or falls back to directory basename.
 */
export async function getProjectName(rootDir: string): Promise<string> {
  const cfg = await readConfig(rootDir);
  if (cfg.name) {
    return cfg.name;
  }
  return basename(rootDir);
}
