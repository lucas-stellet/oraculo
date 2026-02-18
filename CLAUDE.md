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

**5-phase operating model:** Discover > Plan > Execute > Validate > Deliver

Full document: [docs/design.md](docs/design.md)

## Project Structure

### `claude-kit/`

Distributable kit of Claude Code skills/commands. This folder is meant to be copied into the user's project under `.claude/` when they adopt Oraculo. During development, skills live here and are referenced from this repo.

```
claude-kit/
└── skills/
    └── oraculo-brainstorm/
        ├── SKILL.md                       # /oraculo:brainstorm skill
        └── references/
            ├── question-bank.md           # Question templates by phase/level
            ├── frameworks.md              # Framework descriptions
            └── artifact-templates.md      # Output artifact formats
```

Only `SKILL.md` appears as a slash command. Reference files inside the skill directory are internal — they don't pollute the command list.
