# Eval: Approval + Knowledge — with_skill

## Tarefa

O design da story `autenticacao` no epic `user-management` está pronto.

1. Solicitar uma aprovação do tipo `design` com conteúdo `## Design\n\nUsaremos JWT com expiração de 24h`
2. Listar as aprovações pendentes
3. Persistir no knowledge um padrão com os parâmetros fornecidos

---

## Passo 1 — Solicitar aprovação de design

**Comando:**

```bash
printf '## Design\n\nUsaremos JWT com expiração de 24h' | oraculo tools approval request \
  --type design \
  --epic user-management \
  --story autenticacao
```

**Saída esperada:**

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "type": "design",
  "status": "pending",
  "requested_at": "2026-03-10T14:00:00Z"
}
```

---

## Passo 2 — Listar aprovações pendentes

**Comando:**

```bash
oraculo tools approval list --pending
```

**Saída esperada:**

```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "type": "design",
    "epic": "user-management",
    "story": "autenticacao",
    "status": "pending",
    "requested_at": "2026-03-10T14:00:00Z"
  }
]
```

---

## Passo 3 — Persistir padrão no knowledge

**Comando:**

```bash
oraculo tools memory store \
  --domain "auth-service" \
  --category pattern \
  --finding "JWT tokens expiram em 24h; refresh tokens em cookies HttpOnly" \
  --source "design.md" \
  --confidence medium
```

**Saída esperada:**

```json
{
  "id": 1,
  "domain": "auth-service",
  "category": "pattern",
  "finding": "JWT tokens expiram em 24h; refresh tokens em cookies HttpOnly",
  "source": "design.md",
  "confidence": "medium",
  "created_at": "2026-03-10T14:00:01Z"
}
```

---

## Resumo

| Passo | Comando principal | Resultado |
|-------|-------------------|-----------|
| 1 | `approval request --type design --epic user-management --story autenticacao` | Aprovação criada com status `pending` |
| 2 | `approval list --pending` | Lista com a aprovação recém-criada |
| 3 | `memory store --domain auth-service --category pattern ...` | Padrão JWT persistido no knowledge |

### Observações

- O conteúdo da aprovação é passado via **stdin** com `printf` para preservar as quebras de linha (`\n`) literais.
- A flag `--story` é opcional em `approval request`, mas deve ser usada para aprovações com escopo de story.
- O `memory store` não exige `--epic` ou `--story` — o knowledge é global e indexado por `domain`.
- A CLI retorna JSON em todos os comandos `tools`; erros seguem o formato `{"error": "...", "details": {...}}`.
