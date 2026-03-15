/**
 * Promise-based approval bridge for coordinating approval gates
 * between the orchestrator execution loop and external decision makers.
 */
export function createApprovalBridge() {
  const pending = new Map<
    string,
    { resolve: (verdict: string) => void; reject: (err: Error) => void }
  >();

  return {
    /**
     * Waits for a decision on the given approval ID.
     * Rejects with a timeout error if no decision arrives within the deadline.
     */
    wait(id: string, timeout = 3600000): Promise<string> {
      return new Promise((resolve, reject) => {
        const timer = setTimeout(() => {
          pending.delete(id);
          reject(new Error(`Approval ${id} timed out`));
        }, timeout);

        pending.set(id, {
          resolve: (verdict) => {
            clearTimeout(timer);
            resolve(verdict);
          },
          reject: (err) => {
            clearTimeout(timer);
            reject(err);
          },
        });
      });
    },

    /**
     * Records a decision for a pending approval.
     * Returns true if the approval was pending and has been resolved, false otherwise.
     */
    decide(id: string, verdict: string): boolean {
      const entry = pending.get(id);
      if (entry) {
        entry.resolve(verdict);
        pending.delete(id);
        return true;
      }
      return false;
    },

    /**
     * Checks whether an approval is currently awaiting a decision.
     */
    hasPending(id: string): boolean {
      return pending.has(id);
    },
  };
}
