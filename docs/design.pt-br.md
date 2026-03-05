# Oraculo — Design do Sistema

## 1. Modelo de Operação

O Oráculo opera em dois modos dependendo do escopo do trabalho.

**Product Engineering (epics):** Descobrir > Planejar > Executar > Validar

**Software Engineering (stories sem épico):** Planejar > Executar > Validar

Stories sem um epic associado pulam Descobrir — o contexto vem direto do usuário. O caminho mínimo é sempre Planejar > Executar > Validar. Stories standalone usam um epic leve criado automaticamente pela CLI, preservando o modelo de dados hierárquico de forma transparente.

### 1.1 Descobrir (Discover)

O Oráculo instiga o usuário com perguntas. Explora a ideia, identifica riscos, levanta edge cases. O resultado é um documento de requisitos validado pelo usuário.

A fase Discover termina em uma **version review**: o agente chama `oraculo tools epic version <epic-name>` para criar um snapshot versionado, depois monitora reviews via `oraculo tools review list <version-id> --type epic`. O dashboard exibe o documento de requisitos para revisão. O Plan não começa até que um verdict seja recebido.

_Obrigatória para epics. Pulada quando uma story é submetida diretamente com contexto suficiente._

### 1.2 Planejar (Plan)

Os requisitos são decompostos em tarefas. O Oráculo modela as dependências como um DAG — identifica o que é paralelo, o que é sequencial, e onde está o constraint (TOC). O resultado é um plano de execução otimizado.

A fase Plan inclui uma **sub-fase de Design** entre a análise de dependências e a otimização. O orquestrador despacha dois agentes de pesquisa paralelos — um agente de pesquisa de codebase que analisa arquitetura existente, padrões e convenções, e um agente de pesquisa web que investiga tecnologias externas, melhores práticas e bibliotecas relevantes para a feature. Suas descobertas são sintetizadas em um documento de design arquitetural que especifica a abordagem técnica, fronteiras de componentes e pontos de integração.

Um **approval gate** obrigatório (`design`) segue a sub-fase de Design: o agente chama `oraculo tools approval request --type design`, o dashboard apresenta o documento de design arquitetural para revisão humana, e o agente entra em `awaiting_approval`. A otimização e execução não prosseguem até que um verdict seja recebido. Esse gate garante que a abordagem técnica seja validada antes de qualquer código ser escrito.

Um **approval gate** opcional (`execution-plan`) existe entre Plan e Execute: o agente pode chamar `oraculo tools approval request --type execution-plan` para apresentar o DAG para revisão humana antes de os agentes serem despachados. Esse gate é recomendado para epics grandes ou de alto risco.

### 1.3 Executar (Execute)

O Oráculo monta um time de agentes e delega. Cada agente recebe uma tarefa específica com contexto claro: padrões do projeto, arquitetura existente, testes esperados. Agentes trabalham em paralelo seguindo o DAG. Todo código é escrito com TDD.

### 1.4 Validar (Validate)

Um agente de QA dedicado revisa a implementação. Ele verifica: testes passam, padrões do projeto foram respeitados, edge cases estão cobertos, a implementação atende os requisitos documentados. O agente de QA é independente dos agentes que executaram — olhar fresco, sem viés. Se o QA reprova, o Oráculo retorna à fase apropriada — nunca força a barra, nunca aceita com ressalvas.

Quando o QA agent identifica um defeito crítico que não consegue resolver de forma autônoma, ele escala via `oraculo tools approval request --type qa-escalation`. Isso aciona um **approval gate** (`qa-escalation`) que expõe o problema a um revisor humano pelo dashboard. O agente entra em `awaiting_approval` até que o revisor entregue um verdict que direciona a próxima ação.

**Regra de ouro:** O Oráculo nunca pula as fases centrais (Plan, Execute, Validate). Para epics, Discover é obrigatório. Revisão humana entre fases é obrigatória — o fluxo nunca avança além de uma revisão ou gate sem um verdict explícito. Reviews de documentos (requisitos de epic, definições de story) usam o sistema de versionamento com verdicts `approved`/`rejected`. Gates operacionais (`design`, `execution-plan`, `qa-escalation`) usam o sistema de approvals com verdicts `approved`/`rejected`/`needs_revision`.

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

### Ferramentas Internas, Output de Domínio

O Oráculo usa frameworks analíticos (Double Diamond, JTBD, TOC, Assumption Mapping) como ferramentas internas de conversação durante a discovery e refinamento. Sua terminologia guia o pensamento do agente, mas não aparece nos artefatos de saída. Documentos de requisitos e definições de story são escritos na linguagem de domínio do usuário — contexto, regras de negócio, comportamento esperado e critérios de aceite — não em jargão de framework.

## 3. Ecossistema Claude Code

O Oráculo é construído inteiramente sobre o ecossistema do Claude Code. Cada capacidade nativa é uma peça do sistema.

### Skills (Comandos)

Os pontos de entrada do Oráculo. Cada fase do modelo de operação é uma skill invocável — `/oraculo:epic` (modo completo), `/oraculo:story` (modo reduzido), `/oraculo:plan`, `/oraculo:execute`, etc. O usuário interage com o Oráculo através desses comandos.

### Teams (Agentes)

O motor de execução. O Oráculo usa a funcionalidade de times do Claude Code para montar equipes de agentes especializados — agentes de código, agente de QA, agentes de pesquisa. Cada agente recebe contexto e escopo bem definidos. O Oráculo é o líder do time, nunca um membro executor.

### Hooks

Os guardiões automáticos. Hooks garantem que padrões sejam respeitados sem depender da boa vontade — validações pré-commit, verificações de qualidade, formatação. Agem como gates que o código precisa passar.

#### Arquitetura de Dois Canais

O Oráculo usa o sistema de hooks do Claude Code como um de dois canais complementares de comunicação:

**Canal 1 — HTTP Hooks (telemetria automática):** O Claude Code dispara HTTP hooks em eventos do sistema (início/fim de sessão, início/parada de agente, uso de ferramenta). O servidor HTTP do Oráculo recebe cada evento, persiste metadados no SQLite e transmite para clientes WebSocket conectados. São fire-and-forget: o hook retorna `200` com corpo vazio. Se o servidor estiver inacessível, o hook falha silenciosamente — agentes nunca são bloqueados por telemetria.

**Canal 2 — MCP (approval gates interativos):** Quando um agente precisa de aprovação humana, ele chama a ferramenta MCP `request_approval`. Isso bloqueia o agente até que um verdict seja recebido do dashboard. O canal MCP é explícito e bloqueante por design — reservado exclusivamente para momentos que requerem julgamento humano.

Os dois canais são complementares: HTTP hooks são automáticos e não-bloqueantes (ideal para telemetria); ferramentas MCP são explícitas e bloqueantes (ideal para aprovações). Cada canal faz o que faz melhor.

#### Endpoints de HTTP Hook

Os seguintes hooks são registrados por `oraculo install` e configurados automaticamente em `.claude/settings.json`:

| Endpoint | Claude Code Hook | Propósito |
|---|---|---|
| `POST /hooks/session-start` | `SessionStart` (command hook) | Registrar nova sessão com health check |
| `POST /hooks/agent-start` | `SubagentStart` | Agente instanciado |
| `POST /hooks/agent-stop` | `SubagentStop` | Agente concluído ou falhou |
| `POST /hooks/tool-used` | `PostToolUse` (`Bash\|Edit\|Write\|NotebookEdit`) | Evento de mutação (apenas metadados, sem conteúdo) |
| `POST /hooks/task-completed` | `TaskCompleted` | Task finalizada |
| `POST /hooks/stop` | `Stop` | Agente parando |
| `POST /hooks/teammate-idle` | `TeammateIdle` | Teammate ocioso |
| `POST /hooks/session-end` | `SessionEnd` | Sessão encerrada |

Todos os HTTP hooks usam `timeout: 5` (5 segundos) e seguem uma política de degradação graciosa — se o servidor estiver offline, o hook falha silenciosamente e o agente continua sem ser afetado.

O hook `SessionStart` é o único command hook (não HTTP). Ele executa `oraculo hook session-start`, que realiza um health check e imprime um aviso para o usuário se o dashboard estiver offline — HTTP hooks não podem enviar saída para o usuário. Sempre sai com código `0` e nunca bloqueia a sessão.

### CLAUDE.md / Memória

O contexto persistente. Padrões do projeto, convenções de código, arquitetura — tudo que os agentes precisam saber para produzir código que se encaixa no projeto. O Oráculo alimenta seus agentes com esse contexto antes de qualquer delegação.

### Dashboard

A superfície de observação e controle. Um dashboard baseado em browser que fornece visibilidade sobre agentes, tarefas, o DAG, approval gates e conhecimento acumulado. Consome dados pelo CLI Trust Layer (nunca o contorna) e funciona como Mission Control: consciência situacional abrangente com intervenção humana estratégica nos approval gates. Quando um agente entra em `awaiting_approval`, o dashboard exibe o artefato para revisão e coleta o verdict humano. Para reviews de documentos (requisitos de epic, definições de story), os verdicts são `approved` ou `rejected`. Para gates operacionais (design, execution-plan, qa-escalation), os verdicts são `approved`, `rejected` ou `needs_revision`.

Atualizações em tempo real fluem por duas fontes: **HTTP hooks** empurram eventos de telemetria (ciclo de vida de agentes, mutações de ferramentas, conclusões de tasks, rastreamento de sessões) via WebSocket conforme acontecem; **MCP** entrega notificações de approval gate quando um agente bloqueia esperando um verdict humano. O dashboard nunca lê arquivos diretamente ou observa o banco de dados — todos os dados chegam pelo CLI Trust Layer ou push WebSocket do servidor HTTP.

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
3. **Version review** (`requisitos do epic`) — dashboard apresenta os requisitos versionados para revisão humana; fluxo pausa até o verdict
4. Requisitos validados viram um plano com tarefas em DAG
5. **Approval gate** (`design`) — dashboard apresenta o documento de design arquitetural para revisão; fluxo pausa até o verdict
6. **Approval gate** opcional (`execution-plan`) — dashboard apresenta o DAG para revisão antes dos agentes serem despachados
7. Agentes executam em paralelo, seguindo TDD e padrões do projeto
8. Agente de QA valida independentemente — defeitos críticos acionam **approval gate** (`qa-escalation`) para direcionamento humano; se rejeitado, retorna à fase apropriada

**Fluxo de story (Software Engineering):**

1. Um item de trabalho já está definido ou o usuário fornece contexto direto
2. Inicia o Oráculo com `/oraculo:story` — pula Discover, vai direto para Plan
3. **Version review** (`definição da story`) — dashboard apresenta a definição versionada da story para revisão humana; fluxo pausa até o verdict
4. **Approval gate** (`design`) — dashboard apresenta o documento de design arquitetural para revisão; fluxo pausa até o verdict
5. Agentes executam em paralelo, seguindo TDD e padrões do projeto
6. Agente de QA valida independentemente — defeitos críticos acionam **approval gate** (`qa-escalation`) para direcionamento humano; se rejeitado, retorna à fase apropriada

**O Oráculo reduz a distância entre a ideia e o código de qualidade.** Ele não substitui o time — ele amplifica a capacidade do time de pensar bem e executar com rigor.
