export type ApprovalStatus =
  | "none"
  | "pending"
  | "approved"
  | "rejected"
  | "needs_revision";

export type TaskStatus = "pending" | "in_progress" | "completed" | "failed";

export type Phase = "discover" | "plan" | "execute" | "validate";

export interface Epic {
  id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
}

export interface EpicSummary extends Epic {
  phase: Phase;
  phase_status: string;
  story_count: number;
  task_count: number;
  completed_task_count: number;
}
