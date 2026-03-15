import { describe, it, expect } from "bun:test";
import {
  agentForTaskType,
  CODE_AGENT,
  RESEARCH_AGENT,
  QA_AGENT,
} from "../agents.ts";

describe("agentForTaskType", () => {
  it("returns CODE_AGENT for 'code'", () => {
    expect(agentForTaskType("code")).toBe(CODE_AGENT);
  });

  it("returns RESEARCH_AGENT for 'research'", () => {
    expect(agentForTaskType("research")).toBe(RESEARCH_AGENT);
  });

  it("returns QA_AGENT for 'qa'", () => {
    expect(agentForTaskType("qa")).toBe(QA_AGENT);
  });

  it("defaults to CODE_AGENT for unknown type", () => {
    expect(agentForTaskType("unknown")).toBe(CODE_AGENT);
  });

  it("defaults to CODE_AGENT for empty string", () => {
    expect(agentForTaskType("")).toBe(CODE_AGENT);
  });
});

describe("agent definitions", () => {
  it("CODE_AGENT has expected properties", () => {
    expect(CODE_AGENT.name).toBe("code-agent");
    expect(CODE_AGENT.type).toBe("code");
    expect(CODE_AGENT.maxTurns).toBe(50);
    expect(CODE_AGENT.tools).toContain("Edit");
    expect(CODE_AGENT.tools).toContain("Write");
  });

  it("RESEARCH_AGENT has no Write tool", () => {
    expect(RESEARCH_AGENT.tools).not.toContain("Write");
    expect(RESEARCH_AGENT.tools).not.toContain("Edit");
  });

  it("QA_AGENT has no Write tool", () => {
    expect(QA_AGENT.tools).not.toContain("Write");
    expect(QA_AGENT.tools).not.toContain("Edit");
  });
});
