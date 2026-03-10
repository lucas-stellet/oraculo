"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useWebSocket } from "@/lib/ws";
import type { WSEvent } from "@/lib/ws";
import Link from "next/link";
import {
  Compass,
  AlertTriangle,
  ClipboardList,
  CheckCircle2,
  XCircle,
  ArrowRight,
  ShieldAlert,
} from "lucide-react";
import { api } from "@/lib/api";
import { approvalDisplayTitle } from "@/lib/utils";
import type { Approval, ApprovalType } from "@/lib/types";

const typeConfig: Record<
  ApprovalType,
  {
    label: string;
    icon: React.ElementType;
    iconColor: string;
    badgeBg: string;
    badgeText: string;
    resolvedBadgeBg: string;
    resolvedBadgeText: string;
  }
> = {
  design: {
    label: "Design",
    icon: Compass,
    iconColor: "#a78bfa",
    badgeBg: "#7c3aed",
    badgeText: "#f5f3ff",
    resolvedBadgeBg: "#2e1065",
    resolvedBadgeText: "#c4b5fd",
  },
  "qa-escalation": {
    label: "QA Escalation",
    icon: AlertTriangle,
    iconColor: "#fca5a5",
    badgeBg: "#dc2626",
    badgeText: "#fef2f2",
    resolvedBadgeBg: "#450a0a",
    resolvedBadgeText: "#fca5a5",
  },
  "execution-plan": {
    label: "Execution Plan",
    icon: ClipboardList,
    iconColor: "#fbbf24",
    badgeBg: "#059669",
    badgeText: "#ecfdf5",
    resolvedBadgeBg: "#422006",
    resolvedBadgeText: "#fbbf24",
  },
};

export default function ApprovalsPage() {
  const pathname = usePathname();
  const epicName = pathname.split("/").filter(Boolean)[1];
  const [pending, setPending] = useState<Approval[]>([]);
  const [resolved, setResolved] = useState<Approval[]>([]);

  useEffect(() => {
    api.listApprovals().then((all) => {
      setPending(all.filter((a) => a.status === "pending"));
      setResolved(all.filter((a) => a.status !== "pending"));
    }).catch(() => {
      setPending([]);
      setResolved([]);
    });
  }, []);

  const handleWS = useCallback((evt: WSEvent) => {
    if (evt.event !== "approval_requested" && evt.event !== "approval_decided") return;
    if (!evt.id) return;
    api.getApproval(evt.id).then((approval) => {
      if (approval.status === "pending") {
        setPending((prev) => {
          const exists = prev.some((a) => a.id === approval.id);
          return exists
            ? prev.map((a) => (a.id === approval.id ? approval : a))
            : [approval, ...prev];
        });
        setResolved((prev) => prev.filter((a) => a.id !== approval.id));
      } else {
        setResolved((prev) => {
          const exists = prev.some((a) => a.id === approval.id);
          return exists
            ? prev.map((a) => (a.id === approval.id ? approval : a))
            : [approval, ...prev];
        });
        setPending((prev) => prev.filter((a) => a.id !== approval.id));
      }
    }).catch(() => {});
  }, []);

  useWebSocket(handleWS);

  return (
    <div className="flex h-full flex-col gap-6 p-8">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm font-[family-name:var(--font-sans)]">
        <Link
          href={`/epics/${epicName}`}
          className="text-[#8ea2bd] transition-colors hover:text-white"
        >
          Home
        </Link>
        <span className="text-[#525e6e]">&rsaquo;</span>
        <span className="font-medium text-[#f5f9ff]">Approvals</span>
      </nav>

      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
          <p
            className="text-[12px] font-medium tracking-[1.2px] text-[#60a5fa] font-[family-name:var(--font-mono)]"
          >
            APPROVALS
          </p>
          <h1 className="text-[32px] font-bold tracking-tight text-[#f5f9ff] font-[family-name:var(--font-display)]">
            {epicName}
          </h1>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-1.5 rounded-2xl bg-[#dc2626] px-3 py-0 h-8">
            <span className="h-2 w-2 rounded-full bg-[#fca5a5]" />
            <span className="text-[13px] font-medium text-white font-[family-name:var(--font-sans)]">
              {pending.length} Pending
            </span>
          </span>
          <span className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
            {resolved.length} resolved
          </span>
        </div>
      </div>

      {/* Pending Review Section */}
      <div className="flex flex-col gap-3">
        <h2 className="text-[18px] font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
          Pending Review
        </h2>

        {pending.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 py-12 text-zinc-500">
            <ShieldAlert className="h-10 w-10" />
            <p className="text-sm font-[family-name:var(--font-sans)]">
              No pending approvals
            </p>
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            {pending.map((approval) => {
              const config = typeConfig[approval.type];
              const TypeIcon = config.icon;
              const displayName = approvalDisplayTitle(approval.type);

              return (
                <div
                  key={approval.id}
                  className="flex flex-col gap-4 rounded-2xl border border-[#22324a] bg-[#0f172a] p-5"
                >
                  {/* Header */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                      <TypeIcon
                        className="h-[18px] w-[18px] shrink-0"
                        style={{ color: config.iconColor }}
                      />
                      <h3 className="text-[16px] font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
                        {displayName}
                      </h3>
                    </div>
                    <span
                      className="inline-flex h-6 shrink-0 items-center rounded-xl px-2.5 text-[11px] font-medium font-[family-name:var(--font-mono)]"
                      style={{
                        backgroundColor: config.badgeBg,
                        color: config.badgeText,
                      }}
                    >
                      {config.label}
                    </span>
                  </div>

                  {/* Body */}
                  {approval.content && (
                    <p className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)] line-clamp-2">
                      {approval.content.slice(0, 200)}
                    </p>
                  )}

                  {/* Footer */}
                  <div className="flex items-center justify-between">
                    <span className="text-[12px] text-[#71717a] font-[family-name:var(--font-mono)]">
                      Requested{" "}
                      {new Date(approval.requested_at).toLocaleDateString("en-US", {
                        month: "short",
                        day: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                    <Link
                      href={`/epics/${epicName}/approvals/${approval.id}/review`}
                      className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-[#2563eb] px-4 text-[13px] font-semibold text-white transition-colors hover:bg-blue-500 font-[family-name:var(--font-sans)]"
                    >
                      Review
                      <ArrowRight className="h-3.5 w-3.5" />
                    </Link>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Resolved Section */}
      {resolved.length > 0 && (
        <div className="flex flex-col gap-3">
          <h2 className="text-[18px] font-semibold text-[#71717a] font-[family-name:var(--font-display)]">
            Resolved
          </h2>
          <div className="flex flex-col gap-3">
            {resolved.map((approval) => {
              const config = typeConfig[approval.type];
              const isApproved = approval.status === "approved";
              const displayName = approvalDisplayTitle(approval.type);

              return (
                <div
                  key={approval.id}
                  className="flex items-center gap-4 rounded-xl border border-[#22324a] bg-[#0f172a] px-5 py-4"
                >
                  {/* Status icon */}
                  {isApproved ? (
                    <CheckCircle2 className="h-[18px] w-[18px] shrink-0 text-[#22c55e]" />
                  ) : (
                    <XCircle className="h-[18px] w-[18px] shrink-0 text-[#ef4444]" />
                  )}

                  {/* Info column */}
                  <div className="flex flex-1 flex-col gap-0.5">
                    <span className="text-sm font-medium text-[#a1a1aa] font-[family-name:var(--font-sans)]">
                      {displayName}
                    </span>
                    <div className="flex items-center gap-2">
                      <span
                        className="inline-flex h-5 items-center rounded-[10px] px-2 text-[10px] font-medium font-[family-name:var(--font-mono)]"
                        style={{
                          backgroundColor: config.resolvedBadgeBg,
                          color: config.resolvedBadgeText,
                        }}
                      >
                        {config.label}
                      </span>
                      <span
                        className="text-[11px] font-medium font-[family-name:var(--font-mono)]"
                        style={{ color: isApproved ? "#4ade80" : "#f87171" }}
                      >
                        {isApproved ? "Approved" : "Rejected"}
                      </span>
                    </div>
                  </div>

                  {/* Timestamp */}
                  <span className="text-[12px] text-[#71717a] font-[family-name:var(--font-mono)]">
                    {new Date(approval.requested_at).toLocaleDateString("en-US", {
                      month: "short",
                      day: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
