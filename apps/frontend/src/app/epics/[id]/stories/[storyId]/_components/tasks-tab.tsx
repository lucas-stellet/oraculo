"use client";

import { useState } from "react";
import { Check, ChevronRight, ChevronDown } from "lucide-react";
import type { StoryTask, TaskStatus } from "@/lib/types";
import { cn, formatDuration } from "@/lib/utils";

const taskStatusConfig: Record<
  TaskStatus,
  { label: string; bg: string; text: string; checkBg: string; checkIcon: string }
> = {
  completed: {
    label: "completed",
    bg: "bg-green-700",
    text: "text-green-50",
    checkBg: "bg-green-500",
    checkIcon: "text-white",
  },
  in_progress: {
    label: "in_progress",
    bg: "bg-blue-700",
    text: "text-blue-50",
    checkBg: "bg-blue-700",
    checkIcon: "text-blue-400",
  },
  pending: {
    label: "pending",
    bg: "bg-zinc-700",
    text: "text-zinc-100",
    checkBg: "bg-zinc-700",
    checkIcon: "text-transparent",
  },
  failed: {
    label: "failed",
    bg: "bg-red-700",
    text: "text-red-50",
    checkBg: "bg-red-700",
    checkIcon: "text-white",
  },
};

export function TasksTab({ tasks, activeAgents }: { tasks: StoryTask[]; activeAgents?: Record<number, string> }) {
  const [expandedId, setExpandedId] = useState<number | null>(null);

  return (
    <div className="flex flex-col gap-0.5">
      {tasks.map((task) => {
        const config = taskStatusConfig[task.status];
        const isExpanded = expandedId === task.id;
        const hasResult = task.result !== null;
        const duration = formatDuration(task.started_at, task.completed_at);

        return (
          <div key={task.id}>
            <button
              onClick={() => hasResult ? setExpandedId(isExpanded ? null : task.id) : undefined}
              className={cn(
                "flex w-full items-center gap-3 rounded-lg bg-[#0f172a] px-4 py-3 text-left",
                hasResult && "cursor-pointer hover:bg-[#131d30]",
                !hasResult && "cursor-default"
              )}
            >
              {/* Checkbox */}
              <div
                className={cn(
                  "flex h-5 w-5 shrink-0 items-center justify-center rounded",
                  config.checkBg
                )}
              >
                <Check className={cn("h-3.5 w-3.5", config.checkIcon)} />
              </div>

              {/* Name + Description */}
              <div className="flex min-w-0 flex-1 flex-col">
                <span className="truncate text-sm font-medium text-[#f5f9ff] font-[family-name:var(--font-sans)]">
                  {task.name}
                </span>
                <span className="truncate text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                  {task.description}
                </span>
              </div>

              {/* Status Badge */}
              <span
                className={cn(
                  "inline-flex shrink-0 items-center rounded px-2 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)]",
                  config.bg,
                  config.text
                )}
              >
                {config.label}
              </span>

              {/* Agent Badge */}
              {task.status === "in_progress" && activeAgents?.[task.id] && (
                <span className="inline-flex items-center gap-1 rounded-md bg-blue-900/50 px-2 py-0.5 text-[11px] text-blue-300 font-[family-name:var(--font-mono)]">
                  <span className="h-1.5 w-1.5 rounded-full bg-blue-400 animate-pulse" />
                  {activeAgents[task.id]}
                </span>
              )}

              {/* Duration */}
              <span className="w-16 shrink-0 text-right text-xs text-zinc-500 font-[family-name:var(--font-mono)]">
                {duration ?? "\u2014"}
              </span>

              {/* Chevron */}
              <div className="w-5 shrink-0">
                {hasResult && (
                  isExpanded
                    ? <ChevronDown className="h-4 w-4 text-[#8ea2bd]" />
                    : <ChevronRight className="h-4 w-4 text-[#8ea2bd]" />
                )}
              </div>
            </button>

            {/* Expanded Detail Panel */}
            {isExpanded && task.result && (
              <div className="ml-8 mt-1 mb-1 rounded-lg border border-[#22324a] bg-[#0b1120] p-4">
                <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-blue-400 font-[family-name:var(--font-mono)]">
                  Task Result
                </p>
                <p className="mb-4 text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
                  {task.result.summary}
                </p>

                {task.failure_reason && (
                  <div className="mb-4 rounded-lg border border-red-900/50 bg-red-950/30 p-3">
                    <p className="text-xs font-semibold uppercase tracking-wider text-red-400 font-[family-name:var(--font-mono)] mb-1">
                      Failure Reason
                    </p>
                    <p className="text-sm text-red-200 font-[family-name:var(--font-sans)]">
                      {task.failure_reason}
                    </p>
                  </div>
                )}

                {task.result.files_modified.length > 0 && (
                  <div className="mb-3">
                    <p className="mb-1.5 text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                      Files modified
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {task.result.files_modified.map((f) => (
                        <span
                          key={f}
                          className="rounded bg-[#1e293b] px-2 py-0.5 text-xs text-[#f5f9ff] font-[family-name:var(--font-mono)]"
                        >
                          {f}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {task.result.skills_used.length > 0 && (
                  <div className="mb-3">
                    <p className="mb-1.5 text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                      Skills used
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {task.result.skills_used.map((s) => (
                        <span
                          key={s}
                          className="rounded bg-[#1e1b4b] px-2 py-0.5 text-xs text-indigo-300 font-[family-name:var(--font-mono)]"
                        >
                          {s}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {task.depends_on.length > 0 && (
                  <div>
                    <p className="mb-1.5 text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                      Dependencies
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {task.depends_on.map((depId) => (
                        <span
                          key={depId}
                          className="rounded bg-[#1e293b] px-2 py-0.5 text-xs text-[#8ea2bd] font-[family-name:var(--font-mono)]"
                        >
                          Task #{depId}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
