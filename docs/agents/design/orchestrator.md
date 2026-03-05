# Agents Design — Orchestrator

## 1. Role

The orchestrator is the only agent that sees the full picture. It plans, delegates, and coordinates — it never executes. It never writes code, never runs tests, never touches the filesystem. Its context window is reserved exclusively for strategic reasoning.

This is not a guideline — it is an architectural boundary. Mixing planning and execution in the same context measurably degrades both. The orchestrator operates at the level of tasks, dependencies, and agent assignments. Tactical details (syntax errors, test output, compilation logs) are handled entirely by subordinate agents and the CLI.

## 2. DAG Creation

During the Plan phase, the orchestrator decomposes requirements into a DAG:

- **Tasks are nodes.** Each node has a type (code, qa, research, documentation), a description, and a skill assignment.
- **Dependencies are edges.** An edge from A to B means B cannot start until A completes.
- **The graph is acyclic.** The CLI validates this before persisting.

The orchestrator emits the DAG as a structured plan. The CLI validates it (acyclicity, dependency integrity, no self-references) and persists it to SQLite using the schema defined in [`docs/cli/design.md`](../../cli/design.md) §4.3.

The DAG is not a static artifact. When an agent encounters an obstacle or QA rejects a task, the orchestrator mutates the DAG — pruning invalid branches, adding new tasks, adjusting dependencies — without regenerating the entire plan. The CLI validates every mutation before committing.

## 3. Dispatch

The orchestrator evaluates the DAG and dispatches tasks that are ready to execute:

**Readiness rule:** A task is ready when all its dependencies are in a terminal state (completed or failed-and-handled).

**Parallelism:** All ready tasks are dispatched simultaneously when possible. The orchestrator maximizes parallel execution within the constraints of the DAG and file-level coordination.

**File coordination:** Since all agents work on the same branch in the same directory, the orchestrator must ensure no two agents modify the same files simultaneously. This is achieved through the DAG's dependency structure — tasks that touch overlapping files are connected by edges, forcing sequential execution. When the orchestrator creates the DAG, it encodes these file-level constraints as dependencies.

**Sequential fallback:** When true parallelism is not possible (overlapping files, shared state), tasks run sequentially. The orchestrator prefers parallel execution but never at the cost of correctness.

## 4. Skill Assignment

The orchestrator selects skills for each agent based on the nature of the task. This is a first-class responsibility — the right skill determines the agent's entire workflow.

| Task Type | Typical Skills | Rationale |
|-----------|---------------|-----------|
| Code implementation | TDD | Red-green-refactor enforces correctness |
| UI/frontend work | TDD + frontend-design | Design quality + test discipline |
| E2E validation | playwright | Browser-based testing |
| Bug fix | TDD + systematic-debugging | Diagnosis before implementation |
| Refactoring | TDD | Existing tests as the safety net |

The skill is the containment mechanism. It defines:
- **Workflow:** What steps the agent must follow (e.g., red-green-refactor)
- **Constraints:** What the agent must not do (e.g., skip the red phase)
- **Quality gates:** What must be true before the agent can report completion

The orchestrator does not invent workflows. It selects from the available skill library based on task characteristics.

## 5. QA Throughput as Governing Constraint

The orchestrator treats QA capacity as the system's bottleneck — the constraint that sets the pace for the entire system (Theory of Constraints).

**The problem:** If the orchestrator dispatches code tasks as fast as the DAG allows, unvalidated work accumulates. Unvalidated work carries risk: context drift, cascading assumptions, exponential rework when QA eventually catches issues.

**The solution:** The orchestrator limits dispatch rate so that code is produced only as fast as QA can validate it. When the QA agent is reviewing task N, the orchestrator may hold task N+2 even if its dependencies are satisfied, to prevent a growing queue of unvalidated work.

**Practical heuristic:** Keep no more than one unvalidated task ahead of QA at any time. If QA is busy, the orchestrator can dispatch non-code tasks (documentation, research) that don't require QA validation.

## 6. Handling Failures

When a code agent fails or QA rejects a task:

1. **QA rejection:** The orchestrator spawns a **new** code agent with QA's structured findings. The new agent gets a fresh context — no memory of the previous attempt, no inherited assumptions. This prevents agents from defending previous mistakes.

2. **Agent failure:** If an agent fails to complete its task (error, timeout, impossible task), the orchestrator marks the task as failed with a reason and evaluates whether to retry with adjusted context or escalate.

3. **Circuit breaker:** After N failed QA cycles on the same task (configurable, default 3), the orchestrator stops retrying and submits a `qa-escalation` approval request to the dashboard:

   ```
   oraculo tools approval request --type qa-escalation
   ```

   The orchestrator transitions to `awaiting_approval` state. The dashboard displays the full diagnostic context (all QA findings, all attempt summaries) for human review. The human's verdict drives the next step:
   - **`approved`** — Accept the current implementation as-is and advance the task
   - **`rejected`** — Abort the story; mark the task as permanently failed
   - **`needs_revision`** — Spawn a new code agent with the human's comments as additional context and reset the circuit breaker

4. **DAG mutation:** When a failure invalidates downstream tasks, the orchestrator prunes the affected branch and may add new tasks to take a different approach. The CLI validates every mutation.

## 7. Awaiting Approval State

When an approval gate is open, the orchestrator enters `awaiting_approval` state. This state has specific behavioral rules to keep the system productive without advancing a blocked phase:

**What the orchestrator does not do:**
- Does not dispatch any tasks that depend — directly or transitively — on the pending approval
- Does not advance the phase that the approval gate is guarding

**What the orchestrator continues to do:**
- Dispatches independent tasks whose dependencies are already satisfied and that do not depend on the approval outcome
- For operational gates: monitors approval status by polling `oraculo tools approval status`
- For document reviews: monitors review status by polling `oraculo tools review list <version-id> --type <epic|story>`

**Returning to active state:**
Once a verdict is received (via `oraculo tools approval status` for operational gates, or `oraculo tools review list` for document reviews), the orchestrator reads the verdict, applies the appropriate outcome (advance or abort for document reviews; advance, abort, or retry with comments for operational gates), and transitions back to active state. Dependent tasks are re-evaluated against the updated DAG.
