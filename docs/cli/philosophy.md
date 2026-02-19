# Oraculo CLI — Philosophy

## 1. Purpose

The Oraculo CLI is the **Trust Layer** of the system.

Oraculo operates at two fundamentally different levels: interpretive work and deterministic work. Skills and agents handle the interpretive side — Socratic dialogue, codebase analysis, creative reasoning. The CLI handles the deterministic side — persisting data, enforcing schemas, validating operations, returning predictable results.

The CLI draws a hard line between these two levels. The consumer — whether an agent or a human — declares WHAT it wants. The CLI guarantees the HOW with a predictable, validated result. No guessing, no interpretation, no ambiguity.

## 2. Core Belief

**Consumers should trust the tool completely — agents and humans alike.**

When an agent needs to persist a finding about the codebase, it should not need to know that the storage is SQLite, that FTS5 indexes require sync triggers, or that the database lives at `.oraculo/oraculo.db`. It should call `oraculo memory store --domain payments --finding "Uses repository pattern"` and trust the tool to handle the rest.

The CLI is the point in the system where the result is deterministic and validated. It eliminates assumptions: the consumer does not need to guess paths, construct schemas, remember formats, or handle edge cases. The interface is the contract, and the contract is enforced.

The same interface serves any context. Inside Claude Code or outside it. Called by a skill, an agent, a human at the terminal, or a CI pipeline. The trust is unconditional and universal.

## 3. Theoretical Foundations

The CLI's design is grounded in three established concepts from software engineering.

### 3.1 Design by Contract (Bertrand Meyer)

Every CLI command operates under a contract: preconditions that must be met, postconditions that are guaranteed, and invariants that are preserved. The CLI validates strictly — operations that violate preconditions are refused with a clear error, not silently accepted with unpredictable results.

**Why it matters:** Agents are powerful reasoners but unreliable system operators. They hallucinate paths, forget schemas, and produce subtly broken commands. A strict contract means the agent either gets a correct result or a clear rejection. There is no middle ground where bad data silently enters the system.

### 3.2 Pit of Success (Brad Abrams)

The correct path is the easy path. The CLI is designed so that using it correctly requires less effort than using it incorrectly. The consumer does not need to construct file paths, remember database schemas, or deduce output formats. The CLI handles all of that — the consumer just states its intent.

**Why it matters:** When the easy thing to do is also the right thing to do, errors become rare. The CLI does not rely on consumers reading documentation carefully or remembering conventions. It makes the wrong thing hard and the right thing obvious.

### 3.3 Trusted Computing Base (TCB)

The Trusted Computing Base is the minimal core that the entire system depends on for correctness. If the TCB is reliable, everything built on top of it can be built with confidence. The CLI is Oraculo's TCB — the foundation that skills and agents trust without question.

**Why it matters:** Skills handle Socratic dialogue. Agents handle codebase investigation. Both depend on the CLI for persistence, validation, and structured operations. If the CLI is trustworthy, skills and agents can focus on their interpretive work without worrying about data integrity. If the CLI is not trustworthy, nothing built on top of it can be trusted either.

## 4. Design Principles

### 4.1 Zero-Dependency Distribution

The CLI is a single static binary built in Go. No runtime, no package manager, no system libraries. It runs on any machine where Claude Code runs. This is non-negotiable — if the tool requires setup, agents cannot rely on it.

Go was chosen specifically for this: `modernc.org/sqlite` provides a pure-Go SQLite implementation without CGO. The binary compiles, copies, and runs. Nothing else.

### 4.2 Contextual Output

Commands serve two types of questions and return two types of responses:

- **Operational commands** return structured JSON. These are commands that report status, create resources, or manage state. Agents and scripts parse them reliably.
- **Content commands** return raw content. These are commands that retrieve documents, requirements, or knowledge meant to be read. Wrapping human-readable text in JSON adds noise without value.

The output format matches the consumer's need, not an arbitrary convention.

### 4.3 Self-Bootstrapping

The CLI never requires a separate initialization step. The first command that touches storage creates the database, runs migrations, and sets up the schema automatically. `oraculo epic init my-project` on a fresh codebase does everything — no `oraculo init` prerequisite.

This matters because agents cannot reliably remember to run setup commands. If the tool requires setup, it will eventually be called without it, and the agent will waste a turn on an error.

### 4.4 Idempotent Operations

Every write operation is safe to repeat. `oraculo epic init my-project` called twice does not fail — it returns `"created": false` and moves on. This aligns with how agents work: they retry, they re-run, they sometimes forget they already did something. The CLI tolerates this.

### 4.5 Strict Validation

The CLI validates preconditions and refuses invalid operations. It is not permissive — it is a guardian of data integrity.

If a command requires an existing epic, and the epic does not exist, the CLI rejects the call with a clear error. It does not create a placeholder, guess the intent, or proceed with partial data. Invalid input produces an error, never a corrupted state.

This is Design by Contract in practice: the CLI enforces the rules so that consumers do not need to.

### 4.6 Fail Loudly, Recover Quietly

Errors return structured JSON with a non-zero exit code. The agent sees a clear error message and adapts. Successful operations return structured results. There is no ambiguity — the exit code tells the consumer whether to proceed or adjust.

## 5. Independence

The CLI functions standalone. It does not depend on being inside a Claude Code session, a specific shell environment, or any particular runtime context. This is a philosophical principle, not a practical convenience.

Agents use the CLI within skills. Humans use it at the terminal. CI/CD pipelines can use it for automation. The same binary, the same interface, the same contracts — regardless of who or what is calling it.

This independence ensures that the Trust Layer is always available. The CLI's reliability does not depend on the reliability of anything around it.

## 6. Evolution Model

The CLI is opinionated and versioned. It knows Oraculo's operating model — the phases, the artifacts, the schemas — and validates against it. The CLI is not a generic data store; it understands the domain it serves.

When the operating model evolves, the CLI evolves with it. Changes to artifact structures, new phases, new validation rules — these become new CLI versions with explicit migration paths.

There is no plugin system or extension mechanism. The CLI grows with Oraculo as a single, cohesive tool. Currently it covers the Discover phase; as Oraculo matures, it will cover all five phases: Discover, Plan, Execute, Validate, and Deliver.

## 7. What the CLI Is Not

The CLI is **not a general-purpose project management tool**. It does not replace Jira, Linear, or any external tracker. It is Oraculo's internal infrastructure — the nervous system that connects skills, agents, and persistent storage.

The CLI is also **not the entry point for the Socratic workflow**. Users start product discovery through skills (`/oraculo:epic`, `/oraculo:story`), which provide the guided, conversational experience. The CLI is the infrastructure that skills and agents rely on — universal in access, but not a substitute for the Socratic process.
