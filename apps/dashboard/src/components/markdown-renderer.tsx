"use client";

import { useEffect, useRef, useState, useCallback, type ReactNode } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import mermaid from "mermaid";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import type { InlineComment } from "@/lib/types";

// ── Mermaid ──────────────────────────────────────────────────────────────────

let mermaidInitialized = false;

function MermaidBlock({ code }: { code: string }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!mermaidInitialized) {
      mermaid.initialize({
        startOnLoad: false,
        theme: "dark",
        themeVariables: {
          primaryColor: "#1e3a5f",
          primaryTextColor: "#f5f9ff",
          primaryBorderColor: "#22324a",
          lineColor: "#60a5fa",
          secondaryColor: "#0f172a",
          tertiaryColor: "#0b1120",
          fontFamily: "var(--font-sans)",
        },
      });
      mermaidInitialized = true;
    }

    const id = `mermaid-${Math.random().toString(36).slice(2, 9)}`;
    mermaid
      .render(id, code)
      .then(({ svg }) => {
        if (ref.current) ref.current.innerHTML = svg;
      })
      .catch(() => {
        if (ref.current)
          ref.current.innerHTML = `<pre class="text-red-400 text-sm">Failed to render mermaid diagram</pre>`;
      });
  }, [code]);

  return (
    <div
      ref={ref}
      className="my-4 flex justify-center rounded-xl border border-[#22324a] bg-[#0b1120] p-6"
    />
  );
}

// ── Comment Popover ──────────────────────────────────────────────────────────

interface PopoverState {
  x: number;
  y: number;
  text: string;
}

function CommentPopover({
  state,
  onSubmit,
  onCancel,
}: {
  state: PopoverState;
  onSubmit: (text: string, comment: string) => void;
  onCancel: () => void;
}) {
  const [comment, setComment] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  return (
    <div
      className="fixed z-50 flex w-80 flex-col gap-2 rounded-xl border border-[#22324a] bg-[#0f172a] p-3 shadow-2xl"
      style={{ left: state.x, top: state.y }}
    >
      <p className="truncate text-xs text-[#8ea2bd] font-[family-name:var(--font-mono)]">
        &ldquo;{state.text.slice(0, 60)}
        {state.text.length > 60 ? "…" : ""}&rdquo;
      </p>
      <textarea
        ref={textareaRef}
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        placeholder="Add a comment…"
        rows={3}
        className="resize-none rounded-lg border border-[#22324a] bg-[#0b1120] px-3 py-2 text-sm text-[#f5f9ff] placeholder-[#525e6e] outline-none focus:border-blue-500 font-[family-name:var(--font-sans)]"
        onKeyDown={(e) => {
          if (e.key === "Escape") onCancel();
          if (e.key === "Enter" && e.metaKey && comment.trim()) {
            onSubmit(state.text, comment.trim());
          }
        }}
      />
      <div className="flex items-center justify-between">
        <span className="text-[11px] text-[#525e6e] font-[family-name:var(--font-mono)]">
          ⌘ Enter to submit
        </span>
        <div className="flex gap-2">
          <button
            onClick={onCancel}
            className="rounded-lg px-3 py-1.5 text-xs font-medium text-[#8ea2bd] transition-colors hover:bg-[#22324a] font-[family-name:var(--font-sans)]"
          >
            Cancel
          </button>
          <button
            onClick={() => {
              if (comment.trim()) onSubmit(state.text, comment.trim());
            }}
            disabled={!comment.trim()}
            className="rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-blue-500 disabled:opacity-40 font-[family-name:var(--font-sans)]"
          >
            Add Comment
          </button>
        </div>
      </div>
    </div>
  );
}

// ── Inline Comment Highlight ─────────────────────────────────────────────────

function CommentHighlight({
  comment,
  children,
}: {
  comment: InlineComment;
  children: ReactNode;
}) {
  const [showTooltip, setShowTooltip] = useState(false);

  return (
    <span
      className="relative cursor-pointer rounded bg-yellow-500/20 px-0.5"
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      {children}
      {showTooltip && (
        <span className="absolute bottom-full left-0 z-40 mb-1 w-64 rounded-lg border border-[#22324a] bg-[#0f172a] p-2.5 text-xs shadow-xl">
          <span className="block text-[#f5f9ff] font-[family-name:var(--font-sans)]">
            {comment.comment}
          </span>
          <span className="mt-1 block text-[10px] text-[#525e6e] font-[family-name:var(--font-mono)]">
            {comment.created_at}
          </span>
        </span>
      )}
    </span>
  );
}

// ── Markdown Renderer ────────────────────────────────────────────────────────

interface MarkdownRendererProps {
  content: string;
  comments?: InlineComment[];
  onAddComment?: (comment: Omit<InlineComment, "id" | "created_at">) => void;
  onDeleteComment?: (commentId: number) => void;
}

export function MarkdownRenderer({
  content,
  comments,
  onAddComment,
  onDeleteComment,
}: MarkdownRendererProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [popover, setPopover] = useState<PopoverState | null>(null);

  const handleMouseUp = useCallback(() => {
    if (!onAddComment) return;

    const selection = window.getSelection();
    if (!selection || selection.isCollapsed || !selection.toString().trim()) {
      return;
    }

    const text = selection.toString().trim();
    const range = selection.getRangeAt(0);
    const rect = range.getBoundingClientRect();

    setPopover({
      x: Math.min(rect.left, window.innerWidth - 340),
      y: rect.top - 10 + window.scrollY,
      text,
    });
  }, [onAddComment]);

  const handleAddComment = useCallback(
    (selectedText: string, comment: string) => {
      onAddComment?.({ selected_text: selectedText, comment });
      setPopover(null);
      window.getSelection()?.removeAllRanges();
    },
    [onAddComment]
  );

  // Highlight commented text in content
  const highlightedContent = content;
  // For simplicity, comments are shown as a list below the content.
  // True inline highlight requires DOM manipulation after render which adds complexity.

  return (
    <div ref={containerRef} onMouseUp={handleMouseUp} className="relative">
      <article className="prose-custom max-w-none">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeRaw]}
          components={{
            h1: ({ children }) => (
              <h1 className="mb-4 mt-8 text-[28px] font-bold text-[#f5f9ff] font-[family-name:var(--font-display)]">
                {children}
              </h1>
            ),
            h2: ({ children }) => (
              <h2 className="mb-3 mt-7 text-[22px] font-bold text-[#f5f9ff] font-[family-name:var(--font-display)]">
                {children}
              </h2>
            ),
            h3: ({ children }) => (
              <h3 className="mb-2 mt-5 text-[18px] font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
                {children}
              </h3>
            ),
            h4: ({ children }) => (
              <h4 className="mb-2 mt-4 text-[16px] font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
                {children}
              </h4>
            ),
            p: ({ children }) => (
              <p className="mb-4 text-sm leading-relaxed text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                {children}
              </p>
            ),
            a: ({ href, children }) => (
              <a
                href={href}
                className="text-blue-400 underline decoration-blue-400/30 transition-colors hover:text-blue-300"
              >
                {children}
              </a>
            ),
            strong: ({ children }) => (
              <strong className="font-semibold text-[#f5f9ff]">
                {children}
              </strong>
            ),
            em: ({ children }) => (
              <em className="text-[#a5b4c8]">{children}</em>
            ),
            blockquote: ({ children }) => (
              <blockquote className="my-4 border-l-2 border-blue-500 pl-4 text-[#8ea2bd] italic">
                {children}
              </blockquote>
            ),
            ul: ({ children }) => (
              <ul className="mb-4 ml-4 list-disc space-y-1 text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                {children}
              </ul>
            ),
            ol: ({ children }) => (
              <ol className="mb-4 ml-4 list-decimal space-y-1 text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                {children}
              </ol>
            ),
            li: ({ children }) => (
              <li className="text-sm text-[#8ea2bd]">{children}</li>
            ),
            table: ({ children }) => (
              <div className="my-4 overflow-x-auto rounded-xl border border-[#22324a]">
                <table className="w-full text-sm">{children}</table>
              </div>
            ),
            thead: ({ children }) => (
              <thead className="border-b border-[#22324a] bg-[#0b1120]">
                {children}
              </thead>
            ),
            tbody: ({ children }) => (
              <tbody className="divide-y divide-[#22324a]">{children}</tbody>
            ),
            tr: ({ children }) => (
              <tr className="even:bg-[#0b1120]/50">{children}</tr>
            ),
            th: ({ children }) => (
              <th className="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-[#8ea2bd] font-[family-name:var(--font-mono)]">
                {children}
              </th>
            ),
            td: ({ children }) => (
              <td className="px-4 py-2.5 text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                {children}
              </td>
            ),
            code: ({ className, children }) => {
              const match = /language-(\w+)/.exec(className || "");
              const language = match?.[1];

              if (language === "mermaid") {
                return <MermaidBlock code={String(children).trim()} />;
              }

              if (className) {
                return (
                  <div className="my-4 overflow-x-auto rounded-xl border border-[#22324a] bg-[#0b1120]">
                    <div className="flex items-center justify-between border-b border-[#22324a] px-4 py-2">
                      <span className="text-[11px] font-medium uppercase tracking-wider text-[#525e6e] font-[family-name:var(--font-mono)]">
                        {language || "code"}
                      </span>
                    </div>
                    <pre className="overflow-x-auto p-4">
                      <code className="text-[13px] leading-relaxed text-[#e2e8f0] font-[family-name:var(--font-mono)]">
                        {children}
                      </code>
                    </pre>
                  </div>
                );
              }

              return (
                <code className="rounded bg-[#0b1120] px-1.5 py-0.5 text-[13px] text-[#e2e8f0] font-[family-name:var(--font-mono)]">
                  {children}
                </code>
              );
            },
            pre: ({ children }) => <>{children}</>,
            hr: () => <hr className="my-6 border-[#22324a]" />,
            input: ({ checked, ...props }) => (
              <input
                type="checkbox"
                checked={checked}
                readOnly
                className="mr-2 rounded border-[#22324a] bg-[#0b1120] text-blue-500"
                {...props}
              />
            ),
          }}
        >
          {highlightedContent}
        </ReactMarkdown>
      </article>

      {/* Inline comments list */}
      {comments && comments.length > 0 && (
        <div className="mt-8 border-t border-[#22324a] pt-6">
          <h4 className="mb-3 text-sm font-semibold text-[#f5f9ff] font-[family-name:var(--font-display)]">
            Comments ({comments.length})
          </h4>
          <div className="space-y-3">
            {comments.map((c) => (
              <div key={c.id} className="rounded-xl border border-[#22324a] bg-[#0b1120] p-3">
                <div className="flex items-start justify-between">
                  <p className="mb-1 truncate text-xs text-[#525e6e] font-[family-name:var(--font-mono)]">
                    On: &ldquo;{c.selected_text.slice(0, 80)}
                    {c.selected_text.length > 80 ? "…" : ""}&rdquo;
                  </p>
                  {onDeleteComment && (
                    <button
                      onClick={() => onDeleteComment(c.id)}
                      className="ml-2 shrink-0 rounded p-1 text-[#525e6e] transition-colors hover:bg-[#22324a] hover:text-red-400"
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  )}
                </div>
                <p className="text-sm text-[#8ea2bd] font-[family-name:var(--font-sans)]">
                  {c.comment}
                </p>
                <p className="mt-1 text-[10px] text-[#525e6e] font-[family-name:var(--font-mono)]">
                  {c.created_at}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Comment popover */}
      {popover && (
        <CommentPopover
          state={popover}
          onSubmit={handleAddComment}
          onCancel={() => {
            setPopover(null);
            window.getSelection()?.removeAllRanges();
          }}
        />
      )}
    </div>
  );
}
