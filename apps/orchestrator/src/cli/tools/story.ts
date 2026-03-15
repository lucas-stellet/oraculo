import { Command } from "commander";
import { join } from "node:path";
import { mkdirSync } from "node:fs";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { StoryStore } from "../../db/stores/story-store.ts";
import { VersionStore } from "../../db/stores/version-store.ts";
import { ApprovalStore } from "../../db/stores/approval-store.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { APPROVAL_STATUSES } from "../../domain/enums.ts";
import type { ApprovalStatus } from "../../domain/enums.ts";
import { dbFromCmd } from "../context.ts";
import type { Database } from "bun:sqlite";

export function storyCommands(): Command {
  const cmd = new Command("story").description("Manage stories");

  cmd.addCommand(storyInitCmd());
  cmd.addCommand(storySaveCmd());
  cmd.addCommand(storyGetCmd());
  cmd.addCommand(storyListCmd());
  cmd.addCommand(storyUpdateCmd());
  cmd.addCommand(storyUpdateStatusCmd());
  cmd.addCommand(storyDeleteCmd());
  cmd.addCommand(storyVersionCmd());
  cmd.addCommand(storyVersionsCmd());

  return cmd;
}

function resolveEpic(cmd: Command): { db: Database; epicName: string; epicId: number } {
  const db = dbFromCmd(cmd);
  const epicName = cmd.opts().epic ?? cmd.parent?.opts().epic;
  const epicStore = new EpicStore(db);
  const epic = epicStore.getByName(epicName);
  return { db, epicName, epicId: epic.id };
}

function storyInitCmd(): Command {
  return new Command("init")
    .description("Create a new story within an epic")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .option("--description <desc>", "Story description", "")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicName, epicId } = resolveEpic(cmd);
        const store = new StoryStore(db);
        const story = store.create(epicId, name, opts.description);
        writeJSON({
          name,
          epic: epicName,
          path: join(".oraculo", "epics", epicName, "stories", name),
          created: true,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function storySaveCmd(): Command {
  return new Command("save")
    .description("Save requirements markdown for a story (reads stdin)")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicName, epicId } = resolveEpic(cmd);

        // Upsert DB record
        const store = new StoryStore(db);
        try {
          store.create(epicId, name, "");
        } catch {
          // already exists, ignore
        }

        const content = await Bun.stdin.text();
        const dir = join(".oraculo", "epics", epicName, "stories", name);
        mkdirSync(dir, { recursive: true });
        const path = join(dir, "requirements.md");
        await Bun.write(path, content);

        writeJSON({ name, epic: epicName, path, saved: true });
      } catch (err) {
        writeError(err);
      }
    });
}

function storyGetCmd(): Command {
  return new Command("get")
    .description("Get requirements markdown for a story")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const { epicName } = resolveEpic(cmd);
        const path = join(
          ".oraculo",
          "epics",
          epicName,
          "stories",
          name,
          "requirements.md",
        );
        const file = Bun.file(path);
        if (!(await file.exists())) {
          writeError(
            new Error(`story "${name}" requirements: file not found`),
          );
          return;
        }
        process.stdout.write(await file.text());
      } catch (err) {
        writeError(err);
      }
    });
}

function storyListCmd(): Command {
  return new Command("list")
    .description("List all stories for an epic")
    .requiredOption("--epic <name>", "Parent epic name")
    .action((opts: any, cmd: Command) => {
      try {
        const { db, epicId } = resolveEpic(cmd);
        const store = new StoryStore(db);
        writeJSON(store.list(epicId));
      } catch (err) {
        writeError(err);
      }
    });
}

function storyUpdateCmd(): Command {
  return new Command("update")
    .description("Update a story's description")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .option("--description <desc>", "New story description", "")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicId } = resolveEpic(cmd);
        const store = new StoryStore(db);
        const story = store.update(epicId, name, opts.description);
        writeJSON(story);
      } catch (err) {
        writeError(err);
      }
    });
}

function storyUpdateStatusCmd(): Command {
  return new Command("update-status")
    .description("Update a story's approval status")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption(
      "--status <status>",
      "Approval status: none|pending|approved|rejected|needs_revision",
    )
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const status = opts.status as string;
        const validStatuses = ["none", ...APPROVAL_STATUSES];
        if (!validStatuses.includes(status)) {
          writeError(
            new Error(
              `invalid status "${status}": must be one of ${validStatuses.join(", ")}`,
            ),
          );
          return;
        }

        const { db, epicId } = resolveEpic(cmd);
        const store = new StoryStore(db);
        store.updateApprovalStatus(
          epicId,
          name,
          status as ApprovalStatus | "none",
        );
        const story = store.getByName(epicId, name);
        writeJSON(story);
      } catch (err) {
        writeError(err);
      }
    });
}

function storyDeleteCmd(): Command {
  return new Command("delete")
    .description("Delete a story")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicName, epicId } = resolveEpic(cmd);
        const store = new StoryStore(db);
        store.delete(epicId, name);
        writeJSON({ name, epic: epicName, deleted: true });
      } catch (err) {
        writeError(err);
      }
    });
}

function storyVersionCmd(): Command {
  return new Command("version")
    .description("Create a new version of story definition (reads stdin)")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicId } = resolveEpic(cmd);

        const storyStore = new StoryStore(db);
        const story = storyStore.getByName(epicId, name);

        const content = await Bun.stdin.text();

        const versionStore = new VersionStore(db);
        const version = versionStore.CreateStoryVersion(story.id, content);

        const approvalStore = new ApprovalStore(db);
        const approval = approvalStore.Create(
          "design",
          epicId,
          story.id,
          content,
        );

        writeJSON({
          version_id: version.id,
          approval_id: approval.id,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function storyVersionsCmd(): Command {
  return new Command("versions")
    .description("List all versions of a story")
    .argument("<name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const { db, epicId } = resolveEpic(cmd);
        const storyStore = new StoryStore(db);
        const story = storyStore.getByName(epicId, name);

        const versionStore = new VersionStore(db);
        writeJSON(versionStore.ListStoryVersions(story.id));
      } catch (err) {
        writeError(err);
      }
    });
}
