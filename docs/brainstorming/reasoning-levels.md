# Brainstorming Phase — Reasoning Levels

## 1. Two Modes, Same Discipline

The Brainstorming phase operates at two reasoning levels. They share the same Socratic discipline — reframe, diverge, surface assumptions, gate before advancing — but differ in **domain of inquiry** and **epistemic authority**.

- **Light Reasoning** — The user is an executor (typically a developer) who received a task defined by others. They have technical expertise but limited product context. The brainstorming focuses on: does the user understand the task correctly? What are the technical risks? What edge cases exist in the system?

- **Deep Reasoning** — The user is a decision-maker (typically a PM, designer, or tech lead) with full context: user data, business metrics, strategic vision. The brainstorming covers all dimensions — user, business, and technical — with the complete framework toolkit.

**Key distinction:** Light is not a degraded Deep. It is a reorientation. The axis shifts from "user and business" to "code and system." The Socratic principles are invariant. What changes is the domain of the questions and who has the authority to answer them.

## 2. How Reasoning Level is Selected

Oraculo determines the reasoning level at session entry based on signals from the user:

| Signal | Light | Deep |
|--------|-------|------|
| User describes a task received from someone else | Yes | — |
| User describes an idea or problem they identified | — | Yes |
| User has business metrics or user data to share | — | Yes |
| User focuses on implementation approach | Yes | — |
| User's role is primarily technical | Yes | — |
| User's role involves product decisions | — | Yes |

When ambiguous, Oraculo asks: "Are you exploring an idea you want to validate, or implementing something that's already been decided?"

The user can also explicitly request a level, and Oraculo can suggest switching mid-session if the conversation reveals the user has more (or less) context than initially assumed.

## 3. Theoretical Foundations by Level

### Frameworks common to both levels

| Framework | Application |
|-----------|-------------|
| **Double Diamond** (diverge/converge) | Both levels diverge before converging. Light diverges in the technical space; Deep diverges in the product space. |
| **Assumption Mapping** | Both levels surface and score hidden assumptions. The categories differ (see section 4). |
| **Four Forces of Progress** | Both levels map forces of adoption and resistance. Light applies them to technical adoption (system change); Deep applies them to user adoption (behavior change). |

### Frameworks exclusive to Deep Reasoning

| Framework | Why Deep only |
|-----------|---------------|
| **JTBD — full dimensions** (functional, emotional, social) | Requires user research data and behavioral context that executors typically don't have. |
| **OST — construction** | Building the Opportunity Solution Tree requires knowledge of business outcomes and mapped user opportunities. Executors receive tasks already connected to an opportunity. |
| **Working Backwards (PR/FAQ)** | Writing a fictional press release requires clarity about user value and product positioning that executors can't provide. |

### Frameworks with different emphasis

| Framework | Light emphasis | Deep emphasis |
|-----------|---------------|---------------|
| **Job Stories** | System Stories: "When [system state/event], the system should [behavior], so that [observable/testable outcome]." | User Stories: "When [user situation], I want [motivation], so I can [expected outcome]." |
| **OST — consultation** | Queries existing tree to check if the task connects to a mapped opportunity. Passive use. | Builds and extends the tree with new opportunities, solutions, and experiments. Active use. |
| **Four Forces** | Push: "What technical limitation is driving this task?" Pull: "What makes this approach better architecturally?" Anxiety: "What can break for other systems?" Habit: "What code patterns must be preserved?" | Push: "What frustrates the user enough to motivate change?" Pull: "What makes this attractive?" Anxiety: "What makes users hesitate?" Habit: "What current behavior does this disrupt?" |

## 4. Session Flow by Level

### 4.1 Entry — Reframing

| Aspect | Light | Deep |
|--------|-------|------|
| **Trigger questions** | "You received this task. What's the problem it solves, as described by whoever assigned it?" / "How do you know you understood what was asked? What system behavior will change?" / "What is explicitly out of scope?" | "You described what you want to build. What problem does this solve for the user?" / "If this feature didn't exist, what would the user do?" / "What is explicitly out of scope?" |
| **Exit condition** | User can articulate the expected system behavior change independently from the implementation approach. | User has articulated a problem statement independent from any specific solution. |

### 4.2 Divergence

| Aspect | Light | Deep |
|--------|-------|------|
| **JTBD questions** | "What part of the code handles this case today? Why is it insufficient?" / "What pre-conditions must exist for this behavior to trigger?" | Full JTBD: functional, emotional, social dimensions. |
| **Four Forces** | Technical forces (see section 3 table). | User adoption forces (see section 3 table). |
| **Edge cases** | "What happens with invalid inputs, concurrent state, or network errors?" / "What other services, queues, or jobs depend on this component?" | "What happens if the user does [unexpected behavior]?" / "Who could be negatively affected even if they're not the target?" |
| **Exit condition** | At least: technical edge cases explored, system dependencies identified, one risk the user hadn't considered surfaced. | At least: three JTBD dimensions explored, all four forces mapped, one edge case the user hadn't considered surfaced. |

### 4.3 Codebase Analysis (Parallel)

Identical in both levels. In Light, this step carries **higher weight** — it is the center of gravity of the session, not a parallel complement. Results enter the conversation earlier and with more influence on subsequent questions.

### 4.4 Convergence

| Aspect | Light | Deep |
|--------|-------|------|
| **Questions** | "What is the core functional requirement — what must change in the system?" / "Describe the expected system behavior in one sentence." / "What observable behavior changes after deploy?" / "If you could solve only one aspect, which delivers the most value?" | "What is the core pain — root cause, not symptom?" / "Write the problem in one sentence from the user's perspective." / "What business metric does this affect?" / "If you could solve only one aspect, which delivers the most value?" |
| **Exit condition** | Clear system behavior statement + observable verification criteria. | Clear problem statement + target outcome (measurable metric or user behavior change). |

### 4.5 Assumption Mapping

| Category | Light | Deep |
|----------|-------|------|
| **Problem hypothesis** | Recorded as "external dependency" — the PM validated this. Not re-analyzed by the dev. | Fully explored and scored. |
| **Solution hypothesis** | Evaluated: "Does this implementation actually solve the described use case?" | Fully explored and scored. |
| **Value hypothesis** | Recorded as "external dependency" — outside the dev's scope to validate. | Fully explored and scored. |
| **Feasibility hypothesis** | **Central focus.** Architecture fit, dependency risks, regression potential, performance implications. | Explored and scored alongside other categories. |

Risk scoring mechanism is identical in both levels.

### 4.6 Stress Test

| Aspect | Light | Deep |
|--------|-------|------|
| **Method** | Technical Viability Checklist (replaces PR/FAQ) | PR/FAQ (Working Backwards) |
| **Questions** | "Does the chosen approach fit the current architecture without major refactoring?" / "What tests must be written to cover the mapped edge cases?" / "What would another dev's main objection be when reviewing this PR?" / "How do you verify this works on deploy day? What are the technical acceptance criteria?" | "If this launched tomorrow, how would the announcement read?" / "What would a skeptical customer ask?" / "What would engineering's main objection be?" / "How would you measure success in three months?" |
| **Exit condition** | User answers all four without contradiction. | User answers all four without hesitation or contradiction. |

### 4.7 Exit Gate

The four risks are evaluated in both levels, but with different authority and escalation behavior:

| Risk | Light | Deep |
|------|-------|------|
| **Value** | "Is the expected behavior described in the task clear enough to implement?" If unclear: **escalate to PM** (not loop back). | "Is the user value clear and the need validated?" If unclear: return to Divergence. |
| **Usability** | "Is the solution consistent with codebase patterns and maintainable?" If gap: return to edge case exploration. | "Is the solution understandable and usable without excessive friction?" If gap: return to Divergence. |
| **Technical feasibility** | Full authority. Identical to Deep. If gap: return to Codebase Analysis. | If gap: return to Codebase Analysis. |
| **Business viability** | "Are the technical acceptance criteria defined and testable?" If not: **escalate to PM** (not loop back). | "Are success metrics defined and business impact understood?" If gap: return to Convergence. |

**Escalation behavior (Light only):** When the dev cannot answer a gate question because it requires product context, Oraculo does not loop the dev back through brainstorming steps. Instead, it generates a structured "Questions for PM" artifact — specific questions the PM must answer before the dev can proceed. The gate opens only after the PM responds or the team explicitly accepts the risk.

## 5. Output Artifacts by Level

| Artifact | Light | Deep |
|----------|-------|------|
| **Job Stories** | System Stories: "When [system state/event], the system should [behavior], so that [observable/testable outcome]." | User Job Stories: "When [situation], I want [motivation], so I can [outcome]." |
| **OST** | Shallow tree starting from the received requirement: Requirement → Solution → Feasibility Assumption. No Outcome node (outside dev's scope). | Full tree: Outcome → Opportunity → Solution → Assumption. |
| **Assumption Register** | Filtered: only Solution and Feasibility categories scored. Problem and Value categories recorded as "external dependency — outside dev scope to validate." | Complete: all four categories scored and tracked. |
| **Codebase Impact Summary** | **Expanded** — this is the primary artifact. Includes: affected components, interface contracts that change, existing tests that may break, relevant past technical decisions, implementation complexity estimate. | Standard: components, constraints, edge cases, past decisions. |
| **Questions for PM** (Light only) | Generated when the exit gate identifies gaps the dev cannot fill. Structured list of specific questions with context from the brainstorming session. | Not applicable — the Deep user has the authority to answer these questions. |

## 6. Interaction with Complexity Scaling

Reasoning level (Light/Deep) and complexity (Minimal/Standard/Deep) are **two independent axes**. They combine into a matrix:

|  | Minimal | Standard | Deep Complexity |
|--|---------|----------|-----------------|
| **Light** | Reframing + Feasibility assumptions + adapted exit gate. 3-5 technical questions. | Full Light flow: Reframing → Edge Cases → Codebase Analysis → Technical Convergence → Assumption Mapping → Viability Checklist → Exit Gate. | Full Light flow with parallel sub-agents analyzing codebase across multiple domains. Long session, but no business questions. |
| **Deep** | Reframing + basic assumption check + full exit gate. 3-5 questions. | Full flow as described in the main design document. All frameworks, all questions, all gates. | Full flow with parallel sub-agents for codebase, feasibility, and past decisions review. Maximum duration and depth across all dimensions. |

**Special case — Light + Deep Complexity:** A developer receives a task that seems simple but affects multiple codebase domains (e.g., changing a shared model). Oraculo selects Deep Complexity due to the technical breadth of impact, but maintains Light Reasoning because the dev lacks business context. The result is a long session with many technical questions and multiple sub-agents, without any questions about metrics or users.
