import type { Database } from "bun:sqlite";
import type { Knowledge } from "../../domain/entities.ts";
import type { Category, Confidence } from "../../domain/enums.ts";

export class MemoryStore {
  constructor(private db: Database) {}

  Store(
    domain: string,
    category: Category,
    finding: string,
    sourceFiles: string,
    confidence: Confidence,
  ): Knowledge {
    this.db.run(
      "INSERT INTO knowledge (domain, category, finding, source_files, confidence) VALUES (?, ?, ?, ?, ?)",
      [domain, category, finding, sourceFiles, confidence],
    );
    const id = this.db
      .query<{ id: number }, []>("SELECT last_insert_rowid() as id")
      .get()!.id;
    return this.db
      .query<Knowledge, [number]>(
        "SELECT id, domain, category, finding, source_files, confidence, created_at FROM knowledge WHERE id = ?",
      )
      .get(id)!;
  }

  Search(query: string, domain?: string): Knowledge[] {
    if (domain) {
      return this.db
        .query<Knowledge, [string, string]>(
          `SELECT k.id, k.domain, k.category, k.finding, k.source_files, k.confidence, k.created_at
           FROM knowledge k
           WHERE k.id IN (SELECT rowid FROM knowledge_fts WHERE knowledge_fts MATCH ?)
           AND k.domain = ?`,
        )
        .all(query, domain);
    }
    return this.db
      .query<Knowledge, [string]>(
        `SELECT k.id, k.domain, k.category, k.finding, k.source_files, k.confidence, k.created_at
         FROM knowledge k
         WHERE k.id IN (SELECT rowid FROM knowledge_fts WHERE knowledge_fts MATCH ?)`,
      )
      .all(query);
  }

  Domains(): string[] {
    return this.db
      .query<{ domain: string }, []>(
        "SELECT DISTINCT domain FROM knowledge ORDER BY domain",
      )
      .all()
      .map((r) => r.domain);
  }
}
