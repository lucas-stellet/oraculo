"use client";

import { useCallback, useState, useRef, useEffect, type RefObject } from "react";
import {
  Heading1,
  Heading2,
  Heading3,
  Bold,
  Italic,
  Code,
  Table,
  List,
  ListOrdered,
  ListChecks,
  Link2,
  Quote,
  Minus,
  FileCode,
  GitBranch,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface ToolbarAction {
  icon: React.ElementType;
  label: string;
  snippet: string;
  /** If true, wraps the selected text instead of inserting */
  wrap?: { before: string; after: string };
  /** Cursor offset from end of inserted text (negative = move back) */
  cursorOffset?: number;
  /** If true, this action uses a custom handler instead of direct snippet insert */
  custom?: boolean;
}

function generateTable(cols: number): string {
  const header = "| " + Array.from({ length: cols }, (_, i) => `Column ${i + 1}`).join(" | ") + " |";
  const separator = "| " + Array.from({ length: cols }, () => "----------").join(" | ") + " |";
  const row = "| " + Array.from({ length: cols }, () => "Cell").join("      | ") + "      |";
  return `${header}\n${separator}\n${row}\n${row}`;
}

const toolbarGroups: { label: string; actions: ToolbarAction[] }[] = [
  {
    label: "Headings",
    actions: [
      { icon: Heading1, label: "Heading 1", snippet: "# ", cursorOffset: 0 },
      { icon: Heading2, label: "Heading 2", snippet: "## ", cursorOffset: 0 },
      { icon: Heading3, label: "Heading 3", snippet: "### ", cursorOffset: 0 },
    ],
  },
  {
    label: "Inline",
    actions: [
      { icon: Bold, label: "Bold", snippet: "**text**", wrap: { before: "**", after: "**" }, cursorOffset: -2 },
      { icon: Italic, label: "Italic", snippet: "*text*", wrap: { before: "*", after: "*" }, cursorOffset: -1 },
      { icon: Code, label: "Inline code", snippet: "`code`", wrap: { before: "`", after: "`" }, cursorOffset: -1 },
      { icon: Link2, label: "Link", snippet: "[text](url)", cursorOffset: -1 },
    ],
  },
  {
    label: "Blocks",
    actions: [
      {
        icon: FileCode,
        label: "Code block",
        snippet: "```\n\n```",
        cursorOffset: -4,
      },
      {
        icon: GitBranch,
        label: "Mermaid diagram",
        snippet: "```mermaid\ngraph TD\n    A[Start] --> B[End]\n```",
        cursorOffset: -4,
      },
      { icon: Quote, label: "Blockquote", snippet: "> ", cursorOffset: 0 },
      { icon: Minus, label: "Divider", snippet: "\n---\n", cursorOffset: 0 },
    ],
  },
  {
    label: "Lists",
    actions: [
      { icon: List, label: "Bullet list", snippet: "- Item 1\n- Item 2\n- Item 3", cursorOffset: 0 },
      { icon: ListOrdered, label: "Numbered list", snippet: "1. Item 1\n2. Item 2\n3. Item 3", cursorOffset: 0 },
      { icon: ListChecks, label: "Checklist", snippet: "- [ ] Task 1\n- [ ] Task 2\n- [ ] Task 3", cursorOffset: 0 },
    ],
  },
  {
    label: "Table",
    actions: [
      {
        icon: Table,
        label: "Table",
        snippet: "",
        cursorOffset: 0,
        custom: true,
      },
    ],
  },
];

interface MarkdownToolbarProps {
  textareaRef: RefObject<HTMLTextAreaElement | null>;
  onContentChange: (content: string) => void;
}

export function MarkdownToolbar({
  textareaRef,
  onContentChange,
}: MarkdownToolbarProps) {
  const [tablePopover, setTablePopover] = useState<{ top: number; left: number } | null>(null);
  const [tableCols, setTableCols] = useState("3");
  const tableInputRef = useRef<HTMLInputElement>(null);
  const tableButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (tablePopover) {
      tableInputRef.current?.focus();
      tableInputRef.current?.select();
    }
  }, [tablePopover]);

  /** Insert text via execCommand to preserve the browser's native undo stack */
  const execInsert = useCallback(
    (text: string, cursorOffset = 0) => {
      const textarea = textareaRef.current;
      if (!textarea) return;

      textarea.focus();
      document.execCommand("insertText", false, text);

      // Sync the ref-based content after insertion
      onContentChange(textarea.value);

      if (cursorOffset !== 0) {
        const pos = textarea.selectionStart + cursorOffset;
        textarea.setSelectionRange(pos, pos);
      }
    },
    [textareaRef, onContentChange]
  );

  const insertAtCursor = useCallback(
    (snippet: string, cursorOffset = 0) => {
      const textarea = textareaRef.current;
      if (!textarea) return;

      const { selectionStart, value } = textarea;

      const needsNewline =
        selectionStart > 0 &&
        value[selectionStart - 1] !== "\n" &&
        (snippet.startsWith("|") || snippet.startsWith("#") ||
          snippet.startsWith("-") || snippet.startsWith("1.") ||
          snippet.startsWith(">"));
      const prefix = needsNewline ? "\n" : "";

      execInsert(prefix + snippet, cursorOffset);
    },
    [textareaRef, execInsert]
  );

  const insertSnippet = useCallback(
    (action: ToolbarAction) => {
      const textarea = textareaRef.current;
      if (!textarea) return;

      const { selectionStart, selectionEnd, value } = textarea;
      const selectedText = value.slice(selectionStart, selectionEnd);

      if (action.wrap && selectedText) {
        const text = `${action.wrap.before}${selectedText}${action.wrap.after}`;
        execInsert(text);
      } else if (action.wrap && !selectedText) {
        execInsert(action.snippet, action.cursorOffset ?? 0);
      } else {
        const needsNewline =
          selectionStart > 0 &&
          value[selectionStart - 1] !== "\n" &&
          (action.snippet.startsWith("#") ||
            action.snippet.startsWith("-") ||
            action.snippet.startsWith("1.") ||
            action.snippet.startsWith(">"));
        const prefix = needsNewline ? "\n" : "";
        execInsert(prefix + action.snippet, action.cursorOffset ?? 0);
      }
    },
    [textareaRef, execInsert]
  );

  const handleTableClick = useCallback(() => {
    const btn = tableButtonRef.current;
    if (!btn) return;
    const rect = btn.getBoundingClientRect();
    setTablePopover({ top: rect.top, left: rect.right + 8 });
    setTableCols("3");
  }, []);

  const submitTable = useCallback(() => {
    const cols = Math.max(1, Math.min(10, parseInt(tableCols) || 3));
    insertAtCursor(generateTable(cols));
    setTablePopover(null);
  }, [tableCols, insertAtCursor]);

  return (
    <div className="sticky top-4 flex flex-col gap-4">
      {toolbarGroups.map((group) => (
        <div key={group.label} className="flex flex-col gap-1">
          <span className="px-1 text-[9px] uppercase tracking-widest text-[#525e6e] font-[family-name:var(--font-mono)]">
            {group.label}
          </span>
          <div className="flex flex-col gap-0.5">
            {group.actions.map((action) => (
              <button
                key={action.label}
                ref={action.custom ? tableButtonRef : undefined}
                onClick={() => {
                  if (action.custom) {
                    handleTableClick();
                  } else {
                    insertSnippet(action);
                  }
                }}
                title={action.label}
                className={cn(
                  "group flex items-center gap-2.5 rounded-lg px-2.5 py-2 transition-colors",
                  "text-[#8ea2bd] hover:bg-[#22324a] hover:text-[#f5f9ff]"
                )}
              >
                <action.icon className="h-4 w-4 shrink-0" />
                <span className="text-[11px] font-medium font-[family-name:var(--font-sans)]">
                  {action.label}
                </span>
              </button>
            ))}
          </div>
        </div>
      ))}

      {/* Table column popover — rendered with fixed positioning to escape overflow clipping */}
      {tablePopover && (
        <div
          className="fixed z-50 flex items-center gap-2 rounded-xl border border-[#22324a] bg-[#0f172a] px-3 py-2.5 shadow-2xl"
          style={{ top: tablePopover.top, left: tablePopover.left }}
        >
          <label className="text-[11px] text-[#8ea2bd] whitespace-nowrap font-[family-name:var(--font-sans)]">
            Columns
          </label>
          <input
            ref={tableInputRef}
            type="number"
            min={1}
            max={10}
            value={tableCols}
            onChange={(e) => setTableCols(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") submitTable();
              if (e.key === "Escape") setTablePopover(null);
            }}
            className="w-14 rounded-lg border border-[#22324a] bg-[#0b1120] px-2 py-1 text-center text-xs text-[#f5f9ff] outline-none focus:border-blue-500 font-[family-name:var(--font-mono)]"
          />
          <button
            onClick={submitTable}
            className="rounded-lg bg-blue-600 px-2.5 py-1 text-[11px] font-medium text-white transition-colors hover:bg-blue-500 font-[family-name:var(--font-sans)]"
          >
            Insert
          </button>
        </div>
      )}
    </div>
  );
}
