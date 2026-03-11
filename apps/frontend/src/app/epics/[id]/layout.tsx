"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { RefreshCw } from "lucide-react";
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
  const [updateAvailable, setUpdateAvailable] = useState(false);
  const [newVersion, setNewVersion] = useState("");
  const [restarting, setRestarting] = useState(false);
  const [version, setVersion] = useState("");

  useEffect(() => {
    api.listStories(epicName).then(setStories).catch(() => setStories([]));
    api.listApprovals(undefined, "pending").then((approvals) => {
      setPendingCount(approvals.length);
    }).catch(() => setPendingCount(0));
  }, [epicName]);

  // Poll for new binary every 30s
  useEffect(() => {
    function check() {
      api.getSystemStatus()
        .then((s) => {
          setUpdateAvailable(s.update_available);
          if (s.version) setVersion(s.version);
          if (s.new_version) setNewVersion(s.new_version);
        })
        .catch(() => {});
    }
    check();
    const id = setInterval(check, 30_000);
    return () => clearInterval(id);
  }, []);

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
  }, []);

  useWebSocket(handleWS);

  function handleRestart() {
    if (restarting) return;
    setRestarting(true);
    // Fire and forget — the WS drop+reconnect will trigger window.location.reload()
    api.restartServer().catch(() => {});
  }

  return (
    <div className="flex flex-col h-screen overflow-hidden bg-[#020617]">
      {updateAvailable && (
        <div className="flex shrink-0 items-center justify-between gap-3 border-b border-blue-800/50 bg-blue-950/40 px-5 py-2.5">
          <span className="text-sm text-blue-200 font-[family-name:var(--font-sans)]">
            New version available
            {newVersion && (
              <span className="ml-2 font-[family-name:var(--font-mono)] text-blue-400 text-xs">
                {newVersion}
              </span>
            )}
          </span>
          <button
            onClick={handleRestart}
            disabled={restarting}
            className="flex items-center gap-1.5 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-[family-name:var(--font-sans)]"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${restarting ? "animate-spin" : ""}`} />
            {restarting ? "Restarting…" : "Restart server"}
          </button>
        </div>
      )}
      <div className="flex flex-1 overflow-hidden">
        <Sidebar
          epicId={epicName}
          epicName={epicName}
          stories={stories.map((s) => ({ id: s.id, name: s.name, status: deriveStoryStatus(s) }))}
          pendingApprovalCount={pendingCount}
          projectCommit={version}
        />
        <div className="flex-1 overflow-y-auto">{children}</div>
      </div>
    </div>
  );
}
