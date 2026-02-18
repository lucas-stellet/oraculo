# Fase de Brainstorming — Design do Sistema

## 1. Estrutura da Sessão

A fase de Brainstorming segue um fluxo estruturado com dois macro-movimentos: **Divergir** (expandir o espaço do problema) e depois **Convergir** (estreitar para uma definição validada). O Oraculo avança por sete etapas, progredindo apenas quando a condição de saída de cada etapa é satisfeita.

### 1.1 Entrada — Reenquadramento

O usuário chega com uma ideia. O Oraculo separa o problema de qualquer solução proposta antes de prosseguir.

**Perguntas de gatilho:**
- "Você descreveu o que quer construir. Que problema isso resolve para o usuário?"
- "Se essa feature não existisse, o que o usuário faria? Qual é o custo dessa alternativa?"
- "O que está explicitamente fora do escopo desta exploração?"

**Condição de saída:** O usuário articulou uma definição de problema que é independente de qualquer solução específica. O Oraculo registra a ideia crua como originalmente declarada e o problema reenquadrado como artefatos separados.

### 1.2 Divergência — Expandindo o Espaço do Problema

O Oraculo aplica templates de perguntas JTBD e Quatro Forças para explorar dimensões que o usuário não considerou.

**Perguntas JTBD:**
- "Em que situação específica o usuário encontra esse problema?"
- "O que o usuário está tentando realizar — funcionalmente, emocionalmente, socialmente?"
- "Que alternativas existem hoje? Por que são insuficientes?"
- "O que faria o usuário 'contratar' uma nova solução para esse trabalho?"

**Perguntas das Quatro Forças:**
- **Push:** "O que é frustrante o suficiente no estado atual para motivar mudança?"
- **Pull:** "O que torna essa solução mais atraente do que o status quo?"
- **Ansiedade:** "O que pode fazer os usuários hesitarem? Curva de aprendizado? Perda de dados? Dependências?"
- **Hábito:** "O que o usuário faz automaticamente hoje que isso precisaria substituir ou integrar?"

**Perguntas de edge cases:**
- "O que acontece se o usuário fizer [comportamento inesperado]?"
- "Como isso se comporta com estado vazio, dados no limite ou usuários concorrentes?"
- "Quem poderia ser prejudicado por essa mudança mesmo não sendo o usuário-alvo?"

**Condição de saída:** O Oraculo explorou pelo menos as três dimensões JTBD (funcional, emocional, social) e todas as quatro forças. O usuário identificou pelo menos um edge case ou risco que não havia considerado anteriormente.

### 1.3 Análise da Codebase (Paralela)

Enquanto o diálogo avança, o Oraculo dispara sub-agentes para analisar a base de código existente. Isso roda em paralelo — não bloqueia a conversa.

**O que os sub-agentes investigam:**
- Implementações existentes que se sobrepõem à ideia proposta
- Padrões arquiteturais que a nova feature deve respeitar
- Testes que cobrem comportamento relacionado (pontos potenciais de quebra)
- Restrições técnicas que limitam ou habilitam abordagens específicas
- Decisões passadas registradas no SQLite que se relacionam com este domínio

**Como os resultados entram na conversa:**
- "A análise encontrou que [componente X] já trata um caso similar. Você quer construir sobre isso ou criar algo independente? Por quê?"
- "Há um edge case em [área Y] que o código atual trata explicitamente. Como sua ideia se comporta nesse cenário?"
- "Uma decisão anterior rejeitou [abordagem Z] por [motivo]. Esse raciocínio ainda se aplica?"

**Timing:** A análise da codebase é disparada no início da Divergência (etapa 1.2). Resultados são introduzidos no diálogo conforme ficam disponíveis — durante a Divergência se prontos cedo, ou durante a Convergência se a análise demorar mais.

### 1.4 Convergência — Definindo o Problema

O Oraculo muda de expandir para estreitar. Essa transição acontece após a condição de saída da Divergência ser atingida.

**Perguntas de convergência:**
- "De tudo que exploramos, qual é a dor central — a causa raiz, não o sintoma?"
- "Se você tivesse que escrever o problema em uma frase do ponto de vista do usuário, como seria?"
- "Que métrica de negócio esse problema afeta diretamente?"
- "Se você pudesse resolver apenas um aspecto disso, qual entregaria mais valor?"

**Condição de saída:** O usuário produziu uma definição de problema clara (uma frase, do ponto de vista do usuário) e identificou um outcome-alvo (métrica mensurável ou mudança observável de comportamento do usuário).

### 1.5 Mapeamento de Premissas

O Oraculo revela premissas ocultas e as pontua por risco.

**Categorias de premissas:**
- **Hipótese de problema:** "O usuário X sofre com Y na situação Z."
- **Hipótese de solução:** "Se construirmos A, resolveremos Y."
- **Hipótese de valor:** "O usuário perceberá isso como valioso o suficiente para mudar comportamento."
- **Hipótese de viabilidade:** "Isso pode ser construído dentro da arquitetura atual sem refatoração significativa."

**Mecanismo de pontuação de risco:**
Cada premissa é avaliada em dois eixos:
- **Impacto** (1-5): se estiver errada, quão mal quebra a ideia?
- **Evidência** (1-5): quantos dados suportam isso? (1 = intuição pura, 5 = dados validados)
- **Score de risco** = Impacto × (6 - Evidência). Score maior = prioridade maior.

Premissas são registradas no SQLite com status "não validada". O Oraculo declara explicitamente: "Estou registrando isso como premissa não validada. Você tem alguma evidência que já confirme parte disso?"

**Condição de saída:** Todas as premissas com score de risco acima do limiar foram reconhecidas pelo usuário. Cada premissa crítica (alto impacto, baixa evidência) tem evidência de suporte ou está sinalizada como risco conhecido para a fase de Plan.

### 1.6 Stress Test — Working Backwards

O Oraculo roda um PR/FAQ simplificado para testar se a ideia tem clareza suficiente para avançar.

**Perguntas de stress test:**
- "Se essa feature fosse lançada amanhã, como seria o anúncio em duas frases?"
- "Qual seria a primeira pergunta de um cliente cético?"
- "Qual seria a principal objeção da engenharia?"
- "Como você mediria sucesso três meses após o lançamento?"

**Condição de saída:** O usuário consegue responder todas as quatro perguntas sem hesitação significativa ou contradição com respostas anteriores. Se o usuário tiver dificuldade, o Oraculo identifica qual lacuna existe e volta à etapa que a endereça (tipicamente 1.2 ou 1.4).

### 1.7 Portão de Saída

O checkpoint final antes de avançar para o Plan. O Oraculo avalia quatro dimensões de risco:

**Checklist de saída (Quatro Riscos):**
1. **Valor** — O valor para o usuário está claro e a necessidade validada (ou explicitamente marcada como premissa)?
2. **Usabilidade** — A solução é compreensível e utilizável sem atrito excessivo?
3. **Viabilidade técnica** — A análise da codebase confirmou viabilidade (ou sinalizou riscos conhecidos)?
4. **Viabilidade de negócio** — As métricas de sucesso estão definidas e o impacto no negócio é compreendido?

**Comportamento do portão:** O Oraculo avalia cada risco e apresenta a avaliação ao usuário. Se qualquer risco estiver sem resposta, o Oraculo identifica a lacuna específica e redireciona para a etapa apropriada:
- Lacuna de valor → retorna a 1.2 (Divergência)
- Lacuna de usabilidade → retorna a 1.2 (Exploração de edge cases)
- Lacuna de viabilidade técnica → retorna a 1.3 (Análise da Codebase)
- Lacuna de viabilidade de negócio → retorna a 1.4 (Convergência)

Somente quando todos os quatro riscos estão endereçados a sessão produz seus artefatos finais e faz handoff para o Plan.

## 2. Artefatos de Saída

A fase de Brainstorming produz quatro artefatos estruturados que alimentam diretamente a fase de Plan.

### 2.1 Job Stories

Uma ou mais Job Stories no formato:

> "Quando [situação/contexto específico], eu quero [motivação/trabalho a ser feito], para que eu possa [resultado esperado mensurável]."

Job Stories se tornam o input para decomposição de tarefas na fase de Plan. Cada tarefa no DAG deve rastrear de volta a uma Job Story.

### 2.2 Opportunity Solution Tree (OST)

Uma árvore estruturada persistida no SQLite:

```
Outcome (métrica de negócio / comportamento do usuário)
  └── Oportunidade (necessidade do usuário, escrita em primeira pessoa)
        └── Solução (abordagem proposta)
              └── Premissa (hipótese + critério de validação)
```

A árvore cresce entre sessões. Quando o usuário retorna com uma nova ideia, o Oraculo consulta oportunidades existentes: "Em uma sessão anterior, identificamos a oportunidade X ligada ao outcome Y. Essa nova ideia endereça a mesma oportunidade ou uma nova?"

### 2.3 Registro de Premissas

Uma lista priorizada de premissas com:
- Descrição da premissa
- Categoria (problema / solução / valor / viabilidade)
- Score de risco (impacto × falta de evidência)
- Status: não validada / validada / refutada
- Evidência associada (se houver)

Premissas críticas (alto risco) se tornam inputs explícitos para a fase de Plan — podem gerar tarefas de validação antes de qualquer implementação começar.

### 2.4 Resumo de Impacto na Codebase

Resultados da análise dos sub-agentes, registrados como:
- Componentes existentes que se relacionam com a ideia
- Restrições arquiteturais identificadas
- Edge cases do código atual que a nova feature deve tratar
- Decisões passadas do SQLite que são relevantes

## 3. Escalando com a Complexidade

O Oraculo seleciona a profundidade do brainstorming com base nas características da ideia na entrada.

### 3.1 Brainstorming Mínimo

**Quando:** O usuário descreve uma mudança pequena e bem delimitada em comportamento existente. O problema é claro e o espaço de solução é estreito.

**Fluxo:** Reenquadramento (1.1) → verificação básica de premissas (1.5) → Portão de Saída (1.7). Tipicamente 3-5 perguntas. O portão ainda se aplica — todos os quatro riscos devem ser endereçados mesmo para mudanças pequenas.

### 3.2 Brainstorming Padrão

**Quando:** O usuário descreve uma nova feature ou uma mudança significativa em funcionalidade existente. O problema precisa de exploração mas o domínio é compreendido.

**Fluxo:** Sequência completa — Reenquadramento → Divergência → Análise da Codebase → Convergência → Mapeamento de Premissas → Stress Test → Portão de Saída.

### 3.3 Brainstorming Profundo

**Quando:** A ideia é vaga, ambiciosa, transversal, ou o usuário não consegue articular o problema claramente mesmo após o reenquadramento inicial.

**Fluxo:** Sequência completa com sub-agentes paralelos explorando múltiplos ângulos simultaneamente:
- Agente analisando impacto na codebase em múltiplos domínios
- Agente explorando viabilidade técnica de diferentes abordagens
- Agente revisando decisões passadas e features relacionadas no SQLite

Resultados são consolidados e alimentados de volta no diálogo, habilitando questionamento mais rico ancorado em evidências.

### 3.4 Critérios de Seleção de Complexidade

O Oraculo avalia na entrada:

| Sinal | Mínimo | Padrão | Profundo |
|-------|--------|--------|----------|
| Usuário consegue declarar o problema claramente | Sim | Parcialmente | Não |
| Espaço de solução é estreito | Sim | Moderado | Amplo ou desconhecido |
| Área da codebase afetada | Componente único | Múltiplos componentes | Transversal |
| Decisões passadas existem no SQLite | Nenhuma relevante | Algumas relevantes | Muitas relevantes ou contraditórias |

## 4. Integração com o Modelo Operacional do Oraculo

### 4.1 Relacionamento com a Fase de Plan

A fase de Brainstorming produz o documento de requisitos referenciado no design principal:
- Job Stories definem **o que** precisa acontecer
- A OST fornece **por que** (conectado a outcomes)
- O Registro de Premissas define **o que deve ser validado primeiro**
- O Resumo de Impacto na Codebase define **restrições** para o DAG

A fase de Plan recebe esses artefatos e os decompõe em um DAG de tarefas. Nenhuma tarefa entra no DAG sem rastreabilidade a uma Job Story e um Outcome.

### 4.2 Ponto de Entrada

A fase de Brainstorming é invocada via `/oraculo:brainstorm`. Qualquer membro do time pode iniciá-la — Produto, Desenvolvimento ou qualquer pessoa com uma ideia. O Oraculo adapta a profundidade e a linguagem ao contexto do usuário.

### 4.3 Persistência no SQLite

Tudo da fase de Brainstorming é registrado:
- A ideia crua como originalmente declarada
- O problema reenquadrado
- Todas as perguntas feitas e respostas dadas
- A OST construída durante a sessão
- Premissas com seus scores de risco e status
- Resultados da análise da codebase
- A avaliação do portão de saída
- As Job Stories finais

Isso habilita continuidade entre sessões e membros do time. Qualquer membro do time pode consultar o SQLite para entender o raciocínio completo por trás dos requisitos — não apenas o que foi decidido, mas por quê, e o que foi explicitamente rejeitado.
