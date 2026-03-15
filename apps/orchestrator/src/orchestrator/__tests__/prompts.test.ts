import { describe, it, expect } from "bun:test";
import { buildTaskPrompt } from "../prompts.ts";
import type { TaskEnriched } from "../../domain/index.ts";

function makeTask(overrides: Partial<TaskEnriched> = {}): TaskEnriched {
  return {
    id: 1,
    story_id: 1,
    name: "implement-auth",
    description: "Implement JWT authentication",
    status: "pending",
    failure_reason: "",
    depends_on: [],
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("buildTaskPrompt", () => {
  it("includes task name and description", () => {
    const prompt = buildTaskPrompt(makeTask(), {
      epicName: "user-management",
      storyName: "auth-flow",
      knowledge: [],
    });
    expect(prompt).toContain("# Task: implement-auth");
    expect(prompt).toContain("Implement JWT authentication");
  });

  it("includes epic and story context", () => {
    const prompt = buildTaskPrompt(makeTask(), {
      epicName: "user-management",
      storyName: "auth-flow",
      knowledge: [],
    });
    expect(prompt).toContain("Epic: user-management");
    expect(prompt).toContain("Story: auth-flow");
  });

  it("includes knowledge when provided", () => {
    const prompt = buildTaskPrompt(makeTask(), {
      epicName: "e",
      storyName: "s",
      knowledge: ["Use bcrypt for hashing", "JWT expiry is 24h"],
    });
    expect(prompt).toContain("## Accumulated Knowledge");
    expect(prompt).toContain("Use bcrypt for hashing");
    expect(prompt).toContain("JWT expiry is 24h");
  });

  it("omits knowledge section when empty", () => {
    const prompt = buildTaskPrompt(makeTask(), {
      epicName: "e",
      storyName: "s",
      knowledge: [],
    });
    expect(prompt).not.toContain("Knowledge");
  });
});
