"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  FolderOpen,
  ChevronDown,
  House,
  Layers,
  PanelLeftClose,
  PanelLeftOpen,
  ScanEye,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface SidebarProps {
  epicId: number;
  epicName: string;
}

const navItems = [
  { label: "Home", icon: House, href: (id: number) => `/epics/${id}` },
  {
    label: "Stories",
    icon: Layers,
    href: (id: number) => `/epics/${id}`,
    matchPattern: "/stories/",
  },
];

export function Sidebar({ epicId, epicName }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false);
  const pathname = usePathname();

  const isStoryDetail = pathname.includes("/stories/");

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

      {/* Nav Items */}
      <nav className="space-y-0.5 px-2 py-3">
        {navItems.map((item) => {
          const isActive = item.matchPattern
            ? isStoryDetail
            : !isStoryDetail;

          return (
            <Link
              key={item.label}
              href={item.href(epicId)}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                collapsed && "justify-center px-0",
                isActive
                  ? "bg-blue-600 text-white"
                  : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
              )}
              title={collapsed ? item.label : undefined}
            >
              <item.icon className="h-[18px] w-[18px] shrink-0" />
              {!collapsed && (
                <span className="font-[family-name:var(--font-sans)]">
                  {item.label}
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      {/* Spacer */}
      <div className="flex-1" />

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
