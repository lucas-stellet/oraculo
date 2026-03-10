export type ApprovalStatus =
  | "none"
  | "pending"
  | "approved"
  | "rejected"
  | "needs_revision";

export type TaskStatus = "pending" | "in_progress" | "completed" | "failed";

export type Phase = "discover" | "plan" | "execute" | "validate";

export type StoryStatus = "pending" | "in_progress" | "completed" | "failed";

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
  status: StoryStatus;
  task_count: number;
  completed_task_count: number;
  running_status: string;
  last_activity: string;
}

export interface StoryTask {
  id: number;
  story_id: number;
  name: string;
  status: TaskStatus;
  agent: string | null;
  duration: string | null;
}

export type ApprovalType = "epic-version" | "story-version" | "qa-escalation" | "design" | "execution-plan";

export interface Approval {
  id: number;
  epic_id: number;
  story_name: string | null;
  type: ApprovalType;
  title: string;
  description: string;
  status: "pending" | "approved" | "rejected";
  requested_at: string;
}
