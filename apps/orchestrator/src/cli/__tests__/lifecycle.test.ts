import { describe, test, expect } from "bun:test";
import { parseSseLine } from "../lifecycle.ts";

describe("parseSseLine", () => {
  test("extracts time, level, msg from valid SSE data line", () => {
    const line = 'data: {"time":"2025-01-01T00:00:00Z","level":"INFO","msg":"server.started"}';
    expect(parseSseLine(line)).toBe("2025-01-01T00:00:00Z INFO server.started");
  });

  test("returns raw JSON for data lines without time/level/msg", () => {
    const line = 'data: {"foo":"bar"}';
    expect(parseSseLine(line)).toBe('{"foo":"bar"}');
  });

  test("returns raw text for non-JSON data lines", () => {
    const line = "data: hello world";
    expect(parseSseLine(line)).toBe("hello world");
  });

  test("returns empty string for non-data lines", () => {
    expect(parseSseLine("event: log")).toBe("");
    expect(parseSseLine("id: 123")).toBe("");
    expect(parseSseLine("")).toBe("");
    expect(parseSseLine(": comment")).toBe("");
  });
});
