"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { Sidebar } from "@/components/sidebar";
import { SidebarProvider } from "@/lib/sidebar-context";
import { WebSocketProvider, useWebSocket } from "@/lib/ws";
import type { WSEvent } from "@/lib/ws";
import { api } from "@/lib/api";
import { deriveStoryStatus } from "@/lib/utils";
import type { Story } from "@/lib/types";

export default function EpicLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <SidebarProvider>
      <WebSocketProvider>
        <EpicLayoutInner>{children}</EpicLayoutInner>
      </WebSocketProvider>
    </SidebarProvider>
  );
}

function EpicLayoutInner({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const epicName = pathname.split("/").filter(Boolean)[1];
  const [stories, setStories] = useState<Story[]>([]);
  const [pendingCount, setPendingCount] = useState(0);

  useEffect(() => {
    api.listStories(epicName).then(setStories).catch(() => setStories([]));
    api.listApprovals(undefined, "pending").then((approvals) => {
      setPendingCount(approvals.length);
    }).catch(() => setPendingCount(0));
  }, [epicName]);

  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event === "approval_requested" || evt.event === "approval_decided") {
      api.listApprovals(undefined, "pending").then((approvals) => {
        setPendingCount(approvals.length);
      }).catch(() => {});
    }
  }, []);

  useWebSocket(handleWS);

  return (
    <div className="flex h-screen overflow-hidden bg-[#020617]">
      <Sidebar
        epicId={epicName}
        epicName={epicName}
        stories={stories.map((s) => ({ id: s.id, name: s.name, status: deriveStoryStatus(s) }))}
        pendingApprovalCount={pendingCount}
      />
      <div className="flex-1 overflow-y-auto">{children}</div>
    </div>
  );
}
