# Oraculo

## Philosophy

Oraculo is a Socratic guide and team orchestrator for quality product development.

**Core principles:**

- **Ask before doing** — No action without deep understanding
- **Orchestrate, never execute** — The main agent only delegates
- **Maximize parallelism** — Independent tasks run simultaneously
- **Quality over speed** — Every line of code follows the project's standards

Full document: [docs/philosophy.md](docs/philosophy.md)

## Design

Two operating modes:
- **Product Engineering** (with epic): Discover > Plan > Execute > Validate
- **Software Engineering** (story only): Plan > Execute > Validate

Full document: [docs/design.md](docs/design.md)

## Epic Phase

Transforms raw ideas into validated problem definitions through Socratic exploration. Uses Double Diamond, JTBD, TOC, and Assumption Mapping as internal conversation tools (their terminology does not appear in output artifacts).

Full document: [docs/epic/philosophy.md](docs/epic/philosophy.md)

## Story Phase

Transforms work items into well-defined, executable units. Same Socratic discipline as Epic, with less depth and more focus. Produces executable specifications with context, business rules, expected behavior, and acceptance criteria.

Full document: [docs/story/philosophy.md](docs/story/philosophy.md)

## CLI

The CLI is the **Trust Layer** — the deterministic, validated core that skills and agents depend on. Grounded in Design by Contract, Pit of Success, and Trusted Computing Base.

Full document: [docs/cli/philosophy.md](docs/cli/philosophy.md)

## Agents

The agent layer is Oraculo's execution workforce. The orchestrator decomposes work into a DAG and dispatches code agents (with TDD skill), research agents (for codebase analysis), and QA agents (with clean context). All agents work on the same branch — no worktrees. SQLite tracks operational state; the knowledge table persists lessons learned; committed markdowns capture outcomes.

Full document: [docs/agents/philosophy.md](docs/agents/philosophy.md)

Design index: [docs/agents/design.md](docs/agents/design.md)

## UI

The UI is Oraculo's **observation and control surface** — a browser-based dashboard that provides visibility into agents, tasks, DAG, approvals, and accumulated knowledge. It consumes data through the CLI Trust Layer (never bypasses it) and functions as Mission Control: comprehensive situational awareness with strategic intervention at approval gates.

Full document: [docs/ui/philosophy.md](docs/ui/philosophy.md)

Design index: [docs/ui/design.md](docs/ui/design.md)

## Project Structure

### `apps/`

Monorepo application code.

- **`apps/backend/`** — Go backend: CLI binary (`cmd/oraculo/`) and packages (`src/` — cli, db, domain, config, etc.)
- **`apps/dashboard/`** — Web UI (Next.js) — the observation and control surface

### `claude-kit/`

Distributable kit of Claude Code skills/commands. This folder is meant to be copied into the user's project under `.claude/` when they adopt Oraculo. During development, skills live here and are referenced from this repo.

```
apps/
├── backend/
│   ├── cmd/oraculo/        — CLI entrypoint
│   └── src/                — Go packages (cli, db, domain, config, etc.)
└── dashboard/              — Web UI (Next.js)
claude-kit/
└── skills/
    └── oraculo/
        ├── epic/
        │   ├── SKILL.md
        │   ├── phases/
        │   │   ├── 00-setup.md
        │   │   ├── 01-reframing.md
        │   │   ├── 02-divergence.md
        │   │   ├── 03-codebase-analysis.md
        │   │   ├── 04-convergence.md
        │   │   ├── 05-assumptions.md
        │   │   ├── 06-exit-gate.md
        │   │   ├── 07-artifact.md
        │   │   └── 08-approval.md
        │   └── references/
        │       ├── question-bank.md
        │       ├── frameworks.md
        │       └── artifact-templates.md
        └── story/
            ├── SKILL.md
            ├── phases/
            │   ├── 00-setup.md
            │   ├── 01-reframing.md
            │   ├── 02-assumptions.md
            │   ├── 03-exit-gate.md
            │   ├── 04-artifact.md
            │   └── 05-approval.md
            └── references/
                ├── question-bank.md
                └── artifact-templates.md
        ├── plan/
        │   ├── SKILL.md
        │   ├── phases/
        │   │   ├── 00-setup.md
        │   │   ├── 01-decomposition.md
        │   │   ├── 02-dependencies.md
        │   │   ├── 03-optimization.md
        │   │   └── 04-artifact.md
        │   └── references/
        │       └── decomposition-patterns.md
        ├── execute/
        │   ├── SKILL.md
        │   ├── phases/
        │   │   ├── 00-setup.md
        │   │   ├── 01-team-assembly.md
        │   │   ├── 02-monitoring.md
        │   │   └── 03-completion.md
        │   └── references/
        │       └── agent-dispatch.md
        └── validate/
            ├── SKILL.md
            ├── phases/
            │   ├── 00-setup.md
            │   ├── 01-qa-dispatch.md
            │   └── 02-verdict.md
            └── references/
                └── qa-criteria.md
```

Only `SKILL.md` files appear as slash commands. Reference files inside skill directories are internal — they don't pollute the command list.

## Package Manager

**ALWAYS use `bun`** — never npm, yarn, or pnpm. This applies to installing dependencies, running scripts, and any other package manager operation across the entire project.
