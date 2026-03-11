"use client";

import { createContext, useContext, useMemo, useState, useEffect } from "react";
import { createApi } from "./api";
import { isWailsRuntime, ServerService } from "./wails";

interface ServerContextValue {
  baseUrl: string;
  setBaseUrl: (url: string) => void;
}

const ServerContext = createContext<ServerContextValue>({
  baseUrl: "",
  setBaseUrl: () => {},
});

export function ServerProvider({
  baseUrl: initialBaseUrl,
  children,
}: {
  baseUrl?: string;
  children: React.ReactNode;
}) {
  const [baseUrl, setBaseUrl] = useState(
    initialBaseUrl || (typeof window !== "undefined" ? window.location.origin : "")
  );

  // In Wails, pick up the selected server URL from Go-side state.
  useEffect(() => {
    if (!isWailsRuntime()) return;
    ServerService.GetCurrentServer().then((url) => {
      if (url) setBaseUrl(url);
    }).catch(() => {});
  }, []);

  const value = useMemo(() => ({ baseUrl, setBaseUrl }), [baseUrl]);

  return (
    <ServerContext.Provider value={value}>{children}</ServerContext.Provider>
  );
}

export function useServerUrl(): string {
  return useContext(ServerContext).baseUrl;
}

export function useSetServerUrl(): (url: string) => void {
  return useContext(ServerContext).setBaseUrl;
}

export function useApi() {
  const baseUrl = useServerUrl();
  return useMemo(() => createApi(baseUrl), [baseUrl]);
}
