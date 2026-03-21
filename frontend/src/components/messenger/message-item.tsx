import { useState, useCallback, useMemo, memo } from "react";
import type { Message } from "@/api/messenger";
import {
  deleteMessage,
  togglePin,
  addBookmark,
} from "@/api/messenger";
import { useMessengerStore } from "@/stores/messenger-store";
import { useAuthStore } from "@/stores/auth-store";
import { ForwardDialog } from "./forward-dialog";
import { LinkPreview } from "./link-preview";
import { RenderContent } from "./render-content";
import { MessageActions } from "./message-actions";
import { MessageAttachments } from "./message-attachments";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

/**
 * Regular expression for detecting URLs in message text.
 * Supports http/https links.
 */
const URL_REGEX = /https?:\/\/[^\s<>"')\]]+/g;

/** Props for the MessageItem component. */
interface MessageItemProps {
  /** Message to display. */
  message: Message;
  /** Channel ID. */
  channelId: string;
  /** Previous message (for grouping). */
  previousMessage?: Message;
  /** Show date separator before the message. */
  showDateSeparator?: boolean;
  /** Date separator text. */
  dateSeparatorText?: string;
  /** Show new messages divider. */
  showNewMessagesDivider?: boolean;
}

/** Format time from ISO string. */
function formatTime(iso: string): string {
  const date = new Date(iso);
  return date.toLocaleTimeString("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

/**
 * Check if a message should be grouped with the previous one.
 * Grouping: same sender, less than 2 minutes apart.
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
  return diff < 2 * 60 * 1000; // 2 minutes
}

/**
 * MessageItem displays a single chat message in Tinkoff Talk style.
 * Own messages are on the right with gradient background, others on the left with card background.
 * Supports grouping, reactions, hover actions, replies and attachments.
 */
function MessageItemInner({
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
   * Extracts unique URLs from message text content.
   * Searches content_text and recursively in content_json text nodes.
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

  /** Reply to message (set reply-to). */
  const handleReply = useCallback(() => {
    setReplyTo(message);
  }, [message, setReplyTo]);

  return (
    <>
      {/* Date separator */}
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

      {/* New messages divider */}
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

      {/* Message */}
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
          {/* Avatar -- only for others, hidden when grouped */}
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

          {/* Message bubble */}
          <div className="min-w-0 flex-1">
            {/* Name + time -- hidden when grouped */}
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

            {/* Bubble container */}
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

              {/* Message content */}
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

              {/* Voice message */}
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

              {/* Attachments */}
              {message.attachments && message.attachments.length > 0 && (
                <MessageAttachments attachments={message.attachments} isOwn={isOwn} />
              )}

              {/* Time inside bubble (for grouped messages) */}
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

            {/* Link previews */}
            {extractedUrls.length > 0 && (
              <div className={cn("mt-1.5 space-y-1.5", isOwn && "flex flex-col items-end")}>
                {extractedUrls.slice(0, 2).map((url) => (
                  <LinkPreview key={url} url={url} />
                ))}
              </div>
            )}

            {/* Reactions */}
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

            {/* Thread indicator */}
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

        {/* Hover actions -- floating panel */}
        {showActions && (
          <MessageActions
            message={message}
            channelId={channelId}
            isOwn={isOwn}
            showReactionPicker={showReactionPicker}
            onReactionPickerChange={setShowReactionPicker}
            onReply={handleReply}
            onOpenThread={() => openThread(message.id)}
            onPin={handlePin}
            onBookmark={handleBookmark}
            onForward={() => setShowForwardDialog(true)}
            onDelete={handleDelete}
          />
        )}

        {/* Forward dialog */}
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

export const MessageItem = memo(MessageItemInner);
