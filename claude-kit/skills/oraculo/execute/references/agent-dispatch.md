# Execute Agent Dispatch

## Execution Readiness

Only dispatch work when:
- the plan is persisted
- dependencies are satisfied
- any required approvals are satisfied

## Agent Roles

- code agent: implements tasks and follows TDD
- research agent: gathers codebase evidence that unblocks execution

QA is not part of execute. QA belongs to `/oraculo:validate`.

## Parallelism Rules

Parallelize only when:
- dependencies are satisfied
- file scope does not conflict
- the tasks do not need tight coordination

## Retry Discipline

When a task fails, prefer a fresh agent with explicit failure context instead of reusing stale reasoning.

## No Direct Agent Coordination

Agents do not coordinate directly with each other. Oraculo coordinates through the DAG and task state.

## QA Handoff

Execution completion hands off:
- task summaries
- relevant diffs
- test results
- story or epic context

It does not perform QA itself.

## Agent Interactive Questions

When an agent needs clarification before or during a task, it must use the
`AskUserQuestion` tool — never plain text questions.

The orchestrator must include this in every agent prompt:
> "If you have questions before or during the task, use the `AskUserQuestion`
> tool. Never ask questions as plain text."

Constraints (same as orchestrator):
- 2–4 options per question
- header max 12 characters
- `multiSelect: true` when multiple answers can apply simultaneously
