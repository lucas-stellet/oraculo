export interface AgentDefinition {
  name: string;
  type: "code" | "research" | "qa";
  model: string;
  systemPrompt: string;
  tools: string[];
  maxTurns: number;
}

export const CODE_AGENT: AgentDefinition = {
  name: "code-agent",
  type: "code",
  model: "claude-sonnet-4-20250514",
  systemPrompt: [
    "You are a code agent working on a software project.",
    "You write, edit, and refactor code following the project's conventions.",
    "Always run tests after making changes to verify correctness.",
    "Follow TDD: write failing tests first, then implement, then refactor.",
  ].join("\n"),
  tools: ["Read", "Write", "Edit", "Bash", "Glob", "Grep"],
  maxTurns: 50,
};

export const RESEARCH_AGENT: AgentDefinition = {
  name: "research-agent",
  type: "research",
  model: "claude-sonnet-4-20250514",
  systemPrompt: [
    "You are a research agent analyzing a codebase.",
    "You explore code structure, find patterns, and document findings.",
    "You never modify code — only read and analyze.",
    "Provide detailed, structured findings with file paths and code references.",
  ].join("\n"),
  tools: ["Read", "Glob", "Grep", "Bash"],
  maxTurns: 30,
};

export const QA_AGENT: AgentDefinition = {
  name: "qa-agent",
  type: "qa",
  model: "claude-sonnet-4-20250514",
  systemPrompt: [
    "You are a QA agent validating code changes.",
    "You run tests, check for regressions, and verify acceptance criteria.",
    "You have a clean context — no knowledge of implementation details.",
    "Report findings objectively with pass/fail verdicts.",
  ].join("\n"),
  tools: ["Read", "Bash", "Glob", "Grep"],
  maxTurns: 30,
};

/**
 * Returns the appropriate agent definition for a given task type.
 */
export function agentForTaskType(type: string): AgentDefinition {
  switch (type) {
    case "research":
      return RESEARCH_AGENT;
    case "qa":
      return QA_AGENT;
    default:
      return CODE_AGENT;
  }
}
