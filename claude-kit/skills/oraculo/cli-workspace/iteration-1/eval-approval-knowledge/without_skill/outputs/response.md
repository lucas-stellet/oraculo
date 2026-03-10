# Aprovação e Knowledge — story `autenticacao` / epic `user-management`

## Passo 1 — Solicitar aprovação do tipo `design`

O comando `approval request` lê o conteúdo via stdin e exige as flags `--type` e `--epic`.

**Comando:**
```bash
echo '## Design\n\nUsaremos JWT com expiração de 24h' | \
  oraculo tools approval request \
    --type design \
    --epic user-management \
    --story autenticacao
```

**Resultado esperado:**
```json
{
  "id": "apr_01j9z3k2m4p7q8r5s6t9u0v1w2",
  "type": "design",
  "status": "pending",
  "epic_id": 1,
  "story_id": 2,
  "content": "## Design\n\nUsaremos JWT com expiração de 24h",
  "created_at": "2026-03-10T14:22:00Z",
  "updated_at": "2026-03-10T14:22:00Z"
}
```

---

## Passo 2 — Listar aprovações pendentes

**Comando:**
```bash
oraculo tools approval list --pending
```

**Resultado esperado:**
```json
[
  {
    "id": "apr_01j9z3k2m4p7q8r5s6t9u0v1w2",
    "type": "design",
    "status": "pending",
    "epic_id": 1,
    "story_id": 2,
    "content": "## Design\n\nUsaremos JWT com expiração de 24h",
    "created_at": "2026-03-10T14:22:00Z",
    "updated_at": "2026-03-10T14:22:00Z"
  }
]
```

A flag `--pending` filtra apenas registros com `status = "pending"`. Sem ela, o comando retorna todas as aprovações independente do status.

---

## Passo 3 — Persistir padrão no knowledge (memory store)

**Comando:**
```bash
oraculo tools memory store \
  --domain auth-service \
  --category pattern \
  --finding 'JWT tokens expiram em 24h; refresh tokens em cookies HttpOnly' \
  --source design.md \
  --confidence medium
```

**Resultado esperado:**
```json
{
  "id": "mem_01j9z3k2m4p7q8r5s6t9u0v2x3",
  "domain": "auth-service",
  "category": "pattern",
  "finding": "JWT tokens expiram em 24h; refresh tokens em cookies HttpOnly",
  "source": "design.md",
  "confidence": "medium",
  "created_at": "2026-03-10T14:23:15Z"
}
```

---

## Resumo dos comandos

| # | Comando | Finalidade |
|---|---------|------------|
| 1 | `echo '...' \| oraculo tools approval request --type design --epic user-management --story autenticacao` | Solicita aprovação de design com conteúdo via stdin |
| 2 | `oraculo tools approval list --pending` | Lista apenas aprovações aguardando decisão |
| 3 | `oraculo tools memory store --domain auth-service --category pattern --finding '...' --source design.md --confidence medium` | Persiste padrão no knowledge |

### Observações

- O conteúdo da aprovação é lido via **stdin** (`io.ReadAll`), não via flag — por isso é necessário usar `echo ... |` ou um heredoc.
- A flag `--story` é opcional em `approval request`; sem ela a aprovação fica vinculada apenas ao epic.
- As categorias válidas para `memory store` são: `pattern`, `convention`, `constraint`, `dependency`, `test`, `architecture`.
- Os níveis de confiança válidos são: `high`, `medium`, `low`.
