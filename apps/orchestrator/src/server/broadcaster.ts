/**
 * SSE Broadcaster — ring buffer of log entries with fan-out to subscribers.
 *
 * Matches the Go Broadcaster in applog/broadcaster.go:
 * - Fixed-size ring buffer (200 entries)
 * - subscribe() returns an unsubscribe function
 * - push() serialises the entry and fans out to all active subscribers
 * - recent() returns a snapshot of the ring buffer for replay
 */

const RING_SIZE = 200;

export class Broadcaster {
  private buffer: string[] = [];
  private maxSize = RING_SIZE;
  private subscribers = new Set<(data: string) => void>();

  /**
   * Serialises the entry to JSON, appends to the ring buffer,
   * and delivers to all current subscribers.
   */
  push(entry: unknown): void {
    const json = JSON.stringify(entry);
    this.buffer.push(json);
    if (this.buffer.length > this.maxSize) {
      this.buffer.shift();
    }
    for (const sub of this.subscribers) {
      try {
        sub(json);
      } catch {
        this.subscribers.delete(sub);
      }
    }
  }

  /**
   * Registers a subscriber callback. Returns an unsubscribe function.
   */
  subscribe(cb: (data: string) => void): () => void {
    this.subscribers.add(cb);
    return () => {
      this.subscribers.delete(cb);
    };
  }

  /**
   * Returns a copy of the current ring buffer contents (oldest first).
   */
  recent(): string[] {
    return [...this.buffer];
  }

  get subscriberCount(): number {
    return this.subscribers.size;
  }
}
