# Oraculo Agents — Design

## 1. Visao Geral

A camada de agents do Oraculo e a forca de execucao que transforma planos em codigo validado. O orquestrador decompoe os requisitos em um DAG, despacha tarefas para agents especializados e valida todo o resultado por meio de QA independente. Nenhum agent age sem um plano; nenhum codigo e entregue sem revisao.

Dois modos de operacao espelham o design de nivel de sistema:

- **Product Engineering** (epics): Discover > Plan > Execute > Validate
- **Software Engineering** (stories): Plan > Execute > Validate

Todos os agents trabalham no mesmo branch, no mesmo diretorio. O orquestrador coordena o acesso a arquivos por meio das dependencias do DAG. O CLI Trust Layer valida cada transicao de estado. O SQLite rastreia o estado operacional e o conhecimento acumulado; markdowns commitados capturam decisoes e resultados.

Detalhes completos: [`design/overview.md`](design/overview.md)

## 2. O Orquestrador

O orquestrador e o unico agent que enxerga o quadro completo. Ele planeja, delega e coordena — nunca executa. Sua context window e reservada exclusivamente para raciocinio estrategico: erros de sintaxe, logs de compilacao e traces de debug nunca entram em seu contexto.

Durante a fase Plan, o orquestrador decompoe os requisitos em um DAG — tarefas como nos, dependencias como arestas. Ele identifica o que e paralelo, o que e sequencial e onde esta o gargalo. O CLI valida o DAG (aciclicidade, integridade das dependencias) e o persiste no SQLite.

O orquestrador atribui skills a cada agent com base nas necessidades da tarefa. TDD para tarefas de codigo, playwright para validacao E2E, frontend-design para trabalho de UI. A skill e o mecanismo de contencao — ela define o workflow do agent, suas restricoes e os quality gates.

O despacho segue o DAG: todas as tarefas desbloqueadas rodam em paralelo quando possivel, de forma sequencial quando a coordenacao no nivel de arquivo exige. O throughput do QA governa o ritmo — o orquestrador limita o despacho para que codigo seja produzido apenas na velocidade em que o QA consegue validar, evitando um backlog de trabalho nao revisado. Quando uma tarefa entra em `awaiting_approval`, o orquestrador rastreia esse estado e suspende todos os despachos dependentes ate que um verdict humano seja recebido.

Detalhes completos: [`design/orchestrator.md`](design/orchestrator.md)

## 3. Code Agent

Um unico tipo de code agent lida com todas as tarefas de implementacao. Nao ha um autor de testes separado do implementador — um unico agent escreve testes e implementacao, guiado pelo ciclo red-green-refactor da skill TDD.

Cada code agent recebe um contexto focado: a descricao da tarefa, os arquivos relevantes, as convencoes do projeto vindas do CLAUDE.md e os requisitos da story/epic. Menos contexto produz codigo melhor — agents recebem o minimo de informacao necessario, nao o repositorio inteiro.

Todos os code agents trabalham no mesmo branch, no mesmo diretorio. O orquestrador garante que dois agents nunca toquem nos mesmos arquivos simultaneamente por meio das dependencias do DAG. O escopo e aplicado pelas instrucoes da skill e pelas descricoes de tarefa, nao por ACLs de filesystem.

Em caso de rejeicao pelo QA, o orquestrador instancia um **novo** code agent com os achados do QA e um contexto limpo. Agents que recebem feedback de rejeicao sobre seu proprio trabalho tendem a defender decisoes anteriores; um novo agent trata os achados do QA como verdade absoluta.

Detalhes completos: [`design/code-agent.md`](design/code-agent.md)

## 4. QA Agent

O QA agent valida todo o output de codigo com uma context window completamente limpa — sem memoria do raciocinio do code agent, sem acesso ao seu historico conversacional. Ele recebe exatamente o diff, as specs e os resultados dos testes. Esse isolamento quebra o ciclo de sycophancy em que um agent concorda com seus proprios erros.

O QA verifica corretude funcional, conformidade com os padroes, casos de borda, qualidade dos testes e escopo. Ele **nunca corrige codigo** — produz achados estruturados que o orquestrador encaminha para um novo code agent. Essa separacao garante que o QA nunca possa comprometer seus proprios veredictos.

Um circuit breaker limita os ciclos de rejeicao do QA (padrao: 3) antes de o QA agent submeter uma approval request `qa-escalation` ao dashboard. O agent entra em `awaiting_approval` e interrompe o despacho naquela tarefa ate que o humano emita um verdict (`approved`, `rejected` ou `needs_revision`). Todos os achados do QA e resumos de tentativas sao preservados para revisao humana.

Detalhes completos: [`design/qa-agent.md`](design/qa-agent.md)

## 4.5 Research Agent

Research agents investigam o codebase e referencias externas para fornecer evidencias para a tomada de decisao. Eles sao despachados durante o Discover (quando nao existe contexto previo) e durante o Plan (para analise de viabilidade tecnica).

Cada research agent recebe um escopo de investigacao focado: diretorios relevantes, padroes especificos a procurar ou referencias externas a analisar. Eles retornam achados estruturados — nao opinioes. O orquestrador integra os achados ao dialogo ou ao processo de planejamento.

Research agents nao escrevem codigo, nao rodam testes e nao modificam arquivos. Eles observam, analisam e reportam.

Detalhes completos: [`design/research-agent.md`](design/research-agent.md)

## 5. Runtime

### SQLite — Estado Operacional + Conhecimento

Um unico banco de dados SQLite em `.oraculo/oraculo.db` rastreia duas categorias de dados. O estado operacional (tarefas, dependencias, veredictos do QA) e transiente — essencial durante a execucao, limpavel apos a conclusao do epic. A tabela `knowledge` e persistente — acumula licoes aprendidas em todos os epics. O banco de dados usa o schema de [`docs/cli/design.pt-br.md`](../cli/design.pt-br.md) §4.3. Ele esta listado no `.gitignore` e nao e commitado, mas seus dados de conhecimento sao a memoria de longo prazo do projeto.

### Artifact Markdown Unico

Quando uma story e concluida, o sistema gera um unico arquivo markdown commitado resumindo a implementacao: o que foi construido, decisoes-chave, arquivos modificados, resultado do QA. Um arquivo por story — detalhes no nivel de tarefa sao granulares demais para leitura humana.

### Modelo de Memoria

Tres fontes, tres propositos:
- **CLAUDE.md** — Convencoes do projeto, arquitetura, estilo de codigo (persistente, commitado)
- **Epic markdowns** — Requisitos, decisoes e resultados por feature (persistente, commitado)
- **SQLite** — Estado operacional (transiente) + conhecimento acumulado (persistente), nao commitado

Sem arquitetura de memoria de tres camadas, sem pipeline de curacao. Uma tabela de conhecimento simples com busca full-text — nao um sistema de pontuacao de promocao.

Detalhes completos: [`design/runtime.md`](design/runtime.md)

## 6. Trabalho Futuro

A fase de pesquisa produziu achados extensivos em seis topicos. Muitas capacidades foram adiadas em favor de uma arquitetura mais simples. Os itens adiados mais significativos:

- **Fase Deliver/Merge** — Integracao automatizada do codigo validado na linha principal
- **Worktree isolation** — Isolamento fisico de filesystem para execucao verdadeiramente paralela
- **Separacao de autor de testes/implementador** — Garantias TDD mais fortes por meio da separacao de agents
- **QA adversarial** — Testes de exploracao ativa com o padrao Executable Proof
- **Mutation testing** — Validacao automatizada da qualidade dos testes via mutacao de codigo
- **Modelos heterogeneos** — Familias de LLM diferentes para papeis de agents diferentes
- **Sistema de memoria rico** — Arquitetura de tres camadas com pipeline de curacao
- **Despacho avancado** — Pontuacao de caminho critico, deteccao de gargalo variavel
- **Event sourcing** — Log de eventos append-only como fonte unica de verdade
- **Governanca RBAC** — Controle de acesso baseado em papel no blackboard SQLite

Cada item inclui achados de pesquisa que vao informar implementacoes futuras.

Detalhes completos: [`design/future-work.md`](design/future-work.md)
