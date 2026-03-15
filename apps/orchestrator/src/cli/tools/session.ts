import { Command } from "commander";
import { EpicStore } from "../../db/stores/epic-store.ts";
import { SessionStore } from "../../db/stores/session-store.ts";
import { SESSION_TYPES, PHASES } from "../../domain/enums.ts";
import type { SessionType } from "../../domain/enums.ts";
import { writeJSON, writeError } from "../../utils/output.ts";
import { dbFromCmd } from "../context.ts";

export function sessionCommands(): Command {
  const cmd = new Command("session").description("Manage skill sessions");

  cmd.addCommand(sessionInitCmd());
  cmd.addCommand(sessionStatusCmd());
  cmd.addCommand(sessionStateCmd());
  cmd.addCommand(sessionCloseCmd());

  return cmd;
}

function sessionInitCmd(): Command {
  return new Command("init")
    .description("Start a new session")
    .requiredOption(
      "--type <type>",
      "Session type (epic, story, plan, execute, validate)",
    )
    .requiredOption("--epic <name>", "Epic name")
    .option("--description <desc>", "Epic description", "")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);

        // Create or get epic
        let epic;
        try {
          epic = epicStore.create(opts.epic, opts.description);
        } catch {
          epic = epicStore.getByName(opts.epic);
        }

        const sessionStore = new SessionStore(db);
        const session = sessionStore.Create(
          opts.type as SessionType,
          epic.id,
        );

        writeJSON({
          id: session.id,
          type: session.type,
          epic: opts.epic,
          status: session.status,
          created: true,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function sessionStatusCmd(): Command {
  return new Command("status")
    .description("Check for an active session")
    .requiredOption("--type <type>", "Session type")
    .requiredOption("--epic <name>", "Epic name")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const epicStore = new EpicStore(db);

        let epic;
        try {
          epic = epicStore.getByName(opts.epic);
        } catch {
          writeJSON({ active: false });
          return;
        }

        const sessionStore = new SessionStore(db);
        const session = sessionStore.ActiveByEpic(epic.id);
        if (!session || session.type !== opts.type) {
          writeJSON({ active: false });
          return;
        }

        // Get phase info
        const phases = db
          .query<{ name: string }, [string]>(
            "SELECT phase as name FROM session_phases WHERE session_id = ? ORDER BY completed_at",
          )
          .all(session.id);
        const completedNames = phases.map((p) => p.name);

        // Determine current phase from session type phases
        const allPhases: string[] = PHASES[session.type as SessionType] ?? [];
        let currentPhase = "";
        for (const ph of allPhases) {
          if (!completedNames.includes(ph)) {
            currentPhase = ph;
            break;
          }
        }

        writeJSON({
          active: true,
          id: session.id,
          type: session.type,
          epic: opts.epic,
          current_phase: currentPhase,
          completed_phases: completedNames,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function sessionStateCmd(): Command {
  return new Command("state")
    .description("Get full session state with phase data")
    .requiredOption("--session <id>", "Session ID")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const sessionStore = new SessionStore(db);
        const session = sessionStore.Get(opts.session);

        const phases = db
          .query<{ name: string; data: string }, [string]>(
            "SELECT phase as name, data FROM session_phases WHERE session_id = ? ORDER BY completed_at",
          )
          .all(session.id);

        const phaseData: Record<string, unknown> = {};
        for (const p of phases) {
          try {
            phaseData[p.name] = JSON.parse(p.data);
          } catch {
            phaseData[p.name] = p.data;
          }
        }

        // Determine current phase
        const allPhases: string[] =
          PHASES[session.type as SessionType] ?? [];
        const completedNames = phases.map((p) => p.name);
        let currentPhase = "";
        for (const ph of allPhases) {
          if (!completedNames.includes(ph)) {
            currentPhase = ph;
            break;
          }
        }

        writeJSON({
          id: session.id,
          type: session.type,
          status: session.status,
          current_phase: currentPhase,
          phases: phaseData,
        });
      } catch (err) {
        writeError(err);
      }
    });
}

function sessionCloseCmd(): Command {
  return new Command("close")
    .description("Close a session")
    .requiredOption("--session <id>", "Session ID")
    .option("--reason <reason>", "Close reason (abandoned)")
    .action((opts: any, cmd: Command) => {
      try {
        const db = dbFromCmd(cmd);
        const sessionStore = new SessionStore(db);

        if (opts.reason === "abandoned") {
          sessionStore.Abandon(opts.session);
          writeJSON({
            id: opts.session,
            status: "abandoned",
            closed: true,
          });
        } else {
          sessionStore.Complete(opts.session);
          writeJSON({
            id: opts.session,
            status: "completed",
            closed: true,
          });
        }
      } catch (err) {
        writeError(err);
      }
    });
}
