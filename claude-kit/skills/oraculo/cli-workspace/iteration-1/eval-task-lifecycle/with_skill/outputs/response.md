# Task Lifecycle — sistema-pagamentos / checkout

## Objetivo

Registrar 3 tasks no story `checkout` do epic `sistema-pagamentos`, iniciar e completar `create-order`, e listar o estado final.

---

## (1) Registrar as 3 tasks com suas dependências

### Task 1: `create-order` (sem dependências)

```bash
oraculo tools task init create-order \
  --epic sistema-pagamentos \
  --story checkout \
  --description "Create a new order record in the database"
```

**Saída esperada:**
```json
{
  "name": "create-order",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "status": "pending",
  "depends_on": [],
  "created": true
}
```

---

### Task 2: `process-payment` (depende de `create-order`)

```bash
oraculo tools task init process-payment \
  --epic sistema-pagamentos \
  --story checkout \
  --description "Process the payment for the created order" \
  --depends-on create-order
```

**Saída esperada:**
```json
{
  "name": "process-payment",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "status": "pending",
  "depends_on": ["create-order"],
  "created": true
}
```

---

### Task 3: `send-receipt` (depende de `process-payment`)

```bash
oraculo tools task init send-receipt \
  --epic sistema-pagamentos \
  --story checkout \
  --description "Send a receipt to the customer after successful payment" \
  --depends-on process-payment
```

**Saída esperada:**
```json
{
  "name": "send-receipt",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "status": "pending",
  "depends_on": ["process-payment"],
  "created": true
}
```

---

## (2) Iniciar a task `create-order`

```bash
oraculo tools task start create-order \
  --epic sistema-pagamentos \
  --story checkout
```

**Saída esperada:**
```json
{
  "name": "create-order",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "status": "in_progress",
  "started_at": "2026-03-10T14:00:00Z"
}
```

---

## (3) Completar a task `create-order`

```bash
echo '{
  "summary": "Order created successfully",
  "files_modified": ["src/order.go"]
}' | oraculo tools task complete create-order \
  --epic sistema-pagamentos \
  --story checkout
```

**Saída esperada:**
```json
{
  "name": "create-order",
  "epic": "sistema-pagamentos",
  "story": "checkout",
  "status": "completed",
  "summary": "Order created successfully",
  "files_modified": ["src/order.go"],
  "completed_at": "2026-03-10T14:05:00Z"
}
```

---

## (4) Listar todas as tasks para ver o estado atual

```bash
oraculo tools task list \
  --epic sistema-pagamentos \
  --story checkout
```

**Saída esperada:**
```json
[
  {
    "name": "create-order",
    "description": "Create a new order record in the database",
    "status": "completed",
    "depends_on": [],
    "summary": "Order created successfully",
    "files_modified": ["src/order.go"],
    "started_at": "2026-03-10T14:00:00Z",
    "completed_at": "2026-03-10T14:05:00Z"
  },
  {
    "name": "process-payment",
    "description": "Process the payment for the created order",
    "status": "pending",
    "depends_on": ["create-order"],
    "summary": null,
    "files_modified": [],
    "started_at": null,
    "completed_at": null
  },
  {
    "name": "send-receipt",
    "description": "Send a receipt to the customer after successful payment",
    "status": "pending",
    "depends_on": ["process-payment"],
    "summary": null,
    "files_modified": [],
    "started_at": null,
    "completed_at": null
  }
]
```

---

## Resumo

| Task | Status | Dependências |
|---|---|---|
| `create-order` | `completed` | — |
| `process-payment` | `pending` | `create-order` |
| `send-receipt` | `pending` | `process-payment` |

A task `create-order` foi completada com sucesso com o sumário `"Order created successfully"` e o arquivo `src/order.go` registrado como modificado. As tasks `process-payment` e `send-receipt` permanecem `pending`, aguardando suas respectivas dependências serem completadas antes de poderem ser iniciadas.
