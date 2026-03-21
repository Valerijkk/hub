import { useState, useCallback, useRef, type KeyboardEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { sendMessage } from "@/api/messenger";
import { useMessengerStore } from "@/stores/messenger-store";
import { VoiceRecorder } from "./voice-recorder";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/** Пропсы компонента MessageComposer. */
interface MessageComposerProps {
  /** ID канала для отправки сообщения. */
  channelId: string;
  /** ID треда (если отвечаем в тред). */
  threadId?: string;
  /** Заглушка-подсказка. */
  placeholder?: string;
}

/**
 * MessageComposer -- редактор сообщений в стиле Tinkoff Talk.
 * Поддерживает reply preview, автоподгонку высоты (макс 5 строк),
 * голосовые сообщения, кнопку эмодзи и плавные анимации.
 */
export function MessageComposer({
  channelId,
  threadId,
  placeholder = "Написать сообщение...",
}: MessageComposerProps) {
  const queryClient = useQueryClient();
  const [text, setText] = useState("");
  const [isSending, setIsSending] = useState(false);
  const [showVoice, setShowVoice] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const replyToMessage = useMessengerStore((s) => s.replyToMessage);
  const setReplyTo = useMessengerStore((s) => s.setReplyTo);

  /** Отправка текстового сообщения. */
  const handleSend = useCallback(async () => {
    const trimmed = text.trim();
    if (!trimmed) return;

    setIsSending(true);
    try {
      await sendMessage(channelId, {
        content_json: {
          type: "doc",
          children: [{ type: "p", children: [{ text: trimmed }] }],
        },
        content_text: trimmed,
        thread_id: threadId,
      });
      void queryClient.invalidateQueries({
        queryKey: ["messenger", "messages", channelId],
      });
      setText("");
      setReplyTo(null);
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto";
      }
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Не удалось отправить сообщение";
      toast.error(message);
    } finally {
      setIsSending(false);
    }
  }, [channelId, threadId, text, queryClient, setReplyTo]);

  /** Обработка клавиш: Enter -- отправить, Shift+Enter -- перенос строки. */
  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        void handleSend();
      }
    },
    [handleSend],
  );

  /** Автоматическая подгонка высоты textarea (макс 5 строк ~120px). */
  const handleInput = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      setText(e.target.value);
      const el = e.target;
      el.style.height = "auto";
      el.style.height = `${Math.min(el.scrollHeight, 120)}px`;
    },
    [],
  );

  /** Отправка голосового сообщения. */
  const handleVoiceSend = useCallback(
    async (voiceUrl: string, durationMs: number) => {
      try {
        await sendMessage(channelId, {
          content_json: {
            type: "doc",
            children: [{ type: "p", children: [{ text: "" }] }],
          },
          voice_url: voiceUrl,
          voice_duration_ms: durationMs,
          thread_id: threadId,
        });
        void queryClient.invalidateQueries({
          queryKey: ["messenger", "messages", channelId],
        });
      } catch (err) {
        const message =
          err instanceof Error
            ? err.message
            : "Не удалось отправить голосовое сообщение";
        toast.error(message);
      }
      setShowVoice(false);
    },
    [channelId, threadId, queryClient],
  );

  /** Текст превью сообщения для reply. */
  const replyPreviewText = replyToMessage?.content_text
    ? replyToMessage.content_text.slice(0, 80) +
      (replyToMessage.content_text.length > 80 ? "..." : "")
    : "Вложение";

  return (
    <div className="space-y-0">
      {/* Reply preview */}
      {replyToMessage && (
        <div className="flex items-center gap-2 px-1 py-2 mb-1.5 animate-fade-in">
          <div className="w-0.5 h-8 rounded-full bg-primary shrink-0" />
          <div className="flex-1 min-w-0">
            <p className="text-xs font-semibold text-primary truncate">
              {replyToMessage.sender_id}
            </p>
            <p className="text-xs text-muted-foreground truncate">
              {replyPreviewText}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setReplyTo(null)}
            className="shrink-0 h-6 w-6 flex items-center justify-center rounded-full hover:bg-muted transition-colors"
            title="Отменить ответ"
          >
            <svg
              className="h-3.5 w-3.5 text-muted-foreground"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      )}

      {showVoice ? (
        <VoiceRecorder
          onSend={handleVoiceSend}
          onCancel={() => setShowVoice(false)}
        />
      ) : (
        <div className="flex items-end gap-2">
          {/* Кнопка вложения */}
          <button
            type="button"
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="Прикрепить файл"
          >
            <svg
              className="h-5 w-5"
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
          </button>

          {/* Поле ввода */}
          <div className="flex-1 rounded-xl border bg-background/80 backdrop-blur-sm shadow-sm transition-all duration-200 focus-within:shadow-md focus-within:border-ring/50 focus-within:bg-background">
            <textarea
              ref={textareaRef}
              value={text}
              onChange={handleInput}
              onKeyDown={handleKeyDown}
              placeholder={placeholder}
              rows={1}
              disabled={isSending}
              className={cn(
                "w-full resize-none bg-transparent px-4 py-3 text-sm focus:outline-none",
                "min-h-[44px] max-h-[120px]",
                "placeholder:text-muted-foreground/50",
              )}
            />
          </div>

          {/* Кнопка эмодзи */}
          <button
            type="button"
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="Эмодзи"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </button>

          {/* Кнопка голосового сообщения */}
          <button
            type="button"
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="Голосовое сообщение"
            onClick={() => setShowVoice(true)}
          >
            <svg
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
              />
            </svg>
          </button>

          {/* Кнопка отправки -- градиент, анимация нажатия */}
          <Button
            size="icon"
            className={cn(
              "h-10 w-10 shrink-0 rounded-xl text-white shadow-sm transition-all duration-200",
              text.trim()
                ? "gradient-messenger hover:opacity-90 click-scale"
                : "bg-muted text-muted-foreground cursor-not-allowed",
            )}
            onClick={() => void handleSend()}
            disabled={isSending || !text.trim()}
            title="Отправить"
          >
            <svg
              className={cn(
                "h-5 w-5 transition-transform duration-200",
                text.trim() && "translate-x-0.5 -translate-y-0.5",
              )}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
              />
            </svg>
          </Button>
        </div>
      )}
    </div>
  );
}
