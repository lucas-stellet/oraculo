import type { Database } from "bun:sqlite";
import type { Session, SessionPhase } from "../../domain/entities.ts";
import type { SessionType } from "../../domain/enums.ts";
import { PHASES } from "../../domain/enums.ts";
import { ErrNotFound, ErrInvalidTransition, ErrInvalidPhase, ErrAlreadyExists } from "../../domain/errors.ts";

export class SessionStore {
  constructor(private db: Database) {}

  Create(type_: SessionType, epicId: number | null, description: string = ""): Session {
    const id = crypto.randomUUID();
    this.db.run(
      "INSERT INTO sessions (id, type, epic_id) VALUES (?, ?, ?)",
      [id, type_, epicId],
    );
    return this.Get(id);
  }

  Get(id: string): Session {
    const row = this.db
      .query<Session, [string]>(
        "SELECT id, type, epic_id, status, created_at, closed_at FROM sessions WHERE id = ?",
      )
      .get(id);
    if (!row) throw ErrNotFound("session", id);
    return row;
  }

  ActiveByEpic(epicId: number): Session | null {
    return (
      this.db
        .query<Session, [number]>(
          "SELECT id, type, epic_id, status, created_at, closed_at FROM sessions WHERE epic_id = ? AND status = 'active'",
        )
        .get(epicId) ?? null
    );
  }

  Complete(id: string): void {
    const session = this.Get(id);
    if (session.status !== "active") {
      throw ErrInvalidTransition(session.status, "completed");
    }
    this.db.run(
      "UPDATE sessions SET status = 'completed', closed_at = datetime('now') WHERE id = ?",
      [id],
    );
  }

  Abandon(id: string): void {
    const session = this.Get(id);
    if (session.status !== "active") {
      throw ErrInvalidTransition(session.status, "abandoned");
    }
    this.db.run(
      "UPDATE sessions SET status = 'abandoned', closed_at = datetime('now') WHERE id = ?",
      [id],
    );
  }

  CompletePhase(sessionId: string, name: string, data: string): SessionPhase {
    const session = this.Get(sessionId);
    if (session.status !== "active") {
      throw ErrInvalidTransition(session.status, "complete-phase");
    }

    const phases = PHASES[session.type as SessionType];
    const idx = phases?.indexOf(name) ?? -1;
    if (idx < 0) {
      throw ErrInvalidPhase(name);
    }

    // Check predecessor
    if (idx > 0) {
      const predecessor = phases[idx - 1];
      const count = this.db
        .query<{ cnt: number }, [string, string]>(
          "SELECT COUNT(*) as cnt FROM session_phases WHERE session_id = ? AND phase = ?",
        )
        .get(sessionId, predecessor)!.cnt;
      if (count === 0) {
        throw ErrInvalidTransition(predecessor, name);
      }
    }

    // Check for duplicate
    const existing = this.db
      .query<SessionPhase, [string, string]>(
        "SELECT session_id, phase as name, data, completed_at FROM session_phases WHERE session_id = ? AND phase = ?",
      )
      .get(sessionId, name);
    if (existing) {
      throw ErrAlreadyExists("phase", name);
    }

    this.db.run(
      "INSERT INTO session_phases (session_id, phase, data) VALUES (?, ?, ?)",
      [sessionId, name, data],
    );

    // Auto-close if last phase
    if (idx === phases.length - 1) {
      this.db.run(
        "UPDATE sessions SET status = 'completed', closed_at = datetime('now') WHERE id = ?",
        [sessionId],
      );
    }

    return this.db
      .query<SessionPhase, [string, string]>(
        "SELECT session_id, phase as name, data, completed_at FROM session_phases WHERE session_id = ? AND phase = ?",
      )
      .get(sessionId, name)!;
  }
}
