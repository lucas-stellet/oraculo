import type { Database } from "bun:sqlite";
import type { ToolEvent } from "../../domain/entities.ts";

export class ToolEventStore {
  constructor(private db: Database) {}

  Create(
    sessionId: string,
    toolName: string,
    filePath: string,
  ): ToolEvent {
    const now = new Date().toISOString().replace("T", " ").slice(0, 19);
    this.db.run(
      "INSERT INTO tool_events (session_id, tool_name, file_path, timestamp) VALUES (?, ?, ?, ?)",
      [sessionId, toolName, filePath, now],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;
    return this.db
      .query<ToolEvent, [number]>(
        "SELECT id, session_id, tool_name, COALESCE(file_path, '') as file_path, timestamp FROM tool_events WHERE id = ?",
      )
      .get(id)!;
  }

  ListBySession(sessionId: string): ToolEvent[] {
    return this.db
      .query<ToolEvent, [string]>(
        "SELECT id, session_id, tool_name, COALESCE(file_path, '') as file_path, timestamp FROM tool_events WHERE session_id = ? ORDER BY timestamp",
      )
      .all(sessionId);
  }
}
