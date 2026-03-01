# Steering — Filosofia

## 1. Propósito

O Steering existe para dar ao Oraculo a identidade do projeto antes de qualquer ação ser tomada.

Agentes e skills operam no vácuo sem o Steering. Eles sabem como escrever código, explorar problemas, validar implementações — mas não sabem **para quem**, em qual **contexto** ou com quais **restrições** o projeto opera. O Steering preenche esse vácuo com dois documentos fundamentais: **Product** (o quê e para quem) e **Design** (como e com quê).

Sem Steering, o Oraculo faz perguntas genéricas. Com Steering, cada pergunta do Epic é informada pelo domínio da empresa, pela visão do produto, pela arquitetura existente e pelas convenções do time. **A qualidade da exploração é proporcional à qualidade do contexto.**

## 2. Crença Central

**Um agente sem contexto de produto é um programador genérico — competente, mas desconectado do problema real.**

Quando alguém pede "adicionar autenticação", a resposta correta depende de quem usa o produto (desenvolvedores integrando via API? consumidores finais?), em que setor a empresa opera (fintech com conformidade regulatória? B2B SaaS?) e quais decisões técnicas já foram tomadas (JWT? sessões? OAuth?). Sem esse contexto, o Oraculo produz soluções tecnicamente corretas, mas potencialmente desalinhadas.

O Steering é a diferença entre um agente que pergunta "que tipo de autenticação você quer?" e um que pergunta "você já usa JWT para a API pública — a autenticação deste novo módulo deve seguir o mesmo padrão ou há alguma razão para divergir?"

## 3. Fundações Teóricas

O Steering se baseia em princípios estabelecidos de contexto, modelagem de domínio e disciplina arquitetural.

### 3.1 Context-Driven Testing (Cem Kaner) — O Contexto Governa Todas as Decisões

Não existem melhores práticas universais — apenas práticas adequadas a um contexto específico. A mesma decisão que é correta para uma startup de 5 pessoas é irresponsável para um sistema financeiro regulado. O Steering torna o contexto explícito para que cada decisão do Oraculo seja informada pela realidade do projeto.

**Por que importa:** Sem contexto explícito, agentes recorrem a heurísticas genéricas. Um code agent escolhendo entre uma validação simples e um pipeline completo de conformidade não tem base para a decisão a menos que conheça o domínio do projeto, o ambiente regulatório e a tolerância ao risco. O Steering fornece essa base.

### 3.2 Bounded Context (Eric Evans, DDD) — A Linguagem Tem Fronteiras

Cada sistema opera dentro de fronteiras de linguagem e significado. "Pagamento" significa algo diferente para checkout, reconciliação e contabilidade. O Steering captura essas fronteiras: Product define a linguagem do domínio de negócio; Design define a linguagem do domínio técnico. Ambos impedem o Oraculo de confundir termos ou produzir código com abstrações incorretas.

**Por que importa:** Quando um agente encontra um termo ambíguo, ele o resolve com qualquer interpretação que pareça plausível. Se "usuário" significa tanto "consumidor final" quanto "integrador via API" no mesmo projeto, o agente vai confundi-los a menos que o Steering trace a fronteira. O desalinhamento da linguagem de domínio produz código que compila mas modela a realidade errada.

### 3.3 Separation of Concerns — Dois Documentos para Dois Domínios

A decisão de dividir o Steering em dois documentos — Product e Design — reflete a separação fundamental entre o domínio do problema (o quê, para quem, por quê) e o domínio da solução (como, com quê, onde). Misturar ambos em um único documento produz um artefato que nenhum público lê com conforto.

**Por que importa:** Pessoas de produto mantêm visão, personas e restrições de negócio. Engenheiros mantêm arquitetura, convenções e decisões técnicas. Dois documentos, cada um com seu público e propósito, garantem que ambos sejam mantidos. Um documento de contexto monolítico decai porque nenhuma pessoa única possui tudo nele.

## 4. O Que Torna o Steering do Oraculo Único

O Steering não é documentação genérica de projeto. É uma camada de contexto estruturada para consumo por agentes, com quatro propriedades específicas.

### 4.1 Leitura Obrigatória, Não Opcional

O Steering não é documentação de referência que o agente consulta se quiser. É **contexto obrigatório carregado antes de cada Epic e de cada Story.** As skills leem o Steering como pré-condição — se ele não existe, a skill informa e oferece criação. O agente nunca opera sem saber onde está.

### 4.2 Complementar ao CLAUDE.md, Não Redundante

O CLAUDE.md contém instruções operacionais para o Claude Code. O Steering contém a identidade do projeto. São camadas diferentes. **O CLAUDE.md diz "como trabalhar aqui." O Steering diz "o que existe aqui."** Ambos são lidos; nenhum substitui o outro.

### 4.3 Criação Híbrida — Template com Validação

O Oraculo não exige que o usuário escreva do zero, nem gera automaticamente. Ele oferece templates estruturados, o usuário os preenche e então o Oraculo valida e enriquece com perguntas complementares. **O resultado reflete o conhecimento do usuário, não a imaginação do agente.** O que o usuário escreveu é tratado como a fonte da verdade.

### 4.4 Dois Documentos, Dois Domínios, Dois Ritmos

Product e Design mudam em ritmos diferentes. A visão do produto e a missão da empresa raramente mudam. A stack técnica e as convenções de código evoluem com o projeto. **Separar em dois documentos permite que cada um seja atualizado no seu ritmo natural.**

## 5. Princípios Específicos do Steering

Estes princípios estendem os princípios centrais do Oraculo para esta camada específica:

- **Steering antes de tudo** — Nenhum Epic ou Story começa sem que os documentos de Steering sejam lidos. Se eles não existem, o Oraculo informa e oferece criação. Nunca ignora silenciosamente a ausência.

- **O usuário é a fonte, nunca o agente** — O Steering captura o que o usuário sabe sobre seu projeto. O Oraculo pode questionar lacunas, sugerir que uma seção está incompleta ou pedir esclarecimentos — mas nunca preenche conteúdo por conta própria. Um documento de Steering com informações inventadas é pior do que nenhum documento.

- **Estável, não imutável** — O Steering é um documento vivo, não uma constituição. Ele muda quando a realidade do projeto muda. Mas as mudanças são deliberadas e explícitas, nunca silenciosas. O Oraculo não modifica o Steering como efeito colateral de um Epic.

- **Contexto, não instrução** — O Steering informa, não comanda. Ele diz "a empresa opera no setor financeiro" e "a API usa JWT" — não diz "sempre use JWT" ou "nunca faça X." Princípios de negócio são restrições, não regras absolutas. Epic e Story podem questionar um princípio se houver razão — mas o Steering garante que o questionamento seja consciente, não acidental.

- **Lacuna visível é melhor que ficção** — Se o usuário não conhece a visão do produto ou não definiu convenções de código, a seção fica vazia ou marcada como "a definir." O Oraculo trabalha com o que tem. Um campo vazio é um sinal honesto; um campo inventado é uma armadilha.

## 6. Relacionamento com Outras Fases

O Steering é uma camada transversal — não uma fase do modelo operacional, mas um pré-requisito que alimenta todas elas.

### 6.1 Epic — Maior Impacto

A exploração começa a partir de perguntas informadas em vez de genéricas. O documento Product dá à fase Epic a linguagem de domínio, as personas de usuário e as restrições de negócio. O documento Design dá a ela consciência arquitetural. Juntos, eles transformam a exploração socrática de "me fale sobre seus usuários" para "seus clientes B2B integram via API — como esta nova feature afeta os contratos de integração deles?"

### 6.2 Story — Compensa a Lacuna de Contexto

Stories avulsas — aquelas criadas sem um Epic — não têm a exploração profunda que produz contexto rico. O Steering compensa fornecendo o contexto de base que a story de outra forma perderia completamente. A skill de story lê o Steering para entender o que é o projeto antes de perguntar o que é a tarefa.

### 6.3 Plan e Execute — Alimenta o Orquestrador e os Code Agents

O documento Design alimenta o orquestrador com a arquitetura para decomposição e os code agents com convenções, estrutura do projeto e decisões técnicas. Um orquestrador que sabe que o projeto usa arquitetura hexagonal decompõe o trabalho de forma diferente de um que assume uma camada monolítica de controllers.

### 6.4 Validate — Lente de Validação Mais Ampla

O QA agent ganha contexto além do diff imediato. Consciência em nível de produto — requisitos de conformidade, invariantes de domínio, expectativas do usuário — estende a validação de "passa nos testes" para "pertence a este produto." Um QA agent que sabe que o projeto opera no setor de saúde vai sinalizar um campo de PII não criptografado que um agente sem contexto aprovaria.

### 6.5 CLAUDE.md — Coexistem Sem Sobreposição

O CLAUDE.md são instruções operacionais. O Steering é a identidade do projeto. São documentos independentes consumidos em momentos diferentes. O CLAUDE.md diz ao agente como rodar testes, qual linter usar e onde os arquivos ficam. O Steering diz ao agente o que o produto faz, a quem serve e como a arquitetura é organizada. Nenhum documento deve conter conteúdo que pertença ao outro.

## 7. O Que o Steering Não É

- **Não é substituto para o Epic.** O Steering é contexto estável; o Epic é investigação pontual. O Steering diz "nossos usuários são provedores de saúde." O Epic pergunta "qual fluxo de trabalho específico quebra quando um provedor tem mais de 200 pacientes?"

- **Não é uma especificação técnica.** O Design descreve a arquitetura em alto nível — padrões, convenções, fronteiras — não ADRs completos, contratos de API ou schemas de banco de dados. Esses pertencem ao próprio codebase.

- **Não é gerado pelo Oraculo.** O conteúdo vem do usuário; o Oraculo valida, nunca inventa. O agente pode perguntar "esta seção parece incompleta — você gostaria de adicionar detalhes sobre sua estratégia de deploy?" Ele nunca preenche a resposta por conta própria.

- **Não é versionado junto com o código do repositório.** O Steering vive em `.oraculo/steering/`, infraestrutura operacional ao lado do banco de dados SQLite. Pode ser commitado por decisão do time, mas o Oraculo não assume nem exige isso.

- **Não é obrigatório para usar o Oraculo.** Um projeto sem Steering funciona — o Oraculo opera com menos contexto e faz mais perguntas. A ausência é informada, nunca silenciada. A skill informa ao usuário que o Steering não existe e oferece ajuda para criá-lo.
