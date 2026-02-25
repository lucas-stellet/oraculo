# Filosofia de Agentes (DRAFT)

> Sintese consolidada de 6 threads de pesquisa paralelas sobre multi-agent orchestration, execucao, comunicacao, qualidade, memoria e coding agents. Baseada em 7 documentos de pesquisa, ~99 padroes identificados, ~38 lacunas catalogadas.

---

## 1. Axiomas Fundamentais

Estes axiomas emergiram independentemente em multiplas linhas de pesquisa. Sao restricoes de design inegociaveis.

### 1.1 Orchestrate, Never Execute

O agente principal (Oraculo) opera exclusivamente como **Planner**. Ele ingere contexto do projeto, monta equipes, gera o DAG, monitora progresso e coordena validacao. Nunca escreve codigo, roda testes ou toca o filesystem diretamente.

**Por que**: Remover execucao tatica do orquestrador reduz hallucinations e desvio estrategico ao preservar a integridade da context window. Quando erros de sintaxe, logs de compilacao e traces de debug entram no contexto do planner, o raciocinio estrategico degrada de forma mensuravel.

**Evidencia**: OpenCode (separacao plan/build), Cline (modos Plan/Act), MetaGPT (separacao de papeis por SOP), AutoGen (thinkers vs. executors). O padrao Planner-Executor e a arquitetura mais validada entre todos os frameworks pesquisados.

### 1.2 Quality Over Speed (Drum-Buffer-Rope)

O QA Agent e o **Drum** -- a restricao que dita o tempo maximo de todo o sistema. A geracao de codigo deve ser matematicamente subordinada ao throughput de QA.

**Por que**: Quando a producao de codigo ultrapassa a capacidade de verificacao, Work-In-Progress nao validado se acumula, causando "context drift" -- codigo rejeitado retorna para retrabalho mas dependencias e interfaces ja mudaram. Paralelismo ilimitado sem respeitar o gargalo de verificacao leva a colisao catastrofica de contexto e custos exponenciais de retrabalho.

**Mecanismo**:
- **Drum**: Throughput do QA Agent define a velocidade maxima do sistema
- **Buffer**: Fila limitada de tarefas completas aguardando QA (previne inanicao do QA)
- **Rope**: Sinal do QA de volta ao orquestrador para throttle no despacho de tarefas

**Evidencia**: Theory of Constraints (Goldratt) adaptada para agent swarms. TAIS framework. Amdahl's Law for Agents (arXiv:2503.15703) -- velocidade global e ditada pelo caminho sequencial mais longo.

### 1.3 Ask Before Doing (Constraint Satisfaction)

Nenhuma execucao sem entendimento profundo. O framework ACONIC formaliza isso: modelar requisitos como constraint satisfaction problems (3-SAT), usar tree decomposition para particionar o grafo de restricoes. Se o grafo nao puder ser decomposto em Stories de baixo treewidth, o problema ainda nao e compreendido -- mais exploracao Socratica e necessaria.

**Por que**: Decomposicao heuristica falha em escala. Grafos formais de restricoes permitem ao orquestrador minimizar algoritmicamente a carga cognitiva nos agentes de execucao.

**Evidencia**: ACONIC (arXiv:2510.07772), Plan-over-Graph (arXiv:2502.14563).

### 1.4 O Trust Layer e a Realidade Deterministica

O CLI Trust Layer e o arbitro final. Ele nao "acredita" -- ele **verifica**. Compilacao, execucao de testes e assercoes de contrato sao as fronteiras deterministicas que contem o comportamento probabilistico da IA.

**Por que**: O problema recursivo de "who watches the watchmen?" e resolvido pela transicao de validacao puramente probabilistica (LLM julgando LLM) para validacao deterministica (contratos que quebram builds). Um contrato e satisfeito se tanto input quanto output sao validados com sucesso contra especificacoes -- nenhuma opiniao necessaria.

**Evidencia**: Design by Contract (Bertrand Meyer), SymbolicAI `@contract` decorators, resultados empiricos do TDFlow (94.3% com testes ground-truth vs. 69.8% com testes auto-gerados).

---

## 2. Arquitetura de Orquestracao

### 2.1 DAG como Estrutura de Dados de Primeira Classe

O plano de execucao nao e uma lista textual -- e um **grafo computavel** de nos (subtarefas) e arestas (dependencias). O orquestrador avalia continuamente o estado do grafo em O(|V|+|E|), despachando execucoes paralelas para todos os nos com in-degree zero.

**Propriedades-chave**:
- Deteccao programatica de ciclos (sem loops infinitos)
- Largura maxima de paralelismo identificada a qualquer momento
- Topological ordering impoe respeito a dependencias
- Nos condicionais de branch permitem mutacao em runtime sem regeneracao total do DAG

**Mutacao dinamica**: DAGs estaticos inevitavelmente falham porque o caminho otimo de execucao depende de observacoes dinamicas (API depreciada, schema legado inesperado). O Routine Framework embutiu logica de branching formal diretamente na representacao do plano. Quando um agente atinge um no condicional, o orquestrador avalia o novo estado, poda o branch invalido e agenda dinamicamente os nos recem-descobertos.

### 2.2 Roteamento Dinamico de Topologia

A Performance Convergence Scaling Law (2026) revela que quando foundation models atingem paridade de raciocinio, a variavel dominante de otimizacao se desloca da selecao de modelo para a **composicao estrutural** -- como agentes sao coordenados importa mais que a inteligencia individual do LLM.

**Quatro topologias canonicas**:

| Topologia | Quando Usar | Fase do Oraculo |
|-----------|-------------|-----------------|
| Parallel | Alta largura, baixa profundidade, features desacopladas | Execute (stories independentes) |
| Sequential | Baixa largura, alta profundidade, dependencias estritas | Validate (review chain) |
| Hierarchical | Alto acoplamento, delegacao manager-worker | Discover (exploracao Socratica) |
| Hybrid | Largura/profundidade mistas entre fases | Ciclo de vida completo |

O orquestrador analisa propriedades matematicas do DAG (largura, profundidade, acoplamento) e roteia execucao dinamicamente. Arquiteturas de workflow estaticas limitam artificialmente as capacidades dos LLMs.

**Evidencia**: AdaptOrch Topology Router (arXiv:2602.16873).

### 2.3 Re-planejamento Concorrente (PIE)

O orquestrador re-planeja estagios posteriores do DAG **concorrentemente** enquanto estagios atuais ainda estao executando. Quando um agente encontra uma restricao inesperada, o orquestrador absorve o feedback dinamico e re-otimiza o DAG downstream em background -- sem pausar a execucao corrente.

**Evidencia**: PIE Framework (AAAI 2025).

### 2.4 Limites de Tamanho de Squad

Evidencia empirica indica que squads especializados de **4-5 agentes** entregam codigo de producao efetivamente. Swarms maiores geram merge conflicts excessivos. O overhead de coordenacao cresce de forma nao-linear com a quantidade de agentes.

---

## 3. Decomposicao de Tarefas e Execucao Paralela

### 3.1 Plan-over-Graph

Em vez de gerar planos textuais sequenciais, o LLM constroi diretamente um grafo formal de nos e arestas. Dynamic programming computa caminhos de execucao paralela otimos. Isso expoe oportunidades maximas de paralelizacao naturalmente.

### 3.2 Decomposicao Induzida por Constraints

Transformar Epics em grafos de restricoes. Minimizar treewidth para garantir matematicamente que as Stories resultantes nao contem requisitos conflitantes. Alto treewidth = problema ainda nao compreendido = mais exploracao Socratica necessaria.

### 3.3 Critical Path Analysis

Amdahl's Law for Agents: velocidade global e limitada pela cadeia sequencial de dependencias mais longa. O orquestrador deve:
- Identificar o critical path no DAG
- Alocar modelos de maior performance aos agentes no critical path
- Usar modelos menores e mais rapidos para tarefas paralelas nao-criticas

### 3.4 Recuperacao de Falhas

**Circuit Breakers**: Monitoram taxas de falha e latencia em cada handoff de agente. Se um Code Agent falha consistentemente na validacao da CLI, o circuit breaker dispara -- bloqueando processamento adicional naquele branch do DAG. Um Debugger Agent e deployado para avaliar o snapshot.

**Coordinated Checkpointing**: Quando um branch paralelo falha, TODOS os agentes dependentes dentro do sub-grafo revertem para um estado temporal sincronizado. Rollbacks nao coordenados em ambientes paralelos criam divergencia de timeline -- agentes colaborando com suposicoes de estado fundamentalmente incompativeis.

**Graceful Degradation**: Se uma ferramenta secundaria falha, o agente deve documentar a limitacao em vez de crashar o no.

---

## 4. Comunicacao e Contexto

### 4.1 O Padrao Blackboard (SQLite + CLI)

O banco SQLite + CLI funciona como um **blackboard deterministico**. Agentes publicam propostas e evidencias; o lider/QA promove o que se torna "conhecimento canonico". O DAG serve como controller, definindo quais escritas sao permitidas em cada fase.

**Governanca e tao importante quanto armazenamento**: O desafio nao e compartilhar dados -- e governar **quem pode escrever o que e quando**, com politicas explicitas.

**Papeis no blackboard** (da pesquisa bMAS/LbMAS):
- **Planner**: Decompoe e agenda
- **Critic**: Revisa qualidade
- **Conflict-Resolver**: Media contradicoes entre agentes
- **Cleaner**: Mantem integridade e relevancia do conteudo compartilhado
- **Decider**: Determina quando consenso suficiente foi alcancado

### 4.2 Injecao Seletiva de Contexto (Minimal Viable Information)

Contexto deve ser tratado como **dependencias modulares e lazy-loaded**, nao como preambulos monoliticos de prompt. Menos informacao, porem altamente direcionada, produz output significativamente mais deterministico.

**Metas praticas**:
- Concepts < 100 linhas
- Guides < 150 linhas
- Reducao de ~80% em tokens alcancavel (de 8.000 para ~750 tokens por requisicao)

**Mecanismos**:
- CLI fornece comando deterministico "build working set"
- FTS5/BM25 recupera apenas contexto relevante por no do DAG
- Mapeamento topologico de repositorio (estilo Aider): classes, assinaturas de metodos, dependencias -- sem carregar corpos de implementacao
- Graph ranking expoe apenas os nos mais referenciados

### 4.3 Protocolo de Handoff

Handoffs sao **transicoes inspecionaveis de primeira classe**, nao acoes implicitas de "perguntar a outro agente" escondidas em prompts.

**Informacao preservada durante handoff**:
1. Descricao da tarefa (estruturada, nao conversacional)
2. Estado atual (referencia de checkpoint)
3. Outputs esperados (validados por schema)
4. Constraints e preconditions
5. Resultados anteriores (referencias, nao conteudo completo)

**Standard Operating Procedures (SOPs)** codificam formatos de handoff como templates legiveis por maquina. Um no do DAG so pode ser marcado como completo quando o agente produz um artefato de output estruturado que o QA Agent pode deterministicamente parsear e validar.

### 4.4 Resolucao de Divergencia de Contexto

Quando agentes paralelos desenvolvem entendimento contraditorio:

1. **Defesa primaria**: SQLite + CLI como single source of truth -- agentes paralelos sempre validam contra o estado canonico do blackboard
2. **Git Worktrees**: Isolamento fisico previne contaminacao durante execucao paralela
3. **QA Agent como arbitro**: Resolve divergencias com evidencia empirica
4. **Truth Maintenance**: Armazenar conclusoes conflitantes como propostas versionadas com justificativas -- nao forcar consenso prematuro. Promover a "verdade operacional" somente apos validacao QA

---

## 5. Garantia de Qualidade e Validacao

### 5.1 Verificacao Ortogonal

O QA Agent opera com uma **context window completamente limpa**, sem memoria do processo original de geracao. Isso quebra o ciclo destrutivo de sycophancy onde um agente tende a concordar com seus proprios erros.

**Requisito arquitetural**: Nenhuma aresta no DAG conecta Execute diretamente a Deliver sem passar por um no de Validate.

### 5.2 TDD como Mecanismo Anti-Hallucination

TDD nao e meramente uma estrategia de testes -- e um **mecanismo critico para ancorar um modelo probabilistico a realidade deterministica**.

**O loop Red-Green-Refactor forcado pela CLI**:
1. Agente escreve teste a partir dos requisitos Socraticos (Red)
2. Agente observa teste falhando contra dependencias reais (prova de validacao real)
3. Agente escreve implementacao minima para satisfazer o contrato (Green)
4. Agente refatora mantendo testes verdes (Refactor)

**A CLI deve rejeitar qualquer commit se o trace de execucao no SQLite nao demonstrar comprovadamente um estado de teste falhando imediatamente antes da fase de implementacao.**

### 5.3 Separacao de Agentes: Test-Author vs. Implementer

O framework TDFlow demonstrou empiricamente:
- Com testes ground-truth humanos: **94.3%** de taxa de sucesso
- Com testes auto-gerados: **69.8%**
- Sem testes: **30-45%**

**"Test hacking"** e uma patologia documentada: quando o agente nao consegue fazer os testes passarem, ele silenciosamente enfraquece as assercoes dos testes. A solucao arquitetural:

1. Um agente especializado escreve testes a partir das especificacoes Epic/Story
2. Um agente separado e isolado escreve a implementacao
3. **O agente de implementacao NAO tem permissao de escrita em arquivos de teste**
4. O Trust Layer (CLI) impoe essa fronteira de permissao

### 5.4 Mutation Testing (Validando o Validador)

Um LLM pode escrever `assert True` -- um teste que passa mas nao cobre nada. Mutation testing introduz bugs deliberados no codigo e verifica se os testes do QA Agent os detectam.

- Se testes ainda passam apesar da mutacao: testes sao frageis, output do QA e rejeitado
- Meta: **>80% de taxa de deteccao de mutacoes**
- A CLI pode automatizar isso como parte do Trust Layer

### 5.5 Validacao Adversarial

Agentes QA cooperativos sao insuficientes para seguranca robusta. Dois tipos de agentes de validacao:

| Tipo | Papel | O Que Detecta |
|------|-------|---------------|
| **Critic Agent** | Avalia criterios nao-funcionais | Eficiencia algoritmica, aderencia a padroes, manutencao, seguranca |
| **Adversarial Agent** | Tenta ativamente quebrar o sistema | Fuzzing de APIs, inputs maliciosos, prompt injection, falhas silenciosas, edge cases |

### 5.6 Cross-Validation com Modelos Heterogeneos

Usar arquiteturas de LLM diferentes para papeis diferentes no ciclo de validacao. Se o Code Agent usa o Modelo X e o QA Agent usa o Modelo Y, vieses arquiteturais compartilhados nao cascateiam pela review chain.

**Evidencia**: Recursive Knowledge Synthesis (arXiv:2601.08839) -- tri-agent com modelos heterogeneos.

### 5.7 Metricas de Qualidade

| Metrica | O Que Mede | Meta |
|---------|-----------|------|
| Pass@1 / Pass@k | Corretude funcional | >85% |
| Mutation Score | Robustez dos testes | >80% |
| Cyclomatic Complexity | Manutencao | <10 caminhos/metodo |
| Execution Efficacy | Fracao de chamadas CLI retornando estado desejado | Alta |
| Convergence Score | Eficiencia de passos ate a solucao | Baixo e melhor |
| Tool Error Rate | Amplificacao de erros pelo DAG | Baixa |

**Qualidade de processo importa**: Um agente que precisa de 50 tool calls para resolver um erro de sintaxe e fundamentalmente menos confiavel do que um que resolve em 3, mesmo que ambos passem na suite de testes.

### 5.8 Human-on-the-Loop (Regra 40-60)

IA lida autonomamente com 40-60% da carga de revisao (sintaxe, linting, formatacao, checks basicos de seguranca). Humanos reservam carga cognitiva para:
- Logica de negocios complexa
- Desvio arquitetural
- Falhas logicas especificas de IA (sorts que falham silenciosamente em edge cases, convencoes inconsistentes, error handling alucinado)

---

## 6. Memoria Compartilhada e Conhecimento

### 6.1 Arquitetura de Memoria em Tres Camadas

| Camada | O Que Contem | Armazenamento | Mutabilidade |
|--------|-------------|---------------|--------------|
| **Working Memory** | Snapshot minimo para tarefa atual: problema, objetivos, restricoes, decisoes recentes, top-k itens recuperados | Construida pelo comando CLI "build working set" | Efemera por no do DAG |
| **Episodic Memory** | Eventos imutaveis vinculados a `run_id/thread_id` e fase (Discover/Plan/Execute/Validate/Deliver) | SQLite append-only | Imutavel |
| **Semantic Memory** | Padroes validados, convencoes, constraints com categoria/dominio, versionamento, status | SQLite com FTS5 | Versionada (proposed/validated/deprecated) |

### 6.2 Event Sourcing como Fundacao

Armazenar mudancas de estado como uma **sequencia imutavel de eventos**. Nunca perder o passado. Derivar "estado atual" como uma projecao validada.

**Implicacoes praticas**:
- Toda "promocao de sabedoria" deve apontar para um conjunto de eventos fonte (proveniencia)
- Capacidade completa de replay, debug e auditoria
- Separar eventos imutaveis (o que aconteceu) de projecoes (padroes consolidados, constraints atuais)

### 6.3 Concorrencia: N Readers + 1 Writer

O modo WAL do SQLite suporta muitos leitores concorrentes mas apenas um escritor por vez. O padrao de paralelismo saudavel:

- **Muitos agentes lendo em paralelo** (recuperacao de contexto, checks de validacao)
- **Escritas serializadas pela CLI** (transacoes curtas, timeouts/retries para SQLITE_BUSY)
- A CLI concentra commits, impoe checkpoints
- Escrita se torna um recurso compartilhado ("no de serializacao" no DAG) sem matar o paralelismo geral

### 6.4 Versionamento Explicito Contra Corrupcao Semantica

Itens de conhecimento canonico devem ser tratados como **versionados**:

```
fields: valid_from, valid_to, supersedes_id, derived_from_episode_id, status
status: proposed | validated | deprecated
```

Versionamento transforma conflitos em dados diagnosticaveis em vez de perdas silenciosas. Quando agentes discordam, o conflito se torna coexistencia de versoes -- nao destruicao de evidencia.

### 6.5 Prevencao de Corrupcao: Contratos na Borda

Design by Contract aplicado a memoria significa que toda operacao de memoria da CLI tem:
- **Preconditions**: Schema obrigatorio, checks de consistencia
- **Postconditions**: Validacao cruzada com indices/FTS5
- **Invariants**: Regras de promocao (apenas QA promove a verdade operacional)

Nao forcar consenso prematuro. Armazenar justificativas e tratar contradicoes como dados ate a validacao final.

### 6.6 Pipeline de Sabedoria

Sabedoria acumulada e um **pipeline de curadoria**, nao acumulacao bruta:

```
Episode -> Score (impacto, risco, recorrencia) -> Reflection/Consolidation -> Validated Promotion -> Relational Linking
```

Sem esse pipeline, a base de conhecimento cresce em volume mas nao em utilidade.

### 6.7 Estrategia de Busca

Comecar com **FTS5 (lexical, deterministico, auditavel)** e evoluir progressivamente:

| Camada | Tecnologia | Quando |
|--------|-----------|--------|
| Busca lexical | FTS5/BM25 | Default, sempre disponivel |
| Estrutura relacional | Tabelas de entidade/relacao em SQLite | Quando conhecimento tem dependencias, conflitos, substituicoes |
| Busca semantica | Embeddings (sqlite-vec) | Quando descoberta conceitual alem de keywords e necessaria |
| Ranking hibrido | RRF (Reciprocal Rank Fusion) | Combinando resultados lexicais + semanticos |

Para memoria canonica, busca lexical com ranking explicavel e checks de integridade pode ser mais confiavel do que depender somente de embeddings.

---

## 7. Coding Agents

### 7.1 Agent-Computer Interface (ACI)

Agentes se comunicam com o ambiente atraves de uma **interface estruturada, deterministica e otimizada para LLMs** (o CLI Trust Layer). Esta e a manifestacao tecnica do conceito de ACI do SWE-agent.

**Principios**:
- Output restrito (ex.: limites de 100 linhas no file viewer)
- Syntax checker embutido forca correcao antes de prosseguir
- Feedback formatado conciso o suficiente para consumo do LLM
- Comandos tipados com inputs/outputs validados

Limitar a liberdade operacional do agente atraves de interfaces estruturadas melhora taxas de sucesso mais do que escalar parametros do modelo.

### 7.2 Git Worktrees para Isolamento Paralelo

Cada no do DAG recebe um **worktree dedicado e isolado** -- uma copia fisica do codebase em um branch separado.

**Por que nao Jujutsu (jj)?**: Seu modelo de working-copy-is-always-a-commit faz com que sub-agentes concorrentes absorvam mudancas uns dos outros em commits massivos e corrompidos. Agentes requerem as fronteiras explicitas de selecao de arquivo (como `git add`) que humanos acham tediosas.

**Insight-chave**: Sistemas projetados para eliminar friccao para desenvolvedores humanos (staging automatico, absorcao implicita de commits) frequentemente criam friccao catastrofica para agentes autonomos. Agentes requerem fronteiras explicitas e deterministicas.

### 7.3 Harness Engineering

O orquestrador implementa "harness engineering" com fronteiras de protocolo explicitas entre tomada de decisao do agente e execucao de ferramentas. Um protocolo bidirecional desacopla o harness do loop do agente.

### 7.4 Context Engineering > Prompt Engineering

Project brain files (CLAUDE.md, SKILL.md, templates de referencia) somados a memoria persistente em SQLite com indexacao FTS5 constituem o "project brain" que sobrevive a interacoes individuais. Contexto dirigido por configuracao supera abordagens baseadas somente em inferencia para manter consistencia de projeto.

---

## 8. Resiliencia e Recuperacao

### 8.1 Durable Execution (Temporal Pattern)

A logica de orquestracao deve ser **estritamente deterministica**. Comportamento nao-deterministico (chamadas LLM, execucoes de ferramentas) e isolado em "Activities". O Event History e completo, imutavel e append-only.

Se um agente crasha, o orquestrador reconstroi o estado exato pre-crash via replay do Event History e retoma execucao sem divergencia de caminho.

**A funcao primaria do orquestrador e preservacao de estado, nao inteligencia.**

### 8.2 Coordinated Checkpointing

Cada fase do Oraculo (Discover/Plan/Execute/Validate/Deliver) se materializa como um **checkpoint reprodutivel** em SQLite. Com FTS5, checkpoints sao pesquisaveis (por decisao, risco, evidencia).

Em caso de falha, rollback e parcial -- ate o ultimo checkpoint valido do sub-grafo -- sem destruir trabalho de branches paralelos nao afetados.

### 8.3 Circuit Breakers

Monitoram taxas de falha e latencia de processamento em cada handoff de agente. Quando um Code Agent falha consistentemente na validacao do Trust Layer:

1. Circuit breaker dispara -- bloqueia processamento adicional naquele branch do DAG
2. Dados corrompidos sao impedidos de fluir downstream
3. Estado e logado no SQLite
4. Debugger Agent e deployado para avaliar o snapshot
5. Sequencia de recuperacao inicia com prompt refinado

---

## 9. Lacunas Identificadas (Transversais)

Areas de pesquisa que requerem investigacao adicional antes de finalizar esta filosofia:

### 9.1 Economia
- Nenhuma analise sistematica de custos operacionais (custo por DAG executado, custo por agent-hour)
- Anthropic reporta 15x de consumo de tokens para sub-agentes paralelos (com 90.2% de ganho de performance), mas analise de ROI esta ausente
- Pontos de break-even vs. desenvolvedores humanos nao quantificados

### 9.2 Falha do Orquestrador
- Toda pesquisa assume que o orquestrador central e confiavel
- Nenhuma cobertura sobre o que acontece quando o Oraculo entra em estado inconsistente, exaure contexto ou alucina durante o planejamento do DAG
- Estrategias de fallback para falha do orquestrador nao documentadas

### 9.3 Decaimento de Conhecimento
- Nenhum mecanismo concreto para detectar quando conhecimento canonico se torna obsoleto
- Necessidade de politicas de "semantic TTL" ou triggers de re-validacao
- Quando um padrao validado se torna deprecated? (migracao de framework, refatoracao grande)

### 9.4 Deteccao de Divergencia em Tempo Real
- Pesquisa descreve como RESOLVER divergencias mas nao como DETECTA-LAS em tempo real durante execucao paralela
- Nenhum mecanismo para monitoramento continuo de coerencia semantica entre agentes paralelos

### 9.5 Propagacao de Contexto Cross-Phase
- Bem coberto dentro de uma fase, mas como contexto acumulado de Discover propaga otimamente para Plan, e de Plan para Execute, especialmente em intervalos longos de tempo?

### 9.6 Dados Empiricos de Producao
- Maioria da evidencia vem de benchmarks academicos (SWE-bench) e frameworks conceituais
- Falta de case studies detalhados de deployments reais em producao com metricas de throughput, qualidade, custo e satisfacao da equipe

### 9.7 Seguranca
- Sandboxing de agentes e seguranca inter-agente (prevenir que um agente comprometido afete outros) tem cobertura minima
- Controle de acesso a memoria (quais agentes podem ler/escrever quais dominios) nao e abordado

---

## 10. Resumo: Decisoes de Design para o Oraculo

| Decisao | Escolha | Racional |
|---------|---------|----------|
| Papel do orquestrador | Somente Planner, nunca executa | Preserva context window, reduz hallucination |
| Modelo de execucao | DAG com roteamento dinamico de topologia | Maximiza paralelismo respeitando dependencias |
| Isolamento paralelo | Git worktrees por no do DAG | Isolamento fisico previne contaminacao cruzada |
| Controle de throughput | Drum-Buffer-Rope (QA como constraint) | Qualidade matematicamente priorizada sobre velocidade |
| Substrato de comunicacao | SQLite blackboard + CLI gateway | Deterministico, auditavel, governavel |
| Entrega de contexto | Minimal Viable Information via CLI | ~80% de reducao de tokens, output mais deterministico |
| Protocolo de handoff | Contratos JSON estruturados validados pela CLI | Legivel por maquina, inspecionavel, testavel |
| Arquitetura de QA | Agente independente, contexto limpo, sem memoria compartilhada com gerador | Quebra ciclo de sycophancy |
| Imposicao de TDD | Agentes separados de test-author e implementer | Previne test hacking (94.3% vs 69.8%) |
| Trust Layer | Design by Contract via CLI (preconditions, postconditions, invariants) | Resposta deterministica a "who watches the watchmen" |
| Profundidade de validacao | Critic + Adversarial + Mutation Testing | QA cooperativo insuficiente para seguranca robusta |
| Diversidade de modelos | Modelos heterogeneos entre papeis | Previne cascata de vieses compartilhados de treinamento |
| Arquitetura de memoria | 3 camadas (working/episodic/semantic) em SQLite | Event sourcing + conhecimento canonico versionado |
| Busca | FTS5 lexical primeiro, evoluir para hibrido | Deterministico, auditavel, verificavel por integrity-check |
| Recuperacao de falhas | Durable Execution + Coordinated Checkpointing + Circuit Breakers | Rollback parcial sem destruir trabalho paralelo |
| Tamanho de squad | 4-5 agentes por squad | Overhead de coordenacao nao-linear alem disso |
| Re-planejamento | Concorrente (PIE) enquanto executa | Sem pausa de execucao para re-otimizacao |
| Papel humano | On-the-Loop (regra 40-60) | Auditar raciocinio e arquitetura, nao sintaxe |
