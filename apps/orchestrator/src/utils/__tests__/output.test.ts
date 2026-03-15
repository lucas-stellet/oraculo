import { describe, it, expect } from "bun:test";
import { formatJSON, formatError, formatTable } from "../output.ts";
import { DomainError } from "../../domain/errors.ts";

describe("formatJSON", () => {
  it("returns indented JSON", () => {
    const result = formatJSON({ name: "test" });
    expect(result).toBe('{\n  "name": "test"\n}');
  });
});

describe("formatError", () => {
  it("maps domain error to code", () => {
    const err = new DomainError("not_found", "task 'x' not found");
    const result = JSON.parse(formatError(err));
    expect(result.error).toBe("not_found");
    expect(result.message).toBe("task 'x' not found");
  });

  it("maps generic error", () => {
    const err = new Error("something broke");
    const result = JSON.parse(formatError(err));
    expect(result.error).toBe("internal");
    expect(result.message).toBe("something broke");
  });
});

describe("formatTable", () => {
  it("renders aligned table", () => {
    const result = formatTable(
      ["Name", "Status"],
      [
        ["task-1", "pending"],
        ["task-2", "completed"],
      ],
    );
    expect(result).toContain("Name");
    expect(result).toContain("Status");
    expect(result).toContain("task-1");
    expect(result).toContain("pending");
  });
});
