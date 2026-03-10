# Desenvolvimento do Epic `sistema-pagamentos`

Abaixo estão os três comandos executados via CLI do Oraculo para inicializar o epic, salvar os requisitos e listar os stories.

---

## Passo 1 — Criar o epic

```bash
oraculo tools epic init sistema-pagamentos --description "Sistema de processamento de pagamentos"
```

**Resultado:**

```json
{
  "created": true,
  "name": "sistema-pagamentos",
  "path": ".oraculo/epics/sistema-pagamentos"
}
```

O epic foi criado com sucesso. O campo `created: true` indica que o registro foi inserido (não era pré-existente). O diretório `.oraculo/epics/sistema-pagamentos` será usado para armazenar os artefatos do epic.

---

## Passo 2 — Salvar os requisitos no epic

O comando `epic save` lê o conteúdo via stdin, portanto os requisitos são passados com heredoc:

```bash
printf '# Requisitos\n\nO sistema deve processar pagamentos via cartão de crédito e PIX.' \
  | oraculo tools epic save sistema-pagamentos
```

**Resultado:**

```json
{
  "name": "sistema-pagamentos",
  "path": ".oraculo/epics/sistema-pagamentos/requirements.md",
  "saved": true
}
```

Os requisitos foram gravados em `.oraculo/epics/sistema-pagamentos/requirements.md`. O campo `saved: true` confirma que o arquivo foi criado/sobrescrito com sucesso.

---

## Passo 3 — Listar os stories do epic

```bash
oraculo tools story list --epic sistema-pagamentos
```

**Resultado:**

```json
null
```

Nenhum story foi criado ainda para este epic. A lista está vazia — o retorno `null` é o comportamento esperado da CLI quando a tabela de stories não possui registros para o epic informado (o slice Go é `nil` e serializa como `null` em JSON).

---

## Resumo

| Passo | Comando | Status |
|-------|---------|--------|
| 1 | `epic init sistema-pagamentos --description "..."` | Epic criado com sucesso |
| 2 | `epic save sistema-pagamentos` (stdin) | Requisitos salvos em `requirements.md` |
| 3 | `story list --epic sistema-pagamentos` | Lista vazia — nenhum story ainda |
