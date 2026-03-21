import { useEffect, useCallback, useRef } from "react";
import { useWsStore } from "@/stores/ws-store";
import { useAuthStore } from "@/stores/auth-store";

type EventHandler = (event: string, payload: unknown) => void;

/**
 * useWebSocket connects to the WebSocket gateway when authenticated,
 * manages room subscriptions, and provides an event listener interface.
 */
export function useWebSocket() {
  const accessToken = useAuthStore((s) => s.accessToken);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const { connected, connect, disconnect, subscribe, unsubscribe } =
    useWsStore();
  const addListener = useWsStore((s) => s.addListener);
  const removeListener = useWsStore((s) => s.removeListener);

  // Connect when authenticated, disconnect on unmount or logout.
  useEffect(() => {
    if (isAuthenticated && accessToken) {
      connect(accessToken);
    }
    return () => {
      disconnect();
    };
  }, [isAuthenticated, accessToken, connect, disconnect]);

  /**
   * Subscribe to an event handler. The handler receives (event, payload)
   * for every server event. Automatically cleaned up on unmount.
   */
  const onEvent = useCallback(
    (handler: EventHandler) => {
      addListener(handler);
      return () => removeListener(handler);
    },
    [addListener, removeListener],
  );

  return {
    connected,
    subscribe,
    unsubscribe,
    onEvent,
  };
}

/**
 * useRooms subscribes to a list of WebSocket rooms.
 * Automatically unsubscribes when rooms change or component unmounts.
 * Must be called as a hook (not inside callbacks).
 */
export function useRooms(rooms: string[]) {
  const { connected, subscribe, unsubscribe } = useWsStore();
  const prevRooms = useRef<string[]>([]);

  useEffect(() => {
    if (!connected) return;

    // Unsubscribe from rooms no longer needed.
    for (const room of prevRooms.current) {
      if (!rooms.includes(room)) {
        unsubscribe(room);
      }
    }

    // Subscribe to new rooms.
    for (const room of rooms) {
      if (!prevRooms.current.includes(room)) {
        subscribe(room);
      }
    }

    prevRooms.current = rooms;

    return () => {
      for (const room of rooms) {
        unsubscribe(room);
      }
      prevRooms.current = [];
    };
  }, [rooms, connected, subscribe, unsubscribe]);
}
