# Fase Story — Design do Sistema

## 1. Estrutura da Sessão

A fase Story segue um fluxo compacto: quatro fases que vão do entendimento do problema à geração do artefato. O Oraculo avança apenas quando a condição de saída de cada fase é satisfeita.

### 1.0 Setup da Sessão

O usuário chega com um item de trabalho. O Oraculo verifica se há um epic pai, captura a descrição crua, detecta o nível de reasoning e confirma que o trabalho é do tamanho de uma story.

**Workflow do epic pai:**
- Se um nome de projeto foi passado como argumento (`/oraculo:story gastos-app`):
  - Busca `.oraculo/epics/<nome-epic>/requirements.md`
  - Se encontra: verificar `approval_status = approved` — se não aprovado, bloquear a decomposição e informar o usuário que o epic pai ainda está pendente de revisão humana; se aprovado, ler o epic e perguntar qual REC-N trabalhar (ou se é algo novo)
  - Se não encontra: segue como standalone
- Se nenhum argumento: segue como standalone

**Verificação de escalação:** Se o trabalho é grande demais (múltiplas áreas, alta incerteza, escopo vago), sugere `/oraculo:epic` no lugar.

**Condição de saída:** Item de trabalho capturado, nível de reasoning determinado internamente, epic pai carregado e aprovação verificada (se aplicável), escopo de story confirmado.

### 1.1 Reenquadramento

O Oraculo separa o problema de qualquer solução proposta e define limites de escopo.

**Perguntas Deep:**
- "Você descreveu o que quer. Que problema isso resolve para o usuário?"
- "O que está explicitamente fora do escopo desta story?"

**Perguntas Light:**
- "Você recebeu essa tarefa. Qual é o problema que ela resolve?"
- "O que está explicitamente fora do escopo?"

**Perguntas derivadas de epic:**
- "O epic descreve [contexto do REC-N]. Qual é o problema específico que esta story endereça?"
- "Que partes do requisito NÃO são cobertas por esta story?"

**Condição de saída:** Problema claramente declarado e separado da solução. Limites de escopo definidos.

### 1.2 Verificação de Premissas

O Oraculo revela as premissas críticas por trás do trabalho.

**Deep (foco em valor):**
- "O que estamos assumindo sobre o usuário que, se estiver errado, torna esta story inútil?"
- "Qual é o maior risco — construir errado, ou construir a coisa errada?"

**Light (foco em viabilidade):**
- "O que estamos assumindo sobre o sistema atual que isso depende?"
- "Qual é o maior risco técnico aqui?"

**Condição de saída:** 2-3 premissas críticas identificadas e reconhecidas. Cada uma com nível de evidência (alto/médio/baixo).

### 1.3 Portão de Saída

Checkpoint final — avaliar quatro dimensões de risco antes de gerar o artefato.

**Quatro Riscos:**
1. **Valor** — Está claro por que alguém quer isso?
2. **Usabilidade** — Pode ser usado sem atrito?
3. **Viabilidade técnica** — É construível dentro do sistema atual?
4. **Viabilidade de negócio** — Está alinhado com objetivos / critérios são testáveis?

**Escalação Light:** Quando o dev não consegue responder perguntas de Valor ou Viabilidade de Negócio (requer contexto de produto), nota a lacuna no artefato da story em vez de voltar a fases anteriores.

**Condição de saída:** Todos os quatro riscos endereçados (aprovados ou explicitamente aceitos).

### 1.4 Geração de Artefatos

O Oraculo gera o Documento de Story e o salva via CLI, depois cria uma versão para revisão humana. O fluxo tem quatro etapas:

1. **Gerar e salvar** — O Oraculo gera o Documento de Story e o salva em `.oraculo/epics/<nome-epic>/stories/<nome-story>/requirements.md` via `oraculo story save`. Se derivada de um epic, o documento inclui referências ao pai nos comentários do cabeçalho.
2. **Criar versão** — O Oraculo chama `oraculo tools story version <nome-story> --epic <nome-epic>`, passando o markdown via stdin. Isso cria um snapshot versionado e o exibe no dashboard para revisão humana.
3. **Aguardar verdict** — O agente monitora reviews via `oraculo tools review list <version-id> --type story`. O fluxo não avança até que um humano emita um verdict pelo dashboard.
4. **Tratar o verdict:**
   - `approved` — A definição da story é finalizada. A story está pronta para execução.
   - `rejected` — Voltar a §1.1 (Reenquadramento) para reabrir o espaço do problema com base no feedback do revisor.

## 2. Saturação

A skill Story rastreia 5 slots para determinar quando informação suficiente foi coletada:

| Slot | Peso Deep | Peso Light |
|------|:---------:|:----------:|
| Declaração do problema | Standard | **High** |
| Resultado desejado | Standard | **High** |
| Critérios de aceite | Standard | Standard |
| Limites de escopo | Low | Standard |
| Premissas-chave | Standard | Standard |

**Regra de saturação:** Todos os slots High completos + todos os slots Standard ao menos parciais → apresenta o draft.

**Definição dos pesos:**
- **High** — Deve atingir **completo** para saturação. Requer perguntas dedicadas.
- **Standard** — Deve atingir ao menos **parcial** para saturação.
- **Low** — Não bloqueia saturação. Uma menção indireta conta como parcial.

## 3. Níveis de Reasoning

A fase Story opera em dois níveis de reasoning que compartilham a mesma disciplina socrática mas diferem no foco.

### 3.1 Light Reasoning (Executor/Desenvolvedor)

- Recebeu uma tarefa de outra pessoa
- Expertise técnica, contexto de produto limitado
- Foco: entender a tarefa corretamente, viabilidade técnica, edge cases
- Verificação de premissas foca em **viabilidade** (adequação à arquitetura, dependências)
- Escala para PM em lacunas de valor/negócio (não retorna a fases anteriores)

### 3.2 Deep Reasoning (Tomador de Decisão/PM)

- Explorando uma mudança que ele identificou como necessária
- Tem contexto de usuário, entendimento do negócio
- Foco: valor para o usuário, clareza do problema, resultado mensurável
- Verificação de premissas foca em **valor** (é o problema certo?)
- Responde todas as perguntas do gate autonomamente

### 3.3 Sinais de Seleção

| Sinal | Light | Deep |
|-------|-------|------|
| Usuário descreve tarefa recebida de outra pessoa | Sim | — |
| Usuário descreve ideia ou problema que ele identificou | — | Sim |
| Usuário tem métricas de negócio ou dados de usuários | — | Sim |
| Usuário foca na abordagem de implementação | Sim | — |
| Papel primariamente técnico | Sim | — |
| Papel envolve decisões de produto | — | Sim |

## 4. Workflow do Epic Pai

Quando uma story é derivada de um epic:

1. **Leitura:** A skill Story lê `.oraculo/epics/<nome-epic>/requirements.md`
2. **Seleção:** Pergunta ao usuário qual REC-N trabalhar, ou se é algo novo
3. **Herança de contexto:** A declaração de problema, contexto e escopo do epic informam a sessão da story — o usuário não repete o que já foi capturado
4. **Perguntas adaptadas:** Perguntas de reenquadramento referenciam o contexto do REC-N específico
5. **Referência no output:** O documento da story inclui metadados `parent_epic` e `parent_rec`

## 5. Output: Template de Story

O Documento de Story tem 4 seções:

1. **Motivação** — Problema e contexto (da perspectiva do usuário ou do sistema)
2. **Requisito** — Uma única story de usuário/sistema com critérios de aceite GIVEN/WHEN/THEN
3. **Escopo** — O que está dentro, o que está explicitamente fora
4. **Premissas** — Premissas-chave com níveis de evidência

O documento captura O QUÊ e POR QUÊ — nunca COMO. Sem detalhes técnicos de implementação.

## 6. Escalação

Se em qualquer momento durante uma sessão de Story o trabalho se revela do tamanho de um epic, o Oraculo sugere escalar:

**Sinais para escalação:**
- Múltiplas áreas ou disciplinas envolvidas
- Alta incerteza sobre a abordagem
- Escopo continua expandindo durante o reenquadramento
- Usuário não consegue articular uma declaração de problema focada

**Comportamento:** Sugere `/oraculo:epic` com explicação clara de por que a exploração completa seria mais apropriada. Respeita a decisão do usuário se quiser continuar como story.

## 7. Caminho de Saída

Artefatos de story são salvos em `.oraculo/epics/<nome-epic>/stories/<nome-story>/requirements.md` dentro da aplicação-alvo. O nome da story é derivado do título da story (minúsculas, hifenizado). Se nenhum nome de epic foi fornecido e a story é standalone, a CLI cria um epic leve automaticamente.

O artefato não é considerado final até que exista uma version review com verdict `approved`. Um documento de story salvo que ainda não recebeu um verdict humano (ou que recebeu `rejected`) não deve ser usado como input para planejamento de execução.
