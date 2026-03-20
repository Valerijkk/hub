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

  /**
   * Subscribe to a list of rooms. Automatically unsubscribes when the
   * rooms change or the component unmounts.
   */
  const useRooms = useCallback(
    (rooms: string[]) => {
      // eslint-disable-next-line react-hooks/rules-of-hooks
      const prevRooms = useRef<string[]>([]);

      // eslint-disable-next-line react-hooks/rules-of-hooks
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
      }, [rooms, connected]); // eslint-disable-line react-hooks/exhaustive-deps
    },
    [connected, subscribe, unsubscribe],
  );

  return {
    connected,
    subscribe,
    unsubscribe,
    onEvent,
    useRooms,
  };
}
