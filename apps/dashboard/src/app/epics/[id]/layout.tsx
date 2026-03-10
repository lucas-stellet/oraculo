"use client";

import { use } from "react";
import { Sidebar } from "@/components/sidebar";
import { getEpic, getStories, getPendingApprovals } from "@/lib/mock-data";

export default function EpicLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const epicId = Number(id);
  const epic = getEpic();
  const stories = getStories(epicId);
  const pendingApprovals = getPendingApprovals(epicId);

  return (
    <div className="flex h-screen overflow-hidden bg-[#020617]">
      <Sidebar
        epicId={epicId}
        epicName={epic.name}
        stories={stories.map((s) => ({ id: s.id, name: s.name, status: s.status }))}
        pendingApprovalCount={pendingApprovals.length}
      />
      <div className="flex-1 overflow-y-auto">{children}</div>
    </div>
  );
}
