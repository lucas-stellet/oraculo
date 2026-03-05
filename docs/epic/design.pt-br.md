# Fase Epic — Design do Sistema

## 1. Estrutura da Sessão

A fase Epic segue um fluxo estruturado com dois macro-movimentos: **Divergir** (expandir o espaço do problema) e depois **Convergir** (estreitar para uma definição validada). O Oraculo avança por oito fases, progredindo apenas quando a condição de saída de cada fase é satisfeita.

### 1.0 Entrada — Setup da Sessão & Triagem

O usuário chega com uma ideia. O Oraculo captura a ideia crua, detecta o nível de reasoning e confirma que é um trabalho do tamanho de um epic.

**Sinais de triagem para escopo epic:**
- Envolve mais de uma área ou disciplina
- Tem partes que poderiam ser entregues independentemente
- Há incertezas significativas sobre como resolver

Se o trabalho parece pequeno e focado, o Oraculo sugere `/oraculo:story` no lugar.

**Condição de saída:** Ideia crua capturada, nível de reasoning determinado internamente, escopo de epic confirmado.

### 1.1 Reenquadramento

O Oraculo separa o problema de qualquer solução proposta antes de prosseguir.

**Perguntas de gatilho:**
- "Você descreveu o que quer construir. Que problema isso resolve para o usuário?"
- "Se essa feature não existisse, o que o usuário faria? Qual é o custo dessa alternativa?"
- "O que está explicitamente fora do escopo desta exploração?"

**Condição de saída:** O usuário articulou uma definição de problema que é independente de qualquer solução específica. O Oraculo registra a ideia crua como originalmente declarada e o problema reenquadrado como artefatos separados.

### 1.2 Divergência — Expandindo o Espaço do Problema

O Oraculo aplica templates de perguntas JTBD (incluindo Quatro Forças) para explorar dimensões que o usuário não considerou.

**Perguntas de situação e motivação do usuário:**
- "Em que situação específica o usuário encontra esse problema?"
- "O que o usuário está tentando realizar — funcionalmente, emocionalmente, socialmente?"
- "Que alternativas existem hoje? Por que são insuficientes?"
- "O que faria o usuário 'contratar' uma nova solução para esse trabalho?"

**Perguntas de forças de mudança:**
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

Quando não há contexto prévio (sem documentação, sem fontes definidas), o Oraculo dispara agentes de pesquisa para analisar a base de código existente. Isso roda em paralelo — não bloqueia a conversa. Quando contexto prévio existe (fontes definidas, documentação existente), o orquestrador conduz uma análise leve diretamente sem despachar agentes.

**O que os sub-agentes investigam:**
- Implementações existentes que se sobrepõem à ideia proposta
- Padrões arquiteturais que a nova feature deve respeitar
- Testes que cobrem comportamento relacionado (pontos potenciais de quebra)
- Restrições técnicas que limitam ou habilitam abordagens específicas

**Como os resultados entram na conversa:**
- "A análise encontrou que [componente X] já trata um caso similar. Você quer construir sobre isso ou criar algo independente? Por quê?"
- "Há um edge case em [área Y] que o código atual trata explicitamente. Como sua ideia se comporta nesse cenário?"

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
Cada premissa é avaliada por impacto e nível de evidência. Premissas de alto impacto com baixa evidência são priorizadas. O mecanismo de pontuação é interno — categorias de premissas e scores de risco não aparecem no documento de saída.

**Condição de saída:** Todas as premissas com score de risco acima do limiar foram reconhecidas pelo usuário. Cada premissa crítica (alto impacto, baixa evidência) tem evidência de suporte ou está sinalizada como risco conhecido para a fase de Plan.

### 1.6 Portão de Saída

O checkpoint final antes de avançar para o Approval Gate (§1.7). O Oraculo avalia quatro dimensões de risco:

**Checklist de saída (Quatro Riscos):**
1. **Valor** — O valor para o usuário está claro e a necessidade validada (ou explicitamente marcada como premissa)?
2. **Usabilidade** — A solução é compreensível e utilizável sem atrito excessivo?
3. **Viabilidade técnica** — A análise da codebase confirmou viabilidade (ou sinalizou riscos conhecidos)?
4. **Viabilidade de negócio** — As métricas de sucesso estão definidas e o impacto no negócio é compreendido?

**Comportamento do portão:** O Oraculo avalia cada risco e apresenta a avaliação ao usuário. Se qualquer risco estiver sem resposta, o Oraculo identifica a lacuna específica e redireciona para a etapa apropriada:
- Lacuna de valor → voltar a 1.2 (Divergência)
- Lacuna de usabilidade → voltar a 1.2 (Exploração de edge cases)
- Lacuna de viabilidade técnica → voltar a 1.3 (Análise da Codebase)
- Lacuna de viabilidade de negócio → voltar a 1.4 (Convergência)

Apenas quando os quatro riscos estão endereçados a sessão produz seus artefatos finais e passa para o Approval Gate (§1.7).

### 1.7 Geração de Artefatos

O Oraculo gera o Documento de Requisitos e o salva via CLI, depois cria uma versão para revisão humana. O fluxo tem quatro etapas:

1. **Gerar e salvar** — O Oraculo gera o Documento de Requisitos e o salva em `.oraculo/epics/<nome-epic>/requirements.md` via `oraculo epic save`.
2. **Criar versão** — O Oraculo chama `oraculo tools epic version <nome-epic>`, passando o markdown via stdin. Isso cria um snapshot versionado e o exibe no dashboard para revisão humana.
3. **Aguardar verdict** — O agente monitora reviews via `oraculo tools review list <version-id> --type epic`. O fluxo não avança até que um humano emita um verdict pelo dashboard.
4. **Tratar o verdict:**
   - `approved` — Requisitos são finalizados. O Oraculo sugere decompor o epic em stories com `/oraculo:story`.
   - `rejected` — Voltar a §1.2 (Divergência) para reabrir o espaço do problema com base no feedback do revisor.

## 2. Artefatos de Saída

A fase Epic produz um único Documento de Requisitos que alimenta diretamente stories e a fase de Plan.

### 2.1 Documento de Requisitos

O output primário — salvo em `.oraculo/epics/<nome-epic>/requirements.md`. O documento tem 9 seções:

1. Contexto — Narrativa da situação
2. Problema — Consequências concretas
3. Objetivos — O que alcançar, como medir, o que aprender
4. Publico-alvo — Quem é afetado
5. Historias de Usuario — Histórias de usuário em linguagem natural
6. Escopo — Dentro e fora
7. Armadilhas conhecidas — (opcional) Armadilhas de implementação
8. Dependencias — Componentes, serviços, sistemas
9. Questoes em aberto — Decisões não resolvidas

**Regra-chave:** O documento usa apenas linguagem de domínio. Terminologia de metodologia (JTBD, Quatro Forças, Assumption Mapping, etc.) é usada internamente durante a conversa mas nunca aparece no output.

Cada história de usuário na seção 5 é um candidato para decomposição em stories.

## 3. Escalando com a Complexidade

O Oraculo seleciona a profundidade da exploração com base nas características da ideia na entrada. A fase Epic usa dois níveis de complexidade:

### 3.1 Epic Padrão

**Quando:** O usuário descreve uma nova feature ou uma mudança significativa em funcionalidade existente. O problema precisa de exploração mas o domínio é compreendido.

**Fluxo:** Sequência completa — Setup → Reenquadramento → Divergência → Análise da Codebase → Convergência → Mapeamento de Premissas → Portão de Saída → Geração de Artefatos → Approval Gate.

### 3.2 Epic Profundo

**Quando:** A ideia é vaga, ambiciosa, transversal, ou o usuário não consegue articular o problema claramente mesmo após o reenquadramento inicial.

**Fluxo:** Sequência completa com sub-agentes paralelos explorando múltiplos ângulos simultaneamente:
- Agente analisando impacto na codebase em múltiplos domínios
- Agente explorando viabilidade técnica de diferentes abordagens
- Agente revisando padrões e convenções passadas relacionadas

Resultados são consolidados e realimentados no diálogo, possibilitando questionamentos mais ricos com base em evidências.

**Fluxo:** Sequência completa com sub-agentes paralelos — Setup → Reenquadramento → Divergência (+ agentes paralelos) → Análise da Codebase → Convergência → Mapeamento de Premissas → Portão de Saída → Geração de Artefatos → Approval Gate.

## 4. Níveis de Reasoning

A fase Epic opera em dois níveis de reasoning que compartilham a mesma disciplina socrática mas diferem no domínio de investigação.

### 4.1 Light Reasoning (Executor/Desenvolvedor)

- Recebeu uma tarefa de outra pessoa
- Expertise técnica, contexto de produto limitado
- Foco: entendimento da tarefa, riscos técnicos, edge cases
- Análise da Codebase é o **centro de gravidade**
- Escala para PM em lacunas de contexto de produto

### 4.2 Deep Reasoning (Tomador de Decisão/PM)

- Explorando ideia que ele identificou
- Tem dados de usuários, métricas de negócio, visão estratégica
- Foco: comportamento do usuário, impacto no negócio, viabilidade técnica
- Usa profundidade completa de exploração em dimensões do usuário, forças de mudança, constraints e premissas
- Responde todas as perguntas do gate autonomamente

### 4.3 Sinais de Seleção

| Sinal | Light | Deep |
|-------|-------|------|
| Usuário descreve tarefa recebida de outra pessoa | Sim | — |
| Usuário descreve ideia ou problema que ele identificou | — | Sim |
| Usuário tem métricas de negócio ou dados de usuários | — | Sim |
| Usuário foca na abordagem de implementação | Sim | — |
| Papel primariamente técnico | Sim | — |
| Papel envolve decisões de produto | — | Sim |

### 4.4 Matriz Complexidade × Reasoning

|  | Padrão | Complexidade Profunda |
|--|--------|-----------------------|
| **Light** | Fluxo Light completo com todas as fases. Análise da Codebase é centro de gravidade. Sem perguntas de negócio. | Fluxo Light completo com sub-agentes paralelos. Sessão longa, sem perguntas de negócio. |
| **Deep** | Fluxo completo — todos os frameworks, todas as perguntas, todos os gates. | Fluxo completo com sub-agentes paralelos. Duração e profundidade máximas. |

## 5. Integração com Stories

A fase Epic produz um Documento de Requisitos com histórias de usuário. Para implementar:

**Pré-condição:** Antes de decompor em stories, verificar que o `approval_status = approved` do epic. Um epic cujos requisitos ainda não foram aprovados por um revisor humano não deve ser decomposto — o artefato ainda pode mudar com base no verdict pendente.

1. Invocar `/oraculo:story <nome-epic>`
2. A skill Story lê `.oraculo/epics/<nome-epic>/requirements.md`
3. O usuário seleciona qual história de usuário ou requisito trabalhar
4. A skill Story roda uma sessão focada para produzir uma definição de story executável

Cada história de usuário pode gerar uma ou mais stories executáveis dependendo do escopo.

## 6. Caminho de Saída

Todos os artefatos do epic são salvos em `.oraculo/epics/<nome-epic>/requirements.md` dentro da aplicação-alvo. O nome do epic é derivado do domínio da ideia ou explicitamente solicitado ao usuário durante a geração do artefato.
