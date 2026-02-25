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
- **Full** (with epic): Discover > Plan > Execute > Validate
- **Reduced** (story only): Plan > Execute > Validate

Full document: [docs/design.md](docs/design.md)

## Epic Phase

Transforms raw ideas into validated problem definitions through Socratic exploration. Uses Double Diamond, JTBD, OST, and Assumption Mapping frameworks.

Full document: [docs/epic/philosophy.md](docs/epic/philosophy.md)

## Story Phase

Transforms work items into well-defined, executable units. Same Socratic discipline as Epic, with less depth and more focus.

Full document: [docs/story/philosophy.md](docs/story/philosophy.md)

## CLI

The CLI is the **Trust Layer** — the deterministic, validated core that skills and agents depend on. Grounded in Design by Contract, Pit of Success, and Trusted Computing Base.

Full document: [docs/cli/philosophy.md](docs/cli/philosophy.md)

## Project Structure

### `claude-kit/`

Distributable kit of Claude Code skills/commands. This folder is meant to be copied into the user's project under `.claude/` when they adopt Oraculo. During development, skills live here and are referenced from this repo.

```
claude-kit/
└── skills/
    └── oraculo/
        ├── epic/
        │   ├── SKILL.md                       # /oraculo:epic skill
        │   └── references/
        │       ├── question-bank.md           # Question templates by phase/level
        │       ├── frameworks.md              # Framework descriptions
        │       ├── artifact-templates.md      # Intermediate artifact formats (OST, Codebase Impact)
        │       └── requirements-template.md   # Primary output: Requirements document template
        └── story/
            ├── SKILL.md                       # /oraculo:story skill
            └── references/
                ├── question-bank.md           # Question templates by phase/level
                ├── frameworks.md              # Framework descriptions (simplified)
                └── story-template.md          # Primary output: Story document template
```

Only `SKILL.md` files appear as slash commands. Reference files inside skill directories are internal — they don't pollute the command list.
