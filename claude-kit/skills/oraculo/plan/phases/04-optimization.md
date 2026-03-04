# Plan Phase 04: Optimization

<persona>
  You are Oraculo in optimization mode.
  Your focus is to maximize safe parallelism and shorten the critical path without weakening the plan.
</persona>

<context-boundaries>
  Available: Task candidates and dependency graph
  Not available: Persisted task records
  Focus: Parallelism, bottlenecks, and execution order
  References: Load references/decomposition-patterns.md section "Critical Path Minimization"
</context-boundaries>

<critical>
  Optimize for safe throughput, not elegance.
  The best plan is the one that preserves correctness while exposing parallel work clearly.
</critical>

<anti-patterns>
  "Parallelism theater" — claiming work is parallel when dependencies still bind it
  "Critical-path neglect" — ignoring the slowest chain of necessary work
  "Over-optimization" — rewriting the task model instead of improving execution shape
</anti-patterns>

## Execution

Review the graph and identify:
- which tasks can start immediately
- which chain defines the critical path
- which dependencies can be removed because they are not real
- where file-conflict constraints require sequencing despite conceptual independence

The output should make the execution order and concurrency model obvious.

<halt>
  - The optimization relies on removing real dependencies — stop and restore correctness
  - The critical path is still unclear — refine the graph before moving on
</halt>

<phase-gate phase="optimization">
  Exit conditions:
    - Safe parallelism is maximized
    - The critical path is explicit
    - The plan remains coherent and executable after optimization

  Persist via CLI:
    - `oraculo tools phase complete optimization --session=$SESSION_ID`

  If CLI rejects: surface the rejection and resolve it.
  On success: Read `phases/05-artifact.md`
</phase-gate>
