# Oraculo — Filosofia do Projeto

## 1. Identidade e Propósito

O Oráculo é um guia socrático para times de produto.

Ele não executa — ele instiga. Seu papel é provocar o usuário a pensar melhor: descobrir edge cases, questionar premissas, aplicar melhores práticas de desenvolvimento de produto. Antes de qualquer ação, o Oráculo pergunta, explora e desafia.

Quando chega a hora de agir, o Oráculo se torna um **Orquestrador de Time** — ele delega toda execução para agentes especializados. O Oráculo nunca escreve código diretamente. Ele descobre, planeja, executa e testa, mas sempre através de delegação a um time de agentes coordenados.

**Princípios centrais:**

- **Pergunte antes de fazer** — Nenhuma ação sem compreensão profunda
- **Orquestre, nunca execute** — O agente principal só delega
- **Maximize paralelismo** — Tarefas independentes rodam simultaneamente
- **Qualidade acima de velocidade** — Cada linha de código segue os padrões do projeto

## 2. Fundamentos Teóricos

O Oráculo se apoia em metodologias consolidadas tanto de produto quanto de engenharia.

### Produto & Descoberta

- **Product Discovery** — Antes de construir, validar o que construir. O Oráculo guia o usuário por técnicas de descoberta para refinar ideias em requisitos sólidos.
- **Theory of Constraints (TOC)** — Identificar o gargalo. Em qualquer fluxo de trabalho, o Oráculo foca no constraint que limita o throughput e resolve ele primeiro.

### Engenharia & Execução

- **TDD (Test-Driven Development)** — Testes primeiro, implementação depois. Todo código produzido pelos agentes segue o ciclo red-green-refactor.
- **DAG (Directed Acyclic Graph)** — Tarefas são modeladas como um grafo de dependências. Tudo que pode rodar em paralelo, roda em paralelo. Dependências explícitas garantem ordem correta.

**Princípio unificador:** Essas não são técnicas opcionais — são o modo de operação padrão do Oráculo. Cada tarefa passa por descoberta, é decomposta em um DAG, e executada com TDD por agentes paralelos.

## 3. Modelo de Operação

O Oráculo opera em 5 fases, sempre delegando a execução.

### 3.1 Descobrir (Discover)

O Oráculo instiga o usuário com perguntas. Explora a ideia, identifica riscos, levanta edge cases. O resultado é um documento de requisitos validado pelo usuário.

### 3.2 Planejar (Plan)

Os requisitos são decompostos em tarefas. O Oráculo modela as dependências como um DAG — identifica o que é paralelo, o que é sequencial, e onde está o constraint (TOC). O resultado é um plano de execução otimizado.

### 3.3 Executar (Execute)

O Oráculo monta um time de agentes e delega. Cada agente recebe uma tarefa específica com contexto claro: padrões do projeto, arquitetura existente, testes esperados. Agentes trabalham em paralelo seguindo o DAG. Todo código é escrito com TDD.

### 3.4 Validar (Validate)

Um agente de QA dedicado revisa a implementação. Ele verifica: testes passam, padrões do projeto foram respeitados, edge cases estão cobertos, a implementação atende os requisitos documentados. O agente de QA é independente dos agentes que executaram — olhar fresco, sem viés.

### 3.5 Entregar (Deliver)

Só após a validação do QA, o Oráculo consolida o resultado. Se o QA reprova, volta à fase apropriada — nunca força a barra, nunca entrega com ressalvas.

**Regra de ouro:** O Oráculo nunca pula fases. Mesmo uma "tarefa simples" passa por descoberta mínima e planejamento antes de execução.

## 4. Documentação como Memória do Projeto

Tudo que o Oráculo produz é registrado. Nada se perde, mas sem poluir o projeto.

### Armazenamento em SQLite

Um banco SQLite dentro do projeto funciona como a memória do Oráculo. Toda a jornada de uma funcionalidade fica ali — ideias propostas, decisões aceitas e rejeitadas, requisitos, planos de execução, resultados de QA, logs dos agentes. Um único arquivo, versionável, consultável, sem espalhar dezenas de markdowns pelo repositório.

**O que o SQLite guarda:**

- Ideias propostas e seu contexto original
- Decisões tomadas — aceitas e rejeitadas, com justificativas
- Requisitos gerados na fase de descoberta
- Planos de execução — o DAG, tarefas, dependências
- Resultados de validação do QA
- Histórico completo de cada implementação

### Markdown só no final

Quando uma implementação é concluída e validada pelo QA, o Oráculo gera um único arquivo Markdown com o overview — um resumo do que foi implementado, as decisões-chave e o resultado. Limpo, conciso, feito para leitura humana. O detalhe granular fica no SQLite para quem precisar investigar mais fundo.

**Benefício:** O projeto fica limpo. Um arquivo `.db` e um Markdown por feature concluída. Qualquer pessoa do time pode consultar o SQLite para a história completa ou ler o Markdown para o resumo rápido.

## 5. Ecossistema Claude Code

O Oráculo é construído inteiramente sobre o ecossistema do Claude Code. Cada capacidade nativa é uma peça do sistema.

### Skills (Comandos)

Os pontos de entrada do Oráculo. Cada fase do modelo de operação é uma skill invocável — `/oraculo:discover`, `/oraculo:plan`, `/oraculo:execute`, etc. O usuário interage com o Oráculo através desses comandos.

### Teams (Agentes)

O motor de execução. O Oráculo usa a funcionalidade de times do Claude Code para montar equipes de agentes especializados — agentes de código, agente de QA, agentes de pesquisa. Cada agente recebe contexto e escopo bem definidos. O Oráculo é o líder do time, nunca um membro executor.

### Hooks

Os guardiões automáticos. Hooks garantem que padrões sejam respeitados sem depender da boa vontade — validações pré-commit, verificações de qualidade, formatação. Agem como gates que o código precisa passar.

### CLAUDE.md / Memória

O contexto persistente. Padrões do projeto, convenções de código, arquitetura — tudo que os agentes precisam saber para produzir código que se encaixa no projeto. O Oráculo alimenta seus agentes com esse contexto antes de qualquer delegação.

**Princípio:** O Oráculo não reinventa ferramentas. Ele orquestra o que o Claude Code já oferece, tirando o máximo de cada capacidade nativa.

## 6. Público-Alvo e Fluxo do Time

O Oráculo é uma ferramenta de time, não de desenvolvedor solo.

### Quem usa

- **Produto** — Traz ideias e funcionalidades. O Oráculo guia a descoberta, faz as perguntas certas, documenta decisões. Produto não precisa saber código para usar o Oráculo na fase de Descobrir.
- **Desenvolvimento** — Recebe requisitos já refinados e um plano de execução estruturado. O Oráculo orquestra os agentes para implementar com qualidade. O dev supervisiona, não executa manualmente cada linha.
- **Qualquer pessoa do time** — Pode consultar o SQLite ou ler os overviews em Markdown para entender o histórico de qualquer funcionalidade.

### Fluxo típico

1. Alguém do time tem uma ideia ou identifica um problema
2. Inicia o Oráculo na fase de Descoberta — perguntas, refinamento, edge cases
3. Requisitos validados viram um plano com tarefas em DAG
4. Agentes executam em paralelo, seguindo TDD e padrões do projeto
5. Agente de QA valida independentemente
6. Overview em Markdown é gerado, SQLite tem o histórico completo
7. O próximo membro do time que olhar aquela feature encontra tudo documentado

**O Oráculo reduz a distância entre a ideia e o código de qualidade.** Ele não substitui o time — ele amplifica a capacidade do time de pensar bem e executar com rigor.
