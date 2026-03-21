import { useEffect, useCallback, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { getChannels } from "@/api/messenger";
import type { Channel } from "@/api/messenger";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";
import { useAuthStore } from "@/stores/auth-store";

/**
 * useChannels загружает список каналов пользователя через TanStack Query
 * и подписывается на WS-обновления (новый канал, обновление, удаление).
 */
export function useChannels() {
  const userId = useAuthStore((s) => s.user?.id);
  const setChannels = useMessengerStore((s) => s.setChannels);
  const addChannel = useMessengerStore((s) => s.addChannel);
  const updateChannel = useMessengerStore((s) => s.updateChannel);
  const incrementUnread = useMessengerStore((s) => s.incrementUnread);

  const { subscribe, unsubscribe, addListener, removeListener, connected } =
    useWsStore();

  const query = useQuery({
    queryKey: ["messenger", "channels"],
    queryFn: getChannels,
  });

  // Синхронизация данных из query в store.
  useEffect(() => {
    if (query.data) {
      setChannels(query.data);
    }
  }, [query.data, setChannels]);

  // Подписка на WS room для обновлений каналов.
  useEffect(() => {
    if (!connected || !userId) return;
    const room = `user:${userId}:channels`;
    subscribe(room);
    return () => {
      unsubscribe(room);
    };
  }, [connected, userId, subscribe, unsubscribe]);

  // Обработка WS-событий каналов.
  const handlerRef = useRef<((event: string, payload: unknown) => void) | null>(null);
  const handleEvent = useCallback(
    (event: string, payload: unknown) => {
      const data = payload as Record<string, unknown>;
      switch (event) {
        case "channel_created":
          addChannel(data as unknown as Channel);
          break;
        case "channel_updated":
          if (typeof data.id === "string") {
            updateChannel(data.id, data as Partial<Channel>);
          }
          break;
        case "new_message_notification":
          if (typeof data.channel_id === "string") {
            incrementUnread(data.channel_id);
          }
          break;
      }
    },
    [addChannel, updateChannel, incrementUnread],
  );

  useEffect(() => {
    if (handlerRef.current) {
      removeListener(handlerRef.current);
    }
    handlerRef.current = handleEvent;
    addListener(handleEvent);
    return () => {
      removeListener(handleEvent);
    };
  }, [handleEvent, addListener, removeListener]);

  return query;
}
