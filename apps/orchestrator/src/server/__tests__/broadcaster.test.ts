import { describe, test, expect, beforeEach } from "bun:test";
import { Broadcaster } from "../broadcaster.ts";

describe("Broadcaster", () => {
  let bc: Broadcaster;

  beforeEach(() => {
    bc = new Broadcaster();
  });

  test("starts with empty buffer", () => {
    expect(bc.recent()).toEqual([]);
    expect(bc.subscriberCount).toBe(0);
  });

  test("push stores entries in recent()", () => {
    bc.push({ msg: "one" });
    bc.push({ msg: "two" });

    const recent = bc.recent();
    expect(recent).toHaveLength(2);
    expect(JSON.parse(recent[0])).toEqual({ msg: "one" });
    expect(JSON.parse(recent[1])).toEqual({ msg: "two" });
  });

  test("ring buffer caps at 200 entries", () => {
    for (let i = 0; i < 250; i++) {
      bc.push({ i });
    }

    const recent = bc.recent();
    expect(recent).toHaveLength(200);

    // First entry should be i=50 (oldest surviving)
    expect(JSON.parse(recent[0])).toEqual({ i: 50 });
    // Last entry should be i=249
    expect(JSON.parse(recent[199])).toEqual({ i: 249 });
  });

  test("subscribe receives live pushes", () => {
    const received: string[] = [];
    bc.subscribe((data) => received.push(data));

    bc.push({ event: "hello" });

    expect(received).toHaveLength(1);
    expect(JSON.parse(received[0])).toEqual({ event: "hello" });
  });

  test("multiple subscribers all receive", () => {
    const r1: string[] = [];
    const r2: string[] = [];
    bc.subscribe((d) => r1.push(d));
    bc.subscribe((d) => r2.push(d));

    bc.push({ x: 1 });

    expect(r1).toHaveLength(1);
    expect(r2).toHaveLength(1);
  });

  test("unsubscribe stops delivery", () => {
    const received: string[] = [];
    const unsub = bc.subscribe((d) => received.push(d));

    bc.push({ a: 1 });
    unsub();
    bc.push({ b: 2 });

    expect(received).toHaveLength(1);
    expect(bc.subscriberCount).toBe(0);
  });

  test("subscriber that throws is removed", () => {
    let callCount = 0;
    bc.subscribe(() => {
      callCount++;
      throw new Error("boom");
    });

    bc.push({ msg: "first" });
    expect(callCount).toBe(1);
    expect(bc.subscriberCount).toBe(0);

    // Second push should not call removed subscriber
    bc.push({ msg: "second" });
    expect(callCount).toBe(1);
  });

  test("recent() returns a copy", () => {
    bc.push({ x: 1 });
    const copy = bc.recent();
    copy.push("extra");
    expect(bc.recent()).toHaveLength(1);
  });
});
