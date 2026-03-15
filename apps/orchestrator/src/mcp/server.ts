import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import type { Database } from "bun:sqlite";
import { ApprovalStore } from "../db/stores/approval-store.ts";

export async function startMcpServer(db: Database): Promise<void> {
  const server = new McpServer({
    name: "oraculo",
    version: "0.1.0",
  });

  const approvalStore = new ApprovalStore(db);

  // Approval bridge - maps approval ID to resolve function
  const pendingApprovals = new Map<string, (verdict: string) => void>();

  server.tool(
    "request_approval",
    "Submit an artifact for human approval. Blocks until a verdict (approved, rejected, or needs_revision) is recorded. Use for operational gates only (qa-escalation, execution-plan, design).",
    {
      type: z.enum(["qa-escalation", "execution-plan", "design"]),
      description: z.string(),
      epic_id: z.number().optional(),
      story_id: z.number().optional(),
      task_id: z.number().optional(),
      version_id: z.number().optional(),
    },
    async (args) => {
      const approval = approvalStore.Create(
        args.type,
        args.epic_id ?? null,
        args.story_id ?? null,
        args.description,
      );

      // Block until verdict
      const verdict = await new Promise<string>((resolve) => {
        pendingApprovals.set(approval.id, resolve);
      });

      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify({ id: approval.id, verdict }),
          },
        ],
      };
    },
  );

  server.tool(
    "approval_status",
    "Get the current status of an operational approval request.",
    {
      id: z.string(),
    },
    async (args) => {
      const approval = approvalStore.Get(args.id);
      const comments = approvalStore.ListComments(args.id);
      return {
        content: [
          {
            type: "text" as const,
            text: JSON.stringify({ ...approval, comments }),
          },
        ],
      };
    },
  );

  const transport = new StdioServerTransport();
  await server.connect(transport);
}

// Export the pending approvals map for the HTTP server to resolve
export function createApprovalBridge() {
  const pending = new Map<string, (verdict: string) => void>();
  return {
    wait(id: string): Promise<string> {
      return new Promise((resolve) => {
        pending.set(id, resolve);
      });
    },
    resolve(id: string, verdict: string): boolean {
      const resolver = pending.get(id);
      if (resolver) {
        resolver(verdict);
        pending.delete(id);
        return true;
      }
      return false;
    },
    hasPending(id: string): boolean {
      return pending.has(id);
    },
  };
}
