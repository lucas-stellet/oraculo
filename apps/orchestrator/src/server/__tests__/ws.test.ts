import { describe, test, expect, beforeEach } from "bun:test";
import { WebSocketHub, type WSClient } from "../ws.ts";

function makeMockWS(readyState = 1): WSClient & { messages: string[] } {
  const messages: string[] = [];
  return {
    readyState,
    messages,
    send(msg: string) {
      messages.push(msg);
    },
  };
}

describe("WebSocketHub", () => {
  let hub: WebSocketHub;

  beforeEach(() => {
    hub = new WebSocketHub();
  });

  test("starts empty", () => {
    expect(hub.size).toBe(0);
  });

  test("add and remove clients", () => {
    const ws = makeMockWS();
    hub.add(ws);
    expect(hub.size).toBe(1);
    hub.remove(ws);
    expect(hub.size).toBe(0);
  });

  test("remove is idempotent", () => {
    const ws = makeMockWS();
    hub.add(ws);
    hub.remove(ws);
    hub.remove(ws);
    expect(hub.size).toBe(0);
  });

  test("broadcast sends JSON to all open clients", () => {
    const ws1 = makeMockWS();
    const ws2 = makeMockWS();
    hub.add(ws1);
    hub.add(ws2);

    hub.broadcast({ event: "test", value: 42 });

    const expected = JSON.stringify({ event: "test", value: 42 });
    expect(ws1.messages).toEqual([expected]);
    expect(ws2.messages).toEqual([expected]);
  });

  test("broadcast skips clients with readyState != 1", () => {
    const open = makeMockWS(1);
    const closed = makeMockWS(3);
    hub.add(open);
    hub.add(closed);

    hub.broadcast({ msg: "hello" });

    expect(open.messages).toHaveLength(1);
    expect(closed.messages).toHaveLength(0);
  });

  test("broadcast removes clients that throw on send", () => {
    const bad: WSClient = {
      readyState: 1,
      send() {
        throw new Error("connection reset");
      },
    };
    const good = makeMockWS();
    hub.add(bad);
    hub.add(good);

    hub.broadcast({ msg: "test" });

    // bad was removed, good received the message
    expect(hub.size).toBe(1);
    expect(good.messages).toHaveLength(1);
  });

  test("broadcast with no clients does nothing", () => {
    // Should not throw
    hub.broadcast({ msg: "nobody listening" });
    expect(hub.size).toBe(0);
  });

  test("multiple broadcasts accumulate on client", () => {
    const ws = makeMockWS();
    hub.add(ws);

    hub.broadcast({ a: 1 });
    hub.broadcast({ b: 2 });
    hub.broadcast({ c: 3 });

    expect(ws.messages).toHaveLength(3);
  });
});
