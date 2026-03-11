import type {
  EpicSummary, Story, StoryTask, StoryVersion,
  Review, Validation, Approval, InlineComment,
} from "./types";

async function fetchJSON<T>(baseUrl: string, path: string): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${path}`);
  }
  return res.json();
}

export function createApi(baseUrl: string) {
  return {
    listEpics: () =>
      fetchJSON<EpicSummary[]>(baseUrl, "/api/epics"),

    createEpic: (name: string, description: string) =>
      fetch(`${baseUrl}/api/epics`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description }),
      }).then((r) => r.json()),

    listStories: (epicName: string) =>
      fetchJSON<Story[]>(baseUrl, `/api/epics/${encodeURIComponent(epicName)}/stories`),

    listTasks: (epicName: string, storyName: string) =>
      fetchJSON<StoryTask[]>(
        baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/tasks`
      ),

    listStoryVersions: (epicName: string, storyName: string) =>
      fetchJSON<StoryVersion[]>(
        baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/versions`
      ),

    listStoryReviews: (epicName: string, storyName: string) =>
      fetchJSON<Review[]>(
        baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/reviews`
      ),

    listValidations: (epicName: string, storyName: string) =>
      fetchJSON<Validation[]>(
        baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/validations`
      ),

    listApprovals: (epicId?: number, status?: string) => {
      const params = new URLSearchParams();
      if (epicId) params.set("epic_id", String(epicId));
      if (status) params.set("status", status);
      const qs = params.toString();
      return fetchJSON<Approval[]>(baseUrl, `/api/approvals${qs ? `?${qs}` : ""}`);
    },

    getApproval: (id: string) =>
      fetchJSON<Approval>(baseUrl, `/api/approvals/${id}`),

    submitVerdict: (id: string, verdict: string, comment: string) =>
      fetch(`${baseUrl}/api/approvals/${id}/verdict`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ verdict, comment }),
      }).then((r) => r.json()),

    createComment: (approvalId: string, selectedText: string, comment: string) =>
      fetch(`${baseUrl}/api/approvals/${approvalId}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ selected_text: selectedText, comment }),
      }).then((r) => r.json()),

    listComments: (approvalId: string) =>
      fetchJSON<InlineComment[]>(baseUrl, `/api/approvals/${approvalId}/comments`),

    deleteComment: (approvalId: string, commentId: number) =>
      fetch(`${baseUrl}/api/approvals/${approvalId}/comments/${commentId}`, {
        method: "DELETE",
      }),

    getSystemStatus: () =>
      fetchJSON<{ update_available: boolean; started_at: string; version: string; project_commit: string; new_version: string }>(
        baseUrl,
        "/api/system/status"
      ),
  };
}

export const api = createApi("");
