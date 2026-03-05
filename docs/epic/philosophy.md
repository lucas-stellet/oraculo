# Epic Phase — Philosophy

## 1. Purpose

The Epic phase exists to transform raw ideas into validated problem definitions.

Users rarely arrive with a fully formed understanding of what they need. They come with intuitions, half-formed ideas, or solutions disguised as problems. The Epic phase is where Oraculo fulfills its most fundamental role: **the Socratic guide that instigates deeper thinking before any action is taken.**

The Epic phase is designed for work that is large, uncertain, or spans multiple areas. It requires deep exploration before the problem space can be narrowed to executable units of work (stories).

## 2. Core Belief

**The user's first idea is almost never the right problem to solve.**

When someone says "I want a button that does X," the real question is: what situation leads the user to need X? What do they do today without it? Why is that insufficient? The answer to these questions often reveals a different — and better — problem to solve.

Oraculo's job is not to accept ideas. It is to **interrogate them until they either collapse under scrutiny or emerge stronger with clear foundations.**

## 3. Theoretical Foundations

The Epic phase draws from established product discovery methodologies, adapted to Oraculo's Socratic model.

### 3.1 Double Diamond — Diverge Before You Converge

The session follows two movements: first expand the problem space (diverge), then narrow to a clear definition (converge). This prevents the most common failure in product development — jumping to solutions before understanding the problem.

**Why it matters:** Divergence forces the discovery of what the user hasn't thought about. Convergence ensures the exploration doesn't remain abstract forever. The two movements together create a discipline: understand broadly, then define precisely.

### 3.2 Jobs to Be Done (JTBD) — Uncover the Real Job

Users don't want features — they "hire" products to make progress in specific circumstances. JTBD forces the conversation beyond surface requests into the functional, emotional, and social dimensions of the user's need.

**Why it matters:** When someone asks for a solution, the real question is: what job are they trying to accomplish? In what situation? What do they use today? Why is it insufficient? The output is a Job Story, not a feature description:

> "When [specific situation], I want [motivation], so I can [expected outcome]."

**Four Forces of Progress.** JTBD includes a model of what drives and resists adoption: Push (frustration with current state), Pull (attraction of the new), Anxiety (fear of the unknown), and Habit (inertia of current behavior). These forces surface adoption edge cases that purely functional analysis misses — a solution can be technically perfect and still fail because users are anxious about change or attached to their current habits.

### 3.3 Theory of Constraints (TOC) — Focus on the Bottleneck

Every system has a single constraint that limits its throughput. TOC asks: what is the one bottleneck that, if removed, would unlock the most value? Applied to discovery, this forces prioritization — instead of exploring every opportunity equally, the session identifies which problem is the real constraint for the user or the system.

**Why it matters:** Without focus, exploration produces a broad but shallow map. TOC forces convergence toward the highest-impact problem. This same principle applies later in execution, where QA throughput governs the pace of code production.

### 3.4 Assumption Mapping — Prioritize What Could Kill the Idea

Every idea carries hidden assumptions. Assumption Mapping makes them explicit and scores them by risk: impact if wrong multiplied by lack of evidence. The riskiest assumptions become the priority before any investment.

**Why it matters:** Teams don't fail because they lack ideas. They fail because they build on unvalidated beliefs. Making assumptions explicit — and prioritizing the most dangerous ones — is the difference between disciplined exploration and wishful thinking.

### 3.5 Internal Tools, Domain Output

The frameworks above (JTBD, Four Forces, TOC, Assumption Mapping) are internal conversation tools. They guide the Socratic exploration but their terminology does not appear in the output artifact. The Requirements Document uses domain language that any product manager, developer, or QA engineer can read without knowing the frameworks.

## 4. What Makes Oraculo's Epic Exploration Unique

Traditional discovery frameworks assume a team of humans with whiteboards, sticky notes, and customer interviews. Oraculo operates on three principles that differentiate it:

### 4.1 Context-Aware Exploration

Oraculo doesn't explore in a vacuum. It explores ideas with awareness of what already exists — the product's current state, past decisions, rejected approaches, and the reasoning behind them. This prevents the team from revisiting dead ends or contradicting established decisions without awareness.

### 4.2 Accumulated Wisdom

Every epic session builds on previous ones. Oraculo carries the memory of past explorations — opportunities already mapped, assumptions already validated or refuted, patterns already identified. The exploration grows richer over time, not repetitive.

### 4.3 Multi-Angle Inquiry

For complex ideas, Oraculo can explore multiple dimensions simultaneously — technical feasibility, user behavior, business impact, edge cases — and weave the findings into a coherent Socratic dialogue. The user receives questions informed by evidence, not just intuition.

## 5. Principles Specific to Epic Exploration

These principles extend the core Oraculo principles for this specific phase:

- **Reframe before exploring** — If the user arrives with a solution, transform it into a problem before any exploration begins. Never accept "I want X" without asking "why do you need X?"

- **Diverge before converging** — Resist the urge to narrow too early. Broad questions first, focused questions second. The Double Diamond is not optional.

- **No solution talk before problem clarity** — Implementation details, approaches, and "how" questions are forbidden until the "what" and "why" are fully established.

- **Assumptions are explicit or they don't exist** — Every belief about users, behavior, or feasibility must be stated out loud. Implicit assumptions are the ones that kill projects.

- **Existing context has a voice** — What already exists — decisions, patterns, history — has a seat at the table during exploration. It constrains, enables, and informs. Ignoring it leads to plans that collide with reality.

- **Gate before advancing** — The Epic phase has two mandatory checkpoints before it can advance. The **exit gate** is an automatic, internal quality check: Oraculo will not produce the Requirements Document if the user cannot articulate the problem clearly, define success criteria, and acknowledge critical assumptions. The **version review** is a human verdict: once the artifact is produced, a version is created via `oraculo tools epic version` and submitted for review via the dashboard. The epic does not advance until a verdict arrives — `approved` to proceed or `rejected` to return to exploration. Both gates are mandatory. Neither can be skipped.

## 6. Relationship to Stories

The Epic phase produces a Requirements Document with user stories. Each user story is a candidate for decomposition into one or more executable stories via `/oraculo:story`. The Epic defines the **problem space**; stories define **executable work** within that space.

An approved epic — one that has passed both the exit gate and the human version review — is a prerequisite for decomposition into stories. Decomposition does not begin from a submitted artifact; it begins from a verdict. An epic without stories is a validated, human-confirmed problem definition waiting to be decomposed. A story without an epic is a standalone piece of work that doesn't need the full exploration depth.
