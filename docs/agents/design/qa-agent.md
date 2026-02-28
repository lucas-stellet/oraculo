# Agents Design — QA Agent

## 1. Role

The QA agent is the immune system. It validates code agent output with fresh eyes, no shared context, and no incentive to agree. Its independence is the single most important property of the validation pipeline.

A QA agent that merely confirms "looks good" is worthless. Effective validation requires an agent that checks functional correctness, enforces standards, probes edge cases, and evaluates test quality — all without access to the code agent's reasoning about why things were done a certain way.

## 1.1 Two Levels of Validation

QA operates at two distinct levels:

**Task-level:** Each completed task is reviewed independently. The QA agent receives the diff, specs, and test results for that specific task. Verdicts are recorded with `task_id` in the `validations` table.

**Story-level:** Once all tasks in a story pass individual QA, a final integrated review checks cross-task coherence, story requirements as a whole, and overall quality. This verdict is recorded with `task_id = NULL` in the `validations` table and serves as the final gate before generating the story's markdown summary and extracting lessons learned into the knowledge table.

## 2. Clean Context

The QA agent operates with a completely clean context window. It receives exactly three inputs:

1. **The diff** — what changed, nothing more
2. **The specs** — the task description, acceptance criteria, and relevant story/epic requirements
3. **Test results** — output from running the existing test suite

It does **not** receive:
- The code agent's conversational history
- The code agent's intermediate reasoning
- Previous QA verdicts on this task (each QA review is independent)
- Debug logs or compilation traces from the implementation process

This isolation breaks the sycophancy cycle. When a QA agent shares context with the code agent, it tends to agree with the code agent's decisions — even flawed ones. Clean context forces the QA agent to evaluate the work on its merits.

## 3. What QA Checks

The QA agent evaluates the implementation against multiple dimensions:

| Dimension | What the QA agent asks |
|-----------|----------------------|
| **Functional correctness** | Does the implementation satisfy the acceptance criteria? Does it handle the specified inputs and produce the expected outputs? |
| **Standards compliance** | Does the code follow the project's conventions (from CLAUDE.md)? Naming, patterns, structure? |
| **Edge cases** | Are boundary conditions handled? What happens with empty inputs, large inputs, concurrent access? |
| **Test quality** | Do the tests actually verify the behavior? Are assertions meaningful or trivially true? Do tests cover the acceptance criteria? |
| **Scope** | Did the agent stay within the task's boundaries? Did it modify files it shouldn't have? Did it add unrequested features? |

## 4. QA Never Fixes

This is a hard rule. The QA agent produces structured findings — it never modifies code.

When QA finds an issue, it reports:
- **What's wrong:** Clear description of the problem
- **Where:** File and location
- **Why it matters:** Which acceptance criterion or standard is violated
- **Severity:** Whether it's a blocker, a significant issue, or a minor concern

The orchestrator receives these findings and spawns a new code agent to address them. The QA agent's job is done once it delivers its report.

**Why not let QA fix things?** If QA fixes code, it can no longer independently validate the fix. The agent that writes code and the agent that validates code must be different agents — otherwise the validation is compromised.

## 5. Rejection Flow

When QA rejects a task:

```
Code Agent completes task
    |
    v
QA Agent reviews (clean context: diff + specs + test results)
    |
    v
QA produces structured findings
    |
    v
Orchestrator receives findings
    |
    v
Orchestrator spawns NEW Code Agent with:
  - Original task description
  - QA's findings as additional context
  - Fresh context (no memory of previous attempt)
    |
    v
New Code Agent implements fix (TDD skill)
    |
    v
QA Agent reviews again (new clean context)
    |
    v
Repeat until approved or circuit breaker trips
```

Each iteration uses a **new** code agent with fresh context. This prevents defensive behavior where an agent tries to justify its previous work rather than reconsidering from first principles.

## 6. Circuit Breaker

A circuit breaker prevents infinite rejection loops:

- **Threshold:** N failed QA cycles on the same task (configurable, default 3)
- **On trip:** The orchestrator stops spawning new code agents and submits a `qa-escalation` approval request to the dashboard via `oraculo tools approval request --type qa-escalation`. The orchestrator enters `awaiting_approval` state.
- **Context preserved:** All QA findings and attempt summaries are available in the dashboard for human review
- **Resolution:** The human issues a verdict — `approved` (accept as-is), `rejected` (abort the story), or `needs_revision` (retry with the human's guidance)

The circuit breaker exists because some tasks may be genuinely beyond the agents' capability — ambiguous requirements, conflicting constraints, or problems that require human judgment. Escalating early is better than burning tokens on endless retries.

## 7. Skill Loading

The QA agent can receive skills from the orchestrator, just like code agents:

- **playwright:** For E2E validation — the QA agent can run browser-based tests to verify UI behavior
- **Project-specific skills:** Custom validation rules defined by the team

The default QA workflow (diff review + standards check) does not require a special skill. Skills extend QA capabilities for specific task types.
