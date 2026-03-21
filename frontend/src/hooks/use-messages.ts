import { useEffect, useCallback, useRef } from "react";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { getMessages, getThreadReplies } from "@/api/messenger";
import type { Message } from "@/api/messenger";
import { useWsStore } from "@/stores/ws-store";

/** Количество сообщений за один запрос. */
const PAGE_SIZE = 30;

/**
 * useMessages загружает сообщения канала с cursor-пагинацией (вверх)
 * и подписывается на WS room channel:{channelId} для real-time обновлений.
 */
export function useMessages(channelId: string | undefined) {
  const queryClient = useQueryClient();
  const { subscribe, unsubscribe, addListener, removeListener, connected } =
    useWsStore();

  const query = useInfiniteQuery({
    queryKey: ["messenger", "messages", channelId],
    queryFn: ({ pageParam }) =>
      getMessages(channelId!, { before: pageParam, limit: PAGE_SIZE }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => {
      if (lastPage.length < PAGE_SIZE) return undefined;
      // Backend expects RFC3339 timestamp for cursor-based pagination.
      return lastPage[0]?.created_at;
    },
    enabled: !!channelId,
  });

  // Подписка на WS room канала.
  useEffect(() => {
    if (!connected || !channelId) return;
    const room = `channel:${channelId}`;
    subscribe(room);
    return () => {
      unsubscribe(room);
    };
  }, [connected, channelId, subscribe, unsubscribe]);

  // Обработка WS-событий сообщений.
  const handlerRef = useRef<((event: string, payload: unknown) => void) | null>(null);

  const handleEvent = useCallback(
    (event: string, payload: unknown) => {
      if (!channelId) return;
      const data = payload as Record<string, unknown>;
      const msgChannelId = data.channel_id as string | undefined;

      // Игнорируем события для других каналов.
      if (msgChannelId && msgChannelId !== channelId) return;

      const queryKey = ["messenger", "messages", channelId];

      switch (event) {
        case "message_sent": {
          const message = data as unknown as Message;
          queryClient.setQueryData(queryKey, (old: unknown) => {
            if (!old) return old;
            const typed = old as { pages: Message[][]; pageParams: unknown[] };
            const lastPage = typed.pages[typed.pages.length - 1] ?? [];
            return {
              ...typed,
              pages: [
                ...typed.pages.slice(0, -1),
                [...lastPage, message],
              ],
            };
          });
          break;
        }
        case "message_edited": {
          const updated = data as unknown as Message;
          queryClient.setQueryData(queryKey, (old: unknown) => {
            if (!old) return old;
            const typed = old as { pages: Message[][]; pageParams: unknown[] };
            return {
              ...typed,
              pages: typed.pages.map((page) =>
                page.map((msg) => (msg.id === updated.id ? updated : msg)),
              ),
            };
          });
          break;
        }
        case "message_deleted": {
          const deletedId = data.message_id as string;
          queryClient.setQueryData(queryKey, (old: unknown) => {
            if (!old) return old;
            const typed = old as { pages: Message[][]; pageParams: unknown[] };
            return {
              ...typed,
              pages: typed.pages.map((page) =>
                page.filter((msg) => msg.id !== deletedId),
              ),
            };
          });
          break;
        }
        case "reaction_toggled": {
          const msgId = data.message_id as string;
          const reactions = data.reactions as Message["reactions"];
          queryClient.setQueryData(queryKey, (old: unknown) => {
            if (!old) return old;
            const typed = old as { pages: Message[][]; pageParams: unknown[] };
            return {
              ...typed,
              pages: typed.pages.map((page) =>
                page.map((msg) =>
                  msg.id === msgId ? { ...msg, reactions } : msg,
                ),
              ),
            };
          });
          break;
        }
      }
    },
    [channelId, queryClient],
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

  /** Все сообщения из всех страниц (отсортированные по времени). */
  const messages = query.data?.pages.flat() ?? [];

  return {
    ...query,
    messages,
    loadMore: query.fetchNextPage,
    hasMore: query.hasNextPage,
  };
}

/**
 * useThreadMessages загружает ответы в треде с cursor-пагинацией (вверх)
 * и подписывается на WS room для обновлений.
 */
export function useThreadMessages(
  channelId: string | undefined,
  messageId: string | undefined,
) {
  const queryClient = useQueryClient();
  const { addListener, removeListener } = useWsStore();

  const query = useInfiniteQuery({
    queryKey: ["messenger", "thread", channelId, messageId],
    queryFn: ({ pageParam }) =>
      getThreadReplies(channelId!, messageId!, {
        before: pageParam,
        limit: PAGE_SIZE,
      }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => {
      if (lastPage.length < PAGE_SIZE) return undefined;
      // Backend expects RFC3339 timestamp for cursor-based pagination.
      return lastPage[0]?.created_at;
    },
    enabled: !!channelId && !!messageId,
  });

  // Обработка WS-событий для треда.
  const handlerRef = useRef<((event: string, payload: unknown) => void) | null>(null);

  const handleEvent = useCallback(
    (event: string, payload: unknown) => {
      if (!channelId || !messageId) return;
      const data = payload as Record<string, unknown>;
      const threadId = data.thread_id as string | undefined;
      if (threadId !== messageId) return;

      const queryKey = ["messenger", "thread", channelId, messageId];

      if (event === "message_sent") {
        const message = data as unknown as Message;
        queryClient.setQueryData(queryKey, (old: unknown) => {
          if (!old) return old;
          const typed = old as { pages: Message[][]; pageParams: unknown[] };
          const lastPage = typed.pages[typed.pages.length - 1] ?? [];
          return {
            ...typed,
            pages: [...typed.pages.slice(0, -1), [...lastPage, message]],
          };
        });
      }
    },
    [channelId, messageId, queryClient],
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

  const messages = query.data?.pages.flat() ?? [];

  return {
    ...query,
    messages,
    loadMore: query.fetchNextPage,
    hasMore: query.hasNextPage,
  };
}
