# Agents Design — Orquestrador

## 1. Papel

O orquestrador e o unico agent que enxerga o quadro completo. Ele planeja, delega e coordena — nunca executa. Nunca escreve codigo, nunca roda testes, nunca toca no filesystem. Sua context window e reservada exclusivamente para raciocinio estrategico.

Isso nao e uma diretriz — e uma fronteira arquitetural. Misturar planejamento e execucao no mesmo contexto degrada mensuravelmente os dois. O orquestrador opera no nivel de tarefas, dependencias e atribuicoes de agents. Detalhes taticos (erros de sintaxe, output de testes, logs de compilacao) sao tratados inteiramente pelos agents subordinados e pelo CLI.

## 2. Criacao do DAG

Durante a fase Plan, o orquestrador decompoe os requisitos em um DAG:

- **Tarefas sao nos.** Cada no tem um tipo (code, qa, research, documentation), uma descricao e uma atribuicao de skill.
- **Dependencias sao arestas.** Uma aresta de A para B significa que B nao pode iniciar ate A concluir.
- **O grafo e aciclico.** O CLI valida isso antes de persistir.

O orquestrador emite o DAG como um plano estruturado. O CLI o valida (aciclicidade, integridade das dependencias, sem auto-referencias) e o persiste no SQLite usando o schema definido em [`docs/cli/design.md`](../../cli/design.md) §4.3.

O DAG nao e um artifact estatico. Quando um agent encontra um obstaculo ou o QA rejeita uma tarefa, o orquestrador muta o DAG — podando ramos invalidos, adicionando novas tarefas, ajustando dependencias — sem regenerar o plano inteiro. O CLI valida toda mutacao antes de commitar.

## 3. Despacho

O orquestrador avalia o DAG e despacha tarefas que estao prontas para execucao:

**Regra de prontidao:** Uma tarefa esta pronta quando todas as suas dependencias estao em estado terminal (concluidas ou falhas-e-tratadas).

**Paralelismo:** Todas as tarefas prontas sao despachadas simultaneamente quando possivel. O orquestrador maximiza a execucao paralela dentro das restricoes do DAG e da coordenacao de arquivos.

**Coordenacao de arquivos:** Como todos os agents trabalham no mesmo branch no mesmo diretorio, o orquestrador deve garantir que dois agents nunca modifiquem os mesmos arquivos simultaneamente. Isso e obtido por meio da estrutura de dependencias do DAG — tarefas que tocam em arquivos sobrepostos sao conectadas por arestas, forcando execucao sequencial. Quando o orquestrador cria o DAG, ele codifica essas restricoes de nivel de arquivo como dependencias.

**Fallback sequencial:** Quando o paralelismo verdadeiro nao e possivel (arquivos sobrepostos, estado compartilhado), as tarefas rodam sequencialmente. O orquestrador prefere execucao paralela, mas nunca a custo da corretude.

## 4. Atribuicao de Skills

O orquestrador seleciona skills para cada agent com base na natureza da tarefa. Essa e uma responsabilidade de primeira classe — a skill certa determina o workflow inteiro do agent.

| Tipo de Tarefa | Skills Tipicas | Justificativa |
|----------------|---------------|---------------|
| Implementacao de codigo | TDD | Red-green-refactor garante corretude |
| Trabalho de UI/frontend | TDD + frontend-design | Qualidade de design + disciplina de testes |
| Validacao E2E | playwright | Testes baseados em browser |
| Correcao de bug | TDD + systematic-debugging | Diagnostico antes da implementacao |
| Refatoracao | TDD | Testes existentes como rede de seguranca |

A skill e o mecanismo de contencao. Ela define:
- **Workflow:** Quais etapas o agent deve seguir (ex.: red-green-refactor)
- **Restricoes:** O que o agent nao deve fazer (ex.: pular a fase red)
- **Quality gates:** O que deve ser verdadeiro antes de o agent reportar conclusao

O orquestrador nao inventa workflows. Ele seleciona da biblioteca de skills disponivel com base nas caracteristicas da tarefa.

## 5. Throughput de QA como Restricao Governante

O orquestrador trata a capacidade de QA como o gargalo do sistema — a restricao que define o ritmo para o sistema inteiro (Theory of Constraints).

**O problema:** Se o orquestrador despacha tarefas de codigo na velocidade maxima que o DAG permite, trabalho nao validado se acumula. Trabalho nao validado carrega risco: context drift, suposicoes em cascata, retrabalho exponencial quando o QA eventualmente encontra problemas.

**A solucao:** O orquestrador limita a taxa de despacho para que codigo seja produzido apenas na velocidade em que o QA consegue validar. Quando o QA agent esta revisando a tarefa N, o orquestrador pode segurar a tarefa N+2 mesmo que suas dependencias estejam satisfeitas, para evitar uma fila crescente de trabalho nao validado.

**Heuristica pratica:** Manter no maximo uma tarefa nao validada a frente do QA em qualquer momento. Se o QA estiver ocupado, o orquestrador pode despachar tarefas que nao exigem validacao de QA (documentacao, research).

## 6. Tratamento de Falhas

Quando um code agent falha ou o QA rejeita uma tarefa:

1. **Rejeicao do QA:** O orquestrador instancia um **novo** code agent com os achados estruturados do QA. O novo agent recebe um contexto limpo — sem memoria da tentativa anterior, sem suposicoes herdadas. Isso evita que agents defendam erros anteriores.

2. **Falha do agent:** Se um agent falha em concluir sua tarefa (erro, timeout, tarefa impossivel), o orquestrador marca a tarefa como falha com uma razao e avalia se deve tentar novamente com contexto ajustado ou escalar.

3. **Circuit breaker:** Apos N ciclos de QA com falha na mesma tarefa (configuravel, padrao 3), o orquestrador para de tentar e submete uma approval request `qa-escalation` ao dashboard:

   ```
   oraculo tools approval request --type qa-escalation
   ```

   O orquestrador transiciona para o estado `awaiting_approval`. O dashboard exibe o contexto de diagnostico completo (todos os achados do QA, todos os sumarios de tentativas) para revisao humana. O veredicto do humano define o proximo passo:
   - **`approved`** — Aceitar a implementacao atual como esta e avançar a tarefa
   - **`rejected`** — Abortar a story; marcar a tarefa como permanentemente falha
   - **`needs_revision`** — Instanciar um novo code agent com os comentarios do humano como contexto adicional e resetar o circuit breaker

4. **Mutacao do DAG:** Quando uma falha invalida tarefas downstream, o orquestrador poda o ramo afetado e pode adicionar novas tarefas para seguir uma abordagem diferente. O CLI valida toda mutacao.

## 7. Estado Awaiting Approval

Quando um approval gate esta aberto, o orquestrador entra no estado `awaiting_approval`. Esse estado tem regras comportamentais especificas para manter o sistema produtivo sem avançar uma fase bloqueada:

**O que o orquestrador nao faz:**
- Nao despacha tarefas que dependam — direta ou transitivamente — da aprovacao pendente
- Nao avanca a fase que o approval gate esta protegendo

**O que o orquestrador continua fazendo:**
- Despacha tarefas independentes cujas dependencias ja estao satisfeitas e que nao dependem do resultado da aprovacao
- Monitora o status da aprovacao por polling: `oraculo tools approval status`

**Retorno ao estado ativo:**
Uma vez recebido um veredicto via `oraculo tools approval status`, o orquestrador le o veredicto, aplica o resultado adequado (avancar, abortar ou tentar novamente com comentarios) e transiciona de volta ao estado ativo. As tarefas dependentes sao reavaliadas em relacao ao DAG atualizado.
