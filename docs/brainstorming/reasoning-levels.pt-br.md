# Fase de Brainstorming — Níveis de Reasoning

## 1. Dois Modos, Mesma Disciplina

A fase de Brainstorming opera em dois níveis de reasoning. Ambos compartilham a mesma disciplina socrática — reenquadrar, divergir, revelar premissas, gate antes de avançar — mas diferem no **domínio de investigação** e na **autoridade epistêmica**.

- **Light Reasoning** — O usuário é um executor (tipicamente um desenvolvedor) que recebeu uma tarefa definida por outros. Ele tem expertise técnica mas contexto limitado de produto. O brainstorming foca em: o usuário entende a tarefa corretamente? Quais são os riscos técnicos? Que edge cases existem no sistema?

- **Deep Reasoning** — O usuário é um tomador de decisão (tipicamente PM, designer ou líder técnico) com contexto completo: dados de usuários, métricas de negócio, visão estratégica. O brainstorming cobre todas as dimensões — usuário, negócio e técnica — com o toolkit completo de frameworks.

**Distinção fundamental:** Light não é uma versão degradada do Deep. É uma reorientação. O eixo muda de "usuário e negócio" para "código e sistema." Os princípios socráticos são invariantes. O que muda é o domínio das perguntas e quem tem autoridade para respondê-las.

## 2. Como o Nível de Reasoning é Selecionado

O Oraculo determina o nível de reasoning na entrada da sessão com base em sinais do usuário:

| Sinal | Light | Deep |
|-------|-------|------|
| Usuário descreve uma tarefa recebida de outra pessoa | Sim | — |
| Usuário descreve uma ideia ou problema que ele identificou | — | Sim |
| Usuário tem métricas de negócio ou dados de usuários para compartilhar | — | Sim |
| Usuário foca na abordagem de implementação | Sim | — |
| O papel do usuário é primariamente técnico | Sim | — |
| O papel do usuário envolve decisões de produto | — | Sim |

Quando ambíguo, o Oraculo pergunta: "Você está explorando uma ideia que quer validar, ou implementando algo que já foi decidido?"

O usuário também pode solicitar explicitamente um nível, e o Oraculo pode sugerir trocar no meio da sessão se a conversa revelar que o usuário tem mais (ou menos) contexto do que assumido inicialmente.

## 3. Fundações Teóricas por Nível

### Frameworks comuns aos dois níveis

| Framework | Aplicação |
|-----------|-----------|
| **Double Diamond** (divergir/convergir) | Ambos os níveis divergem antes de convergir. Light diverge no espaço técnico; Deep diverge no espaço de produto. |
| **Assumption Mapping** | Ambos os níveis revelam e pontuam premissas ocultas. As categorias diferem (ver seção 4). |
| **Quatro Forças de Progresso** | Ambos os níveis mapeiam forças de adoção e resistência. Light as aplica à adoção técnica (mudança no sistema); Deep as aplica à adoção pelo usuário (mudança de comportamento). |

### Frameworks exclusivos do Deep Reasoning

| Framework | Por que apenas Deep |
|-----------|---------------------|
| **JTBD — dimensões completas** (funcional, emocional, social) | Requer dados de pesquisa com usuários e contexto comportamental que executores tipicamente não têm. |
| **OST — construção** | Construir a Opportunity Solution Tree requer conhecimento de outcomes de negócio e oportunidades mapeadas de usuários. Executores recebem tarefas já conectadas a uma oportunidade. |
| **Working Backwards (PR/FAQ)** | Escrever um press release fictício requer clareza sobre valor para o usuário e posicionamento de produto que executores não conseguem fornecer. |

### Frameworks com ênfase diferente

| Framework | Ênfase Light | Ênfase Deep |
|-----------|--------------|-------------|
| **Job Stories** | System Stories: "Quando [estado/evento do sistema], o sistema deve [comportamento], para que [resultado observável/testável]." | User Stories: "Quando [situação do usuário], eu quero [motivação], para que eu possa [resultado esperado]." |
| **OST — consulta** | Consulta a árvore existente para verificar se a tarefa se conecta a uma oportunidade mapeada. Uso passivo. | Constrói e estende a árvore com novas oportunidades, soluções e experimentos. Uso ativo. |
| **Quatro Forças** | Push: "Que limitação técnica está motivando essa tarefa?" Pull: "O que torna essa abordagem melhor arquiteturalmente?" Ansiedade: "O que pode quebrar para outros sistemas?" Hábito: "Que padrões de código devem ser preservados?" | Push: "O que frustra o usuário o suficiente para motivar mudança?" Pull: "O que torna isso atraente?" Ansiedade: "O que faz usuários hesitarem?" Hábito: "Que comportamento atual isso interrompe?" |

## 4. Fluxo da Sessão por Nível

### 4.1 Entrada — Reenquadramento

| Aspecto | Light | Deep |
|---------|-------|------|
| **Perguntas** | "Você recebeu essa tarefa. Qual é o problema que ela resolve, conforme descrito por quem te passou?" / "Como você sabe que entendeu o que foi pedido? Que comportamento do sistema vai mudar?" / "O que está fora do escopo?" | "Você descreveu o que quer construir. Que problema isso resolve para o usuário?" / "Se essa feature não existisse, o que o usuário faria?" / "O que está fora do escopo?" |
| **Condição de saída** | Usuário articula a mudança esperada no comportamento do sistema independentemente da abordagem de implementação. | Usuário articulou uma definição de problema independente de qualquer solução específica. |

### 4.2 Divergência

| Aspecto | Light | Deep |
|---------|-------|------|
| **Perguntas JTBD** | "Que parte do código trata esse caso hoje? Por que é insuficiente?" / "Que pré-condições devem existir para esse comportamento ser acionado?" | JTBD completo: dimensões funcional, emocional, social. |
| **Quatro Forças** | Forças técnicas (ver tabela da seção 3). | Forças de adoção pelo usuário (ver tabela da seção 3). |
| **Edge cases** | "O que acontece com inputs inválidos, estado concorrente ou erros de rede?" / "Que outros serviços, filas ou jobs dependem desse componente?" | "O que acontece se o usuário fizer [comportamento inesperado]?" / "Quem pode ser prejudicado mesmo não sendo o alvo?" |
| **Condição de saída** | Pelo menos: edge cases técnicos explorados, dependências do sistema identificadas, um risco que o usuário não havia considerado revelado. | Pelo menos: três dimensões JTBD exploradas, todas as quatro forças mapeadas, um edge case que o usuário não havia considerado revelado. |

### 4.3 Análise da Codebase (Paralela)

Idêntica em ambos os níveis. No Light, esta etapa tem **peso maior** — é o centro de gravidade da sessão, não um complemento paralelo. Resultados entram na conversa mais cedo e com mais influência nas perguntas subsequentes.

### 4.4 Convergência

| Aspecto | Light | Deep |
|---------|-------|------|
| **Perguntas** | "Qual é o requisito funcional central — o que obrigatoriamente deve mudar no sistema?" / "Descreva o comportamento esperado do sistema em uma frase." / "Que comportamento observável muda após o deploy?" / "Se pudesse resolver apenas um aspecto, qual entregaria mais valor?" | "Qual é a dor central — causa raiz, não sintoma?" / "Escreva o problema em uma frase do ponto de vista do usuário." / "Que métrica de negócio isso afeta?" / "Se pudesse resolver apenas um aspecto, qual entregaria mais valor?" |
| **Condição de saída** | Declaração clara de comportamento do sistema + critérios de verificação observáveis. | Definição clara do problema + outcome-alvo (métrica mensurável ou mudança de comportamento do usuário). |

### 4.5 Mapeamento de Premissas

| Categoria | Light | Deep |
|-----------|-------|------|
| **Hipótese de problema** | Registrada como "dependência externa" — o PM validou isso. Não é reanalisada pelo dev. | Totalmente explorada e pontuada. |
| **Hipótese de solução** | Avaliada: "essa implementação realmente resolve o caso de uso descrito?" | Totalmente explorada e pontuada. |
| **Hipótese de valor** | Registrada como "dependência externa" — fora do escopo do dev para validar. | Totalmente explorada e pontuada. |
| **Hipótese de viabilidade** | **Foco central.** Adequação à arquitetura, riscos de dependência, potencial de regressão, implicações de performance. | Explorada e pontuada junto com as outras categorias. |

O mecanismo de pontuação de risco é idêntico em ambos os níveis.

### 4.6 Stress Test

| Aspecto | Light | Deep |
|---------|-------|------|
| **Método** | Checklist Técnico de Viabilidade (substitui PR/FAQ) | PR/FAQ (Working Backwards) |
| **Perguntas** | "A abordagem cabe na arquitetura atual sem refatoração significativa?" / "Que testes devem ser escritos para cobrir os edge cases mapeados?" / "Qual seria a principal objeção de outro dev revisando esse PR?" / "Como você verifica que funciona no dia do deploy? Quais são os critérios de aceite técnicos?" | "Se lançasse amanhã, como seria o anúncio?" / "Qual seria a pergunta de um cliente cético?" / "Qual seria a objeção da engenharia?" / "Como mediria sucesso em três meses?" |
| **Condição de saída** | Usuário responde todas as quatro sem contradição. | Usuário responde todas as quatro sem hesitação ou contradição. |

### 4.7 Portão de Saída

Os quatro riscos são avaliados em ambos os níveis, mas com autoridade e comportamento de escalação diferentes:

| Risco | Light | Deep |
|-------|-------|------|
| **Valor** | "O comportamento esperado descrito na tarefa é suficientemente claro para implementar?" Se obscuro: **escalar para PM** (não retornar a etapa anterior). | "O valor para o usuário está claro e a necessidade validada?" Se obscuro: retornar à Divergência. |
| **Usabilidade** | "A solução é consistente com os padrões do codebase e fácil de manter?" Se lacuna: retornar à exploração de edge cases. | "A solução é compreensível e utilizável sem atrito excessivo?" Se lacuna: retornar à Divergência. |
| **Viabilidade técnica** | Autoridade total. Idêntico ao Deep. Se lacuna: retornar à Análise da Codebase. | Se lacuna: retornar à Análise da Codebase. |
| **Viabilidade de negócio** | "Os critérios de aceite técnicos estão definidos e testáveis?" Se não: **escalar para PM** (não retornar a etapa anterior). | "Métricas de sucesso definidas e impacto no negócio compreendido?" Se lacuna: retornar à Convergência. |

**Comportamento de escalação (apenas Light):** Quando o dev não consegue responder uma pergunta do gate porque requer contexto de produto, o Oraculo não o faz voltar por etapas do brainstorming. Em vez disso, gera um artefato estruturado de "Perguntas para o PM" — perguntas específicas que o PM deve responder antes que o dev possa prosseguir. O gate só abre depois que o PM responde ou o time decide aceitar o risco explicitamente.

## 5. Artefatos de Saída por Nível

| Artefato | Light | Deep |
|----------|-------|------|
| **Job Stories** | System Stories: "Quando [estado/evento do sistema], o sistema deve [comportamento], para que [resultado observável/testável]." | User Job Stories: "Quando [situação], eu quero [motivação], para que eu possa [resultado]." |
| **OST** | Árvore rasa partindo do requisito recebido: Requisito → Solução → Premissa de Viabilidade. Sem nó de Outcome (fora do escopo do dev). | Árvore completa: Outcome → Oportunidade → Solução → Premissa. |
| **Registro de Premissas** | Filtrado: apenas categorias Solução e Viabilidade pontuadas. Categorias Problema e Valor registradas como "dependência externa — fora do escopo do dev para validar." | Completo: todas as quatro categorias pontuadas e rastreadas. |
| **Resumo de Impacto na Codebase** | **Expandido** — este é o artefato principal. Inclui: componentes afetados, contratos de interface que mudam, testes existentes que podem quebrar, decisões técnicas anteriores relevantes, estimativa de complexidade de implementação. | Padrão: componentes, restrições, edge cases, decisões passadas. |
| **Perguntas para o PM** (apenas Light) | Gerado quando o portão de saída identifica lacunas que o dev não consegue preencher. Lista estruturada de perguntas específicas com contexto da sessão de brainstorming. | Não se aplica — o usuário Deep tem autoridade para responder essas perguntas. |

## 6. Interação com Escala de Complexidade

Nível de reasoning (Light/Deep) e complexidade (Mínimo/Padrão/Profundo) são **dois eixos independentes**. Eles se combinam em uma matriz:

|  | Mínimo | Padrão | Complexidade Profunda |
|--|--------|--------|-----------------------|
| **Light** | Reenquadramento + premissas de viabilidade + gate adaptado. 3-5 perguntas técnicas. | Fluxo Light completo: Reenquadramento → Edge Cases → Análise da Codebase → Convergência Técnica → Mapeamento de Premissas → Checklist de Viabilidade → Portão de Saída. | Fluxo Light completo com sub-agentes paralelos analisando codebase em múltiplos domínios. Sessão longa, mas sem perguntas de negócio. |
| **Deep** | Reenquadramento + verificação básica de premissas + gate completo. 3-5 perguntas. | Fluxo completo conforme descrito no documento de design principal. Todos os frameworks, todas as perguntas, todos os gates. | Fluxo completo com sub-agentes paralelos para codebase, viabilidade e revisão de decisões passadas. Duração e profundidade máximas em todas as dimensões. |

**Caso especial — Light + Complexidade Profunda:** Um dev recebe uma tarefa que parece simples mas afeta múltiplos domínios da codebase (ex: mudança em um modelo compartilhado). O Oraculo seleciona Complexidade Profunda pela amplitude do impacto técnico, mas mantém Light porque o dev não tem contexto de negócio. O resultado é uma sessão longa com muitas perguntas técnicas e múltiplos sub-agentes, sem nenhuma pergunta sobre métricas ou usuários.
