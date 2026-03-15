import { describe, it, expect } from "bun:test";
import { nextWave } from "../dag.ts";
import type { TaskEnriched } from "../../domain/index.ts";

function makeTask(overrides: Partial<TaskEnriched> & { id: number }): TaskEnriched {
  return {
    story_id: 1,
    name: `task-${overrides.id}`,
    description: "",
    status: "pending",
    failure_reason: "",
    depends_on: [],
    created_at: "",
    updated_at: "",
    ...overrides,
  };
}

describe("nextWave", () => {
  it("returns tasks with no dependencies", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1 }),
      makeTask({ id: 2, depends_on: [1] }),
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks.map((t) => t.id)).toEqual([1]);
  });

  it("returns tasks whose deps are completed", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1, status: "completed" }),
      makeTask({ id: 2, depends_on: [1] }),
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks.map((t) => t.id)).toEqual([2]);
  });

  it("returns empty wave when deadlocked", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1, depends_on: [2] }),
      makeTask({ id: 2, depends_on: [1] }),
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks).toHaveLength(0);
  });

  it("returns multiple independent tasks", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1 }),
      makeTask({ id: 2 }),
      makeTask({ id: 3, depends_on: [1, 2] }),
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks.map((t) => t.id)).toEqual([1, 2]);
  });

  it("skips in_progress tasks", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1, status: "in_progress" }),
      makeTask({ id: 2, depends_on: [1] }),
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks).toHaveLength(0);
  });

  it("skips failed tasks as dependencies", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1, status: "failed" }),
      makeTask({ id: 2, depends_on: [1] }),
    ];
    const wave = nextWave(tasks);
    // task 2 depends on 1, which is failed (not completed), so not ready
    expect(wave.tasks).toHaveLength(0);
  });

  it("returns empty wave when all tasks are completed", () => {
    const tasks: TaskEnriched[] = [
      makeTask({ id: 1, status: "completed" }),
      makeTask({ id: 2, status: "completed", depends_on: [1] }),
    ];
    const wave = nextWave(tasks);
    expect(wave.tasks).toHaveLength(0);
  });

  it("handles empty task list", () => {
    const wave = nextWave([]);
    expect(wave.tasks).toHaveLength(0);
  });
});
