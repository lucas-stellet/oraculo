# Oraculo UI — Design

## 1. Visão Geral

A camada de UI oferece um dashboard no navegador para monitorar e interagir com projetos Oraculo. Ela expõe os dados que a Trust Layer do CLI gerencia — épicos, stories, tarefas, dependências do DAG, atividade dos agentes, veredictos do QA e conhecimento acumulado — através de uma interface web em tempo real.

O dashboard nunca contorna o CLI. Toda operação de dados passa pelas funções internas do CLI em Go — as mesmas operações validadas e contratadas que agentes e humanos usam pelo terminal. A UI é uma consumidora pesada de leitura e leve de escrita: lê frequentemente (polling + monitoramento de arquivos), escreve apenas para aprovações e configurações.

```
oraculo (single Go binary)
├── CLI commands          (já existe)
├── HTTP server           (serve REST API + assets do frontend embutidos)
├── WebSocket server      (push em tempo real para o navegador)
├── MCP server (stdio)    (comunicação com o Claude Code)
└── Auto-open browser     (inicialização sem atrito)

Human (Browser) ←→ HTTP/WS ←→ Go HTTP server ←→ CLI internals ←→ SQLite + Markdown
Claude Code     ←→ stdio   ←→ Go MCP server  ←→ CLI internals ←→ SQLite + Markdown
```

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
| `GET /api/epics/:name/versions` | `oraculo tools epic versions <name>` | Listar versões do epic |
| `GET /api/epics/:epic/stories/:story/versions` | `oraculo tools story versions <story> --epic <epic>` | Listar versões da story |
| `GET /api/reviews/:versionId` | `oraculo tools review list <version-id> --type <epic\|story>` | Listar reviews de uma versão |
| `POST /api/reviews` | `oraculo tools review create <version-id> --type <type> --verdict <verdict> --comment <comment>` | Enviar review de documento |
| `POST /api/approvals/:id/verdict` | (estado interno) | Enviar decisão de aprovação operacional |

O servidor gerencia uma única responsabilidade com estado: a **fila de aprovações**. Solicitações de aprovação chegam via chamadas de ferramenta MCP dos agentes e ficam em memória até que o humano responda pelo dashboard. O veredicto é então retransmitido ao agente solicitante.

### 2.2 Fluxo de Inicialização

Quando o Claude Code inicia uma sessão, ele lança `oraculo mcp` como um MCP server registrado (configurado durante o `oraculo install`). A sequência de inicialização:

1. `oraculo mcp` inicia o MCP server via stdio (transporte padrão de MCP)
2. O MCP server inicia o HTTP server em uma porta disponível
3. O HTTP server começa a servir o frontend embutido e a REST API
4. O binário abre o navegador padrão do usuário em `http://localhost:<porta>`
5. File watchers começam a monitorar `.oraculo/` para detectar mudanças

Se o HTTP server já estiver em execução (de outra sessão), o novo MCP server se conecta à instância existente em vez de iniciar uma duplicata. O usuário nunca precisa executar um comando separado.

### 2.3 MCP Server

O MCP server expõe ferramentas que os agentes de IA chamam durante a orquestração. Ele roda no mesmo processo que o HTTP server e compartilha o canal de broadcast via WebSocket.

| Ferramenta MCP | Direção | Propósito |
|---|---|---|
| `request_approval` | Agente -> Dashboard | Enviar um documento para revisão humana (requisitos, story, escalação de QA) |
| `notify_agent_state` | Agente -> Dashboard | Reportar início, conclusão, falha de agente ou veredicto de QA |
| `register_project` | Agente -> Dashboard | Registrar um diretório de projeto para descoberta pelo dashboard |

`request_approval` aceita um payload com: tipo do documento, conteúdo em markdown, versão anterior opcional (para diff) e um mecanismo de callback. O dashboard retém a solicitação até que o humano aja.

`notify_agent_state` envia atualizações em tempo real que a tela do Agent Monitor consome. O payload inclui: ID do agente, tipo do agente (orchestrator/code/qa/research), referência de tarefa, tipo de evento (started/completed/failed) e metadados opcionais.

### 2.4 Comunicação em Tempo Real

WebSocket fornece atualizações push do servidor para o navegador. Duas fontes de eventos alimentam o canal WebSocket:

1. **File watcher** — Monitora `.oraculo/epics/**/*.md` para mudanças em markdowns e `.oraculo/oraculo.db-wal` para atividade no write-ahead log do SQLite. Ao detectar uma mudança, o servidor re-consulta os comandos CLI relevantes e envia os deltas.

2. **Notificações MCP** — Mudanças de estado dos agentes e solicitações de aprovação são transmitidas imediatamente ao serem recebidas.

Formato das mensagens WebSocket:

```json
{
  "type": "task_updated" | "approval_requested" | "agent_state" | "knowledge_added" | "file_changed",
  "payload": { ... }
}
```

O frontend se inscreve em tipos de evento por tela. A DAG View se inscreve em `task_updated`. O Agent Monitor se inscreve em `agent_state`. A tela de Aprovações se inscreve em `approval_requested`.

### 2.5 Fluxo de Dados

Todos os dados fluem pelas funções internas do CLI. O HTTP server chama as mesmas funções Go que os comandos CLI usam — operações validadas e contratadas com pré-condições estritas. O servidor nunca abre o SQLite diretamente e nunca escreve no sistema de arquivos fora das funções do CLI. Isso preserva o CLI como a única Trust Layer — toda validação, imposição de ciclo de vida e migração de schema permanecem em um único lugar.

```
[Browser] --HTTP--> [Go HTTP server] --in-process--> [CLI internals] --SQLite--> [.oraculo/oraculo.db]
[Browser] <--WS---- [Go HTTP server] <--file watch-- [.oraculo/epics/**/*.md]
[Agent]   --stdio-> [Go MCP server]  --WS broadcast-> [Browser]
```

## 3. Telas

### 3.1 Landing (Epic Selection)

**Propósito:** Ponto de entrada do dashboard. O usuário escolhe em qual épico quer trabalhar. Todas as demais telas ficam escopadas ao épico selecionado.

**Fontes de dados:**
- `oraculo tools epic list` — Todos os épicos com fase, contagem de stories e status agregado das tarefas

**Layout:** Título de página "Select an Epic" com subtítulo "Choose an Epic to view its stories, tasks, and agent activity." Abaixo, um grid responsivo de cards de épico. Cada card mostra: nome do épico, badge de fase (`Discover` / `Plan` / `Execute` / `Validate`), barra de progresso com conclusão de tarefas, linha de estatísticas (quantidade de stories, quantidade de tarefas, percentual concluído), badge de status (`completed` / `in_progress` / `pending`) e ação principal para abrir o épico. A tela também inclui uma ação clara para criar um novo épico.

**Interações principais:** Clique em um card ou em sua ação principal para selecionar o épico. O dropdown de épico na sidebar é atualizado e a navegação segue para a tela de Stories daquele épico. Ao criar um novo épico, a UI chama o backend, que deve materializar a pasta do épico no workspace e persistir o registro no banco através da mesma trust layer do CLI.

**Quando nenhum épico está selecionado:** O dropdown da sidebar mostra "Select Epic...". Os itens de navegação dependentes de contexto (Stories até Knowledge Base) ficam desabilitados ou visualmente atenuados. Settings permanece sempre acessível.

### 3.2 Stories

**Propósito:** Navegar pela hierarquia Story > Tarefa dentro do épico selecionado e ler documentos de requisitos.

**Fontes de dados:**
- `oraculo tools epic list` + `oraculo tools epic get <name>` — Metadados do épico e markdown
- `oraculo tools story list --epic <epic>` + `oraculo tools story get <name> --epic <epic>` — Metadados da story e markdown
- `oraculo tools task list --epic <epic> --story <story>` — Lista de tarefas com status

**Layout:** Divisão master-detail. Painel esquerdo: árvore colapsável mostrando Stories > Tarefas do épico selecionado. Painel direito: visualizador de markdown que renderiza o documento de requisitos da entidade selecionada ou o detalhe da tarefa. Badges de fase (pending, in_progress, completed, failed) aparecem ao lado de cada nó da árvore.

**Interações principais:** Selecione um nó da árvore para carregar seu conteúdo no visualizador. Expanda/colapse níveis. Filtre a árvore por status. Link de qualquer tarefa para sua posição na DAG View.

### 3.3 DAG View

**Propósito:** Visualizar grafos de dependências de tarefas para o plano de execução de uma story.

**Fontes de dados:**
- `oraculo tools task list --epic <epic> --story <story>` — Retorna tarefas com referências `depends_on`

**Layout:** Grafo dirigido em largura total. Nós representam tarefas, arestas representam dependências. O layout usa ordenação topológica da esquerda para direita. Cores dos nós codificam o status: cinza (pending), azul (in_progress), verde (completed), vermelho (failed). Um destaque de caminho crítico traça a cadeia de dependências mais longa. Rótulos de atribuição de agente aparecem abaixo de cada nó.

**Interações principais:** Passe o mouse sobre um nó para ver o tooltip de detalhe da tarefa. Clique em um nó para abrir seu detalhe em um drawer lateral. Alterne o destaque do caminho crítico. Filtre por status. Zoom e pan.

**Nota:** Este é um diferencial exclusivo do Oraculo. O grafo é computado no cliente a partir do JSON da lista de tarefas — nenhum comando CLI adicional é necessário. O campo `depends_on` em cada tarefa fornece a lista de arestas.

### 3.4 Agent Monitor

**Propósito:** Visibilidade em tempo real sobre o que os agentes estão fazendo agora.

**Fontes de dados:**
- Eventos WebSocket `agent_state` oriundos de chamadas MCP `notify_agent_state`

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
- Eventos WebSocket `agent_state` com payloads de veredicto do QA

**Layout:** Linha superior: cards de resumo (total de validações, taxa de aprovação, média de ciclos de rejeição, escalações ativas). Abaixo: uma tabela de veredictos recentes mostrando nome da tarefa, veredicto (approved/rejected), número da tentativa e timestamp. Badges de contagem de rejeição destacam tarefas se aproximando do limiar do circuit breaker (padrão: 3). Um indicador de escalação marca tarefas que ultrapassaram o limiar e requerem intervenção humana.

**Interações principais:** Clique em uma linha de veredicto para ver os achados completos do QA. Ordene/filtre por veredicto, story ou data. Link para a tarefa relacionada na DAG View.

### 3.7 Knowledge Base

**Propósito:** Navegar e buscar o conhecimento acumulado do projeto a partir da tabela de conhecimento.

**Fontes de dados:**
- `oraculo tools memory search <query>` — Busca full-text em todos os achados
- `oraculo tools memory domains` — Lista de domínios disponíveis para filtragem

**Layout:** Barra de busca no topo com filtros dropdown de domínio e categoria. Resultados aparecem como cards mostrando: badge de domínio, tag de categoria, texto do achado (truncado), indicador de confiança, arquivos de origem e timestamp. Clique em um card para expandir o achado completo.

**Interações principais:** Busca com type-ahead dispara `memory search` com debounce. Filtre por domínio (da resposta de `memory domains`) e categoria (`pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`). Ordene por confiança ou recência.

### 3.8 Configurações

**Propósito:** Configuração do projeto e preferências do dashboard.

**Layout:** Formulário em abas. Abas: Projeto (nome, caminho do diretório, caminho do binário CLI), Dashboard (intervalo de refresh, tema claro/escuro, preferências de notificação), Projetos Conectados (lista multi-projeto com adicionar/remover).

**Interações principais:** Salvar persiste em um arquivo de configuração local (`~/.oraculo/dashboard.json`). Projetos conectados permitem que o dashboard sirva múltiplos repositórios com Oraculo habilitado a partir de uma única instância.

## 4. Navegação e Layout

**Shell:** Barra lateral esquerda persistente com links de ícone + rótulo para cada tela. A sidebar colapsa para ícones apenas em viewports estreitos. A barra superior mostra o nome do projeto atual e um dropdown de troca de projeto (para o modo multi-projeto).

**Ordem da sidebar:**
1. Landing (Epic Selection)
2. Stories
3. DAG View
4. Agent Monitor
5. Aprovações (com badge de contagem não lida)
6. QA Dashboard
7. Knowledge Base
8. Configurações

**Comportamento responsivo:** A sidebar colapsa abaixo de 1024px de largura de viewport. Visualizações master-detail (Stories, Aprovações) empilham verticalmente em telas estreitas. A DAG View permanece em largura total com scroll horizontal.

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
- Radius base em degraus pequenos, médios e amplos para controlar densidade sem cair em visual “pill” por padrão
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
| Monitoramento de arquivos | `fsnotify/fsnotify` | Monitoramento de sistema de arquivos cross-platform para mudanças em `.oraculo/` |
| Embedding estático | Go `embed.FS` | Assets do frontend compilados no binário em tempo de build |
| MCP server | Go MCP SDK ou handler stdio customizado | Protocolo MCP via stdin/stdout para integração com o Claude Code |
| Integração com CLI | Chamadas de função in-process | HTTP handlers chamam as mesmas funções Go que os comandos CLI — sem overhead de subprocesso |

## 6. Fontes de Dados

Resumo de como cada tela obtém seus dados:

| Tela | Fonte primária | Mecanismo de atualização |
|---|---|---|
| Landing (Epic Selection) | CLI/API: `epic list` com agregados de stories/tarefas/fase | Polling no mount + WebSocket para invalidação leve |
| Stories | CLI: `story list`, `story get`, `task list` | File watcher em `.oraculo/epics/**/*.md` |
| DAG View | CLI: `task list` (inclui `depends_on`) | WebSocket `task_updated` |
| Agent Monitor | WebSocket: eventos `agent_state` | Somente push (sem polling) |
| Aprovações | MCP: `request_approval` | Somente push (sem polling) |
| QA Dashboard | CLI: `task list`, `task get` + WebSocket `agent_state` | Híbrido: carga inicial via CLI, atualizações via WebSocket |
| Knowledge Base | CLI: `memory search`, `memory domains` | Sob demanda (usuário dispara a busca) |
| Configurações | Arquivo de configuração local | Leitura no mount, escrita ao salvar |

## 7. Trabalho Futuro

Capacidades adiadas para iterações futuras:

- **Edição inline** — Editar o markdown de requisitos diretamente no visualizador de Stories e salvar de volta via `oraculo tools epic save` / `oraculo tools story save`.
- **Visualização de timeline** — Visualização estilo Gantt da execução de tarefas ao longo do tempo, usando os timestamps `started_at` e `completed_at`.
- **Streaming de logs dos agentes** — Transmitir stdout/stderr dos agentes em tempo real para o Agent Monitor, além dos eventos estruturados de estado.
- **Sistema de notificações** — Notificações desktop para solicitações de aprovação, escalações de QA e conclusão de épicos.
- **Controle de acesso por papel** — Restringir autoridade de aprovação e acesso às configurações por papel do usuário quando o Oraculo for usado em contextos de time.
- **Terminal embutido** — Lançar comandos do CLI `oraculo` diretamente pela interface do dashboard.
- **API do dashboard para CI** — Expor endpoints REST somente leitura para que pipelines de CI/CD consultem o status do projeto.
