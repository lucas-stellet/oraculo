import { Command } from "commander";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { StoryStore } from "../../db/stores/story-store.ts";
import { ValidationStore } from "../../db/stores/validation-store.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function validationCommands(): Command {
  const cmd = new Command("validation").description("Manage QA validations");

  cmd.addCommand(validationSaveCmd());

  return cmd;
}

function validationSaveCmd(): Command {
  return new Command("save")
    .description("Save a QA validation verdict for a story")
    .argument("<story-name>", "Story name")
    .requiredOption("--epic <name>", "Parent epic name")
    .requiredOption(
      "--verdict <verdict>",
      "Validation verdict: approved|rejected",
    )
    .option("--findings <findings>", "QA findings as JSON string", "")
    .action((name: string, opts: any, cmd: Command) => {
      try {
        const verdict = opts.verdict as string;
        if (verdict !== "approved" && verdict !== "rejected") {
          writeError(
            new Error(
              `invalid verdict "${verdict}": must be approved or rejected`,
            ),
          );
          return;
        }

        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        const epic = epicStore.getByName(opts.epic);

        const storyStore = new StoryStore(db);
        const story = storyStore.getByName(epic.id, name);

        const valStore = new ValidationStore(db);
        const validation = valStore.Create(
          story.id,
          verdict,
          opts.findings,
        );
        writeJSON(validation);
      } catch (err) {
        writeError(err);
      }
    });
}
