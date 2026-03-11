"use client";

import { useState, useRef, useEffect } from "react";
import { FileText, ChevronRight } from "lucide-react";
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
  const [selectedVersionIndex, setSelectedVersionIndex] = useState(0);
  const [showPopup, setShowPopup] = useState(false);
  const popupRef = useRef<HTMLDivElement>(null);

  // Sort versions by number descending (latest first)
  const sorted = [...versions].sort((a, b) => b.number - a.number);

  // Close popup on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (popupRef.current && !popupRef.current.contains(e.target as Node)) {
        setShowPopup(false);
      }
    }
    if (showPopup) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [showPopup]);

  if (sorted.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
        No requirements versions yet
      </div>
    );
  }

  const displayed = sorted[selectedVersionIndex];
  const displayedStatus = deriveVersionApprovalStatus(displayed.id, reviews);
  const badge = approvalBadge[displayedStatus];
  const isViewingOldVersion = selectedVersionIndex > 0;

  return (
    <div className="flex flex-col gap-6">
      {/* Version indicator */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <FileText className="h-4 w-4 text-[#8ea2bd]" />
          <span className="text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
            Version {displayed.number}
          </span>
          <span className="text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
            Created{" "}
            {new Date(displayed.created_at).toLocaleDateString("en-US", {
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
          <div className="flex items-center gap-3">
            {isViewingOldVersion && (
              <button
                onClick={() => setSelectedVersionIndex(0)}
                className="text-xs font-medium text-amber-400 hover:text-amber-300 font-[family-name:var(--font-sans)]"
              >
                Back to latest (v{sorted[0].number})
              </button>
            )}
            <div className="relative" ref={popupRef}>
              <button
                onClick={() => setShowPopup(!showPopup)}
                className="flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300 font-[family-name:var(--font-sans)]"
              >
                <ChevronRight className={cn("h-3 w-3 transition-transform", showPopup && "rotate-90")} />
                View previous versions ({sorted.length - 1})
              </button>
              {showPopup && (
                <div className="absolute right-0 top-full mt-1 z-50 min-w-[200px] rounded-lg border border-[#22324a] bg-[#0d1525] shadow-xl">
                  <div className="p-2">
                    <p className="px-2 py-1 text-[11px] font-medium text-[#525e6e] uppercase tracking-wider font-[family-name:var(--font-mono)]">
                      Versions
                    </p>
                    {sorted.map((v, idx) => {
                      const vStatus = deriveVersionApprovalStatus(v.id, reviews);
                      const vBadge = approvalBadge[vStatus];
                      const isSelected = idx === selectedVersionIndex;
                      return (
                        <button
                          key={v.id}
                          onClick={() => {
                            setSelectedVersionIndex(idx);
                            setShowPopup(false);
                          }}
                          className={cn(
                            "flex w-full items-center justify-between gap-3 rounded-md px-2 py-1.5 text-left text-xs transition-colors",
                            isSelected
                              ? "bg-blue-500/10 text-blue-400"
                              : "text-[#8ea2bd] hover:bg-[#1a2435]"
                          )}
                        >
                          <span className="font-[family-name:var(--font-sans)]">
                            Version {v.number}
                            {idx === 0 && (
                              <span className="ml-1.5 text-[10px] text-[#525e6e]">(latest)</span>
                            )}
                          </span>
                          <span
                            className={cn(
                              "inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium font-[family-name:var(--font-mono)]",
                              vBadge.bg,
                              vBadge.text
                            )}
                          >
                            {vBadge.label}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Displayed version content */}
      {isViewingOldVersion && (
        <div className="rounded-md border border-amber-500/20 bg-amber-500/5 px-3 py-2 text-xs text-amber-400/80 font-[family-name:var(--font-sans)]">
          You are viewing version {displayed.number}. This is not the latest version.
        </div>
      )}
      <MarkdownRenderer content={displayed.content} />

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
