"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Eye, Pencil, Check, X } from "lucide-react";
import { cn, approvalDisplayTitle } from "@/lib/utils";
import { useSidebar } from "@/lib/sidebar-context";
import { api } from "@/lib/api";
import { MarkdownRenderer } from "@/components/markdown-renderer";
import { MarkdownToolbar } from "@/components/markdown-toolbar";
import type { Approval, ApprovalType, InlineComment } from "@/lib/types";

const typeBadge: Record<ApprovalType, { label: string; bg: string; text: string }> = {
  design: { label: "Design", bg: "#7c3aed", text: "#f5f3ff" },
  "qa-escalation": { label: "QA Escalation", bg: "#dc2626", text: "#fef2f2" },
  "execution-plan": { label: "Execution Plan", bg: "#059669", text: "#ecfdf5" },
};

export default function ReviewPage() {
  const pathname = usePathname();
  const segs = pathname.split("/").filter(Boolean);
  // /epics/{id}/approvals/{approvalId}/review
  const epicName = segs[1];
  const approvalId = segs[3];

  const { collapsed, setCollapsed } = useSidebar();
  const prevCollapsed = useRef(collapsed);

  const [approval, setApproval] = useState<Approval | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    prevCollapsed.current = collapsed;
    setCollapsed(true);
    return () => setCollapsed(prevCollapsed.current);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    api.getApproval(approvalId)
      .then(setApproval)
      .catch(() => setApproval(null))
      .finally(() => setLoading(false));
  }, [approvalId]);

  const [mode, setMode] = useState<"preview" | "edit">("preview");
  const editContentRef = useRef("");
  const [, forceRender] = useState(0);
  const [comments, setComments] = useState<InlineComment[]>([]);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Initialize editContentRef once approval loads
  useEffect(() => {
    if (approval) {
      editContentRef.current = approval.content;
    }
  }, [approval]);

  const syncContent = useCallback(() => {
    if (textareaRef.current) {
      editContentRef.current = textareaRef.current.value;
    }
  }, []);

  const handleContentChange = useCallback(
    (newValue: string) => {
      editContentRef.current = newValue;
      forceRender((n) => n + 1);
    },
    []
  );

  const handleAddComment = useCallback(
    (comment: Omit<InlineComment, "id" | "created_at">) => {
      setComments((prev) => [
        ...prev,
        {
          ...comment,
          id: prev.length + 1,
          created_at: "Just now",
        },
      ]);
    },
    []
  );

  const handleVerdict = useCallback(
    async (verdict: "approved" | "rejected") => {
      if (!approval || submitting) return;
      setSubmitting(true);
      try {
        const commentText = comments.map((c) => `[${c.selected_text}]: ${c.comment}`).join("\n") || "";
        await api.submitVerdict(approval.id, verdict, commentText);
        // Refetch to update UI
        const updated = await api.getApproval(approvalId);
        setApproval(updated);
      } catch {
        // Silently handle error
      } finally {
        setSubmitting(false);
      }
    },
    [approval, approvalId, comments, submitting]
  );

  if (loading) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-[#8ea2bd] font-[family-name:var(--font-sans)]">
          Loading...
        </p>
      </div>
    );
  }

  if (!approval) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-[#8ea2bd] font-[family-name:var(--font-sans)]">
          Approval not found
        </p>
      </div>
    );
  }

  const badge = typeBadge[approval.type];
  const displayContent = editContentRef.current || approval.content;
  const displayTitle = approvalDisplayTitle(approval.type);

  return (
    <div className="flex h-full flex-col">
      {/* Toolbar */}
      <div className="flex shrink-0 items-center justify-between border-b border-[#22324a] bg-[#0f172a] px-6 py-3">
        <div className="flex items-center gap-4">
          <Link
            href={`/epics/${epicName}/approvals`}
            className="inline-flex items-center gap-1.5 text-sm text-[#8ea2bd] transition-colors hover:text-white font-[family-name:var(--font-sans)]"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Approvals
          </Link>

          <div className="h-5 w-px bg-[#22324a]" />

          <h1 className="text-[15px] font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
            {displayTitle}
          </h1>

          <span
            className="inline-flex h-5 shrink-0 items-center rounded-lg px-2 text-[10px] font-medium font-[family-name:var(--font-mono)]"
            style={{ backgroundColor: badge.bg, color: badge.text }}
          >
            {badge.label}
          </span>
        </div>

        <div className="flex items-center gap-3">
          {/* Mode toggle */}
          <div className="flex rounded-lg border border-[#22324a] bg-[#0b1120] p-0.5">
            <button
              onClick={() => setMode("preview")}
              className={cn(
                "inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors font-[family-name:var(--font-sans)]",
                mode === "preview"
                  ? "bg-[#22324a] text-[#f5f9ff]"
                  : "text-[#8ea2bd] hover:text-[#f5f9ff]"
              )}
            >
              <Eye className="h-3.5 w-3.5" />
              Preview
            </button>
            <button
              onClick={() => setMode("edit")}
              className={cn(
                "inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors font-[family-name:var(--font-sans)]",
                mode === "edit"
                  ? "bg-[#22324a] text-[#f5f9ff]"
                  : "text-[#8ea2bd] hover:text-[#f5f9ff]"
              )}
            >
              <Pencil className="h-3.5 w-3.5" />
              Edit
            </button>
          </div>

          <div className="h-5 w-px bg-[#22324a]" />

          {/* Actions */}
          <button
            onClick={() => handleVerdict("approved")}
            disabled={submitting}
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-[#22c55e]/30 bg-[#22c55e]/10 px-3 text-xs font-semibold text-[#22c55e] transition-colors hover:bg-[#22c55e]/20 disabled:opacity-50 font-[family-name:var(--font-sans)]"
          >
            <Check className="h-3.5 w-3.5" />
            Approve
          </button>
          <button
            onClick={() => handleVerdict("rejected")}
            disabled={submitting}
            className="inline-flex h-8 items-center gap-1.5 rounded-lg border border-[#ef4444]/30 bg-[#ef4444]/10 px-3 text-xs font-semibold text-[#ef4444] transition-colors hover:bg-[#ef4444]/20 disabled:opacity-50 font-[family-name:var(--font-sans)]"
          >
            <X className="h-3.5 w-3.5" />
            Reject
          </button>
        </div>
      </div>

      {/* Content area */}
      <div className="flex-1 overflow-y-auto">
        {mode === "preview" ? (
          <div className="mx-auto max-w-4xl px-8 py-8">
            <MarkdownRenderer
              content={displayContent}
              comments={comments}
              onAddComment={handleAddComment}
            />
          </div>
        ) : (
          <div className="flex h-full">
            {/* Floating toolbar */}
            <div className="w-40 shrink-0 overflow-y-auto border-r border-[#22324a] bg-[#0f172a]/60 px-2 py-4">
              <MarkdownToolbar
                textareaRef={textareaRef}
                onContentChange={handleContentChange}
              />
            </div>
            {/* Editor */}
            <div className="flex flex-1 flex-col px-6 py-4">
              <textarea
                ref={textareaRef}
                defaultValue={editContentRef.current}
                onInput={syncContent}
                className="flex-1 w-full resize-none rounded-xl border border-[#22324a] bg-[#0b1120] p-6 text-[13px] leading-relaxed text-[#f5f9ff] placeholder-[#525e6e] outline-none focus:border-blue-500 font-[family-name:var(--font-mono)]"
                spellCheck={false}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
