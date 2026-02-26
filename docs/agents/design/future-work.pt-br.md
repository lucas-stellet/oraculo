# Agents Design — Trabalho Futuro

Este documento lista capacidades que foram exploradas durante a pesquisa mas adiadas do design atual. Cada item tem uma justificativa clara para o adiamento e um resumo dos achados de pesquisa que vao informar a implementacao futura.

## 1. Fase Deliver/Merge

**O que:** Uma quinta fase no modo de operacao completo — apos o Validate, o sistema faz merge do trabalho validado para a linha principal, resolve conflitos e produz um sumario final.

**Por que adiado:** O sistema atual valida codigo mas depende do humano para integra-lo. O merge automatizado introduz riscos (resolucao de conflitos, interacao com CI pipeline, regras de branch protection) que exigem um design cuidadoso.

**Achados de pesquisa:** O draft de design descreveu integracao serial com validacao pre-merge e um Merge Agent dedicado para resolucao de conflitos. O CLI garantiria que apenas codigo validado chegasse a linha principal.

## 2. Worktree Isolation

**O que:** Cada code agent opera em um Git worktree dedicado em um branch separado, com isolamento fisico do filesystem.

**Por que adiado:** Worktrees adicionam complexidade operacional significativa (gerenciamento do ciclo de vida do worktree, proliferacao de branches, resolucao de conflitos de merge entre agents paralelos). O modelo mais simples de mesmo branch com coordenacao de arquivos baseada em DAG e suficiente para o escopo atual.

**Achados de pesquisa:** O Synthesis 02 explorou o provisionamento de worktrees, a convencao do diretorio `.oraculo/worktrees/`, o ciclo de vida efemero do worktree (criacao no despacho, destruicao na conclusao) e o `base_sha` fixado para ambientes de execucao deterministicos.

## 3. Agents Separados de Test-Author/Implementer

**O que:** Dois tipos distintos de agent para TDD — um escreve testes a partir de especificacoes (sem conhecimento da implementacao), outro escreve a implementacao para fazer esses testes passarem (sem acesso de modificacao dos testes).

**Por que adiado:** O modelo de dois agents oferece garantias mais fortes contra test hacking (o implementer literalmente nao pode modificar os testes), mas dobra o overhead de agents e introduz complexidade de handoff. O modelo de agent unico com TDD skill preserva a disciplina TDD por meio das instrucoes da skill em vez de separacao fisica.

**Achados de pesquisa:** O Synthesis 03 forneceu designs detalhados para isolamento de contexto do test-author (especificacao como unico input, bloqueio de prompts anti-implementacao), o formato de handoff `tdd_bundle.json`, duas camadas de enforcement de permissoes (filesystem ACLs + Claude Code hooks) e o mecanismo de branch "Before", no qual o branch do implementer e derivado do branch de testes.

## 4. QA Adversarial (Executable Proof Pattern)

**O que:** Um QA agent adversarial especializado que ativamente tenta quebrar o codigo por meio de injecao, fuzzing, teste de race conditions e deteccao de semantic drift. Suas afirmacoes sao validadas pelo padrao "Executable Proof" — apenas scripts de teste executaveis que efetivamente induzem falhas sao aceitos como evidencia.

**Por que adiado:** O QA agent atual lida com corretude funcional, padroes e casos de borda. O teste adversarial adiciona custo significativo de tokens e exige mecanismos de escopo (adversarial focus matrix) para evitar exploracao ilimitada.

**Achados de pesquisa:** O Synthesis 04 definiu quatro padroes de prompt adversarial (manipulacao de estado, spoofing de autoridade, protocolo de revisao round-trip, concorrencia/fuzzing). O Executable Proof pattern elimina falsos positivos ao ignorar afirmacoes textuais — apenas falhas de teste executadas pelo CLI constituem achados validos.

## 5. Mutation Testing

**O que:** Mutacao automatizada do codigo de implementacao (ex.: trocar `&&` por `||`, `+` por `-`) seguida de re-execucao do test suite para validar a qualidade dos testes. Testes que ainda passam apos mutacoes sao "mutantes sobreviventes", indicando assertivas fracas.

**Por que adiado:** Mutation testing valida o test suite, nao a implementacao. E um quality gate de segunda ordem poderoso, mas exige integracao de ferramenta por linguagem (Stryker para JS/TS, mutmut para Python, PIT para JVM) e adiciona tempo de execucao significativo.

**Achados de pesquisa:** O Synthesis 03 e 04 recomendaram mutation testing incremental por story (com escopo no diff) com um threshold configuravel (80% para o padrao, 90% para mudancas de alto risco). Relatorios de mutantes sobreviventes seriam roteados de volta ao test-author para fortalecimento das assertivas.

## 6. Selecao Heterogenea de Modelos

**O que:** Atribuicao de diferentes familias de modelos LLM a diferentes roles de agents — ex.: Claude para geracao de codigo, Gemini para revisao funcional, DeepSeek para testes adversariais. Dados de treinamento e abordagens de alinhamento distintos detectam pontos cegos diferentes.

**Por que adiado:** Exige infraestrutura multi-provedor, gerenciamento de custo por role e cadeias de fallback para indisponibilidade de provedor. O sistema atual funciona com uma unica familia de modelos em todos os agents.

**Achados de pesquisa:** O Synthesis 04 citou o estudo X-MAS-Bench confirmando que configuracoes multi-agent heterogeneas superam as homogeneas. A restricao forte recomendada: QA agents devem usar uma familia de modelos diferente dos code agents que validam.

## 7. Sistema de Memoria Rico

**O que:** Uma arquitetura de memoria em tres camadas — working memory (montada sob demanda), episodic memory (log de eventos imutavel) e semantic memory (conhecimento validado). Inclui um pipeline de curadoria (pontuacao de episodios, reflexao por LLM, juri multi-agent para promocao) e ligacao relacional entre entradas de conhecimento.

**Por que adiado:** O modelo simples atual (CLAUDE.md + markdowns + SQLite com operacoes transientes e conhecimento persistente) cobre as necessidades essenciais. O sistema de tres camadas adiciona complexidade significativa: design de schema, triggers de curadoria, thresholds de severidade, geracao de embeddings e evolucao de busca (FTS5 → sqlite-vec → graph queries).

**Achados de pesquisa:** O Synthesis 05 forneceu schemas prontos para producao para as tres camadas, pontuacao composta FTS5 com decaimento temporal, um pipeline de curadoria completo (Episode → Score → Reflection → MAR Consensus → Promotion) e um roadmap de evolucao de busca em tres fases dentro do SQLite puro.

## 8. Dispatch Avancado

**O que:** Algoritmos sofisticados de dispatch incluindo pontuacao de prioridade de caminho critico (`(critical_path_length, fan_out)`), deteccao de gargalo dinamico (medias moveis de tempo de conclusao por tipo de no) e limites WIP armazenados em tabelas de configuracao do SQLite.

**Por que adiado:** O modelo de dispatch atual (o orquestrador avalia o DAG e despacha tarefas prontas) e direto e suficiente. Otimizacoes avancadas de dispatch importam quando se executam muitos agents paralelos com restricoes rigidas de recursos — uma escala que o sistema ainda nao atingiu.

**Achados de pesquisa:** O Synthesis 01 definiu o loop de dispatch baseado em fronteira, a heuristica de prioridade de caminho critico e o mapeamento Drum-Buffer-Rope (QA como Drum, tarefas concluidas aguardando QA como Buffer, throttling de dispatch como Rope). O loop de dispatch rodaria de forma autonoma no CLI sem envolvimento de LLM.

## 9. Event Sourcing

**O que:** Um log de eventos append-only como unica fonte de verdade para todas as mudancas de estado. Toda mudanca de status de tarefa, veredicto de QA e mutacao de DAG e registrada como um evento imutavel. Tabelas materializadas sao projecoes derivadas do stream de eventos.

**Por que adiado:** O schema SQLite atual usa atualizacoes diretas de status (mais simples de implementar e consultar). Event sourcing oferece auditabilidade e recuperacao de falhas superiores, mas adiciona complexidade em design de schema, versionamento de eventos e gerenciamento de snapshots.

**Achados de pesquisa:** O Synthesis 01 e 05 recomendaram event sourcing com materializacao sincrona — atualizando tabelas materializadas na mesma transacao da insercao do evento. O modo WAL (`PRAGMA journal_mode = WAL`) foi recomendado para acesso concorrente de leitura/escrita.

## 10. RBAC e Governanca

**O que:** Controle de acesso baseado em role para o blackboard SQLite — restringindo quais agents podem ler/escrever quais tabelas, aplicado por CLI middleware e triggers SQLite.

**Por que adiado:** O sistema atual depende do orquestrador para atribuir tarefas e contexto adequados aos agents. RBAC formal adiciona enforcement em nivel de banco de dados, mas e desnecessario quando o orquestrador e o unico despachante e o CLI e o unico gateway de escrita.

**Achados de pesquisa:** O Synthesis 06 projetou um modelo de governanca em tres camadas: armazenamento de politicas no SQLite (tabela `security_policies`), enforcement no banco de dados via triggers `BEFORE INSERT/UPDATE/DELETE` com `RAISE(FAIL)` e CLI middleware para isolamento de permissoes de ferramentas por role.
