# Task Lifecycle — CLI Demo

Epic: `sistema-pagamentos` | Story: `checkout`

---

## (1) Registrar as 3 tasks com dependências

### create-order (sem dependências)

```
oraculo task init create-order \
  --epic sistema-pagamentos \
  --story checkout
```

**Saída:**
```json
{
  "name": "create-order",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "created": true
}
```

### process-payment (depende de create-order)

```
oraculo task init process-payment \
  --epic sistema-pagamentos \
  --story checkout \
  --depends-on create-order
```

**Saída:**
```json
{
  "name": "process-payment",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "created": true
}
```

### send-receipt (depende de process-payment)

```
oraculo task init send-receipt \
  --epic sistema-pagamentos \
  --story checkout \
  --depends-on process-payment
```

**Saída:**
```json
{
  "name": "send-receipt",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "created": true
}
```

---

## (2) Iniciar a task `create-order`

```
oraculo task start create-order \
  --epic sistema-pagamentos \
  --story checkout
```

**Saída:**
```json
{
  "id": 1,
  "story_id": 42,
  "name": "create-order",
  "description": "",
  "status": "in_progress",
  "failure_reason": "",
  "created_at": "2026-03-10T10:00:00Z",
  "updated_at": "2026-03-10T10:00:05Z",
  "started_at": "2026-03-10T10:00:05Z"
}
```

---

## (3) Completar `create-order`

O comando `task complete` lê o payload JSON via stdin. O JSON não inclui `logs` nem `skills_used`.

```
echo '{"summary":"Order created successfully","files_modified":["src/order.go"]}' | \
  oraculo task complete create-order \
    --epic sistema-pagamentos \
    --story checkout
```

**Saída:**
```json
{
  "id": 1,
  "story_id": 42,
  "name": "create-order",
  "description": "",
  "status": "completed",
  "failure_reason": "",
  "created_at": "2026-03-10T10:00:00Z",
  "updated_at": "2026-03-10T10:00:30Z",
  "started_at": "2026-03-10T10:00:05Z",
  "completed_at": "2026-03-10T10:00:30Z",
  "depends_on": [],
  "result": {
    "id": 1,
    "task_id": 1,
    "summary": "Order created successfully",
    "logs": "",
    "skills_used": null,
    "files_modified": ["src/order.go"],
    "created_at": "2026-03-10T10:00:30Z"
  }
}
```

---

## (4) Listar todas as tasks

```
oraculo task list \
  --epic sistema-pagamentos \
  --story checkout
```

**Saída:**
```json
[
  {
    "id": 1,
    "story_id": 42,
    "name": "create-order",
    "description": "",
    "status": "completed",
    "failure_reason": "",
    "created_at": "2026-03-10T10:00:00Z",
    "updated_at": "2026-03-10T10:00:30Z",
    "started_at": "2026-03-10T10:00:05Z",
    "completed_at": "2026-03-10T10:00:30Z",
    "depends_on": [],
    "result": {
      "id": 1,
      "task_id": 1,
      "summary": "Order created successfully",
      "logs": "",
      "skills_used": null,
      "files_modified": ["src/order.go"],
      "created_at": "2026-03-10T10:00:30Z"
    }
  },
  {
    "id": 2,
    "story_id": 42,
    "name": "process-payment",
    "description": "",
    "status": "pending",
    "failure_reason": "",
    "created_at": "2026-03-10T10:00:01Z",
    "updated_at": "2026-03-10T10:00:01Z",
    "depends_on": [1],
    "result": null
  },
  {
    "id": 3,
    "story_id": 42,
    "name": "send-receipt",
    "description": "",
    "status": "pending",
    "failure_reason": "",
    "created_at": "2026-03-10T10:00:02Z",
    "updated_at": "2026-03-10T10:00:02Z",
    "depends_on": [2],
    "result": null
  }
]
```

---

## Resumo do estado final

| Task             | Status      | depends_on       |
|------------------|-------------|------------------|
| create-order     | completed   | —                |
| process-payment  | pending     | create-order (1) |
| send-receipt     | pending     | process-payment (2) |

- `create-order` foi completada com summary `"Order created successfully"`, sem logs, sem skills_used, e com `files_modified: ["src/order.go"]`.
- `process-payment` e `send-receipt` permanecem em `pending`, aguardando suas dependências serem completadas antes de poderem ser iniciadas.
