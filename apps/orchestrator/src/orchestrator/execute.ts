import type { Database } from "bun:sqlite";
import { TaskStore } from "../db/stores/task-store.ts";
import { nextWave } from "./dag.ts";
import { agentForTaskType } from "./agents.ts";
import { buildTaskPrompt } from "./prompts.ts";
import { dispatchAgent, type RunnerInput } from "./dispatch.ts";

export interface ExecuteOptions {
  storyId: number;
  sessionId: string;
  epicName: string;
  storyName: string;
  db: Database;
  maxFailures?: number;
  onEvent?: (event: unknown) => void;
  /** Injected runner for testability. */
  runner?: (opts: RunnerInput) => Promise<string>;
}

export interface ExecuteResult {
  success: boolean;
  results: Array<{ taskId: number; success: boolean }>;
}

/**
 * Executes all tasks for a story in dependency order.
 *
 * Processing proceeds in waves: each wave contains tasks whose dependencies
 * are all completed. Tasks within a wave run concurrently.
 *
 * A circuit breaker aborts execution after `maxFailures` consecutive failures.
 * Deadlocks (pending tasks with unsatisfiable dependencies) are detected and
 * reported as failures.
 */
export async function executeStory(
  opts: ExecuteOptions,
): Promise<ExecuteResult> {
  const {
    storyId,
    sessionId,
    epicName,
    storyName,
    db,
    maxFailures = 3,
    onEvent,
    runner,
  } = opts;

  const taskStore = new TaskStore(db);
  const results: Array<{ taskId: number; success: boolean }> = [];
  let consecutiveFailures = 0;

  while (true) {
    const tasks = taskStore.listEnriched(storyId);
    const wave = nextWave(tasks);

    if (wave.tasks.length === 0) {
      // Check if all done or deadlocked
      const pending = tasks.filter((t) => t.status === "pending");
      if (pending.length === 0) break; // All tasks completed or failed
      // Deadlocked — pending tasks exist but none are ready
      return { success: false, results };
    }

    // Dispatch all tasks in wave concurrently
    const promises = wave.tasks.map(async (task) => {
      // Task type is not currently on the TaskEnriched interface;
      // default to "code" until the domain entity is extended.
      const agentDef = agentForTaskType("code");
      const prompt = buildTaskPrompt(task, {
        epicName,
        storyName,
        knowledge: [],
      });
      return dispatchAgent({
        task,
        agentDef,
        prompt,
        sessionId,
        db,
        onEvent,
        runner,
      });
    });

    const waveResults = await Promise.all(promises);

    for (const result of waveResults) {
      results.push({ taskId: result.taskId, success: result.success });
      if (!result.success) {
        consecutiveFailures++;
        if (consecutiveFailures >= maxFailures) {
          return { success: false, results };
        }
      } else {
        consecutiveFailures = 0;
      }
    }
  }

  return { success: true, results };
}
