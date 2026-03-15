import { describe, it, expect } from "bun:test";
import { createApprovalBridge } from "../server.ts";

describe("ApprovalBridge", () => {
  it("resolves pending approval", async () => {
    const bridge = createApprovalBridge();
    const promise = bridge.wait("test-1");
    bridge.resolve("test-1", "approved");
    const result = await promise;
    expect(result).toBe("approved");
  });

  it("returns false for non-pending approval", () => {
    const bridge = createApprovalBridge();
    expect(bridge.resolve("nonexistent", "approved")).toBe(false);
  });

  it("tracks pending approvals", () => {
    const bridge = createApprovalBridge();
    bridge.wait("test-2");
    expect(bridge.hasPending("test-2")).toBe(true);
    expect(bridge.hasPending("other")).toBe(false);
  });

  it("resolves with rejected verdict", async () => {
    const bridge = createApprovalBridge();
    const promise = bridge.wait("test-3");
    bridge.resolve("test-3", "rejected");
    const result = await promise;
    expect(result).toBe("rejected");
  });

  it("resolves with needs_revision verdict", async () => {
    const bridge = createApprovalBridge();
    const promise = bridge.wait("test-4");
    bridge.resolve("test-4", "needs_revision");
    const result = await promise;
    expect(result).toBe("needs_revision");
  });

  it("removes entry after resolve", () => {
    const bridge = createApprovalBridge();
    bridge.wait("test-5");
    expect(bridge.hasPending("test-5")).toBe(true);
    bridge.resolve("test-5", "approved");
    expect(bridge.hasPending("test-5")).toBe(false);
  });
});
