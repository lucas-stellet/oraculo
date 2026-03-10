"use client";

import { useState } from "react";
import { AlertOctagon, ChevronDown, ChevronRight } from "lucide-react";
import { getValidations } from "@/lib/mock-data";
import { cn } from "@/lib/utils";

const ESCALATION_THRESHOLD = 2;

export function QATab({ storyId }: { storyId: number }) {
  const validations = getValidations(storyId);
  const [expandedId, setExpandedId] = useState<number | null>(null);

  if (validations.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-zinc-600 font-[family-name:var(--font-sans)]">
        No validations yet
      </div>
    );
  }

  const totalValidations = validations.length;
  const approvedCount = validations.filter((v) => v.verdict === "approved").length;
  const rejectedCount = validations.filter((v) => v.verdict === "rejected").length;
  const approvalRate = Math.round((approvedCount / totalValidations) * 100);
  const needsEscalation = rejectedCount > ESCALATION_THRESHOLD;

  return (
    <div className="flex flex-col gap-6">
      {/* Metric Cards */}
      <div className="grid grid-cols-3 gap-4">
        <div className="rounded-xl border border-[#22324a] bg-[#0f172a] p-4">
          <p className="text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)] mb-1">
            Total Validations
          </p>
          <p className="text-2xl font-bold text-[#f5f9ff] font-[family-name:var(--font-display)]">
            {totalValidations}
          </p>
        </div>
        <div className="rounded-xl border border-[#22324a] bg-[#0f172a] p-4">
          <p className="text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)] mb-1">
            Approval Rate
          </p>
          <p
            className={cn(
              "text-2xl font-bold font-[family-name:var(--font-display)]",
              approvalRate >= 60 ? "text-green-400" : approvalRate < 40 ? "text-red-400" : "text-amber-400"
            )}
          >
            {approvalRate}%
          </p>
        </div>
        <div className="rounded-xl border border-[#22324a] bg-[#0f172a] p-4">
          <p className="text-xs text-[#8ea2bd] font-[family-name:var(--font-sans)] mb-1">
            Rejection Cycles
          </p>
          <div className="flex items-baseline gap-2">
            <p
              className={cn(
                "text-2xl font-bold font-[family-name:var(--font-display)]",
                rejectedCount > ESCALATION_THRESHOLD ? "text-red-400" : rejectedCount > 1 ? "text-amber-400" : "text-[#f5f9ff]"
              )}
            >
              {rejectedCount}
            </p>
            {rejectedCount > ESCALATION_THRESHOLD && (
              <span className="text-xs text-amber-400 font-[family-name:var(--font-mono)]">
                above threshold
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Escalation Banner */}
      {needsEscalation && (
        <div className="flex items-center gap-3 rounded-lg border border-red-800/50 bg-red-950/30 px-4 py-3">
          <AlertOctagon className="h-5 w-5 shrink-0 text-red-400" />
          <span className="text-sm text-red-200 font-[family-name:var(--font-sans)]">
            QA Escalation — requires human intervention
          </span>
        </div>
      )}

      {/* Validation History */}
      <div>
        <h3 className="mb-3 text-sm font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
          Validation History
        </h3>
        <div className="flex flex-col gap-2">
          {validations.map((v) => {
            const isExpanded = expandedId === v.id;
            const isRejected = v.verdict === "rejected";
            const verdictBg = isRejected ? "bg-red-700" : "bg-green-700";
            const verdictText = isRejected ? "text-red-50" : "text-green-50";
            const verdictLabel = isRejected ? "Rejected" : "Approved";

            return (
              <div
                key={v.id}
                className="rounded-lg border border-[#22324a] bg-[#0b1120]"
              >
                <button
                  onClick={() => setExpandedId(isExpanded ? null : v.id)}
                  className="flex w-full items-center gap-3 px-4 py-3 text-left"
                >
                  <span
                    className={cn(
                      "inline-flex shrink-0 items-center rounded px-2 py-0.5 text-[11px] font-medium font-[family-name:var(--font-mono)]",
                      verdictBg,
                      verdictText
                    )}
                  >
                    {verdictLabel}
                  </span>
                  <span className="flex-1 truncate text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                    {v.findings.slice(0, 80)}{v.findings.length > 80 ? "..." : ""}
                  </span>
                  <span className="shrink-0 text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
                    {new Date(v.created_at).toLocaleDateString("en-US", {
                      month: "short",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </span>
                  {isExpanded
                    ? <ChevronDown className="h-4 w-4 shrink-0 text-[#8ea2bd]" />
                    : <ChevronRight className="h-4 w-4 shrink-0 text-[#8ea2bd]" />}
                </button>
                {isExpanded && (
                  <div className="border-t border-[#22324a] px-4 py-3">
                    <p className="text-sm leading-relaxed text-[#f5f9ff] font-[family-name:var(--font-sans)]">
                      {v.findings}
                    </p>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
