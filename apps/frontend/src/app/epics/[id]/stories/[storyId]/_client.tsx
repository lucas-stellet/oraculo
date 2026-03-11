"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { deriveStoryStatus } from "@/lib/utils";
import type { Story, StoryTask, StoryVersion, Review, Validation, ApprovalStatus } from "@/lib/types";
import { cn } from "@/lib/utils";
import { useWebSocket } from "@/lib/ws";
import type { WSEvent } from "@/lib/ws";
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

export default function StoryDetailPage() {
  const pathname = usePathname();
  const segs = pathname.split("/").filter(Boolean);
  // /epics/{id}/stories/{storyId}
  const epicName = segs[1];
  const storyNameEncoded = segs[3];
  const storyName = decodeURIComponent(storyNameEncoded);

  const [story, setStory] = useState<Story | null>(null);
  const [tasks, setTasks] = useState<StoryTask[]>([]);
  const [versions, setVersions] = useState<StoryVersion[]>([]);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [validations, setValidations] = useState<Validation[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<Tab>("Tasks");
  const [activeAgents, setActiveAgents] = useState<Record<number, string>>({});

  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event === "task_started" || evt.event === "task_completed") {
      api.listTasks(epicName, storyName).then(setTasks).catch(() => {});
      api.listStories(epicName).then((stories) => {
        const found = stories.find((s) => s.name === storyName);
        if (found) setStory(found);
      }).catch(() => {});
    }
    if (evt.event === "agent_started") {
      const data = evt.data as { task_id?: number; name?: string } | undefined;
      if (data?.task_id && data?.name) {
        setActiveAgents((prev) => ({ ...prev, [data.task_id!]: data.name! }));
      }
    }
    if (evt.event === "agent_stopped") {
      const data = evt.data as { task_id?: number } | undefined;
      if (data?.task_id) {
        setActiveAgents((prev) => {
          const next = { ...prev };
          delete next[data.task_id!];
          return next;
        });
      }
    }
  }, [epicName, storyName]);

  useWebSocket(handleWS);

  useEffect(() => {
    Promise.all([
      api.listStories(epicName).then((stories) => {
        const found = stories.find((s) => s.name === storyName);
        setStory(found ?? null);
      }),
      api.listTasks(epicName, storyName).then(setTasks).catch(() => setTasks([])),
      api.listStoryVersions(epicName, storyName).then(setVersions).catch(() => setVersions([])),
      api.listStoryReviews(epicName, storyName).then(setReviews).catch(() => setReviews([])),
      api.listValidations(epicName, storyName).then(setValidations).catch(() => setValidations([])),
    ])
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [epicName, storyName]);

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center text-zinc-500">
        Loading...
      </div>
    );
  }

  if (!story) {
    return (
      <div className="flex h-full items-center justify-center text-zinc-500">
        Story not found
      </div>
    );
  }

  const approval = approvalConfig[story.approval_status];
  const lastActivity = story.updated_at
    ? new Date(story.updated_at).toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    : "\u2014";

  return (
    <div className="flex h-full flex-col gap-6 p-8">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm font-[family-name:var(--font-sans)]">
        <Link
          href={`/epics/${epicName}`}
          className="text-[#8ea2bd] transition-colors hover:text-white"
        >
          Home
        </Link>
        <span className="text-[#525e6e]">&rsaquo;</span>
        <Link
          href={`/epics/${epicName}`}
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
          &middot; Last activity: {lastActivity}
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
      {activeTab === "Tasks" && <TasksTab tasks={tasks} activeAgents={activeAgents} />}
      {activeTab === "Requirements" && <RequirementsTab versions={versions} reviews={reviews} />}
      {activeTab === "Design" && <DesignTab epicName={epicName} />}
      {activeTab === "QA" && <QATab validations={validations} />}
    </div>
  );
}
