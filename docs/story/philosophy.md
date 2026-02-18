# Story Phase — Philosophy

## 1. Purpose

The Story phase exists to transform work items into well-defined, executable units of work.

Not every piece of work needs the deep exploration of an Epic. When the problem is understood, the scope is clear, and the work is focused, what's needed is a disciplined process to define it precisely — clear enough for a developer to implement without ambiguity.

The Story phase applies the same Socratic discipline as the Epic, but with less depth and more focus. It asks the essential questions, surfaces the critical assumptions, and produces a compact artifact that captures WHAT and WHY.

## 2. Core Belief

**Even clear tasks carry hidden assumptions.**

When someone says "add a delete button to the profile page," the implicit beliefs are: users want to delete their profiles, deletion is reversible (or isn't), the button should be visible (not hidden behind a menu), and the action doesn't cascade to other data. These assumptions are invisible until someone asks about them.

The Story phase makes the essential questions unavoidable — not through exhaustive exploration, but through focused inquiry that catches the most dangerous blind spots.

## 3. Theoretical Foundations

The Story phase draws from two frameworks, simplified for focused work:

### 3.1 Double Diamond (Simplified)

The session follows a compressed diamond: a brief divergence during reframing (explore the problem from different angles), then quick convergence to a focused definition. Unlike the Epic's full diamond, the Story's divergence is narrow — it opens just enough to catch missed angles, then closes rapidly.

**Why it matters:** Even for small work, the instinct to jump straight to "how" skips the "why." A brief divergence catches problems that would become expensive to fix after implementation.

### 3.2 Assumption Mapping (Simplified)

Identifies the top 2-3 assumptions and assesses their evidence level (high/medium/low). Unlike the Epic's full scored register, the Story focuses on catching the most critical assumptions without exhaustive analysis.

**Why it matters:** Every piece of work carries assumptions. The Story doesn't need to map all of them — it needs to surface the ones that would cause the work to fail or need redoing if they're wrong.

## 4. What Makes Oraculo's Story Definition Unique

### 4.1 Same Discipline, Less Depth

The Story phase is not a shortcut or a skip. It's the same Socratic approach — reframe, check assumptions, gate before generating — applied at the right depth for the work's complexity. The discipline never drops; the scope narrows.

### 4.2 Epic-Aware

Stories can be standalone or derived from an Epic. When derived, the Story skill reads the parent Epic, understands the context of the specific REC-N being worked on, and adapts its questions accordingly. The user doesn't repeat context already captured in the Epic.

### 4.3 Escalation-Ready

If during a Story session the work reveals itself to be larger than expected — multiple areas, high uncertainty, vague scope — Oraculo suggests escalating to `/oraculo:epic` rather than trying to force complex work into a story-sized definition.

## 5. Principles Specific to Story Definition

These principles extend the core Oraculo principles for this specific phase:

- **Reframe before defining** — Even for clear tasks, separate the problem from the proposed solution. "Add a delete button" becomes "users need a way to remove their profile."

- **No solution talk before problem clarity** — Implementation details are forbidden until the problem and desired outcome are clear.

- **Assumptions are explicit** — Every critical assumption must be stated and acknowledged. Not all assumptions need scoring — but the dangerous ones must be visible.

- **Gate before generating** — The Story phase has an exit gate. If the four risks aren't addressed, the artifact isn't generated.

- **Escalate, don't force** — If the work is too big for a story, suggest an Epic. Don't compress complex work into a simple format.

## 6. Relationship to Epics

A story can exist in two modes:

- **Standalone:** Work that doesn't need an Epic. The problem is clear, the scope is small, and the Story skill provides enough structure to define it.

- **Derived from Epic:** A specific REC-N from an Epic is being decomposed into an executable story. The Story skill reads the parent Epic and uses it as context for more focused questions.

In both cases, the output is the same: a Story Document that captures motivation, requirement, scope, and assumptions.
