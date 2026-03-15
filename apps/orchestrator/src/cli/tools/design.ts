import { Command } from "commander";
import { join } from "node:path";
import { mkdirSync } from "node:fs";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { StoryStore } from "../../db/stores/story-store.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function designCommands(): Command {
  const cmd = new Command("design").description("Manage design artifacts");

  cmd.addCommand(designSaveCmd());
  cmd.addCommand(designGetCmd());

  return cmd;
}

function designSaveCmd(): Command {
  return new Command("save")
    .description("Save design markdown for a story (reads stdin)")
    .argument("<story-name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        const epic = epicStore.getByName(opts.epic);

        // Verify story exists
        const storyStore = new StoryStore(db);
        storyStore.getByName(epic.id, name);

        const content = await Bun.stdin.text();
        const dir = join(
          ".oraculo",
          "epics",
          opts.epic,
          "stories",
          name,
        );
        mkdirSync(dir, { recursive: true });
        const path = join(dir, "design.md");
        await Bun.write(path, content);

        writeJSON({ name, epic: opts.epic, path, saved: true });
      } catch (err) {
        writeError(err);
      }
    });
}

function designGetCmd(): Command {
  return new Command("get")
    .description("Get design markdown for a story")
    .argument("<story-name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        epicStore.getByName(opts.epic); // validate epic exists

        const path = join(
          ".oraculo",
          "epics",
          opts.epic,
          "stories",
          name,
          "design.md",
        );
        const file = Bun.file(path);
        if (!(await file.exists())) {
          writeError(
            new Error(`story "${name}" design: file not found`),
          );
          return;
        }
        process.stdout.write(await file.text());
      } catch (err) {
        writeError(err);
      }
    });
}
