import { Command } from "commander";
import { SessionStore } from "../../db/stores/session-store.ts";
import { PHASES } from "../../domain/enums.ts";
import type { SessionType } from "../../domain/enums.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function phaseCommands(): Command {
  const cmd = new Command("phase").description("Manage session phases");

  cmd.addCommand(phaseCompleteCmd());

  return cmd;
}

function phaseCompleteCmd(): Command {
  return new Command("complete")
    .description("Complete a phase (reads JSON data from stdin)")
    .argument("<phase>", "Phase name")
    .requiredOption("--session <id>", "Session ID")
    .action(async (phase: string, opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const sessionStore = new SessionStore(db);

        let data = await Bun.stdin.text();
        if (!data.trim()) {
          data = "{}";
        }

        sessionStore.CompletePhase(opts.session, phase, data);

        // Determine next phase
        const session = sessionStore.Get(opts.session);
        const allPhases: string[] =
          PHASES[session.type as SessionType] ?? [];
        const completedPhases = db
          .query<{ name: string }, [string]>(
            "SELECT phase as name FROM session_phases WHERE session_id = ?",
          )
          .all(session.id)
          .map((p) => p.name);

        let next = "";
        for (const ph of allPhases) {
          if (!completedPhases.includes(ph)) {
            next = ph;
            break;
          }
        }

        const result: Record<string, unknown> = {
          phase,
          completed: true,
          next,
        };

        if (session.status === "completed") {
          result.session_closed = true;
        }

        writeJSON(result);
      } catch (err) {
        writeError(err);
      }
    });
}
