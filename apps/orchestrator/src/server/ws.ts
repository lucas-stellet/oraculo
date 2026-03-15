/**
 * WebSocket Hub — manages connections and broadcasts JSON messages.
 *
 * Accepts any object that satisfies the minimal WebSocket interface
 * (send + readyState) so it works with both Bun's native WS and test doubles.
 */

export interface WSClient {
  send(msg: string): void;
  readyState: number;
}

/** WebSocket readyState for OPEN */
const WS_OPEN = 1;

export class WebSocketHub {
  private clients = new Set<WSClient>();

  add(ws: WSClient): void {
    this.clients.add(ws);
  }

  remove(ws: WSClient): void {
    this.clients.delete(ws);
  }

  get size(): number {
    return this.clients.size;
  }

  broadcast(data: unknown): void {
    const msg = JSON.stringify(data);
    for (const ws of this.clients) {
      if (ws.readyState === WS_OPEN) {
        try {
          ws.send(msg);
        } catch {
          this.clients.delete(ws);
        }
      }
    }
  }
}
