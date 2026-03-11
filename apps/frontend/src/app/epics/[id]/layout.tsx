"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { Sidebar } from "@/components/sidebar";
import { SidebarProvider } from "@/lib/sidebar-context";
import { WebSocketProvider, useWebSocket } from "@/lib/ws";
import type { WSEvent } from "@/lib/ws";
import { useServerUrl, useApi } from "@/lib/server-context";
import { deriveStoryStatus } from "@/lib/utils";
import type { Story } from "@/lib/types";

export default function EpicLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const serverUrl = useServerUrl();
  return (
    <SidebarProvider>
      <WebSocketProvider serverUrl={serverUrl}>
        <EpicLayoutInner>{children}</EpicLayoutInner>
      </WebSocketProvider>
    </SidebarProvider>
  );
}

function EpicLayoutInner({ children }: { children: React.ReactNode }) {
  const api = useApi();
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
    if (evt.event === "ws_reconnected") {
      window.location.reload();
      return;
    }
    if (evt.event === "approval_requested" || evt.event === "approval_decided") {
      api.listApprovals(undefined, "pending").then((approvals) => {
        setPendingCount(approvals.length);
      }).catch(() => {});
    }
  }, [api]);

  useWebSocket(handleWS);

  return (
    <div className="flex overflow-hidden bg-[#020617]" style={{ height: "calc(100vh - var(--titlebar-height))" }}>
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
