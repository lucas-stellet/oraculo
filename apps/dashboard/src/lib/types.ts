export type ApprovalStatus = "none" | "pending" | "approved" | "rejected" | "needs_revision";
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

export interface Story {
  id: number;
  epic_id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
  task_count: number;
  completed_task_count: number;
  failed_task_count: number;
}

export interface StoryTask {
  id: number;
  story_id: number;
  name: string;
  description: string;
  status: TaskStatus;
  failure_reason: string;
  created_at: string;
  updated_at: string;
  started_at: string | null;
  completed_at: string | null;
  depends_on: number[];
  result: TaskResult | null;
}

export interface TaskResult {
  id: number;
  task_id: number;
  summary: string;
  logs: string;
  skills_used: string[];
  files_modified: string[];
  created_at: string;
}

export interface StoryVersion {
  id: number;
  story_id: number;
  number: number;
  content: string;
  created_at: string;
}

export interface Review {
  id: number;
  version_id: number;
  version_type: "epic" | "story";
  verdict: "approved" | "rejected";
  comment: string;
  created_at: string;
}

export interface Validation {
  id: number;
  story_id: number;
  task_id: number | null;
  verdict: "approved" | "rejected";
  findings: string;
  created_at: string;
}

export type ApprovalType = "qa-escalation" | "design" | "execution-plan";

export interface Approval {
  id: string;
  type: ApprovalType;
  epic_id: number | null;
  story_id: number | null;
  content: string;
  previous_version: string;
  status: ApprovalStatus;
  verdict_comment: string;
  requested_at: string;
  decided_at: string | null;
}

export interface InlineComment {
  id: number;
  selected_text: string;
  comment: string;
  created_at: string;
}
