# Oraculo Agents — Filosofia

## 1. Proposito

A camada de agentes e a forca de execucao do Oraculo.

O Oraculo opera atraves de um time de agentes especializados — cada um com papel claro, fronteiras estritas e nenhuma autoridade alem do seu escopo. O orquestrador planeja e delega. Code agents implementam. QA agents validam. Nenhum agente faz o trabalho de outro.

Este documento define as crencas que governam como agentes trabalham juntos. E a fundacao para todas as decisoes de design sobre papeis de agentes, ferramentas, comunicacao e coordenacao.

## 2. Crenca Central

**Agentes sao raciocinadores poderosos mas operadores nao confiaveis. Todo agente deve ser contido por fronteiras deterministicas.**

Um agente sem restricoes vai alucinar caminhos, enfraquecer suas proprias assercoes de teste, desviar de convencoes do projeto e produzir output quebrado com confianca. O sistema nao previne isso esperando prompts melhores — previne por design. O CLI Trust Layer impoe contratos. Git worktrees impoem isolamento. O DAG impoe ordenacao. TDD impoe corretude. Todo agente opera dentro de uma jaula de restricoes deterministicas que torna o comportamento correto o unico comportamento possivel.

## 3. O Orquestrador

O orquestrador e o unico agente que ve o quadro completo. Ele recebe contexto do projeto, decompoe trabalho em um DAG, monta times de agentes, despacha tarefas e coordena validacao. Nunca escreve codigo, roda testes ou toca o filesystem.

**A context window do orquestrador e sagrada.** Erros de sintaxe, logs de compilacao e traces de debug destroem raciocinio estrategico. O orquestrador planeja e delega — execucao tatica e empurrada inteiramente para agentes subordinados e o CLI Trust Layer. Isso nao e preferencia; misturar planejamento e execucao no mesmo contexto degrada ambos de forma mensuravel.

**O plano de execucao e um grafo computavel, nao uma lista textual.** Tarefas sao nos, dependencias sao arestas. O orquestrador avalia o grafo continuamente, despachando todos os nos desbloqueados em paralelo. Quando um agente encontra um obstaculo inesperado, o orquestrador muta o DAG dinamicamente — podando branches invalidos e agendando novos nos — sem regenerar o plano inteiro.

**Geracao de codigo e subordinada ao throughput de QA.** O orquestrador nao despacha todas as tarefas paralelas so porque o DAG permite. Ele limita a taxa de despacho para que codigo seja produzido apenas tao rapido quanto o QA consegue validar. Work-in-progress nao validado acumula context drift e custos exponenciais de retrabalho. O orquestrador trata a capacidade de QA como a restricao governante do sistema.

**Squads sao pequenos.** 4-5 agentes por squad entregam codigo de producao efetivamente. Alem disso, overhead de coordenacao cresce nao-linearmente e merge conflicts dominam. O orquestrador decompoe trabalho para minimizar sobreposicao entre agentes e estabelece fronteiras explicitas de responsabilidade.

## 4. Code Agents

Code agents sao as maos do sistema. Recebem uma tarefa focada, um contexto restrito e um conjunto de testes para fazer passar. Implementam, nada mais.

**Menos contexto produz codigo melhor.** Code agents recebem a informacao minima viavel para sua tarefa — interfaces relevantes, padroes arquiteturais, arquivos de teste — nao o repositorio inteiro. Contexto direcionado reduz custo de tokens e produz output mais deterministico. A CLI constroi esse working set.

**Cada agente trabalha em isolamento fisico.** Todo code agent opera em um git worktree dedicado em um branch separado. Nenhum agente ve o trabalho em progresso de outro agente. Isso elimina contencao de file-locking e previne agentes de alucinar contra codigo meio-escrito de agentes irmaos.

**Code agents nao escrevem seus proprios testes.** Um agente separado escreve testes a partir das especificacoes Epic/Story. O code agent recebe esses testes e escreve implementacao para faze-los passar. O code agent nao tem permissao de escrita em arquivos de teste. Isso previne "test hacking" — a patologia documentada onde um agente silenciosamente enfraquece assercoes em vez de escrever implementacao correta.

**TDD e o mecanismo anti-hallucination.** O loop red-green-refactor forca o agente a observar um teste falhando contra dependencias reais antes de escrever implementacao. Isso ancora um modelo probabilistico a realidade deterministica. A CLI rejeita qualquer commit onde o trace de execucao nao mostre um teste falhando imediatamente antes da fase de implementacao.

## 5. QA Agents

QA agents sao o sistema imunologico. Validam todo output com olhos frescos, sem memoria compartilhada com o agente que o produziu e sem incentivo para concordar.

**Independencia e arquitetural, nao aspiracional.** O QA agent opera com uma context window completamente limpa — sem memoria do processo de geracao, sem acesso ao raciocinio do code agent. Isso e o que quebra o ciclo de sycophancy onde um agente tende a concordar com seus proprios erros. Nenhuma aresta no DAG conecta Execute diretamente a Deliver sem passar por Validate.

**Validacao e adversarial, nao cooperativa.** Um QA agent que meramente confirma "parece bom" e inutil. Validacao efetiva requer agentes com expertise ortogonal — um que verifica corretude funcional, outro que ativamente tenta quebrar o sistema com inputs maliciosos e edge cases. Consenso sem desafio amplifica erros em fatos aceitos.

**O Trust Layer e o arbitro final, nao o QA agent.** O problema recursivo de "who watches the watchmen" nao e resolvido adicionando mais vigias. E resolvido pela realidade deterministica. Compilacao, execucao de testes e assercoes de contrato nao tem opinioes. O QA agent produz testes e evidencias; a CLI os verifica. Se o output viola o contrato, e rejeitado — sem debate.

**Modelos diferentes capturam blind spots diferentes.** Quando o code agent e o QA agent usam a mesma arquitetura de LLM, vieses compartilhados de treinamento cascateiam pela review chain sem serem detectados. Usar modelos heterogeneos entre papeis previne isso.

## 6. Memoria

Memoria e a inteligencia acumulada do sistema. Persiste entre sessoes, cresce por curadoria e e governada pelos mesmos contratos que governam todo o resto.

**Memoria tem tres camadas, nao uma.** Working memory e o snapshot minimo montado para a tarefa atual — efemero, construido sob demanda pela CLI. Episodic memory e o log imutavel de tudo que aconteceu — eventos vinculados a execucoes, fases e agentes, nunca modificados ou deletados. Semantic memory e conhecimento validado — padroes, convencoes, constraints — versionado e promovido por revisao explicita. Cada camada tem regras diferentes. Mistura-las corrompe as tres.

**Nada entra na memoria sem um contrato.** A CLI valida toda escrita. Schema e obrigatorio. Proveniencia e obrigatoria. Contradicoes sao armazenadas como propostas versionadas com justificativas, nao forcadas em consenso prematuro. Apenas conhecimento validado por QA e promovido a verdade operacional.

**Sabedoria e curada, nao acumulada.** Acumulacao bruta produz ruido. O sistema pontua episodios por impacto e recorrencia, consolida-os em reflexoes, valida as reflexoes e vincula-as relacionalmente. Sem esse pipeline de curadoria, a base de conhecimento cresce em volume mas nao em utilidade.

**Escritas sao serializadas, leituras sao paralelas.** Muitos agentes leem memoria concorrentemente. Todas as escritas passam pela CLI como transacoes curtas e serializadas. Isso nao e uma limitacao — e o modelo de governanca. A CLI e o gateway unico que impede agentes de escrever conhecimento incorreto ou contraditorio.

## 7. O Que Este Documento Nao E

Este documento define crencas, nao implementacao. Nao especifica comandos CLI, schemas de banco de dados, formatos de DAG ou templates de prompt de agentes. Isso pertence ao documento de design.

Este documento tambem nao substitui os documentos de filosofia existentes. Os principios core (docs/philosophy.md) e a filosofia do CLI Trust Layer (docs/cli/philosophy.md) permanecem autoritativos. Este documento os estende para o dominio de coordenacao multi-agente.
