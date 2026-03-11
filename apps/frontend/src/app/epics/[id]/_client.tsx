"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { useApi } from "@/lib/server-context";
import { deriveStoryStatus } from "@/lib/utils";
import type { Story } from "@/lib/types";
import { useWebSocket } from "@/lib/ws";
import type { WSEvent } from "@/lib/ws";

type DerivedStatus = "pending" | "in_progress" | "completed" | "failed";

const statusConfig: Record<
  DerivedStatus,
  { label: string; bg: string; text: string; dotColor: string }
> = {
  in_progress: {
    label: "In Progress",
    bg: "bg-blue-700",
    text: "text-blue-50",
    dotColor: "bg-blue-400",
  },
  completed: {
    label: "Completed",
    bg: "bg-green-700",
    text: "text-green-50",
    dotColor: "bg-green-400",
  },
  pending: {
    label: "Pending",
    bg: "bg-zinc-700",
    text: "text-zinc-100",
    dotColor: "bg-zinc-400",
  },
  failed: {
    label: "Failed",
    bg: "bg-red-700",
    text: "text-red-50",
    dotColor: "bg-red-400",
  },
};

function deriveRunningLabel(status: DerivedStatus): string {
  switch (status) {
    case "in_progress":
      return "Executing now";
    case "completed":
      return "Completed";
    case "failed":
      return "Failed";
    default:
      return "Not started";
  }
}

const runningStatusColor: Record<string, string> = {
  "Executing now": "text-green-500",
  Completed: "text-[#8ea2bd]",
  Failed: "text-red-500",
  "Not started": "text-zinc-500",
};

export default function EpicOverviewPage() {
  const api = useApi();
  const pathname = usePathname();
  const epicName = pathname.split("/").filter(Boolean)[1];
  const [stories, setStories] = useState<Story[]>([]);

  useEffect(() => {
    api.listStories(epicName).then(setStories).catch(() => setStories([]));
  }, [epicName]);

  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event !== "task_started" && evt.event !== "task_completed") return;
    api.listStories(epicName).then(setStories).catch(() => {});
  }, [epicName]);

  useWebSocket(handleWS);

  const completedStories = stories.filter(
    (s) => deriveStoryStatus(s) === "completed"
  ).length;

  return (
    <div className="flex flex-col gap-6 p-8">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
          <span className="text-xs font-medium tracking-[1.2px] text-blue-400 font-[family-name:var(--font-mono)]">
            EPIC OVERVIEW
          </span>
          <h1 className="text-[32px] font-bold tracking-tight text-[#f5f9ff] font-[family-name:var(--font-display)]">
            {epicName}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-1.5 rounded-full bg-blue-700 px-3 py-1">
            <span className="h-2 w-2 rounded-full bg-blue-400" />
            <span className="text-[13px] font-medium text-blue-50 font-[family-name:var(--font-sans)]">
              In Progress
            </span>
          </span>
          <span className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
            {completedStories} / {stories.length} stories completed
          </span>
        </div>
      </div>

      {/* Story Grid */}
      <div className="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
        {stories.map((story) => {
          const pct =
            story.task_count > 0
              ? Math.round(
                  (story.completed_task_count / story.task_count) * 100
                )
              : 0;
          const storyStatus = deriveStoryStatus(story);
          const config = statusConfig[storyStatus];
          const progressColor =
            storyStatus === "completed"
              ? "bg-green-500"
              : storyStatus === "in_progress"
                ? "bg-blue-600"
                : "bg-zinc-600";
          const runningLabel = deriveRunningLabel(storyStatus);
          const lastActivity = story.updated_at
            ? new Date(story.updated_at).toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })
            : "—";

          return (
            <div
              key={story.id}
              className="flex flex-col overflow-hidden rounded-2xl border border-[#22324a] bg-[#0f172a]"
            >
              {/* Card Header */}
              <div className="flex items-center justify-between px-5 py-4">
                <h3 className="text-base font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
                  {story.name}
                </h3>
                <span
                  className={`inline-flex items-center rounded-xl px-2.5 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)] ${config.bg} ${config.text}`}
                >
                  {config.label}
                </span>
              </div>

              {/* Card Body */}
              <div className="flex flex-1 flex-col justify-center gap-2 px-5">
                <span className="text-[13px] text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                  {pct}% complete ({story.completed_task_count}/{story.task_count}{" "}
                  tasks)
                </span>
                <div className="h-1.5 w-full overflow-hidden rounded-sm bg-[#22324a]">
                  <div
                    className={`h-full rounded-sm transition-all ${progressColor}`}
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </div>

              {/* Card Footer */}
              <div className="mt-auto flex items-center justify-between border-t border-[#22324a] px-5 py-3">
                <div className="flex flex-col gap-0.5">
                  <span
                    className={`text-xs font-medium font-[family-name:var(--font-sans)] ${runningStatusColor[runningLabel] ?? "text-zinc-500"}`}
                  >
                    {runningLabel}
                  </span>
                  <span className="text-[11px] text-zinc-500 font-[family-name:var(--font-mono)]">
                    Last: {lastActivity}
                  </span>
                </div>
                <Link
                  href={`/epics/${epicName}/stories/${encodeURIComponent(story.name)}`}
                  className="inline-flex items-center gap-1 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white transition-colors hover:bg-blue-500"
                >
                  Open
                  <ArrowRight className="h-3.5 w-3.5" />
                </Link>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
