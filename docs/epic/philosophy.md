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

### 3.3 Opportunity Solution Tree (OST) — Structure the Exploration

Teresa Torres' framework provides the structural backbone: desired Outcome at the top, branching into Opportunities (real user needs), then potential Solutions, then Experiments to validate assumptions. This tree prevents solutions from appearing without evidence-based anchoring.

**Why it matters:** Without structure, exploration becomes a wandering conversation. The OST keeps every idea connected to a real user need and a measurable outcome. Solutions that can't trace back to an opportunity have no foundation.

### 3.4 Assumption Mapping — Prioritize What Could Kill the Idea

Every idea carries hidden assumptions. Assumption Mapping makes them explicit and scores them by risk: impact if wrong multiplied by lack of evidence. The riskiest assumptions become the priority before any investment.

**Why it matters:** Teams don't fail because they lack ideas. They fail because they build on unvalidated beliefs. Making assumptions explicit — and prioritizing the most dangerous ones — is the difference between disciplined exploration and wishful thinking.

### 3.5 Working Backwards (PR/FAQ) — Stress Test the Idea

Amazon's technique of writing a fictional press release before building forces clarity about who benefits, what changes, and why it matters. If the announcement doesn't sound compelling, the idea isn't ready.

**Why it matters:** Describing the end result as if it already exists forces the user to confront gaps in their thinking that abstract discussion misses. It answers: "Would anyone care about this?"

### 3.6 Four Forces of Progress — Map Behavioral Edge Cases

The JTBD Four Forces model maps what drives and resists adoption: Push (frustration with current state), Pull (attraction of the new), Anxiety (fear of the unknown), and Habit (inertia of current behavior). This surfaces edge cases that purely functional analysis misses.

**Why it matters:** A solution can be technically perfect and still fail because users are anxious about change or attached to their current habits. These forces reveal the human side of adoption that feature lists ignore.

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

- **Gate before advancing** — The Epic phase has an exit gate. If the user cannot articulate the problem clearly, define success criteria, and acknowledge critical assumptions, Oraculo does not advance to Plan. This is not bureaucracy — it is quality.

## 6. Relationship to Stories

The Epic phase produces a Requirements Document with REC-N requirements. Each REC-N is a candidate for decomposition into one or more stories via `/oraculo:story`. The Epic defines the **problem space**; stories define **executable work** within that space.

An epic without stories is a validated problem definition waiting to be decomposed. A story without an epic is a standalone piece of work that doesn't need the full exploration depth.
