# Artifact Templates

Exact output formats for the Brainstorming phase. Each artifact adapts to the reasoning level (Light/Deep).

---

## 1. Job Stories

### Deep — User Job Stories

```
When [specific user situation/context],
I want [motivation/job to be done],
so I can [expected measurable outcome].
```

**Example:**
> When I'm reviewing pull requests with 500+ lines of changes, I want to see a structured summary of what changed and why, so I can give meaningful feedback in under 15 minutes instead of 45.

### Light — System Stories

```
When [system state/event/trigger],
the system should [expected behavior],
so that [observable/testable outcome].
```

**Example:**
> When a webhook payload exceeds the 256KB size limit, the system should reject it with a 413 status and log the payload size, so that downstream consumers are never blocked by oversized messages.

---

## 2. Opportunity Solution Tree (OST)

### Deep — Full Tree

```
Outcome: [business metric or user behavior change]
└── Opportunity: [user need, written in first person]
      ├── Solution A: [proposed approach]
      │     ├── Assumption: [hypothesis]
      │     │     └── Validation: [how to test]
      │     └── Assumption: [hypothesis]
      │           └── Validation: [how to test]
      └── Solution B: [alternative approach]
            └── Assumption: [hypothesis]
                  └── Validation: [how to test]
```

### Light — Shallow Tree

```
Requirement: [task as received]
└── Solution: [proposed implementation approach]
      ├── Feasibility Assumption: [hypothesis about architecture fit]
      │     └── Validation: [how to verify]
      └── Feasibility Assumption: [hypothesis about risk]
            └── Validation: [how to verify]
```

No Outcome node — outside the developer's scope.

---

## 3. Assumption Register

Format is identical for both levels. Content scope differs.

| # | Description | Category | Impact (1-5) | Evidence (1-5) | Risk Score | Status |
|---|-------------|----------|:------------:|:--------------:|:----------:|--------|
| 1 | [assumption text] | Problem / Solution / Value / Feasibility | N | N | Impact x (6 - Evidence) | unvalidated / validated / refuted |

### Scoring

- **Impact** (1-5): If wrong, how badly does it break the idea?
- **Evidence** (1-5): How much data supports this? (1 = pure intuition, 5 = validated data)
- **Risk Score** = Impact x (6 - Evidence). Higher = higher priority.

### Light-specific rules

- **Problem** and **Value** categories: recorded as "external dependency — outside dev scope to validate"
- **Solution** and **Feasibility** categories: fully scored
- Critical feasibility assumptions become explicit inputs for the Plan phase

### Deep-specific rules

- All four categories are fully explored and scored
- Critical assumptions (high risk) may generate validation tasks before implementation

---

## 4. Codebase Impact Summary

### Deep — Standard

```markdown
### Codebase Impact Summary

**Related components:**
- [component]: [how it relates to the idea]

**Architectural constraints:**
- [constraint]: [implication for the solution]

**Edge cases from existing code:**
- [edge case]: [how the new feature must handle it]

**Relevant past decisions:**
- [decision]: [original reasoning] — [still applies? / needs revisiting?]
```

### Light — Expanded (primary artifact)

```markdown
### Codebase Impact Summary (Expanded)

**Affected components:**
- [component]: [current responsibility] → [what changes]

**Interface contracts that change:**
- [interface/API]: [current contract] → [new contract]

**Existing tests that may break:**
- [test file/suite]: [what it covers] — [why it may break]

**Relevant past technical decisions:**
- [decision]: [original reasoning] — [still applies? / needs revisiting?]

**Implementation complexity estimate:**
- Scope: [single component / multiple components / cross-cutting]
- Risk areas: [list]
- Suggested approach: [brief]
```

---

## 5. Questions for PM (Light only)

Generated when the Exit Gate identifies gaps the developer cannot fill.

```markdown
### Questions for PM

Context: During brainstorming for "[task description]", the following gaps
were identified that require product context to resolve.

1. **[Risk dimension]:** [specific question with context from the session]
   - Why it matters: [what is blocked without this answer]

2. **[Risk dimension]:** [specific question]
   - Why it matters: [what is blocked]

Session reference: [link or identifier to the brainstorming session]
```

The Exit Gate remains open until these questions are answered or the team explicitly accepts the risk.
