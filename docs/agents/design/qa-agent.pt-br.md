# Agents Design — QA Agent

## 1. Papel

O QA agent e o sistema imunologico. Ele valida o output do code agent com olhos frescos, sem contexto compartilhado e sem incentivo para concordar. A sua independencia e a propriedade mais importante do pipeline de validacao.

Um QA agent que simplesmente confirma "esta bom" nao tem valor. Uma validacao eficaz exige um agent que verifique corretude funcional, imponha padroes, sonde casos de borda e avalie a qualidade dos testes — tudo isso sem acesso ao raciocinio do code agent sobre por que as coisas foram feitas de determinada forma.

## 1.1 Dois Niveis de Validacao

O QA opera em dois niveis distintos:

**Nivel de tarefa:** Cada tarefa concluida e revisada de forma independente. O QA agent recebe o diff, as specs e os resultados dos testes para aquela tarefa especifica. Os veredictos sao registrados com `task_id` na tabela `validations`.

**Nivel de story:** Assim que todas as tarefas de uma story passam no QA individual, uma revisao integrada final verifica a coerencia entre tarefas, os requisitos da story como um todo e a qualidade geral. Esse veredicto e registrado com `task_id = NULL` na tabela `validations` e serve como gate final antes de gerar o sumario markdown da story e extrair licoes aprendidas para a tabela de conhecimento.

## 2. Contexto Limpo

O QA agent opera com uma context window completamente limpa. Ele recebe exatamente tres inputs:

1. **O diff** — o que mudou, nada mais
2. **As specs** — a descricao da tarefa, criterios de aceitacao e requisitos relevantes da story/epic
3. **Resultados dos testes** — output da execucao da suite de testes existente

Ele **nao** recebe:
- O historico conversacional do code agent
- O raciocinio intermediario do code agent
- Veredictos anteriores de QA sobre essa tarefa (cada revisao de QA e independente)
- Logs de debug ou traces de compilacao do processo de implementacao

Esse isolamento quebra o ciclo de sycophancy. Quando um QA agent compartilha contexto com o code agent, ele tende a concordar com as decisoes do code agent — mesmo as equivocadas. O contexto limpo forca o QA agent a avaliar o trabalho pelo seu proprio merito.

## 3. O Que o QA Verifica

O QA agent avalia a implementacao em multiplas dimensoes:

| Dimensao | O que o QA agent pergunta |
|----------|--------------------------|
| **Corretude funcional** | A implementacao satisfaz os criterios de aceitacao? Ela lida com os inputs especificados e produz os outputs esperados? |
| **Conformidade com padroes** | O codigo segue as convencoes do projeto (do CLAUDE.md)? Nomenclatura, padroes, estrutura? |
| **Casos de borda** | As condicoes de contorno sao tratadas? O que acontece com inputs vazios, inputs grandes, acesso concorrente? |
| **Qualidade dos testes** | Os testes verificam de fato o comportamento? As assertivas sao significativas ou trivialmente verdadeiras? Os testes cobrem os criterios de aceitacao? |
| **Escopo** | O agent ficou dentro dos limites da tarefa? Ele modificou arquivos que nao deveria? Ele adicionou features nao solicitadas? |

## 4. QA Nunca Corrige

Esta e uma regra imutavel. O QA agent produz achados estruturados — ele nunca modifica codigo.

Quando o QA identifica um problema, ele reporta:
- **O que esta errado:** Descricao clara do problema
- **Onde:** Arquivo e localizacao
- **Por que importa:** Qual criterio de aceitacao ou padrao foi violado
- **Severidade:** Se e um blocker, um problema significativo ou uma preocupacao menor

O orquestrador recebe esses achados e instancia um novo code agent para trata-los. O trabalho do QA agent esta encerrado assim que ele entrega seu relatorio.

**Por que nao deixar o QA corrigir?** Se o QA corrige o codigo, ele nao pode mais validar a correcao de forma independente. O agent que escreve codigo e o agent que valida codigo precisam ser agentes diferentes — caso contrario, a validacao fica comprometida.

## 5. Fluxo de Rejeicao

Quando o QA rejeita uma tarefa:

```
Code Agent conclui a tarefa
    |
    v
QA Agent revisa (contexto limpo: diff + specs + resultados dos testes)
    |
    v
QA produz achados estruturados
    |
    v
Orquestrador recebe os achados
    |
    v
Orquestrador instancia NOVO Code Agent com:
  - Descricao original da tarefa
  - Achados do QA como contexto adicional
  - Contexto limpo (sem memoria da tentativa anterior)
    |
    v
Novo Code Agent implementa a correcao (skill TDD)
    |
    v
QA Agent revisa novamente (novo contexto limpo)
    |
    v
Repetir ate ser aprovado ou o circuit breaker ser acionado
```

Cada iteracao usa um **novo** code agent com contexto limpo. Isso evita comportamento defensivo, onde um agent tenta justificar seu trabalho anterior em vez de reconsiderar a partir dos primeiros principios.

## 6. Circuit Breaker

Um circuit breaker previne loops infinitos de rejeicao:

- **Limite:** N ciclos de QA com falha na mesma tarefa (configuravel, padrao 3)
- **Ao ser acionado:** O orquestrador para de instanciar novos code agents e escala para o humano
- **Contexto preservado:** Todos os achados de QA e resumos das tentativas ficam disponiveis para o humano revisar
- **Resolucao:** O humano pode fornecer orientacao, ajustar a tarefa ou assumir a implementacao

O circuit breaker existe porque algumas tarefas podem estar genuinamente alem da capacidade dos agents — requisitos ambiguos, restricoes conflitantes ou problemas que exigem julgamento humano. Escalar cedo e melhor do que queimar tokens em retries infinitos.

## 7. Carregamento de Skills

O QA agent pode receber skills do orquestrador, assim como os code agents:

- **playwright:** Para validacao E2E — o QA agent pode executar testes baseados em browser para verificar o comportamento da UI
- **Skills especificas do projeto:** Regras de validacao customizadas definidas pelo time

O workflow padrao de QA (revisao de diff + verificacao de padroes) nao requer uma skill especial. As skills ampliam as capacidades do QA para tipos especificos de tarefa.
