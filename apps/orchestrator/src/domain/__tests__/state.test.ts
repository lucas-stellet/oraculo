import { describe, it, expect } from "bun:test";
import { canTransition } from "../state.ts";

describe("canTransition", () => {
  it("allows pending → in_progress", () => {
    expect(canTransition("pending", "in_progress")).toBe(true);
  });

  it("allows in_progress → completed", () => {
    expect(canTransition("in_progress", "completed")).toBe(true);
  });

  it("allows in_progress → failed", () => {
    expect(canTransition("in_progress", "failed")).toBe(true);
  });

  it("allows failed → pending (retry)", () => {
    expect(canTransition("failed", "pending")).toBe(true);
  });

  it("rejects pending → completed (skip)", () => {
    expect(canTransition("pending", "completed")).toBe(false);
  });

  it("rejects completed → anything", () => {
    expect(canTransition("completed", "pending")).toBe(false);
    expect(canTransition("completed", "in_progress")).toBe(false);
    expect(canTransition("completed", "failed")).toBe(false);
  });

  it("rejects same-state transitions", () => {
    expect(canTransition("pending", "pending")).toBe(false);
  });
});
