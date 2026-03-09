import type { EpicSummary, Story, StoryTask } from "./types";

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
    { id: 1, story_id: 1, name: "Create expense model schema", status: "completed", agent: "code-01", duration: "3m 42s" },
    { id: 2, story_id: 1, name: "Implement repository layer", status: "completed", agent: "code-02", duration: "5m 18s" },
    { id: 3, story_id: 1, name: "Build REST API endpoints", status: "completed", agent: "code-01", duration: "4m 07s" },
    { id: 4, story_id: 1, name: "Write input validation rules", status: "in_progress", agent: "code-03", duration: "1m 22s" },
    { id: 5, story_id: 1, name: "Create unit tests for model", status: "pending", agent: null, duration: null },
    { id: 6, story_id: 1, name: "Integration tests for API", status: "pending", agent: null, duration: null },
    { id: 7, story_id: 1, name: "Add currency formatting", status: "completed", agent: "code-01", duration: "2m 15s" },
    { id: 8, story_id: 1, name: "Implement date range filtering", status: "completed", agent: "code-02", duration: "3m 50s" },
    { id: 9, story_id: 1, name: "Build expense list pagination", status: "completed", agent: "code-01", duration: "4m 33s" },
    { id: 10, story_id: 1, name: "Add bulk delete endpoint", status: "completed", agent: "code-03", duration: "2m 58s" },
    { id: 11, story_id: 1, name: "Implement search by description", status: "completed", agent: "code-02", duration: "1m 45s" },
    { id: 12, story_id: 1, name: "E2E tests for expense flow", status: "pending", agent: null, duration: null },
  ],
  2: [
    { id: 13, story_id: 2, name: "Define category taxonomy", status: "completed", agent: "research-01", duration: "6m 12s" },
    { id: 14, story_id: 2, name: "Implement keyword matcher", status: "completed", agent: "code-01", duration: "4m 30s" },
    { id: 15, story_id: 2, name: "Build category assignment logic", status: "completed", agent: "code-02", duration: "5m 44s" },
    { id: 16, story_id: 2, name: "Create manual override API", status: "completed", agent: "code-01", duration: "3m 10s" },
    { id: 17, story_id: 2, name: "Add category CRUD endpoints", status: "completed", agent: "code-03", duration: "2m 55s" },
    { id: 18, story_id: 2, name: "Implement learning from overrides", status: "completed", agent: "code-02", duration: "7m 20s" },
    { id: 19, story_id: 2, name: "Unit tests for matcher", status: "completed", agent: "code-01", duration: "3m 15s" },
    { id: 20, story_id: 2, name: "Unit tests for assignment logic", status: "completed", agent: "code-02", duration: "2m 48s" },
    { id: 21, story_id: 2, name: "Integration tests for override flow", status: "completed", agent: "code-03", duration: "4m 02s" },
    { id: 22, story_id: 2, name: "Add confidence scoring", status: "completed", agent: "code-01", duration: "5m 30s" },
    { id: 23, story_id: 2, name: "Build batch categorization endpoint", status: "completed", agent: "code-02", duration: "3m 40s" },
    { id: 24, story_id: 2, name: "E2E tests for categorization", status: "completed", agent: "code-03", duration: "5m 55s" },
  ],
  3: [
    { id: 25, story_id: 3, name: "Design dashboard layout", status: "pending", agent: null, duration: null },
    { id: 26, story_id: 3, name: "Implement monthly summary aggregation", status: "pending", agent: null, duration: null },
    { id: 27, story_id: 3, name: "Build chart data endpoints", status: "pending", agent: null, duration: null },
    { id: 28, story_id: 3, name: "Create category breakdown view", status: "pending", agent: null, duration: null },
    { id: 29, story_id: 3, name: "Add date range selector", status: "pending", agent: null, duration: null },
    { id: 30, story_id: 3, name: "Unit tests for aggregation", status: "pending", agent: null, duration: null },
  ],
  4: [
    { id: 31, story_id: 4, name: "Define export formats (CSV, PDF)", status: "pending", agent: null, duration: null },
    { id: 32, story_id: 4, name: "Implement CSV generator", status: "pending", agent: null, duration: null },
    { id: 33, story_id: 4, name: "Implement PDF generator", status: "pending", agent: null, duration: null },
    { id: 34, story_id: 4, name: "Build export API endpoint", status: "pending", agent: null, duration: null },
    { id: 35, story_id: 4, name: "Add date range filter for exports", status: "pending", agent: null, duration: null },
    { id: 36, story_id: 4, name: "Implement category filter for exports", status: "pending", agent: null, duration: null },
    { id: 37, story_id: 4, name: "Create export queue system", status: "pending", agent: null, duration: null },
    { id: 38, story_id: 4, name: "Add email delivery option", status: "pending", agent: null, duration: null },
    { id: 39, story_id: 4, name: "Unit tests for generators", status: "pending", agent: null, duration: null },
    { id: 40, story_id: 4, name: "Integration tests for export API", status: "pending", agent: null, duration: null },
    { id: 41, story_id: 4, name: "E2E tests for export flow", status: "pending", agent: null, duration: null },
    { id: 42, story_id: 4, name: "Add export history tracking", status: "pending", agent: null, duration: null },
  ],
};

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

export function getTasks(storyId: number): StoryTask[] {
  return mockTasks[storyId] ?? [];
}
