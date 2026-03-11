"use client";

import { useEffect, useRef, useState, useCallback } from "react";
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

// ── Comment View Popover ─────────────────────────────────────────────────────

interface ViewPopoverState {
  x: number;
  y: number;
  comment: InlineComment;
}

function CommentViewPopover({
  state,
  onDelete,
  onClose,
}: {
  state: ViewPopoverState;
  onDelete?: (id: number) => void;
  onClose: () => void;
}) {
  return (
    <div
      className="fixed z-50 flex w-80 flex-col gap-2 rounded-xl border border-[#22324a] bg-[#0f172a] p-3 shadow-2xl"
      style={{ left: state.x, top: state.y }}
    >
      <div className="flex items-start justify-between gap-2">
        <p className="truncate text-xs text-[#8ea2bd] font-[family-name:var(--font-mono)]">
          &ldquo;{state.comment.selected_text.slice(0, 60)}
          {state.comment.selected_text.length > 60 ? "…" : ""}&rdquo;
        </p>
        <button
          onClick={onClose}
          className="shrink-0 rounded p-1 text-[#525e6e] transition-colors hover:bg-[#22324a] hover:text-[#8ea2bd]"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
      <div className="max-h-48 min-h-[3rem] overflow-y-auto resize-y rounded-lg border border-[#22324a] bg-[#0b1120] px-3 py-2 text-sm text-[#f5f9ff] font-[family-name:var(--font-sans)]">
        {state.comment.comment}
      </div>
      <div className="flex items-center justify-between">
        <span className="text-[10px] text-[#525e6e] font-[family-name:var(--font-mono)]">
          {state.comment.created_at}
        </span>
        {onDelete && (
          <button
            onClick={() => {
              onDelete(state.comment.id);
              onClose();
            }}
            className="rounded-lg px-3 py-1.5 text-xs font-medium text-red-400 transition-colors hover:bg-[#22324a] font-[family-name:var(--font-sans)]"
          >
            Delete
          </button>
        )}
      </div>
    </div>
  );
}

// ── Highlight Helpers ─────────────────────────────────────────────────────────

function highlightText(container: HTMLElement, comment: InlineComment) {
  const article = container.querySelector('article');
  if (!article) return;

  const walker = document.createTreeWalker(article, NodeFilter.SHOW_TEXT);
  const textNodes: Text[] = [];

  let node: Text | null;
  while ((node = walker.nextNode() as Text | null)) {
    textNodes.push(node);
  }

  const searchText = comment.selected_text;

  for (let i = 0; i < textNodes.length; i++) {
    const textNode = textNodes[i];
    const idx = textNode.textContent?.indexOf(searchText) ?? -1;

    if (idx !== -1) {
      // Found within a single text node — split and wrap
      const range = document.createRange();
      range.setStart(textNode, idx);
      range.setEnd(textNode, idx + searchText.length);

      const highlight = document.createElement('span');
      highlight.setAttribute('data-comment-highlight', String(comment.id));
      highlight.className = 'cursor-pointer rounded bg-yellow-500/20 px-0.5 transition-colors hover:bg-yellow-500/30';
      highlight.title = comment.comment;

      range.surroundContents(highlight);
      return; // Only highlight first occurrence
    }
  }

  // If not found in a single node, try across adjacent nodes
  for (let i = 0; i < textNodes.length; i++) {
    let combined = '';
    const nodeSpans: Array<{ node: Text; start: number; length: number }> = [];

    for (let j = i; j < textNodes.length && combined.length < searchText.length * 2; j++) {
      const text = textNodes[j].textContent || '';
      nodeSpans.push({ node: textNodes[j], start: combined.length, length: text.length });
      combined += text;

      const idx = combined.indexOf(searchText);
      if (idx !== -1) {
        // Found across multiple nodes — wrap each portion
        const endIdx = idx + searchText.length;

        for (const span of nodeSpans) {
          const spanEnd = span.start + span.length;
          const overlapStart = Math.max(idx, span.start);
          const overlapEnd = Math.min(endIdx, spanEnd);

          if (overlapStart < overlapEnd) {
            const localStart = overlapStart - span.start;
            const localEnd = overlapEnd - span.start;

            const range = document.createRange();
            range.setStart(span.node, localStart);
            range.setEnd(span.node, localEnd);

            const highlight = document.createElement('span');
            highlight.setAttribute('data-comment-highlight', String(comment.id));
            highlight.className = 'cursor-pointer rounded bg-yellow-500/20 px-0.5 transition-colors hover:bg-yellow-500/30';
            highlight.title = comment.comment;

            range.surroundContents(highlight);
          }
        }
        return;
      }
    }
  }
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
  const [viewPopover, setViewPopover] = useState<ViewPopoverState | null>(null);

  const handleMouseUp = useCallback(() => {
    if (!onAddComment) return;

    const selection = window.getSelection();
    if (!selection || selection.isCollapsed || !selection.toString().trim()) {
      return;
    }

    const text = selection.toString().trim();

    // Check if this text already has a comment
    if (comments?.some(c => c.selected_text === text)) {
      return; // Don't create duplicate comment
    }

    const range = selection.getRangeAt(0);
    const rect = range.getBoundingClientRect();

    setPopover({
      x: Math.min(rect.left, window.innerWidth - 340),
      y: rect.top - 10 + window.scrollY,
      text,
    });
  }, [onAddComment, comments]);

  const handleHighlightClick = useCallback((e: React.MouseEvent) => {
    const target = e.target as HTMLElement;
    const commentId = target.getAttribute('data-comment-highlight');
    if (!commentId || !comments) return;

    const comment = comments.find(c => String(c.id) === commentId);
    if (!comment) return;

    setViewPopover({
      x: Math.min(e.clientX, window.innerWidth - 340),
      y: e.clientY - 10,
      comment,
    });
  }, [comments]);

  useEffect(() => {
    if (!containerRef.current || !comments || comments.length === 0) return;

    // Clean up previous highlights
    const existing = containerRef.current.querySelectorAll('[data-comment-highlight]');
    existing.forEach((el) => {
      const parent = el.parentNode;
      if (parent) {
        // Replace highlight span with its text content
        const text = document.createTextNode(el.textContent || '');
        parent.replaceChild(text, el);
        parent.normalize(); // Merge adjacent text nodes
      }
    });

    // Apply highlights for each comment
    comments.forEach((comment) => {
      highlightText(containerRef.current!, comment);
    });
  }, [comments, content]);

  const handleAddComment = useCallback(
    (selectedText: string, comment: string) => {
      onAddComment?.({ selected_text: selectedText, comment });
      setPopover(null);
      window.getSelection()?.removeAllRanges();
    },
    [onAddComment]
  );

  return (
    <div ref={containerRef} onMouseUp={handleMouseUp} onClick={handleHighlightClick} className="relative">
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
          {content}
        </ReactMarkdown>
      </article>

      {/* Comment creation popover */}
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

      {/* Comment view popover */}
      {viewPopover && (
        <CommentViewPopover
          state={viewPopover}
          onDelete={onDeleteComment}
          onClose={() => setViewPopover(null)}
        />
      )}
    </div>
  );
}
