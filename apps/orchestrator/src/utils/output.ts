import { DomainError } from "../domain/errors.ts";

export function writeJSON(data: unknown): void {
  console.log(formatJSON(data));
}

export function formatJSON(data: unknown): string {
  return JSON.stringify(data, null, 2);
}

export function writeError(err: unknown): void {
  console.error(formatError(err instanceof Error ? err : new Error(String(err))));
  process.exit(1);
}

export function formatError(err: Error): string {
  const code = err instanceof DomainError ? err.code : "internal";
  return JSON.stringify({ error: code, message: err.message }, null, 2);
}

export function writeTable(headers: string[], rows: string[][]): void {
  console.log(formatTable(headers, rows));
}

export function formatTable(headers: string[], rows: string[][]): string {
  const widths = headers.map((h, i) =>
    Math.max(h.length, ...rows.map((r) => (r[i] ?? "").length)),
  );

  const line = widths.map((w) => "─".repeat(w + 2)).join("┼");
  const header = headers.map((h, i) => ` ${h.padEnd(widths[i])} `).join("│");
  const body = rows
    .map((row) => row.map((cell, i) => ` ${(cell ?? "").padEnd(widths[i])} `).join("│"))
    .join("\n");

  return `${header}\n${line}\n${body}`;
}
