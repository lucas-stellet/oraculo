import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import type { Story, ApprovalType } from "./types"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function deriveStoryStatus(story: Story): "pending" | "in_progress" | "completed" | "failed" {
  if (story.task_count === 0) return "pending";
  if (story.completed_task_count === story.task_count) return "completed";
  if (story.failed_task_count > 0 && story.completed_task_count + story.failed_task_count === story.task_count) return "failed";
  if (story.completed_task_count > 0 || story.failed_task_count > 0) return "in_progress";
  return "pending";
}

export function formatDuration(startedAt: string | null, completedAt: string | null): string | null {
  if (!startedAt) return null;
  const start = new Date(startedAt);
  const end = completedAt ? new Date(completedAt) : new Date();
  const diffMs = end.getTime() - start.getTime();
  const totalSec = Math.floor(diffMs / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  if (min > 0) return `${min}m ${sec}s`;
  return `${sec}s`;
}

export function approvalDisplayTitle(type: ApprovalType): string {
  const labels: Record<string, string> = {
    "qa-escalation": "QA Escalation",
    "execution-plan": "Execution Plan",
    "design": "Design Approval",
  };
  return labels[type] || type;
}
