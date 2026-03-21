import { useState, useCallback, useMemo } from "react";
import type { Message } from "@/api/messenger";
import {
  deleteMessage,
  togglePin,
  addBookmark,
} from "@/api/messenger";
import { useMessengerStore } from "@/stores/messenger-store";
import { useAuthStore } from "@/stores/auth-store";
import { ReactionPicker } from "./reaction-picker";
import { ForwardDialog } from "./forward-dialog";
import { LinkPreview } from "./link-preview";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/**
 * Регулярное выражение для обнаружения URL в тексте сообщения.
 * Поддерживает http/https ссылки.
 */
const URL_REGEX = /https?:\/\/[^\s<>"')\]]+/g;

/** Пропсы компонента MessageItem. */
interface MessageItemProps {
  /** Сообщение для отображения. */
  message: Message;
  /** ID канала. */
  channelId: string;
  /** Предыдущее сообщение (для группировки). */
  previousMessage?: Message;
  /** Показать разделитель даты перед сообщением. */
  showDateSeparator?: boolean;
  /** Текст разделителя даты. */
  dateSeparatorText?: string;
  /** Показать разделитель новых сообщений. */
  showNewMessagesDivider?: boolean;
}

/** Форматирование времени из ISO строки. */
function formatTime(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

/**
 * Проверяет, нужно ли группировать сообщение с предыдущим.
 * Группировка: тот же отправитель, менее 2 минут разницы.
 */
function shouldGroupWithPrevious(
  message: Message,
  previous?: Message,
): boolean {
  if (!previous) return false;
  if (previous.sender_id !== message.sender_id) return false;
  const diff =
    new Date(message.created_at).getTime() -
    new Date(previous.created_at).getTime();
  return diff < 2 * 60 * 1000; // 2 минуты
}

/**
 * Проверяет, является ли вложение изображением по MIME-типу.
 */
function isImageAttachment(mimeType: string): boolean {
  return mimeType.startsWith("image/");
}

/** Рендер контента сообщения из Plate JSON (упрощённый). */
function RenderContent({
  contentJson,
}: {
  contentJson?: Record<string, unknown>;
}) {
  if (!contentJson) return null;

  const children = (contentJson as { children?: unknown[] }).children;
  if (!Array.isArray(children)) return null;

  return (
    <div className="prose prose-sm dark:prose-invert max-w-none">
      {children.map((node, i) => (
        <RenderNode key={i} node={node as Record<string, unknown>} />
      ))}
    </div>
  );
}

/** Рендер одной ноды Plate JSON. */
function RenderNode({ node }: { node: Record<string, unknown> }) {
  // Текстовая нода.
  if ("text" in node && typeof node.text === "string") {
    let el: React.ReactNode = node.text;
    if (node.bold) el = <strong>{el}</strong>;
    if (node.italic) el = <em>{el}</em>;
    if (node.code)
      el = (
        <code className="rounded bg-muted px-1.5 py-0.5 text-xs font-mono">
          {el}
        </code>
      );
    if (node.strikethrough) el = <s>{el}</s>;
    return <>{el}</>;
  }

  const children =
    (node.children as Record<string, unknown>[] | undefined) ?? [];
  const renderedChildren = children.map((child, i) => (
    <RenderNode key={i} node={child} />
  ));

  switch (node.type) {
    case "p":
      return <p className="mb-1 leading-relaxed">{renderedChildren}</p>;
    case "h1":
      return <h1 className="text-lg font-bold">{renderedChildren}</h1>;
    case "h2":
      return <h2 className="text-base font-bold">{renderedChildren}</h2>;
    case "ul":
      return <ul className="list-disc pl-4">{renderedChildren}</ul>;
    case "ol":
      return <ol className="list-decimal pl-4">{renderedChildren}</ol>;
    case "li":
      return <li>{renderedChildren}</li>;
    case "blockquote":
      return (
        <blockquote className="border-l-2 border-primary/30 pl-3 italic text-muted-foreground">
          {renderedChildren}
        </blockquote>
      );
    case "a":
      return (
        <a
          href={node.url as string}
          className="text-primary underline decoration-primary/30 hover:decoration-primary"
          target="_blank"
          rel="noopener noreferrer"
        >
          {renderedChildren}
        </a>
      );
    case "code_block":
      return (
        <pre className="rounded-lg bg-muted p-3 text-xs font-mono overflow-x-auto">
          <code>{renderedChildren}</code>
        </pre>
      );
    default:
      return <>{renderedChildren}</>;
  }
}

/**
 * MessageItem отображает одно сообщение чата в стиле Tinkoff Talk.
 * Собственные сообщения -- справа с градиентным фоном, чужие -- слева с фоном card.
 * Поддерживает группировку, реакции, hover-действия, ответы и вложения.
 */
export function MessageItem({
  message,
  channelId,
  previousMessage,
  showDateSeparator,
  dateSeparatorText,
  showNewMessagesDivider,
}: MessageItemProps) {
  const [showActions, setShowActions] = useState(false);
  const [showReactionPicker, setShowReactionPicker] = useState(false);
  const [showForwardDialog, setShowForwardDialog] = useState(false);
  const openThread = useMessengerStore((s) => s.openThread);
  const setReplyTo = useMessengerStore((s) => s.setReplyTo);
  const currentUserId = useAuthStore((s) => s.user?.id);
  const isOwn = currentUserId === message.sender_id;
  const isGrouped = shouldGroupWithPrevious(message, previousMessage);

  /**
   * Извлекает уникальные URL из текстового содержимого сообщения.
   * Ищет в content_text и рекурсивно в текстовых нодах content_json.
   */
  const extractedUrls = useMemo(() => {
    const urls = new Set<string>();
    if (message.content_text) {
      const matches = message.content_text.match(URL_REGEX);
      if (matches) matches.forEach((u) => urls.add(u));
    }
    if (message.content_json) {
      const extractFromNode = (node: unknown) => {
        if (!node || typeof node !== "object") return;
        const n = node as Record<string, unknown>;
        if (n.type === "a" && typeof n.url === "string") {
          urls.add(n.url);
        }
        if (typeof n.text === "string") {
          const matches = n.text.match(URL_REGEX);
          if (matches) matches.forEach((u) => urls.add(u));
        }
        if (Array.isArray(n.children)) {
          n.children.forEach(extractFromNode);
        }
      };
      extractFromNode(message.content_json);
    }
    return Array.from(urls);
  }, [message.content_text, message.content_json]);

  const handleDelete = useCallback(async () => {
    await deleteMessage(channelId, message.id);
  }, [channelId, message.id]);

  const handlePin = useCallback(async () => {
    await togglePin(channelId, message.id);
  }, [channelId, message.id]);

  const handleBookmark = useCallback(async () => {
    await addBookmark(message.id, channelId);
  }, [channelId, message.id]);

  /** Ответить на сообщение (установить reply-to). */
  const handleReply = useCallback(() => {
    setReplyTo(message);
  }, [message, setReplyTo]);

  return (
    <>
      {/* Разделитель даты */}
      {showDateSeparator && dateSeparatorText && (
        <div className="flex items-center justify-center py-3 px-5">
          <div className="flex items-center gap-3 w-full">
            <div className="flex-1 h-px bg-border" />
            <span className="text-xs font-medium text-muted-foreground px-2 select-none">
              {dateSeparatorText}
            </span>
            <div className="flex-1 h-px bg-border" />
          </div>
        </div>
      )}

      {/* Разделитель новых сообщений */}
      {showNewMessagesDivider && (
        <div className="flex items-center justify-center py-2 px-5">
          <div className="flex items-center gap-3 w-full">
            <div className="flex-1 h-px bg-primary" />
            <span className="text-xs font-semibold text-primary px-2 select-none">
              Новые сообщения
            </span>
            <div className="flex-1 h-px bg-primary" />
          </div>
        </div>
      )}

      {/* Сообщение */}
      <div
        className={cn(
          "group relative flex px-5 transition-colors animate-message-in",
          isOwn ? "justify-end" : "justify-start",
          isGrouped ? "pt-0.5" : "pt-3",
          "hover:bg-muted/30",
        )}
        onMouseEnter={() => setShowActions(true)}
        onMouseLeave={() => {
          setShowActions(false);
          setShowReactionPicker(false);
        }}
      >
        <div
          className={cn(
            "flex gap-2.5 max-w-[75%] min-w-0",
            isOwn ? "flex-row-reverse" : "flex-row",
          )}
        >
          {/* Аватар -- только для чужих, скрыт при группировке */}
          {!isOwn && (
            <div className="w-9 shrink-0">
              {!isGrouped && (
                <Avatar className="h-9 w-9 shadow-sm">
                  <AvatarFallback className="text-xs font-medium bg-primary/10 text-primary">
                    {message.sender_id?.charAt(0)?.toUpperCase() ?? "?"}
                  </AvatarFallback>
                </Avatar>
              )}
            </div>
          )}

          {/* Пузырь сообщения */}
          <div className="min-w-0 flex-1">
            {/* Имя + время -- скрыты при группировке */}
            {!isGrouped && (
              <div
                className={cn(
                  "flex items-baseline gap-2 mb-1",
                  isOwn ? "justify-end" : "justify-start",
                )}
              >
                {!isOwn && (
                  <span className="text-[13px] font-semibold text-foreground">
                    {message.sender_id}
                  </span>
                )}
                <span className="text-[11px] text-muted-foreground">
                  {formatTime(message.created_at)}
                </span>
                {message.edited_at && (
                  <span className="text-[11px] text-muted-foreground italic">
                    (изм.)
                  </span>
                )}
              </div>
            )}

            {/* Bubble контейнер */}
            <div
              className={cn(
                "relative rounded-2xl px-3.5 py-2 text-sm shadow-sm transition-shadow",
                isOwn
                  ? "bg-primary text-primary-foreground rounded-tr-md"
                  : "bg-card border border-border/50 rounded-tl-md",
                isGrouped && isOwn && "rounded-tr-2xl",
                isGrouped && !isOwn && "rounded-tl-2xl",
              )}
            >
              {/* Forwarded badge */}
              {message.forwarded_from_id && (
                <div className="mb-1.5 flex items-center gap-1 text-xs opacity-70">
                  <svg
                    className="h-3 w-3"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    strokeWidth={2}
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      d="M13 5l7 7-7 7M5 5l7 7-7 7"
                    />
                  </svg>
                  <span className="italic">Переслано</span>
                </div>
              )}

              {/* Контент сообщения */}
              {message.content_json && (
                <div className={cn(isOwn && "[&_p]:text-primary-foreground [&_a]:text-primary-foreground/90 [&_a]:underline")}>
                  <RenderContent contentJson={message.content_json} />
                </div>
              )}
              {!message.content_json && message.content_text && (
                <p className="whitespace-pre-wrap leading-relaxed">
                  {message.content_text}
                </p>
              )}

              {/* Голосовое сообщение */}
              {message.voice_url && (
                <div className="mt-1.5 flex items-center gap-2 rounded-lg bg-black/5 dark:bg-white/5 p-2">
                  <audio controls preload="metadata" className="h-8 max-w-xs">
                    <source src={message.voice_url} type="audio/webm" />
                  </audio>
                  {message.voice_duration_ms && (
                    <span className="text-xs opacity-70">
                      {Math.floor(message.voice_duration_ms / 1000)}с
                    </span>
                  )}
                </div>
              )}

              {/* Вложения -- изображения инлайн, файлы как ссылки */}
              {message.attachments && message.attachments.length > 0 && (
                <div className="mt-2 space-y-2">
                  {/* Превью изображений */}
                  {message.attachments
                    .filter((att) => isImageAttachment(att.mime_type))
                    .map((att) => (
                      <a
                        key={att.file_id}
                        href={att.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block overflow-hidden rounded-lg"
                      >
                        <img
                          src={att.url}
                          alt={att.file_name}
                          className="max-w-full max-h-64 rounded-lg object-cover hover:opacity-90 transition-opacity"
                          loading="lazy"
                        />
                      </a>
                    ))}
                  {/* Файлы */}
                  <div className="flex flex-wrap gap-1.5">
                    {message.attachments
                      .filter((att) => !isImageAttachment(att.mime_type))
                      .map((att) => (
                        <a
                          key={att.file_id}
                          href={att.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={cn(
                            "inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs transition-colors",
                            isOwn
                              ? "bg-white/15 hover:bg-white/25"
                              : "border bg-card hover:bg-muted shadow-sm",
                          )}
                        >
                          <svg
                            className="h-3.5 w-3.5 opacity-60"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            strokeWidth={2}
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
                            />
                          </svg>
                          <span className="font-medium truncate max-w-[200px]">
                            {att.file_name}
                          </span>
                        </a>
                      ))}
                  </div>
                </div>
              )}

              {/* Время внутри пузыря (для сгруппированных) */}
              {isGrouped && (
                <span
                  className={cn(
                    "mt-0.5 block text-[10px] text-right",
                    isOwn ? "text-primary-foreground/60" : "text-muted-foreground",
                  )}
                >
                  {formatTime(message.created_at)}
                </span>
              )}
            </div>

            {/* Превью ссылок */}
            {extractedUrls.length > 0 && (
              <div className={cn("mt-1.5 space-y-1.5", isOwn && "flex flex-col items-end")}>
                {extractedUrls.slice(0, 2).map((url) => (
                  <LinkPreview key={url} url={url} />
                ))}
              </div>
            )}

            {/* Реакции */}
            {message.reactions && message.reactions.length > 0 && (
              <div
                className={cn(
                  "mt-1.5 flex flex-wrap gap-1",
                  isOwn ? "justify-end" : "justify-start",
                )}
              >
                {message.reactions.map((reaction) => {
                  const hasReacted = currentUserId
                    ? reaction.user_ids.includes(currentUserId)
                    : false;
                  return (
                    <button
                      key={reaction.emoji}
                      type="button"
                      className={cn(
                        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs transition-all hover:scale-105",
                        hasReacted
                          ? "border-primary/50 bg-primary/10 text-primary"
                          : "border-border bg-card hover:bg-accent",
                      )}
                      onClick={() =>
                        void import("@/api/messenger").then((m) =>
                          m.toggleReaction(channelId, message.id, reaction.emoji),
                        )
                      }
                    >
                      <span>{reaction.emoji}</span>
                      <span className="font-medium">
                        {reaction.user_ids.length}
                      </span>
                    </button>
                  );
                })}
              </div>
            )}

            {/* Индикатор треда */}
            {message.reply_count > 0 && (
              <button
                type="button"
                className="mt-1.5 inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium text-primary hover:bg-primary/10 transition-colors"
                onClick={() => openThread(message.id)}
              >
                <svg
                  className="h-3.5 w-3.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"
                  />
                </svg>
                {message.reply_count}{" "}
                {message.reply_count === 1 ? "ответ" : "ответов"}
              </button>
            )}
          </div>
        </div>

        {/* Hover-действия -- плавающая панель */}
        {showActions && (
          <div
            className={cn(
              "absolute -top-3 flex items-center gap-0.5 rounded-lg border bg-card p-1 shadow-lg animate-fade-in z-10",
              isOwn ? "left-4" : "right-4",
            )}
          >
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Ответить"
              onClick={handleReply}
            >
              <svg
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"
                />
              </svg>
            </Button>

            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Тред"
              onClick={() => openThread(message.id)}
            >
              <svg
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                />
              </svg>
            </Button>

            <ReactionPicker
              channelId={channelId}
              messageId={message.id}
              open={showReactionPicker}
              onOpenChange={setShowReactionPicker}
            />

            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Закрепить"
              onClick={handlePin}
            >
              <svg
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
                />
              </svg>
            </Button>

            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Сохранить"
              onClick={handleBookmark}
            >
              <svg
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
                />
              </svg>
            </Button>

            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Переслать"
              onClick={() => setShowForwardDialog(true)}
            >
              <svg
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M13 5l7 7-7 7M5 5l7 7-7 7"
                />
              </svg>
            </Button>

            {isOwn && (
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 text-destructive hover:text-destructive"
                title="Удалить"
                onClick={handleDelete}
              >
                <svg
                  className="h-3.5 w-3.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </Button>
            )}
          </div>
        )}

        {/* Диалог пересылки */}
        <ForwardDialog
          message={message}
          channelId={channelId}
          open={showForwardDialog}
          onOpenChange={setShowForwardDialog}
        />
      </div>
    </>
  );
}
