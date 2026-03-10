"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
} from "react";

export interface WSEvent {
  event: string;
  data?: unknown;
  id?: string;
}

type Handler = (evt: WSEvent) => void;

interface WSContextValue {
  subscribe: (handler: Handler) => () => void;
}

const WSContext = createContext<WSContextValue | null>(null);

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const handlersRef = useRef<Set<Handler>>(new Set());
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // Guard: prevents scheduling reconnects after unmount.
  const mountedRef = useRef(true);

  const connect = useCallback(() => {
    // Determine WS URL: same origin, /ws path.
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${protocol}//${window.location.host}/ws`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (msg) => {
      try {
        const evt: WSEvent = JSON.parse(msg.data);
        handlersRef.current.forEach((h) => h(evt));
      } catch {
        // Ignore malformed messages
      }
    };

    ws.onclose = () => {
      // ws.close() in cleanup fires onclose asynchronously — guard against
      // scheduling reconnects after the component has unmounted.
      if (!mountedRef.current) return;
      reconnectTimerRef.current = setTimeout(connect, 2000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    connect();
    return () => {
      mountedRef.current = false;
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
      wsRef.current?.close();
    };
  }, [connect]);

  const subscribe = useCallback((handler: Handler) => {
    handlersRef.current.add(handler);
    return () => {
      handlersRef.current.delete(handler);
    };
  }, []);

  // Memoize context value to prevent spurious re-subscriptions when the
  // Provider re-renders (a new object literal would change ctx identity).
  const value = useMemo(() => ({ subscribe }), [subscribe]);

  return (
    <WSContext.Provider value={value}>
      {children}
    </WSContext.Provider>
  );
}

// useWebSocket registers a handler for incoming WS events.
// The handler is called for every event received while the component is mounted.
// Re-renders do NOT re-subscribe — use a stable callback (useCallback) if needed.
export function useWebSocket(handler: Handler) {
  const ctx = useContext(WSContext);
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    if (!ctx) return;
    // Wrap in stable ref so the subscription never changes identity.
    const stable: Handler = (evt) => handlerRef.current(evt);
    return ctx.subscribe(stable);
  }, [ctx]);
}
