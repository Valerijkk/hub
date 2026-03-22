import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect } from "react";
import { getUnreadCounts, markChannelRead } from "@/api/messenger";
import { useMessengerStore } from "@/stores/messenger-store";

/** Интервал обновления непрочитанных (60 секунд). */
const UNREAD_POLL_INTERVAL = 60_000;

/**
 * useUnread загружает и синхронизирует счётчики непрочитанных сообщений.
 * Обновляет messenger-store при получении данных.
 */
export function useUnread() {
	const updateUnread = useMessengerStore((s) => s.updateUnread);

	const query = useQuery({
		queryKey: ["messenger", "unread"],
		queryFn: getUnreadCounts,
		refetchInterval: UNREAD_POLL_INTERVAL,
		staleTime: UNREAD_POLL_INTERVAL - 5_000,
	});

	// Синхронизация данных в store.
	useEffect(() => {
		if (query.data) {
			for (const [channelId, count] of Object.entries(query.data)) {
				updateUnread(channelId, count);
			}
		}
	}, [query.data, updateUnread]);

	return query;
}

/**
 * useMarkRead возвращает функцию для пометки канала как прочитанного.
 * Сбрасывает локальный store и вызывает API.
 */
export function useMarkRead() {
	const resetUnread = useMessengerStore((s) => s.resetUnread);
	const queryClient = useQueryClient();

	return useCallback(
		async (channelId: string) => {
			resetUnread(channelId);
			try {
				await markChannelRead(channelId);
				await queryClient.invalidateQueries({
					queryKey: ["messenger", "unread"],
				});
			} catch {
				// Игнорируем ошибки -- локальный сброс уже выполнен.
			}
		},
		[resetUnread, queryClient],
	);
}
