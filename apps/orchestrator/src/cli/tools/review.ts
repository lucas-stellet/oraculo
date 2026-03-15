import { Command } from "commander";
import { ReviewStore } from "../../db/stores/review-store.ts";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { StoryStore } from "../../db/stores/story-store.ts";
import { REVIEW_VERDICTS, VERSION_TYPES } from "../../domain/enums.ts";
import type { ReviewVerdict, VersionType } from "../../domain/enums.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function reviewCommands(): Command {
  const cmd = new Command("review").description("Manage document reviews");

  cmd.addCommand(reviewCreateCmd());
  cmd.addCommand(reviewGetCmd());
  cmd.addCommand(reviewListCmd());

  return cmd;
}

function reviewCreateCmd(): Command {
  return new Command("create")
    .description("Create a review for a document version")
    .argument("<version-id>", "Version ID")
    .requiredOption("--type <type>", "Version type (epic or story)")
    .requiredOption(
      "--verdict <verdict>",
      "Review verdict (approved or rejected)",
    )
    .option("--comment <comment>", "Optional review comment", "")
    .action((versionIdStr: string, opts: any, cmd: Command) => {
      try {
        const versionId = parseInt(versionIdStr, 10);
        if (isNaN(versionId)) {
          writeError(
            new Error(`invalid version-id "${versionIdStr}"`),
          );
          return;
        }

        const versionType = opts.type as string;
        if (!VERSION_TYPES.includes(versionType as VersionType)) {
          writeError(
            new Error(
              `invalid type "${versionType}": must be "epic" or "story"`,
            ),
          );
          return;
        }

        const verdict = opts.verdict as string;
        if (!REVIEW_VERDICTS.includes(verdict as ReviewVerdict)) {
          writeError(
            new Error(
              `invalid verdict "${verdict}": must be "approved" or "rejected"`,
            ),
          );
          return;
        }

        const db = dbFromCmd(cmd);
        const store = new ReviewStore(db);
        const review = store.Create(
          versionId,
          versionType as VersionType,
          verdict as ReviewVerdict,
          opts.comment,
        );
        writeJSON(review);
      } catch (err) {
        writeError(err);
      }
    });
}

function reviewGetCmd(): Command {
  return new Command("get")
    .description("Get a review by ID")
    .argument("<review-id>", "Review ID")
    .action((reviewIdStr: string, opts: any, cmd: Command) => {
      try {
        const id = parseInt(reviewIdStr, 10);
        if (isNaN(id)) {
          writeError(new Error(`invalid review-id "${reviewIdStr}"`));
          return;
        }

        const db = dbFromCmd(cmd);
        const store = new ReviewStore(db);
        writeJSON(store.Get(id));
      } catch (err) {
        writeError(err);
      }
    });
}

function reviewListCmd(): Command {
  return new Command("list")
    .description("List reviews for an epic or story")
    .requiredOption("--epic <name>", "Epic name")
    .option("--story <name>", "Story name (optional, narrows to story)")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        const epic = epicStore.getByName(opts.epic);

        const store = new ReviewStore(db);

        if (opts.story) {
          const storyStore = new StoryStore(db);
          const story = storyStore.getByName(epic.id, opts.story);
          writeJSON(store.ListByStory(story.id));
        } else {
          writeJSON(store.ListByEpic(epic.id));
        }
      } catch (err) {
        writeError(err);
      }
    });
}
