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
 * useRooms подписывается на список WebSocket-комнат.
 * Автоматически отписывается при изменении списка комнат или размонтировании компонента.
 * Должен вызываться как хук (не внутри callbacks).
 *
 * @param rooms - массив имён комнат (например, ["channel:abc", "project:xyz"])
 */
export function useRooms(rooms: string[]) {
	const { connected, subscribe, unsubscribe } = useWsStore();
	const prevRooms = useRef<string[]>([]);

	useEffect(() => {
		if (!connected) return;

		// Отписываемся от комнат, которые больше не нужны.
		for (const room of prevRooms.current) {
			if (!rooms.includes(room)) {
				unsubscribe(room);
			}
		}

		// Подписываемся на новые комнаты.
		for (const room of rooms) {
			if (!prevRooms.current.includes(room)) {
				subscribe(room);
			}
		}

		prevRooms.current = rooms;

		return () => {
			for (const room of rooms) {
				unsubscribe(room);
			}
			prevRooms.current = [];
		};
	}, [rooms, connected, subscribe, unsubscribe]);
}
