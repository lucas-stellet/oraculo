"use client";

import { createContext, useContext, useMemo } from "react";
import { createApi } from "./api";

interface ServerContextValue {
  baseUrl: string;
}

const ServerContext = createContext<ServerContextValue>({
  baseUrl: "",
});

export function ServerProvider({
  baseUrl,
  children,
}: {
  baseUrl?: string;
  children: React.ReactNode;
}) {
  const value = useMemo(
    () => ({
      baseUrl: baseUrl || (typeof window !== "undefined" ? window.location.origin : ""),
    }),
    [baseUrl]
  );

  return (
    <ServerContext.Provider value={value}>{children}</ServerContext.Provider>
  );
}

export function useServerUrl(): string {
  return useContext(ServerContext).baseUrl;
}

export function useApi() {
  const baseUrl = useServerUrl();
  return useMemo(() => createApi(baseUrl), [baseUrl]);
}
