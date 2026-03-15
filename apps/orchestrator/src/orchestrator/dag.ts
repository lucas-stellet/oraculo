import type { TaskEnriched } from "../domain/index.ts";

export interface Wave {
  tasks: TaskEnriched[];
}

/**
 * Returns the next wave of tasks that are ready to execute.
 * A task is ready when it is pending and all its dependencies are completed.
 */
export function nextWave(tasks: TaskEnriched[]): Wave {
  const ready = tasks.filter(
    (t) =>
      t.status === "pending" &&
      t.depends_on.every((depId) => {
        const dep = tasks.find((d) => d.id === depId);
        return dep?.status === "completed";
      }),
  );
  return { tasks: ready };
}
