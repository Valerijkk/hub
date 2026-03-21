import { useState, useCallback } from "react";
import type { Message } from "@/api/messenger";
import {
  deleteMessage,
  togglePin,
  addBookmark,
} from "@/api/messenger";
import { useMessengerStore } from "@/stores/messenger-store";
import { useAuthStore } from "@/stores/auth-store";
import { ReactionPicker } from "./reaction-picker";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/** Пропсы компонента MessageItem. */
interface MessageItemProps {
  /** Сообщение для отображения. */
  message: Message;
  /** ID канала. */
  channelId: string;
}

/** Форматирование времени из ISO строки. */
function formatTime(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
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
    if (node.code) el = <code className="rounded bg-muted px-1 py-0.5 text-xs">{el}</code>;
    if (node.strikethrough) el = <s>{el}</s>;
    return <>{el}</>;
  }

  const children = (node.children as Record<string, unknown>[] | undefined) ?? [];
  const renderedChildren = children.map((child, i) => (
    <RenderNode key={i} node={child} />
  ));

  switch (node.type) {
    case "p":
      return <p className="mb-1">{renderedChildren}</p>;
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
        <blockquote className="border-l-2 border-muted-foreground/30 pl-3 italic">
          {renderedChildren}
        </blockquote>
      );
    case "a":
      return (
        <a
          href={node.url as string}
          className="text-primary underline"
          target="_blank"
          rel="noopener noreferrer"
        >
          {renderedChildren}
        </a>
      );
    case "code_block":
      return (
        <pre className="rounded bg-muted p-2 text-xs">
          <code>{renderedChildren}</code>
        </pre>
      );
    default:
      return <>{renderedChildren}</>;
  }
}

/**
 * MessageItem отображает одно сообщение чата.
 * Включает аватар, имя, время, контент (Plate JSON / голосовое / forwarded),
 * реакции, индикатор треда и hover-действия.
 */
export function MessageItem({ message, channelId }: MessageItemProps) {
  const [showActions, setShowActions] = useState(false);
  const [showReactionPicker, setShowReactionPicker] = useState(false);
  const openThread = useMessengerStore((s) => s.openThread);
  const currentUserId = useAuthStore((s) => s.user?.id);
  const isAuthor = currentUserId === message.sender_id;

  const handleDelete = useCallback(async () => {
    await deleteMessage(channelId, message.id);
  }, [channelId, message.id]);

  const handlePin = useCallback(async () => {
    await togglePin(channelId, message.id);
  }, [channelId, message.id]);

  const handleBookmark = useCallback(async () => {
    await addBookmark(message.id, channelId);
  }, [channelId, message.id]);

  return (
    <div
      className={cn(
        "group relative flex gap-3 px-4 py-2 hover:bg-muted/50 transition-colors",
      )}
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => {
        setShowActions(false);
        setShowReactionPicker(false);
      }}
    >
      {/* Аватар */}
      <Avatar className="h-8 w-8 shrink-0 mt-0.5">
        <AvatarImage src={message.sender_avatar_url} />
        <AvatarFallback className="text-xs">
          {message.sender_name?.charAt(0)?.toUpperCase() ?? "?"}
        </AvatarFallback>
      </Avatar>

      {/* Контент */}
      <div className="flex-1 min-w-0">
        {/* Заголовок: имя + время */}
        <div className="flex items-baseline gap-2">
          <span className="text-sm font-medium">{message.sender_name}</span>
          <span className="text-xs text-muted-foreground">
            {formatTime(message.created_at)}
          </span>
          {message.edited_at && (
            <span className="text-xs text-muted-foreground">(изменено)</span>
          )}
          {message.is_pinned && (
            <span className="text-xs text-yellow-500">закреплено</span>
          )}
        </div>

        {/* Forwarded badge */}
        {message.forwarded_from_name && (
          <div className="mb-1 text-xs text-muted-foreground italic">
            Переслано от {message.forwarded_from_name}
          </div>
        )}

        {/* Контент сообщения */}
        {message.content_json && (
          <RenderContent contentJson={message.content_json} />
        )}
        {!message.content_json && message.content_text && (
          <p className="text-sm whitespace-pre-wrap">{message.content_text}</p>
        )}

        {/* Голосовое сообщение */}
        {message.voice_url && (
          <div className="mt-1 flex items-center gap-2">
            <audio controls preload="metadata" className="h-8 max-w-xs">
              <source src={message.voice_url} type="audio/webm" />
            </audio>
            {message.voice_duration_ms && (
              <span className="text-xs text-muted-foreground">
                {Math.floor(message.voice_duration_ms / 1000)}s
              </span>
            )}
          </div>
        )}

        {/* Вложения */}
        {message.attachments && message.attachments.length > 0 && (
          <div className="mt-1 flex flex-wrap gap-2">
            {message.attachments.map((att) => (
              <a
                key={att.id}
                href={att.url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 rounded bg-muted px-2 py-1 text-xs hover:bg-muted/80"
              >
                <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
                </svg>
                {att.filename}
              </a>
            ))}
          </div>
        )}

        {/* Реакции */}
        {message.reactions.length > 0 && (
          <div className="mt-1 flex flex-wrap gap-1">
            {message.reactions.map((reaction) => (
              <button
                key={reaction.emoji}
                type="button"
                className={cn(
                  "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs transition-colors hover:bg-accent",
                  reaction.has_reacted && "border-primary bg-primary/10",
                )}
                onClick={() =>
                  void import("@/api/messenger").then((m) =>
                    m.toggleReaction(channelId, message.id, reaction.emoji),
                  )
                }
              >
                <span>{reaction.emoji}</span>
                <span>{reaction.count}</span>
              </button>
            ))}
          </div>
        )}

        {/* Индикатор треда */}
        {message.reply_count > 0 && (
          <button
            type="button"
            className="mt-1 text-xs text-primary hover:underline"
            onClick={() => openThread(message.id)}
          >
            {message.reply_count}{" "}
            {message.reply_count === 1 ? "ответ" : "ответов"}
          </button>
        )}
      </div>

      {/* Hover actions */}
      {showActions && (
        <div className="absolute -top-3 right-4 flex items-center gap-0.5 rounded-md border bg-background p-0.5 shadow-sm">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            title="Ответить в тред"
            onClick={() => openThread(message.id)}
          >
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
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
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
            </svg>
          </Button>

          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            title="Сохранить"
            onClick={handleBookmark}
          >
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
            </svg>
          </Button>

          {isAuthor && (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 text-destructive hover:text-destructive"
              title="Удалить"
              onClick={handleDelete}
            >
              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
