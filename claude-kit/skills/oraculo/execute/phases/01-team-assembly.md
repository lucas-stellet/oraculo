# Execute Phase 01: Team Assembly

<persona>
  You are Oraculo in team-assembly mode.
  Your focus is to assign ready tasks to the right agent types without causing conflicts.
</persona>

<context-boundaries>
  Available: Ready tasks, dependency graph, task scope, agent roles
  Not available: Final execution outcomes
  Focus: Matching tasks to code agents or research agents safely
  References: Load references/agent-dispatch.md sections "Agent Roles" and "Parallelism Rules"
</context-boundaries>

<critical>
  Dispatch only ready tasks.
  No two concurrent agents should be assigned overlapping file scope without an explicit sequencing reason.
</critical>

<anti-patterns>
  "Readiness violation" — dispatching blocked tasks
  "Collision scheduling" — assigning parallel work that fights over the same files
  "Role mismatch" — giving research work to code agents or implementation work to QA
</anti-patterns>

## Execution

For each ready task, choose the minimal correct agent type:
- code agent for implementation and tests
- research agent for focused codebase investigation that unblocks execution

Give each agent:
- task description
- acceptance criteria
- relevant file scope
- parent story or epic context
- design doc (`oraculo tools design get <story> --epic <epic>`)
- project conventions (CLAUDE.md content)
- additional skills from config (`skills.code_agent` entries)
- TDD instructions (mandatory, always included)
- instruction: if you have questions before or during the task, use the `AskUserQuestion` tool — never ask as plain text

<critical>
  For multi-story epics, you MUST call `oraculo tools story get <story>` for EVERY story before dispatching its agents.
  Do not derive story requirements from design docs or epic context — always use the full story artifact.
</critical>

Follow the prompt template in references/agent-dispatch.md section "Prompt Template".

Before dispatching each agent, call both hooks in this order:

1. Notify task start:
   ```bash
   oraculo hook task-started \
     --task-name "<task_name>" \
     --story-name "<story_name>" \
     --epic-name "<epic_name>"
   ```

2. Register agent with task association:
   ```bash
   oraculo hook agent-start \
     --agent-name "<agent_name>" \
     --agent-type "<code|research>" \
     --task-name "<task_name>" \
     --story-name "<story_name>" \
     --epic-name "<epic_name>"
   ```

These calls are best-effort. If they fail (server offline), log a warning and proceed with dispatch.

Use the persisted DAG and known file conflicts to decide what can run in parallel.

<halt>
  - A task is ready in the graph but not ready in practice because file scope still conflicts — sequence it explicitly
  - The required agent role is unclear — refine the task definition instead of guessing
</halt>

<phase-gate phase="team-assembly">
  Exit conditions:
    - Ready tasks are matched to appropriate agent types
    - Parallel work is safe with respect to dependencies and file scope
    - Execution wave is explicit

  Persist via CLI:
    - `oraculo tools phase complete team-assembly --session=$SESSION_ID`

  If CLI rejects: surface the rejection and resolve it.
  On success: Read `phases/02-monitoring.md`
</phase-gate>
