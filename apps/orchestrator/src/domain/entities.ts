import type {
  TaskStatus,
  ApprovalStatus,
  ApprovalType,
  SessionType,
  SessionStatus,
  Category,
  Confidence,
  ReviewVerdict,
  VersionType,
} from "./enums.ts";

// --- Epic ---

export interface Epic {
  id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
}

export interface EpicSummary extends Epic {
  phase: string;
  phase_status: string;
  story_count: number;
  task_count: number;
  completed_task_count: number;
}

// --- Story ---

export interface Story {
  id: number;
  epic_id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
}

export interface StorySummary extends Story {
  task_count: number;
  completed_task_count: number;
  failed_task_count: number;
}

// --- Task ---

export interface Task {
  id: number;
  story_id: number;
  name: string;
  description: string;
  status: TaskStatus;
  failure_reason: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
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

export interface TaskEnriched extends Task {
  depends_on: number[];
  result?: TaskResult;
}

// --- Session ---

export interface Session {
  id: string;
  type: SessionType;
  epic_id?: number;
  status: SessionStatus;
  created_at: string;
  closed_at?: string;
}

export interface SessionPhase {
  session_id: string;
  name: string;
  data: string;
  completed_at: string;
}

// --- Approval ---

export interface Approval {
  id: string;
  type: ApprovalType;
  epic_id?: number;
  story_id?: number;
  content: string;
  previous_version: string;
  status: ApprovalStatus;
  verdict_comment: string;
  requested_at: string;
  decided_at?: string;
}

export interface ApprovalComment {
  id: number;
  selected_text: string;
  comment: string;
  created_at: string;
}

// --- Knowledge ---

export interface Knowledge {
  id: number;
  domain: string;
  category: Category;
  finding: string;
  source_files: string;
  confidence: Confidence;
  created_at: string;
}

// --- Validation ---

export interface Validation {
  id: number;
  story_id: number;
  task_id?: number;
  verdict: string;
  findings: string;
  created_at: string;
}

// --- Agent ---

export interface Agent {
  id: number;
  session_id: string;
  name: string;
  type: string;
  status: string;
  started_at: string;
  stopped_at?: string;
  task_id?: number;
}

// --- ToolEvent ---

export interface ToolEvent {
  id: number;
  session_id: string;
  tool_name: string;
  file_path: string;
  timestamp: string;
}

// --- SessionEvent ---

export interface SessionEvent {
  id: number;
  session_id: string;
  event_type: string;
  payload: string;
  created_at: string;
}

// --- Review ---

export interface Review {
  id: number;
  version_id: number;
  version_type: VersionType;
  verdict: ReviewVerdict;
  comment: string;
  created_at: string;
}

// --- Version ---

export interface EpicVersion {
  id: number;
  epic_id: number;
  number: number;
  content: string;
  created_at: string;
}

export interface StoryVersion {
  id: number;
  story_id: number;
  number: number;
  content: string;
  created_at: string;
}
