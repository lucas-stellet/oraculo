import type {
  EpicSummary, Story, StoryTask, StoryVersion,
  Review, Validation, Approval, InlineComment,
} from "./types";

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(path);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${path}`);
  }
  return res.json();
}

export const api = {
  listEpics: () =>
    fetchJSON<EpicSummary[]>("/api/epics"),

  createEpic: (name: string, description: string) =>
    fetch("/api/epics", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, description }),
    }).then((r) => r.json()),

  listStories: (epicName: string) =>
    fetchJSON<Story[]>(`/api/epics/${encodeURIComponent(epicName)}/stories`),

  listTasks: (epicName: string, storyName: string) =>
    fetchJSON<StoryTask[]>(
      `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/tasks`
    ),

  listStoryVersions: (epicName: string, storyName: string) =>
    fetchJSON<StoryVersion[]>(
      `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/versions`
    ),

  listStoryReviews: (epicName: string, storyName: string) =>
    fetchJSON<Review[]>(
      `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/reviews`
    ),

  listValidations: (epicName: string, storyName: string) =>
    fetchJSON<Validation[]>(
      `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/validations`
    ),

  listApprovals: (epicId?: number, status?: string) => {
    const params = new URLSearchParams();
    if (epicId) params.set("epic_id", String(epicId));
    if (status) params.set("status", status);
    const qs = params.toString();
    return fetchJSON<Approval[]>(`/api/approvals${qs ? `?${qs}` : ""}`);
  },

  getApproval: (id: string) =>
    fetchJSON<Approval>(`/api/approvals/${id}`),

  submitVerdict: (id: string, verdict: string, comment: string) =>
    fetch(`/api/approvals/${id}/verdict`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ verdict, comment }),
    }).then((r) => r.json()),

  createComment: (approvalId: string, selectedText: string, comment: string) =>
    fetch(`/api/approvals/${approvalId}/comments`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ selected_text: selectedText, comment }),
    }).then((r) => r.json()),

  listComments: (approvalId: string) =>
    fetchJSON<InlineComment[]>(`/api/approvals/${approvalId}/comments`),

  deleteComment: (approvalId: string, commentId: number) =>
    fetch(`/api/approvals/${approvalId}/comments/${commentId}`, {
      method: "DELETE",
    }),

  getSystemStatus: () =>
    fetchJSON<{ update_available: boolean; started_at: string; version: string; project_commit: string; new_version: string }>("/api/system/status"),

  restartServer: () =>
    fetch("/api/system/restart", { method: "POST" }).then((r) => r.json()),
};
