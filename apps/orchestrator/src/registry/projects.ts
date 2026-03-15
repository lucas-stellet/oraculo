import { join } from "node:path";
import { homedir } from "node:os";
import { mkdir } from "node:fs/promises";
import { existsSync } from "node:fs";

export interface KnownProject {
  path: string;
  name: string;
}

function projectsPath(): string {
  return join(homedir(), ".oraculo", "projects.json");
}

async function readProjects(): Promise<KnownProject[]> {
  const p = projectsPath();
  if (!existsSync(p)) {
    return [];
  }
  return await Bun.file(p).json();
}

async function writeProjects(projects: KnownProject[]): Promise<void> {
  const p = projectsPath();
  await mkdir(join(homedir(), ".oraculo"), { recursive: true });
  await Bun.write(p, JSON.stringify(projects, null, 2) + "\n");
}

/**
 * Registers a project in ~/.oraculo/projects.json if not already present.
 * Updates the name if the path already exists but the name changed.
 */
export async function registerProject(directory: string, name: string): Promise<void> {
  const projects = await readProjects();
  const idx = projects.findIndex((p) => p.path === directory);

  if (idx >= 0) {
    if (projects[idx].name === name) return; // already up to date
    projects[idx].name = name;
  } else {
    projects.push({ path: directory, name });
  }

  await writeProjects(projects);
}
