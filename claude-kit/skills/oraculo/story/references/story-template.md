# Story: {titulo}

<!--
  TEMPLATE RULES:
  - Output of the Story skill (/oraculo:story).
  - Captures a focused, executable unit of work.
  - WHAT and WHY only — never HOW.
  - Language: match the user's language.
-->

<!-- parent_epic: {nome-projeto}/epic.md -->
<!-- parent_rec: REC-N -->

---

## 1. Motivation

<!--
  <format>
    - **Problem:** 1-2 sentences describing the problem this story addresses
    - **Context:** the specific situation that triggers this need
    - If derived from an epic, reference the REC-N and quote the relevant context
  </format>
  <rules>
    - Must be from the user's perspective (Deep) or system perspective (Light)
    - Independent from any solution
  </rules>
-->

---

## 2. Requirement

<!--
  <format>
    - Single user/system story:
      - Deep: "When [user situation], I want [motivation], so I can [outcome]"
      - Light: "When [system state/event], the system should [behavior], so that [outcome]"
    - Acceptance criteria in GIVEN/WHEN/THEN (2-4 items)
    - At least one unhappy path criterion
  </format>
  <rules>
    - User language only — no implementation details
    - Observable outcomes only
    - 2-4 acceptance criteria. Fewer = vague. More = should be split.
  </rules>
-->

---

## 3. Scope

<!--
  <format>
    - **In:** what this story covers (bullet list)
    - **Out:** what is explicitly excluded and why
  </format>
  <rules>
    - Everything discussed and excluded must appear in "Out"
    - No technical scope (backend only, API first) — that's Plan phase
  </rules>
-->

---

## 4. Assumptions

<!--
  <format>
    - Bullet list of key assumptions (2-4)
    - Each with: description + evidence level (high/medium/low)
    - Flag any that need validation before implementation
  </format>
  <rules>
    - Only assumptions that emerged from the dialogue
    - Light: focus on Solution/Feasibility
    - Deep: focus on Problem/Value
  </rules>
-->
