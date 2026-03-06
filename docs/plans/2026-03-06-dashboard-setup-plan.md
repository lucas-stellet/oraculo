# Dashboard Frontend Setup — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Initialize the Next.js frontend in `apps/dashboard/` with design system foundations and a functional Landing page (epic selection + creation modal).

**Architecture:** Next.js 15 App Router with static export (`output: 'export'`). shadcn/ui provides composable components on top of Radix primitives. Tailwind v4 handles styling with custom tokens matching the Mission Control design direction. Mock data mirrors the Go `domain.Epic` shape for seamless future API integration.

**Tech Stack:** Bun, Next.js 15, shadcn/ui, Tailwind CSS v4, Radix UI, TypeScript

---

### Task 1: Scaffold Next.js project with Bun

**Files:**
- Create: `apps/dashboard/package.json` (via CLI)
- Create: `apps/dashboard/next.config.ts` (via CLI, then modify)
- Create: `apps/dashboard/tsconfig.json` (via CLI)
- Create: `apps/dashboard/src/app/layout.tsx` (via CLI)
- Create: `apps/dashboard/src/app/page.tsx` (via CLI)
- Create: `apps/dashboard/src/app/globals.css` (via CLI)

**Step 1: Create Next.js app with Bun**

```bash
cd apps/dashboard && bunx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir --no-import-alias --turbopack
```

When prompted, accept defaults. This scaffolds a Next.js 15 project with App Router, TypeScript, Tailwind CSS v4, and ESLint inside `apps/dashboard/`.

**Step 2: Configure static export**

Edit `apps/dashboard/next.config.ts`:

```ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
};

export default nextConfig;
```

**Step 3: Verify the app builds**

```bash
cd apps/dashboard && bun run build
```

Expected: Build succeeds and creates `apps/dashboard/out/` directory with static HTML.

**Step 4: Verify dev server runs**

```bash
cd apps/dashboard && bun run dev
```

Expected: Dev server starts on localhost:3000. Visit and see the default Next.js page.

**Step 5: Commit**

```bash
git add apps/dashboard
git commit -m "feat(dashboard): scaffold Next.js 15 app with static export"
```

---

### Task 2: Initialize shadcn/ui

**Files:**
- Modify: `apps/dashboard/src/app/globals.css` (shadcn tokens)
- Create: `apps/dashboard/components.json` (via CLI)
- Create: `apps/dashboard/src/lib/utils.ts` (via CLI)
- Create: `apps/dashboard/src/components/ui/` (via CLI)

**Step 1: Run shadcn init**

```bash
cd apps/dashboard && bunx shadcn@latest init -d
```

This configures shadcn/ui for the project: creates `components.json`, `src/lib/utils.ts` (with `cn()` helper), and updates `globals.css` with CSS variable tokens.

**Step 2: Add required shadcn components**

```bash
cd apps/dashboard && bunx shadcn@latest add button card badge dialog input textarea progress
```

This installs the 7 components we need for the Landing page into `src/components/ui/`.

**Step 3: Verify build still works**

```bash
cd apps/dashboard && bun run build
```

Expected: Build succeeds.

**Step 4: Commit**

```bash
git add apps/dashboard
git commit -m "feat(dashboard): initialize shadcn/ui with landing page components"
```

---

### Task 3: Configure design system tokens and typography

**Files:**
- Modify: `apps/dashboard/src/app/globals.css`
- Modify: `apps/dashboard/src/app/layout.tsx`

**Step 1: Install Google Fonts via `next/font`**

Edit `apps/dashboard/src/app/layout.tsx`:

```tsx
import type { Metadata } from "next";
import { Space_Grotesk, IBM_Plex_Sans, IBM_Plex_Mono } from "next/font/google";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  variable: "--font-display",
  subsets: ["latin"],
  weight: ["600", "700"],
});

const ibmPlexSans = IBM_Plex_Sans({
  variable: "--font-sans",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

const ibmPlexMono = IBM_Plex_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
  weight: ["500"],
});

export const metadata: Metadata = {
  title: "Oraculo",
  description: "Mission Control for AI-agent-driven development",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${spaceGrotesk.variable} ${ibmPlexSans.variable} ${ibmPlexMono.variable} font-sans antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
```

**Step 2: Configure design tokens in globals.css**

Replace the shadcn default theme variables in `apps/dashboard/src/app/globals.css` with Mission Control tokens. Keep the shadcn `@theme inline` block but update the color values to use Slate base + Blue accent:

- `--background`: slate-950 (dark) / white (light)
- `--foreground`: slate-50 (dark) / slate-950 (light)
- `--card`: slate-900 (dark) / white (light)
- `--primary`: blue-600 (dark) / blue-600 (light)
- `--muted`: slate-800 (dark) / slate-100 (light)
- Semantic status colors as utility classes:
  - `--status-pending`: amber-500
  - `--status-in-progress`: blue-500
  - `--status-completed`: green-500
  - `--status-failed`: red-500

Add typography utility classes:

```css
.text-display-lg {
  font-family: var(--font-display);
  font-size: 2.5rem;
  line-height: 3rem;
  font-weight: 700;
}

.text-heading-md {
  font-family: var(--font-display);
  font-size: 1.5rem;
  line-height: 2rem;
  font-weight: 600;
}

.text-body-md {
  font-family: var(--font-sans);
  font-size: 1rem;
  line-height: 1.5rem;
  font-weight: 400;
}

.text-label-sm {
  font-family: var(--font-sans);
  font-size: 0.875rem;
  line-height: 1.25rem;
  font-weight: 500;
}

.text-mono-sm {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  line-height: 1.125rem;
  font-weight: 500;
}
```

**Step 3: Verify fonts render correctly**

```bash
cd apps/dashboard && bun run dev
```

Visit localhost:3000. Inspect that `--font-display`, `--font-sans`, `--font-mono` CSS variables are applied to the `<body>` tag.

**Step 4: Commit**

```bash
git add apps/dashboard/src/app/globals.css apps/dashboard/src/app/layout.tsx
git commit -m "feat(dashboard): configure Mission Control design tokens and typography"
```

---

### Task 4: Create types and mock data

**Files:**
- Create: `apps/dashboard/src/lib/types.ts`
- Create: `apps/dashboard/src/lib/mock-data.ts`

**Step 1: Define TypeScript types mirroring Go domain**

Create `apps/dashboard/src/lib/types.ts`:

```ts
export type ApprovalStatus =
  | "none"
  | "pending"
  | "approved"
  | "rejected"
  | "needs_revision";

export type TaskStatus = "pending" | "in_progress" | "completed" | "failed";

export type Phase = "discover" | "plan" | "execute" | "validate";

export interface Epic {
  id: number;
  name: string;
  description: string;
  approval_status: ApprovalStatus;
  created_at: string;
  updated_at: string;
}

export interface EpicSummary extends Epic {
  phase: Phase;
  story_count: number;
  task_count: number;
  completed_task_count: number;
}
```

**Step 2: Create mock data**

Create `apps/dashboard/src/lib/mock-data.ts`:

```ts
import { EpicSummary } from "./types";

export const mockEpics: EpicSummary[] = [
  {
    id: 1,
    name: "user-authentication",
    description: "Implement secure authentication with OAuth2 and session management",
    approval_status: "approved",
    phase: "execute",
    story_count: 4,
    task_count: 12,
    completed_task_count: 8,
    created_at: "2026-02-15T10:00:00Z",
    updated_at: "2026-03-05T14:30:00Z",
  },
  {
    id: 2,
    name: "notification-system",
    description: "Real-time notification pipeline for agent events and approval requests",
    approval_status: "pending",
    phase: "discover",
    story_count: 2,
    task_count: 0,
    completed_task_count: 0,
    created_at: "2026-03-01T09:00:00Z",
    updated_at: "2026-03-04T16:00:00Z",
  },
  {
    id: 3,
    name: "dashboard-ui",
    description: "Browser-based Mission Control for monitoring agents, tasks, and approvals",
    approval_status: "approved",
    phase: "plan",
    story_count: 3,
    task_count: 7,
    completed_task_count: 0,
    created_at: "2026-03-02T11:00:00Z",
    updated_at: "2026-03-06T08:00:00Z",
  },
  {
    id: 4,
    name: "knowledge-base",
    description: "Persistent memory layer for cross-session project learnings",
    approval_status: "approved",
    phase: "validate",
    story_count: 2,
    task_count: 5,
    completed_task_count: 5,
    created_at: "2026-01-20T08:00:00Z",
    updated_at: "2026-02-28T17:00:00Z",
  },
];
```

**Step 3: Verify types compile**

```bash
cd apps/dashboard && bunx tsc --noEmit
```

Expected: No type errors.

**Step 4: Commit**

```bash
git add apps/dashboard/src/lib/types.ts apps/dashboard/src/lib/mock-data.ts
git commit -m "feat(dashboard): add Epic types and mock data"
```

---

### Task 5: Build the Sidebar component

**Files:**
- Create: `apps/dashboard/src/components/sidebar.tsx`
- Modify: `apps/dashboard/src/app/layout.tsx`

**Step 1: Create the Sidebar**

Create `apps/dashboard/src/components/sidebar.tsx`:

```tsx
"use client";

import {
  FolderTree,
  GitBranch,
  Bot,
  Activity,
  ShieldAlert,
  ShieldCheck,
  Brain,
  Layers,
  Settings,
} from "lucide-react";

const navItems = [
  { label: "Stories", icon: FolderTree },
  { label: "DAG View", icon: GitBranch },
  { label: "Agent Monitor", icon: Bot },
  { label: "Activity Feed", icon: Activity },
  { label: "Approvals", icon: ShieldAlert },
  { label: "QA Dashboard", icon: ShieldCheck },
  { label: "Knowledge Base", icon: Brain },
  { label: "Sessions", icon: Layers },
];

export function Sidebar() {
  return (
    <aside className="flex h-screen w-64 flex-col border-r bg-card">
      <div className="border-b px-6 py-5">
        <h1 className="text-display-lg text-xl font-bold tracking-tight">
          Oraculo
        </h1>
        <p className="text-label-sm text-muted-foreground mt-1">
          Select Epic...
        </p>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {navItems.map((item) => (
          <button
            key={item.label}
            disabled
            className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground/50 cursor-not-allowed"
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </button>
        ))}
      </nav>

      <div className="border-t px-3 py-4">
        <button
          disabled
          className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground/50 cursor-not-allowed"
        >
          <Settings className="h-4 w-4" />
          Settings
        </button>
      </div>
    </aside>
  );
}
```

**Step 2: Install lucide-react**

```bash
cd apps/dashboard && bun add lucide-react
```

**Step 3: Wire Sidebar into layout**

Update `apps/dashboard/src/app/layout.tsx` — wrap `{children}` in a flex container with the Sidebar:

```tsx
// Inside the <body> tag, replace {children} with:
<div className="flex h-screen">
  <Sidebar />
  <main className="flex-1 overflow-y-auto bg-background">
    {children}
  </main>
</div>
```

Add the import: `import { Sidebar } from "@/components/sidebar";`

**Step 4: Verify visually**

```bash
cd apps/dashboard && bun run dev
```

Expected: Sidebar appears on the left with "Oraculo" header, muted/disabled nav items, and Settings at the bottom.

**Step 5: Commit**

```bash
git add apps/dashboard/src/components/sidebar.tsx apps/dashboard/src/app/layout.tsx apps/dashboard/package.json apps/dashboard/bun.lock
git commit -m "feat(dashboard): add sidebar shell with muted navigation"
```

---

### Task 6: Build the EpicCard component

**Files:**
- Create: `apps/dashboard/src/components/epic-card.tsx`

**Step 1: Create the EpicCard**

Create `apps/dashboard/src/components/epic-card.tsx`:

```tsx
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import type { EpicSummary, Phase } from "@/lib/types";

const phaseColors: Record<Phase, string> = {
  discover: "bg-amber-500/10 text-amber-500 border-amber-500/20",
  plan: "bg-blue-500/10 text-blue-500 border-blue-500/20",
  execute: "bg-purple-500/10 text-purple-500 border-purple-500/20",
  validate: "bg-green-500/10 text-green-500 border-green-500/20",
};

interface EpicCardProps {
  epic: EpicSummary;
  onOpen: (epic: EpicSummary) => void;
}

export function EpicCard({ epic, onOpen }: EpicCardProps) {
  const completionPct =
    epic.task_count > 0
      ? Math.round((epic.completed_task_count / epic.task_count) * 100)
      : 0;

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-3">
        <CardTitle className="text-heading-md text-lg font-semibold">
          {epic.name}
        </CardTitle>
        <Badge variant="outline" className={phaseColors[epic.phase]}>
          {epic.phase}
        </Badge>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-4">
        <p className="text-body-md text-sm text-muted-foreground line-clamp-2">
          {epic.description}
        </p>

        <div className="space-y-2">
          <Progress value={completionPct} className="h-2" />
          <div className="flex items-center justify-between text-mono-sm text-xs text-muted-foreground">
            <span>{epic.story_count} stories</span>
            <span>
              {epic.completed_task_count}/{epic.task_count} tasks
            </span>
            <span>{completionPct}%</span>
          </div>
        </div>

        <Button className="mt-auto w-full" onClick={() => onOpen(epic)}>
          Open
        </Button>
      </CardContent>
    </Card>
  );
}
```

**Step 2: Verify types compile**

```bash
cd apps/dashboard && bunx tsc --noEmit
```

Expected: No type errors.

**Step 3: Commit**

```bash
git add apps/dashboard/src/components/epic-card.tsx
git commit -m "feat(dashboard): add EpicCard component with phase badge and progress"
```

---

### Task 7: Build the CreateEpicDialog component

**Files:**
- Create: `apps/dashboard/src/components/create-epic-dialog.tsx`

**Step 1: Create the dialog**

Create `apps/dashboard/src/components/create-epic-dialog.tsx`:

```tsx
"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Plus } from "lucide-react";

interface CreateEpicDialogProps {
  onCreate: (name: string, description: string) => void;
}

export function CreateEpicDialog({ onCreate }: CreateEpicDialogProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    onCreate(name.trim(), description.trim());
    setName("");
    setDescription("");
    setOpen(false);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="gap-2">
          <Plus className="h-4 w-4" />
          New Epic
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Epic</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="epic-name" className="text-label-sm text-sm font-medium">
              Name
            </label>
            <Input
              id="epic-name"
              placeholder="e.g. user-authentication"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <label htmlFor="epic-desc" className="text-label-sm text-sm font-medium">
              Description
            </label>
            <Textarea
              id="epic-desc"
              placeholder="Brief description of the epic..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button type="button" variant="ghost" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={!name.trim()}>
              Create
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Verify types compile**

```bash
cd apps/dashboard && bunx tsc --noEmit
```

Expected: No type errors.

**Step 3: Commit**

```bash
git add apps/dashboard/src/components/create-epic-dialog.tsx
git commit -m "feat(dashboard): add CreateEpicDialog with name and description fields"
```

---

### Task 8: Assemble the Landing page

**Files:**
- Modify: `apps/dashboard/src/app/page.tsx`

**Step 1: Build the Landing page**

Replace the contents of `apps/dashboard/src/app/page.tsx`:

```tsx
"use client";

import { useState } from "react";
import { EpicCard } from "@/components/epic-card";
import { CreateEpicDialog } from "@/components/create-epic-dialog";
import { mockEpics } from "@/lib/mock-data";
import type { EpicSummary } from "@/lib/types";

export default function LandingPage() {
  const [epics, setEpics] = useState<EpicSummary[]>(mockEpics);

  function handleCreate(name: string, description: string) {
    const newEpic: EpicSummary = {
      id: Date.now(),
      name,
      description,
      approval_status: "none",
      phase: "discover",
      story_count: 0,
      task_count: 0,
      completed_task_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    setEpics((prev) => [...prev, newEpic]);
  }

  function handleOpen(epic: EpicSummary) {
    // Future: navigate to stories view for this epic
    console.log("Open epic:", epic.name);
  }

  return (
    <div className="mx-auto max-w-6xl px-8 py-12">
      {/* Hero */}
      <section className="mb-12">
        <h1 className="text-display-lg text-4xl font-bold tracking-tight">
          Oraculo
        </h1>
        <p className="text-body-md mt-3 max-w-2xl text-muted-foreground">
          Mission Control for AI-agent-driven development. Monitor your agents,
          track task progress, review artifacts, and approve decisions — all from
          one place.
        </p>
      </section>

      {/* Epic Grid Header */}
      <section>
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-heading-md text-2xl font-semibold">Epics</h2>
          <CreateEpicDialog onCreate={handleCreate} />
        </div>

        {epics.length === 0 ? (
          <div className="rounded-lg border border-dashed p-12 text-center">
            <p className="text-muted-foreground">
              No epics yet. Create your first one to get started.
            </p>
          </div>
        ) : (
          <div className="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
            {epics.map((epic) => (
              <EpicCard key={epic.id} epic={epic} onOpen={handleOpen} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
```

**Step 2: Verify the full page renders**

```bash
cd apps/dashboard && bun run dev
```

Expected: Landing page shows "Oraculo" hero, 4 mock epic cards in a responsive grid, and a "New Epic" button that opens the creation dialog. Cards show phase badges, progress bars, and stats.

**Step 3: Verify static build**

```bash
cd apps/dashboard && bun run build
```

Expected: Build succeeds. `out/` directory contains `index.html`.

**Step 4: Commit**

```bash
git add apps/dashboard/src/app/page.tsx
git commit -m "feat(dashboard): assemble Landing page with hero, epic grid, and create dialog"
```

---

### Task 9: Add dark mode support

**Files:**
- Modify: `apps/dashboard/src/app/layout.tsx`

**Step 1: Install next-themes**

```bash
cd apps/dashboard && bun add next-themes
```

**Step 2: Create a ThemeProvider**

Create `apps/dashboard/src/components/theme-provider.tsx`:

```tsx
"use client";

import { ThemeProvider as NextThemesProvider } from "next-themes";

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider attribute="class" defaultTheme="dark" enableSystem>
      {children}
    </NextThemesProvider>
  );
}
```

**Step 3: Wrap layout with ThemeProvider**

In `apps/dashboard/src/app/layout.tsx`, wrap the `<div className="flex h-screen">` with `<ThemeProvider>`:

```tsx
import { ThemeProvider } from "@/components/theme-provider";

// Inside the <body>:
<ThemeProvider>
  <div className="flex h-screen">
    <Sidebar />
    <main className="flex-1 overflow-y-auto bg-background">
      {children}
    </main>
  </div>
</ThemeProvider>
```

**Step 4: Verify dark mode works**

```bash
cd apps/dashboard && bun run dev
```

Expected: Dashboard renders in dark mode by default. The `<html>` tag has `class="dark"`.

**Step 5: Verify static build**

```bash
cd apps/dashboard && bun run build
```

Expected: Build succeeds (next-themes is compatible with static export).

**Step 6: Commit**

```bash
git add apps/dashboard/src/components/theme-provider.tsx apps/dashboard/src/app/layout.tsx apps/dashboard/package.json apps/dashboard/bun.lock
git commit -m "feat(dashboard): add dark mode support via next-themes"
```

---

### Task 10: Final verification and cleanup

**Step 1: Run full type check**

```bash
cd apps/dashboard && bunx tsc --noEmit
```

Expected: No type errors.

**Step 2: Run linter**

```bash
cd apps/dashboard && bun run lint
```

Expected: No lint errors (fix any that appear).

**Step 3: Run production build**

```bash
cd apps/dashboard && bun run build
```

Expected: Build succeeds. `out/` directory has `index.html` and all static assets.

**Step 4: Add `out/` to .gitignore**

Append to `apps/dashboard/.gitignore` (created by create-next-app):

```
out/
```

**Step 5: Final commit**

```bash
git add apps/dashboard/.gitignore
git commit -m "chore(dashboard): add out/ to gitignore"
```
