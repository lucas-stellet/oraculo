# Tarefa: Iniciar desenvolvimento do epic `sistema-pagamentos`

## Passo 1 — Criar o epic

```bash
oraculo tools epic init sistema-pagamentos --description "Sistema de processamento de pagamentos"
```

**Resultado:**
```json
{
  "name": "sistema-pagamentos",
  "path": ".oraculo/epics/sistema-pagamentos",
  "created": true
}
```

---

## Passo 2 — Salvar os requisitos no epic

```bash
cat <<'EOF' | oraculo tools epic save sistema-pagamentos
# Requisitos

O sistema deve processar pagamentos via cartão de crédito e PIX.
EOF
```

**Resultado:**

_(sem saída — comando bem-sucedido; requisitos gravados em `.oraculo/epics/sistema-pagamentos/requirements.md`)_

Para confirmar, é possível ler de volta os requisitos com:

```bash
oraculo tools epic get sistema-pagamentos
```

**Saída:**
```
# Requisitos

O sistema deve processar pagamentos via cartão de crédito e PIX.
```

---

## Passo 3 — Listar os stories do epic

```bash
oraculo tools story list --epic sistema-pagamentos
```

**Resultado:**
```json
[]
```

Nenhum story foi criado ainda para o epic `sistema-pagamentos`. A lista está vazia porque o epic acabou de ser inicializado e nenhum story foi adicionado.

---

## Resumo

| Passo | Comando | Status |
|-------|---------|--------|
| 1 | `oraculo tools epic init sistema-pagamentos --description "..."` | Epic criado com sucesso |
| 2 | `cat <<'EOF' \| oraculo tools epic save sistema-pagamentos` | Requisitos salvos com sucesso |
| 3 | `oraculo tools story list --epic sistema-pagamentos` | Lista retornada — 0 stories existentes |
