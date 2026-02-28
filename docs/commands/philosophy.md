# Oraculo Commands — Philosophy

## 1. Purpose

Commands are the conversational surface of Oraculo.

They are the point where a human intention — "I have a feature idea", "I need to implement this story" — enters the system and becomes structured work. Each command maps to a phase of Oraculo's operating model. The user invokes a slash command, the command guides a Socratic dialogue, and the output feeds the CLI Trust Layer and the agent workforce.

Commands are not scripts. They do not automate a checklist. They encode a discipline — a sequence of questions, frameworks, and quality gates that forces the user (and the LLM) to think before acting. The command is the boundary between casual intent and rigorous execution.

## 2. Core Belief

**A command must make the disciplined path feel effortless and the undisciplined path feel impossible.**

LLMs are eager to act. Given a vague prompt, they will generate code, skip discovery, weaken constraints, and produce confident garbage. Commands exist to prevent this. They channel the LLM's energy through a structured workflow where skipping steps is harder than following them. The command does not rely on the model's judgment to maintain rigor — it structures the interaction so that rigor is the path of least resistance.

## 3. Composition Over Monoliths

**A command is a thin entry point, not a self-contained program.** The SKILL.md file defines the workflow, the persona, and the quality gates. It does not contain the frameworks, question banks, or artifact templates inline. Those live in reference files, loaded only when needed.

This is a deliberate constraint. A monolithic prompt of hundreds of lines degrades LLM performance — attention dilutes, instructions at the top are forgotten by the time the model reaches the bottom. Smaller, focused documents loaded at the right moment produce better adherence than a single massive file loaded upfront.

**Reference files are internal, not commands.** Only SKILL.md files surface as slash commands. Supporting files — question banks, framework descriptions, templates — live alongside the skill but never pollute the command list. The user sees a clean interface; the skill has access to deep reference material when the workflow demands it.

**Commands never duplicate what the CLI enforces.** Schema validation, state persistence, idempotent operations — these belong to the Trust Layer. A command calls the CLI; it does not reimplement its guarantees. If a command contains logic that could be a CLI contract, it is in the wrong place.

**Vocabulary is minimal and consistent.** Every concept a command introduces — a phase name, a framework label, an anti-pattern handle — must earn its place. Domain-specific jargon and extended metaphors create a learning curve that competes with the actual workflow. Commands use the smallest vocabulary that achieves precision. New terms are introduced only when plain language would be ambiguous.

## 4. Behavioral Containment

**Commands shape LLM behavior through structure, not hope.** Telling an LLM "be thorough" is decoration. Giving it a phase gate that requires three validated assumptions before proceeding is containment. Commands use structural techniques:

- **Phase gates** prevent forward progress until quality criteria are met. The model cannot skip discovery to start planning — the workflow does not offer that path.
- **Named anti-patterns** give the LLM cognitive handles for failure modes. "Premature solutioning" is easier to detect and avoid than "don't jump to conclusions." When a failure mode has a name, the model can pattern-match against it.
- **Explicit deviation rules** define what happens when the user pushes back or asks to skip. The command knows which steps are compressible and which are inviolable. This is not rigidity — it is clarity about where flexibility exists and where it does not.

**Repetition of critical rules is intentional, not sloppy.** LLMs lose instruction adherence over long conversations. Rules that govern safety, quality, or workflow integrity appear at the point of use — not only in a preamble that the model has long forgotten. This is not redundancy; it is reinforcement at the moment of maximum drift risk.

**Approval is a human verdict, not a verbal signal.** When a command reaches an approval gate — requirements ready for review, a plan ready for execution — it calls `oraculo tools approval request` to submit the artifact and enters `awaiting_approval` state. The dashboard presents the artifact to a human reviewer, who issues a verdict: `approved` (advance), `rejected` (return to generator phase), or `needs_revision` (return with comments). "Looks good" in conversation is not approval. A verdict delivered through the dashboard's approval gate is approval. Commands never proceed on verbal confirmation alone for irreversible transitions. The workflow does not advance until the verdict is received.

## 5. Context as a Finite Resource

**Every token in a command's context has a cost.** The context window is not infinite storage — it is working memory. Commands treat it with the same discipline a real-time system applies to RAM.

**Load late, unload early.** Reference material enters the context only when the workflow phase demands it. A question bank for assumption mapping is not loaded during problem framing. A requirements template is not loaded until the user has validated the exploration. This just-in-time loading keeps the active context focused on the current task.

**The command's own instructions compete with the conversation.** As dialogue grows, the ratio of instruction tokens to conversation tokens shrinks. Commands are written to be concise — every sentence earns its place. Verbose instructions do not produce better adherence; they produce diluted attention.

**State belongs in storage, not in the conversation.** Intermediate results — validated assumptions, explored constraints, partially formed requirements — are persisted through the CLI as soon as they stabilize. If the context window overflows or the session is interrupted, no validated work is lost. The conversation can be rebuilt from persisted state; persisted state cannot be rebuilt from a lost conversation.

## 6. The Three Layers

Commands operate within a strict hierarchy. Each layer has a single responsibility.

**Commands** are the conversational interface. They own the Socratic workflow — what questions to ask, what frameworks to apply, what artifacts to produce. They interact with the user, guide discovery, and produce structured output. Commands are interpretive — they reason, adapt, and respond to nuance.

**The CLI Trust Layer** is the deterministic foundation. Commands delegate all persistence, validation, and state management to the CLI. The command declares intent; the CLI guarantees correctness. This boundary is absolute — a command never writes directly to the database, constructs file paths, or manages schemas. If the CLI rejects an operation, the command does not retry with looser constraints — it surfaces the rejection to the user.

**Agents** are the execution workforce. When a command's output (requirements, stories, plans) is ready for implementation, the orchestrator dispatches agents. Commands do not execute code, run tests, or modify the codebase. They produce the artifacts that agents consume.

The flow is always: **Command discovers and defines -> CLI persists and validates -> Agents execute and deliver.** At approval gates, the flow pauses: the Dashboard presents the artifact and awaits a human verdict before the workflow advances. No layer does another layer's job.

This separation is not organizational convenience — it is a reliability guarantee. When each layer does exactly one thing, failures are local, debuggable, and recoverable. A command that also persists data is untestable. An agent that also does discovery is uncontainable. The layers exist because mixing responsibilities produces systems that are fragile in ways that are invisible until they break.

## 7. Inputs and Outputs

**Commands accept arguments, not configuration.** A command receives what it needs through `$ARGUMENTS` — the user's intent, a reference to an existing artifact, or a scope declaration. There are no config files, no environment variables, no implicit state that changes behavior. The command's behavior is fully determined by its SKILL.md and its arguments.

**Command output is an artifact, not a conversation log.** The Socratic dialogue is the process; the artifact is the product. An epic command produces a requirements document. A story command produces a story definition. These artifacts are structured, validated, and persisted through the CLI. The conversation that produced them is ephemeral — the artifact is what survives.

**Handoffs are explicit and gated.** When a command completes its phase and the workflow transitions — from discovery to planning, from planning to execution — the handoff requires an approval verdict before completing. The command submits the artifact to the Approval Gate via `oraculo tools approval request`, enters `awaiting_approval`, and does not advance until the dashboard delivers a verdict. Only an `approved` verdict allows the next phase to begin with a clean context and a validated input. This prevents context bleed and ensures each phase operates on agreed-upon facts, not conversational drift or unreviewed artifacts.

## 8. Portability

**Commands are plain markdown files. No build step, no compilation, no runtime.** A SKILL.md and its reference files are copied into a project's `.claude/` directory and they work. There is no transpilation from YAML to XML, no template engine, no plugin registry. The file is the command.

This constraint is non-negotiable. Build steps create version drift, toolchain dependencies, and failure modes that have nothing to do with the command's purpose. A command that requires a compiler is a command that will break in environments the author never tested.

## 9. What This Document Is Not

This document defines the principles that govern command design. It does not specify individual commands, their SKILL.md contents, or their reference file structures. Those belong in the design document.

This document does not replace the CLI philosophy (docs/cli/philosophy.md) or the agent philosophy (docs/agents/philosophy.md). Commands depend on the CLI for deterministic operations and on agents for execution. This document defines how commands occupy the interpretive layer between user intent and system action.
