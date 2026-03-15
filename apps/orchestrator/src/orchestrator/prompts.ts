import type { TaskEnriched } from "../domain/index.ts";

export interface TaskContext {
  epicName: string;
  storyName: string;
  knowledge: string[];
}

/**
 * Builds the user prompt sent to an agent for a specific task.
 * Includes task details, hierarchy context, and accumulated knowledge.
 */
export function buildTaskPrompt(
  task: TaskEnriched,
  context: TaskContext,
): string {
  const sections = [
    `# Task: ${task.name}`,
    `Epic: ${context.epicName}`,
    `Story: ${context.storyName}`,
    ``,
    `## Description`,
    task.description,
  ];

  if (context.knowledge.length > 0) {
    sections.push(
      ``,
      `## Accumulated Knowledge`,
      ...context.knowledge,
    );
  }

  return sections.join("\n");
}
