<div align="center">
  <img src="apps/desktop/build/appicon.png" alt="Oraculo" width="120" />
  <h1>Oraculo</h1>
  <p><strong>A Socratic guide and team orchestrator for quality product development.</strong></p>
  <p>Asks before doing. Plans before building. Validates before shipping.</p>

  <p>
    <a href="#install"><img src="https://img.shields.io/badge/Claude_Code-Plugin-2563eb?style=flat-square" alt="Claude Code Plugin" /></a>
    <img src="https://img.shields.io/badge/TypeScript-Bun-f472b6?style=flat-square&logo=bun" alt="TypeScript/Bun" />
    <img src="https://img.shields.io/badge/Platform-macOS-000000?style=flat-square&logo=apple" alt="macOS" />
    <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="License" />
  </p>
</div>

---

## The problem with "just do it" AI

Most AI coding tools have the same default posture: you describe something, they build it. Fast. Confidently. Often wrong in ways you don't discover until the sprint review.

They skip the part where someone asks *"is this actually the right problem to solve?"* They skip the architecture review. They skip independent QA. And when you need human judgment — a tradeoff decision, a scope call — there's nowhere to pause.

**Oraculo takes the opposite stance.**

It treats code as the last step, not the first. Before any implementation, it guides your team through structured discovery and planning. When it's time to build, it assembles a coordinated team of agents — not a single actor rushing to completion. And at every critical decision point, it stops and waits for a human verdict.

---

## How it works

Oraculo operates as a **Socratic guide** during discovery and a **team orchestrator** during execution. It never writes code directly — it delegates to specialized agents while remaining the single source of coordination.

### The workflow

```
Epic  →  Discover  →  Plan  →  Execute  →  Validate
Story            →  Plan  →  Execute  →  Validate
```

**Discover** — Oraculo questions your idea. It surfaces edge cases, identifies risks, and challenges assumptions before anything is committed to. The output is a requirements document reviewed and approved by your team.

**Plan** — Requirements are decomposed into a dependency graph (DAG). Two parallel research agents analyze the codebase and external best practices, producing an architectural design. The design goes through a mandatory approval gate before any code is written.

**Execute** — A team of agents works in parallel, respecting the DAG's dependency order. Every task follows TDD — tests first, implementation second. Agents are self-contained: each receives its full context, project patterns, and expected behavior.

**Validate** — A dedicated QA agent reviews the implementation with fresh eyes. No bias from having written the code. If it rejects, the workflow returns to the appropriate phase — nothing is forced through.

### Human-in-the-loop gates

Oraculo doesn't automate past the decisions that matter. Three mandatory gates pause the workflow until a human delivers a verdict through the dashboard:

| Gate | When | Verdicts |
|---|---|---|
| **Design** | After architecture is drafted, before any code | `approved` / `rejected` / `needs_revision` |
| **Execution Plan** | Before agents are dispatched (large epics) | `approved` / `rejected` / `needs_revision` |
| **QA Escalation** | When QA finds a critical defect it can't resolve | `approved` / `rejected` / `needs_revision` |

Document reviews (requirements, story definitions) use a separate versioning system with `approved` / `rejected` verdicts.

---

## The dashboard

<div align="center">
  <img src="apps/desktop/build/appicon.png" alt="Dashboard" width="64" />
  <br/>
  <em>Mission Control — not a log viewer.</em>
</div>

The dashboard is Oraculo's observation and control surface. It shows:

- **Live agent activity** — which agents are running, what they're doing, when they finish
- **DAG visualization** — the full task graph with real-time status
- **Approval gates** — surfaces artifacts for review and collects human verdicts
- **Knowledge base** — accumulated lessons learned across all epics

Real-time updates flow through two channels: HTTP hooks push telemetry events as they happen; MCP delivers blocking approval gate notifications when an agent waits for a verdict.

---

## How Oraculo compares

| | Oraculo | Get-shit-done tools | Superpowers | spec-workflow-mcp |
|---|:---:|:---:|:---:|:---:|
| Discovery & Socratic questioning | ✅ | ❌ | ❌ | Partial |
| Architecture review gate | ✅ | ❌ | ❌ | ❌ |
| Parallel agent orchestration (DAG) | ✅ | ❌ | ❌ | ❌ |
| TDD enforced across all agents | ✅ | ❌ | Depends on skill | ❌ |
| Independent QA agent | ✅ | ❌ | ❌ | ❌ |
| Human-in-the-loop approval gates | ✅ | ❌ | ❌ | ❌ |
| Persistent knowledge accumulation | ✅ | ❌ | ❌ | ❌ |
| Real-time dashboard (Mission Control) | ✅ | ❌ | ❌ | ❌ |
| Works with your existing project | ✅ | ✅ | ✅ | ✅ |

**Get-shit-done tools** are optimized for speed. They execute immediately, often well, but skip discovery, architecture review, and independent validation. You get code fast; correctness is your problem.

**Superpowers / skill kits** enhance how a single agent works. They add structure to prompts, enforce patterns, and improve output quality. But they're still one agent, no persistent state, no team coordination, no approval gates.

**spec-workflow-mcp** handles specification generation — valuable, but the workflow ends where Oraculo's begins. There's no execution engine, no QA, no knowledge that accumulates across features.

---

## Install

### Prerequisites / Pré-requisitos

- [Claude Code](https://claude.ai/code) installed / instalado
- [Bun](https://bun.sh) runtime (for building from source / para compilar do código-fonte)
- macOS (Linux support in progress / suporte a Linux em progresso)

### Via Claude Code plugin marketplace

Search for **Oraculo** in the Claude Code plugin marketplace, or install directly:

Procure por **Oraculo** no marketplace de plugins do Claude Code, ou instale diretamente:

```bash
# Coming soon via marketplace
# Em breve via marketplace
```

### Via npm (CLI only / apenas CLI)

```bash
bun install -g @oraculo/cli
oraculo install
```

### From source / Do código-fonte

```bash
git clone https://github.com/your-org/oraculo
cd oraculo
cd apps/orchestrator && bun install && cd ../..
make install
```

`oraculo install` registers HTTP hooks in your project's `.claude/settings.json`, copies skills, and starts the local server.

`oraculo install` registra HTTP hooks no `.claude/settings.json` do seu projeto, copia skills e inicia o servidor local.

---

## Usage / Uso

Once installed in a project, Oraculo is invoked through Claude Code slash commands:

Uma vez instalado no projeto, o Oraculo é invocado através de slash commands do Claude Code:

```
/oraculo:epic    — Start product discovery / Iniciar descoberta de produto
/oraculo:story   — Define and plan a work item / Definir e planejar um item de trabalho
/oraculo:plan    — Decompose into an executable DAG / Decompor em um DAG executável
/oraculo:execute — Dispatch agents to implement / Despachar agentes para implementar
/oraculo:validate — Run independent QA / Executar QA independente
```

Open the dashboard to monitor agents, review artifacts, and deliver verdicts at approval gates.

Abra o dashboard para monitorar agentes, revisar artefatos e entregar veredictos nos gates de aprovação.

---

## Architecture / Arquitetura

```
claude-kit/skills/       — Claude Code skills (slash commands)
apps/orchestrator/       — TypeScript/Bun binary: CLI + HTTP server + MCP server
apps/frontend/           — Next.js dashboard (observation & control)
apps/desktop/            — Wails macOS app (bundles binary + dashboard)
npm/                     — Cross-platform binary distribution via npm
```

### Tech stack

| Layer | Technology |
|---|---|
| Runtime | [Bun](https://bun.sh) (native TypeScript, single-binary compilation) |
| HTTP / WebSocket | [Hono](https://hono.dev) on `Bun.serve()` |
| Database | `bun:sqlite` (native, sync API) |
| CLI | [Commander](https://github.com/tj/commander.js) with typed options |
| MCP Server | `@modelcontextprotocol/sdk` |
| Agent Orchestration | `@anthropic-ai/claude-agent-sdk` |
| Schema Validation | [Zod](https://zod.dev) |
| Logging | [Pino](https://getpino.io) + `pino-roll` |
| Frontend | [Next.js](https://nextjs.org) (static export) |
| Desktop | [Wails v3](https://wails.io) (Go, embeds frontend + bundles binary) |

### Trust Layer

The orchestrator binary is the **Trust Layer** — all data access goes through it. The dashboard never reads files or queries SQLite directly; everything flows through the CLI's validated commands.

O binário do orquestrador é a **Camada de Confiança** — todo acesso a dados passa por ele. O dashboard nunca lê arquivos ou consulta SQLite diretamente; tudo flui pelos comandos validados do CLI.

### Data

SQLite (`.oraculo/oraculo.db`) holds two kinds of data: transient operational state (tasks, approvals, agent lifecycle) and a persistent `knowledge` table that accumulates lessons learned across all epics.

SQLite (`.oraculo/oraculo.db`) armazena dois tipos de dados: estado operacional transitório (tarefas, aprovações, ciclo de vida de agentes) e uma tabela persistente `knowledge` que acumula lições aprendidas de todos os épicos.

---

## Development / Desenvolvimento

### Build

```bash
# Install dependencies / Instalar dependências
cd apps/orchestrator && bun install

# Build the CLI binary / Compilar o binário CLI
make build

# Build everything (frontend + backend) / Compilar tudo
make rebuild

# Run tests / Executar testes
make test

# Run the dashboard in dev mode / Dashboard em modo de desenvolvimento
make web-dev

# Cross-compile for all platforms / Compilar para todas as plataformas
make cross-compile
```

### Project structure / Estrutura do projeto

```
apps/orchestrator/src/
├── cli/            — Commander subcommands (lifecycle, install, hooks, tools)
├── db/             — bun:sqlite, migrations, data stores
├── domain/         — TypeScript interfaces, enums, state machine
├── server/         — Hono HTTP, WebSocket hub, SSE broadcaster
├── mcp/            — MCP server (approval gates)
├── orchestrator/   — DAG evaluation, agent dispatch, execution loop
├── config/         — config.json management
├── logging/        — Pino + execution logs
├── registry/       — servers.json management
└── utils/          — Helpers (output, env, UUID)
```

---

## Philosophy / Filosofia

> *Ask before doing. Orchestrate, never execute. Maximize parallelism. Quality over speed. Human in the loop.*

> *Perguntar antes de fazer. Orquestrar, nunca executar. Maximizar paralelismo. Qualidade sobre velocidade. Humano no controle.*

Oraculo is built on the conviction that most AI-generated code problems are requirements problems. The agent understood the wrong thing, made a scope assumption, or skipped a constraint that wasn't written down. The fix isn't faster agents — it's a better process before the agents start.

O Oraculo é construído na convicção de que a maioria dos problemas com código gerado por IA são problemas de requisitos. O agente entendeu errado, fez uma suposição de escopo ou pulou uma restrição que não foi documentada. A solução não são agentes mais rápidos — é um processo melhor antes dos agentes começarem.

Full philosophy / Filosofia completa: [docs/philosophy.md](docs/philosophy.md)

---

<div align="center">
  <sub>Built on the <a href="https://claude.ai/code">Claude Code</a> ecosystem.</sub>
</div>
