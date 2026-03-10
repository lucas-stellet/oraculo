import type { EpicSummary, Story, StoryTask, StoryDetail, StoryVersion, Review, Validation, Approval } from "./types";

// ── Epic ────────────────────────────────────────────────────────────────────
export const mockEpic: EpicSummary = {
  id: 1,
  name: "Gastos Pessoais",
  description:
    "Sistema completo de controle de gastos pessoais com categorização automática e dashboard de resumo.",
  approval_status: "approved",
  phase: "execute",
  phase_status: "in_progress",
  story_count: 4,
  task_count: 42,
  completed_task_count: 24,
  created_at: "2026-03-01T10:00:00Z",
  updated_at: "2026-03-09T14:30:00Z",
};

// ── Stories ──────────────────────────────────────────────────────────────────
export const mockStories: Story[] = [
  {
    id: 1,
    epic_id: 1,
    name: "Registro de Gastos",
    status: "in_progress",
    task_count: 12,
    completed_task_count: 8,
    running_status: "Executing now",
    last_activity: "2 min ago",
  },
  {
    id: 2,
    epic_id: 1,
    name: "Categorização Automática",
    status: "completed",
    task_count: 12,
    completed_task_count: 12,
    running_status: "Completed",
    last_activity: "1 hour ago",
  },
  {
    id: 3,
    epic_id: 1,
    name: "Dashboard de Resumo",
    status: "pending",
    task_count: 6,
    completed_task_count: 0,
    running_status: "Not started",
    last_activity: "Waiting",
  },
  {
    id: 4,
    epic_id: 1,
    name: "Exportação de Relatórios",
    status: "pending",
    task_count: 12,
    completed_task_count: 0,
    running_status: "Not started",
    last_activity: "Waiting",
  },
];

// ── Tasks ───────────────────────────────────────────────────────────────────
export const mockTasks: Record<number, StoryTask[]> = {
  1: [
    { id: 1, story_id: 1, name: "Create expense model schema", description: "Define SQLite schema for expenses table with amount, description, category_id, date, and timestamps.", status: "completed", agent: "code-01", duration: "3m 42s", depends_on: [], result: { summary: "Created expenses table with 7 columns, added migration file and model struct with JSON tags.", files_modified: ["src/domain/expense.go", "src/db/migrations/001_expenses.sql"], skills_used: ["tdd"], failure_reason: null } },
    { id: 2, story_id: 1, name: "Implement repository layer", description: "Build CRUD repository for expenses with prepared statements and transaction support.", status: "completed", agent: "code-02", duration: "5m 18s", depends_on: [1], result: { summary: "Implemented ExpenseRepository with Create, Get, List, Update, Delete methods using sqlx.", files_modified: ["src/db/expense_repo.go", "src/db/expense_repo_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 3, story_id: 1, name: "Build REST API endpoints", description: "Create HTTP handlers for POST/GET/PUT/DELETE /expenses with proper error responses.", status: "completed", agent: "code-01", duration: "4m 07s", depends_on: [2], result: { summary: "Added 5 REST endpoints with request validation, error mapping, and OpenAPI-compatible responses.", files_modified: ["src/cli/api/expenses.go", "src/cli/api/routes.go", "src/cli/api/expenses_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 4, story_id: 1, name: "Write input validation rules", description: "Add validation for amount (positive, max 2 decimals), description (non-empty, max 255), and date format.", status: "failed", agent: "code-03", duration: "1m 22s", depends_on: [3], result: { summary: "Attempted to add validation middleware but encountered conflicts with existing error handler.", files_modified: ["src/cli/api/validation.go"], skills_used: ["tdd"], failure_reason: "Validation middleware conflicts with the global error handler — error responses are double-wrapped. Needs architectural decision on error handling strategy." } },
    { id: 5, story_id: 1, name: "Create unit tests for model", description: "Table-driven tests for expense model validation, serialization, and edge cases.", status: "pending", agent: null, duration: null, depends_on: [1], result: null },
    { id: 6, story_id: 1, name: "Integration tests for API", description: "Test full request cycle with in-memory SQLite, covering happy paths and error cases.", status: "pending", agent: null, duration: null, depends_on: [3, 4], result: null },
    { id: 7, story_id: 1, name: "Add currency formatting", description: "Implement locale-aware currency formatting for BRL with proper decimal and thousand separators.", status: "completed", agent: "code-01", duration: "2m 15s", depends_on: [1], result: { summary: "Added FormatCurrency helper with BRL locale support and unit tests for edge cases.", files_modified: ["src/domain/currency.go", "src/domain/currency_test.go"], skills_used: [], failure_reason: null } },
    { id: 8, story_id: 1, name: "Implement date range filtering", description: "Add start_date/end_date query params to GET /expenses with index-backed filtering.", status: "completed", agent: "code-02", duration: "3m 50s", depends_on: [3], result: { summary: "Extended List query with optional date range filter, added composite index on (date, id).", files_modified: ["src/db/expense_repo.go", "src/cli/api/expenses.go", "src/db/migrations/002_date_index.sql"], skills_used: ["tdd"], failure_reason: null } },
    { id: 9, story_id: 1, name: "Build expense list pagination", description: "Cursor-based pagination for expense list with configurable page size and next_cursor response.", status: "completed", agent: "code-01", duration: "4m 33s", depends_on: [3], result: { summary: "Implemented cursor-based pagination using (date, id) as cursor, default page size 20.", files_modified: ["src/db/expense_repo.go", "src/cli/api/expenses.go", "src/cli/api/pagination.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 10, story_id: 1, name: "Add bulk delete endpoint", description: "DELETE /expenses/bulk accepting array of IDs with transactional deletion.", status: "completed", agent: "code-03", duration: "2m 58s", depends_on: [2], result: { summary: "Added bulk delete with transaction, returns count of deleted records. Max 100 IDs per request.", files_modified: ["src/cli/api/expenses.go", "src/db/expense_repo.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 11, story_id: 1, name: "Implement search by description", description: "Full-text search on expense description using SQLite FTS5 extension.", status: "completed", agent: "code-02", duration: "1m 45s", depends_on: [2], result: { summary: "Created FTS5 virtual table for expense descriptions, added search endpoint with ranking.", files_modified: ["src/db/migrations/003_fts.sql", "src/db/expense_repo.go", "src/cli/api/expenses.go"], skills_used: [], failure_reason: null } },
    { id: 12, story_id: 1, name: "E2E tests for expense flow", description: "End-to-end tests covering create, list, filter, paginate, search, and delete flows.", status: "pending", agent: null, duration: null, depends_on: [4, 5, 6], result: null },
  ],
  2: [
    { id: 13, story_id: 2, name: "Define category taxonomy", description: "Research and define hierarchical category taxonomy for personal expenses.", status: "completed", agent: "research-01", duration: "6m 12s", depends_on: [], result: { summary: "Defined 3-level taxonomy with 8 top-level categories and 32 subcategories.", files_modified: ["docs/category-taxonomy.md"], skills_used: [], failure_reason: null } },
    { id: 14, story_id: 2, name: "Implement keyword matcher", description: "Build keyword-based matcher for automatic category suggestion.", status: "completed", agent: "code-01", duration: "4m 30s", depends_on: [13], result: { summary: "Implemented trie-based keyword matcher with Portuguese stemming support.", files_modified: ["src/domain/matcher.go", "src/domain/matcher_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 15, story_id: 2, name: "Build category assignment logic", description: "Core logic for assigning categories to expenses based on matcher results.", status: "completed", agent: "code-02", duration: "5m 44s", depends_on: [13, 14], result: { summary: "Built assignment pipeline with fallback chain: exact match → keyword → historical.", files_modified: ["src/domain/categorizer.go", "src/domain/categorizer_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 16, story_id: 2, name: "Create manual override API", description: "API for users to manually override category assignments.", status: "completed", agent: "code-01", duration: "3m 10s", depends_on: [15], result: { summary: "Added PUT /expenses/:id/category endpoint with override tracking.", files_modified: ["src/cli/api/categories.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 17, story_id: 2, name: "Add category CRUD endpoints", description: "Full CRUD for managing custom categories.", status: "completed", agent: "code-03", duration: "2m 55s", depends_on: [13], result: { summary: "Added REST endpoints for category management with parent-child validation.", files_modified: ["src/cli/api/categories.go", "src/db/category_repo.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 18, story_id: 2, name: "Implement learning from overrides", description: "System learns from manual overrides to improve future suggestions.", status: "completed", agent: "code-02", duration: "7m 20s", depends_on: [16], result: { summary: "Added feedback loop that updates keyword weights based on override patterns.", files_modified: ["src/domain/matcher.go", "src/domain/feedback.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 19, story_id: 2, name: "Unit tests for matcher", description: "Comprehensive tests for keyword matcher edge cases.", status: "completed", agent: "code-01", duration: "3m 15s", depends_on: [14], result: { summary: "Added 24 test cases covering stemming, accents, and multi-word matches.", files_modified: ["src/domain/matcher_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 20, story_id: 2, name: "Unit tests for assignment logic", description: "Tests for category assignment pipeline and fallback chain.", status: "completed", agent: "code-02", duration: "2m 48s", depends_on: [15], result: { summary: "Added 18 test cases for assignment pipeline including ambiguous inputs.", files_modified: ["src/domain/categorizer_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 21, story_id: 2, name: "Integration tests for override flow", description: "Test full override cycle: assign → override → verify learning.", status: "completed", agent: "code-03", duration: "4m 02s", depends_on: [16, 18], result: { summary: "Added integration tests verifying feedback loop improves subsequent assignments.", files_modified: ["src/cli/api/categories_integration_test.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 22, story_id: 2, name: "Add confidence scoring", description: "Return confidence score with each category suggestion.", status: "completed", agent: "code-01", duration: "5m 30s", depends_on: [15], result: { summary: "Added 0-100 confidence score based on match quality and historical accuracy.", files_modified: ["src/domain/categorizer.go", "src/cli/api/categories.go"], skills_used: [], failure_reason: null } },
    { id: 23, story_id: 2, name: "Build batch categorization endpoint", description: "Endpoint to categorize multiple expenses in a single request.", status: "completed", agent: "code-02", duration: "3m 40s", depends_on: [15], result: { summary: "Added POST /expenses/categorize/batch with max 50 items per request.", files_modified: ["src/cli/api/categories.go"], skills_used: ["tdd"], failure_reason: null } },
    { id: 24, story_id: 2, name: "E2E tests for categorization", description: "End-to-end tests for the full categorization workflow.", status: "completed", agent: "code-03", duration: "5m 55s", depends_on: [19, 20, 21], result: { summary: "Added 8 E2E scenarios covering auto-categorize, override, and batch flows.", files_modified: ["tests/e2e/categorization_test.go"], skills_used: ["tdd"], failure_reason: null } },
  ],
  3: [
    { id: 25, story_id: 3, name: "Design dashboard layout", description: "Design the visual layout for the expenses dashboard.", status: "pending", agent: null, duration: null, depends_on: [], result: null },
    { id: 26, story_id: 3, name: "Implement monthly summary aggregation", description: "Build aggregation queries for monthly expense summaries.", status: "pending", agent: null, duration: null, depends_on: [25], result: null },
    { id: 27, story_id: 3, name: "Build chart data endpoints", description: "API endpoints providing chart-ready data.", status: "pending", agent: null, duration: null, depends_on: [26], result: null },
    { id: 28, story_id: 3, name: "Create category breakdown view", description: "Visual breakdown of expenses by category.", status: "pending", agent: null, duration: null, depends_on: [27], result: null },
    { id: 29, story_id: 3, name: "Add date range selector", description: "Interactive date range selector for filtering dashboard data.", status: "pending", agent: null, duration: null, depends_on: [25], result: null },
    { id: 30, story_id: 3, name: "Unit tests for aggregation", description: "Tests for summary aggregation logic.", status: "pending", agent: null, duration: null, depends_on: [26], result: null },
  ],
  4: [
    { id: 31, story_id: 4, name: "Define export formats (CSV, PDF)", description: "Define format specifications for CSV and PDF exports.", status: "pending", agent: null, duration: null, depends_on: [], result: null },
    { id: 32, story_id: 4, name: "Implement CSV generator", description: "Generate CSV files from expense data.", status: "pending", agent: null, duration: null, depends_on: [31], result: null },
    { id: 33, story_id: 4, name: "Implement PDF generator", description: "Generate PDF reports from expense data.", status: "pending", agent: null, duration: null, depends_on: [31], result: null },
    { id: 34, story_id: 4, name: "Build export API endpoint", description: "REST endpoint for triggering exports.", status: "pending", agent: null, duration: null, depends_on: [32, 33], result: null },
    { id: 35, story_id: 4, name: "Add date range filter for exports", description: "Filter exported data by date range.", status: "pending", agent: null, duration: null, depends_on: [34], result: null },
    { id: 36, story_id: 4, name: "Implement category filter for exports", description: "Filter exported data by category.", status: "pending", agent: null, duration: null, depends_on: [34], result: null },
    { id: 37, story_id: 4, name: "Create export queue system", description: "Background queue for processing large exports.", status: "pending", agent: null, duration: null, depends_on: [34], result: null },
    { id: 38, story_id: 4, name: "Add email delivery option", description: "Send completed exports via email.", status: "pending", agent: null, duration: null, depends_on: [37], result: null },
    { id: 39, story_id: 4, name: "Unit tests for generators", description: "Tests for CSV and PDF generation.", status: "pending", agent: null, duration: null, depends_on: [32, 33], result: null },
    { id: 40, story_id: 4, name: "Integration tests for export API", description: "Integration tests for export endpoints.", status: "pending", agent: null, duration: null, depends_on: [34], result: null },
    { id: 41, story_id: 4, name: "E2E tests for export flow", description: "End-to-end tests for complete export workflows.", status: "pending", agent: null, duration: null, depends_on: [35, 36], result: null },
    { id: 42, story_id: 4, name: "Add export history tracking", description: "Track export history and allow re-downloads.", status: "pending", agent: null, duration: null, depends_on: [34], result: null },
  ],
};

// ── Approvals ────────────────────────────────────────────────────────────────
export const mockApprovals: Approval[] = [
  {
    id: 1,
    epic_id: 1,
    story_name: "Registro de Gastos",
    type: "design",
    title: "Design Approval: Registro de Gastos",
    description: "Technical architecture document for expense tracking. Covers data model, API contracts, and component structure.",
    status: "pending",
    requested_at: "3 min ago",
    content: `# Design: Registro de Gastos

## Technical Approach

SQLite with sqlx for type-safe queries. Repository pattern for data access. HTTP handlers using standard library with chi router. FTS5 virtual table for full-text search.

## Data Model

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| amount | REAL | NOT NULL, CHECK(amount > 0) |
| description | TEXT | NOT NULL, MAX 255 |
| category_id | INTEGER | FK → categories(id), NULLABLE |
| date | TEXT | NOT NULL, ISO 8601 |
| created_at | TEXT | DEFAULT CURRENT_TIMESTAMP |
| updated_at | TEXT | DEFAULT CURRENT_TIMESTAMP |

## API Contracts

\`\`\`
POST   /expenses          → 201 Created
GET    /expenses          → 200 OK (paginated)
GET    /expenses/:id      → 200 OK
PUT    /expenses/:id      → 200 OK
DELETE /expenses/bulk     → 200 OK { deleted: number }
\`\`\`

### Request/Response Examples

\`\`\`json
{
  "amount": 42.50,
  "description": "Almoço no restaurante",
  "category_id": 3,
  "date": "2026-03-09"
}
\`\`\`

## Component Architecture

\`\`\`mermaid
graph TD
    A[HTTP Handler] --> B[Validation Middleware]
    B --> C[Service Layer]
    C --> D[Repository]
    D --> E[(SQLite)]
    C --> F[Currency Formatter]
    C --> G[FTS5 Search]
\`\`\`

## Component Boundaries

- \`domain/expense.go\` — model struct with JSON tags and validation methods
- \`db/expense_repo.go\` — repository with prepared statements and transaction support
- \`db/migrations/\` — numbered SQL migrations (001_expenses, 002_date_index, 003_fts)
- \`cli/api/expenses.go\` — HTTP handlers with request/response DTOs
- \`cli/api/pagination.go\` — cursor-based pagination using (date, id) composite key
- \`cli/api/validation.go\` — input validation middleware
- \`domain/currency.go\` — BRL locale formatting (R$ 1.234,56)

## Pagination Strategy

Cursor-based pagination using composite key \`(date, id)\`:

- [x] Avoids OFFSET performance degradation
- [x] Stable ordering even with concurrent inserts
- [x] Default page size: 20, max: 100
- [ ] Client-side infinite scroll support (future)

## Integration Points

- **Category assignment** (Story 2) will call into expense repository for category updates
- **Dashboard aggregation** (Story 3) will use expense list with date range filters
- **Export** (Story 4) will use expense list with all filter combinations
- Validation middleware needs architectural decision before integration tests can proceed
`,
    previous_content: "",
  },
  {
    id: 2,
    epic_id: 1,
    story_name: "Dashboard de Resumo",
    type: "story-version",
    title: "Story Update: Dashboard de Resumo",
    description: "Story specification with control, business rules, expected behavior, and acceptance criteria needing review.",
    status: "pending",
    requested_at: "12 min ago",
    content: `# Story: Dashboard de Resumo

## Context

Users need a visual overview of their spending patterns. The dashboard provides monthly summaries, category breakdowns, and trend visualization to help users understand where their money goes.

## Business Rules

- Dashboard shows data for the **current month** by default
- Date range can be changed via selector (min: 1 day, max: 1 year)
- All monetary values displayed in **BRL** (R$ 1.234,56)
- Categories with zero spending are hidden from breakdown
- "Uncategorized" expenses grouped under a special category

## Expected Behavior

| Action | Result |
|--------|--------|
| Open dashboard | Shows current month summary |
| Change date range | Reloads all charts and summary |
| Click category slice | Filters expense list by category |
| Hover chart element | Shows tooltip with exact amount |

## API Endpoints

\`\`\`
GET /dashboard/summary?start=2026-03-01&end=2026-03-31
GET /dashboard/by-category?start=2026-03-01&end=2026-03-31
GET /dashboard/trend?months=6
\`\`\`

### Summary Response

\`\`\`json
{
  "total": 3245.80,
  "count": 47,
  "average": 69.06,
  "highest": { "amount": 890.00, "description": "Aluguel" },
  "lowest": { "amount": 2.50, "description": "Café" }
}
\`\`\`

## Acceptance Criteria

- [ ] Monthly summary shows total, count, average, highest, and lowest
- [ ] Category breakdown pie chart renders correctly
- [ ] 6-month trend line chart shows spending over time
- [ ] Date range selector updates all dashboard components
- [ ] Empty state shown when no expenses exist for selected range
- [ ] Loading skeleton displayed while data is being fetched
- [ ] All amounts formatted in BRL locale
`,
    previous_content: "",
  },
  {
    id: 3,
    epic_id: 1,
    story_name: "Categorização Automática",
    type: "qa-escalation",
    title: "QA Escalation: Categorização Automática",
    description: "QA rejected implementation 3 times. Recurring issue with category matching accuracy. Needs human decision.",
    status: "pending",
    requested_at: "25 min ago",
    content: `# QA Escalation: Categorização Automática

## Summary

The category matching algorithm has been rejected by QA **3 consecutive times**. The core issue is accuracy degradation with Portuguese compound words and regional variations.

## Rejection History

| Attempt | Finding | Severity |
|---------|---------|----------|
| 1 | Keyword matcher fails on accented characters (café → not matched) | Major |
| 2 | Fixed accents, but compound words split incorrectly ("cartão de crédito" → 3 separate tokens) | Major |
| 3 | Compound handling improved, but regional variations still fail ("açaí" vs "acai") | Major |

## Root Cause Analysis

The trie-based matcher treats each word independently. Portuguese compound expressions (e.g., "cartão de crédito", "conta de luz", "plano de saúde") need to be matched as single units.

### Current Matching Pipeline

\`\`\`mermaid
flowchart LR
    A[Input Text] --> B[Tokenizer]
    B --> C[Stemmer]
    C --> D[Trie Lookup]
    D --> E{Match?}
    E -->|Yes| F[Return Category]
    E -->|No| G[Historical Fallback]
    G --> H{Match?}
    H -->|Yes| F
    H -->|No| I[Uncategorized]
\`\`\`

## Options for Resolution

### Option A: N-gram matching
Add bigram and trigram matching before single-word lookup.

**Pros:** Simple to implement, handles most compound expressions
**Cons:** Increases memory usage, may create false positives

### Option B: Dictionary-based compound detection
Maintain a dictionary of known Portuguese compound expressions.

**Pros:** Precise matching, no false positives
**Cons:** Requires manual curation, won't handle new expressions

### Option C: Fuzzy matching with Levenshtein distance
Add fuzzy matching as fallback before historical lookup.

**Pros:** Handles regional variations and typos
**Cons:** Slower, needs careful threshold tuning

## Decision Needed

Which approach should we take? This blocks Story 2 completion and downstream Story 3 (dashboard needs accurate categories for breakdown chart).
`,
    previous_content: "",
  },
  {
    id: 4,
    epic_id: 1,
    story_name: null,
    type: "epic-version",
    title: "Epic Update: Gastos Pessoais",
    description: "Epic requirements document with problem definition, objectives, user stories, scope, and dependencies.",
    status: "pending",
    requested_at: "1 hour ago",
    content: `# Epic: Gastos Pessoais

## Problem Definition

Individuals lack a simple, offline-capable tool to track and understand their daily spending. Existing solutions are either too complex (full banking apps) or too simple (plain spreadsheets with no automation).

## Objectives

1. **Record** daily expenses with minimal friction
2. **Categorize** automatically based on description patterns
3. **Visualize** spending trends and category breakdown
4. **Export** reports for personal finance review

## User Stories

| # | Story | Status | Dependencies |
|---|-------|--------|--------------|
| 1 | Registro de Gastos | 🔄 In Progress | None |
| 2 | Categorização Automática | ✅ Completed | Story 1 |
| 3 | Dashboard de Resumo | ⏳ Pending | Stories 1, 2 |
| 4 | Exportação de Relatórios | ⏳ Pending | Stories 1, 2, 3 |

## Scope

### In Scope
- Single-user expense tracking
- BRL currency support
- SQLite local storage
- REST API for all operations
- Automatic categorization with learning

### Out of Scope
- Multi-user / authentication
- Cloud sync
- Mobile app (API-first enables this later)
- Budget planning and alerts
- Bank statement import

## Technical Constraints

- **Runtime:** Go with SQLite (no external DB dependency)
- **Search:** FTS5 for description matching
- **API:** RESTful with chi router
- **Testing:** Table-driven tests, minimum 80% coverage

## Architecture Overview

\`\`\`mermaid
graph TB
    subgraph "CLI Layer"
        A[HTTP Server] --> B[Router]
        B --> C[Expense Handlers]
        B --> D[Category Handlers]
        B --> E[Dashboard Handlers]
        B --> F[Export Handlers]
    end
    subgraph "Domain Layer"
        C --> G[Expense Service]
        D --> H[Category Service]
        E --> I[Dashboard Service]
        F --> J[Export Service]
        G --> K[Currency Formatter]
        H --> L[Keyword Matcher]
    end
    subgraph "Data Layer"
        G --> M[Expense Repo]
        H --> N[Category Repo]
        M --> O[(SQLite)]
        N --> O
    end
\`\`\`

## Success Metrics

- [ ] All 4 stories completed and validated
- [ ] 42 tasks across all stories
- [ ] QA validation passes for each story
- [ ] No major findings in final validation
`,
    previous_content: "",
  },
  {
    id: 5,
    epic_id: 1,
    story_name: "Registro de Gastos",
    type: "execution-plan",
    title: "Plan Approval: Registro de Gastos",
    description: "Execution plan with 12 tasks across 3 agents.",
    status: "approved",
    requested_at: "2 hours ago",
    content: "",
    previous_content: "",
  },
  {
    id: 6,
    epic_id: 1,
    story_name: "Categorização Automática",
    type: "design",
    title: "Design Approval: Categorização Automática",
    description: "Category taxonomy and matching algorithm design.",
    status: "approved",
    requested_at: "3 hours ago",
    content: "",
    previous_content: "",
  },
  {
    id: 7,
    epic_id: 1,
    story_name: "Dashboard de Resumo",
    type: "story-version",
    title: "Story Update: Dashboard de Resumo",
    description: "Initial story definition rejected — too broad.",
    status: "rejected",
    requested_at: "5 hours ago",
    content: "",
    previous_content: "",
  },
];

// ── Story Versions (Requirements) ───────────────────────────────────────────
export const mockStoryVersions: StoryVersion[] = [
  {
    id: 1,
    story_id: 1,
    version: 1,
    type: "requirements",
    created_at: "2026-03-02T09:00:00Z",
    approval_status: "rejected",
    content: `## Context\nUsers need a way to record daily expenses with amount, description, category, and date. The system must support BRL currency with proper formatting.\n\n## Business Rules\n- Amount must be positive with max 2 decimal places\n- Description is required, max 255 characters\n- Date defaults to today if not provided\n- Category is optional (auto-categorization will handle it)\n\n## Expected Behavior\n- POST /expenses creates a new expense record\n- GET /expenses returns paginated list with cursor-based navigation\n- GET /expenses supports date range and search filters\n- DELETE /expenses/bulk removes multiple records in a transaction\n\n## Acceptance Criteria\n- [ ] Expense CRUD operations work correctly\n- [ ] Currency formatting follows BRL locale\n- [ ] Pagination uses cursor-based approach\n- [ ] Search uses FTS5 for description matching\n- [ ] All endpoints return proper error responses`,
  },
  {
    id: 2,
    story_id: 1,
    version: 2,
    type: "requirements",
    created_at: "2026-03-03T11:30:00Z",
    approval_status: "rejected",
    content: `## Context\nUsers need a way to record daily expenses with amount, description, category, and date. The system must support BRL currency with proper formatting. Must handle offline scenarios gracefully.\n\n## Business Rules\n- Amount must be positive with max 2 decimal places\n- Description is required, max 255 characters\n- Date defaults to today if not provided\n- Category is optional (auto-categorization will handle it)\n- Bulk operations limited to 100 items per request\n\n## Expected Behavior\n- POST /expenses creates a new expense record\n- GET /expenses returns paginated list with cursor-based navigation\n- GET /expenses supports date range and search filters\n- DELETE /expenses/bulk removes multiple records in a transaction\n- All mutations return the updated resource\n\n## Acceptance Criteria\n- [ ] Expense CRUD operations work correctly\n- [ ] Currency formatting follows BRL locale (R$ 1.234,56)\n- [ ] Pagination uses cursor-based approach with configurable page size\n- [ ] Search uses FTS5 for description matching with ranking\n- [ ] All endpoints return proper error responses\n- [ ] Input validation rejects malformed requests with 422`,
  },
  {
    id: 3,
    story_id: 1,
    version: 3,
    type: "requirements",
    created_at: "2026-03-04T14:00:00Z",
    approval_status: "approved",
    content: `## Context\nUsers need a way to record daily expenses with amount, description, category, and date. The system must support BRL currency with proper formatting. The expense module is the foundation for categorization and reporting stories.\n\n## Business Rules\n- Amount must be positive with max 2 decimal places\n- Description is required, max 255 characters\n- Date defaults to today if not provided\n- Category is optional (auto-categorization handled by Story 2)\n- Bulk operations limited to 100 items per request\n- Cursor-based pagination with default page size of 20\n\n## Expected Behavior\n- POST /expenses — creates expense, returns 201 with resource\n- GET /expenses — paginated list with cursor, date range, and search filters\n- GET /expenses/:id — single expense by ID\n- PUT /expenses/:id — update expense, returns 200\n- DELETE /expenses/bulk — transactional bulk delete, returns count\n\n## Acceptance Criteria\n- [ ] Expense CRUD operations work correctly\n- [ ] Currency formatting follows BRL locale (R$ 1.234,56)\n- [ ] Pagination uses cursor-based approach (date, id) with configurable page size\n- [ ] Search uses FTS5 for description matching with ranking\n- [ ] All endpoints return structured error responses (code, message, details)\n- [ ] Input validation rejects malformed requests with 422\n- [ ] Bulk delete is transactional — all or nothing`,
  },
];

// ── Story Versions (Design) ─────────────────────────────────────────────────
export const mockDesignVersions: StoryVersion[] = [
  {
    id: 4,
    story_id: 1,
    version: 1,
    type: "design",
    created_at: "2026-03-05T10:00:00Z",
    approval_status: "approved",
    content: `## Technical Approach\nSQLite with sqlx for type-safe queries. Repository pattern for data access. HTTP handlers using standard library with chi router.\n\n## Component Boundaries\n- domain/expense.go — model + validation\n- db/expense_repo.go — repository with prepared statements\n- cli/api/expenses.go — HTTP handlers\n- cli/api/pagination.go — cursor-based pagination helper\n- domain/currency.go — BRL formatting\n\n## Integration Points\n- Category assignment (Story 2) will call into expense repository\n- Dashboard aggregation (Story 3) will use expense list with date filters\n- Export (Story 4) will use expense list with all filters`,
  },
  {
    id: 5,
    story_id: 1,
    version: 2,
    type: "design",
    created_at: "2026-03-06T16:00:00Z",
    approval_status: "pending",
    content: `## Technical Approach\nSQLite with sqlx for type-safe queries. Repository pattern for data access. HTTP handlers using standard library with chi router. FTS5 virtual table for full-text search.\n\n## Component Boundaries\n- domain/expense.go — model struct with JSON tags and validation methods\n- db/expense_repo.go — repository with prepared statements and transaction support\n- db/migrations/ — numbered SQL migrations (001_expenses, 002_date_index, 003_fts)\n- cli/api/expenses.go — HTTP handlers with request/response DTOs\n- cli/api/pagination.go — cursor-based pagination using (date, id) composite key\n- cli/api/validation.go — input validation middleware (BLOCKED: conflicts with error handler)\n- domain/currency.go — BRL locale formatting (R$ 1.234,56)\n\n## Integration Points\n- Category assignment (Story 2) will call into expense repository for category updates\n- Dashboard aggregation (Story 3) will use expense list with date range filters\n- Export (Story 4) will use expense list with all filter combinations\n- Validation middleware needs architectural decision before integration tests can proceed`,
  },
];

// ── Reviews ─────────────────────────────────────────────────────────────────
export const mockReviews: Review[] = [
  {
    id: 1,
    version_id: 1,
    verdict: "rejected",
    comment: "Missing error response format specification. Need to define the structure for 4xx/5xx responses consistently across all endpoints.",
    created_at: "2026-03-02T15:30:00Z",
  },
  {
    id: 2,
    version_id: 2,
    verdict: "rejected",
    comment: "Good improvements on error format. However, the bulk operation limit and pagination details need to be explicit in acceptance criteria, not just business rules.",
    created_at: "2026-03-03T16:45:00Z",
  },
  {
    id: 3,
    version_id: 3,
    verdict: "approved",
    comment: "All feedback addressed. Clear acceptance criteria, explicit API contracts, and proper scope boundaries with other stories.",
    created_at: "2026-03-04T17:00:00Z",
  },
];

// ── QA Validations ──────────────────────────────────────────────────────────
export const mockValidations: Validation[] = [
  {
    id: 1,
    story_id: 1,
    verdict: "approved",
    findings: "CRUD operations pass all acceptance criteria. Currency formatting correct for BRL. Pagination cursor works with composite key.",
    created_at: "2026-03-07T10:00:00Z",
  },
  {
    id: 2,
    story_id: 1,
    verdict: "rejected",
    findings: "Input validation middleware is not integrated — endpoint accepts negative amounts and empty descriptions without error. Task 4 (validation rules) failed and was not retried.",
    created_at: "2026-03-07T14:30:00Z",
  },
  {
    id: 3,
    story_id: 1,
    verdict: "approved",
    findings: "FTS5 search returns ranked results correctly. Tested with Portuguese accented characters and compound words.",
    created_at: "2026-03-08T09:15:00Z",
  },
  {
    id: 4,
    story_id: 1,
    verdict: "rejected",
    findings: "Bulk delete endpoint does not validate array size — sending 200+ IDs causes timeout. Business rule says max 100, but no server-side enforcement.",
    created_at: "2026-03-08T11:45:00Z",
  },
  {
    id: 5,
    story_id: 1,
    verdict: "approved",
    findings: "Date range filtering uses index correctly. Tested boundary conditions (same day, empty range, future dates). Pagination integrates properly with filters.",
    created_at: "2026-03-08T16:00:00Z",
  },
];

// ── Helpers ─────────────────────────────────────────────────────────────────
export function getEpic(): EpicSummary {
  return mockEpic;
}

export function getStories(epicId: number): Story[] {
  return mockStories.filter((s) => s.epic_id === epicId);
}

export function getStory(storyId: number): Story | undefined {
  return mockStories.find((s) => s.id === storyId);
}

export function getStoryDetail(storyId: number): StoryDetail | undefined {
  const story = mockStories.find((s) => s.id === storyId);
  if (!story) return undefined;
  const approvalMap: Record<number, StoryDetail["approval_status"]> = {
    1: "pending",
    2: "approved",
    3: "none",
    4: "none",
  };
  return { ...story, approval_status: approvalMap[storyId] ?? "none" };
}

export function getTasks(storyId: number): StoryTask[] {
  return mockTasks[storyId] ?? [];
}

export function getStoryVersions(storyId: number, type: "requirements" | "design"): StoryVersion[] {
  const source = type === "requirements" ? mockStoryVersions : mockDesignVersions;
  return source.filter((v) => v.story_id === storyId).sort((a, b) => b.version - a.version);
}

export function getReviewsForVersion(versionId: number): Review[] {
  return mockReviews.filter((r) => r.version_id === versionId);
}

export function getReviewsForStory(storyId: number): Review[] {
  const versionIds = [...mockStoryVersions, ...mockDesignVersions]
    .filter((v) => v.story_id === storyId)
    .map((v) => v.id);
  return mockReviews.filter((r) => versionIds.includes(r.version_id));
}

export function getValidations(storyId: number): Validation[] {
  return mockValidations.filter((v) => v.story_id === storyId);
}

export function getApproval(approvalId: number): Approval | undefined {
  return mockApprovals.find((a) => a.id === approvalId);
}

export function getApprovals(epicId: number): Approval[] {
  return mockApprovals.filter((a) => a.epic_id === epicId);
}

export function getPendingApprovals(epicId: number): Approval[] {
  return mockApprovals.filter((a) => a.epic_id === epicId && a.status === "pending");
}

export function getResolvedApprovals(epicId: number): Approval[] {
  return mockApprovals.filter((a) => a.epic_id === epicId && a.status !== "pending");
}
