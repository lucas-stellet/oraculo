# Oraculo Command Skills Design

## Summary

Implement a complete first-pass Oraculo skill kit under `claude-kit/skills/oraculo` as the single checked-in source of truth for distributable Claude Code commands.

The kit will contain five commands:
- `oraculo:epic`
- `oraculo:story`
- `oraculo:plan`
- `oraculo:execute`
- `oraculo:validate`

Each command will follow the command system design in `docs/commands/design.md`:
- thin `SKILL.md` entrypoint
- phase files loaded just in time from `phases/`
- supporting material loaded on demand from `references/`
- XML guardrails embedded in each phase file

## Decisions

### Source of truth

The checked-in source of truth is `claude-kit/skills/oraculo`. Runtime copies in `.claude/skills/oraculo` are created by `oraculo install` and are out of scope for manual editing.

### Scope

This first package implements all five commands, not just discovery commands.

### Language

The first shipped version is English only.

### CLI contracts

Skills must rely only on existing CLI commands:
- `oraculo tools session status|init|state|close`
- `oraculo tools phase complete`
- `oraculo tools approval list|request|status`
- `oraculo tools epic save|get|init`
- `oraculo tools story save|get|init`
- `oraculo tools task init|start|complete|fail|list|get`

No new CLI APIs are introduced for this work.

### Frontmatter

All `SKILL.md` files include:
- `name`
- `description`
- `argument-hint`
- `disable-model-invocation: true`

`epic`, `story`, `plan`, and `validate` also include:
- `allowed-tools: Read, Bash, Task, AskUserQuestion`

`execute` intentionally omits `allowed-tools` because the repository does not define a stable frontmatter contract for team orchestration tools.

## Command Structure

### Epic

`epic` is the full discovery command. It uses:
- 9 phase files
- question, framework, and artifact references
- approval handling after artifact generation

### Story

`story` is the compact discovery command. It supports:
- standalone story mode
- epic-derived story mode
- escalation guidance back to `oraculo:epic`

### Plan

`plan` converts approved requirements into a DAG-oriented execution plan using existing task CLI contracts. Optional `execution-plan` approval remains part of the artifact phase rather than a separate phase file.

### Execute

`execute` is an orchestration command. It describes team assembly, monitoring, and QA handoff without claiming direct code execution.

### Validate

`validate` dispatches clean-context QA, processes verdicts, and recommends the correct return path when work must go back to execution, planning, or discovery.

## Acceptance Criteria

- The repository contains the full command tree under `claude-kit/skills/oraculo`.
- Every command has `SKILL.md`, `phases/`, and `references/`.
- Every phase file contains the required XML guardrails and ends with exactly one `<phase-gate>`.
- `oraculo install` can copy the skill kit into `.claude/skills/oraculo`.
- `CLAUDE.md` documents the shipped command kit accurately.
