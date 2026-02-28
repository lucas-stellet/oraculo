# Oraculo Commands — Filosofia

## 1. Propósito

Commands são a superfície conversacional do Oraculo.

São o ponto onde uma intenção humana — "tenho uma ideia de feature", "preciso implementar esta story" — entra no sistema e se torna trabalho estruturado. Cada command mapeia para uma fase do modelo operacional do Oraculo. O usuário invoca um slash command, o command conduz um diálogo socrático, e o resultado alimenta o CLI Trust Layer e a força de trabalho de agentes.

Commands não são scripts. Eles não automatizam um checklist. Eles codificam uma disciplina — uma sequência de perguntas, frameworks e quality gates que forçam o usuário (e o LLM) a pensar antes de agir. O command é a fronteira entre intenção casual e execução rigorosa.

## 2. Crença Central

**Um command deve tornar o caminho disciplinado fácil e o caminho indisciplinado impossível.**

LLMs são ansiosos para agir. Dado um prompt vago, eles geram código, pulam a descoberta, enfraquecem restrições e produzem lixo com confiança. Commands existem para prevenir isso. Eles canalizam a energia do LLM por um fluxo de trabalho estruturado onde pular etapas é mais difícil do que segui-las. O command não depende do julgamento do modelo para manter o rigor — ele estrutura a interação de forma que o rigor seja o caminho de menor resistência.

## 3. Composição em Vez de Monolitos

**Um command é um ponto de entrada enxuto, não um programa autossuficiente.** O arquivo SKILL.md define o fluxo de trabalho, a persona e os quality gates. Ele não contém os frameworks, question banks ou templates de artefatos inline. Esses vivem em arquivos de referência, carregados apenas quando necessário.

Essa é uma restrição deliberada. Um prompt monolítico de centenas de linhas degrada a performance do LLM — a atenção se dilui, instruções no topo são esquecidas quando o modelo chega ao final. Documentos menores e focados, carregados no momento certo, produzem melhor aderência do que um único arquivo massivo carregado de antemão.

**Arquivos de referência são internos, não commands.** Apenas arquivos SKILL.md aparecem como slash commands. Arquivos de suporte — question banks, descrições de frameworks, templates — vivem ao lado do skill mas nunca poluem a lista de commands. O usuário vê uma interface limpa; o skill tem acesso a material de referência profundo quando o fluxo de trabalho exige.

**Commands nunca duplicam o que o CLI impõe.** Validação de schema, persistência de estado, operações idempotentes — esses pertencem ao Trust Layer. Um command chama o CLI; ele não reimplementa suas garantias. Se um command contém lógica que poderia ser um contrato do CLI, ela está no lugar errado.

**O vocabulário é mínimo e consistente.** Todo conceito que um command introduz — um nome de fase, um rótulo de framework, um nome de anti-padrão — deve merecer seu lugar. Jargão específico de domínio e metáforas extensas criam uma curva de aprendizado que compete com o fluxo de trabalho real. Commands usam o menor vocabulário que alcança precisão. Novos termos são introduzidos apenas quando linguagem simples seria ambígua.

## 4. Contenção Comportamental

**Commands moldam o comportamento do LLM através de estrutura, não de esperança.** Dizer a um LLM "seja minucioso" é decoração. Dar a ele um phase gate que exige três premissas validadas antes de prosseguir é contenção. Commands usam técnicas estruturais:

- **Phase gates** impedem progresso para frente até que critérios de qualidade sejam atendidos. O modelo não pode pular a descoberta para começar o planejamento — o fluxo de trabalho não oferece esse caminho.
- **Anti-padrões nomeados** dão ao LLM handles cognitivos para modos de falha. "Premature solutioning" é mais fácil de detectar e evitar do que "não pule para conclusões". Quando um modo de falha tem nome, o modelo consegue fazer pattern-match contra ele.
- **Regras de desvio explícitas** definem o que acontece quando o usuário recua ou pede para pular. O command sabe quais etapas são comprimíveis e quais são invioláveis. Isso não é rigidez — é clareza sobre onde a flexibilidade existe e onde ela não existe.

**Repetição de regras críticas é intencional, não descuidada.** LLMs perdem aderência a instruções em conversas longas. Regras que governam segurança, qualidade ou integridade do fluxo de trabalho aparecem no ponto de uso — não apenas em um preâmbulo que o modelo já esqueceu há muito tempo. Isso não é redundância; é reforço no momento de maior risco de deriva.

**Aprovação é um veredicto humano, não um sinal verbal.** Quando um command chega a um approval gate — requisitos prontos para revisão, um plano pronto para execução — ele chama `oraculo tools approval request` para submeter o artefato e entra no estado `awaiting_approval`. O dashboard apresenta o artefato a um revisor humano, que emite um veredicto: `approved` (avançar), `rejected` (retornar à fase geradora) ou `needs_revision` (retornar com comentários). "Parece bom" em conversa não é aprovação. Um veredicto entregue pelo approval gate do dashboard é aprovação. Commands nunca avançam com confirmação verbal para transições irreversíveis. O fluxo não avança até que o veredicto seja recebido.

## 5. Contexto como Recurso Finito

**Cada token no contexto de um command tem um custo.** O context window não é armazenamento infinito — é memória de trabalho. Commands o tratam com a mesma disciplina que um sistema em tempo real aplica à RAM.

**Carregue tarde, descarregue cedo.** Material de referência entra no contexto apenas quando a fase do fluxo de trabalho exige. Um question bank para assumption mapping não é carregado durante o enquadramento do problema. Um template de requisitos não é carregado até que o usuário tenha validado a exploração. Esse carregamento just-in-time mantém o contexto ativo focado na tarefa atual.

**As próprias instruções do command competem com a conversa.** À medida que o diálogo cresce, a proporção de tokens de instrução para tokens de conversa diminui. Commands são escritos para ser concisos — cada frase merece seu lugar. Instruções verbosas não produzem melhor aderência; produzem atenção diluída.

**Estado pertence ao armazenamento, não à conversa.** Resultados intermediários — premissas validadas, restrições exploradas, requisitos parcialmente formados — são persistidos pelo CLI assim que se estabilizam. Se o context window transbordar ou a sessão for interrompida, nenhum trabalho validado é perdido. A conversa pode ser reconstruída a partir do estado persistido; estado persistido não pode ser reconstruído a partir de uma conversa perdida.

## 6. As Três Camadas

Commands operam dentro de uma hierarquia estrita. Cada camada tem uma única responsabilidade.

**Commands** são a interface conversacional. Eles detêm o fluxo socrático — quais perguntas fazer, quais frameworks aplicar, quais artefatos produzir. Eles interagem com o usuário, guiam a descoberta e produzem saída estruturada. Commands são interpretativos — eles raciocinam, adaptam e respondem a nuances.

**O CLI Trust Layer** é a fundação determinística. Commands delegam toda persistência, validação e gerenciamento de estado ao CLI. O command declara intenção; o CLI garante correção. Essa fronteira é absoluta — um command nunca escreve diretamente no banco de dados, constrói caminhos de arquivo ou gerencia schemas. Se o CLI rejeita uma operação, o command não tenta novamente com restrições mais frouxas — ele expõe a rejeição ao usuário.

**Agents** são a força de trabalho de execução. Quando a saída de um command (requisitos, stories, planos) está pronta para implementação, o orchestrator despacha agentes. Commands não executam código, rodam testes ou modificam a codebase. Eles produzem os artefatos que os agentes consomem.

O fluxo é sempre: **Command descobre e define -> CLI persiste e valida -> Agents executam e entregam.** Nos approval gates, o fluxo pausa: o Dashboard apresenta o artefato e aguarda um veredicto humano antes de o workflow avançar. Nenhuma camada faz o trabalho de outra.

Essa separação não é conveniência organizacional — é uma garantia de confiabilidade. Quando cada camada faz exatamente uma coisa, falhas são locais, depuráveis e recuperáveis. Um command que também persiste dados é impossível de testar. Um agente que também faz descoberta é incontrolável. As camadas existem porque misturar responsabilidades produz sistemas que são frágeis de formas invisíveis até que quebrem.

## 7. Entradas e Saídas

**Commands aceitam argumentos, não configuração.** Um command recebe o que precisa por `$ARGUMENTS` — a intenção do usuário, uma referência a um artefato existente ou uma declaração de escopo. Não há arquivos de config, variáveis de ambiente ou estado implícito que altere o comportamento. O comportamento do command é completamente determinado pelo seu SKILL.md e seus argumentos.

**A saída do command é um artefato, não um log de conversa.** O diálogo socrático é o processo; o artefato é o produto. Um epic command produz um documento de requisitos. Um story command produz uma definição de story. Esses artefatos são estruturados, validados e persistidos pelo CLI. A conversa que os produziu é efêmera — o artefato é o que sobrevive.

**Handoffs são explícitos e controlados por approval gate.** Quando um command completa sua fase e o fluxo de trabalho faz transição — da descoberta para o planejamento, do planejamento para a execução — o handoff requer um veredicto de aprovação antes de se completar. O command submete o artefato ao Approval Gate via `oraculo tools approval request`, entra em `awaiting_approval` e não avança até que o dashboard entregue um veredicto. Apenas um veredicto `approved` permite que a próxima fase comece com contexto limpo e uma entrada validada. Isso previne context bleed e garante que cada fase opere sobre fatos acordados, não artefatos não revisados ou deriva conversacional.

## 8. Portabilidade

**Commands são arquivos markdown simples. Sem build step, sem compilação, sem runtime.** Um SKILL.md e seus arquivos de referência são copiados para o diretório `.claude/` de um projeto e funcionam. Não há transpilação de YAML para XML, nenhum template engine, nenhum plugin registry. O arquivo é o command.

Essa restrição é inegociável. Build steps criam version drift, dependências de toolchain e modos de falha que não têm nada a ver com o propósito do command. Um command que requer um compilador é um command que vai quebrar em ambientes que o autor nunca testou.

## 9. O Que Este Documento Não É

Este documento define os princípios que governam o design de commands. Ele não especifica commands individuais, seus conteúdos de SKILL.md ou suas estruturas de arquivos de referência. Esses pertencem ao documento de design.

Este documento não substitui a filosofia do CLI (docs/cli/philosophy.md) nem a filosofia de agentes (docs/agents/philosophy.md). Commands dependem do CLI para operações determinísticas e dos agentes para execução. Este documento define como commands ocupam a camada interpretativa entre a intenção do usuário e a ação do sistema.
