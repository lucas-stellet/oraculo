import type { Database } from "bun:sqlite";
import type { SessionEvent } from "../../domain/entities.ts";

export class SessionEventStore {
  constructor(private db: Database) {}

  Create(sessionId: string, type_: string, data: string): SessionEvent {
    // Must have a claude_session row first
    this.db.run(
      "INSERT OR IGNORE INTO claude_sessions (id) VALUES (?)",
      [sessionId],
    );
    this.db.run(
      "INSERT INTO session_events (session_id, event_type, payload) VALUES (?, ?, ?)",
      [sessionId, type_, data],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;
    return this.db
      .query<SessionEvent, [number]>(
        "SELECT id, session_id, event_type, payload, created_at FROM session_events WHERE id = ?",
      )
      .get(id)!;
  }

  ListBySession(sessionId: string): SessionEvent[] {
    return this.db
      .query<SessionEvent, [string]>(
        "SELECT id, session_id, event_type, payload, created_at FROM session_events WHERE session_id = ? ORDER BY created_at",
      )
      .all(sessionId);
  }
}
