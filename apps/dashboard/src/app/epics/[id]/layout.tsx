"use client";

import { use, useEffect, useState } from "react";
import { Sidebar } from "@/components/sidebar";
import { SidebarProvider } from "@/lib/sidebar-context";
import { api } from "@/lib/api";
import { deriveStoryStatus } from "@/lib/utils";
import type { Story } from "@/lib/types";

export default function EpicLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const { id: epicName } = use(params);
  const [stories, setStories] = useState<Story[]>([]);
  const [pendingCount, setPendingCount] = useState(0);

  useEffect(() => {
    api.listStories(epicName).then(setStories).catch(() => setStories([]));
    api.listApprovals(undefined, "pending").then((approvals) => {
      setPendingCount(approvals.length);
    }).catch(() => setPendingCount(0));
  }, [epicName]);

  return (
    <SidebarProvider>
      <div className="flex h-screen overflow-hidden bg-[#020617]">
        <Sidebar
          epicId={epicName}
          epicName={epicName}
          stories={stories.map((s) => ({ id: s.id, name: s.name, status: deriveStoryStatus(s) }))}
          pendingApprovalCount={pendingCount}
        />
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>
    </SidebarProvider>
  );
}
