"use client";

import { use, useState } from "react";
import Link from "next/link";
import { Check, Bot } from "lucide-react";
import { getEpic, getStory, getTasks } from "@/lib/mock-data";
import type { TaskStatus } from "@/lib/types";
import { cn } from "@/lib/utils";

const tabs = ["Tasks", "Design", "Requirements", "QA"] as const;
type Tab = (typeof tabs)[number];

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

export default function StoryDetailPage({
  params,
}: {
  params: Promise<{ id: string; storyId: string }>;
}) {
  const { id, storyId } = use(params);
  const epicId = Number(id);
  const sId = Number(storyId);

  const epic = getEpic();
  const story = getStory(sId);
  const tasks = getTasks(sId);

  const [activeTab, setActiveTab] = useState<Tab>("Tasks");

  if (!story) {
    return (
      <div className="flex h-full items-center justify-center text-zinc-500">
        Story not found
      </div>
    );
  }

  const statusBg =
    story.status === "in_progress"
      ? "bg-blue-700"
      : story.status === "completed"
        ? "bg-green-700"
        : "bg-zinc-700";
  const statusDot =
    story.status === "in_progress"
      ? "bg-blue-400"
      : story.status === "completed"
        ? "bg-green-400"
        : "bg-zinc-400";
  const statusLabel =
    story.status === "in_progress"
      ? "In Progress"
      : story.status === "completed"
        ? "Completed"
        : story.status === "failed"
          ? "Failed"
          : "Pending";

  return (
    <div className="flex h-full flex-col gap-6 p-8">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm font-[family-name:var(--font-sans)]">
        <Link
          href={`/epics/${epicId}`}
          className="text-[#8ea2bd] transition-colors hover:text-white"
        >
          Home
        </Link>
        <span className="text-[#525e6e]">&rsaquo;</span>
        <Link
          href={`/epics/${epicId}`}
          className="text-[#8ea2bd] transition-colors hover:text-white"
        >
          Stories
        </Link>
        <span className="text-[#525e6e]">&rsaquo;</span>
        <span className="font-medium text-[#f5f9ff]">{story.name}</span>
      </nav>

      {/* Story Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-[28px] font-bold tracking-tight text-[#f5f9ff] font-[family-name:var(--font-display)]">
            {story.name}
          </h1>
          <span
            className={`inline-flex items-center gap-1.5 rounded-xl px-2.5 py-1 ${statusBg}`}
          >
            <span className={`h-2 w-2 rounded-full ${statusDot}`} />
            <span className="text-[11px] font-medium text-blue-50 font-[family-name:var(--font-mono)]">
              {statusLabel}
            </span>
          </span>
        </div>
        <span className="text-[13px] text-[#8ea2bd] font-[family-name:var(--font-mono)]">
          {story.completed_task_count}/{story.task_count} tasks &middot; Last
          run: {story.last_activity}
        </span>
      </div>

      {/* Tab Bar */}
      <div className="flex gap-1 border-b border-[#22324a] pb-2">
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={cn(
              "rounded-lg px-4 py-2 text-sm font-medium transition-colors font-[family-name:var(--font-sans)]",
              activeTab === tab
                ? "bg-blue-600 text-white"
                : "text-[#8ea2bd] hover:bg-zinc-800 hover:text-zinc-200"
            )}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      {activeTab === "Tasks" && (
        <div className="flex flex-col gap-0.5">
          {tasks.map((task) => {
            const config = taskStatusConfig[task.status];
            return (
              <div
                key={task.id}
                className="flex items-center gap-3 rounded-lg bg-[#0f172a] px-4 py-3"
              >
                {/* Checkbox */}
                <div
                  className={cn(
                    "flex h-5 w-5 shrink-0 items-center justify-center rounded",
                    config.checkBg
                  )}
                >
                  <Check
                    className={cn("h-3.5 w-3.5", config.checkIcon)}
                  />
                </div>

                {/* Task Name */}
                <span className="flex-1 truncate text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
                  {task.name}
                </span>

                {/* Status Badge */}
                <span
                  className={cn(
                    "inline-flex items-center rounded px-2 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)]",
                    config.bg,
                    config.text
                  )}
                >
                  {config.label}
                </span>

                {/* Agent */}
                <div className="flex items-center gap-1">
                  <Bot className="h-3.5 w-3.5 text-[#8ea2bd]" />
                  <span className="text-xs text-zinc-500 font-[family-name:var(--font-mono)]">
                    {task.agent ?? "\u2014"}
                  </span>
                </div>

                {/* Duration */}
                <span className="w-16 text-right text-xs text-zinc-500 font-[family-name:var(--font-mono)]">
                  {task.duration ?? "\u2014"}
                </span>
              </div>
            );
          })}
        </div>
      )}

      {activeTab !== "Tasks" && (
        <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
          {activeTab} content coming soon
        </div>
      )}
    </div>
  );
}
