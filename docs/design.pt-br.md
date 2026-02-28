# Oraculo — Design do Sistema

## 1. Modelo de Operação

O Oráculo opera em dois modos, sempre delegando a execução.

**Modo completo (para épicos):** Descobrir > Planejar > Executar > Validar

**Modo reduzido (para stories sem épico):** Planejar > Executar > Validar

No modo reduzido, a fase de Descobrir é pulada — o contexto vem direto do usuário. O caminho mínimo é Planejar → Executar → Validar.

### 1.1 Descobrir (Discover)

O Oráculo instiga o usuário com perguntas. Explora a ideia, identifica riscos, levanta edge cases. O resultado é um documento de requisitos validado pelo usuário.

A fase Discover termina em um **approval gate** (`epic-requirements`): o agente chama `oraculo tools approval request --type epic-requirements`, o dashboard exibe o documento de requisitos para revisão, e o agente entra em `awaiting_approval`. O Plan não começa até que um verdict seja recebido.

_Obrigatória para epics. Pulada quando uma story é submetida diretamente com contexto suficiente._

### 1.2 Planejar (Plan)

Os requisitos são decompostos em tarefas. O Oráculo modela as dependências como um DAG — identifica o que é paralelo, o que é sequencial, e onde está o constraint (TOC). O resultado é um plano de execução otimizado.

Um **approval gate** opcional (`execution-plan`) existe entre Plan e Execute: o agente pode chamar `oraculo tools approval request --type execution-plan` para apresentar o DAG para revisão humana antes de os agentes serem despachados. Esse gate é recomendado para epics grandes ou de alto risco.

### 1.3 Executar (Execute)

O Oráculo monta um time de agentes e delega. Cada agente recebe uma tarefa específica com contexto claro: padrões do projeto, arquitetura existente, testes esperados. Agentes trabalham em paralelo seguindo o DAG. Todo código é escrito com TDD.

### 1.4 Validar (Validate)

Um agente de QA dedicado revisa a implementação. Ele verifica: testes passam, padrões do projeto foram respeitados, edge cases estão cobertos, a implementação atende os requisitos documentados. O agente de QA é independente dos agentes que executaram — olhar fresco, sem viés. Se o QA reprova, volta à fase apropriada — nunca força a barra, nunca conclui com ressalvas.

Quando o QA agent identifica um defeito crítico que não consegue resolver de forma autônoma, ele submete uma approval request `qa-escalation` ao dashboard via `oraculo tools approval request --type qa-escalation`. Isso aciona um **approval gate** (`qa-escalation`) que expõe o problema a um revisor humano pelo dashboard. O agente entra em `awaiting_approval` até que o revisor entregue um verdict que direciona a próxima ação.

**Regra de ouro:** O Oráculo nunca pula as fases centrais (Plan, Execute, Validate). Para epics, Discover é obrigatório. Approval gates entre fases são obrigatórios — o fluxo nunca avança além de um gate sem um verdict explícito.

## 2. Documentação como Memória do Projeto

Tudo que o Oráculo produz é registrado. Nada se perde, mas sem poluir o projeto.

### SQLite — Estado Operacional + Conhecimento Acumulado

Um único banco SQLite (`.oraculo/oraculo.db`) dentro do projeto serve dois propósitos:

**Dados operacionais transientes** — Status de tarefas, dependências, vereditos de QA, approval requests e verdicts, e estado de execução. Esses dados são essenciais durante o desenvolvimento ativo, mas podem ser limpos após a conclusão de um epic.

**Conhecimento persistente** — A tabela `knowledge` acumula lições aprendidas, padrões de codebase e convenções descobertas em todos os epics. Esses dados sobrevivem ao ciclo de vida do epic e ficam mais ricos com o tempo.

Ao concluir um epic/story:
1. Gerar um markdown de overview (resumo do que foi implementado, decisões-chave, resultado do QA)
2. Extrair lições aprendidas para a tabela `knowledge`
3. Dados operacionais daquele epic podem ser limpos (opcional)

### Markdown como Saída de Fase

Cada fase do modelo de operação produz seu próprio artefato markdown:
- **Discover** gera um documento de requisitos (`requirements.md` no nível do epic)
- **Plan** consome os requisitos e produz o DAG (rastreado no SQLite)
- **Validate** aciona a geração de um markdown de overview resumindo a implementação

O SQLite mantém estado operacional e conhecimento acumulado. Arquivos markdown capturam definições de produto e resumos de implementação. O projeto permanece limpo — um arquivo `.db`, um `requirements.md` por epic/story, e um overview por implementação validada.

## 3. Ecossistema Claude Code

O Oráculo é construído inteiramente sobre o ecossistema do Claude Code. Cada capacidade nativa é uma peça do sistema.

### Skills (Comandos)

Os pontos de entrada do Oráculo. Cada fase do modelo de operação é uma skill invocável — `/oraculo:epic` (modo completo), `/oraculo:story` (modo reduzido), `/oraculo:plan`, `/oraculo:execute`, etc. O usuário interage com o Oráculo através desses comandos.

### Teams (Agentes)

O motor de execução. O Oráculo usa a funcionalidade de times do Claude Code para montar equipes de agentes especializados — agentes de código, agente de QA, agentes de pesquisa. Cada agente recebe contexto e escopo bem definidos. O Oráculo é o líder do time, nunca um membro executor.

### Hooks

Os guardiões automáticos. Hooks garantem que padrões sejam respeitados sem depender da boa vontade — validações pré-commit, verificações de qualidade, formatação. Agem como gates que o código precisa passar.

### CLAUDE.md / Memória

O contexto persistente. Padrões do projeto, convenções de código, arquitetura — tudo que os agentes precisam saber para produzir código que se encaixa no projeto. O Oráculo alimenta seus agentes com esse contexto antes de qualquer delegação.

### Dashboard

A superfície de observação e controle. Um dashboard baseado em browser que fornece visibilidade sobre agentes, tarefas, o DAG, approval gates e conhecimento acumulado. Consome dados pelo CLI Trust Layer (nunca o contorna) e funciona como Mission Control: consciência situacional abrangente com intervenção humana estratégica nos approval gates. Quando um agente entra em `awaiting_approval`, o dashboard exibe o artefato para revisão e coleta o verdict humano (`approved`, `rejected` ou `needs_revision`).

**Princípio:** O Oráculo não reinventa ferramentas. Ele orquestra o que o Claude Code já oferece, tirando o máximo de cada capacidade nativa.

## 4. Público-Alvo e Fluxo do Time

O Oráculo é uma ferramenta de time, não de desenvolvedor solo.

### Quem usa

- **Produto** — Traz ideias e funcionalidades. O Oráculo guia a descoberta, faz as perguntas certas, documenta decisões. Produto não precisa saber código para usar o Oráculo na fase de Descobrir.
- **Desenvolvimento** — Recebe requisitos já refinados e um plano de execução estruturado. O Oráculo orquestra os agentes para implementar com qualidade. O dev supervisiona, não executa manualmente cada linha.
- **Qualquer pessoa do time** — Pode consultar o SQLite ou ler os overviews em Markdown para entender o histórico de qualquer funcionalidade.

### Fluxo típico

**Fluxo de epic (Product Engineering):**

1. Alguém do time tem uma ideia ou identifica um problema
2. Inicia o Oráculo com `/oraculo:epic` — perguntas, refinamento, edge cases
3. **Approval gate** (`epic-requirements`) — dashboard apresenta os requisitos para revisão humana; fluxo pausa até o verdict
4. Requisitos validados viram um plano com tarefas em DAG
5. **Approval gate** opcional (`execution-plan`) — dashboard apresenta o DAG para revisão antes dos agentes serem despachados
6. Agentes executam em paralelo, seguindo TDD e padrões do projeto
7. Agente de QA valida independentemente — defeitos críticos acionam **approval gate** (`qa-escalation`) para direcionamento humano; se rejeitado, retorna à fase apropriada

**Fluxo de story (Software Engineering):**

1. Um item de trabalho já está definido ou o usuário fornece contexto direto
2. Inicia o Oráculo com `/oraculo:story` — pula Discover, vai direto para Plan
3. **Approval gate** (`story-definition`) — dashboard apresenta a definição da story para revisão humana; fluxo pausa até o verdict
4. Agentes executam em paralelo, seguindo TDD e padrões do projeto
5. Agente de QA valida independentemente — defeitos críticos acionam **approval gate** (`qa-escalation`) para direcionamento humano; se rejeitado, retorna à fase apropriada

**O Oráculo reduz a distância entre a ideia e o código de qualidade.** Ele não substitui o time — ele amplifica a capacidade do time de pensar bem e executar com rigor.
