"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { LauncherService, ServerService, isWailsRuntime } from "@/lib/wails";
import { useSetServerUrl } from "@/lib/server-context";
import type { ProjectWithStatus } from "@/lib/wails";

// ── icons ────────────────────────────────────────────────────────────────────

function FolderIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
    </svg>
  );
}

function PlusIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  );
}

function ScanEyeIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <path d="M3 7V5a2 2 0 0 1 2-2h2" />
      <path d="M17 3h2a2 2 0 0 1 2 2v2" />
      <path d="M21 17v2a2 2 0 0 1-2 2h-2" />
      <path d="M7 21H5a2 2 0 0 1-2-2v-2" />
      <circle cx="12" cy="12" r="1" fill="currentColor" />
      <path d="M18.944 12.33a1 1 0 0 0 0-.66 7.5 7.5 0 0 0-13.888 0 1 1 0 0 0 0 .66 7.5 7.5 0 0 0 13.888 0" />
    </svg>
  );
}

// ── component ─────────────────────────────────────────────────────────────────

export default function LauncherPage() {
  const router = useRouter();
  const setServerUrl = useSetServerUrl();
  const [projects, setProjects] = useState<ProjectWithStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<Record<string, boolean>>({});
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!isWailsRuntime()) {
      setLoading(false);
      return;
    }
    try {
      const list = await LauncherService.ListProjects();
      setProjects(list ?? []);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  async function handleOpen(project: ProjectWithStatus) {
    if (!project.port) return;
    setBusy((b) => ({ ...b, [project.path]: true }));
    try {
      const url = await ServerService.SelectServer(project.port);
      setServerUrl(url);
      router.push("/epics");
    } finally {
      setBusy((b) => ({ ...b, [project.path]: false }));
    }
  }

  async function handleStart(project: ProjectWithStatus) {
    setBusy((b) => ({ ...b, [project.path]: true }));
    try {
      await LauncherService.StartServer(project.path);
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy((b) => ({ ...b, [project.path]: false }));
    }
  }

  async function handleStop(project: ProjectWithStatus) {
    setBusy((b) => ({ ...b, [project.path]: true }));
    try {
      await LauncherService.StopServer(project.path);
      await refresh();
    } catch (e) {
      setError(String(e));
    } finally {
      setBusy((b) => ({ ...b, [project.path]: false }));
    }
  }

  async function handleRemove(project: ProjectWithStatus) {
    setBusy((b) => ({ ...b, [project.path]: true }));
    try {
      await LauncherService.RemoveProject(project.path);
      await refresh();
    } finally {
      setBusy((b) => ({ ...b, [project.path]: false }));
    }
  }

  async function handleAdd() {
    try {
      await LauncherService.AddProject();
      await refresh();
    } catch (e) {
      setError(String(e));
    }
  }

  return (
    <div className="flex flex-col min-h-screen bg-background">
      {/* Header */}
      <header className="flex items-center gap-3 px-8 py-6 border-b border-border">
        <div className="flex items-center justify-center w-9 h-9 rounded-xl bg-blue-600">
          <ScanEyeIcon className="w-5 h-5 text-white" />
        </div>
        <span className="text-xl font-semibold font-[family-name:var(--font-display)] tracking-tight">
          Oraculo
        </span>
      </header>

      {/* Content */}
      <main className="flex-1 px-8 py-8 max-w-2xl mx-auto w-full">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
            Projects
          </h2>
          <button
            onClick={handleAdd}
            className="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg border border-border hover:bg-accent transition-colors"
          >
            <PlusIcon className="w-3.5 h-3.5" />
            Add Project
          </button>
        </div>

        {error && (
          <div className="mb-4 px-4 py-3 rounded-lg bg-destructive/10 text-destructive text-sm">
            {error}
            <button
              onClick={() => setError(null)}
              className="ml-2 underline"
            >
              dismiss
            </button>
          </div>
        )}

        {loading ? (
          <div className="flex items-center justify-center py-20 text-muted-foreground text-sm">
            Loading…
          </div>
        ) : projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 gap-4 text-center">
            <FolderIcon className="w-12 h-12 text-muted-foreground/40" />
            <p className="text-muted-foreground text-sm max-w-xs">
              No projects yet. Add an Oraculo project to get started.
            </p>
            <button
              onClick={handleAdd}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
            >
              <PlusIcon className="w-4 h-4" />
              Add Project
            </button>
          </div>
        ) : (
          <ul className="space-y-2">
            {projects.map((project) => (
              <li
                key={project.path}
                className="flex items-center gap-4 px-4 py-3 rounded-xl border border-border bg-card"
              >
                {/* Status dot */}
                <span
                  className={`w-2 h-2 rounded-full flex-shrink-0 ${
                    project.online ? "bg-green-500" : "bg-muted-foreground/30"
                  }`}
                />

                {/* Name + path */}
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-sm truncate">{project.name}</p>
                  <p className="text-xs text-muted-foreground truncate">{project.path}</p>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-2 flex-shrink-0">
                  {project.online ? (
                    <>
                      <button
                        onClick={() => handleOpen(project)}
                        disabled={busy[project.path]}
                        className="px-3 py-1.5 rounded-lg bg-primary text-primary-foreground text-xs font-medium hover:bg-primary/90 disabled:opacity-50 transition-colors"
                      >
                        Open Dashboard
                      </button>
                      <button
                        onClick={() => handleStop(project)}
                        disabled={busy[project.path]}
                        className="px-3 py-1.5 rounded-lg border border-border text-xs hover:bg-accent disabled:opacity-50 transition-colors"
                      >
                        Stop
                      </button>
                    </>
                  ) : (
                    <>
                      <button
                        onClick={() => handleStart(project)}
                        disabled={busy[project.path]}
                        className="px-3 py-1.5 rounded-lg border border-border text-xs hover:bg-accent disabled:opacity-50 transition-colors"
                      >
                        {busy[project.path] ? "Starting…" : "Start"}
                      </button>
                      <button
                        onClick={() => handleRemove(project)}
                        disabled={busy[project.path]}
                        className="px-3 py-1.5 rounded-lg text-xs text-muted-foreground hover:text-destructive disabled:opacity-50 transition-colors"
                      >
                        Remove
                      </button>
                    </>
                  )}
                </div>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  );
}
