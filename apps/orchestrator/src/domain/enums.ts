// --- Task ---
export type TaskStatus = "pending" | "in_progress" | "completed" | "failed";
export const TASK_STATUSES: TaskStatus[] = ["pending", "in_progress", "completed", "failed"];

// --- Approval ---
export type ApprovalStatus = "pending" | "approved" | "rejected" | "needs_revision";
export const APPROVAL_STATUSES: ApprovalStatus[] = ["pending", "approved", "rejected", "needs_revision"];

export type ApprovalType = "qa-escalation" | "execution-plan" | "design";
export const APPROVAL_TYPES: ApprovalType[] = ["qa-escalation", "execution-plan", "design"];

export type Verdict = "approved" | "rejected" | "needs_revision";
export const VERDICTS: Verdict[] = ["approved", "rejected", "needs_revision"];

// --- Session ---
export type SessionType = "epic" | "story" | "plan" | "execute" | "validate";
export const SESSION_TYPES: SessionType[] = ["epic", "story", "plan", "execute", "validate"];

export type SessionStatus = "active" | "completed" | "abandoned";
export const SESSION_STATUSES: SessionStatus[] = ["active", "completed", "abandoned"];

// --- Knowledge ---
export type Category = "pattern" | "convention" | "constraint" | "dependency" | "test" | "architecture";
export const CATEGORIES: Category[] = ["pattern", "convention", "constraint", "dependency", "test", "architecture"];

export type Confidence = "high" | "medium" | "low";
export const CONFIDENCES: Confidence[] = ["high", "medium", "low"];

// --- Review ---
export type ReviewVerdict = "approved" | "rejected";
export const REVIEW_VERDICTS: ReviewVerdict[] = ["approved", "rejected"];

// --- Version ---
export type VersionType = "epic" | "story";
export const VERSION_TYPES: VersionType[] = ["epic", "story"];

// --- Phase mapping ---
export const SESSION_TYPE_TO_PHASE: Record<SessionType, string> = {
  epic: "discover",
  story: "define",
  plan: "plan",
  execute: "execute",
  validate: "validate",
};

export const PHASES: Record<SessionType, string[]> = {
  epic: ["discover", "define", "plan", "execute", "validate"],
  story: ["define", "plan", "execute", "validate"],
  plan: ["plan", "execute", "validate"],
  execute: ["execute", "validate"],
  validate: ["validate"],
};

export function phaseIndex(sessionType: SessionType, phase: string): number {
  return PHASES[sessionType]?.indexOf(phase) ?? -1;
}
