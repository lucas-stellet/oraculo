"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BookOpen,
  FolderOpen,
  ChevronDown,
  LayoutDashboard,
  PanelLeftClose,
  PanelLeftOpen,
  ScanEye,
  ShieldAlert,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useSidebar } from "@/lib/sidebar-context";

interface SidebarProps {
  epicId: number;
  epicName: string;
  stories: { id: number; name: string; status: string }[];
  pendingApprovalCount: number;
}

function storyDotColor(status: string) {
  switch (status) {
    case "completed":
      return "bg-green-500";
    case "in_progress":
      return "bg-blue-500";
    case "failed":
      return "bg-red-500";
    default:
      return "bg-zinc-500";
  }
}

export function Sidebar({
  epicId,
  epicName,
  stories,
  pendingApprovalCount,
}: SidebarProps) {
  const { collapsed, setCollapsed } = useSidebar();
  const pathname = usePathname();

  const homeHref = `/epics/${epicId}`;
  const isHomeActive = pathname === `/epics/${epicId}`;
  const isApprovalsActive = pathname.includes("/approvals");

  return (
    <aside
      className={cn(
        "flex h-screen flex-col bg-zinc-900 transition-[width] duration-200",
        collapsed ? "w-16" : "w-[260px]"
      )}
    >
      {/* Logo */}
      <div
        className={cn(
          "flex h-12 shrink-0 items-center gap-2.5 border-b border-white/5",
          collapsed ? "justify-center px-0" : "px-4"
        )}
      >
        <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-blue-600">
          <ScanEye className="h-4 w-4 text-white" />
        </div>
        {!collapsed && (
          <span className="text-[15px] font-bold tracking-[2.5px] text-[#f5f9ff] font-[family-name:var(--font-display)]">
            ORACULO
          </span>
        )}
      </div>

      {/* Context Switcher */}
      <div
        className={cn(
          "flex h-16 items-center gap-3 border-b border-white/10",
          collapsed ? "justify-center px-0" : "px-4"
        )}
      >
        <FolderOpen className="h-5 w-5 shrink-0 text-blue-400" />
        {!collapsed && (
          <>
            <span className="truncate text-sm font-semibold font-[family-name:var(--font-display)] text-[#f5f9ff]">
              {epicName}
            </span>
            <ChevronDown className="ml-auto h-4 w-4 shrink-0 text-[#8ea2bd]" />
          </>
        )}
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto">
        {/* Overview */}
        <div className={cn("px-2 py-3", collapsed && "px-2 py-3")}>
          <Link
            href={homeHref}
            className={cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              collapsed && "justify-center px-0",
              isHomeActive
                ? "bg-blue-600 text-white"
                : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
            )}
            title={collapsed ? "Overview" : undefined}
          >
            <LayoutDashboard className="h-[18px] w-[18px] shrink-0" />
            {!collapsed && (
              <span className="font-[family-name:var(--font-sans)]">
                Overview
              </span>
            )}
          </Link>
        </div>

        {/* STORIES section */}
        <div
          className={cn(
            collapsed && "border-t border-white/5 px-2 py-2",
            !collapsed && "px-2"
          )}
        >
          {!collapsed && (
            <div className="px-3 pt-4 pb-1 text-[11px] uppercase tracking-widest text-zinc-500 font-[family-name:var(--font-mono)]">
              Stories
            </div>
          )}

          {collapsed ? (
            <div
              className="flex items-center justify-center rounded-lg py-2.5 text-[#8ea2bd] hover:bg-zinc-800 transition-colors cursor-default"
              title="Stories"
            >
              <div className="flex items-start">
                <BookOpen className="h-[18px] w-[18px]" />
                {stories.length > 0 && (
                  <span className="-ml-1.5 -mt-2 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-blue-500 px-1 text-[9px] font-semibold text-white font-[family-name:var(--font-mono)]">
                    {stories.length}
                  </span>
                )}
              </div>
            </div>
          ) : (
            <div className="space-y-0.5">
              {stories.map((story) => {
                const storyHref = `/epics/${epicId}/stories/${story.id}`;
                const isActive = pathname.includes(`/stories/${story.id}`);

                return (
                  <Link
                    key={story.id}
                    href={storyHref}
                    className={cn(
                      "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                      isActive
                        ? "bg-blue-600/20 text-blue-400"
                        : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
                    )}
                  >
                    <div
                      className={cn(
                        "h-1.5 w-1.5 shrink-0 rounded-full",
                        storyDotColor(story.status)
                      )}
                    />
                    <span className="truncate text-[13px] font-[family-name:var(--font-sans)]">
                      {story.name}
                    </span>
                  </Link>
                );
              })}
            </div>
          )}
        </div>

        {/* MONITOR section */}
        <div
          className={cn(
            collapsed && "border-t border-white/5 px-2 py-2",
            !collapsed && "px-2"
          )}
        >
          {!collapsed && (
            <div className="px-3 pt-5 pb-1 text-[11px] uppercase tracking-widest text-zinc-500 font-[family-name:var(--font-mono)]">
              Monitor
            </div>
          )}

          <Link
            href={`/epics/${epicId}/approvals`}
            className={cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              collapsed && "justify-center px-0",
              isApprovalsActive
                ? "bg-zinc-800 text-white"
                : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
            )}
            title={collapsed ? "Approvals" : undefined}
          >
            {collapsed ? (
              <div className="flex items-start">
                <ShieldAlert className="h-[18px] w-[18px] shrink-0" />
                {pendingApprovalCount > 0 && (
                  <span className="-ml-1.5 -mt-2 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-red-600 px-1 text-[9px] font-semibold text-white font-[family-name:var(--font-mono)]">
                    {pendingApprovalCount}
                  </span>
                )}
              </div>
            ) : (
              <>
                <ShieldAlert className="h-[18px] w-[18px] shrink-0" />
                <span className="font-[family-name:var(--font-sans)]">
                  Approvals
                </span>
                {pendingApprovalCount > 0 && (
                  <span className="ml-auto flex items-center justify-center rounded-full bg-red-600 text-[10px] text-white min-w-[18px] h-[18px] px-1 font-[family-name:var(--font-mono)]">
                    {pendingApprovalCount}
                  </span>
                )}
              </>
            )}
          </Link>
        </div>
      </nav>

      {/* Sidebar Toggle */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className={cn(
          "flex h-12 items-center gap-2 border-t border-white/10 transition-colors hover:bg-zinc-800",
          collapsed ? "justify-center px-0" : "px-4"
        )}
      >
        {collapsed ? (
          <PanelLeftOpen className="h-[18px] w-[18px] text-[#8ea2bd]" />
        ) : (
          <>
            <PanelLeftClose className="h-[18px] w-[18px] text-[#8ea2bd]" />
            <span className="text-[13px] font-medium text-zinc-500 font-[family-name:var(--font-sans)]">
              Collapse
            </span>
          </>
        )}
      </button>
    </aside>
  );
}
