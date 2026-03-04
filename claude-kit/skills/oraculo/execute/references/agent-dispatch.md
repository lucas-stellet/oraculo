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
- `{task_name}`, `{task_description}`, `{acceptance_criteria}`: from the persisted task record
