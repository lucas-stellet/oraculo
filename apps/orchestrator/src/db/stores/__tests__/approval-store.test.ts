import { describe, test, expect, beforeEach } from "bun:test";
import type { Database } from "bun:sqlite";
import { openMemory } from "../../database.ts";
import { ApprovalStore } from "../approval-store.ts";
import { DomainError } from "../../../domain/errors.ts";

describe("ApprovalStore", () => {
  let db: Database;
  let store: ApprovalStore;

  beforeEach(() => {
    db = openMemory();
    store = new ApprovalStore(db);
    db.run("INSERT INTO epics (name, description) VALUES ('test-epic', 'desc')");
  });

  test("Create creates a pending approval with UUID", () => {
    const a = store.Create("qa-escalation", 1, null, "some content");
    expect(a.id).toBeTruthy();
    expect(a.id).toContain("-"); // UUID format
    expect(a.type).toBe("qa-escalation");
    expect(a.epic_id).toBe(1);
    expect(a.story_id).toBeNull();
    expect(a.content).toBe("some content");
    expect(a.status).toBe("pending");
    expect(a.requested_at).toBeTruthy();
  });

  test("Get retrieves an approval by id", () => {
    const created = store.Create("execution-plan", 1, null, "plan content");
    const fetched = store.Get(created.id);
    expect(fetched.id).toBe(created.id);
    expect(fetched.content).toBe("plan content");
  });

  test("Get throws not_found for missing id", () => {
    expect(() => store.Get("nonexistent")).toThrow(DomainError);
  });

  test("List returns all approvals", () => {
    store.Create("qa-escalation", 1, null, "content 1");
    store.Create("execution-plan", 1, null, "content 2");
    const all = store.List();
    expect(all).toHaveLength(2);
  });

  test("List filters by epicId", () => {
    db.run("INSERT INTO epics (name, description) VALUES ('epic-2', 'desc')");
    store.Create("qa-escalation", 1, null, "content 1");
    store.Create("execution-plan", 2, null, "content 2");

    const forEpic1 = store.List(1);
    expect(forEpic1).toHaveLength(1);
    expect(forEpic1[0].epic_id).toBe(1);
  });

  test("ListPending returns only pending approvals", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    store.Create("execution-plan", 1, null, "content 2");
    store.SetVerdict(a.id, "approved");

    const pending = store.ListPending();
    expect(pending).toHaveLength(1);
    expect(pending[0].status).toBe("pending");
  });

  test("SetVerdict updates status and decided_at", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    store.SetVerdict(a.id, "approved", "looks good");
    const updated = store.Get(a.id);
    expect(updated.status).toBe("approved");
    expect(updated.verdict_comment).toBe("looks good");
    expect(updated.decided_at).toBeTruthy();
  });

  test("SetVerdict throws when approval already decided", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    store.SetVerdict(a.id, "approved");
    expect(() => store.SetVerdict(a.id, "rejected")).toThrow(DomainError);
  });

  test("SetVerdict with needs_revision copies content to previous_version", () => {
    const a = store.Create("qa-escalation", 1, null, "original content");
    store.SetVerdict(a.id, "needs_revision", "fix this");
    const updated = store.Get(a.id);
    expect(updated.status).toBe("needs_revision");
    expect(updated.previous_version).toBe("original content");
  });

  test("SetVerdict with approved deletes comments", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    store.AddComment(a.id, "selected text", "my comment");
    expect(store.ListComments(a.id)).toHaveLength(1);

    store.SetVerdict(a.id, "approved");
    expect(store.ListComments(a.id)).toHaveLength(0);
  });

  test("AddComment creates and returns a comment", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    const c = store.AddComment(a.id, "some text", "fix this part");
    expect(c.id).toBeGreaterThan(0);
    expect(c.selected_text).toBe("some text");
    expect(c.comment).toBe("fix this part");
    expect(c.created_at).toBeTruthy();
  });

  test("ListComments returns comments ordered by created_at", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    store.AddComment(a.id, "text1", "comment1");
    store.AddComment(a.id, "text2", "comment2");
    const comments = store.ListComments(a.id);
    expect(comments).toHaveLength(2);
    expect(comments[0].comment).toBe("comment1");
    expect(comments[1].comment).toBe("comment2");
  });

  test("DeleteCommentsByApproval removes all comments for an approval", () => {
    const a = store.Create("qa-escalation", 1, null, "content");
    store.AddComment(a.id, "t1", "c1");
    store.AddComment(a.id, "t2", "c2");
    store.DeleteCommentsByApproval(a.id);
    expect(store.ListComments(a.id)).toHaveLength(0);
  });
});
