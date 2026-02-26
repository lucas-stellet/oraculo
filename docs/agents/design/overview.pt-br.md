# Agents Design — Visao Geral

## 1. Modos de Operacao

A camada de agents do Oraculo suporta dois modos de operacao, espelhando os modos de nivel de sistema definidos em [`docs/design.md`](../../design.md).

**Product Engineering** (epics): Discover > Plan > Execute > Validate

O modo completo. Uma ideia entra como um enunciado de problema bruto. O orquestrador guia o usuario por uma descoberta Socratica, produz requisitos validados, os decompoe em um DAG de tarefas, despacha agents para executar e valida todo o output por meio de QA independente. A fase Deliver/Merge e trabalho futuro.

**Software Engineering** (stories sem epic): Plan > Execute > Validate

O caminho mais curto. O contexto vem diretamente do usuario — um bug report, uma solicitacao de feature com detalhes suficientes, uma tarefa de refatoracao. A descoberta e ignorada. O orquestrador decompoe o trabalho em um DAG e segue diretamente para execucao e validacao.

Ambos os modos compartilham a mesma infraestrutura de agents, o mesmo pipeline de QA e o mesmo CLI Trust Layer. A unica diferenca e se o Discover e executado.

## 2. Fases da Perspectiva dos Agents

### 2.1 Discover

**Atores:** Orquestrador + Humano (+ Research Agents quando nao existe contexto previo)

O orquestrador faz perguntas, nao code agents ou QA agents. Quando nao existe contexto previo (sem documentacao, sem fontes definidas), research agents sao despachados em paralelo para analisar o codebase e trazer evidencias. Quando existe contexto previo, o orquestrador conduz uma analise leve diretamente. O output e um documento de requisitos validado armazenado como markdown em `.oraculo/epics/<name>/requirements.md`.

### 2.2 Plan

**Atores:** Orquestrador

O orquestrador decompoe os requisitos em um DAG — tarefas como nos, dependencias como arestas. Ele identifica o que e paralelo, o que e sequencial e onde esta o gargalo. Ele atribui skills a cada tarefa com base no trabalho exigido. O CLI valida o DAG (aciclicidade, integridade das dependencias) e o persiste no SQLite.

### 2.3 Execute

**Atores:** Code Agents (com skills carregadas pelo orquestrador)

Cada code agent recebe uma tarefa especifica com contexto focado: arquivos relevantes, padroes arquiteturais, convencoes do CLAUDE.md. Agents trabalham no mesmo branch, no mesmo diretorio — o orquestrador garante que dois agents nunca toquem nos mesmos arquivos simultaneamente. Todas as tarefas de codigo usam TDD por meio da skill TDD.

### 2.4 Validate

**Atores:** QA Agent

A validacao opera em dois niveis:

**Validacao em nivel de tarefa:** Apos cada code agent concluir uma tarefa, o QA agent revisa a implementacao com uma context window limpa. Ele recebe o diff, as specs e os resultados dos testes — nunca o raciocinio do code agent. O QA nunca corrige; ele reporta achados estruturados. Se o QA rejeitar, o orquestrador instancia um novo code agent com o feedback do QA.

**Validacao em nivel de story:** Assim que todas as tarefas de uma story passam no QA individual, o QA agent realiza uma revisao integrada — coerencia entre tarefas, requisitos da story como um todo e qualidade geral. O veredicto da story e o gate final antes da geracao de markdown e conhecimento.

## 3. Ciclo de Vida de uma Tarefa de Ponta a Ponta

Uma tarefa percorre este caminho pelo sistema:

**1. Nascimento.** Durante o Plan, o orquestrador emite um DAG. O CLI valida e o persiste no SQLite. Cada tarefa recebe o status `pending`.

**2. Despacho.** O orquestrador avalia o DAG, identifica tarefas com todas as dependencias satisfeitas e as despacha. Quando possivel, multiplas tarefas rodam em paralelo. Quando tarefas compartilham dependencias de arquivo, rodam de forma sequencial.

**3. Execucao.** Um code agent e instanciado com as skills adequadas e contexto focado. Ele trabalha no mesmo branch, no mesmo diretorio que todos os outros agents. A skill TDD impos o ciclo red-green-refactor. Ao concluir, o agent reporta seus resultados pelo CLI.

**4. Validacao.** O QA agent recebe o diff, as specs e os resultados dos testes em um contexto limpo. Ele verifica corretude funcional, conformidade com os padroes, casos de borda e qualidade dos testes. O CLI registra o veredicto no SQLite.

**5. Resolucao.** Se aprovada, a tarefa e marcada como completa. Se rejeitada, o orquestrador instancia um novo code agent com o feedback do QA — contexto limpo, sem memoria da tentativa anterior. Um circuit breaker limita os ciclos de rejeicao antes de escalar para o humano.

**6. Integracao.** Assim que todas as tarefas de uma story sao validadas, um unico sumario em markdown e gerado e commitado. Os dados operacionais transientes no SQLite cumpriram seu proposito. A fase Deliver/Merge (merge para a linha principal) e trabalho futuro.

## 4. Invariantes Fundamentais

Estes valem em todas as fases e todos os agents:

1. **O CLI valida tudo.** Toda transicao de estado, toda mutacao de DAG, todo veredicto passa pelo CLI. O CLI e deterministico — suas decisoes sao baseadas em dados, nao em raciocinio de LLM.

2. **QA e obrigatorio.** Nenhuma tarefa e completa sem validacao independente de QA. O QA agent sempre opera com uma context window limpa.

3. **Estado append-only.** Tarefas nunca sao deletadas — elas sao marcadas como falhas ou concluidas. Veredictos de QA sao registros imutaveis. O historico completo e preservado no SQLite durante a duracao do epic.

4. **Execucao no mesmo branch.** Todos os agents trabalham no mesmo branch, no mesmo diretorio. O orquestrador coordena o acesso a arquivos por meio das dependencias do DAG. Sem worktrees, sem proliferacao de branches.

5. **Skills definem o comportamento dos agents.** O orquestrador atribui skills a agents com base nas necessidades da tarefa. A skill e o mecanismo de contencao — ela define o workflow, as restricoes e os quality gates do agent.

6. **Artifact markdown unico.** Quando uma story e concluida, o sistema produz um unico arquivo markdown commitado resumindo o que foi feito. Os detalhes operacionais ficam no SQLite (dados transientes, limpaveis apos a conclusao do epic).
