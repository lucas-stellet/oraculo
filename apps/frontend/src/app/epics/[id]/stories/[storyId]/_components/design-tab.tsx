"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { AlertTriangle, FileText } from "lucide-react";
import { api } from "@/lib/api";
import { approvalDisplayTitle } from "@/lib/utils";
import type { Approval } from "@/lib/types";
import { cn } from "@/lib/utils";
import { MarkdownRenderer } from "@/components/markdown-renderer";

export function DesignTab({ epicName }: { epicName: string }) {
  const [approvals, setApprovals] = useState<Approval[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.listApprovals()
      .then((all) => {
        const designApprovals = all.filter((a) => a.type === "design");
        setApprovals(designApprovals);
      })
      .catch(() => setApprovals([]))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
        Loading...
      </div>
    );
  }

  if (approvals.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
        No design versions yet
      </div>
    );
  }

  const current = approvals[0];
  const hasPendingApproval = current.status === "pending";

  const statusBadge = {
    pending: { label: "Pending", bg: "bg-amber-800", text: "text-amber-50" },
    approved: { label: "Approved", bg: "bg-green-700", text: "text-green-50" },
    rejected: { label: "Rejected", bg: "bg-red-700", text: "text-red-50" },
    none: { label: "No Review", bg: "bg-zinc-700", text: "text-zinc-100" },
    needs_revision: { label: "Needs Revision", bg: "bg-amber-800", text: "text-amber-50" },
  } as const;

  const badge = statusBadge[current.status];

  return (
    <div className="flex flex-col gap-6">
      {/* Approval gate banner */}
      {hasPendingApproval && (
        <div className="flex items-center gap-3 rounded-lg border border-amber-800/50 bg-amber-950/30 px-4 py-3">
          <AlertTriangle className="h-5 w-5 shrink-0 text-amber-400" />
          <span className="flex-1 text-sm text-amber-200 font-[family-name:var(--font-sans)]">
            Awaiting design approval
          </span>
          <Link
            href={`/epics/${epicName}/approvals/${current.id}/review`}
            className="text-sm font-medium text-amber-400 hover:text-amber-300 font-[family-name:var(--font-sans)]"
          >
            Go to Approvals &rarr;
          </Link>
        </div>
      )}

      {/* Version indicator */}
      <div className="flex items-center gap-3">
        <FileText className="h-4 w-4 text-[#8ea2bd]" />
        <span className="text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
          {approvalDisplayTitle(current.type)}
        </span>
        <span className="text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
          Requested{" "}
          {new Date(current.requested_at).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric",
          })}
        </span>
        <span
          className={cn(
            "inline-flex items-center rounded px-2 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)]",
            badge.bg,
            badge.text
          )}
        >
          {badge.label}
        </span>
      </div>

      {/* Content */}
      {current.content && <MarkdownRenderer content={current.content} />}
    </div>
  );
}
