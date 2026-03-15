import { describe, it, expect } from "bun:test";
import { createApprovalBridge } from "../approval-bridge.ts";

describe("ApprovalBridge", () => {
  it("resolves pending approval", async () => {
    const bridge = createApprovalBridge();
    const promise = bridge.wait("test-1");
    bridge.decide("test-1", "approved");
    const result = await promise;
    expect(result).toBe("approved");
  });

  it("returns false for unknown approval", () => {
    const bridge = createApprovalBridge();
    expect(bridge.decide("unknown", "approved")).toBe(false);
  });

  it("tracks pending state", () => {
    const bridge = createApprovalBridge();
    bridge.wait("test-2");
    expect(bridge.hasPending("test-2")).toBe(true);
    expect(bridge.hasPending("other")).toBe(false);
  });

  it("removes entry after decide", () => {
    const bridge = createApprovalBridge();
    bridge.wait("test-3");
    expect(bridge.hasPending("test-3")).toBe(true);
    bridge.decide("test-3", "rejected");
    expect(bridge.hasPending("test-3")).toBe(false);
  });

  it("resolves with different verdicts", async () => {
    const bridge = createApprovalBridge();

    const p1 = bridge.wait("v1");
    bridge.decide("v1", "rejected");
    expect(await p1).toBe("rejected");

    const p2 = bridge.wait("v2");
    bridge.decide("v2", "needs_revision");
    expect(await p2).toBe("needs_revision");
  });

  it("times out if no decision arrives", async () => {
    const bridge = createApprovalBridge();
    try {
      await bridge.wait("timeout-1", 10); // 10ms timeout
      throw new Error("Expected timeout");
    } catch (err) {
      expect((err as Error).message).toContain("timed out");
    }
  });
});
