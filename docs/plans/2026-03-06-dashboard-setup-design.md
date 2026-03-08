# Dashboard Frontend Setup + Landing Page

## Decision

Initialize the Next.js frontend application in `apps/dashboard/` with the design system foundations and a functional Landing page (epic selection + creation).

## Stack

| Layer | Technology |
|---|---|
| Runtime | Bun |
| Framework | Next.js 15 (App Router) |
| Components | shadcn/ui + Radix primitives |
| Styling | Tailwind CSS v4 |
| Typography | Space Grotesk + IBM Plex Sans + IBM Plex Mono |
| Theme | Light/dark, Mission Control direction (Slate base, Blue accent) |

## Location

`apps/dashboard/` — static export via `output: 'export'` in next.config.ts. The Go binary will embed the output directory via `embed.FS`.

## Directory Structure

```
apps/dashboard/
├── src/
│   ├── app/
│   │   ├── layout.tsx          # Shell: sidebar + main area
│   │   ├── page.tsx            # Landing (epic selection)
│   │   └── globals.css         # Tailwind + tokens + fonts
│   ├── components/
│   │   ├── ui/                 # shadcn/ui components
│   │   ├── sidebar.tsx         # Nav sidebar (muted when no epic selected)
│   │   ├── epic-card.tsx       # Epic card with phase badge + progress
│   │   └── create-epic-dialog.tsx  # Creation modal
│   ├── lib/
│   │   └── utils.ts            # cn() helper (shadcn standard)
│   └── hooks/                  # Custom hooks (empty for now)
├── public/
├── next.config.ts              # output: 'export', basePath config
├── tailwind.config.ts
├── tsconfig.json
├── package.json
└── components.json             # shadcn/ui config
```

## Landing Page

Three vertical sections:

1. **Hero** — "Oraculo" title, brief subtitle explaining the system (Mission Control for AI-agent-driven development). Clean, `Space Grotesk 700` typography.

2. **Epic Grid** — Responsive cards (1-3 columns). Each card shows:
   - Epic name (heading)
   - Phase badge (Discover / Plan / Execute / Validate)
   - Progress bar (completed tasks / total)
   - Stats row: stories, tasks, completion %
   - "Open" button

3. **New Epic** — Button in grid header or special `+` card. Opens a Dialog (Radix/shadcn) with:
   - Name field
   - Brief description field
   - Cancel / Create buttons

## Data

Hardcoded mock data matching the `GET /api/epics` response shape. Types defined to match future API integration.

## Design System (partial)

Only what the Landing needs:
- Color tokens: Slate base, Blue accent, semantic status colors (blue=in_progress, green=completed, red=failed, amber=pending)
- Typography: 3 font families, tokens (display-lg, heading-md, body-md, label-sm, mono-sm)
- shadcn components: Button, Card, Badge, Dialog, Input, Textarea, Progress
- Dark mode via `class` strategy

## Out of Scope

- Routes for other screens (Stories, DAG, Agent Monitor, etc.)
- WebSocket or real API integration
- Functional sidebar navigation (visible but muted/disabled)
- Sidebar responsive collapse
- Go embed.FS integration (separate task)
