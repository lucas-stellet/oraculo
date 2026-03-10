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

### Retry Prompt Additions

When constructing a retry prompt, use the standard Prompt Template above and add these sections after `## Rules`:

```
## Previous Failure Context
{failure_reason}

### What went wrong
{orchestrator_analysis_of_failure}

### Guidance
{suggested_alternative_approach}

## File State
The previous agent may have left partial files on disk. Before starting:
1. Check which files in your scope already exist
2. Read any existing files to understand what was attempted
3. Decide: fix in place or delete and start fresh
4. If starting fresh, delete the partial files before writing new ones
Do NOT assume the filesystem is clean.
```

### How to fill retry slots

- `{failure_reason}`: the `--reason` string from `oraculo tools task fail`
- `{orchestrator_analysis_of_failure}`: your analysis of root cause based on the failure signal
- `{suggested_alternative_approach}`: concrete guidance for an alternative approach that avoids the same failure

## Pre-Dispatch Hooks

Before dispatching each agent via the Agent tool, the orchestrator MUST call two hooks:

### 1. `oraculo hook task-started`
Broadcasts `task_started` WS event so the dashboard re-fetches and shows the task as in-progress.

```bash
oraculo hook task-started \
  --task-name "<task_name>" \
  --story-name "<story_name>" \
  --epic-name "<epic_name>"
```

### 2. `oraculo hook agent-start`
Registers the agent with its task association. The dashboard uses this to show "executing · {agent_name}" on the task row.

```bash
oraculo hook agent-start \
  --agent-name "<agent_name>" \
  --agent-type "<code|research>" \
  --task-name "<task_name>" \
  --story-name "<story_name>" \
  --epic-name "<epic_name>"
```

`--session-id` is optional. If provided, it should be the Claude Code session ID of the dispatched agent (not always known before dispatch — omit if unknown).

Both calls are best-effort: failure logs a warning but does not block dispatch.

> **Ordem obrigatória:** O frontend, ao receber `task_started`, chama `api.listTasks()` e espera ver o status `in_progress` no banco. `handleTaskStarted` no backend apenas faz broadcast — não altera o DB. Portanto, `oraculo task start` **deve ser chamado antes** de `oraculo hook task-started`, para que o banco já esteja atualizado quando o frontend consultar. A sequência completa antes de cada dispatch é:
>
> 1. `oraculo task start --story-id <id> --name <task_name>` — muda status para `in_progress` no DB
> 2. `oraculo hook task-started ...` — broadcast WS para o dashboard atualizar
> 3. `oraculo hook agent-start ...` — registra associação agente↔task
> 4. Dispatch via Agent tool

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

## Prompt Template

Use this template for every code agent prompt. Fill all slots before dispatching.

```
You are a code agent executing a task for the "{story_name}" story
in the "{epic_name}" epic.

## Task: {task_name}
{task_description}

## Acceptance Criteria
{acceptance_criteria}

## Design
{design_doc}

## Story Context
{story_requirements}

## Project Conventions
{claude_md_content}

## Additional Skills
{additional_skills_from_config}

## TDD
You MUST follow red-green-refactor:
1. Write the failing test first (red)
2. Write the minimum code to make it pass (green)
3. Refactor without breaking tests (refactor)
Never write implementation before a failing test exists.

## Working Directory
{working_directory}

## Rules
- Use `AskUserQuestion` for clarifications — never as plain text.
- Do NOT commit changes.
- Do NOT modify files outside this task's scope.
```

### How to fill the slots

- `{design_doc}`: run `oraculo tools design get <story> --epic <epic>` and paste the output
- `{story_requirements}`: run `oraculo tools story get <story> --epic <epic>` and paste the output
- `{claude_md_content}`: read CLAUDE.md from the project root
- `{additional_skills_from_config}`: read `skills.code_agent` from `.oraculo/config.json`; for each skill name, resolve it in `.claude/skills/` (local) or `~/.claude/skills/` (global) and include its content
- `{working_directory}`: absolute path to the project root
- `{task_name}`, `{task_description}`, `{acceptance_criteria}`: from the persisted task record. `{acceptance_criteria}` maps to the "Criterios de Aceite" section of the story artifact.
- `{story_requirements}`: the story artifact contains the following sections: Contexto, O que essa historia entrega, Regras de negocio, Comportamento esperado, Criterios de Aceite, Dependencias, Notas tecnicas. Include all sections so the agent has full context.
