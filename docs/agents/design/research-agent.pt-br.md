# Agents Design — Research Agent

## 1. Papel

O research agent e o investigador. Ele analisa o codebase existente, explora referencias externas e levanta evidencias que informam a descoberta e o planejamento. Ele nao escreve codigo, nao roda testes e nao modifica arquivos — ele observa, analisa e reporta achados estruturados.

Research agents sao despachados quando o orquestrador precisa de contexto baseado em evidencias: padroes do codebase, restricoes arquiteturais, capacidades de bibliotecas ou decisoes passadas que afetam o trabalho atual.

## 2. Quando e Despachado

### Durante o Discover (Epic Phase)

Quando nao existe contexto previo (sem documentacao, sem fontes definidas), o orquestrador despacha research agents em paralelo com o dialogo Socratico. Os resultados sao introduzidos na conversa como perguntas embasadas em evidencias.

Quando existe contexto previo (documentacao existente, fontes definidas), o orquestrador conduz uma analise leve diretamente — nenhum research agent e necessario.

### Durante o Plan

O orquestrador pode despachar research agents para investigar a viabilidade tecnica de abordagens especificas, identificar dependencias ou analisar padroes existentes antes de decompor os requisitos em um DAG.

## 3. Contexto Recebido

Research agents recebem um escopo de investigacao focado:

- **Area alvo:** Diretorios, arquivos ou padroes especificos a investigar
- **Pergunta de investigacao:** O que o orquestrador precisa saber (ex.: "Como o sistema de auth atual trata o expiry de sessao?")
- **Convencoes do projeto:** Do CLAUDE.md — estilo de codigo, padroes, decisoes arquiteturais
- **Contexto da epic/story:** Os requisitos mais amplos para entender a intencao

## 4. O que os Research Agents Produzem

Achados estruturados — nunca opinioes ou recomendacoes:

- **Componentes relacionados:** Implementacoes existentes que se sobrepoe ao trabalho atual
- **Restricoes arquiteturais:** Padroes que a nova feature deve respeitar
- **Edge cases:** Comportamentos no codigo atual que a nova feature deve tratar
- **Decisoes passadas:** Decisoes tecnicas relevantes com o raciocinio original
- **Referencias externas:** Capacidades de bibliotecas, documentacao de API, padroes relevantes

## 5. O que os Research Agents Nao Fazem

- **Nao escrevem codigo** — apenas investigacao
- **Nao modificam arquivos** — acesso somente leitura
- **Nao rodam testes** — observacao, nao execucao
- **Nao fazem recomendacoes** — reportam achados, deixam o orquestrador decidir
- **Nao se comunicam com outros agents** — reportam apenas ao orquestrador
- **Nao escolhem seu proprio escopo** — o orquestrador define a investigacao

## 6. Skills

Research agents podem receber skills do orquestrador para investigacoes especializadas:

- **Padrao:** Exploracao de codebase (glob, grep, read)
- **Research externo:** Web search para documentacao de bibliotecas, referencias de API, analise de padroes

O workflow de investigacao padrao (analise de codebase + achados estruturados) nao exige uma skill especial.
