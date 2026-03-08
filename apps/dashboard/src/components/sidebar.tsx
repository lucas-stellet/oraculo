"use client";

import {
  FolderTree,
  GitBranch,
  Bot,
  Activity,
  ShieldAlert,
  ShieldCheck,
  Brain,
  Layers,
  Settings,
} from "lucide-react";

const navItems = [
  { label: "Stories", icon: FolderTree },
  { label: "DAG View", icon: GitBranch },
  { label: "Agent Monitor", icon: Bot },
  { label: "Activity Feed", icon: Activity },
  { label: "Approvals", icon: ShieldAlert },
  { label: "QA Dashboard", icon: ShieldCheck },
  { label: "Knowledge Base", icon: Brain },
  { label: "Sessions", icon: Layers },
];

export function Sidebar() {
  return (
    <aside className="flex h-screen w-64 flex-col border-r bg-card">
      <div className="border-b px-6 py-5">
        <h1 className="text-xl font-bold tracking-tight font-[family-name:var(--font-display)]">
          Oraculo
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          Select Epic...
        </p>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {navItems.map((item) => (
          <button
            key={item.label}
            disabled
            className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground/50 cursor-not-allowed"
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </button>
        ))}
      </nav>

      <div className="border-t px-3 py-4">
        <button
          disabled
          className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground/50 cursor-not-allowed"
        >
          <Settings className="h-4 w-4" />
          Settings
        </button>
      </div>
    </aside>
  );
}
