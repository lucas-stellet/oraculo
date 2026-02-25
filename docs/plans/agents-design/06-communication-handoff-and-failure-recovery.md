# Research Prompt 6: Communication, Handoff, and Failure Recovery

## Context: Oraculo Agent Philosophy

The SQLite + CLI functions as a deterministic blackboard. Handoffs are first-class inspectable transitions with structured JSON contracts. The orchestrator treats QA capacity as the governing constraint. Circuit breakers isolate failing branches. Coordinated checkpointing enables partial rollback without destroying parallel work.

Key beliefs from the philosophy:
- Governance: who can write what and when, with explicit policies
- Handoffs preserve task description, state, expected outputs, constraints, and prior results
- SOPs codify handoff formats as machine-readable templates
- Writes serialized through CLI; reads parallel
- Circuit breakers block corrupted data from flowing downstream

Architecture constraints:
- Agents communicate exclusively through CLI commands and SQLite state
- No direct agent-to-agent messaging (all goes through the blackboard)
- The orchestrator monitors agent health and DAG progress
- Each phase (Discover/Plan/Execute/Validate/Deliver) materializes as a checkpoint

## Research Mission

Design the communication protocol, handoff mechanism, and failure recovery system for Oraculo. I need to understand how agents exchange work products, how the orchestrator detects and recovers from failures, and how checkpoints enable resilient execution.

## Guiding Questions

1. **Handoff contract schema**: What's the JSON schema for an agent handoff? What fields are mandatory (task_id, source_agent, target_agent, payload, expected_output_schema, constraints)? How does the CLI validate handoff contracts before delivery? How to version handoff schemas as the system evolves?

2. **Blackboard governance model**: How to implement write policies on the SQLite blackboard? Role-based access (orchestrator can write task assignments, code agents can write implementation results, QA agents can write validation verdicts)? How to enforce these policies in the CLI? What's the authorization model?

3. **SOP templates for common workflows**: What are the standard handoff sequences for common operations? (e.g., "new feature": Epic -> Stories -> test-author -> implementer -> QA -> merge). How to represent SOPs as reusable DAG templates? How to parameterize them per project?

4. **Circuit breaker implementation**: How to monitor failure rates per agent/node? What thresholds trigger the circuit breaker? (N consecutive failures? Error rate above X%? Latency above Y seconds?). What's the recovery sequence: pause -> log state -> deploy debugger agent -> retry with refined prompt? How to implement this in the CLI + SQLite architecture?

5. **Coordinated checkpointing**: How to implement phase-level checkpoints in SQLite? What state must a checkpoint capture (DAG state, all node outputs, memory state)? How to rollback to a checkpoint without affecting unrelated parallel branches? How to detect that a rollback is needed (vs. a simple retry)?

6. **Cross-phase context propagation**: How does context from Discover propagate to Plan, and from Plan to Execute? What gets summarized vs. passed verbatim? Who creates the summary -- the orchestrator, a dedicated summarizer agent, or the CLI? How to ensure no critical context is lost during phase transitions?

## Expected Output

For each finding, provide:
- **Pattern/Concept**: Name
- **Source**: URL, paper, repository, or documentation
- **Summary**: 2-3 sentences describing the implementation approach
- **Applicability to Oraculo**: How it maps to CLI blackboard + SQLite checkpoints + Claude Code sub-agents
- **Key design decision**: The one implementation choice this informs

Focus on resilient distributed system patterns (2024-2026). Include patterns from workflow engines (Temporal's durable execution, Prefect's failure handling), circuit breaker implementations, and any agent-specific failure recovery mechanisms.

## Reference Implementations to Study

Study these projects for concrete communication and recovery patterns:

- **Gastown** (https://github.com/steveyegge/gastown) -- Uses convoys (agent groups), queues, and escalation mechanisms for multi-agent coordination. Git-backed issue tracking stores structured work state. Dashboard provides single-page overview of agents, convoys, hooks, queues, issues, and escalations. Relevant for: agent coordination protocols, escalation patterns, work state handoff between agents.

- **spec-workflow-mcp** (https://github.com/Pimzino/spec-workflow-mcp) -- Structured workflow with approval gates and revision cycles (Requirements -> Design -> Tasks -> Implementation). Real-time dashboard for monitoring specs, tasks, and progress. Relevant for: phase transitions with approval gates, progress monitoring, structured revision workflows when tasks fail validation.

- **BMAD-METHOD** (https://github.com/bmad-code-org/BMAD-METHOD) -- Agent workflows invocable via slash commands with direct workflow invocation. Structured development phases with clear handoff points. Relevant for: workflow-as-code patterns, phase-based handoff protocols, command-driven agent coordination.
