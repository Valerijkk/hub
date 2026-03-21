import { useRef, useEffect, useCallback, useState } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useMessages } from "@/hooks/use-messages";
import { useMarkRead } from "@/hooks/use-unread";
import { useMessengerStore } from "@/stores/messenger-store";
import { getChannel } from "@/api/messenger";
import type { Channel } from "@/api/messenger";
import { useQuery } from "@tanstack/react-query";
import { MessageItem } from "./message-item";
import { MessageComposer } from "./message-composer";
import { ThreadPanel } from "./thread-panel";
import { SearchDialog } from "./search-dialog";
import { PinnedMessages } from "./pinned-messages";
import { CallButton } from "@/components/calls/CallButton";
import { cn } from "@/lib/utils";

/** Пропсы компонента ChatView. */
interface ChatViewProps {
  /** ID канала для отображения. */
  channelId: string;
}

/**
 * ChatView -- основной компонент чата с виртуализацией списка сообщений.
 * Использует @tanstack/react-virtual для рендеринга только видимых сообщений.
 * Поддерживает infinite scroll вверх и автоскролл вниз для новых сообщений.
 */
export function ChatView({ channelId }: ChatViewProps) {
  const { messages, loadMore, hasMore, isLoading } = useMessages(channelId);
  const markRead = useMarkRead();
  const activeThreadMessageId = useMessengerStore(
    (s) => s.activeThreadMessageId,
  );
  const closeThread = useMessengerStore((s) => s.closeThread);

  // Пометить канал как прочитанный при открытии.
  useEffect(() => {
    if (channelId) {
      void markRead(channelId);
    }
  }, [channelId, markRead]);

  const { data: channel } = useQuery<Channel>({
    queryKey: ["messenger", "channel", channelId],
    queryFn: () => getChannel(channelId),
  });

  const parentRef = useRef<HTMLDivElement>(null);
  const [isAtBottom, setIsAtBottom] = useState(true);

  const virtualizer = useVirtualizer({
    count: messages.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 80,
    overscan: 5,
  });

  // Автоскролл вниз при новых сообщениях (если пользователь внизу).
  const prevCountRef = useRef(messages.length);
  useEffect(() => {
    if (messages.length > prevCountRef.current && isAtBottom) {
      virtualizer.scrollToIndex(messages.length - 1, { align: "end" });
    }
    prevCountRef.current = messages.length;
  }, [messages.length, isAtBottom, virtualizer]);

  // Первоначальная прокрутка вниз.
  useEffect(() => {
    if (messages.length > 0 && prevCountRef.current === messages.length) {
      virtualizer.scrollToIndex(messages.length - 1, { align: "end" });
    }
  }, [channelId]); // eslint-disable-line react-hooks/exhaustive-deps

  // Определение позиции скролла (для автоскролла и infinite scroll).
  const handleScroll = useCallback(() => {
    const el = parentRef.current;
    if (!el) return;

    // Проверка прокрутки вниз.
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50;
    setIsAtBottom(atBottom);

    // Infinite scroll вверх: загрузка старых сообщений.
    if (el.scrollTop < 100 && hasMore) {
      loadMore();
    }
  }, [hasMore, loadMore]);

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      {/* Заголовок канала */}
      <div className="flex h-14 shrink-0 items-center justify-between border-b bg-card/60 backdrop-blur-sm px-5">
        <div className="flex items-center gap-3">
          {/* Иконка канала */}
          {channel && (
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
              {channel.channel_type === "public" ? (
                <span className="text-sm font-bold">#</span>
              ) : channel.channel_type === "dm" ? (
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              ) : (
                <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              )}
            </div>
          )}
          <div>
            <h2 className="text-sm font-semibold leading-tight">
              {channel?.name ?? "Загрузка..."}
            </h2>
            {channel && (
              <span className="text-[11px] text-muted-foreground">
                {channel.members_count} участников
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-1">
          <CallButton channelId={channelId} />
          <SearchDialog channelId={channelId} />
          <PinnedMessages channelId={channelId} />
        </div>
      </div>

      {/* Область сообщений */}
      <div className="flex flex-1 overflow-hidden">
        <div className="flex flex-1 flex-col overflow-hidden">
          {isLoading ? (
            /* Скелетон загрузки сообщений */
            <div className="flex flex-1 flex-col justify-end p-4 space-y-4">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex gap-3">
                  <div className="h-8 w-8 rounded-full animate-shimmer shrink-0" />
                  <div className="space-y-2 flex-1">
                    <div className="h-3 w-24 rounded animate-shimmer" />
                    <div className="h-4 w-3/4 rounded animate-shimmer" />
                  </div>
                </div>
              ))}
            </div>
          ) : messages.length === 0 ? (
            /* Пустое состояние — центрированное, приятное */
            <div className="flex flex-1 items-center justify-center">
              <div className="text-center max-w-xs">
                <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-muted">
                  <svg className="h-8 w-8 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                </div>
                <h3 className="text-base font-medium text-foreground mb-1">
                  Нет сообщений
                </h3>
                <p className="text-sm text-muted-foreground">
                  Начните общение! Напишите первое сообщение в этом канале.
                </p>
              </div>
            </div>
          ) : (
            <div
              ref={parentRef}
              className="flex-1 overflow-y-auto"
              onScroll={handleScroll}
            >
              <div
                className="relative w-full"
                style={{ height: `${virtualizer.getTotalSize()}px` }}
              >
                {virtualizer.getVirtualItems().map((virtualItem) => {
                  const message = messages[virtualItem.index];
                  if (!message) return null;
                  return (
                    <div
                      key={virtualItem.key}
                      data-index={virtualItem.index}
                      ref={virtualizer.measureElement}
                      className="absolute left-0 top-0 w-full"
                      style={{
                        transform: `translateY(${virtualItem.start}px)`,
                      }}
                    >
                      <MessageItem
                        message={message}
                        channelId={channelId}
                      />
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Composer — фиксированный внизу */}
          {/* Композер — стеклянный фон */}
          {!channel?.is_readonly && (
            <div className="shrink-0 border-t glass-subtle px-4 py-3">
              <MessageComposer channelId={channelId} />
            </div>
          )}
        </div>

        {/* Thread panel */}
        {activeThreadMessageId && (
          <div
            className={cn(
              "w-96 shrink-0 border-l",
              "max-md:fixed max-md:inset-y-0 max-md:right-0 max-md:z-50 max-md:w-full max-md:bg-background",
            )}
          >
            <ThreadPanel
              channelId={channelId}
              messageId={activeThreadMessageId}
              onClose={closeThread}
            />
          </div>
        )}
      </div>
    </div>
  );
}
