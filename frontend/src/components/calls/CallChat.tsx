import { useState, useEffect, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Send } from "lucide-react";
import { listChatMessages, sendChatMessage, type CallChatMessage } from "@/api/calls";
import { useAuthStore } from "@/stores/auth-store";
import { useWsStore } from "@/stores/ws-store";

/** Пропсы компонента эфемерного чата звонка. */
interface CallChatProps {
  /** ID активного звонка. */
  callId: string;
}

/** Эфемерный чат звонка — исчезает после завершения звонка. */
export function CallChat({ callId }: CallChatProps) {
  const [text, setText] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);

  /** Загрузка истории сообщений. */
  const { data: messages = [] } = useQuery({
    queryKey: ["call-chat", callId],
    queryFn: () => listChatMessages(callId),
    refetchInterval: false,
  });

  /** Подписка на новые сообщения через WebSocket. */
  const addListener = useWsStore((s) => s.addListener);
  const removeListener = useWsStore((s) => s.removeListener);

  useEffect(() => {
    const handler = (_event: string, payload: unknown) => {
      const msg = payload as CallChatMessage;
      if (msg.call_id === callId) {
        queryClient.setQueryData<CallChatMessage[]>(
          ["call-chat", callId],
          (old = []) => [...old, msg]
        );
      }
    };
    addListener(handler);
    return () => removeListener(handler);
  }, [callId, queryClient, addListener, removeListener]);

  /** Отправка сообщения. */
  const sendMutation = useMutation({
    mutationFn: (msg: string) => sendChatMessage(callId, msg),
    onSuccess: () => setText(""),
  });

  /** Авто-скролл при новых сообщениях. */
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [messages.length]);

  /** Обработка отправки по Enter. */
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (text.trim()) sendMutation.mutate(text.trim());
  };

  return (
    <div className="flex h-full flex-col border-l">
      {/* Заголовок */}
      <div className="border-b px-3 py-2 text-sm font-medium">
        Чат звонка
      </div>

      {/* Сообщения */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto p-3 space-y-2">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`text-sm ${msg.sender_id === user?.id ? "text-right" : ""}`}
          >
            <span className="text-xs text-muted-foreground">
              {msg.sender_name}
            </span>
            <p className="rounded-md bg-muted px-2 py-1 inline-block max-w-[80%]">
              {msg.text}
            </p>
          </div>
        ))}
      </div>

      {/* Ввод сообщения */}
      <form onSubmit={handleSubmit} className="flex gap-2 border-t p-2">
        <Input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Сообщение..."
          className="flex-1 h-8 text-sm"
        />
        <Button type="submit" size="sm" variant="ghost" disabled={!text.trim()}>
          <Send className="h-4 w-4" />
        </Button>
      </form>
    </div>
  );
}
