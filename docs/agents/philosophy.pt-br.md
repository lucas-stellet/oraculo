# Oraculo Agents — Filosofia

## 1. Proposito

A camada de agentes e a forca de execucao do Oraculo.

O Oraculo opera atraves de um time de agentes especializados — cada um com papel claro, fronteiras estritas e nenhuma autoridade alem do seu escopo. O orquestrador planeja e delega. Code agents implementam. QA agents validam. Nenhum agente faz o trabalho de outro.

Este documento define as crencas que governam como agentes trabalham juntos. E a fundacao para todas as decisoes de design sobre papeis de agentes, ferramentas, comunicacao e coordenacao.

## 2. Crenca Central

**Agentes sao raciocinadores poderosos mas operadores nao confiaveis. Todo agente deve ser contido por fronteiras deterministicas.**

Um agente sem restricoes vai alucinar caminhos, enfraquecer suas proprias assercoes de teste, desviar de convencoes do projeto e produzir output quebrado com confianca. O sistema nao previne isso esperando prompts melhores — previne por design. O CLI Trust Layer impoe contratos. Skills impoem disciplina (TDD, padroes de codigo). O DAG impoe ordenacao. Execucao no mesmo branch com contencao baseada em skills substitui a necessidade de isolamento fisico — agentes recebem instrucoes precisas sobre quais arquivos tocar, quais testes rodar e quais convencoes seguir. Todo agente opera dentro de uma jaula de restricoes deterministicas que torna o comportamento correto o unico comportamento possivel.

## 3. O Orquestrador

O orquestrador e o unico agente que ve o quadro completo. Ele recebe contexto do projeto, decompoe trabalho em um DAG, despacha tarefas, atribui skills aos agentes baseado nas necessidades da tarefa e coordena validacao. Nunca escreve codigo, roda testes ou toca o filesystem.

**A context window do orquestrador e sagrada.** Erros de sintaxe, logs de compilacao e traces de debug destroem raciocinio estrategico. O orquestrador planeja e delega — execucao tatica e empurrada inteiramente para agentes subordinados e o CLI Trust Layer. Isso nao e preferencia; misturar planejamento e execucao no mesmo contexto degrada ambos de forma mensuravel.

**O plano de execucao e um grafo computavel, nao uma lista textual.** Tarefas sao nos, dependencias sao arestas. O orquestrador avalia o grafo continuamente, despachando todos os nos desbloqueados quando possivel. Quando um agente encontra um obstaculo inesperado, o orquestrador muta o DAG dinamicamente — podando branches invalidos e agendando novos nos — sem regenerar o plano inteiro.

**Geracao de codigo e subordinada ao throughput de QA.** O orquestrador nao despacha todas as tarefas paralelas so porque o DAG permite. Ele limita a taxa de despacho para que codigo seja produzido apenas tao rapido quanto o QA consegue validar. Work-in-progress nao validado acumula context drift e custos exponenciais de retrabalho. O orquestrador trata a capacidade de QA como a restricao governante do sistema.

**O orquestrador seleciona skills para cada agente.** Baseado na natureza da tarefa, o orquestrador carrega as skills apropriadas — TDD para tarefas de codigo, playwright para validacao E2E, frontend-design para trabalho de UI. A skill e o mecanismo de contencao: define o workflow do agente, restricoes e quality gates.

## 4. Code Agents

Code agents sao as maos do sistema. Recebem uma tarefa focada, um contexto restrito e instrucoes claras. Implementam, nada mais.

**Menos contexto produz codigo melhor.** Code agents recebem a informacao minima viavel para sua tarefa — interfaces relevantes, padroes arquiteturais, contexto de testes — nao o repositorio inteiro. Contexto direcionado reduz custo de tokens e produz output mais deterministico. A CLI constroi esse working set.

**Agentes trabalham no mesmo branch no mesmo diretorio.** Todos os code agents operam no branch de trabalho compartilhado. O orquestrador gerencia coordenacao a nivel de arquivo — garantindo que dois agentes nao modifiquem os mesmos arquivos simultaneamente. Isso elimina a complexidade de worktrees enquanto a estrutura de dependencias do DAG previne conflitos.

**TDD e imposto atraves de skills, nao agentes separados.** Um unico code agent recebe a skill de TDD, que impoe o loop red-green-refactor: escrever um teste falhando, implementar para faze-lo passar, refatorar. As instrucoes da skill impedem o agente de pular etapas ou enfraquecer assercoes. Isso e mais simples que agentes separados de test-author/implementer enquanto preserva a disciplina de TDD.

**TDD e o mecanismo anti-hallucination.** O loop red-green-refactor forca o agente a observar um teste falhando contra dependencias reais antes de escrever implementacao. Isso ancora um modelo probabilistico a realidade deterministica.

## 5. QA Agents

QA agents sao o sistema imunologico. Validam todo output com olhos frescos, sem memoria compartilhada com o agente que o produziu e sem incentivo para concordar.

**Independencia e arquitetural, nao aspiracional.** O QA agent opera com uma context window completamente limpa — sem memoria do processo de geracao, sem acesso ao raciocinio do code agent. Recebe o diff, as specs e resultados de testes. Isso e o que quebra o ciclo de sycophancy onde um agente tende a concordar com seus proprios erros. Nenhuma tarefa pode ser marcada como completa sem passar por QA.

**QA nunca corrige — apenas reporta.** Quando o QA agent encontra problemas, produz um finding estruturado. Nunca modifica codigo. O orquestrador cria um novo code agent com o feedback do QA para resolver os problemas. Essa separacao impede o QA de comprometer seus proprios vereditos.

**O Trust Layer e o arbitro final, nao o QA agent.** O problema recursivo de "who watches the watchmen" nao e resolvido adicionando mais vigias. E resolvido pela realidade deterministica. Compilacao, execucao de testes e assercoes de contrato nao tem opinioes. O QA agent produz findings e evidencias; a CLI os verifica. Se o output viola o contrato, e rejeitado — sem debate.

## 6. Memoria

Memoria e simples e pratica. Serve os agentes sem sobrecarregar o sistema.

**CLAUDE.md e o contexto persistente.** Padroes do projeto, convencoes de codigo, decisoes de arquitetura — tudo que agentes precisam saber vive no CLAUDE.md. E a fonte unica de contexto a nivel de projeto, sempre disponivel, sempre atualizada.

**Markdowns de epic sao a inteligencia acumulada.** Documentos de requisitos, definicoes de story e sumarios de tarefas capturam as decisoes e raciocinio de cada feature. Sao commitados no repositorio e servem como registro historico.

**SQLite efemero rastreia estado de execucao.** Status de tarefas, dependencias, vereditos de QA e dados operacionais vivem em SQLite — por epic, nao commitado, descartavel apos o epic completar. Isso e infraestrutura, nao conhecimento.

**Sem sistema de memoria complexo.** Nao ha arquitetura de tres camadas, nao ha pipeline de curadoria, nao ha knowledge store semantico. O sistema mantem o que precisa na forma mais simples que funciona: markdown para humanos, SQLite para maquinas, CLAUDE.md para agentes.

## 7. O Que Este Documento Nao E

Este documento define crencas, nao implementacao. Nao especifica comandos CLI, schemas de banco de dados, formatos de DAG ou templates de prompt de agentes. Isso pertence ao documento de design.

Este documento tambem nao substitui os documentos de filosofia existentes. Os principios core (docs/philosophy.md) e a filosofia do CLI Trust Layer (docs/cli/philosophy.md) permanecem autoritativos. Este documento os estende para o dominio de coordenacao multi-agente.
