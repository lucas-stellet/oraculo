# Fase Story — Filosofia

## 1. Propósito

A fase Story existe para transformar itens de trabalho em unidades de trabalho bem definidas e executáveis.

Nem todo trabalho precisa da exploração profunda de um Epic. Quando o problema é compreendido, o escopo é claro e o trabalho é focado, o que se precisa é de um processo disciplinado para defini-lo com precisão — claro o suficiente para um desenvolvedor implementar sem ambiguidade.

A fase Story aplica a mesma disciplina socrática do Epic, mas com menos profundidade e mais foco. Ela faz as perguntas essenciais, revela as premissas críticas e produz um artefato compacto que captura O QUÊ e POR QUÊ.

## 2. Crença Central

**Mesmo tarefas claras carregam premissas ocultas.**

Quando alguém diz "adicionar um botão de deletar na página de perfil", as crenças implícitas são: usuários querem deletar seus perfis, a deleção é reversível (ou não é), o botão deve ser visível (não escondido atrás de um menu), e a ação não propaga para outros dados. Essas premissas são invisíveis até que alguém pergunte sobre elas.

A fase Story torna as perguntas essenciais inevitáveis — não através de exploração exaustiva, mas através de investigação focada que captura os pontos cegos mais perigosos.

## 3. Fundações Teóricas

A fase Story se baseia em dois frameworks, simplificados para trabalho focado:

### 3.1 Double Diamond (Simplificado)

A sessão segue um diamante comprimido: uma breve divergência durante o reenquadramento (explorar o problema de diferentes ângulos), depois convergência rápida para uma definição focada. Diferente do diamante completo do Epic, a divergência da Story é estreita — abre apenas o suficiente para capturar ângulos perdidos, depois fecha rapidamente.

**Por que importa:** Mesmo para trabalho pequeno, o instinto de pular direto para o "como" ignora o "por quê". Uma breve divergência captura problemas que seriam caros de corrigir após a implementação.

### 3.2 Assumption Mapping (Simplificado)

Identifica as 2-3 premissas mais importantes e avalia seu nível de evidência (alto/médio/baixo). Diferente do registro completo pontuado do Epic, a Story foca em capturar as premissas mais críticas sem análise exaustiva.

**Por que importa:** Todo trabalho carrega premissas. A Story não precisa mapear todas — precisa revelar as que fariam o trabalho falhar ou precisar ser refeito se estiverem erradas.

## 4. O Que Torna a Definição de Story do Oraculo Única

### 4.1 Mesma Disciplina, Menos Profundidade

A fase Story não é um atalho ou um skip. É a mesma abordagem socrática — reenquadrar, verificar premissas, gate antes de gerar — aplicada na profundidade certa para a complexidade do trabalho. A disciplina nunca cai; o escopo estreita.

### 4.2 Consciência do Epic

Stories podem ser standalone ou derivadas de um Epic. Quando derivadas, a skill Story lê o Epic pai, entende o contexto do REC-N específico sendo trabalhado e adapta suas perguntas de acordo. O usuário não repete contexto já capturado no Epic.

### 4.3 Pronta para Escalação

Se durante uma sessão de Story o trabalho se revela maior do que o esperado — múltiplas áreas, alta incerteza, escopo vago — o Oraculo sugere escalar para `/oraculo:epic` em vez de tentar forçar trabalho complexo em uma definição do tamanho de uma story.

### 4.4 Pronta para Aprovação

O artefato de Story é produzido com revisão humana em mente. Um revisor que encontra a story pelo dashboard não tem contexto da sessão — não participou do diálogo. O artefato deve ser autocontido: motivação, requisito, escopo e premissas críticas devem ser legíveis para alguém que lê apenas o documento. O Oraculo trata a prontidão para aprovação como critério de qualidade ao lado da completude. Um artefato que um revisor não consegue avaliar de forma independente não está finalizado.

## 5. Princípios Específicos da Definição de Story

Estes princípios estendem os princípios centrais do Oraculo para esta fase específica:

- **Reenquadrar antes de definir** — Mesmo para tarefas claras, separar o problema da solução proposta. "Adicionar um botão de deletar" se torna "usuários precisam de uma forma de remover seu perfil."

- **Proibido falar de solução antes de clareza no problema** — Detalhes de implementação são proibidos até que o problema e o resultado desejado estejam claros.

- **Premissas são explícitas** — Toda premissa crítica deve ser declarada e reconhecida. Nem todas as premissas precisam de pontuação — mas as perigosas devem ser visíveis.

- **Gate antes de gerar** — A fase Story tem dois checkpoints obrigatórios antes que o artefato chegue ao agente. O **exit gate** é uma verificação de qualidade automática e interna: se os quatro riscos não forem endereçados, o artefato não é gerado. A **version review** é um verdict humano: uma vez que o artefato é produzido, uma versão é criada via `oraculo tools story version` e submetida para revisão pelo dashboard. A story não avança para execução até que um verdict chegue — `approved` para prosseguir ou `rejected` para retornar à definição. Ambos os gates são obrigatórios. Nenhum pode ser pulado.

- **Escalar, não forçar** — Se o trabalho é grande demais para uma story, sugerir um Epic. Não comprimir trabalho complexo em um formato simples.

## 6. Relacionamento com Epics

Uma story pode existir em dois modos:

- **Standalone:** Trabalho que não precisa de um Epic. O problema é claro, o escopo é pequeno, e a skill Story fornece estrutura suficiente para defini-lo.

- **Derivada de Epic:** Um REC-N específico de um Epic está sendo decomposto em uma story executável. A skill Story lê o Epic pai e usa como contexto para perguntas mais focadas.

Em ambos os casos, o output é o mesmo: um Documento de Story que captura motivação, requisito, escopo e premissas.
