import { describe, it, expect } from "bun:test";
import { DomainError, ErrNotFound, ErrCyclicDependency, ErrInvalidTransition } from "../errors.ts";

describe("DomainError", () => {
  it("creates error with code", () => {
    const err = ErrNotFound("task", "abc");
    expect(err).toBeInstanceOf(DomainError);
    expect(err.code).toBe("not_found");
    expect(err.message).toBe("task 'abc' not found");
  });

  it("creates cyclic dependency error", () => {
    const err = ErrCyclicDependency("a", "b");
    expect(err.code).toBe("cyclic_dependency");
  });

  it("creates invalid transition error", () => {
    const err = ErrInvalidTransition("pending", "completed");
    expect(err.code).toBe("invalid_transition");
  });
});
