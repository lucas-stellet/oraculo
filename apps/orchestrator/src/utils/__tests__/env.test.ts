import { describe, it, expect } from "bun:test";
import { cleanEnv } from "../env.ts";

describe("cleanEnv", () => {
  it("strips CLAUDECODE from env", () => {
    const original = process.env.CLAUDECODE;
    process.env.CLAUDECODE = "1";
    const env = cleanEnv();
    expect(env.CLAUDECODE).toBeUndefined();
    if (original) process.env.CLAUDECODE = original;
    else delete process.env.CLAUDECODE;
  });

  it("preserves other env vars", () => {
    const env = cleanEnv();
    expect(env.PATH).toBeDefined();
  });
});
