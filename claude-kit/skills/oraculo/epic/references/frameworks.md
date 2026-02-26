# Frameworks Reference

Concise descriptions of each framework used in Epic exploration and when it applies.

---

## Double Diamond (Diverge/Converge)

The session follows two movements: first expand the problem space (diverge), then narrow to a clear definition (converge). This prevents the most common failure — jumping to solutions before understanding the problem. Both Light and Deep levels use this structure; Light diverges in the technical space, Deep in the product space.

**When:** Always. This is the macro-structure of every epic session.

---

## Jobs to Be Done (JTBD)

Users don't want features — they "hire" products to make progress in specific circumstances. JTBD explores the functional, emotional, and social dimensions of the user's need, producing Job Stories instead of feature descriptions.

Includes the **Four Forces of Progress** — Push (frustration with current state), Pull (attraction of the new), Anxiety (fear of the unknown), Habit (inertia of current behavior). These forces surface adoption edge cases that functional analysis misses.

**When:** Divergence phase. Full dimensions (functional, emotional, social) and all four forces in Deep only. Light uses a technical variant focused on system behavior, pre-conditions, and technical adoption forces.

---

## Theory of Constraints (TOC)

Every system has a single constraint that limits its throughput. TOC asks: what is the one bottleneck that, if removed, would unlock the most value? Applied to discovery, this forces prioritization toward the highest-impact problem.

**When:** Convergence phase, both levels. Deep identifies the user/business constraint. Light identifies the technical constraint. The same principle applies later in execution, where QA throughput governs the pace.

---

## Assumption Mapping

Makes hidden assumptions explicit and scores them by risk: Impact (if wrong, how bad?) x (6 - Evidence). The riskiest assumptions become priorities before any investment.

**When:** Assumption Mapping phase, both levels. Light focuses on Solution and Feasibility categories; Deep scores all four categories (Problem, Solution, Value, Feasibility).

---

## Technical Viability Checklist

Light-mode stress test. Tests whether the implementation approach fits the current architecture, what tests are needed, what code review objections might arise, and what technical acceptance criteria apply.

**When:** Exit Gate phase, Light only.
