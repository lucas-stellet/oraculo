# Comandos Oraculo — Design do Sistema

## 1. Anatomia de um Comando

Um comando é composto por 3 camadas de arquivos:

```
skills/oraculo/<command>/
  SKILL.md                    # Ponto de entrada — frontmatter + persona + visão geral do workflow
  phases/                     # Arquivos de fase carregados just-in-time
    00-setup.md
    01-phase-name.md
    ...
  references/                 # Material de suporte carregado sob demanda
    question-bank.md
    frameworks.md
    artifact-templates.md
```

**SKILL.md** — O ponto de entrada. Contém:
- Frontmatter enriquecido (name, description, argument-hint, allowed-tools, disable-model-invocation)
- A única persona funcional do Oraculo
- Visão geral do workflow (lista de fases com uma descrição de uma linha cada)
- Instruções de bootstrap (verificar sessão ativa via CLI, carregar a fase correta)

**phases/*.md** — Um arquivo por fase. Contém:
- Tags XML guardrail relevantes para aquela fase
- Perguntas e lógica específicas da fase
- Phase gate de saída com critérios de saída
- Anti-patterns e condições de HALT específicos daquela fase

**references/*.md** — Material de suporte. Nunca carregado automaticamente. Lido explicitamente quando uma fase precisa (ex.: question-bank durante a Divergência, artifact-template durante a geração).

**Regra fundamental:** SKILL.md nunca ultrapassa ~100 linhas. A complexidade vive nos arquivos de fase e nas referências. O ponto de entrada é enxuto.

## 2. Frontmatter

O frontmatter define os metadados que o Claude Code usa para registro e despacho. O campo `description` descreve QUANDO usar o comando — nunca resume o workflow.

```yaml
---
name: oraculo:epic
description: >
  Use when exploring a new feature idea that spans multiple areas,
  has significant uncertainty, or needs structured discovery before planning
argument-hint: <idea or problem description>
allowed-tools:
  - Read
  - Bash
  - Task
  - AskUserQuestion
disable-model-invocation: true
---
```

**Campos obrigatórios:**

| Campo | Propósito |
|-------|-----------|
| `name` | Identificador do slash command. Formato: `oraculo:<phase>` |
| `description` | Condições de disparo — quando invocar. Nunca resume o workflow. |

**Campos opcionais:**

| Campo | Propósito | Quando usar |
|-------|-----------|-------------|
| `argument-hint` | Placeholder exibido ao usuário | Quando o comando aceita input |
| `allowed-tools` | Lista de ferramentas permitidas | Quando o comando deve ser restrito |
| `disable-model-invocation` | Impede auto-invocação pelo modelo | Sempre `true` para comandos Oraculo — só o usuário invoca |

**O que NÃO vai no frontmatter:**
- Lógica de workflow (vai no corpo do SKILL.md)
- Referências de arquivos (vão nas instruções de cada fase)
- Estado ou variáveis de runtime (gerenciados pela CLI)

## 3. Vocabulário XML

10 tags divididas em duas categorias. Cada tag tem uma origem documentada em pesquisa, justificativa na filosofia e regras de uso.

### Tags Semânticas (organizam seções do prompt)

**`<persona>`** — Declara a identidade do Oraculo para a fase atual.

```xml
<persona>
  You are Oraculo, a Socratic guide for quality product development.
  You ask before doing. You orchestrate, never execute.
  You prioritize understanding over speed.

  In this phase (Divergence), your focus is expanding the problem space —
  surfacing dimensions the user hasn't considered.
</persona>
```

Regra: Aparece uma vez no SKILL.md (identidade base) e uma vez em cada arquivo de fase (foco da fase). O segundo não repete o primeiro — adiciona contexto de fase.

**`<workflow>`** — Declara a sequência de fases no SKILL.md. Não contém lógica — apenas a visão geral.

```xml
<workflow>
  00-setup:       Session triage and reasoning level detection
  01-reframing:   Separate problem from solution
  02-divergence:  Expand problem space (JTBD, Four Forces, edge cases)
  03-codebase:    Dispatch research agents (parallel)
  04-convergence: Narrow to validated problem definition
  05-assumptions: Map and score hidden assumptions
  06-exit-gate:   Validate four risks before proceeding
  07-artifact:    Generate requirements document via CLI
</workflow>
```

Regra: Aparece apenas no SKILL.md. Nunca nos arquivos de fase. Serve como mapa, não como instrução.

**`<anti-patterns>`** — Lista nomeada de modos de falha específicos da fase.

```xml
<anti-patterns>
  "Premature solutioning" — proposing HOW before validating WHAT/WHY
  "Assumption blindness" — accepting user hypotheses without questioning evidence
  "Shallow reframing" — accepting the first problem statement without pushing deeper
</anti-patterns>
```

Regra: Cada anti-pattern tem um nome entre aspas seguido de uma descrição. Os nomes devem ser memoráveis. Cada arquivo de fase define os seus — não há lista global.

**`<context-boundaries>`** — Declara explicitamente o que o LLM pode e não pode acessar nesta fase.

```xml
<context-boundaries>
  Available: Raw idea from user, reasoning level from setup
  Not available: Codebase analysis (runs in parallel, results arrive later)
  Focus: Problem exploration, not solution design
  References: Load question-bank.md section "JTBD" when starting Four Forces
</context-boundaries>
```

Regra: Aparece em cada arquivo de fase. Impede que o LLM invente contexto que não possui ou carregue referências desnecessárias.

### Tags de Fluxo (controlam a execução em pontos críticos)

**`<phase-gate>`** — Define critérios de saída e persistência via CLI. O LLM não avança sem satisfazê-los.

```xml
<phase-gate phase="divergence">
  Exit conditions:
    - At least 3 JTBD dimensions explored (functional, emotional, social)
    - All four forces addressed (push, pull, anxiety, habit)
    - At least 1 edge case or risk the user hadn't previously considered

  Persist via CLI:
    - oraculo phase complete divergence --session=$SESSION_ID

  If CLI rejects: surface rejection to user. Do not retry with looser criteria.
  On success: Read phases/04-convergence.md
</phase-gate>
```

Regra: Exatamente um por arquivo de fase, sempre no final. Contém condições de saída, comando CLI e instrução de navegação.

**`<halt>`** — Condição de parada automática quando algo está errado.

```xml
<halt>
  - Zero assumptions identified after Divergence — something was ignored, re-analyze
  - User cannot articulate problem in one sentence after Convergence — return to Reframing
  - Codebase Analysis found zero related components — verify analysis was sufficient
</halt>
```

Regra: Cada condição é uma frase declarativa. A ação é sempre parar e reagir, nunca continuar silenciosamente.

**`<self-check>`** — Verificação de afirmações antes de entregar um artefato ou completar uma fase.

```xml
<self-check phase="exit-gate">
  Before declaring all four risks addressed:
  1. Reread problem statement — still coherent with what was explored?
  2. Verify each critical assumption has evidence or explicit flag
  3. Confirm Codebase Analysis results were incorporated (not ignored)
  4. Check that scope boundaries defined in Reframing are still respected

  If any check fails — do not generate artifact. Return to appropriate phase.
</self-check>
```

Regra: Aparece antes de transições irreversíveis (geração de artefato, handoff para a próxima fase do modelo operacional). As verificações são objetivas, não subjetivas.

**`<rationalization-table>`** — Mapeia desculpas do LLM para a realidade. Inoculação cognitiva.

```xml
<rationalization-table>
  "The user already knows what they want"    → If they did, they wouldn't need Oraculo. Explore.
  "This idea is too simple for 8 phases"     → Simple ideas hide the most assumptions. Follow the process.
  "I already understand the problem"         → Understanding is not validating. The user must confirm.
  "We can skip Assumption Mapping, it's obvious" → Obvious assumptions are the most dangerous ones.
</rationalization-table>
```

Regra: Formato fixo: desculpa entre aspas seguida da realidade. Cada entrada tem uma frase. Aparece nas fases onde o risco de atalho é maior.

**`<iteration-limit>`** — Define o número máximo de tentativas antes de mudar a abordagem.

```xml
<iteration-limit phase="reframing" max="3">
  If after 3 attempts the user cannot separate problem from solution:
  1. Record the proposed solution as-is
  2. Ask: "What outcome do you expect this solution to produce for the user?"
  3. Use the answer as a proxy problem statement
  Do not insist indefinitely — adapt the approach.
</iteration-limit>
```

Regra: Sempre inclui `max` e a ação de escalação. Escalação nunca é "desistir" — é "adaptar".

**`<critical>`** — Regra inviolável repetida no ponto de maior risco de desvio.

```xml
<critical>
  The command does NOT generate requirements without user validation at each phase.
  Verbal confirmation is process, not approval.
  Approval is a human verdict issued through the dashboard approval gate after the artifact is submitted via `oraculo tools approval request`.
</critical>
```

Regra: Reservado para regras que, se violadas, comprometem a integridade do sistema. Máximo de 2-3 por arquivo de fase. Se tudo é `<critical>`, nada é.

### Resumo do Vocabulário

| Tag | Categoria | Onde aparece | Frequência |
|-----|-----------|--------------|------------|
| `<persona>` | Semântica | SKILL.md + cada arquivo de fase | 1 por arquivo |
| `<workflow>` | Semântica | Apenas SKILL.md | 1 total |
| `<anti-patterns>` | Semântica | Cada arquivo de fase | 1 por fase |
| `<context-boundaries>` | Semântica | Cada arquivo de fase | 1 por fase |
| `<phase-gate>` | Fluxo | Final de cada arquivo de fase | 1 por fase |
| `<halt>` | Fluxo | Arquivos de fase onde relevante | 0-1 por fase |
| `<self-check>` | Fluxo | Antes de transições irreversíveis | 0-1 por fase |
| `<rationalization-table>` | Fluxo | Fases com maior risco de atalho | 0-1 por fase |
| `<iteration-limit>` | Fluxo | Fases com interação iterativa | 0-1 por fase |
| `<critical>` | Fluxo | Ponto de maior risco de desvio | 0-3 por fase |

## 4. Estrutura do Arquivo de Fase

Cada arquivo de fase segue uma estrutura fixa. A ordem das seções não é arbitrária — reflete a prioridade de atenção do LLM (instruções no topo têm maior aderência).

Exemplo: `phases/01-reframing.md`

```markdown
# Phase 01: Reframing

<persona>
  You are Oraculo in the Reframing phase.
  Your focus: separate the problem from any proposed solution.
</persona>

<context-boundaries>
  Available: Raw idea from user, reasoning level from 00-setup
  Not available: Codebase analysis (not yet started)
  Focus: Problem clarity, not solution exploration
  References: Load references/question-bank.md section "Reframing" if needed
</context-boundaries>

<critical>
  Do NOT accept a solution description as a problem statement.
  "I want to add a cache" is a solution. "Pages load too slowly" is a problem.
</critical>

<rationalization-table>
  "The user clearly described the problem" → Did they describe a problem or a solution they already chose?
  "Reframing will frustrate the user"     → Skipping reframing produces requirements that solve the wrong problem.
</rationalization-table>

<anti-patterns>
  "Premature solutioning" — accepting HOW as the starting point instead of WHY
  "Shallow reframing" — accepting the first restatement without probing deeper
  "Scope inflation" — allowing the problem statement to absorb adjacent problems
</anti-patterns>

## Questions

### Deep Reasoning
- "You described what you want to build. What problem does this solve for the user?"
- "If this feature didn't exist, what would the user do? What's the cost of that alternative?"
- "What is explicitly out of scope for this exploration?"

### Light Reasoning
- "You received this task. What's the problem it solves?"
- "What is explicitly out of scope?"

### Epic-Derived
- "The epic describes [REC-N context]. What's the specific problem this story addresses?"

## Execution

Ask one question at a time. Wait for the user's answer before proceeding.

After each answer, evaluate:
- Did the user articulate a PROBLEM (not a solution)?
- Is the problem statement independent from implementation?
- Are scope boundaries emerging?

When all three are satisfied, proceed to the phase gate.

<iteration-limit phase="reframing" max="3">
  If after 3 attempts the user cannot separate problem from solution:
  1. Record the proposed solution as-is
  2. Ask: "What outcome do you expect this solution to produce for the user?"
  3. Use the answer as a proxy problem statement
  Do not insist indefinitely — adapt the approach.
</iteration-limit>

<halt>
  - User explicitly refuses to engage with reframing — note the gap, do not force
  - Problem statement is identical to the raw idea after 3 iterations — flag as risk
</halt>

<phase-gate phase="reframing">
  Exit conditions:
    - Problem statement articulated, independent from any specific solution
    - Raw idea preserved as originally stated (separate from reframed problem)
    - Scope boundaries defined (at least one explicit exclusion)

  Persist via CLI:
    - oraculo phase complete reframing --session=$SESSION_ID

  If CLI rejects: surface rejection to user. Do not retry with looser criteria.
  On success: Read phases/02-divergence.md
</phase-gate>
```

**Ordem fixa das seções:**

| Ordem | Seção | Obrigatória | Propósito |
|-------|-------|:-----------:|-----------|
| 1 | `<persona>` | Sim | Foco da fase — lida primeiro |
| 2 | `<context-boundaries>` | Sim | O que o LLM tem/não tem disponível |
| 3 | `<critical>` | Não | Regras invioláveis — reforço imediato |
| 4 | `<rationalization-table>` | Não | Inoculação antes da execução |
| 5 | `<anti-patterns>` | Sim | Modos de falha nomeados |
| 6 | Perguntas / Execução | Sim | O trabalho real da fase (Markdown livre) |
| 7 | `<iteration-limit>` | Não | Limites de tentativas |
| 8 | `<halt>` | Não | Condições de parada |
| 9 | `<phase-gate>` | Sim | Sempre por último — critérios de saída |

As tags semânticas (persona, boundaries, critical) vêm primeiro porque o LLM precisa delas antes de agir. As tags de fluxo (iteration-limit, halt, phase-gate) vêm depois porque controlam quando parar, não como começar. O conteúdo executivo (perguntas, lógica de execução) fica no meio — o trabalho real, emoldurado por guardrails dos dois lados.

## 5. Gerenciamento de Contexto

Três regras governam como o contexto é usado ao longo de uma sessão.

### Regra 1: Apenas um arquivo de fase em foco por vez

O SKILL.md carrega o arquivo da fase atual via `Read`. Quando a fase é concluída e o `<phase-gate>` avança, o arquivo da próxima fase é carregado. O anterior naturalmente sai do foco do LLM conforme a conversa cresce.

```
Session starts → Read SKILL.md → Read phases/00-setup.md
Setup completes → Read phases/01-reframing.md
Reframing completes → Read phases/02-divergence.md
...
```

O LLM nunca carrega múltiplos arquivos de fase simultaneamente. Nunca antecipa fases futuras. Nunca cria "listas mentais de tarefas" do workflow completo — o `<workflow>` no SKILL.md é o mapa, não a instrução.

### Regra 2: Referências são carregadas por instrução explícita

Cada `<context-boundaries>` declara quais referências podem ser carregadas e quando. O arquivo de fase nunca diz "leia todo o question-bank.md" — diz qual seção.

```xml
<context-boundaries>
  References: Load references/question-bank.md section "JTBD" when starting Four Forces
  References: Load references/frameworks.md section "Assumption Mapping" when scoring begins
</context-boundaries>
```

Se uma referência não está declarada no `<context-boundaries>` da fase atual, o LLM não a carrega. Isso impede carregamento especulativo que desperdiça contexto.

### Regra 3: O estado vive na CLI, não na conversa

Resultados validados são persistidos imediatamente via CLI no `<phase-gate>`. Se a sessão cair, a CLI tem tudo que foi validado. A conversa é efêmera — o artefato é o que sobrevive.

Consequências práticas:
- O comando nunca acumula estado internamente ("até agora identificamos X premissas...")
- Se o LLM precisa de dados de fases anteriores, consulta a CLI (`oraculo session state`)
- Se o contexto comprime (estouro da janela de contexto), o arquivo da fase atual + estado da CLI são suficientes para continuar

### Retomada de Sessão

O SKILL.md inclui uma instrução de bootstrap que verifica uma sessão ativa antes de qualquer outra coisa:

```xml
<phase-gate phase="bootstrap">
  1. Call: oraculo session status --type=epic
  1.5 Call: oraculo tools approval list --pending
      If any approvals pending: display warning, do NOT proceed
  2. If active session found:
     - Read phase file for current phase
     - Display to user: "Active session found. Resuming from [phase name]."
  3. If no active session:
     - Read phases/00-setup.md
     - CLI creates session on first phase completion
</phase-gate>
```

A CLI é a fonte de verdade. O comando é stateless.

### Imposição da Ordem de Carregamento

A ordem de carregamento é imposta por três camadas:

**Camada 1 — Navegação explícita (prompt):** Cada `<phase-gate>` termina com a instrução exata do próximo arquivo (`On success: Read phases/02-divergence.md`). O LLM não precisa decidir o que vem a seguir — está escrito.

**Camada 2 — Sinal contextual (estrutura):** Se o LLM pular à frente, o `<context-boundaries>` da fase-alvo descreve inputs que ainda não existem, criando atrito natural.

**Camada 3 — CLI rejeita (determinístico):** A CLI valida a sequência. `oraculo phase complete divergence` falha se `reframing` não foi concluído primeiro. O comando orienta o comportamento; a CLI garante a correção.

## 6. Comandos Oraculo

Cada comando mapeia para uma fase do modelo operacional definido em `docs/design.md`.

### Catálogo

| Comando | Fase do Modelo Operacional | Input | Output |
|---------|---------------------------|-------|--------|
| `/oraculo:epic` | Discover | Ideia crua ou problema | `requirements.md` aprovado (veredicto humano via dashboard) |
| `/oraculo:story` | Discover (light) | Item de trabalho ou REC-N do epic | `requirements.md` de story aprovado (veredicto humano via dashboard) |
| `/oraculo:plan` | Plan | Requisitos aprovados | DAG de tarefas em SQLite via CLI |
| `/oraculo:execute` | Execute | DAG pronto | Agentes executam, código commitado |
| `/oraculo:validate` | Validate | Implementação completa | Veredicto de QA via CLI |

### `/oraculo:epic`

O comando mais denso. Guia a exploração socrática completa do Double Diamond.

```
skills/oraculo/epic/
  SKILL.md
  phases/
    00-setup.md              # Triage, reasoning level, session check
    01-reframing.md          # Separate problem from solution
    02-divergence.md         # JTBD, Four Forces, edge cases
    03-codebase-analysis.md  # Dispatch research agents
    04-convergence.md        # Narrow to validated problem
    05-assumptions.md        # Map and score assumptions
    06-exit-gate.md          # Four risks validation + self-check
    07-artifact.md           # Generate requirements via CLI
    08-approval.md           # Submit requirements for human approval via dashboard
  references/
    question-bank.md         # Questions by phase and reasoning level
    frameworks.md            # JTBD, Four Forces, TOC, Assumption Mapping
    artifact-templates.md    # Requirements document template
```

9 fases. Sessão longa. Todas as tags XML são utilizadas.

### `/oraculo:story`

Versão compacta do epic. Mesmo padrão estrutural, menos fases.

```
skills/oraculo/story/
  SKILL.md
  phases/
    00-setup.md              # Triage, parent epic check, reasoning level
    01-reframing.md          # Focused problem/scope definition
    02-assumptions.md        # Quick assumption check (2-3 critical)
    03-exit-gate.md          # Four risks (light version)
    04-artifact.md           # Generate story document via CLI
    05-approval.md           # Submit story definition for human approval via dashboard
  references/
    question-bank.md         # Story-specific questions
    artifact-templates.md    # Story document template
```

6 fases. Sessão mais curta. Herda contexto do epic quando derivado. O `00-setup.md` lê o epic da CLI quando passado como argumento.

### `/oraculo:plan`

Consome os requisitos e produz o DAG de execução. Aqui o Oraculo muda de modo — de guia socrático para planejador técnico.

```
skills/oraculo/plan/
  SKILL.md
  phases/
    00-setup.md              # Load requirements, validate completeness
    01-decomposition.md      # Break requirements into tasks
    02-dependencies.md       # Model DAG, identify constraint (TOC)
    03-optimization.md       # Maximize parallelism, minimize critical path
    04-artifact.md           # Persist DAG via CLI
  references/
    decomposition-patterns.md  # How to break requirements into tasks
```

5 fases. Input: requisitos aprovados (veredicto humano via dashboard). Output: DAG em SQLite. O Oraculo pode despachar agentes de pesquisa para analisar a codebase durante a decomposição. Um approval gate opcional para o plano de execução pode ser inserido após `04-artifact.md` quando o time exige aprovação humana antes que os agentes comecem o trabalho.

### `/oraculo:execute`

O comando mais diferente. O Oraculo não executa — ele orquestra. Este comando monta o time de agentes e delega.

```
skills/oraculo/execute/
  SKILL.md
  phases/
    00-setup.md              # Load DAG from CLI, verify readiness
    01-team-assembly.md      # Select agents, assign tasks per DAG
    02-monitoring.md         # Track progress, handle failures
    03-completion.md         # Verify all tasks done, prepare for QA
  references/
    agent-dispatch.md        # How to spawn and configure agents
```

4 fases. Este comando interage mais com o sistema de agentes (`docs/agents/design.md`) do que com o usuário. O `<persona>` muda: o Oraculo aqui é orquestrador, não guia socrático.

### `/oraculo:validate`

Despacha o agente de QA com contexto limpo (olhos frescos). Recebe o veredicto e decide o próximo passo.

```
skills/oraculo/validate/
  SKILL.md
  phases/
    00-setup.md              # Load implementation state, prepare QA context
    01-qa-dispatch.md        # Spawn QA agent with clean context
    02-verdict.md            # Process QA result, decide next step
  references/
    qa-criteria.md           # What the QA agent checks
```

3 fases. Se o QA rejeitar, o comando identifica a fase correta para retorno e instrui o usuário.

### Padrão Comum

Todos os comandos compartilham:
- Mesma estrutura de diretório (SKILL.md + phases/ + references/)
- Mesma ordem de seções nos arquivos de fase
- Mesmo vocabulário XML (10 tags)
- Mesma mecânica de phase-gate (CLI persiste, CLI valida)
- Mesma persona base (adaptada por fase)

O que varia:
- Número de fases (3 a 8)
- Densidade de guardrails (epic > story > plan > execute > validate)
- Tom da persona (socrático em epic/story, orquestrador em execute/validate)
- Referências disponíveis (question banks em epic/story, padrões em plan, regras de despacho em execute)

## 7. Interação entre Camadas

A filosofia define três camadas com responsabilidades absolutas: Command, CLI e Agents.

### O Contrato

```
Command (interpretativo)   CLI (determinístico)      Agents (execução)
────────────────────────   ────────────────────────  ──────────────────
Guia a conversa            Valida e persiste         Executam código
Decide QUANDO perguntar    Decide SE aceita          Seguem TDD
Produz artefatos           Garante schema/seq        Trabalham no DAG
Chama a CLI                Rejeita ou confirma       Reportam resultados
```

**Regra inviolável:** Nenhuma camada faz o trabalho de outra camada.

- O comando nunca escreve diretamente no SQLite, constrói file paths ou valida schemas
- A CLI nunca faz perguntas ao usuário, interpreta intenção ou adapta tom
- Os agentes nunca fazem discovery, modificam requisitos ou decidem escopo

### Handoffs Entre Comandos

Cada comando termina sugerindo o próximo, mas nunca o invoca automaticamente. O usuário decide quando avançar.

```xml
<!-- In epic's 07-artifact.md -->
<phase-gate phase="artifact">
  ...
  On success:
    Call: oraculo tools approval request --type=epic-requirements --session=$SESSION_ID
    Display: "Requirements document submitted for approval. Awaiting human verdict via dashboard."
    Enter awaiting_approval — do NOT advance until verdict is received.

  On verdict=approved:
    Display: "Requirements approved. To decompose into stories, run /oraculo:story <epic-name>"
    Do NOT invoke /oraculo:story automatically.
    The user decides when to proceed.

  On verdict=rejected:
    Display the rejection reason.
    Return to phases/06-exit-gate.md for re-validation.

  On verdict=needs_revision:
    Display the reviewer's comments.
    Return to the appropriate phase to address comments, then re-generate the artifact.
</phase-gate>
```

Cada transição é um ponto de decisão do usuário controlado por um veredicto humano. O usuário pode querer revisar os requisitos com o time, aguardar ou ajustar — mas o fluxo não avança sem um veredicto `approved` explícito do dashboard. O Oraculo sugere os próximos passos; o humano e o approval gate decidem.

### Interação com Agentes

Os comandos `plan`, `execute` e `validate` interagem com o sistema de agentes definido em `docs/agents/design.md`.

| Comando | Tipo de Agente | Quando | Como |
|---------|---------------|--------|------|
| `epic` | Agente de pesquisa | Fase 03 (Codebase Analysis) | Task tool, paralelo à conversa |
| `story` | Nenhum | — | Sessão puramente socrática |
| `plan` | Agente de pesquisa (opcional) | Fase 01 (Decomposition) | Análise de codebase para informar tarefas |
| `execute` | Agentes de código + Agentes de pesquisa | Fases 01-02 | Montagem do time via TeamCreate |
| `validate` | Agente de QA | Fase 01 | Task tool, contexto limpo |

O comando nunca configura o agente diretamente. Declara a intenção e o despacho segue as regras em `docs/agents/design.md`.

Entre comandos nos approval gates, o agente orquestrador entra em `awaiting_approval` — um estado distinto ao lado de `active`, `idle`, `completed` e `failed`. Nenhum comando downstream é invocado e nenhum agente é despachado enquanto em `awaiting_approval`. O estado se resolve apenas quando o dashboard entrega um veredicto.

### Rejeição e Retorno

Quando o QA rejeita em `validate`, o comando identifica a fase de retorno:

```xml
<phase-gate phase="verdict">
  If QA verdict is REJECTED:
    Analyze rejection reasons:
    - Implementation bug → return to /oraculo:execute (re-execute failed tasks)
    - Missing requirement → return to /oraculo:plan (re-decompose)
    - Wrong requirement → return to /oraculo:story or /oraculo:epic (re-discover)

    Display the verdict, the analysis, and the suggested return point.
    The user decides which command to re-invoke.
    Do NOT re-invoke automatically.
</phase-gate>
```

O Oraculo nunca força um caminho. Analisa, recomenda, e o usuário decide.

## 8. Critérios de Evolução

O vocabulário XML tem 10 tags hoje. A estrutura tem 5 comandos. Esta seção define como o sistema evolui sem degradar.

### Adicionando uma Nova Tag XML

Uma nova tag deve passar por três critérios. Se falhar em qualquer um, não entra no vocabulário.

| Critério | Teste |
|----------|-------|
| **Mapeia para a filosofia** | A tag resolve um problema descrito em `docs/commands/philosophy.md`? Se nenhum princípio a sustenta, é invenção sem fundamento. |
| **Markdown não resolve** | O mesmo efeito pode ser alcançado com cabeçalhos, negrito, tabelas ou blocos de código? Se sim, use Markdown — uma tag extra é custo sem benefício. |
| **Evidência de necessidade** | Houve desvio ou falha documentada que essa tag teria prevenido? Tags criadas preventivamente, sem evidência de um problema real, diluem o vocabulário. |

**Processo:** Documente a justificativa no design doc antes de implementar. A tag entra no vocabulário com sua origem (qual projeto a inspirou ou qual falha a motivou) e um exemplo de uso.

**Limite suave:** Se o vocabulário ultrapassar 15 tags, revise as existentes. Alguma provavelmente não está justificando seu custo.

### Adicionando uma Nova Fase a um Comando

| Critério | Teste |
|----------|-------|
| **Condição de saída distinta** | A fase tem uma condição de saída que não se sobrepõe às fases adjacentes? Se a condição de saída é igual à da fase anterior ou seguinte, não é uma fase — é parte de outra. |
| **Carregamento justifica a separação** | O conteúdo da fase é grande o suficiente para justificar um arquivo separado? Se cabe em 20 linhas, é uma seção dentro de outra fase, não uma nova fase. |
| **CLI valida a transição** | A CLI precisa validar algo entre esta fase e a próxima? Se não há nada a persistir ou validar, a separação é cosmética. |

### Adicionando um Novo Comando

Comandos mapeiam para fases do modelo operacional. Um novo comando implica uma nova fase no modelo — uma mudança arquitetural.

| Critério | Teste |
|----------|-------|
| **Fase do modelo operacional** | O comando mapeia para uma fase não coberta por nenhum comando existente? Se coberta, é uma extensão de um comando existente, não um novo. |
| **Artefato distinto** | O comando produz um artefato que nenhum outro comando produz? Se o output é do mesmo tipo que outro comando, provavelmente é uma variação, não um novo comando. |
| **Handoff claro** | Você consegue definir o que vem antes e depois deste comando no fluxo? Comandos sem conexão com o fluxo são features soltas. |

### Removendo Elementos

A remoção segue o critério inverso: se uma tag, fase ou comando não foi usado nos últimos N ciclos de desenvolvimento, ou se o problema que motivou sua criação não existe mais, é candidato à remoção.

### O que Este Documento NÃO Governa

- Conteúdo dos arquivos de fase (perguntas, frameworks, templates) — esse é o domínio dos designs do Epic e da Story
- Comportamento da CLI (validações, schemas, persistência) — esse é o domínio de `docs/cli/design.md`
- Comportamento dos agentes (TDD, despacho, QA) — esse é o domínio de `docs/agents/design.md`

Este documento governa como os comandos são construídos — a anatomia, o vocabulário, a estrutura e as regras de evolução. O conteúdo vive nos designs de cada fase.
