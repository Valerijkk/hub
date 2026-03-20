import { create } from "zustand";

type EventHandler = (event: string, payload: unknown) => void;

interface ServerMessage {
  type: string;
  event?: string;
  room?: string;
  payload?: unknown;
}

interface WsState {
  connected: boolean;
  socket: WebSocket | null;
  listeners: Set<EventHandler>;

  connect: (token: string) => void;
  disconnect: () => void;
  subscribe: (room: string) => void;
  unsubscribe: (room: string) => void;
  addListener: (handler: EventHandler) => void;
  removeListener: (handler: EventHandler) => void;
}

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
