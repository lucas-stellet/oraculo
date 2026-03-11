// Wails v3 service bindings — only functional inside Wails runtime.
// In browser/dev, calls will throw; callers must guard accordingly.

export type { KnownProject, ProjectWithStatus } from "./types";

// Dynamic imports so Next.js doesn't try to bundle @wailsio/runtime for SSR.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let _launcher: any = null;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let _server: any = null;

async function launcher() {
  if (!_launcher) _launcher = await import("./launcherservice.js");
  return _launcher;
}

async function server() {
  if (!_server) _server = await import("./serverservice.js");
  return _server;
}

export const LauncherService = {
  ListProjects: () => launcher().then((m) => m.ListProjects()),
  AddProject: () => launcher().then((m) => m.AddProject()),
  RemoveProject: (path: string) => launcher().then((m) => m.RemoveProject(path)),
  StartServer: (path: string) => launcher().then((m) => m.StartServer(path)),
  StopServer: (path: string) => launcher().then((m) => m.StopServer(path)),
};

export const ServerService = {
  SelectServer: (port: number): Promise<string> => server().then((m) => m.SelectServer(port)),
  GetCurrentServer: (): Promise<string> => server().then((m) => m.GetCurrentServer()),
};

export function isWailsRuntime(): boolean {
  return typeof window !== "undefined" && "_wails" in window;
}
