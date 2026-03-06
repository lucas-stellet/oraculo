# Oraculo Dashboard Home Design

**Date:** 2026-03-06

**Scope:** Tela A do dashboard do Oraculo, criada do zero no Figma, baseada nos documentos de design e filosofia da UI.

## Goal

Definir a primeira tela do dashboard do Oraculo como uma Home/Dashboard em estilo Mission Control, com foco principal em progresso do projeto, alinhada aos contratos de dados descritos na especificação e aos princípios de observação, situational awareness e information radiator descritos na filosofia da UI.

## Constraints

- A tela deve ser desktop-first em `1920x1080`.
- O resultado deve ser hi-fi pronto para produto, não wireframe.
- A composição precisa seguir os dados reais da Home/Dashboard definidos em `docs/ui/design.pt-br.md`.
- A linguagem visual precisa seguir `docs/ui/philosophy.md`: mission control, não cockpit; observar, não interferir; clareza acima de controle.
- A tela será criada no Figma; o mockup do Pencil será tratado depois como uma segunda tela separada.

## Source Documents

- `docs/ui/design.pt-br.md`
- `docs/ui/philosophy.pt-br.md`
- `docs/ui/philosophy.md`

## Approved Direction

### Visual Direction

- Estética `Mission Control`.
- Superfície densa, sóbria e operacional.
- Visual hi-fi pronto para produto.
- Desktop fixed canvas em `1920x1080`.

### Functional Priority

- A Home prioriza `progresso do projeto`.
- A tela deve responder rapidamente:
  - quanto já avançou,
  - em quais épicos o trabalho está concentrado,
  - qual é a fase atual do épico mais recentemente ativo,
  - quantos agentes estão trabalhando agora.

## Screen Composition

### Shell

- Sidebar esquerda persistente com a navegação principal do dashboard.
- Topbar no conteúdo com contexto do projeto e estado de atualização.

### Summary Row

Quatro summary cards no topo, exatamente alinhados ao contrato da Home:

- `Total Epics`
- `Total Stories`
- `Task Completion`
- `Active Agents`

Esses cards funcionam como information radiator de relance. Eles não devem parecer widgets financeiros ou painéis de BI; devem comunicar estado do sistema com legibilidade imediata.

### Main Area

O núcleo da tela é uma `Epic Progress Table`.

Cada linha da tabela deve enfatizar:

- nome do épico,
- fase atual,
- contagem de stories,
- estado agregado das tarefas,
- progresso inline por barras ou trilhas de conclusão.

Essa tabela é o principal mecanismo de compreensão da tela. A Home existe para tornar o estado global do projeto observável sem exigir navegação inicial.

### Right Rail

A lateral direita mostra `Current Phase Focus` para o épico mais recentemente ativo.

Esse painel deve:

- destacar visualmente a fase atual (`Discover`, `Plan`, `Execute`, `Validate`),
- mostrar a trilha completa das fases,
- comunicar o significado da fase atual de forma curta e contextual.

Esse bloco existe para apoiar comprehension e projection, não para introduzir alertas genéricos de gestão.

### Lower Pulse Area

Uma faixa inferior enxuta mostra `System Pulse`.

Esse bloco pode incluir:

- última sincronização,
- atividade recente de agentes em nível resumido,
- pequenos sinais de movimento do sistema.

Ele não deve competir com a tabela principal nem se transformar em uma auditoria completa.

## Interaction Model

- Clique em uma linha de épico navega ao Epic Explorer.
- Clique em `Active Agents` navega ao Agent Monitor.
- A Home não deve expor ações operacionais pesadas ou controles que sugiram micromanagement de agentes.

## Visual Rules

- Base escura grafite, evitando preto puro.
- Contraste controlado para suportar alta densidade.
- Cor semântica disciplinada:
  - azul para andamento,
  - verde para concluído,
  - âmbar para pendência/atenção,
  - vermelho apenas para bloqueios ou falhas reais.
- Grid rígido e alinhamento preciso.
- Elemento hero é a leitura estruturada do progresso por épico, não gráficos decorativos.

## Explicitly Rejected Elements

Os seguintes elementos foram descartados por não refletirem bem o modelo do Oraculo para a Home:

- `Attention Required`
- `Upcoming Milestones`
- `Stories Without Progress`

Esses blocos introduzem semântica genérica de project management e desviam do contrato funcional definido para a Home/Dashboard.

## Deliverable

Criar no Figma uma primeira tela `Tela A` com composição hi-fi em `1920x1080`, pronta para servir como base visual da implementação futura da Home/Dashboard do Oraculo.
