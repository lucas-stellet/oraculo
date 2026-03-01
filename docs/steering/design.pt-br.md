# Oraculo Steering — Design

## 1. Visão Geral

Steering é a camada de contexto do Oraculo — dois documentos markdown que capturam a identidade do produto e a identidade técnica. Eles ficam em `.oraculo/steering/` e são lidos obrigatoriamente pelo `/oraculo:epic` e pelo `/oraculo:story` antes de qualquer exploração ou definição começar.

Dois documentos, dois domínios:

- **`product.md`** — Empresa, produto, visão, princípios de negócio
- **`design.md`** — Tech stack, arquitetura, convenções de código, estrutura do projeto

```
.oraculo/
├── oraculo.db          # Estado operacional + conhecimento
├── config              # Porta do dashboard, configurações do projeto
└── steering/
    ├── product.md      # Identidade do produto
    └── design.md       # Identidade técnica
```

A criação segue um modelo híbrido: o Oraculo oferece templates estruturados, o usuário os preenche, o Oraculo valida a completude e faz perguntas sobre lacunas — nunca inventa conteúdo. O usuário é o único autor. O Oraculo é o validador.

O Steering é complementar ao CLAUDE.md. O CLAUDE.md contém instruções operacionais — como rodar o projeto, quais comandos usar, quais convenções seguir no código. O Steering contém a identidade do projeto — o que é o produto, a quem ele serve, para onde está indo e como é construído. Ambos são lidos por skills e agents. Nenhum substitui o outro.

## 2. Estrutura dos Documentos

### 2.1 Documento de Produto (`product.md`)

Quatro seções obrigatórias:

| Seção | Conteúdo | Propósito |
|---|---|---|
| **Company** | Nome, setor, missão/propósito, tamanho/estágio | Calibra o Oraculo para o domínio de negócio |
| **Product** | Nome do produto, propósito (uma frase), usuários-alvo, problema que resolve, estado atual (produção/desenvolvimento/stage) | Define o que existe e para quem |
| **Vision** | Para onde o produto quer ir — não um roadmap, a direção que guia a priorização. Um parágrafo ou duas frases. Opcionalmente inclui horizonte de tempo | Ancora exploração e priorização |
| **Business Principles** | Regras inegociáveis que governam decisões de produto. Restrições, não instruções. Cada uma tem nome + descrição curta | Tratados como guardrails durante a exploração |

### 2.2 Documento de Design (`design.md`)

Quatro seções obrigatórias:

| Seção | Conteúdo | Propósito |
|---|---|---|
| **Tech Stack** | Linguagens, frameworks, banco de dados, infra, CI/CD | Code agents produzem código na linguagem e ferramentas corretas |
| **Architecture** | Padrão arquitetural, comunicação entre componentes, armazenamento, integrações externas relevantes | O orquestrador decompõe o trabalho corretamente |
| **Code Conventions** | Tratamento de erros, nomenclatura, estilo de testes, formato de commits, convenções específicas do projeto | Code agents seguem essas convenções; QA agents validam contra elas |
| **Project Structure** | Organização de diretórios e onde cada tipo de código reside. Pode ser uma árvore de diretórios com descrições curtas | Agents sabem onde criar arquivos e encontrar código existente |

### 2.3 Seções Opcionais (Futuro)

- **Produto:** Contexto de mercado, métricas de sucesso
- **Design:** Decisões técnicas (ADR simplificado), restrições técnicas

Seções opcionais nunca são obrigatórias. O Oraculo trabalha com o que existe.

## 3. Templates

O Oraculo distribui templates para ambos os documentos. Os templates definem a estrutura esperada sem preencher conteúdo — cada seção tem uma descrição breve e um placeholder.

### 3.1 Template de Produto

```markdown
# Steering — Product

## Company
<!-- Name, sector, mission/purpose, size/stage -->

## Product
<!-- Product name, purpose (what it does), target users,
     problem it solves, current state -->

## Vision
<!-- Where the product wants to go. Not a roadmap — the direction
     that guides prioritization. -->

## Business Principles
<!-- Non-negotiable rules governing product decisions.
     Format: name + short description. -->
```

### 3.2 Template de Design

```markdown
# Steering — Design

## Tech Stack
<!-- Languages, frameworks, database, infra, CI/CD -->

## Architecture
<!-- Architectural pattern, component communication,
     storage, external integrations -->

## Code Conventions
<!-- Error handling, naming, test style, commit format,
     project-specific conventions -->

## Project Structure
<!-- Directory organization. Can be a tree with
     short descriptions of each main directory. -->
```

### 3.3 Localização dos Templates

Os templates ficam dentro do kit distribuível do Oraculo:

```
claude-kit/
└── skills/
    └── oraculo/
        └── steering/
            └── references/
                ├── product-template.md
                └── design-template.md
```

A skill de Steering lê o template da sua pasta `references/`, copia para `.oraculo/steering/` e informa o usuário. A partir daí, o usuário preenche e o Oraculo valida.

## 4. Fluxo de Criação e Validação

### 4.1 Quando o Steering é Criado

Dois caminhos:

- **Explícito** — O usuário invoca um fluxo de criação de Steering. Os templates são copiados e a validação começa.
- **Implícito** — O usuário invoca `/oraculo:epic` ou `/oraculo:story` e o Steering não existe. A skill detecta a ausência, informa o usuário e oferece a criação. O usuário pode aceitar (criação inline) ou recusar (a skill continua com menos contexto).

### 4.2 Fluxo de Criação Híbrido

Três etapas:

1. **Template** — Copiar o template para `.oraculo/steering/`.
2. **Preenchimento** — O usuário preenche as seções (fora do Oraculo ou ditando na sessão). O conteúdo é sempre de autoria do usuário.
3. **Validação** — O Oraculo lê o documento preenchido e verifica a completude por seção:
   - **Preenchida:** Confirma e prossegue.
   - **Vazia/placeholder:** Pergunta se o usuário quer preencher agora ou deixar como lacuna explícita.
   - **Ambígua/superficial:** Faz perguntas complementares para aprofundar.

O Oraculo nunca rejeita um documento por estar incompleto. Lacunas são aceitas — o que importa é que sejam visíveis.

### 4.3 Atualizações

O Steering é atualizado manualmente pelo usuário. O Oraculo não modifica documentos como efeito colateral. Se uma divergência for detectada durante a exploração, o Oraculo informa, mas não se auto-corrige.

### 4.4 Diagrama de Fluxo

```
/oraculo:epic ou /oraculo:story invocado
        │
        ▼
Steering existe em .oraculo/steering/?
   ├── SIM → Ler product.md e design.md → Continuar skill
   └── NÃO → Informar usuário
                │
                ▼
        Usuário quer criar agora?
           ├── SIM → Copiar templates → Preencher → Validar → Continuar skill
           └── NÃO → Continuar sem Steering (menos contexto)
```

## 5. Integração com Skills

### 5.1 Como Skills Leem o Steering

Skills leem via CLI: `oraculo tools steering get --doc product` e `oraculo tools steering get --doc design`. O CLI retorna markdown puro (convenção de comando de conteúdo). Se o arquivo não existir, o CLI retorna um erro e a skill segue o fluxo de detecção da Seção 4.1.

### 5.2 O Que Muda no Epic com o Steering

O Steering não altera os frameworks (Double Diamond, JTBD, etc.) — ele enriquece o ponto de partida. Sem Steering, as perguntas são genéricas. Com Steering, as perguntas são informadas e calibradas para o domínio.

O Steering é injetado como um bloco de referência. A skill o utiliza para:

- Calibrar perguntas para o domínio de negócio e a base de usuários
- Verificar conflitos com princípios de negócio
- Referenciar a arquitetura existente ao explorar viabilidade técnica
- Evitar perguntas já respondidas pelos documentos de steering

### 5.3 O Que Muda na Story com o Steering

Stories — especialmente stories avulsas fora de um epic — chegam com menos contexto. O Steering compensa de forma mais aguda. A skill cruza a solicitação do usuário com o contexto do projeto: tech stack, restrições arquiteturais, princípios de negócio. Uma story que diz "adicionar um webhook de pagamento" é interpretada de forma diferente quando o Steering indica que o produto usa Stripe versus quando indica que o produto usa um sistema de cobrança próprio.

### 5.4 O Que Muda em Plan e Execute

O orquestrador lê o Design antes de decompor o trabalho:

- **Architecture** informa a estrutura das tarefas — monolito vs. serviços determina como as tarefas são divididas.
- **Conventions** informam as instruções do code agent — nomenclatura, tratamento de erros e estilo de testes são passados como restrições.
- **Project structure** informa o posicionamento dos arquivos — agents sabem onde criar arquivos sem precisar adivinhar.

O Design é passado como contexto a cada code agent junto com a descrição da tarefa e os arquivos relevantes.

### 5.5 O Que Muda em Validate

O QA agent recebe o Design como parte do seu contexto de validação. Além de verificar os resultados dos testes, o QA valida:

- Conformidade com as convenções de código (nomenclatura, tratamento de erros, estilo de testes)
- Alinhamento arquitetural (camadas corretas, padrões de comunicação corretos)
- Implicações dos princípios de negócio (quando a mudança tem consequências voltadas ao produto)

## 6. Comandos CLI

### 6.1 Comandos

| Comando | Saída | Propósito |
|---|---|---|
| `oraculo tools steering init` | JSON | Copiar templates para `.oraculo/steering/`. Idempotente |
| `oraculo tools steering init --doc product` | JSON | Copiar apenas o template de Produto |
| `oraculo tools steering init --doc design` | JSON | Copiar apenas o template de Design |
| `oraculo tools steering get --doc product` | Conteúdo puro | Retornar o conteúdo de `product.md`. Erro se o arquivo não existir |
| `oraculo tools steering get --doc design` | Conteúdo puro | Retornar o conteúdo de `design.md`. Erro se o arquivo não existir |
| `oraculo tools steering status` | JSON | Estado de cada documento: existe ou não, seções preenchidas vs. vazias |

### 6.2 Comportamento

**`steering init`** segue o padrão idempotente. Chamá-lo quando os templates já existem não os sobrescreve:

```json
{
  "product": { "created": true, "path": ".oraculo/steering/product.md" },
  "design": { "created": true, "path": ".oraculo/steering/design.md" }
}
```

Segunda chamada:

```json
{
  "product": { "created": false, "path": ".oraculo/steering/product.md" },
  "design": { "created": false, "path": ".oraculo/steering/design.md" }
}
```

**`steering get`** retorna markdown puro — a convenção de comando de conteúdo. Sem encapsulamento JSON. O consumidor lê o documento como está.

**`steering status`** reporta o estado por seção:

```json
{
  "product": {
    "exists": true,
    "sections": {
      "Company": "filled",
      "Product": "filled",
      "Vision": "empty",
      "Business Principles": "filled"
    }
  },
  "design": {
    "exists": true,
    "sections": {
      "Tech Stack": "filled",
      "Architecture": "filled",
      "Code Conventions": "empty",
      "Project Structure": "empty"
    }
  }
}
```

**Lógica de detecção:** O CLI verifica a existência de texto não-comentário abaixo de cada heading. Uma seção com apenas comentários HTML ou espaços em branco é reportada como `"empty"`. Uma seção com qualquer outro conteúdo é `"filled"`. Trata-se de uma heurística simples — sem análise semântica.

### 6.3 Integração com `oraculo install`

`oraculo install` **não** cria automaticamente documentos de Steering. O Steering é opt-in. Isso mantém a instalação enxuta e evita criar arquivos que o usuário pode não querer preencher imediatamente.

## 7. Trabalho Futuro

Capacidades adiadas para iterações futuras:

- **Skill dedicada (`/oraculo:steering`)** — Sessão Socrática para criação guiada, similar ao `/oraculo:epic`
- **Detecção de divergência** — Detectar contradições entre o Steering e declarações do usuário durante Epic/Story, oferecendo atualização inline
- **Seções opcionais de Produto** — Contexto de mercado, métricas de sucesso
- **Seções opcionais de Design** — Decisões técnicas (ADR simplificado), restrições técnicas
- **Versionamento do Steering** — Rastrear mudanças com timestamps e justificativas
- **Comando de validação CLI** — `oraculo tools steering validate` para verificação de coerência
- **Importação de contexto existente** — Popular a partir de CLAUDE.md, README.md, `package.json` como sugestões para validação do usuário
