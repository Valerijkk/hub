import { useCallback, useEffect, useRef } from "react";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

/** Таймаут автоочистки индикатора набора текста (мс). */
const TYPING_TIMEOUT_MS = 4000;

/**
 * useTypingIndicator подписывается на WS-события typing для указанного канала
 * и управляет автоочисткой печатающих пользователей в messenger-store.
 *
 * @param channelId - ID канала для отслеживания
 */
export function useTypingIndicator(channelId: string | undefined) {
	const currentUserId = useAuthStore((s) => s.user?.id);
	const addTypingUser = useMessengerStore((s) => s.addTypingUser);
	const removeTypingUser = useMessengerStore((s) => s.removeTypingUser);
	const { addListener, removeListener } = useWsStore();

	/** Таймеры автоочистки по userId. */
	const timersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(
		new Map(),
	);

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (event !== "typing" || !channelId) return;
			// W11-A3-011 / P2-FE-10: битый/пустой WS typing не валит общий listener.
			if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
				return;
			}

			const data = payload as Record<string, unknown>;
			const eventChannelId = data.channel_id as string | undefined;
			if (eventChannelId !== channelId) return;

			const userId = data.user_id as string | undefined;
			const userName = data.user_name as string | undefined;
			if (!userId || !userName) return;

			// Не показываем свой собственный индикатор.
			if (userId === currentUserId) return;

			addTypingUser(channelId, userId, userName);

			// Сбросить предыдущий таймер для этого пользователя.
			const existingTimer = timersRef.current.get(userId);
			if (existingTimer) clearTimeout(existingTimer);

			// Установить новый таймер автоочистки.
			const timer = setTimeout(() => {
				removeTypingUser(channelId, userId);
				timersRef.current.delete(userId);
			}, TYPING_TIMEOUT_MS);
			timersRef.current.set(userId, timer);
		},
		[channelId, currentUserId, addTypingUser, removeTypingUser],
	);

	useEffect(() => {
		addListener(handleEvent);
		return () => {
			removeListener(handleEvent);
			// Очистить все таймеры при размонтировании.
			for (const timer of timersRef.current.values()) {
				clearTimeout(timer);
			}
			timersRef.current.clear();
		};
	}, [handleEvent, addListener, removeListener]);
}
