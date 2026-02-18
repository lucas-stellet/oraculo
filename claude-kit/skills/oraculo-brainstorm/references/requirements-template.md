# Requirements: {feature_name}

<!--
  TEMPLATE RULES:
  - This document is the OUTPUT of the Discovery phase (brainstorm skill).
  - It captures WHAT and WHY — never HOW.
  - No technical implementation details: no data models, no API design, no stack choices, no architecture.
  - Technical details belong to a separate "Code Design" document produced in the Plan phase.
  - Language: match the user's language throughout.
-->

---

## 1. Summary

<!--
  <section id="summary">
    <purpose>A quick overview for anyone picking up this document without context. Should be scannable in 30 seconds.</purpose>
    <format>
      - **Feature:** one-line description of what's being built
      - **Problem:** the validated problem statement (one sentence)
      - **Expected outcome:** what changes if this succeeds (one sentence)
      - **Stakeholders:** who cares about this — users, teams, decision-makers
    </format>
    <rules>
      - Maximum 4 lines. This is an index card, not a narrative.
      - Every field must be filled — if you can't write the expected outcome in one sentence, the Discovery isn't done.
    </rules>
  </section>
-->

---

## 2. Problem

<!--
  <section id="problem">
    <purpose>Articulate the core problem in the user's own words. This is the validated problem statement that emerged from the Socratic dialogue — not the original idea, but the refined understanding.</purpose>
    <format>
      - Open with a single quote that captures the pain in first person ("Não sei para onde vai nosso dinheiro")
      - Follow with 2-4 bullet points describing the concrete consequences of this problem
      - End with relevant context: failed past attempts, current workarounds, triggers that make the problem acute
    </format>
    <rules>
      - Must describe the problem from the USER's perspective, not the builder's
      - Must be independent from any solution — if you removed the app entirely, this problem would still exist
      - No feature language ("the app should..."), no implementation hints
    </rules>
  </section>
-->

---

## 3. Context & Motivation

<!--
  <section id="context">
    <purpose>Describe who experiences this problem, when, and why it matters now. Grounds the problem in reality and prevents building for abstract personas.</purpose>
    <format>
      <subsection id="who">
        ### Quem
        - Describe the real users: their role, relationship to the problem, level of comfort with technology
        - If multiple user types exist, list each with a short description and their JTBD:
          - **[User type]:** who they are, their context, what "job" they're trying to get done
        - Do NOT use formal persona templates — use natural language that emerged from the conversation
      </subsection>
      <subsection id="when">
        ### Quando
        - List the specific situations/moments when the problem manifests
        - Each situation should be concrete and recognizable ("quando surge uma despesa inesperada", not "quando o usuário precisa do app")
      </subsection>
      <subsection id="current-alternatives">
        ### Alternativas atuais
        - What do users do TODAY without this product?
        - What's the cost of each alternative? (time, friction, errors, frustration)
        - If the answer is "nothing" — state that explicitly, it's a valid and important finding
      </subsection>
      <subsection id="why-now">
        ### Por que agora
        - What changed or what trigger makes this problem worth solving NOW?
        - This can be: a failed past attempt, accumulated frustration, a new context, a team decision, etc.
        - If there's no specific urgency, state that — "accumulated pain, no specific trigger" is a valid answer
      </subsection>
    </format>
    <rules>
      - Everything here must come from the conversation, not from assumptions
      - If the user couldn't answer "who suffers" clearly, flag it as a gap
    </rules>
  </section>
-->

---

## 4. Requirements

<!--
  <section id="requirements">
    <purpose>The core deliverable. Each requirement describes something the user needs to be able to DO, written from their perspective. These are the "contract" between Discovery and Plan — the Plan phase consumes these directly.</purpose>
    <format>
      - Each requirement has a unique ID: REC-1, REC-2, REC-3...
      - IDs are sequential and permanent — never reordered, never reused
      - Each requirement follows the structure:

        ### REC-N — {short title}

        > {user-centric statement: "Eu como [tipo de usuário], quero [ação/capacidade], para [benefício/resultado]"}

        **Contexto:** {situation from which this need emerged — links back to the "when" in Context}

        **Acceptance criteria:**
        - GIVEN {context} WHEN {action} THEN {expected observable outcome}
        - GIVEN {context} WHEN {action} THEN {expected observable outcome}

      - Group requirements by theme/area if there are many (e.g., "Registro", "Visualização", "Compartilhamento")
      - Each group can have a short intro sentence, but no group-level IDs
    </format>
    <acceptance-criteria-rules>
      - Acceptance criteria use GIVEN/WHEN/THEN format — product language, not test code
      - They describe OBSERVABLE outcomes from the user's perspective
      - "THEN the expense appears in the monthly list" — good (observable)
      - "THEN the database inserts a row with status=active" — bad (implementation detail)
      - 2-4 criteria per requirement. Fewer means the requirement is vague. More means it should be split.
      - Include at least one "unhappy path" criterion where relevant (empty state, error, edge case)
    </acceptance-criteria-rules>
    <rules>
      - User language only — "registrar uma despesa", not "POST /expenses"
      - Describe WHAT the user achieves, not HOW the system implements it
      - Each requirement must trace back to at least one situation from the Context section
      - No implementation details: no field names, no data types, no UI specifics, no technical constraints
      - No decomposition into sub-requirements — that's the Plan phase's job
      - Priority is NOT assigned here — prioritization happens in Planning with full technical context
      - Aim for 4-10 requirements. Fewer than 4 suggests the problem isn't well-explored. More than 10 suggests premature decomposition.
    </rules>
    <examples>
      <good>
        ### REC-1 — Registro rápido de despesas
        > Eu como membro da família, quero registrar um gasto no momento em que ele acontece com o mínimo de informação necessária, para manter o hábito sem que isso vire uma tarefa pesada.

        **Contexto:** O usuário já tentou controlar gastos antes e abandonou por excesso de fricção.

        **Acceptance criteria:**
        - GIVEN estou no app WHEN registro uma despesa com apenas nome, categoria e valor THEN a despesa é salva com a data de hoje
        - GIVEN estou registrando uma despesa WHEN deixo campos opcionais em branco THEN a despesa é salva sem erros
        - GIVEN acabei de registrar uma despesa WHEN olho a lista THEN vejo a despesa que acabei de registrar
      </good>
      <bad>
        ### REC-1 — Criar despesa via API
        > O sistema deve aceitar POST /expenses com campos name, category_id, amount, date (obrigatórios).

        **Acceptance criteria:**
        - WHEN POST /expenses with valid JSON THEN return 201 with expense ID
        (Motivo: linguagem técnica, descreve implementação, não descreve valor para o usuário)
      </bad>
    </examples>
  </section>
-->

---

## 5. User Experience

<!--
  <section id="user-experience">
    <purpose>Describe the key user flows and states without prescribing UI design. This section captures product decisions about behavior — what the user experiences, not how it looks.</purpose>
    <format>
      <subsection id="main-flow">
        ### Fluxo principal
        - Describe the happy path: step by step, what the user does and what they see/get back
        - Written as a narrative or numbered steps, NOT as technical sequence diagrams
        - Example: "1. O usuário abre o app → 2. Vê a lista de despesas do mês → 3. Toca em adicionar → 4. Preenche nome, categoria e valor → 5. Salva → 6. Volta para a lista atualizada"
      </subsection>
      <subsection id="empty-states">
        ### Estados vazios
        - What does the user see when there's no data yet? (first use, no expenses this month, no categories)
        - This is a product decision — the empty state is often the first impression
      </subsection>
      <subsection id="error-states">
        ### Estados de erro
        - What happens when something goes wrong? (validation error, connection lost, duplicate entry)
        - Describe from the user's perspective: what they see, what they can do about it
      </subsection>
    </format>
    <rules>
      - No wireframes or mockups — describe behavior, not layout
      - No technical flow (API calls, database queries) — describe user-visible steps only
      - If the brainstorming didn't explore UX deeply, include only what was discussed. Don't invent flows.
      - Mark uncertain flows as "to be defined in design phase"
    </rules>
  </section>
-->

---

## 6. Scope

<!--
  <section id="scope">
    <purpose>Draw explicit boundaries. What's IN for this version, what's planned for LATER, and what's deliberately OUT. Prevents scope creep and aligns expectations.</purpose>
    <format>
      <subsection id="in">
        ### Dentro (v1)
        - Bullet list of what this version covers
        - Each item should be traceable to one or more REC-N
        - Written in plain language, not feature specs
      </subsection>
      <subsection id="later">
        ### Depois (v2+)
        - Things discussed that have value but aren't needed to validate the core hypothesis
        - Each item should note WHY it's deferred, not just that it is
      </subsection>
      <subsection id="out">
        ### Fora
        - Things explicitly excluded — not deferred, but out of scope for this problem
        - For each exclusion, briefly state WHY it's out
        - This list is as important as the "in" list — it's a documented decision, not a backlog
      </subsection>
    </format>
    <rules>
      - Distinguish "later" (deferred value) from "out" (not relevant to this problem)
      - If something was discussed during brainstorming and deliberately excluded, it MUST appear in "out" or "later"
      - No technical scope (e.g., "backend only", "API first") — that's a Code Design decision
    </rules>
  </section>
-->

---

## 7. Assumptions & Risks

<!--
  <section id="assumptions-risks">
    <purpose>Surface everything we're taking for granted. Every assumption is a potential failure point. Making them explicit transforms invisible risks into manageable ones.</purpose>
    <format>
      - ALWAYS start with the scoring explanation so readers understand the numbers:

        > **Risco = Impacto × (6 − Evidência)**

        | Fator | Escala | Significado |
        |-------|:------:|-------------|
        | Impacto | 1–5 | Se a premissa estiver errada, quão grave é? (1 = irrelevante, 5 = inviabiliza o projeto) |
        | Evidência | 1–5 | Quanta evidência temos de que é verdade? (1 = pura intuição, 5 = dado validado) |
        | **Risco** | **1–25** | Quanto maior, mais perigoso. **≥ 15 = crítico** — precisa de mitigação antes de construir. |

      - Then the assumptions table with columns: #, Assumption, Category, Impact (1-5), Evidence (1-5), Risk Score, Mitigation, Status
      - Categories: Problem (is the problem real?), Value (will users care?), Viability (can we sustain this?)
      - NO "Feasibility" category here — technical feasibility is assessed in the Code Design phase
      - Mitigation: what can be done to reduce the risk BEFORE building (validation, research, prototype, conversation)
      - Status: unvalidated / validated / refuted
      - After the table, highlight critical risks (score >= 15) with a short paragraph each:
        what's the risk, what's the evidence, what's the mitigation strategy
    </format>
    <rules>
      - Every assumption must have been discussed with the user during the session
      - The agent does NOT invent assumptions — they emerge from the dialogue
      - "Validated" requires actual evidence, not just user confidence
      - Technical assumptions (architecture fit, performance, etc.) belong to Code Design, not here
    </rules>
  </section>
-->

---

## 8. Success Criteria

<!--
  <section id="success-criteria">
    <purpose>Define what "this worked" AND "this failed" looks like — from the user's and business perspective.</purpose>
    <format>
      - Table with columns: Criterion, Success indicator, Failure signal, Horizon
      - Criterion: what we're measuring (e.g., "Hábito de registro")
      - Success indicator: observable signal that it's working (e.g., "Família registra despesas pelo menos 5 dias por semana")
      - Failure signal: observable signal that it's NOT working (e.g., "Menos de 2 membros registram despesas após 2 semanas")
      - Horizon: when we expect to see this (e.g., "1 mês", "3 meses")
      - Aim for 3-5 criteria
    </format>
    <rules>
      - Indicators must be observable — something a human can verify without analytics tooling
      - No vanity metrics ("number of API calls") — focus on user behavior change
      - At least one criterion must directly address the core problem from Section 2
      - Every success indicator must have a corresponding failure signal
      - Technical metrics (uptime, latency, error rate) belong to Code Design, not here
    </rules>
  </section>
-->

---

## 9. Open Questions

<!--
  <section id="open-questions">
    <purpose>Capture what we still don't know. These are unresolved items that emerged during brainstorming but couldn't be answered. They feed into the Plan phase as inputs to resolve before or during implementation.</purpose>
    <format>
      - Numbered list
      - Each question includes:
        1. The question itself
        2. Why it matters (what is blocked or at risk without an answer)
        3. Suggested owner (who should answer this — PM, user, dev team, etc.)
    </format>
    <rules>
      - Only include questions that genuinely couldn't be answered during brainstorming
      - Do NOT include implementation questions ("which database?", "REST or GraphQL?") — those are Code Design scope
      - Product questions that a dev couldn't answer (Light level) should be flagged for PM escalation
    </rules>
  </section>
-->
