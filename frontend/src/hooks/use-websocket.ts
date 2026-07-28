import { useCallback, useEffect, useRef } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useWsStore } from "@/stores/ws-store";

type EventHandler = (event: string, payload: unknown) => void;

/**
 * useWebSocket подключается к WebSocket-шлюзу при аутентификации,
 * управляет подписками на комнаты и предоставляет интерфейс для слушателей событий.
 *
 * @returns connected — статус подключения, subscribe/unsubscribe — управление комнатами,
 *          onEvent — подписка на события (возвращает функцию отписки)
 */
export function useWebSocket() {
	const accessToken = useAuthStore((s) => s.accessToken);
	const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
	const { connected, connect, disconnect, subscribe, unsubscribe } =
		useWsStore();
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);

	// Подключаемся при аутентификации, отключаемся при размонтировании или logout.
	useEffect(() => {
		if (isAuthenticated && accessToken) {
			connect(accessToken);
		}
		return () => {
			disconnect();
		};
	}, [isAuthenticated, accessToken, connect, disconnect]);

	/**
	 * Подписка на обработчик событий. Обработчик получает (event, payload)
	 * для каждого серверного события. Автоматически отписывается при размонтировании.
	 */
	const onEvent = useCallback(
		(handler: EventHandler) => {
			addListener(handler);
			return () => removeListener(handler);
		},
		[addListener, removeListener],
	);

	return {
		connected,
		subscribe,
		unsubscribe,
		onEvent,
	};
}

/**
 * useWsEventListener подписывает обработчик на WS-события.
 * Автоматически обновляет подписку при изменении handler и отписывается при размонтировании.
 */
export function useWsEventListener(handler: EventHandler) {
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);
	const handlerRef = useRef<EventHandler | null>(null);

	useEffect(() => {
		if (handlerRef.current) {
			removeListener(handlerRef.current);
		}
		handlerRef.current = handler;
		addListener(handler);
		return () => {
			removeListener(handler);
		};
	}, [handler, addListener, removeListener]);
}
