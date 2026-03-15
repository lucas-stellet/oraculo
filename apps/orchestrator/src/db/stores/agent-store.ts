import type { Database } from "bun:sqlite";
import type { Agent } from "../../domain/entities.ts";

export class AgentStore {
  constructor(private db: Database) {}

  Create(
    sessionId: string,
    name: string,
    type_: string,
    taskId?: number,
  ): Agent {
    const now = new Date().toISOString().replace("T", " ").slice(0, 19);
    this.db.run(
      "INSERT INTO agents (session_id, name, type, status, started_at, task_id) VALUES (?, ?, ?, 'active', ?, ?)",
      [sessionId, name, type_, now, taskId ?? null],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;
    return this.db
      .query<Agent, [number]>(
        "SELECT id, session_id, name, type, status, started_at, stopped_at, task_id FROM agents WHERE id = ?",
      )
      .get(id)!;
  }

  Stop(id: number): void {
    const now = new Date().toISOString().replace("T", " ").slice(0, 19);
    this.db.run(
      "UPDATE agents SET status = 'completed', stopped_at = ? WHERE id = ?",
      [now, id],
    );
  }

  ListBySession(sessionId: string): Agent[] {
    return this.db
      .query<Agent, [string]>(
        "SELECT id, session_id, name, type, status, started_at, stopped_at, task_id FROM agents WHERE session_id = ? ORDER BY started_at",
      )
      .all(sessionId);
  }

  Active(): Agent[] {
    return this.db
      .query<Agent, []>(
        "SELECT id, session_id, name, type, status, started_at, stopped_at, task_id FROM agents WHERE status = 'active' ORDER BY started_at",
      )
      .all();
  }
}
