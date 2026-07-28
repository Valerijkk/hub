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
	const setUnreadCounts = useMessengerStore((s) => s.setUnreadCounts);

	const query = useQuery({
		queryKey: ["messenger", "unread"],
		queryFn: getUnreadCounts,
		refetchInterval: UNREAD_POLL_INTERVAL,
		staleTime: UNREAD_POLL_INTERVAL - 5_000,
	});

	// Полная замена карты: каналы, исчезнувшие из Redis HGetAll (MarkRead/HDel), обнуляются.
	useEffect(() => {
		if (query.data) {
			setUnreadCounts(query.data);
		}
	}, [query.data, setUnreadCounts]);

	return query;
}

/**
 * useMarkRead возвращает функцию для пометки канала как прочитанного.
 * Optimistic resetUnread; при ошибке API — rollback badge (W11-A4-004).
 */
export function useMarkRead() {
	const resetUnread = useMessengerStore((s) => s.resetUnread);
	const updateUnread = useMessengerStore((s) => s.updateUnread);
	const queryClient = useQueryClient();

	return useCallback(
		async (channelId: string) => {
			const previous =
				useMessengerStore.getState().unreadCounts[channelId] ?? 0;
			resetUnread(channelId);
			try {
				await markChannelRead(channelId);
				await queryClient.invalidateQueries({
					queryKey: ["messenger", "unread"],
				});
			} catch {
				updateUnread(channelId, previous);
			}
		},
		[resetUnread, updateUnread, queryClient],
	);
}
