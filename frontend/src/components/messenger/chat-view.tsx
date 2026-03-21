import { useRef, useEffect, useCallback, useState } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useMessages } from "@/hooks/use-messages";
import { useMessengerStore } from "@/stores/messenger-store";
import { getChannel } from "@/api/messenger";
import type { Channel } from "@/api/messenger";
import { useQuery } from "@tanstack/react-query";
import { MessageItem } from "./message-item";
import { MessageComposer } from "./message-composer";
import { ThreadPanel } from "./thread-panel";
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
  const activeThreadMessageId = useMessengerStore(
    (s) => s.activeThreadMessageId,
  );
  const closeThread = useMessengerStore((s) => s.closeThread);

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
      <div className="flex h-12 shrink-0 items-center justify-between border-b px-4">
        <div className="flex items-center gap-2">
          <h2 className="text-sm font-semibold">
            {channel?.name ?? "Загрузка..."}
          </h2>
          {channel && (
            <span className="text-xs text-muted-foreground">
              {channel.members_count} участников
            </span>
          )}
        </div>
      </div>

      {/* Область сообщений */}
      <div className="flex flex-1 overflow-hidden">
        <div className="flex flex-1 flex-col overflow-hidden">
          {isLoading ? (
            <div className="flex flex-1 items-center justify-center">
              <span className="text-sm text-muted-foreground">
                Загрузка сообщений...
              </span>
            </div>
          ) : messages.length === 0 ? (
            <div className="flex flex-1 items-center justify-center">
              <span className="text-sm text-muted-foreground">
                Нет сообщений. Начните общение!
              </span>
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

          {/* Composer */}
          {!channel?.is_read_only && (
            <div className="shrink-0 border-t p-3">
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
