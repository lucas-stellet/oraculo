# Oraculo Steering — Design

## 1. Overview

Steering is Oraculo's context layer — two markdown documents that capture product identity and technical identity. They live in `.oraculo/steering/` and are read mandatorily by `/oraculo:epic` and `/oraculo:story` before any exploration or definition begins.

Two documents, two domains:

- **`product.md`** — Company, product, vision, business principles
- **`design.md`** — Tech stack, architecture, code conventions, project structure

```
.oraculo/
├── oraculo.db          # Operational state + knowledge
├── config              # Dashboard port, project settings
└── steering/
    ├── product.md      # Product identity
    └── design.md       # Technical identity
```

Creation follows a hybrid model: Oraculo offers structured templates, the user fills them, Oraculo validates completeness and asks questions about gaps — it never invents content. The user is the sole author. Oraculo is the validator.

Steering is complementary to CLAUDE.md. CLAUDE.md contains operational instructions — how to run the project, which commands to use, what conventions to follow in code. Steering contains project identity — what the product is, who it serves, where it is going, and how it is built. Both are read by skills and agents. Neither replaces the other.

## 2. Document Structure

### 2.1 Product Document (`product.md`)

Four mandatory sections:

| Section | Content | Purpose |
|---|---|---|
| **Company** | Name, sector, mission/purpose, size/stage | Calibrates Oraculo for the business domain |
| **Product** | Product name, purpose (one sentence), target users, problem it solves, current state (production/development/stage) | Defines what exists and for whom |
| **Vision** | Where the product wants to go — not a roadmap, the direction that guides prioritization. A paragraph or two sentences. Optionally includes time horizon | Anchors exploration and prioritization |
| **Business Principles** | Non-negotiable rules governing product decisions. Constraints, not instructions. Each has name + short description | Treated as guardrails during exploration |

### 2.2 Design Document (`design.md`)

Four mandatory sections:

| Section | Content | Purpose |
|---|---|---|
| **Tech Stack** | Languages, frameworks, database, infra, CI/CD | Code agents produce code in the right language and tools |
| **Architecture** | Architectural pattern, component communication, storage, relevant external integrations | Orchestrator decomposes work correctly |
| **Code Conventions** | Error handling, naming, test style, commit format, project-specific conventions | Code agents follow these; QA agents validate against them |
| **Project Structure** | Directory organization and where each type of code lives. Can be a directory tree with short descriptions | Agents know where to create files and find existing code |

### 2.3 Optional Sections (Future)

- **Product:** Market context, success metrics
- **Design:** Technical decisions (simplified ADR), technical constraints

Optional sections are never required. Oraculo works with what exists.

## 3. Templates

Oraculo distributes templates for both documents. Templates define the expected structure without filling content — each section has a brief description and a placeholder.

### 3.1 Product Template

```markdown
# Steering — Product

## Company
<!-- Name, sector, mission/purpose, size/stage -->

## Product
<!-- Product name, purpose (what it does), target users,
     problem it solves, current state -->

## Vision
<!-- Where the product wants to go. Not a roadmap — the direction
     that guides prioritization. -->

## Business Principles
<!-- Non-negotiable rules governing product decisions.
     Format: name + short description. -->
```

### 3.2 Design Template

```markdown
# Steering — Design

## Tech Stack
<!-- Languages, frameworks, database, infra, CI/CD -->

## Architecture
<!-- Architectural pattern, component communication,
     storage, external integrations -->

## Code Conventions
<!-- Error handling, naming, test style, commit format,
     project-specific conventions -->

## Project Structure
<!-- Directory organization. Can be a tree with
     short descriptions of each main directory. -->
```

### 3.3 Template Location

Templates live inside the distributable Oraculo kit:

```
claude-kit/
└── skills/
    └── oraculo/
        └── steering/
            └── references/
                ├── product-template.md
                └── design-template.md
```

The Steering skill reads the template from its `references/` folder, copies it to `.oraculo/steering/`, and informs the user. From there, the user fills and Oraculo validates.

## 4. Creation and Validation Flow

### 4.1 When Steering Is Created

Two paths:

- **Explicit** — User invokes a Steering creation workflow. Templates are copied, validation starts.
- **Implicit** — User invokes `/oraculo:epic` or `/oraculo:story` and Steering does not exist. The skill detects the absence, informs the user, and offers creation. The user can accept (inline creation) or decline (skill continues with less context).

### 4.2 Hybrid Creation Flow

Three steps:

1. **Template** — Copy template to `.oraculo/steering/`.
2. **Filling** — User fills sections (outside Oraculo or dictated in session). Content is always from the user.
3. **Validation** — Oraculo reads the filled document and checks completeness per section:
   - **Filled:** Confirm and proceed.
   - **Empty/placeholder:** Ask if the user wants to fill now or leave as an explicit gap.
   - **Ambiguous/superficial:** Ask complementary questions to deepen.

Oraculo never rejects a document for being incomplete. Gaps are accepted — what matters is that they are visible.

### 4.3 Updates

Steering is updated manually by the user. Oraculo does not modify documents as a side effect. If divergence is detected during exploration, Oraculo informs but does not self-correct.

### 4.4 Flow Diagram

```
/oraculo:epic or /oraculo:story invoked
        │
        ▼
Steering exists in .oraculo/steering/?
   ├── YES → Read product.md and design.md → Continue skill
   └── NO  → Inform user
                │
                ▼
        User wants to create now?
           ├── YES → Copy templates → Fill → Validate → Continue skill
           └── NO  → Continue without Steering (less context)
```

## 5. Integration with Skills

### 5.1 How Skills Read Steering

Skills read via CLI: `oraculo tools steering get --doc product` and `oraculo tools steering get --doc design`. The CLI returns raw markdown (content command convention). If the file does not exist, the CLI returns an error and the skill follows the detection flow from Section 4.1.

### 5.2 What Changes in Epic with Steering

Steering does not alter frameworks (Double Diamond, JTBD, etc.) — it enriches the starting point. Without Steering, questions are generic. With Steering, questions are informed and domain-calibrated.

Steering is injected as a reference block. The skill uses it to:

- Calibrate questions to the business domain and user base
- Check for conflicts with business principles
- Reference existing architecture when exploring technical feasibility
- Avoid asking questions already answered by the steering documents

### 5.3 What Changes in Story with Steering

Stories — especially standalone stories outside an epic — arrive with less context. Steering compensates more acutely. The skill cross-references the user's request with project context: tech stack, architectural constraints, business principles. A story that says "add a payment webhook" is interpreted differently when Steering says the product uses Stripe vs. when it says the product uses a custom billing system.

### 5.4 What Changes in Plan and Execute

The orchestrator reads Design before decomposing work:

- **Architecture** informs task structure — monolith vs. services determines how tasks are split.
- **Conventions** inform code agent instructions — naming, error handling, test style are passed as constraints.
- **Project structure** informs file placement — agents know where to create files without guessing.

Design is passed as context to each code agent alongside the task description and relevant files.

### 5.5 What Changes in Validate

The QA agent receives Design as part of its validation context. Beyond verifying test results, QA validates:

- Code convention compliance (naming, error handling, test style)
- Architecture alignment (correct layers, correct communication patterns)
- Business principle implications (when the change has product-facing consequences)

## 6. CLI Commands

### 6.1 Commands

| Command | Output | Purpose |
|---|---|---|
| `oraculo tools steering init` | JSON | Copy templates to `.oraculo/steering/`. Idempotent |
| `oraculo tools steering init --doc product` | JSON | Copy only Product template |
| `oraculo tools steering init --doc design` | JSON | Copy only Design template |
| `oraculo tools steering get --doc product` | Raw content | Return `product.md` content. Error if file does not exist |
| `oraculo tools steering get --doc design` | Raw content | Return `design.md` content. Error if file does not exist |
| `oraculo tools steering status` | JSON | State of each document: exists or not, filled vs. empty sections |

### 6.2 Behavior

**`steering init`** follows the idempotent pattern. Calling it when templates already exist does not overwrite:

```json
{
  "product": { "created": true, "path": ".oraculo/steering/product.md" },
  "design": { "created": true, "path": ".oraculo/steering/design.md" }
}
```

Second call:

```json
{
  "product": { "created": false, "path": ".oraculo/steering/product.md" },
  "design": { "created": false, "path": ".oraculo/steering/design.md" }
}
```

**`steering get`** returns raw markdown — the content command convention. No JSON wrapping. The consumer reads the document as-is.

**`steering status`** reports per-section state:

```json
{
  "product": {
    "exists": true,
    "sections": {
      "Company": "filled",
      "Product": "filled",
      "Vision": "empty",
      "Business Principles": "filled"
    }
  },
  "design": {
    "exists": true,
    "sections": {
      "Tech Stack": "filled",
      "Architecture": "filled",
      "Code Conventions": "empty",
      "Project Structure": "empty"
    }
  }
}
```

**Detection logic:** The CLI checks for non-comment text below each heading. A section with only HTML comments or whitespace is reported as `"empty"`. A section with any other content is `"filled"`. This is a simple heuristic — no semantic analysis.

### 6.3 Integration with `oraculo install`

`oraculo install` does **not** automatically create Steering documents. Steering is opt-in. This keeps install lean and avoids creating files the user may not want to fill immediately.

## 7. Future Work

Deferred capabilities for later iterations:

- **Dedicated skill (`/oraculo:steering`)** — Socratic session for guided creation, similar to `/oraculo:epic`
- **Divergence detection** — Detect contradictions between Steering and user statements during Epic/Story, offer inline update
- **Optional Product sections** — Market context, success metrics
- **Optional Design sections** — Technical decisions (simplified ADR), technical constraints
- **Steering versioning** — Track changes with timestamps and reasons
- **CLI validation command** — `oraculo tools steering validate` for coherence checking
- **Import from existing context** — Populate from CLAUDE.md, README.md, `package.json` as suggestions for user validation
