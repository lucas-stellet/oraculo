"use client";

import { use, useState } from "react";
import Link from "next/link";
import { getEpic, getStoryDetail, getTasks } from "@/lib/mock-data";
import type { ApprovalStatus } from "@/lib/types";
import { cn } from "@/lib/utils";
import { TasksTab } from "./_components/tasks-tab";
import { RequirementsTab } from "./_components/requirements-tab";
import { DesignTab } from "./_components/design-tab";
import { QATab } from "./_components/qa-tab";

const tabs = ["Tasks", "Requirements", "Design", "QA"] as const;
type Tab = (typeof tabs)[number];

const approvalConfig: Record<
  ApprovalStatus,
  { label: string; bg: string; text: string }
> = {
  none: { label: "No Review", bg: "bg-zinc-700", text: "text-zinc-100" },
  pending: { label: "Pending Approval", bg: "bg-amber-800", text: "text-amber-50" },
  approved: { label: "Approved", bg: "bg-green-700", text: "text-green-50" },
  rejected: { label: "Rejected", bg: "bg-red-700", text: "text-red-50" },
  needs_revision: { label: "Needs Revision", bg: "bg-amber-800", text: "text-amber-50" },
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
  const story = getStoryDetail(sId);
  const tasks = getTasks(sId);

  const [activeTab, setActiveTab] = useState<Tab>("Tasks");

  if (!story) {
    return (
      <div className="flex h-full items-center justify-center text-zinc-500">
        Story not found
      </div>
    );
  }

  const approval = approvalConfig[story.approval_status];

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
            className={cn(
              "inline-flex items-center rounded-xl px-2.5 py-1 text-[11px] font-medium font-[family-name:var(--font-mono)]",
              approval.bg,
              approval.text
            )}
          >
            {approval.label}
          </span>
        </div>
        <span className="text-[13px] text-[#8ea2bd] font-[family-name:var(--font-mono)]">
          {story.completed_task_count}/{story.task_count} tasks completed
          &middot; Last activity: {story.last_activity}
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
      {activeTab === "Tasks" && <TasksTab tasks={tasks} />}
      {activeTab === "Requirements" && <RequirementsTab storyId={sId} />}
      {activeTab === "Design" && <DesignTab storyId={sId} epicId={epicId} />}
      {activeTab === "QA" && <QATab storyId={sId} />}
    </div>
  );
}
