import { describe, it, expect } from "bun:test";
import {
  type TaskStatus,
  type ApprovalStatus,
  type ApprovalType,
  type Verdict,
  type SessionType,
  type SessionStatus,
  type Category,
  type Confidence,
  type ReviewVerdict,
  type VersionType,
  TASK_STATUSES,
  APPROVAL_STATUSES,
  APPROVAL_TYPES,
  VERDICTS,
  SESSION_TYPES,
  SESSION_STATUSES,
  CATEGORIES,
  CONFIDENCES,
  SESSION_TYPE_TO_PHASE,
  PHASES,
  phaseIndex,
} from "../enums.ts";

describe("enums", () => {
  it("defines all task statuses", () => {
    expect(TASK_STATUSES).toEqual(["pending", "in_progress", "completed", "failed"]);
  });

  it("defines all approval types", () => {
    expect(APPROVAL_TYPES).toEqual(["qa-escalation", "execution-plan", "design"]);
  });

  it("defines all session types", () => {
    expect(SESSION_TYPES).toEqual(["epic", "story", "plan", "execute", "validate"]);
  });

  it("maps session types to dashboard phases", () => {
    expect(SESSION_TYPE_TO_PHASE.epic).toBe("discover");
    expect(SESSION_TYPE_TO_PHASE.story).toBe("define");
    expect(SESSION_TYPE_TO_PHASE.plan).toBe("plan");
    expect(SESSION_TYPE_TO_PHASE.execute).toBe("execute");
    expect(SESSION_TYPE_TO_PHASE.validate).toBe("validate");
  });

  it("returns correct phase index", () => {
    expect(phaseIndex("epic", "discover")).toBe(0);
    expect(phaseIndex("epic", "define")).toBe(1);
    expect(phaseIndex("plan", "plan")).toBe(0);
    expect(phaseIndex("plan", "execute")).toBe(1);
  });

  it("returns -1 for unknown phase", () => {
    expect(phaseIndex("epic", "nonexistent")).toBe(-1);
  });

  it("defines phases per session type", () => {
    expect(PHASES.epic).toEqual(["discover", "define", "plan", "execute", "validate"]);
    expect(PHASES.story).toEqual(["define", "plan", "execute", "validate"]);
    expect(PHASES.plan).toEqual(["plan", "execute", "validate"]);
    expect(PHASES.execute).toEqual(["execute", "validate"]);
    expect(PHASES.validate).toEqual(["validate"]);
  });
});
