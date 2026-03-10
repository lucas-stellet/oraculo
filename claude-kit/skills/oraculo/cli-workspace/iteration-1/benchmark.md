# Benchmark — oraculo:cli — Iteration 1

## Summary

| Config | Pass Rate | Passed/Total | Mean Tokens | Mean Duration |
|--------|-----------|-------------|-------------|---------------|
| **with_skill** | **100%** | 17/17 | 17,195 | 31.4s |
| without_skill | 82.4% | 14/17 | 25,774 | 51.5s |
| **Delta** | **+17.6pp** | +3 | **-33%** | **-39%** |

## Per-eval breakdown

| Eval | with_skill | without_skill | Delta |
|------|-----------|--------------|-------|
| eval-epic-story | 5/5 (100%) | 5/5 (100%) | 0pp |
| eval-task-lifecycle | 6/6 (100%) | 3/6 (50%) | **+50pp** |
| eval-approval-knowledge | 6/6 (100%) | 6/6 (100%) | 0pp |

## Failed assertions (without_skill)

All 3 failures in `eval-task-lifecycle`:
- `creates-all-three-tasks` — used `oraculo task init` instead of `oraculo tools task init`
- `uses-task-start` — used `oraculo task start` instead of `oraculo tools task start`
- `uses-task-list` — used `oraculo task list` instead of `oraculo tools task list`

**Root cause:** Baseline read source code but still applied incorrect mental model of the CLI namespace. Even with source access, agents may drop the `tools` subcommand for non-epic/story domains.

## Analyst Notes

- **Eval-0 and Eval-2** are non-discriminating on correctness in this test environment because the baseline agent read the project source code to discover commands. In a real Oraculo installation (different project, no source access), both evals would likely fail completely without the skill.
- **Eval-1** is the most valuable assertion set — it caught a real correctness error even in source-access conditions.
- **Token efficiency**: With skill uses 33% fewer tokens and completes 39% faster. In production, this difference would be larger (baseline would spend even more tokens searching an unfamiliar codebase).
- **Recommendation**: The `correct-command-namespace` assertion should be replicated in all evals (currently only in eval-0) since it's the most fundamental correctness check.
