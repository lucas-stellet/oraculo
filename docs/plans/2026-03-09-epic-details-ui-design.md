# Epic Details UI Design Plan

> **For Claude:** Use superpowers:executing-plans ou mcp__pencil__batch_design para implementar este design no arquivo dashboard.pen.

**Goal:** Criar design visual da tela Epic Details com navegação Opção B (Context Switcher), incluindo sidebar colapsável, Home/Overview com grid de stories, e Story Detail Page com tabs horizontais.

**Architecture:** O design será criado no arquivo `docs/ui/dashboard.pen` usando o MCP Pencil, adicionando 3 novas seções após o Design System existente:
1. Sidebar & Navigation — estrutura de navegação com Context Switcher
2. Home / Overview — grid de stories do épico selecionado
3. Story Detail Page — página dedicada com tabs (Tasks, Design, Requirements, QA)

**Tech Stack:**
- Pencil MCP para edição do arquivo .pen
- Design System existente: Mission Control palette (Slate base, Blue accent)
- Typography: Space Grotesk (títulos), IBM Plex Sans (corpo), IBM Plex Mono (operacional)

---

## Design Foundations (do Design System existente)

### Cores (Mission Control Palette)
- Background: `#020617` (deep slate)
- Card: `#0f172a` (slate card surface)
- Primary: `#2563eb` (blue accent)
- Ring: `#60a5fa` (light blue)
- Success: `#22c55e` (green)
- Warning: `#f59e0b` (amber)
- Foreground: `#f5f9ff` (text primary)
- Muted Foreground: `#8ea2bd` (text secondary)
- Border: `#22324a` (card stroke)

### Typography
- Display: Space Grotesk 700, 40/48 (títulos de tela)
- Heading: Space Grotesk 600, 24/32 (títulos de seção)
- Body: IBM Plex Sans 400, 16/24 (corpo)
- Label: IBM Plex Sans 500, 14/20 (labels, buttons)
- Mono: IBM Plex Mono 500, 13/18 (IDs, timestamps)

### Spacing
- Scale: `8 / 12 / 16 / 24 / 32 / 40`
- Radius: `8 / 16 / 24` (small, medium, large)
- Card padding: 20
- Gap entre cards: 20

### Status Badges (existing)
- pending: `#3f3f46` com texto `#fafafa`
- in_progress: `#1d4ed8` com texto `#eff6ff`
- completed: `#15803d` com texto `#f0fdf4`
- failed: `#b91c1c` com texto `#fef2f2`

---

## Seção 1: Sidebar & Navigation

### Estrutura da Sidebar
- Largura expandida: 260px
- Largura colapsada: 64px (apenas ícones)
- Background: `--sidebar` (tema Dark: `#18181b`)
- Border: `--sidebar-border` (`#ffffff1a`)

### Context Switcher (topo)
- Dropdown com nome do Épico selecionado
- Ícone de chevron para expandir
- Quando colapsado: apenas ícone de "pasta"

### Itens de Navegação
1. 🏠 Home / Overview
2. 📋 Stories
3. 🕸️ DAG View
4. 🤖 Agent Monitor
5. 📄 Approvals (com badge de contador)
6. 🧪 QA Dashboard
7. 🧠 Knowledge Base
8. Sessions (novo)
— separator —
9. ⚙️ Settings

### Comportamento de Colapso
- Colapsada: mostra apenas ícones (48x48px centralizados)
- Hover em ícone: tooltip flutuante com nome + ícone maior
- Toggle button: ícone de sidebar no canto inferior esquerdo
  - Sidebar aberta: toggle no topo
  - Sidebar fechada: toggle embaixo

---

## Seção 2: Home / Overview

### Layout
- Grid 3 colunas de Story Cards
- Cada card mostra:
  - Header: Nome da story + badge de status
  - Corpo: Barra de progresso (tarefas completadas/total)
  - Footer: Última execução, status de execução ativa, botão "Open"

### Story Card Spec
```
┌─────────────────────────────────────┐
│ 📝 Registro de Gastos      [🟡]    │  ← Header
├─────────────────────────────────────┤
│                                     │
│  ████████░░░░  67% (8/12)          │  ← Progresso
│                                     │
├─────────────────────────────────────┤
│ 🔄 Executando agora                 │  ← Status
│ 🕐 Última: há 2 min                 │
│ [Open →]                            │  ← Ação
└─────────────────────────────────────┘
```

### Dimensões
- Card width: fill_container (grid column)
- Card height: ~180-200px
- Card padding: 24
- Gap entre cards: 20
- Radius: 20 (large)
- Stroke: 1px `#22324a`

---

## Seção 3: Story Detail Page

### Breadcrumb Navigation
```
Home › Stories › Registro de Gastos
```
- Font: IBM Plex Sans 500, 14/20
- Cor: `--muted-foreground` (`#8ea2bd`)
- Separator: `›` (chevron right)

### Horizontal Tabs
```
[ Tasks ] [ Design ] [ Requirements ] [ QA ]
```
- Tab ativa: fundo `--primary` (`#2563eb`), texto white
- Tab inativa: fundo transparente, texto `--muted-foreground`
- Hover: fundo `--sidebar-accent`
- Font: IBM Plex Sans 500, 14/20
- Padding: 10/16
- Radius: 8 (small)

### Tab Content Areas

#### Tasks Tab
- Lista vertical de tarefas com:
  - Checkbox de completude
  - Nome da tarefa
  - Badge de status
  - Agente responsável (ícone)
  - Tempo estimado/conclusão

#### Design Tab
- Markdown viewer do documento de design
- Header: "Design Document" + badge de versão
- Footer: "Last updated by [agent] at [timestamp]"

#### Requirements Tab
- Markdown viewer do documento de requirements
- Header: "Requirements Document" + badge de versão
- Footer: Review status (approved/rejected/pending)

#### QA Tab
- Lista de ciclos de QA
- Cada ciclo mostra:
  - Tentativa #N
  - Verdict (approved/rejected)
  - Data/hora
  - Agente QA
  - Link para logs

---

## Componentes Novos Necessários

1. **Context Switcher Dropdown** — seleção de épico no topo da sidebar
2. **Sidebar Toggle Button** — ícone para colapsar/expandir
3. **Tooltip** — hover em ícones da sidebar colapsada
4. **Breadcrumb** — navegação hierárquica
5. **Horizontal Tabs** — navegação entre seções da story
6. **Story Card (Home)** — card resumo da story
7. **Task List Item** — item de tarefa na tab Tasks
8. **QA Cycle Card** — card de ciclo de QA na tab QA

---

## Critérios de Aceite

- [ ] Sidebar com 2 estados (colapsada/expandada) visualmente claros
- [ ] Context Switcher no topo da sidebar com dropdown funcional
- [ ] Tooltip aparece no hover de ícones quando colapsada
- [ ] Toggle button posicionado corretamente (topo quando aberta, baixo quando fechada)
- [ ] Home/Overview grid com 3 colunas responsivas
- [ ] Story Card mostra todas as informações requeridas
- [ ] Breadcrumb navegação visualmente clara
- [ ] Horizontal tabs com estado ativo/inativo claro
- [ ] 4 tabs de conteúdo (Tasks, Design, Requirements, QA) com layouts distintos
- [ ] Cores, fontes e espaçamentos seguem Design System existente

---

## Próximos Passos

1. Criar novo frame "Epic Details Screen" no dashboard.pen
2. Adicionar Sidebar com Context Switcher
3. Adicionar Home/Overview com grid de story cards
4. Adicionar Story Detail Page com breadcrumb + tabs
5. Revisar com screenshot
6. Commit do arquivo .pen
