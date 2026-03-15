import { Command } from "commander";
import { join } from "node:path";
import { mkdirSync } from "node:fs";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { VersionStore } from "../../db/stores/version-store.ts";
import { ApprovalStore } from "../../db/stores/approval-store.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function epicCommands(): Command {
  const cmd = new Command("epic").description("Manage epics");

  cmd.addCommand(epicInitCmd());
  cmd.addCommand(epicSaveCmd());
  cmd.addCommand(epicGetCmd());
  cmd.addCommand(epicListCmd());
  cmd.addCommand(epicUpdateCmd());
  cmd.addCommand(epicDeleteCmd());
  cmd.addCommand(epicVersionCmd());
  cmd.addCommand(epicVersionsCmd());

  return cmd;
}

function epicInitCmd(): Command {
  return new Command("init")
    .description("Create a new epic")
    .argument("<name>", "Epic name")
    .option("--description <desc>", "Epic description", "")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new EpicStore(db);
        store.create(name, opts.description);
        writeJSON({
          name,
          path: join(".oraculo", "epics", name),
          created: true,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function epicSaveCmd(): Command {
  return new Command("save")
    .description("Save requirements markdown for an epic (reads stdin)")
    .argument("<name>", "Epic name")
    .action(async (name: string) => {
      try {
        const content = await Bun.stdin.text();
        const dir = join(".oraculo", "epics", name);
        mkdirSync(dir, { recursive: true });
        const path = join(dir, "requirements.md");
        await Bun.write(path, content);
        writeJSON({ name, path, saved: true });
      } catch (err) {
        writeError(err);
      }
    });
}

function epicGetCmd(): Command {
  return new Command("get")
    .description("Get requirements markdown for an epic")
    .argument("<name>", "Epic name")
    .action(async (name: string) => {
      try {
        const path = join(".oraculo", "epics", name, "requirements.md");
        const file = Bun.file(path);
        if (!(await file.exists())) {
          writeError(new Error(`epic "${name}" requirements: file not found`));
          return;
        }
        const content = await file.text();
        process.stdout.write(content);
      } catch (err) {
        writeError(err);
      }
    });
}

function epicListCmd(): Command {
  return new Command("list")
    .description("List all epics")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new EpicStore(db);
        writeJSON(store.list());
      } catch (err) {
        writeError(err);
      }
    });
}

function epicUpdateCmd(): Command {
  return new Command("update")
    .description("Update an epic's description")
    .argument("<name>", "Epic name")
    .option("--description <desc>", "New epic description", "")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new EpicStore(db);
        const epic = store.update(name, opts.description);
        writeJSON(epic);
      } catch (err) {
        writeError(err);
      }
    });
}

function epicDeleteCmd(): Command {
  return new Command("delete")
    .description("Delete an epic")
    .argument("<name>", "Epic name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new EpicStore(db);
        store.delete(name);
        writeJSON({ name, deleted: true });
      } catch (err) {
        writeError(err);
      }
    });
}

function epicVersionCmd(): Command {
  return new Command("version")
    .description("Create a new version of epic requirements (reads stdin)")
    .argument("<name>", "Epic name")
    .action(async (name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        const epic = epicStore.getByName(name);

        const content = await Bun.stdin.text();

        const versionStore = new VersionStore(db);
        const version = versionStore.CreateEpicVersion(epic.id, content);

        const approvalStore = new ApprovalStore(db);
        const approval = approvalStore.Create("design", epic.id, null, content);

        writeJSON({
          version_id: version.id,
          approval_id: approval.id,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function epicVersionsCmd(): Command {
  return new Command("versions")
    .description("List all versions of an epic")
    .argument("<name>", "Epic name")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        const epic = epicStore.getByName(name);

        const versionStore = new VersionStore(db);
        writeJSON(versionStore.ListEpicVersions(epic.id));
      } catch (err) {
        writeError(err);
      }
    });
}
