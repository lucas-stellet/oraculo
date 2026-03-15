import { Command } from "commander";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { StoryStore } from "../../db/stores/story-store.ts";
import { ApprovalStore } from "../../db/stores/approval-store.ts";
import { APPROVAL_TYPES, VERDICTS } from "../../domain/enums.ts";
import type { ApprovalType, Verdict } from "../../domain/enums.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function approvalCommands(): Command {
  const cmd = new Command("approval").description("Manage approval gates");

  cmd.addCommand(approvalRequestCmd());
  cmd.addCommand(approvalStatusCmd());
  cmd.addCommand(approvalListCmd());
  cmd.addCommand(approvalVerdictCmd());

  return cmd;
}

function approvalRequestCmd(): Command {
  return new Command("request")
    .description("Create a new approval request (reads content from stdin)")
    .requiredOption(
      "--type <type>",
      "Approval type (qa-escalation|execution-plan|design)",
    )
    .requiredOption("--epic <name>", "Parent epic name")
    .option("--story <name>", "Story name")
    .action(async (opts: any, cmd: Command) => {
      try {
        const approvalType = opts.type as string;
        if (!APPROVAL_TYPES.includes(approvalType as ApprovalType)) {
          writeError(
            new Error(
              `invalid approval type "${approvalType}": must be one of ${APPROVAL_TYPES.join(", ")}`,
            ),
          );
          return;
        }

        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);
        const epic = epicStore.getByName(opts.epic);

        let storyId: number | null = null;
        if (opts.story) {
          const storyStore = new StoryStore(db);
          const story = storyStore.getByName(epic.id, opts.story);
          storyId = story.id;
        }

        const content = await Bun.stdin.text();
        const store = new ApprovalStore(db);
        const approval = store.Create(
          approvalType as ApprovalType,
          epic.id,
          storyId,
          content,
        );
        writeJSON(approval);
      } catch (err) {
        writeError(err);
      }
    });
}

function approvalStatusCmd(): Command {
  return new Command("status")
    .description(
      "Get the status of an approval request (includes inline comments)",
    )
    .argument("<id>", "Approval ID")
    .action((id: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new ApprovalStore(db);
        const approval = store.Get(id);
        const comments = store.ListComments(id);
        writeJSON({ ...approval, comments });
      } catch (err) {
        writeError(err);
      }
    });
}

function approvalListCmd(): Command {
  return new Command("list")
    .description("List approval requests")
    .option("--pending", "Show only pending approvals")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const store = new ApprovalStore(db);
        if (opts.pending) {
          writeJSON(store.ListPending());
        } else {
          writeJSON(store.List());
        }
      } catch (err) {
        writeError(err);
      }
    });
}

function approvalVerdictCmd(): Command {
  return new Command("verdict")
    .description("Record a verdict on an approval request")
    .argument("<id>", "Approval ID")
    .requiredOption(
      "--verdict <verdict>",
      "Verdict (approved|rejected|needs_revision)",
    )
    .option("--comment <comment>", "Optional comment", "")
    .action((id: string, opts: any, cmd: Command) => {
      try {
        const verdict = opts.verdict as string;
        if (!VERDICTS.includes(verdict as Verdict)) {
          writeError(
            new Error(
              `invalid verdict "${verdict}": must be one of ${VERDICTS.join(", ")}`,
            ),
          );
          return;
        }

        const db = dbFromCmd(cmd);
        const store = new ApprovalStore(db);
        store.SetVerdict(id, verdict as Verdict, opts.comment);
        const approval = store.Get(id);
        writeJSON(approval);
      } catch (err) {
        writeError(err);
      }
    });
}
