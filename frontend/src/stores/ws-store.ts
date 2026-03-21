import { create } from "zustand";

/** Обработчик WS-событий: получает имя события и payload. */
type EventHandler = (event: string, payload: unknown) => void;

/** Сообщение от WebSocket-сервера. */
interface ServerMessage {
  /** Тип сообщения: "event" для событий. */
  type: string;
  /** Имя события (например, "message_sent", "issue_updated"). */
  event?: string;
  /** WS-комната, из которой пришло событие. */
  room?: string;
  /** Данные события. */
  payload?: unknown;
}

/**
 * Состояние стора WebSocket-подключения.
 * Управляет единственным WS-соединением для всего приложения.
 */
interface WsState {
  /** Подключён ли WebSocket. */
  connected: boolean;
  /** Экземпляр WebSocket (null если не подключён). */
  socket: WebSocket | null;
  /** Набор зарегистрированных слушателей событий. */
  listeners: Set<EventHandler>;

  /** Подключиться к WS-серверу с указанным JWT-токеном. */
  connect: (token: string) => void;
  /** Отключиться от WS-сервера. */
  disconnect: () => void;
  /** Подписаться на WS-комнату (например, "channel:abc"). */
  subscribe: (room: string) => void;
  /** Отписаться от WS-комнаты. */
  unsubscribe: (room: string) => void;
  /** Добавить слушатель событий. */
  addListener: (handler: EventHandler) => void;
  /** Удалить слушатель событий. */
  removeListener: (handler: EventHandler) => void;
}

/** Стор WebSocket-подключения. Без персиста — пересоздаётся при каждой загрузке. */
export const useWsStore = create<WsState>((set, get) => ({
  connected: false,
  socket: null,
  listeners: new Set(),

  connect: (token: string) => {
    const { socket } = get();
    if (socket) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(
      `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(token)}`,
    );

    ws.onopen = () => {
      set({ connected: true, socket: ws });
    };

    ws.onclose = () => {
      set({ connected: false, socket: null });
    };

    ws.onerror = () => {
      ws.close();
    };

    ws.onmessage = (event: MessageEvent) => {
      try {
        const msg = JSON.parse(event.data as string) as ServerMessage;
        if (msg.type === "event" && msg.event) {
          const { listeners } = get();
          for (const handler of listeners) {
            handler(msg.event, msg.payload);
          }
        }
      } catch {
        // Ignore malformed messages.
      }
    };

    set({ socket: ws });
  },

  disconnect: () => {
    const { socket } = get();
    if (socket) {
      socket.close();
      set({ connected: false, socket: null });
    }
  },

  subscribe: (room: string) => {
    const { socket } = get();
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: "subscribe", room }));
    }
  },

  unsubscribe: (room: string) => {
    const { socket } = get();
    if (socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: "unsubscribe", room }));
    }
  },

  addListener: (handler: EventHandler) => {
    set((state) => {
      const next = new Set(state.listeners);
      next.add(handler);
      return { listeners: next };
    });
  },

  removeListener: (handler: EventHandler) => {
    set((state) => {
      const next = new Set(state.listeners);
      next.delete(handler);
      return { listeners: next };
    });
  },
}));
