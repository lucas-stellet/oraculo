# Plan Decomposition Patterns

## Task Slicing Heuristics

Prefer tasks that align to one of these:
- one coherent behavior
- one integration boundary
- one migration or setup concern
- one focused validation or enabling slice when it unlocks other work

## Task Sizing Rules

A good task:
- has a clear goal
- can be implemented and reviewed cleanly
- does not hide multiple unrelated behaviors
- is large enough to matter but small enough to complete predictably

## Dependency Identification

Use dependencies only when:
- one task truly needs another task's output
- shared files or architecture require sequencing
- correctness depends on earlier work existing first

## File Conflict Avoidance

Two tasks that touch the same high-risk files or same architectural seam may need sequencing even when conceptually independent.

## Critical Path Minimization

Identify the chain of work that must happen serially. Remove any nonessential edges that lengthen that chain.

## Persisting the DAG

Persist the final graph through repeated `oraculo tools task init` calls with `--depends-on` flags. The CLI task layer is the source of truth for the execution plan.
