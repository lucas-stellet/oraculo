# Oraculo Epic Selection Design

**Date:** 2026-03-06

**Scope:** Primeira tela do dashboard do Oraculo, responsável por selecionar o épico ativo e permitir a criação de um novo épico sem sair da UI.

## Goal

Definir a primeira entrega de UI como `Landing (Epic Selection)`, preservando a filosofia de `Mission Control` e respeitando o princípio de que toda mutação passa pela trust layer do CLI.

## Constraints

- A primeira tela precisa funcionar com a base técnica que já existe hoje: servidor HTTP em Go, stores SQLite, hooks e WebSocket.
- O repositório ainda não possui app frontend separado, pipeline de build web ou assets embutidos para a UI.
- A criação de épico pela UI deve produzir o mesmo efeito esperado pelo workflow do produto: pasta materializada no workspace e registro persistido no banco.
- A implementação deve seguir o design system já definido em `docs/ui/design.pt-br.md`: tipografia `Space Grotesk` + `IBM Plex Sans` + `IBM Plex Mono`, spacing `8 / 12 / 16 / 24 / 32 / 40`, superfícies frias e semântica disciplinada.

## Source Documents

- `docs/ui/design.md`
- `docs/ui/design.pt-br.md`
- `docs/ui/philosophy.md`
- `docs/ui/philosophy.pt-br.md`

## Approaches Considered

### 1. Go-rendered HTML + embedded CSS/JS

Renderizar a tela diretamente no servidor Go com `html/template`, assets estáticos embutidos e uma camada leve de JavaScript para buscar dados e submeter criação de épico.

**Vantagens**
- Aproveita a stack já existente.
- Mantém o binário único sem introduzir toolchain frontend agora.
- É suficiente para uma tela principalmente orientada a leitura, com pouca interatividade.

**Desvantagens**
- Menos ergonomia para componentes ricos no futuro.
- Exige disciplina manual para manter componentes e tokens consistentes.

### 2. SPA pequena embutida no binário

Criar um frontend dedicado mínimo, compilar assets estáticos e servi-los pelo Go.

**Vantagens**
- Separa melhor UI e backend.
- Facilita a expansão para telas futuras.

**Desvantagens**
- Introduz bootstrap de frontend antes do primeiro valor entregue.
- Aumenta escopo da primeira implementação.

### 3. Fundar a stack completa de produto agora

Introduzir desde já uma stack web completa, mais próxima da visão final do documento de arquitetura.

**Vantagens**
- Aproxima a base do estado final pretendido.

**Desvantagens**
- Escopo grande demais para a primeira tela.
- Alto risco de gastar esforço em infraestrutura antes de validar a superfície inicial.

## Recommended Direction

Seguir a **Abordagem 1** para a primeira entrega.

Ela é a opção mais pragmática porque respeita a arquitetura existente, entrega a tela inicial rapidamente e não bloqueia uma evolução posterior para uma camada frontend mais sofisticada. A Landing de seleção de épico é um bom corte vertical para isso: ela exige layout, leitura de dados agregados e uma mutação simples, mas importante, de criação de épico.

## Screen Definition

### Purpose

A tela responde a uma única pergunta operacional: **"em qual épico eu quero trabalhar agora?"**

Ela também precisa acomodar o caso em que o épico ainda não existe, oferecendo criação imediata sem obrigar o usuário a alternar para o terminal.

### Layout

- Título principal `Select an Epic`
- Subtítulo explicando que a seleção define o escopo das demais telas
- Ação primária `Create Epic`
- Grid responsivo de cards de épico
- Estado vazio com chamada clara para criação do primeiro épico

Cada card deve mostrar:

- nome do épico
- badge de fase (`Discover`, `Plan`, `Execute`, `Validate`)
- barra de progresso de tarefas
- linha de métricas com stories, tarefas e percentual concluído
- badge de status agregado
- ação `Open`

### Sidebar Behavior

- O dropdown de épico mostra `Select Epic...` enquanto nada foi escolhido.
- Itens de navegação dependentes de contexto permanecem desabilitados até a escolha.
- `Settings` continua acessível sem contexto.

## Design System Application

### Typography

- `display-lg` para o título da página
- `body-md` para o subtítulo e microcopy
- `heading-md` para o nome do épico no card
- `label-sm` para labels de métricas e badges
- `mono-sm` para percentuais, timestamps e counters operacionais

### Surfaces and Spacing

- Canvas escuro em tom grafite, sem preto puro
- Cards como superfície principal, com borda fina e espaçamento interno generoso
- Escala de spacing restrita a `8 / 12 / 16 / 24 / 32 / 40`
- Ícones outline apenas onde realmente ajudarem a leitura

### Semantic Color

- Azul para andamento
- Verde para concluído
- Âmbar para atenção ou pendência
- Vermelho apenas para falha real

## Data Contract for the Screen

O contrato atual de `GET /api/epics` é insuficiente para a Landing porque hoje ele retorna apenas os campos crus do épico. A tela precisa de um payload agregado por card, com pelo menos:

- `name`
- `description`
- `approval_status`
- `current_phase`
- `story_count`
- `task_count`
- `completed_task_count`
- `completion_ratio`
- `overall_status`

### Phase Derivation

A fase exibida no card deve ser derivada do estado já persistido no projeto. A heurística inicial recomendada:

- sessão ativa `validate` -> `Validate`
- sessão ativa `execute` -> `Execute`
- sessão ativa `plan` -> `Plan`
- sessão ativa `story` ou `epic` -> `Discover`
- sem sessão ativa -> fallback para `Discover`

Isso entrega uma regra objetiva sem inventar novo estado só para a UI.

## Create Epic Flow

Ao criar um novo épico pela UI:

1. a UI envia um `POST /api/epics`
2. o servidor chama uma operação interna compartilhada com o CLI
3. essa operação garante:
   - criação do registro do épico no SQLite
   - criação da pasta `.oraculo/epics/<epic-name>`
4. a resposta retorna o payload do épico criado
5. a UI atualiza a grade e pode opcionalmente selecionar o novo épico

### Explicit Rejection

Não usar a UI para escrever direto no filesystem nem no SQLite. A mutation precisa passar por uma operação interna que também possa ser reaproveitada pelo CLI.

## Error and Empty States

- nome inválido ou vazio -> erro inline no modal/formulário
- épico duplicado -> feedback claro sem quebrar a navegação
- erro de criação de diretório -> erro bloqueante e nenhuma seleção automática
- lista vazia -> estado vazio com CTA principal para criar o primeiro épico

## Deliverable

Produzir uma primeira implementação navegável da tela `Landing (Epic Selection)`, servida pelo binário Go, com:

- leitura de épicos agregados
- criação de épico via trust layer
- aplicação fiel dos foundations do design system
- navegação preparada para escopo por épico nas telas seguintes
