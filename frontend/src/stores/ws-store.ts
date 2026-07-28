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

	/** Подключиться к WS-серверу с указанным JWT-токеном (через ticket). */
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

/** Базовая задержка reconnect (мс); растёт экспоненциально. */
const RECONNECT_BASE_MS = 1_000;
/** Потолок backoff (мс). */
const RECONNECT_MAX_MS = 30_000;

/** Префикс Sec-WebSocket-Protocol для WS ticket (ADR-WS-JWT-HANDSHAKE-2026). */
export const WS_TICKET_PROTOCOL_PREFIX = "hub.ws.ticket.";

/** Модульные поля вне zustand — таймеры не должны жить в React state. */
let lastToken: string | null = null;
let intentionalClose = false;
let reconnectAttempt = 0;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
/** Поколение connect — отменяет устаревшие async ticket-fetch. */
let connectGeneration = 0;

/** Сброс таймера reconnect (logout / intentional disconnect). */
export function clearWsReconnectTimer(): void {
	if (reconnectTimer !== null) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
}

/** Экспоненциальный backoff с потолком (для тестов и schedule). */
export function wsReconnectDelayMs(attempt: number): number {
	return Math.min(RECONNECT_MAX_MS, RECONNECT_BASE_MS * 2 ** attempt);
}

function scheduleReconnect(connectFn: (token: string) => void): void {
	clearWsReconnectTimer();
	if (intentionalClose || !lastToken) return;
	const delay = wsReconnectDelayMs(reconnectAttempt);
	reconnectAttempt += 1;
	const token = lastToken;
	reconnectTimer = setTimeout(() => {
		reconnectTimer = null;
		if (intentionalClose || !token) return;
		connectFn(token);
	}, delay);
}

/** Запрос one-shot ticket для WS handshake (без access JWT в URL). */
export async function fetchWSTicket(accessToken: string): Promise<string> {
	const res = await fetch("/api/auth/ws-ticket", {
		method: "POST",
		headers: {
			Authorization: `Bearer ${accessToken}`,
			"Content-Type": "application/json",
		},
	});
	if (!res.ok) {
		throw new Error(`ws ticket failed: ${res.status}`);
	}
	const data = (await res.json()) as { ticket?: string };
	if (!data.ticket) {
		throw new Error("ws ticket missing in response");
	}
	return data.ticket;
}

function openSocket(ticket: string): WebSocket {
	const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
	return new WebSocket(
		`${protocol}//${window.location.host}/ws`,
		`${WS_TICKET_PROTOCOL_PREFIX}${ticket}`,
	);
}

/** Стор WebSocket-подключения. Без персиста — пересоздаётся при каждой загрузке. */
export const useWsStore = create<WsState>((set, get) => ({
	connected: false,
	socket: null,
	listeners: new Set(),

	connect: (token: string) => {
		lastToken = token;
		intentionalClose = false;

		const { socket } = get();
		if (
			socket &&
			(socket.readyState === WebSocket.OPEN ||
				socket.readyState === WebSocket.CONNECTING)
		) {
			return;
		}

		const gen = ++connectGeneration;
		void (async () => {
			try {
				const ticket = await fetchWSTicket(token);
				if (gen !== connectGeneration || intentionalClose) return;

				const ws = openSocket(ticket);

				ws.onopen = () => {
					if (gen !== connectGeneration) {
						ws.close();
						return;
					}
					reconnectAttempt = 0;
					clearWsReconnectTimer();
					set({ connected: true, socket: ws });
				};

				// A10-003: обрыв сети не должен убивать real-time до remount/logout.
				ws.onclose = () => {
					if (gen !== connectGeneration) return;
					set({ connected: false, socket: null });
					if (!intentionalClose) {
						scheduleReconnect(get().connect);
					}
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
						// Некорректный JSON от WS — пропускаем, чтобы не ронять обработчики.
					}
				};

				set({ socket: ws });
			} catch {
				if (gen !== connectGeneration || intentionalClose) return;
				set({ connected: false, socket: null });
				scheduleReconnect(get().connect);
			}
		})();
	},

	disconnect: () => {
		intentionalClose = true;
		lastToken = null;
		reconnectAttempt = 0;
		connectGeneration += 1;
		clearWsReconnectTimer();
		const { socket } = get();
		if (socket) {
			socket.close();
			set({ connected: false, socket: null });
		} else {
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
