# Agents Design — Code Agent

## 1. Um Unico Agent, Skills Carregadas

O Oraculo usa um unico tipo de code agent. Nao existe um agent separado para escrever testes ou um agent separado para implementar — um unico agent cuida de ambos, guiado pela skill de TDD.

O orquestrador instancia um code agent para cada tarefa e carrega as skills apropriadas com base na natureza da tarefa. A skill define o workflow, as restricoes e os quality gates do agent. O agent segue as instrucoes da skill — ele nao inventa seu proprio processo.

Esse design e mais simples do que pipelines de TDD multi-agent (com test-author e implementer separados), preservando a disciplina central: os testes sao escritos antes da implementacao, e o agent nao pode pular etapas.

## 2. TDD via Skill

A skill de TDD impoe o loop red-green-refactor:

**Red:** Escreva um teste que falha e que captura o requisito. Rode-o. Observe a falha. O teste deve falhar pelo motivo certo — nao por um erro de sintaxe ou import faltando, mas porque a feature ainda nao existe.

**Green:** Escreva a implementacao minima para fazer o teste passar. Rode o teste novamente. Observe o sucesso. Nenhum codigo alem do necessario.

**Refactor:** Limpe a implementacao sem mudar o comportamento. Rode os testes novamente para confirmar que nada quebrou.

As instrucoes da skill sao explicitas sobre o que o agent nao deve fazer:
- Nao escreva a implementacao antes de o teste falhar
- Nao enfraqueça as assertions para fazer os testes passar
- Nao pule a fase de refactor
- Nao modifique as assertions do teste apos escrever a implementacao

Essas restricoes sao impostas pelo workflow da skill — o agent recebe instrucoes passo a passo que tornam pular fases algo nao natural. O CLI pode validar adicionalmente o trace de TDD verificando que as falhas de teste precedem as mudancas de implementacao no historico de commits.

## 3. Contexto que o Agent Recebe

Cada code agent recebe um bundle de contexto focado, montado pelo orquestrador:

- **Descricao da tarefa:** O que implementar, criterios de aceitacao, comportamento esperado
- **Arquivos relevantes:** Apenas os arquivos que o agent precisa ler ou modificar — nao o repositorio inteiro
- **Convencoes do projeto:** Do CLAUDE.md — estilo de codigo, padroes, decisoes arquiteturais
- **Contexto de testes:** Testes existentes na area, padroes de teste usados no projeto
- **Contexto da story/epic:** O documento de requisitos mais amplo para entender a intencao

**Menos e mais.** Agents performam melhor com contexto focado do que com um dump completo do repositorio. O orquestrador cuida desse contexto com base no escopo da tarefa e nos arquivos que ela vai tocar.

## 4. Limites de Escopo

O escopo do agent e definido pela descricao da tarefa e pelas instrucoes da skill — nao por ACLs de filesystem ou sistemas de permissao:

- A descricao da tarefa especifica quais arquivos modificar
- As instrucoes da skill definem o workflow (TDD, frontend-design, etc.)
- O orquestrador garante que dois agents concorrentes nunca toquem nos mesmos arquivos por meio das dependencias do DAG

Como todos os agents trabalham no mesmo branch no mesmo diretorio, o escopo e uma convencao imposta por instrucoes, nao uma fronteira de isolamento fisico. Isso e mais simples e suficiente quando o orquestrador sequencia corretamente as tarefas que compartilham arquivos.

## 5. Em Caso de Rejeicao pelo QA

Quando o QA rejeita uma tarefa, o orquestrador **nao** envia a rejeicao de volta para o mesmo agent. Em vez disso:

1. A sessao do agent original e encerrada
2. Um **novo** code agent e instanciado com:
   - A descricao original da tarefa
   - Os achados estruturados do QA (o que falhou, o que estava errado, o que corrigir)
   - Contexto limpo — sem memoria da tentativa anterior
3. O novo agent segue a mesma skill de TDD do zero

**Por que um novo agent?** Agents que recebem feedback de rejeicao sobre seu proprio trabalho tendem a defender suas decisoes anteriores em vez de reconsiderar a partir dos principios basicos. Um agent novo com os achados do QA trata o feedback como verdade absoluta, nao como um desafio ao seu raciocinio anterior.

Esse padrao continua ate que:
- A tarefa passe no QA, ou
- O circuit breaker seja acionado apos N falhas e o caso seja escalado para o humano

## 6. O que o Code Agent Nao Faz

- **Nao despacha outros agents.** Apenas o orquestrador despacha.
- **Nao escreve no SQLite.** Apenas o CLI escreve o estado operacional.
- **Nao escolhe suas proprias skills.** O orquestrador atribui as skills.
- **Nao se comunica com outros agents.** Cada agent trabalha de forma independente. A coordenacao acontece por meio do orquestrador e do DAG.
- **Nao faz merge de codigo.** A integracao e responsabilidade do orquestrador e do CLI.
