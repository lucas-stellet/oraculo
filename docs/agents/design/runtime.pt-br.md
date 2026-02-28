# Agents Design — Runtime

## 1. SQLite — Operational State + Knowledge

Um unico banco de dados SQLite em `.oraculo/oraculo.db` serve todo o projeto. Ele esta listado no `.gitignore` e nao e commitado no repositorio. O banco de dados armazena duas categorias de dados: estado operacional transiente (tasks, dependencias, QA verdicts) que pode ser limpo apos a conclusao do epic, e conhecimento persistente (licoes aprendidas, padroes do codebase) que se acumula ao longo de todos os epics.

O banco de dados usa o schema definido em [`docs/cli/design.md`](../../cli/design.md) §4.3:

- **`epics`** — Metadados do epic (name, description, timestamps)
- **`stories`** — Stories pertencentes a epics
- **`tasks`** — Ciclo de vida do status das tasks (`pending → in_progress → completed | failed`)
- **`task_dependencies`** — Arestas do DAG (qual task depende de qual)
- **`task_results`** — Dados ricos de conclusao (summary, logs, skills usadas, arquivos modificados)
- **`validations`** — QA verdicts por task e por story (approved/rejected)
- **`approvals`** — Registros de approval gate (type, status, artifact snapshot, verdict, comments) — veja [`docs/cli/design.md`](../../cli/design.md) §4.3 para o schema
- **`knowledge`** — Conhecimento do codebase com full-text search

### O Que e Rastreado

| Dado | Tabela | Proposito |
|------|--------|-----------|
| Status da task | `tasks` | O orquestrador consulta isso para determinar o que despachar a seguir |
| Dependencias | `task_dependencies` | A estrutura do DAG — quais tasks bloqueiam quais |
| Dados de conclusao | `task_results` | Summary, logs, arquivos modificados — usados para contexto do QA e geracao de markdown |
| QA verdicts | `validations` | Se uma task ou story passou ou falhou na validacao |
| Conhecimento do codebase | `knowledge` | Padroes, convencoes, restricoes descobertas durante a execucao |

### Ciclo de Vida

O banco de dados e criado automaticamente pelo CLI no primeiro comando `oraculo tools`. Ele persiste durante toda a vida do projeto.

**Dados transientes** (tasks, dependencias, QA verdicts, task results, approvals) sao valiosos durante a execucao, mas nao apos. Uma vez que o epic e concluido e seus artifacts markdown sao gerados, esses dados operacionais podem ser limpos. Os registros de approval seguem o mesmo ciclo de vida — sao estado operacional, nao historico de longo prazo.

**Dados persistentes** (tabela knowledge) se acumulam ao longo de todos os epics. Quando um epic/story e concluido, as licoes aprendidas sao extraidas para a tabela knowledge. Esses dados nunca sao limpos — sao a memoria de longo prazo do projeto.

O banco de dados nao e commitado (`.gitignore`) porque contem estado operacional que poluiria o repositorio. O output significativo — requisitos, decisoes, summaries e conhecimento — e capturado em arquivos markdown commitados e na tabela knowledge persistente.

## 2. Single Markdown Artifact

Quando uma story e concluida e validada pelo QA, o sistema gera um unico arquivo markdown resumindo a implementacao. Esse e o unico artifact commitado gerado pelo processo de execucao.

O markdown contem:
- O que foi implementado (summary da story)
- Decisoes-chave tomadas durante a implementacao
- Arquivos criados ou modificados
- Resultado do QA

Esse arquivo reside na estrutura de diretorios do epic:

```
.oraculo/
└── epics/
    └── <epic-name>/
        ├── requirements.md          # Epic requirements (output of /oraculo:epic)
        └── stories/
            └── <story-name>/
                └── requirements.md  # Story requirements (output of /oraculo:story)
```

**Um arquivo por story, nao por task.** Os resultados individuais de tasks sao granulares demais para leitura humana. O summary em nivel de story captura o que importa: o que foi feito, por que e o que o QA agent verificou.

## 3. Memory Model

A memoria do Oraculo e deliberadamente simples. Tres fontes, tres propositos:

### CLAUDE.md — Project Context

CLAUDE.md e o contexto persistente para todos os agents. Ele contem:
- Estrutura e convencoes do projeto
- Decisoes arquiteturais
- Estilo de codigo e padroes
- Estrategias de teste
- Dependencias e restricoes

O orquestrador alimenta o conteudo do CLAUDE.md para cada agent que instancia. Isso garante que todos os agents sigam as mesmas convencoes sem precisar de um sistema complexo de recuperacao de conhecimento.

O CLAUDE.md e atualizado pelo time de desenvolvimento (humano ou assistido por agents) quando as convencoes mudam. E a unica fonte de verdade para "como as coisas sao feitas neste projeto."

### Epic Markdowns — Feature History

Documentos de requisitos, definicoes de stories e summaries de tasks capturam a inteligencia acumulada de cada feature:

- **Epic requirements** — A definicao de problema validada, criterios de aceitacao, casos de borda
- **Story requirements** — Escopo e restricoes especificos de implementacao
- **Completion summaries** — O que foi construido, decisoes tomadas, resultados do QA

Esses arquivos sao commitados no repositorio. Servem como registro historico e como contexto para trabalhos futuros na mesma area.

### SQLite — Operational State + Knowledge

Dados transientes (status de tasks, dependencias, QA verdicts) vivem no SQLite durante a duracao do epic — essenciais durante a execucao, limpaveis apos. A tabela knowledge persiste ao longo de todos os epics, acumulando licoes aprendidas. Essa e a memoria de trabalho do sistema mais a sua memoria de longo prazo.

### O Que Isso Nao E

Nao ha uma arquitetura de memoria de tres camadas (working/episodic/semantic). Nao ha pipeline de curadoria ou scoring de promocao. A tabela knowledge com full-text search oferece um repositorio simples e consultavel para descobertas do codebase — nao o sistema de memoria rico e deferido descrito em future-work.md. O modelo — CLAUDE.md + markdowns + SQLite (operational state + knowledge) — cobre as necessidades essenciais:

- Agents conhecem as convencoes do projeto (CLAUDE.md)
- Agents conhecem os requisitos da feature (epic markdowns)
- O sistema rastreia o que esta feito e o que esta pendente (SQLite)

Se a experiencia revelar que um sistema de memoria mais rico e necessario, ele pode ser adicionado depois. Veja [`future-work.md`](future-work.md) para capacidades de memoria deferidas.
