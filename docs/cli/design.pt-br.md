# Oraculo CLI — Design

## 1. Arquitetura de Comandos

O CLI possui duas camadas de comandos, separadas por audiência e propósito:

| Camada | Audiência | Saída | Propósito |
|--------|-----------|-------|-----------|
| **Root** (`oraculo <cmd>`) | Humanos | Tabelas/texto formatado | Ciclo de vida, dashboard |
| **Tools** (`oraculo tools <domain> <action>`) | Agentes | JSON | Operações determinísticas |

Ambas as camadas compartilham pacotes internos em Go, mas possuem handlers de comando e formatação de saída independentes. Comandos root priorizam legibilidade. Comandos tools priorizam parseabilidade.

## 2. Comandos Root (Para humanos)

```
oraculo install [--global|--local]    # Instala skills, hooks, settings no Claude Code
oraculo update                        # Atualiza o binário do CLI para a última versão
oraculo status                        # Dashboard: epics, stories, tasks, progresso
oraculo version                       # Versão do CLI
```

**`install`** — Configura o Claude Code para funcionar como Oraculo. Dois modos:

- `--global`: copia skills para `~/.claude/skills/`, hooks para `~/.claude/hooks/`, modifica settings globais. Todos os projetos ganham acesso ao Oraculo.
- `--local`: o mesmo, mas dentro de `.claude/` do projeto atual. Apenas aquele projeto recebe o Oraculo.

**`update`** — Atualiza o binário do CLI do Oraculo para a última versão. Distinto de `install` — `install` gerencia a configuração do Claude Code (skills, hooks, settings), `update` gerencia o binário.

**`status`** — Dashboard legível mostrando epics, contagem de stories, progresso de tasks (pending/in_progress/completed/failed), percentuais e contagem de aprovações pendentes.

**Sem comando `init`.** Auto-bootstrapping: o primeiro comando `tools` que toca o armazenamento cria `.oraculo/` e o banco de dados automaticamente.

## 3. Comandos Tools (Para agentes)

Convenção: o nome da entidade é sempre o primeiro argumento posicional. Referências ao pai usam flags (`--epic`, `--story`).

Todos os comandos `init` são idempotentes — chamá-los duas vezes retorna `"created": false` em vez de falhar. Comandos de transição de status (`start`, `complete`, `fail`) aplicam o ciclo de vida e rejeitam transições inválidas com um erro estruturado.

### 3.1 Epic

```bash
oraculo tools epic init <name> [--description "..."]
oraculo tools epic save <name>                          # stdin: markdown
oraculo tools epic get <name>                           # stdout: raw markdown
oraculo tools epic list
oraculo tools epic update <name> [--description "..."]
oraculo tools epic delete <name>
```

### 3.2 Story

```bash
oraculo tools story init <name> --epic <epic>
oraculo tools story save <name> --epic <epic>            # stdin: markdown
oraculo tools story get <name> --epic <epic>             # stdout: raw markdown
oraculo tools story list --epic <epic>
oraculo tools story update <name> --epic <epic> [--description "..."]
oraculo tools story delete <name> --epic <epic>
```

### 3.3 Task

```bash
oraculo tools task init <name> --epic <epic> --story <story> [--description "..."] [--depends-on <task>]
oraculo tools task start <name> --epic <epic> --story <story>
oraculo tools task complete <name> --epic <epic> --story <story>    # stdin: JSON
oraculo tools task fail <name> --epic <epic> --story <story> --reason "..."
oraculo tools task get <name> --epic <epic> --story <story>
oraculo tools task list --epic <epic> --story <story>
oraculo tools task delete <name> --epic <epic> --story <story>
```

**Ciclo de vida da task:** `pending → in_progress → completed | failed`

**`task init`** — Os IDs das tasks vêm do documento da story. Quando o usuário aprova uma story, o agente cria tasks baseadas nos IDs definidos lá. `--depends-on` constrói o DAG.

**Formato stdin de `task complete`:**

```json
{
  "summary": "Implemented login form with validation",
  "logs": "Created src/components/LoginForm.tsx...",
  "skills_used": ["test-driven-development", "frontend-design"],
  "files_modified": ["src/components/LoginForm.tsx", "src/tests/login.test.tsx"]
}
```

**`task fail`** — `--reason` é obrigatório. O contrato recusa uma falha sem justificativa.

### 3.4 Memory

```bash
oraculo tools memory store --domain <d> --category <c> --finding "..." [--source "..."] [--confidence high|medium|low]
oraculo tools memory search <query> [--domain <d>] [--limit N]
oraculo tools memory domains
```

Categorias: `pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`.

### 3.5 Approval

```bash
oraculo tools approval request --type <type> --epic <epic> [--story <story>]   # stdin: markdown
oraculo tools approval status <id>
oraculo tools approval list [--pending]
oraculo tools approval verdict <id> --verdict <approved|rejected|needs_revision> [--comment "..."]
```

`--type` aceita: `epic-requirements`, `story-definition`, `qa-escalation`, `execution-plan`.

## 4. Modelo de Dados

### 4.1 Divisão de Responsabilidade

| Tipo de dado | Onde | Por quê |
|--------------|------|---------|
| Estrutural (DAG, status, vereditos) | SQLite | Determinístico, consultável |
| Conteúdo (análise, decisões, raciocínio) | Markdown | Criativo, narrativo |
| Ponte | `save` + `memory` | CLI indexa, não substitui |

### 4.2 Sistema de Arquivos

```
.oraculo/
├── oraculo.db                              # SQLite — infraestrutura, .gitignore
└── epics/
    └── gastos-control/
        ├── requirements.md                 # Requisitos do epic (saída de /oraculo:epic)
        └── stories/
            └── login-flow/
                ├── requirements.md         # Requisitos da story (saída de /oraculo:story)
                └── tasks/                  # Resultados de tasks armazenados no DB, não em arquivos
```

Regras:

- Stories sempre pertencem a um epic. O CLI aplica isso. Quando uma story é criada sem um epic explícito, o CLI auto-cria um **epic leve** — um epic mínimo com o nome derivado da story e sem markdown de requisitos. Isso preserva o modelo de dados hierárquico enquanto mantém a UX de story standalone sem atrito.
- O `requirements.md` de cada nível é a definição de produto — O QUE e POR QUE, nunca COMO.
- O banco de dados SQLite é infraestrutura — `.gitignore`.
- Arquivos Markdown são versionáveis no git.

### 4.3 Schema SQLite

```sql
-- Entidades centrais
CREATE TABLE epics (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT UNIQUE NOT NULL,
    description     TEXT DEFAULT '',
    approval_status TEXT DEFAULT 'none'
                    CHECK (approval_status IN ('none','pending','approved','rejected','needs_revision')),
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now'))
);

CREATE TABLE stories (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    epic_id         INTEGER NOT NULL REFERENCES epics(id),
    name            TEXT NOT NULL,
    description     TEXT DEFAULT '',
    approval_status TEXT DEFAULT 'none'
                    CHECK (approval_status IN ('none','pending','approved','rejected','needs_revision')),
    created_at      TEXT DEFAULT (datetime('now')),
    updated_at      TEXT DEFAULT (datetime('now')),
    UNIQUE(epic_id, name)
);

CREATE TABLE tasks (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id       INTEGER NOT NULL REFERENCES stories(id),
    name           TEXT NOT NULL,
    description    TEXT DEFAULT '',
    status         TEXT DEFAULT 'pending'
                   CHECK (status IN ('pending','in_progress','completed','failed')),
    failure_reason TEXT DEFAULT '',
    created_at     TEXT DEFAULT (datetime('now')),
    updated_at     TEXT DEFAULT (datetime('now')),
    started_at     TEXT,
    completed_at   TEXT,
    UNIQUE(story_id, name)
);

-- DAG: dependências entre tasks
CREATE TABLE task_dependencies (
    task_id    INTEGER NOT NULL REFERENCES tasks(id),
    depends_on INTEGER NOT NULL REFERENCES tasks(id),
    PRIMARY KEY (task_id, depends_on),
    CHECK (task_id != depends_on)
);

-- Dados ricos de conclusão (1:1 com task, apenas ao completar)
CREATE TABLE task_results (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id        INTEGER UNIQUE NOT NULL REFERENCES tasks(id),
    summary        TEXT NOT NULL,
    logs           TEXT DEFAULT '',
    skills_used    TEXT DEFAULT '',     -- JSON array armazenado como texto
    files_modified TEXT DEFAULT '',     -- JSON array armazenado como texto
    created_at     TEXT DEFAULT (datetime('now'))
);

-- Vereditos de validação QA (parte estrutural; análise fica em markdown)
CREATE TABLE validations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id    INTEGER NOT NULL REFERENCES stories(id),
    task_id     INTEGER REFERENCES tasks(id),              -- NULL = validação em nível de story
    verdict     TEXT NOT NULL CHECK (verdict IN ('approved','rejected')),
    created_at  TEXT DEFAULT (datetime('now'))
);

-- Conhecimento da codebase
CREATE TABLE knowledge (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    domain       TEXT NOT NULL,
    category     TEXT NOT NULL
                 CHECK (category IN ('pattern','convention','constraint','dependency','test','architecture')),
    finding      TEXT NOT NULL,
    source_files TEXT DEFAULT '',
    confidence   TEXT DEFAULT 'medium'
                 CHECK (confidence IN ('high','medium','low')),
    created_at   TEXT DEFAULT (datetime('now'))
);

-- Índice de busca full-text
CREATE VIRTUAL TABLE knowledge_fts USING fts5(
    domain, category, finding, source_files,
    content=knowledge, content_rowid=id
);

CREATE TRIGGER knowledge_ins AFTER INSERT ON knowledge BEGIN
    INSERT INTO knowledge_fts(rowid, domain, category, finding, source_files)
    VALUES (new.id, new.domain, new.category, new.finding, new.source_files);
END;

CREATE TRIGGER knowledge_del AFTER DELETE ON knowledge BEGIN
    INSERT INTO knowledge_fts(knowledge_fts, rowid, domain, category, finding, source_files)
    VALUES ('delete', old.id, old.domain, old.category, old.finding, old.source_files);
END;

-- Approval gates HITL
CREATE TABLE approvals (
    id               TEXT PRIMARY KEY,
    type             TEXT NOT NULL CHECK (type IN ('epic-requirements','story-definition',
                                                   'qa-escalation','execution-plan')),
    epic_id          INTEGER REFERENCES epics(id),
    story_id         INTEGER REFERENCES stories(id),
    content          TEXT NOT NULL,
    previous_version TEXT DEFAULT '',
    status           TEXT DEFAULT 'pending'
                     CHECK (status IN ('pending','approved','rejected','needs_revision')),
    verdict_comment  TEXT DEFAULT '',
    requested_at     TEXT DEFAULT (datetime('now')),
    decided_at       TEXT
);
```

Decisões de design:

- `epics` e `stories` rastreiam apenas metadados. O conteúdo vive em arquivos Markdown.
- `tasks` rastreiam ciclo de vida de status e metadados. Dados ricos de conclusão em `task_results`.
- `task_dependencies` modela o DAG. O CLI pode validar que não existem ciclos.
- `validations` armazena vereditos estruturais em dois níveis: por task (`task_id` preenchido) durante a execução e por story (`task_id` NULL) como gate final. A análise de QA vive em Markdown.
- `knowledge` usa FTS5 para busca full-text. Triggers de sincronização mantêm o índice consistente.
- Arrays JSON (`skills_used`, `files_modified`) armazenados como TEXT — simples, consultável com `json_each()`.
- `approvals` usa uma chave primária TEXT (UUID) para que os IDs sejam estáveis e compartilháveis entre fronteiras de agentes. `epic_id` e `story_id` podem ser NULL para tipos transversais como `execution-plan`. `previous_version` armazena o conteúdo anterior quando uma revisão é solicitada, permitindo exibição de diff na UI.
- `approval_status` em `epics` e `stories` espelha o resultado mais recente do approval gate para aquela entidade, evitando joins no caminho comum de verificação de status.

## 5. Formato de Saída

**Comandos tools:** JSON no stdout, exit code 0 para sucesso, exit code 1 para erro.

Exemplo de sucesso:

```json
{
  "name": "gastos-control",
  "path": ".oraculo/epics/gastos-control",
  "created": true
}
```

Exemplo de erro:

```json
{
  "error": "epic_not_found",
  "message": "No epic found with name 'nonexistent'"
}
```

**Recuperação de conteúdo** (comandos `get`): Markdown bruto no stdout, exit code 0. Erros ainda retornam JSON com exit code 1.

**Comandos root:** saída formatada legível (tabelas, barras de progresso, texto alinhado). Não destinada a parsing programático.

## 6. Estrutura do Projeto Go

```
cmd/oraculo/
├── main.go
├── cmd/
│   ├── root.go                 # Comando root Cobra
│   ├── install.go              # install --global|--local
│   ├── status.go               # Dashboard humano
│   ├── version.go              # Informação de versão
│   └── tools/
│       ├── tools.go            # Comando pai tools + middleware de auto-bootstrap
│       ├── epic.go             # tools epic init|save|get|list|update|delete
│       ├── story.go            # tools story init|save|get|list|update|delete
│       ├── task.go             # tools task init|start|complete|fail|get|list|delete
│       ├── memory.go           # tools memory store|search|domains
│       └── approval.go         # tools approval request|status|list|verdict
├── internal/
│   ├── db/
│   │   ├── db.go               # Open, auto-criar .oraculo/, migrar
│   │   └── migrations.go       # Versionamento de schema
│   ├── epic/
│   │   └── epic.go             # Lógica de negócio de epic
│   ├── story/
│   │   └── story.go            # Lógica de negócio de story
│   ├── task/
│   │   └── task.go             # Lógica de negócio de task + validação de DAG
│   ├── memory/
│   │   └── memory.go           # Knowledge store/search
│   ├── approval/
│   │   └── approval.go         # Lógica de request/verdict/query de approval
│   ├── installer/
│   │   └── installer.go        # Lógica de install (copiar skills, hooks, settings)
│   └── output/
│       ├── json.go             # Helpers de resposta JSON (tools)
│       └── table.go            # Formatação legível (root)
├── go.mod
└── go.sum
```

**Dependências:**

- `github.com/spf13/cobra` — Framework CLI
- `modernc.org/sqlite` — SQLite em Go puro (sem CGO)

**Notas de arquitetura:**

- `cmd/tools/tools.go` contém um `PersistentPreRun` do Cobra que lida com auto-bootstrapping (criar `.oraculo/` e DB se ausentes) para todos os subcomandos tools.
- Pacotes em `internal/` contêm lógica de negócio pura, sem preocupações de CLI. Tanto os comandos `tools` quanto o comando `status` usam os mesmos pacotes internos.
- Validação de pré-condições (epic existe, story pertence ao epic, transição de status válida) vive nos pacotes `internal/`, não nos handlers de comando.
- Handlers de comando são finos: parse de flags → chamar lógica interna → formatar saída.
