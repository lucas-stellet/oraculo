import { Command } from "commander";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { StoryStore } from "../../db/stores/story-store.ts";
import { TaskStore } from "../../db/stores/task-store.ts";
import type { CompleteInput } from "../../db/stores/task-store.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";
import type { Database } from "bun:sqlite";

export function taskCommands(): Command {
  const cmd = new Command("task").description("Manage tasks within a story");

  cmd.addCommand(taskInitCmd());
  cmd.addCommand(taskStartCmd());
  cmd.addCommand(taskCompleteCmd());
  cmd.addCommand(taskFailCmd());
  cmd.addCommand(taskGetCmd());
  cmd.addCommand(taskListCmd());
  cmd.addCommand(taskDeleteCmd());

  return cmd;
}

function resolveStory(cmd: Command): {
  db: Database;
  epicName: string;
  storyName: string;
  storyId: number;
} {
  const db = dbFromCmd(cmd);
  const epicName = cmd.opts().epic;
  const storyName = cmd.opts().story;

  const epicStore = new EpicStore(db);
  const epic = epicStore.getByName(epicName);

  const storyStore = new StoryStore(db);
  const story = storyStore.getByName(epic.id, storyName);

  return { db, epicName, storyName, storyId: story.id };
}

function taskInitCmd(): Command {
  return new Command("init")
    .description("Create a new task within a story")
    .argument("<name>", "Task name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .option("--description <desc>", "Task description", "")
    .option("--depends-on <names...>", "Task dependencies")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicName, storyName, storyId } = resolveStory(cmd);
        const store = new TaskStore(db);
        const dependsOn = opts.dependsOn ?? [];
        const [task, created] = store.create(
          storyId,
          name,
          opts.description,
          dependsOn,
        );
        writeJSON({
          name,
          epic: epicName,
          story: storyName,
          created,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function taskStartCmd(): Command {
  return new Command("start")
    .description("Transition a task from pending to in_progress")
    .argument("<name>", "Task name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, storyId } = resolveStory(cmd);
        const store = new TaskStore(db);
        const task = store.start(storyId, name);
        writeJSON(task);
      } catch (err) {
        writeError(err);
      }
    });
}

function taskCompleteCmd(): Command {
  return new Command("complete")
    .description(
      "Transition a task from in_progress to completed (reads JSON from stdin)",
    )
    .argument("<name>", "Task name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const { db, storyId } = resolveStory(cmd);
        const raw = await Bun.stdin.text();
        const input: CompleteInput = JSON.parse(raw);
        const store = new TaskStore(db);
        const task = store.complete(storyId, name, input);
        writeJSON(task);
      } catch (err) {
        writeError(err);
      }
    });
}

function taskFailCmd(): Command {
  return new Command("fail")
    .description("Transition a task from in_progress to failed")
    .argument("<name>", "Task name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .requiredOption("--reason <reason>", "Failure reason")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, storyId } = resolveStory(cmd);
        const store = new TaskStore(db);
        const task = store.fail(storyId, name, opts.reason);
        writeJSON(task);
      } catch (err) {
        writeError(err);
      }
    });
}

function taskGetCmd(): Command {
  return new Command("get")
    .description("Get a task by name")
    .argument("<name>", "Task name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, storyId } = resolveStory(cmd);
        const store = new TaskStore(db);
        const task = store.getByName(storyId, name);
        writeJSON(task);
      } catch (err) {
        writeError(err);
      }
    });
}

function taskListCmd(): Command {
  return new Command("list")
    .description("List all tasks for a story")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .action((opts: any, cmd: Command) => {
      try {
        const { db, storyId } = resolveStory(cmd);
        const store = new TaskStore(db);
        writeJSON(store.listEnriched(storyId));
      } catch (err) {
        writeError(err);
      }
    });
}

function taskDeleteCmd(): Command {
  return new Command("delete")
    .description("Delete a task")
    .argument("<name>", "Task name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption("--story <name>", "Parent story name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicName, storyName, storyId } = resolveStory(cmd);
        const store = new TaskStore(db);
        store.delete(storyId, name);
        writeJSON({ name, epic: epicName, story: storyName, deleted: true });
      } catch (err) {
        writeError(err);
      }
    });
}
