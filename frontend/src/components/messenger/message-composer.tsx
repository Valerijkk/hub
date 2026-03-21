import { useState, useCallback, useRef } from "react";
import type { Value } from "@udecode/plate";
import {
  Plate,
  PlateContent,
  usePlateEditor,
} from "@udecode/plate/react";
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

/** Пустое значение Plate. */
const EMPTY_VALUE: Value = [{ type: "p", children: [{ text: "" }] }];

/** Извлечение plain text из Plate JSON. */
function extractText(value: Value): string {
  const parts: string[] = [];
  function walk(nodes: unknown[]) {
    for (const node of nodes) {
      const n = node as Record<string, unknown>;
      if (typeof n.text === "string") {
        parts.push(n.text);
      }
      if (Array.isArray(n.children)) {
        walk(n.children);
      }
    }
  }
  walk(value);
  return parts.join("").trim();
}

/**
 * MessageComposer -- редактор сообщений на базе Plate.
 * Полноширинный ввод с кнопками, фиксирован внизу чата.
 * Поддерживает rich text, отправку по Enter, голосовые сообщения.
 */
export function MessageComposer({
  channelId,
  threadId,
  placeholder = "Написать сообщение...",
}: MessageComposerProps) {
  const [isSending, setIsSending] = useState(false);
  const [showVoice, setShowVoice] = useState(false);
  const editorValueRef = useRef<Value>(EMPTY_VALUE);

  const editor = usePlateEditor({
    value: EMPTY_VALUE,
  });

  const handleSend = useCallback(async () => {
    const value = editorValueRef.current;
    const text = extractText(value);
    if (!text) return;

    setIsSending(true);
    try {
      await sendMessage(channelId, {
        content_json: { children: value } as unknown as Record<string, unknown>,
        content_text: text,
        thread_id: threadId,
      });
      // Очистка редактора.
      editor.tf.reset();
      editorValueRef.current = EMPTY_VALUE;
    } finally {
      setIsSending(false);
    }
  }, [channelId, threadId, editor]);

  /** Отправка голосового сообщения. */
  const handleVoiceSend = useCallback(
    async (voiceUrl: string, durationMs: number) => {
      await sendMessage(channelId, {
        voice_url: voiceUrl,
        voice_duration_ms: durationMs,
        thread_id: threadId,
      });
      setShowVoice(false);
    },
    [channelId, threadId],
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
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
            </svg>
          </button>

          {/* Поле ввода — полная ширина, скруглённое */}
          <div className="flex-1 rounded-xl border bg-background shadow-sm transition-shadow focus-within:shadow-md focus-within:border-ring/50">
            <Plate
              editor={editor}
              onValueChange={({ value }) => {
                editorValueRef.current = value;
              }}
            >
              <PlateContent
                placeholder={placeholder}
                className={cn(
                  "min-h-[44px] max-h-[200px] overflow-y-auto px-4 py-3 text-sm focus:outline-none",
                  "[&_[data-slate-placeholder]]:!text-muted-foreground/50",
                )}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    void handleSend();
                  }
                }}
              />
            </Plate>
          </div>

          {/* Кнопка голосового сообщения */}
          <button
            type="button"
            className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            title="Голосовое сообщение"
            onClick={() => setShowVoice(true)}
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
            </svg>
          </button>

          {/* Кнопка отправки — градиент */}
          <Button
            size="icon"
            className="h-10 w-10 shrink-0 rounded-xl gradient-messenger text-white shadow-sm hover:opacity-90"
            onClick={() => void handleSend()}
            disabled={isSending}
            title="Отправить"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </Button>
        </div>
      )}
    </div>
  );
}
