import { useState, useCallback, useRef, type KeyboardEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { sendMessage } from "@/api/messenger";
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
 * MessageComposer -- редактор сообщений.
 * Полноширинный textarea с отправкой по Enter (Shift+Enter -- перенос строки).
 * Поддерживает голосовые сообщения и показывает ошибки через toast.
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

  /** Отправка текстового сообщения. */
  const handleSend = useCallback(async () => {
    const trimmed = text.trim();
    if (!trimmed) return;

    setIsSending(true);
    try {
      await sendMessage(channelId, {
        content_text: trimmed,
        thread_id: threadId,
      });
      // Инвалидация кеша сообщений — страховка на случай отсутствия WS.
      void queryClient.invalidateQueries({
        queryKey: ["messenger", "messages", channelId],
      });
      // Очистка поля после успешной отправки.
      setText("");
      // Сброс высоты textarea.
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
  }, [channelId, threadId, text, queryClient]);

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

  /** Автоматическая подгонка высоты textarea. */
  const handleInput = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      setText(e.target.value);
      const el = e.target;
      el.style.height = "auto";
      el.style.height = `${Math.min(el.scrollHeight, 200)}px`;
    },
    [],
  );

  /** Отправка голосового сообщения. */
  const handleVoiceSend = useCallback(
    async (voiceUrl: string, durationMs: number) => {
      try {
        await sendMessage(channelId, {
          voice_url: voiceUrl,
          voice_duration_ms: durationMs,
          thread_id: threadId,
        });
        // Инвалидация кеша сообщений — страховка на случай отсутствия WS.
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

  return (
    <div className="space-y-2">
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

          {/* Поле ввода — полная ширина, скруглённое, стеклянный эффект */}
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
                "min-h-[44px] max-h-[200px]",
                "placeholder:text-muted-foreground/50",
              )}
            />
          </div>

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

          {/* Кнопка отправки -- градиент */}
          <Button
            size="icon"
            className="h-10 w-10 shrink-0 rounded-xl gradient-messenger text-white shadow-sm hover:opacity-90"
            onClick={() => void handleSend()}
            disabled={isSending || !text.trim()}
            title="Отправить"
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
                d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
              />
            </svg>
          </Button>
        </div>
      )}
    </div>
  );
}
