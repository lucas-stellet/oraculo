# Oraculo UI — Design

## 1. Visão Geral

A camada de UI oferece um dashboard no navegador para monitorar e interagir com projetos Oraculo. Ela expõe os dados que a Trust Layer do CLI gerencia — épicos, stories, tarefas, dependências do DAG, atividade dos agentes, veredictos do QA e conhecimento acumulado — através de uma interface web em tempo real.

O dashboard nunca contorna o CLI. Toda operação de dados passa pelas funções internas do CLI em Go — as mesmas operações validadas e contratadas que agentes e humanos usam pelo terminal. A UI é uma consumidora pesada de leitura e leve de escrita: lê frequentemente (polling + file watching), escreve apenas para aprovações e configurações.

```
oraculo (single Go binary)
├── CLI commands          (já existe)
├── HTTP server           (serve REST API + assets do frontend embutidos)
├── WebSocket server      (push em tempo real para o navegador)
├── MCP server (stdio)    (comunicação com o Claude Code)
└── Auto-open browser     (inicialização sem atrito)

Human (Browser) ←→ HTTP/WS ←→ Go HTTP server ←→ CLI internals ←→ SQLite + Markdown
Claude Code     ←→ stdio   ←→ Go MCP server  ←→ CLI internals ←→ SQLite + Markdown
Claude Code     ──POST──>  HTTP server ──> SQLite + broadcast WebSocket (HTTP hooks)
```

**Arquitetura de dois canais:** O binário opera dois canais de comunicação. Canal 1 — HTTP Hooks — trata telemetria automática: o Claude Code dispara hooks em eventos de sistema (início/fim de sessão, início/fim de agente, uso de ferramenta), o HTTP server persiste metadados no SQLite e faz broadcast para clientes WebSocket. São fire-and-forget; hooks nunca bloqueiam agentes. Canal 2 — MCP — trata gates de aprovação interativos: agentes chamam `request_approval`, que bloqueia até um humano responder pelo dashboard. Cada canal faz o que faz melhor: telemetria precisa ser automática e não-bloqueante (HTTP hooks), aprovações precisam ser explícitas e bloqueantes (MCP).

Todo o sistema — CLI, HTTP server, MCP server e frontend embutido — vive em um único binário Go. O frontend é compilado em tempo de build (Next.js + shadcn/ui gera assets estáticos) e embutido via `embed.FS` do Go. Em runtime, não há Node.js, npm ou dependências externas.

## 2. Arquitetura

### 2.1 Dashboard Server

O binário Go inclui um HTTP server que serve tanto a REST API quanto os assets estáticos do frontend. Cada endpoint de API chama as funções internas do CLI diretamente (chamadas de função in-process, sem spawning de subprocesso) e retorna JSON ao frontend. O servidor não mantém estado próprio — o CLI e o SQLite são a fonte única de verdade.

| Padrão de endpoint | Comando(s) CLI | Propósito |
|---|---|---|
| `GET /api/epics` | `oraculo tools epic list` | Listar todos os épicos |
| `GET /api/epics/:name` | `oraculo tools epic get <name>` | Markdown de requisitos do épico |
| `GET /api/epics/:epic/stories` | `oraculo tools story list --epic <epic>` | Stories de um épico |
| `GET /api/epics/:epic/stories/:story` | `oraculo tools story get <name> --epic <epic>` | Markdown de requisitos da story |
| `GET /api/epics/:epic/stories/:story/tasks` | `oraculo tools task list --epic <epic> --story <story>` | Tarefas com status e dependências |
| `GET /api/epics/:epic/stories/:story/tasks/:task` | `oraculo tools task get <name> --epic <epic> --story <story>` | Detalhe de uma única tarefa com resultado |
| `GET /api/knowledge` | `oraculo tools memory search <query>` | Busca na base de conhecimento |
| `GET /api/knowledge/domains` | `oraculo tools memory domains` | Domínios disponíveis |
| `GET /api/status` | `oraculo status` (parsed) | Estatísticas agregadas do projeto |
| `GET /api/sessions` | Tabela `sessions` (SQLite) | Histórico de sessões |
| `GET /api/sessions/:id/agents` | Tabela `agents` (SQLite) | Agentes de uma sessão |
| `GET /api/sessions/:id/tool-events` | Tabela `tool_events` (SQLite) | Eventos de ferramenta de uma sessão |
| `GET /api/epics/:name/versions` | `oraculo tools epic versions <name>` | Listar versões do epic |
| `GET /api/epics/:epic/stories/:story/versions` | `oraculo tools story versions <story> --epic <epic>` | Listar versões da story |
| `GET /api/reviews/:versionId` | `oraculo tools review list <version-id> --type <epic\|story>` | Listar reviews de uma versão |
| `POST /api/reviews` | `oraculo tools review create <version-id> --type <type> --verdict <verdict> --comment <comment>` | Enviar review de documento |
| `POST /api/approvals/:id/verdict` | (estado interno) | Enviar decisão de aprovação operacional |

O servidor gerencia uma única responsabilidade com estado: a **fila de aprovações**. Solicitações de aprovação chegam via chamadas de ferramenta MCP dos agentes e ficam em memória até que o humano responda pelo dashboard. O veredicto é então retransmitido ao agente solicitante.

Para dados de telemetria (sessions, agents, tool events), o servidor consulta as tabelas do SQLite diretamente — esses registros são inseridos pelos HTTP hooks, não por comandos CLI.

### 2.2 Fluxo de Inicialização

Quando o Claude Code inicia uma sessão, ele lança `oraculo mcp` como um MCP server registrado (configurado durante o `oraculo install`). A sequência de inicialização:

1. `oraculo mcp` inicia o MCP server via stdio (transporte padrão de MCP)
2. O MCP server lê a porta alocada do projeto em `.oraculo/config` → `dashboard.port`
3. Se nenhuma porta estiver configurada, aloca a primeira porta disponível no range 3100-3199
4. O HTTP server tenta fazer bind na porta configurada
5. Se a porta estiver ocupada, faz uma varredura sequencial no range 3100-3199 pela primeira porta livre
6. A nova porta é salva de volta em `.oraculo/config`
7. Se nenhuma porta estiver disponível no range, o servidor encerra com um erro claro: "All ports in range 3100-3199 are in use"
8. O HTTP server começa a servir o frontend embutido e a REST API
9. O binário abre o navegador padrão do usuário em `http://localhost:<porta>`

O command hook `SessionStart` (`oraculo hook session-start`) dispara após o servidor estar pronto. Ele envia um health check para confirmar que o servidor está online, depois envia metadados da sessão para `/hooks/session-start`. Se o servidor estiver offline, imprime um aviso em stderr e sai com código 0 — a sessão nunca é bloqueada.

**Sistema de Portas:**

- **Range:** 3100-3199 (constantes hardcoded no binário Go)
- **Configuração:** `.oraculo/config` por projeto armazena `dashboard.port`
- **Estabilidade:** Cada projeto recebe uma porta dedicada e estável — a URL do dashboard é bookmarkable entre sessões
- **Sem config global:** O range de portas vive no binário, a alocação de porta vive no projeto

Se o HTTP server já estiver em execução (de outra sessão), o novo MCP server se conecta à instância existente em vez de iniciar uma duplicata. O usuário nunca precisa executar um comando separado.

### 2.3 MCP Server

O MCP server expõe ferramentas que os agentes de IA chamam durante a orquestração. Ele roda no mesmo processo que o HTTP server e compartilha o canal de broadcast via WebSocket.

| Ferramenta MCP | Direção | Propósito |
|---|---|---|
| `request_approval` | Agente -> Dashboard | Enviar um documento para revisão humana (requisitos, story, escalação de QA) |
| `approval_status` | Agente -> Dashboard | Consultar veredicto (fallback não-bloqueante para reconexão) |

`request_approval` aceita um payload com: tipo do documento, conteúdo em markdown, epic, story opcional. O dashboard retém a solicitação até que o humano aja. Esta é a única operação bloqueante do sistema — o agente aguarda até receber um veredicto.

`approval_status` é um fallback de polling não-bloqueante para recuperação de crash. Se um agente desconecta e reconecta, ele pode verificar se uma aprovação pendente já foi decidida.

**Ferramentas removidas:** `notify_agent_state` e `register_project` não fazem mais parte do MCP server. Telemetria de estado dos agentes agora é tratada automaticamente por HTTP hooks (`SubagentStart`/`SubagentStop`) — agentes não precisam fazer chamadas MCP explícitas para telemetria. Registro de projeto é tratado pelo `oraculo install`.

### 2.4 Comunicação em Tempo Real

WebSocket fornece atualizações push do servidor para o navegador. Dois canais alimentam o WebSocket:

1. **HTTP Hooks** — O Claude Code dispara hooks em eventos de sistema. O HTTP server persiste metadados no SQLite e faz broadcast para todos os clientes WebSocket conectados. Hooks são fire-and-forget: `200` com body vazio, não-bloqueantes. Tipos de evento:

   | Evento WebSocket | Endpoint do hook | Hook do Claude Code | Gatilho |
   |---|---|---|---|
   | `session_started` | `POST /hooks/session-start` | `SessionStart` (command) | Sessão iniciada |
   | `agent_started` | `POST /hooks/agent-start` | `SubagentStart` | Agente criado |
   | `agent_stopped` | `POST /hooks/agent-stop` | `SubagentStop` | Agente concluído/falhou |
   | `tool_used` | `POST /hooks/tool-used` | `PostToolUse` | Ferramenta de mutação usada (Bash\|Edit\|Write\|NotebookEdit) |
   | `task_completed` | `POST /hooks/task-completed` | `TaskCompleted` | Tarefa marcada como completa |
   | `agent_stopping` | `POST /hooks/stop` | `Stop` | Agente parando |
   | `teammate_idle` | `POST /hooks/teammate-idle` | `TeammateIdle` | Teammate ocioso |
   | `session_ended` | `POST /hooks/session-end` | `SessionEnd` | Sessão encerrada |

2. **MCP + CLI** — Solicitações de aprovação e reviews de versão são transmitidos quando ocorrem.

   | Evento WebSocket | Origem |
   |---|---|
   | `approval_requested` | Ferramenta MCP `request_approval` (gates operacionais) |
   | `version_created` | CLI `epic version` / `story version` (versionamento de documentos) |
   | `review_submitted` | CLI `review create` (reviews de documentos) |

Formato das mensagens WebSocket:

```json
{
  "type": "agent_started" | "agent_stopped" | "tool_used" | "task_completed" | "session_started" | "session_ended" | "approval_requested" | "version_created" | "review_submitted" | "teammate_idle" | "agent_stopping",
  "payload": { ... }
}
```

O frontend se inscreve em tipos de evento por tela. A DAG View se inscreve em `task_completed`. O Agent Monitor se inscreve em `agent_started` e `agent_stopped`. A tela de Aprovações se inscreve em `approval_requested`, `version_created` e `review_submitted`. A tela de Sessions se inscreve em `session_started` e `session_ended`.

### 2.5 Fluxo de Dados

A maior parte dos dados flui pelas funções internas do CLI. O HTTP server chama as mesmas funções Go que os comandos CLI usam — operações validadas e contratadas com pré-condições estritas. Para dados de telemetria (sessions, agents, tool events), o servidor consulta as tabelas do SQLite diretamente — esses registros são inseridos pelos HTTP hooks, não por comandos CLI. Isso preserva o CLI como a única Trust Layer para toda a lógica de negócio, ao mesmo tempo que permite ao HTTP server ler telemetria persistida por hooks diretamente.

```
[Browser] --HTTP--> [Go HTTP server] --in-process--> [CLI internals] --SQLite--> [.oraculo/oraculo.db]
[Browser] --HTTP--> [Go HTTP server] --SQLite query-> [sessions, agents, tool_events tables]
[Browser] <--WS---- [Go HTTP server] <--HTTP hooks--- [Claude Code automatic hooks]
[Agent]   --stdio-> [Go MCP server]  --WS broadcast-> [Browser]
```

## 3. Telas

### 3.1 Landing (Epic Selection)

**Propósito:** Ponto de entrada do dashboard. O usuário escolhe em qual épico quer trabalhar. Todas as demais telas ficam escopadas ao épico selecionado.

**Fontes de dados:**
- `oraculo tools epic list` — Todos os épicos com fase, contagem de stories e status de tarefas

**Layout:** Título de página "Select an Epic" com subtítulo "Choose an Epic to view its stories, tasks, and agent activity." Abaixo, um grid responsivo de cards de épico. Cada card mostra: nome do épico (header, semibold), badge de fase (Discover / Plan / Execute / Validate), barra de progresso com conclusão de tarefas, linha de estatísticas (quantidade de stories, quantidade de tarefas, percentual concluído), badge de status (completed / in_progress / pending) e um botão "Open".

**Interações principais:** Clique em um card ou em seu botão "Open" para selecionar o épico. O dropdown de épico na sidebar é atualizado e a navegação segue para a tela de Stories daquele épico. Os itens de navegação da sidebar ficam ativos assim que um épico é selecionado.

**Quando nenhum épico está selecionado:** O dropdown de épico na sidebar mostra "Select Epic...", os itens de navegação (Stories até Knowledge Base) ficam visualmente atenuados/desabilitados, e Settings permanece sempre acessível.

### 3.1.1 Home / Overview

**Propósito:** Visão geral do épico selecionado, mostrando todas as stories em formato de grid para navegação rápida.

**Fontes de dados:**
- `oraculo tools story list --epic <epic>` — Todas as stories do épico com status e progresso
- `oraculo tools task list --epic <epic> --story <story>` — Progresso de tarefas por story

**Layout:** Grid de cards de story. Cada card mostra: nome da story (título), badge de status, barra de progresso de tarefas, contagem de tarefas (completas/total), link para abrir a página dedicada da story. No topo da página, um cabeçalho com o nome do épico e fase atual.

**Interações principais:** Clique em um card de story para navegar à sua página dedicada com tabs (Tasks, Design, Requirements, QA). Filtre stories por status. A página Home é a view padrão ao selecionar um épico.

### 3.2 Stories

**Propósito:** Listar todas as stories do épico selecionado em formato de lista ou grid. Cada story é uma página dedicada com seus documentos e tarefas.

**Fontes de dados:**
- `oraculo tools story list --epic <epic>` — Stories do épico selecionado
- `oraculo tools story get <name> --epic <epic>` — Markdown de requisitos da story (página dedicada)
- `oraculo tools task list --epic <epic> --story <story>` — Tarefas com status

**Layout (lista/grid de stories):** Grid ou lista vertical de cards de story. Cada card mostra: nome da story (título), badge de status (pending, in_progress, completed, failed), barra de progresso de tarefas, contagem de tarefas completas/total. Clique no card navega para a página dedicada da story.

**Página dedicada da story:** Ao clicar em uma story, abre-se uma página exclusiva com:
- **Breadcrumb no topo:** Epic → Story → [tab atual]
- **Tabs horizontais:** [Tasks] [Design] [Requirements] [QA]
  - **Tab Tasks:** Lista de tarefas da story com status, ações e links para DAG View
  - **Tab Design:** Visualização de documentos de design associados
  - **Tab Requirements:** Markdown de requisitos da story (renderização completa)
  - **Tab QA:** Veredictos de QA, histórico de validações, tentativas

**Interações principais:** Clique em uma story para abrir sua página dedicada. Navegue entre tabs para acessar diferentes documentos e visões. Use o breadcrumb para navegar de volta ao épico ou à lista de stories.

### 3.3 DAG View

**Propósito:** Visualizar grafos de dependências de tarefas para o plano de execução de uma story.

**Fontes de dados:**
- `oraculo tools task list --epic <epic> --story <story>` — Retorna tarefas com referências `depends_on`

**Mecanismo de atualização:** Push via WebSocket a partir do evento `task_completed` (disparado pelo hook `TaskCompleted`). Mais confiável que monitoramento de arquivo WAL — o evento dispara exatamente quando uma tarefa é marcada como completa, não em cada escrita no banco.

**Layout:** Grafo dirigido em largura total. Nós representam tarefas, arestas representam dependências. O layout usa ordenação topológica da esquerda para direita. Cores dos nós codificam o status: cinza (pending), azul (in_progress), verde (completed), vermelho (failed). Um destaque de caminho crítico traça a cadeia de dependências mais longa. Rótulos de atribuição de agente aparecem abaixo de cada nó.

**Interações principais:** Passe o mouse sobre um nó para ver o tooltip de detalhe da tarefa. Clique em um nó para abrir seu detalhe em um drawer lateral. Alterne o destaque do caminho crítico. Filtre por status. Zoom e pan.

**Nota:** Este é um diferencial exclusivo do Oraculo. O grafo é computado no cliente a partir do JSON da lista de tarefas — nenhum comando CLI adicional é necessário. O campo `depends_on` em cada tarefa fornece a lista de arestas.

### 3.4 Agent Monitor

**Propósito:** Visibilidade em tempo real sobre o que os agentes estão fazendo agora.

**Fontes de dados:**
- Tabela `agents` (SQLite) — Estado atual e histórico dos agentes, populada por HTTP hooks
- Eventos WebSocket `agent_started` e `agent_stopped` — Push dos hooks `SubagentStart`/`SubagentStop`

**Mecanismo de atualização:** Somente push via WebSocket (sem polling). Agentes não precisam fazer chamadas MCP explícitas — o Claude Code dispara os hooks automaticamente em cada evento de ciclo de vida dos agentes.

**Layout:** Grade de cards de agentes, um por agente ativo. Cada card mostra: ícone do tipo de agente (orchestrator/code/qa/research), nome da tarefa atual, tempo decorrido, indicador de status. Abaixo da grade, um feed de atividade com scroll mostrando eventos com timestamp (agente iniciado, tarefa concluída, veredicto do QA emitido, escalação acionada).

**Interações principais:** Clique em um card de agente para ver seu histórico completo da sessão atual. Clique em uma referência de tarefa no feed para navegar à DAG View. Filtre o feed por tipo de agente ou tipo de evento.

### 3.5 Aprovações

**Propósito:** Gate de revisão humana para versões de documentos (requisitos de epic, definições de story) e aprovações operacionais (escalações de QA, design, execution-plan).

**Fontes de dados:**
- Para reviews de documentos: `oraculo tools epic versions <nome-epic>` / `oraculo tools story versions <nome-story> --epic <nome-epic>` — Carregam histórico de versões
- Para gates operacionais: chamadas MCP `request_approval` preenchem a fila
- Diff entre versões usa as tabelas `epic_versions` / `story_versions`

**Layout:** Coluna esquerda: fila de reviews/aprovações ordenada por hora de chegada, com badges de tipo (epic-version/story-version/qa-escalation/design/execution-plan) e indicadores de tempo de espera. Coluna direita: visualizador rico de markdown com syntax highlighting. Para versões de documentos, um toggle alterna entre a visualização renderizada e o diff lado a lado contra a versão anterior. Abaixo do visualizador: campo de comentário inline e botões de ação. Para reviews de documentos: dois botões (Aprovar, Rejeitar). Para gates operacionais: três botões (Aprovar, Rejeitar, Precisa de Revisão).

**Interações principais:** Selecione um item da fila para carregar seu conteúdo. Alterne o modo diff. Adicione comentários (anexados ao veredicto). Envie o veredicto — o servidor o retransmite ao agente solicitante via callback MCP.

### 3.6 QA Dashboard

**Propósito:** Acompanhar resultados de validação em todo o projeto.

**Fontes de dados:**
- `oraculo tools task list --epic <epic> --story <story>` — Status das tarefas (completed/failed implica ciclo de QA)
- `oraculo tools task get <name> --epic <epic> --story <story>` — Resultado da tarefa com resumo e logs
- Eventos WebSocket `agent_stopped` com payloads de veredicto do QA

**Mecanismo de atualização:** Híbrido — carga inicial via CLI, atualizações via push WebSocket dos hooks `SubagentStop`.

**Layout:** Linha superior: cards de resumo (total de validações, taxa de aprovação, média de ciclos de rejeição, escalações ativas). Abaixo: uma tabela de veredictos recentes mostrando nome da tarefa, veredicto (approved/rejected), número da tentativa e timestamp. Badges de contagem de rejeição destacam tarefas se aproximando do limiar do circuit breaker (padrão: 3). Um indicador de escalação marca tarefas que ultrapassaram o limiar e requerem intervenção humana.

**Interações principais:** Clique em uma linha de veredicto para ver os achados completos do QA. Ordene/filtre por veredicto, story ou data. Link para a tarefa relacionada na DAG View.

### 3.7 Knowledge Base

**Propósito:** Navegar e buscar o conhecimento acumulado do projeto a partir da tabela de conhecimento.

**Fontes de dados:**
- `oraculo tools memory search <query>` — Busca full-text em todos os achados
- `oraculo tools memory domains` — Lista de domínios disponíveis para filtragem

**Mecanismo de atualização:** Sob demanda (usuário dispara a busca). Sem push em tempo real necessário.

**Layout:** Barra de busca no topo com filtros dropdown de domínio e categoria. Resultados aparecem como cards mostrando: badge de domínio, tag de categoria, texto do achado (truncado), indicador de confiança, arquivos de origem e timestamp. Clique em um card para expandir o achado completo.

**Interações principais:** Busca com type-ahead dispara `memory search` com debounce. Filtre por domínio (da resposta de `memory domains`) e categoria (`pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`). Ordene por confiança ou recência.

### 3.8 Sessions

**Propósito:** Histórico e observabilidade de sessões do Claude Code. Mostra quais sessões foram rastreadas, quais modelos foram usados, e a atividade associada de agentes e ferramentas.

**Fontes de dados:**
- Tabela `sessions` (SQLite) — Sessões registradas via hooks `session-start`/`session-end`
- Tabela `agents` (SQLite) — Agentes associados a cada sessão
- Tabela `tool_events` (SQLite) — Eventos de ferramenta por sessão

**Mecanismo de atualização:** REST API + push WebSocket nos eventos `session_started` e `session_ended`.

**Schema SQLite:**

```sql
-- sessions: rastreia sessões do Claude Code observadas via hooks
CREATE TABLE sessions (
    id          TEXT PRIMARY KEY,   -- Identificador da sessão do Claude Code
    model       TEXT,               -- Modelo em uso (ex: claude-opus-4-6)
    cwd         TEXT,               -- Diretório de trabalho
    started_at  TEXT NOT NULL,      -- Timestamp ISO 8601
    ended_at    TEXT,               -- Timestamp ISO 8601 (null enquanto ativa)
    end_reason  TEXT                -- Motivo do encerramento da sessão
);

-- agents: rastreia eventos de ciclo de vida dos agentes via hooks SubagentStart/Stop
CREATE TABLE agents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    name        TEXT NOT NULL,      -- Nome/identificador do agente
    type        TEXT NOT NULL,      -- code, qa, research, orchestrator
    status      TEXT NOT NULL DEFAULT 'active',  -- active, completed, failed
    started_at  TEXT NOT NULL,      -- Timestamp ISO 8601
    stopped_at  TEXT               -- Timestamp ISO 8601 (null enquanto ativo)
);

-- tool_events: rastreia uso de ferramentas de mutação (apenas metadados — sem conteúdo, sem diffs)
CREATE TABLE tool_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id),
    tool_name   TEXT NOT NULL,      -- Bash, Edit, Write, NotebookEdit
    file_path   TEXT,              -- Caminho do arquivo (null para Bash)
    timestamp   TEXT NOT NULL      -- Timestamp ISO 8601
);
```

**Layout:** Lista de sessões com indicador de status (active/ended), modelo utilizado, diretório de trabalho e duração. Clicar em uma sessão expande os agentes e eventos de ferramenta associados.

**Interações principais:** Filtre por status (active/ended), modelo ou data. Visualize detalhes de agentes e ferramentas por sessão.

### 3.9 Configurações

**Propósito:** Configuração do projeto e preferências do dashboard.

**Layout:** Formulário em abas. Abas: Projeto (nome, caminho do diretório, caminho do binário CLI), Dashboard (intervalo de refresh, tema claro/escuro, preferências de notificação).

**Interações principais:** Salvar persiste em `.oraculo/config`. A tela de Configurações é sempre acessível independente da seleção de épico.

## 4. Navegação e Layout

**Shell:** Barra lateral esquerda persistente com links de ícone + rótulo para cada tela. A sidebar colapsa para ícones apenas em viewports estreitos. O header da sidebar contém o nome da marca ("Oraculo") e um dropdown seletor de épico no topo.

**Dropdown de épico (Context Switcher):** Localizado no topo da sidebar, sempre visível. Mostra o épico atual com um chevron. Clicar abre um dropdown para alternar entre épicos. Quando nenhum épico está selecionado, exibe "Selecionar Épico..." e os itens de navegação abaixo ficam visualmente atenuados.

**Ordem da sidebar:**
1. 🏠 Home / Overview (ícone: home) ← Visão geral do épico (todas stories em grid)
2. 📋 Stories (ícone: folder-tree) ← Lista vertical ou grid de stories
3. 🕸️ DAG View (ícone: git-branch)
4. 🤖 Agent Monitor (ícone: bot)
5. 📄 Approvals (ícone: shield-alert, com badge de contagem não lida)
6. 🧪 QA Dashboard (ícone: shield-check)
7. 🧠 Knowledge Base (ícone: brain)
— separador —
8. Configurações (ícone: settings)

**Navegação plana (não nested):** Cada story é uma página dedicada. Clique em uma story na lista/grid abre sua página dedicada. Dentro da página da story: tabs horizontais [Tasks] [Design] [Requirements] [QA]. Breadcrumb no topo: Epic → Story → [tab atual].

**Quando um épico está selecionado:** Todos os itens de navegação (1-7) ficam ativos. A view padrão é Home/Overview. Alternar épicos via dropdown navega para Home do novo épico. Agent Monitor e Sessions mostram dados da sessão atual independente do épico selecionado.

**Quando nenhum épico está selecionado (Landing):** Os itens de navegação 1-7 ficam cinza/desabilitados. Apenas Configurações é interativo. A área de conteúdo principal mostra o grid de seleção de épico.

**Comportamento responsivo:** A sidebar colapsa abaixo de 1024px de largura de viewport. Visualizações master-detail empilham verticalmente em telas estreitas. A DAG View permanece em largura total com scroll horizontal.

**Tema:** Modos claro e escuro. A direção visual default é `Mission Control`: base `Slate`, acento `Blue`, superfícies frias e contraste alto para leitura prolongada. Duas direções alternativas ficam documentadas no design system para usos específicos: `Operations Warm` (base `Stone`, acento `Orange`) para fluxos mais editoriais de revisão, e `Signal Green` (base `Zinc`, acento `Green`) para monitoramento e QA. Cores semânticas de status permanecem separadas do acento primário: azul para em andamento, verde para concluído, vermelho para falhou, âmbar para atenção/aprovação pendente.

### 4.1 Foundations do Design System

O dashboard adota um design system explícito antes da implementação tela por tela. Ele define foundations, tokens e componentes-base reutilizáveis para evitar decisões visuais ad-hoc conforme novas telas forem sendo adicionadas.

**Tipografia:**
- `Space Grotesk` — família de destaque para títulos de página, títulos de seção e métricas de alto contraste
- `IBM Plex Sans` — família base para navegação, corpo, labels, botões, tabelas e leitura contínua
- `IBM Plex Mono` — família operacional para IDs, timestamps, paths, counters técnicos e telemetria

**Pesos e uso:**
- `Space Grotesk 700` — títulos de tela e valores de métrica em destaque
- `Space Grotesk 600` — títulos de seção e títulos principais de card
- `IBM Plex Sans 400` — body text, descrições e preview de markdown
- `IBM Plex Sans 500` — labels, botões, tabs e cabeçalhos de tabela
- `IBM Plex Sans 600` — active nav, selected states e ênfase curta
- `IBM Plex Mono 500` — valores operacionais e strings técnicas

**Tokens tipográficos iniciais:**
- `display-lg` — `40/48`, `Space Grotesk 700`, uso em títulos de tela
- `heading-md` — `24/32`, `Space Grotesk 600`, uso em títulos de seção
- `body-md` — `16/24`, `IBM Plex Sans 400`, uso em texto corrido da interface
- `label-sm` — `14/20`, `IBM Plex Sans 500`, uso em labels e metadata curta
- `mono-sm` — `13/18`, `IBM Plex Mono 500`, uso em IDs, timestamps e paths

**Escalas e superfícies:**
- Spacing base em `8 / 12 / 16 / 24 / 32 / 40`
- Radius base em degraus pequenos, médios e amplos para controlar densidade sem cair em visual "pill" por padrão
- `background` como tom mais profundo do canvas, nunca usado para enfatizar interação
- `card` como superfície padrão de leitura, com borda fina e respiro generoso
- `secondary` como surface de apoio para metadata, estados inset e agrupamentos discretos

**Semântica e iconografia:**
- O acento primário nunca substitui cores semânticas de status
- Status continuam explícitos em `pending`, `in_progress`, `completed` e `failed`
- Iconografia deve permanecer outline, com labels em `IBM Plex Sans 500` e uso disciplinado de cor semântica

**Inventário inicial de componentes:**
- `sidebar item`
- `topbar slot`
- `metric card`
- `status badge`
- `table shell`
- `markdown panel`

Componentes futuros previstos:
- `DAG node`
- `approval queue item`
- `diff viewer`
- `agent activity card`
- `knowledge result card`

## 5. Stack Tecnológica

| Camada | Tecnologia | Justificativa |
|---|---|---|
| **Frontend (build-time)** | | |
| Framework | Next.js | Export estático gera bundle otimizado de HTML/CSS/JS |
| Biblioteca de componentes | shadcn/ui | Componentes de alta qualidade e composíveis; sem dependência em runtime |
| Estilização | Tailwind CSS | Utility-first para iteração rápida de UI, dark mode embutido |
| Tipografia | Space Grotesk + IBM Plex Sans + IBM Plex Mono | Separa hierarquia de display, legibilidade de interface e telemetria operacional |
| Renderização de grafos | D3.js ou dagre | Computação de layout do DAG e renderização SVG para o grafo de dependências |
| Renderização de markdown | react-markdown + remark-gfm | Renderiza documentos de requisitos com markdown estilo GitHub |
| **Backend (runtime)** | | |
| HTTP server | Go `net/http` | Biblioteca padrão, sem dependência externa, pronto para produção |
| WebSocket | `gorilla/websocket` ou `nhooyr.io/websocket` | Implementações maduras de WebSocket em Go |
| Hook endpoints | Go `net/http` (mesmo servidor) | Recebe requisições POST fire-and-forget dos hooks do Claude Code |
| Embedding estático | Go `embed.FS` | Assets do frontend compilados no binário em tempo de build |
| MCP server | Go MCP SDK ou handler stdio customizado | Protocolo MCP via stdin/stdout para integração com o Claude Code |
| Integração com CLI | Chamadas de função in-process | HTTP handlers chamam as mesmas funções Go que os comandos CLI — sem overhead de subprocesso |

**Dependência removida:** `fsnotify/fsnotify` — File watcher não é mais necessário. Dados em tempo real chegam via HTTP hooks em vez de monitoramento de sistema de arquivos.

## 6. Fontes de Dados

Resumo de como cada tela obtém seus dados sob a arquitetura de dois canais:

| Tela | Fonte primária | Mecanismo de atualização |
|---|---|---|
| Landing | CLI: `epic list` | Polling no mount (sem mudança) |
| Home / Overview | CLI: `story list --epic`, `task list` | Polling + WebSocket `task_completed` |
| Stories (página dedicada) | CLI: `story get`, `task list` | Polling + WebSocket `task_completed` |
| DAG View | CLI: `task list` (inclui `depends_on`) | Push WebSocket do hook `TaskCompleted` |
| Agent Monitor | Tabela `agents` (SQLite) + WebSocket | Push WebSocket dos hooks `SubagentStart`/`SubagentStop` |
| Aprovações | Tabela `approvals` + tabelas `epic_versions`/`story_versions` (SQLite) + WebSocket | Push WebSocket do MCP `request_approval` (gates operacionais) e `version_created`/`review_submitted` (reviews de documentos) |
| QA Dashboard | CLI: `task list`, `task get` + WebSocket | Híbrido: carga inicial via CLI, atualizações via WebSocket |
| Knowledge Base | CLI: `memory search`, `memory domains` | Sob demanda (usuário dispara a busca) |
| Sessions | Tabela `sessions` (SQLite) | REST API + push WebSocket nos eventos session start/end |
| Configurações | Arquivo de configuração local | Leitura no mount, escrita ao salvar |

**O que mudou em relação ao design anterior:**

| Aspecto | Antes | Depois |
|---|---|---|
| Agent Monitor | Eventos MCP `notify_agent_state` (exigia instrumentação do agente) | HTTP hooks automáticos `SubagentStart`/`SubagentStop` |
| DAG View | File watcher em `.oraculo/oraculo.db-wal` | Hook `TaskCompleted` via push WebSocket |
| Home / Overview | Não existia | Novo — visão geral do épico com grid de stories |
| Stories | Tree view master-detail | Página dedicada por story com tabs (Tasks, Design, Requirements, QA) |
| Navegação | Lista nested de stories/tarefas | Navegação plana, cada story é uma página |
| Sessions | Não existia | Novo — tabela `sessions` via hooks `session-start`/`session-end` |
| Aprovações | MCP `request_approval` | Sem mudança — MCP permanece para gates bloqueantes |

## 7. Referência de Endpoints de Hook

Esta seção documenta os endpoints de HTTP hook que alimentam dados em tempo real no dashboard. Cada endpoint é invocado pelo Claude Code automaticamente — sem necessidade de instrumentação dos agentes. O servidor persiste metadados no SQLite e faz broadcast de um evento WebSocket para todos os clientes do dashboard conectados.

Todos os endpoints aceitam requisições POST com `timeout: 5` (5 segundos) e retornam `200` com body vazio. Se o servidor estiver inacessível, o Claude Code registra um aviso não-bloqueante e o agente continua sem ser afetado.

| Endpoint | Hook do Claude Code | Evento WebSocket | Persiste em |
|---|---|---|---|
| `POST /hooks/session-start` | `SessionStart` (command hook) | `session_started` | Tabela `sessions` |
| `POST /hooks/agent-start` | `SubagentStart` | `agent_started` | Tabela `agents` |
| `POST /hooks/agent-stop` | `SubagentStop` | `agent_stopped` | Tabela `agents` |
| `POST /hooks/tool-used` | `PostToolUse` (Bash\|Edit\|Write\|NotebookEdit) | `tool_used` | Tabela `tool_events` |
| `POST /hooks/task-completed` | `TaskCompleted` | `task_completed` | (metadados do evento) |
| `POST /hooks/stop` | `Stop` | `agent_stopping` | (metadados do evento) |
| `POST /hooks/teammate-idle` | `TeammateIdle` | `teammate_idle` | (metadados do evento) |
| `POST /hooks/session-end` | `SessionEnd` | `session_ended` | Tabela `sessions` |

**Nota sobre `session-start`:** O hook `SessionStart` é o único command hook (não um HTTP hook direto). Ele executa `oraculo hook session-start`, que realiza um health check, imprime um aviso em stderr se o servidor estiver offline, e envia metadados da sessão via HTTP se o servidor estiver online. Sempre sai com código 0 — a sessão nunca é bloqueada.

**Nota sobre `tool-used`:** Apenas ferramentas de mutação são rastreadas (Bash, Edit, Write, NotebookEdit). Ferramentas somente leitura (Read, Glob, Grep, WebFetch) são excluídas para reduzir ruído. Apenas metadados são armazenados — sem conteúdo, sem diffs, sem texto de comando.

**Tratamento de erros:** Todos os endpoints de HTTP hook seguem degradação graciosa. Se o servidor estiver offline, o Claude Code registra um aviso não-bloqueante e o agente continua. Hooks nunca bloqueiam agentes — telemetria é best-effort.

## 8. Trabalho Futuro

Capacidades adiadas para iterações futuras:

- **Edição inline** — Editar o markdown de requisitos diretamente no visualizador de Stories e salvar de volta via `oraculo tools epic save` / `oraculo tools story save`.
- **Visualização de timeline** — Visualização estilo Gantt da execução de tarefas ao longo do tempo, usando os timestamps `started_at` e `completed_at`.
- **Streaming de logs dos agentes** — Transmitir stdout/stderr dos agentes em tempo real para o Agent Monitor, além dos eventos estruturados de estado.
- **Sistema de notificações** — Notificações desktop para solicitações de aprovação, escalações de QA e conclusão de épicos.
- **Controle de acesso por papel** — Restringir autoridade de aprovação e acesso às configurações por papel do usuário quando o Oraculo for usado em contextos de time.
- **Terminal embutido** — Lançar comandos do CLI `oraculo` diretamente pela interface do dashboard.
- **API do dashboard para CI** — Expor endpoints REST somente leitura para que pipelines de CI/CD consultem o status do projeto.
- **Enforcement de políticas via PreToolUse** — Usar o hook `PreToolUse` para bloquear operações perigosas com base em regras definidas no dashboard (adiado até a arquitetura de dois canais ser validada em produção).
