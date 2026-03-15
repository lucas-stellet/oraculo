import type { Database } from "bun:sqlite";
import type { TaskEnriched } from "../domain/index.ts";
import type { AgentDefinition } from "./agents.ts";
import { TaskStore } from "../db/stores/task-store.ts";
import { AgentStore } from "../db/stores/agent-store.ts";

export interface DispatchResult {
  taskId: number;
  success: boolean;
  summary: string;
  log: string;
}

export interface DispatchOptions {
  task: TaskEnriched;
  agentDef: AgentDefinition;
  prompt: string;
  sessionId: string;
  db: Database;
  onEvent?: (event: unknown) => void;
  /** Injected runner for testability — defaults to a stub. */
  runner?: (opts: RunnerInput) => Promise<string>;
}

export interface RunnerInput {
  model: string;
  prompt: string;
  system: string;
  maxTurns: number;
  tools: string[];
  onEvent?: (event: unknown) => void;
}

/**
 * Default runner stub — in production this would call the Claude Agent SDK.
 * Kept as a separate function so the dispatch logic can be tested without
 * an actual API key or network call.
 */
async function defaultRunner(_opts: RunnerInput): Promise<string> {
  return "Task completed (stub)";
}

/**
 * Dispatches an agent to execute a task.
 * Creates agent/task records, runs the agent, and updates status on completion or failure.
 */
export async function dispatchAgent(
  opts: DispatchOptions,
): Promise<DispatchResult> {
  const { task, agentDef, prompt, sessionId, db, onEvent } = opts;
  const runner = opts.runner ?? defaultRunner;

  const taskStore = new TaskStore(db);
  const agentStore = new AgentStore(db);

  // Transition task to in_progress
  taskStore.start(task.story_id, task.name);

  // Create agent record
  const agent = agentStore.Create(sessionId, agentDef.name, agentDef.type, task.id);

  try {
    const result = await runner({
      model: agentDef.model,
      prompt,
      system: agentDef.systemPrompt,
      maxTurns: agentDef.maxTurns,
      tools: agentDef.tools,
      onEvent,
    });

    const summary = typeof result === "string" ? result : JSON.stringify(result);

    taskStore.complete(task.story_id, task.name, {
      summary,
      logs: "",
      skills_used: [],
      files_modified: [],
    });

    agentStore.Stop(agent.id);

    return { taskId: task.id, success: true, summary, log: "" };
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);

    taskStore.fail(task.story_id, task.name, msg);
    agentStore.Stop(agent.id);

    return { taskId: task.id, success: false, summary: msg, log: "" };
  }
}
