"use client";

import { useState } from "react";
import Link from "next/link";
import { AlertTriangle, FileText, ChevronDown, ChevronRight } from "lucide-react";
import { getStoryVersions, getReviewsForVersion } from "@/lib/mock-data";
import type { ApprovalStatus } from "@/lib/types";
import { cn } from "@/lib/utils";

const approvalBadge: Record<ApprovalStatus, { label: string; bg: string; text: string }> = {
  none: { label: "No Review", bg: "bg-zinc-700", text: "text-zinc-100" },
  pending: { label: "Pending", bg: "bg-amber-800", text: "text-amber-50" },
  approved: { label: "Approved", bg: "bg-green-700", text: "text-green-50" },
  rejected: { label: "Rejected", bg: "bg-red-700", text: "text-red-50" },
  needs_revision: { label: "Needs Revision", bg: "bg-amber-800", text: "text-amber-50" },
};

function MarkdownPanel({ content }: { content: string }) {
  const sections = content.split(/^## /m).filter(Boolean);

  return (
    <div className="rounded-xl border border-[#22324a] bg-[#0f172a] p-6 space-y-5">
      {sections.map((section, i) => {
        const lines = section.split("\n");
        const title = lines[0].trim();
        const body = lines.slice(1).join("\n").trim();

        return (
          <div key={i}>
            <h2 className="mb-2 text-base font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
              {title}
            </h2>
            <div className="text-sm leading-relaxed text-[#8ea2bd] font-[family-name:var(--font-sans)] whitespace-pre-line">
              {body}
            </div>
          </div>
        );
      })}
    </div>
  );
}

export function DesignTab({ storyId, epicId }: { storyId: number; epicId: number }) {
  const versions = getStoryVersions(storyId, "design");
  const [showHistory, setShowHistory] = useState(false);

  if (versions.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
        No design versions yet
      </div>
    );
  }

  const current = versions[0];
  const badge = approvalBadge[current.approval_status];
  const hasPendingApproval = current.approval_status === "pending";

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
            href={`/epics/${epicId}/approvals`}
            className="text-sm font-medium text-amber-400 hover:text-amber-300 font-[family-name:var(--font-sans)]"
          >
            Go to Approvals &rarr;
          </Link>
        </div>
      )}

      {/* Version indicator */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <FileText className="h-4 w-4 text-[#8ea2bd]" />
          <span className="text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
            Version {current.version}
          </span>
          <span className="text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
            Created{" "}
            {new Date(current.created_at).toLocaleDateString("en-US", {
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
        {versions.length > 1 && (
          <button
            onClick={() => setShowHistory(!showHistory)}
            className="flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300 font-[family-name:var(--font-sans)]"
          >
            {showHistory ? (
              <>
                <ChevronDown className="h-3 w-3" />
                Hide previous versions
              </>
            ) : (
              <>
                <ChevronRight className="h-3 w-3" />
                View previous versions ({versions.length - 1})
              </>
            )}
          </button>
        )}
      </div>

      {/* Current version content */}
      <MarkdownPanel content={current.content} />

      {/* Previous versions */}
      {showHistory && versions.slice(1).map((v) => {
        const vBadge = approvalBadge[v.approval_status];
        return (
          <div key={v.id} className="opacity-60">
            <div className="flex items-center gap-3 mb-3">
              <span className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                Version {v.version}
              </span>
              <span className="text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
                {new Date(v.created_at).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                })}
              </span>
              <span
                className={cn(
                  "inline-flex items-center rounded px-2 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)]",
                  vBadge.bg,
                  vBadge.text
                )}
              >
                {vBadge.label}
              </span>
            </div>
            <MarkdownPanel content={v.content} />
          </div>
        );
      })}
    </div>
  );
}
