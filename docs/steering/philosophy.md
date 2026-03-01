# Steering — Philosophy

## 1. Purpose

Steering exists to give Oraculo the project's identity before any action is taken.

Agents and skills operate in a vacuum without Steering. They know how to write code, explore problems, validate implementations — but they don't know **for whom**, in what **context**, or with which **constraints** the project operates. Steering fills this vacuum with two foundational documents: **Product** (what and for whom) and **Design** (how and with what).

Without Steering, Oraculo asks generic questions. With Steering, every Epic question is informed by the company's domain, the product's vision, the existing architecture, and the team's conventions. **The quality of exploration is proportional to the quality of context.**

## 2. Core Belief

**An agent without product context is a generic programmer — competent, but disconnected from the real problem.**

When someone asks "add authentication," the correct answer depends on who uses the product (developers integrating via API? end consumers?), what sector the company operates in (fintech with regulatory compliance? B2B SaaS?), and what technical decisions have already been made (JWT? sessions? OAuth?). Without this context, Oraculo produces technically correct but potentially misaligned solutions.

Steering is the difference between an agent that asks "what type of authentication do you want?" and one that asks "you already use JWT for the public API — should this new module's authentication follow the same pattern or is there a reason to diverge?"

## 3. Theoretical Foundations

Steering draws from established principles of context, domain modeling, and architectural discipline.

### 3.1 Context-Driven Testing (Cem Kaner) — Context Governs All Decisions

No universal best practices — only practices adequate to a specific context. The same decision that is correct for a 5-person startup is irresponsible for a regulated financial system. Steering makes context explicit so every Oraculo decision is informed by the project's reality.

**Why it matters:** Without explicit context, agents default to generic heuristics. A code agent choosing between a simple validation and a full compliance pipeline has no basis for the decision unless it knows the project's domain, regulatory environment, and risk tolerance. Steering provides that basis.

### 3.2 Bounded Context (Eric Evans, DDD) — Language Has Boundaries

Each system operates within boundaries of language and meaning. "Payment" means something different for checkout, reconciliation, and accounting. Steering captures these boundaries: Product defines business domain language; Design defines technical domain language. Both prevent Oraculo from confusing terms or producing code with incorrect abstractions.

**Why it matters:** When an agent encounters an ambiguous term, it resolves it with whatever interpretation seems plausible. If "user" means both "end consumer" and "API integrator" in the same project, the agent will conflate them unless Steering draws the boundary. Domain language misalignment produces code that compiles but models the wrong reality.

### 3.3 Separation of Concerns — Two Documents for Two Domains

The decision to split Steering into two documents — Product and Design — reflects the fundamental separation between the problem domain (what, for whom, why) and the solution domain (how, with what, where). Mixing both in a single document produces an artifact no audience reads comfortably.

**Why it matters:** Product people maintain vision, personas, and business constraints. Engineers maintain architecture, conventions, and technical decisions. Two documents, each with its audience and purpose, ensure both are maintained. A monolithic context document decays because no single person owns all of it.

## 4. What Makes Oraculo's Steering Unique

Steering is not generic project documentation. It is a structured context layer designed for agent consumption with four specific properties.

### 4.1 Mandatory Reading, Not Optional

Steering is not reference documentation the agent consults if it wants to. It is **mandatory context loaded before every Epic and every Story.** The skills read Steering as a precondition — if it doesn't exist, the skill informs and offers creation. The agent never operates without knowing where it is.

### 4.2 Complementary to CLAUDE.md, Not Redundant

CLAUDE.md contains operational instructions for Claude Code. Steering contains project identity. They are different layers. **CLAUDE.md says "how to work here." Steering says "what is here."** Both are read, neither replaces the other.

### 4.3 Hybrid Creation — Template with Validation

Oraculo doesn't require the user to write from scratch, nor does it auto-generate. It offers structured templates, the user fills them, then Oraculo validates and enriches with complementary questions. **The result reflects the user's knowledge, not the agent's imagination.** What the user wrote is treated as the source of truth.

### 4.4 Two Documents, Two Domains, Two Cadences

Product and Design change at different rates. Product vision and company mission rarely change. Tech stack and code conventions evolve with the project. **Separating into two documents allows each to be updated at its natural cadence.**

## 5. Principles Specific to Steering

These principles extend the core Oraculo principles for this specific layer:

- **Steering before everything** — No Epic or Story starts without Steering documents being read. If they don't exist, Oraculo informs and offers creation. Never silently ignores the absence.

- **The user is the source, never the agent** — Steering captures what the user knows about their project. Oraculo can question gaps, suggest a section is incomplete, or ask for clarification — but never fills content on its own. A Steering document with invented information is worse than no document.

- **Stable, not immutable** — Steering is a living document, not a constitution. It changes when project reality changes. But changes are deliberate and explicit, never silent. Oraculo does not modify Steering as a side effect of an Epic.

- **Context, not instruction** — Steering informs, it does not command. It says "the company operates in the financial sector" and "the API uses JWT" — it doesn't say "always use JWT" or "never do X." Business principles are constraints, not absolute rules. Epic and Story can challenge a principle if there's reason — but Steering ensures the challenge is conscious, not accidental.

- **Visible gap is better than fiction** — If the user doesn't know the product vision or hasn't defined code conventions, the section stays empty or marked as "to be defined." Oraculo works with what it has. An empty field is an honest signal; an invented field is a trap.

## 6. Relationship to Other Phases

Steering is a cross-cutting layer — not a phase of the operating model, but a prerequisite that feeds all of them.

### 6.1 Epic — Greatest Impact

Exploration starts from informed questions instead of generic ones. The Product document gives the Epic phase domain language, user personas, and business constraints. The Design document gives it architectural awareness. Together, they transform Socratic exploration from "tell me about your users" to "your B2B customers integrate via API — how does this new feature affect their integration contracts?"

### 6.2 Story — Compensates the Context Gap

Standalone stories — those created without an Epic — lack the deep exploration that produces rich context. Steering compensates by providing the baseline context the story would otherwise miss entirely. The story skill reads Steering to understand what the project is before asking what the task is.

### 6.3 Plan and Execute — Feeds the Orchestrator and Code Agents

The Design document feeds the orchestrator with architecture for decomposition and code agents with conventions, project structure, and technical decisions. An orchestrator that knows the project uses a hexagonal architecture decomposes work differently than one that assumes a monolithic controller layer.

### 6.4 Validate — Broader Validation Lens

The QA agent gains context beyond the immediate diff. Product-level awareness — compliance requirements, domain invariants, user expectations — extends validation beyond "does it pass tests" to "does it belong in this product." A QA agent that knows the project operates in healthcare will flag an unencrypted PII field that a context-free agent would approve.

### 6.5 CLAUDE.md — Coexist Without Overlap

CLAUDE.md is operational instructions. Steering is project identity. They are independent documents consumed at different moments. CLAUDE.md tells the agent how to run tests, what linter to use, and where files go. Steering tells the agent what the product does, who it serves, and how the architecture is organized. Neither document should contain content that belongs in the other.

## 7. What Steering Is Not

- **Not a substitute for the Epic.** Steering is stable context; Epic is pointed investigation. Steering says "our users are healthcare providers." Epic asks "what specific workflow breaks when a provider has 200+ patients?"

- **Not a technical specification.** Design describes architecture at a high level — patterns, conventions, boundaries — not full ADRs, API contracts, or database schemas. Those belong in the codebase itself.

- **Not generated by Oraculo.** Content comes from the user; Oraculo validates, never invents. The agent may ask "this section seems incomplete — would you like to add details about your deployment strategy?" It never fills the answer itself.

- **Not versioned with repository code.** Steering lives in `.oraculo/steering/`, operational infrastructure alongside the SQLite database. It can be committed by team decision, but Oraculo doesn't assume or require it.

- **Not required to use Oraculo.** A project without Steering works — Oraculo operates with less context and asks more questions. The absence is informed, never silenced. The skill tells the user Steering doesn't exist and offers to help create it.
