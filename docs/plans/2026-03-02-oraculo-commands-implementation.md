# Oraculo Command Skills Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the full first-pass Oraculo command skill kit in `claude-kit/skills/oraculo` for `epic`, `story`, `plan`, `execute`, and `validate`, faithfully following the command design and philosophy docs and the existing CLI trust-layer contracts.

**Architecture:** The source of truth is a distributable Markdown skill kit under `claude-kit/skills/oraculo`. Each command is composed of a thin `SKILL.md`, phase files loaded just in time, and narrowly scoped reference files. Skills call only existing CLI contracts for session, phase, approval, and artifact persistence.

**Tech Stack:** Markdown skills, Claude Code slash-command frontmatter, Go CLI install flow, Cobra-based `oraculo tools` commands, Go tests.

---

## Execution Order

1. Persist the approved design and implementation docs.
2. Add an install test proving `oraculo install` copies a local `claude-kit/skills/oraculo` tree.
3. Scaffold the full command directory tree.
4. Implement shared `SKILL.md` conventions.
5. Implement the `epic` and `story` command sets.
6. Implement the `plan`, `execute`, and `validate` command sets.
7. Update `CLAUDE.md` to match the shipped kit.
8. Run structural verification and targeted tests.

## Required Deliverables

Create this checked-in tree under `claude-kit/skills/oraculo`:

- `epic/SKILL.md`
- `epic/phases/00-setup.md`
- `epic/phases/01-reframing.md`
- `epic/phases/02-divergence.md`
- `epic/phases/03-codebase-analysis.md`
- `epic/phases/04-convergence.md`
- `epic/phases/05-assumptions.md`
- `epic/phases/06-exit-gate.md`
- `epic/phases/07-artifact.md`
- `epic/phases/08-approval.md`
- `epic/references/question-bank.md`
- `epic/references/frameworks.md`
- `epic/references/artifact-templates.md`
- `story/SKILL.md`
- `story/phases/00-setup.md`
- `story/phases/01-reframing.md`
- `story/phases/02-assumptions.md`
- `story/phases/03-exit-gate.md`
- `story/phases/04-artifact.md`
- `story/phases/05-approval.md`
- `story/references/question-bank.md`
- `story/references/artifact-templates.md`
- `plan/SKILL.md`
- `plan/phases/00-setup.md`
- `plan/phases/01-decomposition.md`
- `plan/phases/02-dependencies.md`
- `plan/phases/03-optimization.md`
- `plan/phases/04-artifact.md`
- `plan/references/decomposition-patterns.md`
- `execute/SKILL.md`
- `execute/phases/00-setup.md`
- `execute/phases/01-team-assembly.md`
- `execute/phases/02-monitoring.md`
- `execute/phases/03-completion.md`
- `execute/references/agent-dispatch.md`
- `validate/SKILL.md`
- `validate/phases/00-setup.md`
- `validate/phases/01-qa-dispatch.md`
- `validate/phases/02-verdict.md`
- `validate/references/qa-criteria.md`

## Constraints

- English only
- `claude-kit` is the only checked-in source of truth
- no new CLI APIs
- no checked-in `.claude` runtime copy
- `execute` omits `allowed-tools`
- `plan` keeps optional `execution-plan` approval inside `04-artifact.md`

## Verification

Run:

```bash
find claude-kit/skills/oraculo -maxdepth 4 -type f | sort
rg -n "<phase-gate|<persona|<context-boundaries|<anti-patterns" claude-kit/skills/oraculo
go test ./src/cli -run Install
go test ./...
```

Verify:
- every command has the expected file set
- every phase file ends with one `<phase-gate>`
- referenced sections and files exist
- `oraculo install` copies the skill kit successfully
