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

## Dashboard Static Assets and SPA Routing

The dashboard is a Next.js app with `output: "export"`, embedded in the Go binary via `embed.FS`. Dynamic routes are pre-rendered with the `__placeholder__` param (e.g. `/epics/__placeholder__/approvals.html`).

### How the Go server routes requests (`apps/backend/src/server/server.go`)

The SPA handler follows three steps in order:

1. **Exact file** — if `fs.Stat(assets, fsPath)` finds the file, serve it directly.
2. **Placeholder substitution** — if the file doesn't exist, replace dynamic segments in the URL with `__placeholder__` (via `withPlaceholders`) and serve the resulting file if it exists.
3. **HTML/TXT shell** — fallback to `spaShell`, which maps the route to the `.html` shell (direct navigation) or `.txt` payload (RSC requests with `?_rsc=`).

### Why step 2 is necessary

The Next.js app router fetches RSC segment files by direct path — without the `?_rsc=` query param. Examples:

```
GET /epics/gastos-pessoais/approvals/__next.epics.$d$id.approvals.__PAGE__.txt
GET /epics/gastos-pessoais/approvals/__next._full.txt
```

These files live under `epics/__placeholder__/approvals/`, not `epics/gastos-pessoais/approvals/`. Without substitution, the handler falls through to `spaShell` and returns HTML — the router receives HTML where it expects RSC data and navigates to the wrong page.

### When adding new dynamic routes to the dashboard

- Add the dynamic segment to `withPlaceholders` in `server.go`.
- Also update `spaShell` with the corresponding shell mapping.
- Every new dynamic page must use `generateStaticParams()` returning `[{ id: "__placeholder__" }]` (or the equivalent param name).
- Never add client-side redirect logic based on SSR params — always use `usePathname()` to read the real ID from the URL.

## Package Manager

**ALWAYS use `bun`** — never npm, yarn, or pnpm. This applies to installing dependencies, running scripts, and any other package manager operation across the entire project.
