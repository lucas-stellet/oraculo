# Research Prompt 4: QA Architecture and Validation Pipeline

## Context: Oraculo Agent Philosophy

QA agents are the immune system. They operate with completely clean context windows, no shared memory with generators. Validation is adversarial, not cooperative. The Trust Layer (CLI) is the final arbiter. Different LLM models across roles prevent shared bias cascading.

Key beliefs from the philosophy:
- Independence is architectural: clean context, no access to code agent's reasoning
- No DAG edge connects Execute to Deliver without passing through Validate
- Effective validation needs orthogonal expertise: functional correctness + adversarial testing
- The CLI verifies, not the QA agent -- deterministic reality over probabilistic opinion
- Heterogeneous models across roles prevent shared training biases

Architecture constraints:
- QA agents are separate Claude Code sub-agents with their own worktrees
- The CLI runs tests and validates contracts deterministically
- SQLite logs all validation events (append-only)
- The orchestrator can use different models for different agent roles

## Research Mission

Design the QA architecture and validation pipeline for Oraculo. I need to understand how to implement independent validation, adversarial testing, mutation testing as validator-of-validators, and heterogeneous model selection.

## Guiding Questions

1. **QA agent types and responsibilities**: What distinct QA roles are needed? (functional reviewer, adversarial tester, security auditor, style/convention checker). How many QA agents per code agent? What's the optimal review chain topology -- sequential, parallel, or hybrid?

2. **Clean context enforcement**: How to guarantee the QA agent has no context from the generation process? Separate worktree is necessary but sufficient? Should the QA agent receive only the diff + test results, or the full implementation + specs? What context produces the best review quality?

3. **Adversarial testing implementation**: How to build an agent that actively tries to break code? What prompt patterns produce effective adversarial behavior? How to scope adversarial testing (fuzzing inputs? edge cases? security vulnerabilities? race conditions?)? How to prevent the adversarial agent from generating false positives that waste cycles?

4. **Mutation testing integration**: How to automate mutation testing as part of the validation pipeline? Which mutation testing tools support which languages? How to integrate mutation results into the CLI's accept/reject decision? What mutation score threshold is practical (the philosophy suggests >80%)?

5. **Heterogeneous model selection**: How to choose which model for which role? What's the API/configuration for specifying model per agent type? Should model selection be static (configured) or dynamic (based on task complexity)? What combinations have been shown to reduce shared bias effectively?

6. **Validation pipeline orchestration**: How does the validation pipeline fit into the DAG? Is it a sub-DAG within each Validate node? What's the flow: code agent completes -> functional QA -> adversarial QA -> mutation testing -> CLI contract check -> approved/rejected? How to handle partial approval (functional passes but adversarial finds an edge case)?

## Expected Output

For each finding, provide:
- **Pattern/Concept**: Name
- **Source**: URL, paper, repository, or documentation
- **Summary**: 2-3 sentences describing the implementation approach
- **Applicability to Oraculo**: How it maps to QA agents with CLI Trust Layer and SQLite logging
- **Key design decision**: The one implementation choice this informs

Focus on QA automation in agent systems (2024-2026). Include mutation testing tools, adversarial testing patterns, and any production implementations of multi-model validation chains.

## Reference Implementations to Study

Study these projects for concrete QA and validation patterns:

- **BMAD-METHOD** (https://github.com/bmad-code-org/BMAD-METHOD) -- Breakthrough Method for Agile AI-Driven Development. Defines explicit agent roles with specialized responsibilities in a structured agile workflow. Relevant for: role separation between development and validation agents, structured review workflows, phase-gated quality checks.

- **spec-workflow-mcp** (https://github.com/Pimzino/spec-workflow-mcp) -- MCP server with approval workflow and complete revision process. Tasks go through structured approval before progressing. Relevant for: approval gates as validation checkpoints, revision workflows when validation rejects output, tracking validation state per task.
