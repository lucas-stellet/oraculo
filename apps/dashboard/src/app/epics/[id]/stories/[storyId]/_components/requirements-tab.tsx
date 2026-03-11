"use client";

import { useState } from "react";
import { FileText, ChevronDown, ChevronRight } from "lucide-react";
import type { StoryVersion, Review, ApprovalStatus } from "@/lib/types";
import { cn } from "@/lib/utils";
import { MarkdownRenderer } from "@/components/markdown-renderer";

const approvalBadge: Record<ApprovalStatus, { label: string; bg: string; text: string }> = {
  none: { label: "No Review", bg: "bg-zinc-700", text: "text-zinc-100" },
  pending: { label: "Pending", bg: "bg-amber-800", text: "text-amber-50" },
  approved: { label: "Approved", bg: "bg-green-700", text: "text-green-50" },
  rejected: { label: "Rejected", bg: "bg-red-700", text: "text-red-50" },
  needs_revision: { label: "Needs Revision", bg: "bg-amber-800", text: "text-amber-50" },
};

function deriveVersionApprovalStatus(versionId: number, reviews: Review[]): ApprovalStatus {
  const versionReviews = reviews.filter((r) => r.version_id === versionId);
  if (versionReviews.length === 0) return "none";
  const latest = versionReviews[versionReviews.length - 1];
  return latest.verdict === "approved" ? "approved" : "rejected";
}

function ReviewCard({ review }: { review: Review }) {
  const badge = review.verdict === "approved" ? approvalBadge.approved : approvalBadge.rejected;

  return (
    <div className="rounded-lg border border-[#22324a] bg-[#0b1120] p-4">
      <div className="flex items-center gap-2 mb-2">
        <span
          className={cn(
            "inline-flex items-center rounded px-2 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)]",
            badge.bg,
            badge.text
          )}
        >
          {badge.label}
        </span>
        <span className="text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
          {new Date(review.created_at).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
          })}
        </span>
      </div>
      <p className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
        {review.comment}
      </p>
    </div>
  );
}

interface RequirementsTabProps {
  versions: StoryVersion[];
  reviews: Review[];
}

export function RequirementsTab({ versions, reviews }: RequirementsTabProps) {
  const [showHistory, setShowHistory] = useState(false);

  // Sort versions by number descending (latest first)
  const sorted = [...versions].sort((a, b) => b.number - a.number);

  if (sorted.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
        No requirements versions yet
      </div>
    );
  }

  const current = sorted[0];
  const currentStatus = deriveVersionApprovalStatus(current.id, reviews);
  const badge = approvalBadge[currentStatus];

  return (
    <div className="flex flex-col gap-6">
      {/* Version indicator */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <FileText className="h-4 w-4 text-[#8ea2bd]" />
          <span className="text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
            Version {current.number}
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
        {sorted.length > 1 && (
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
                View previous versions ({sorted.length - 1})
              </>
            )}
          </button>
        )}
      </div>

      {/* Current version content */}
      <MarkdownRenderer content={current.content} />

      {/* Previous versions */}
      {showHistory && sorted.slice(1).map((v) => {
        const vStatus = deriveVersionApprovalStatus(v.id, reviews);
        const vBadge = approvalBadge[vStatus];
        return (
          <div key={v.id} className="opacity-60">
            <div className="flex items-center gap-3 mb-3">
              <span className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                Version {v.number}
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
            <MarkdownRenderer content={v.content} />
          </div>
        );
      })}

      {/* Review History */}
      {reviews.length > 0 && (
        <div>
          <h3 className="mb-3 text-sm font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
            Review History
          </h3>
          <div className="flex flex-col gap-2">
            {reviews.map((review) => (
              <ReviewCard key={review.id} review={review} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
