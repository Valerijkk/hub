import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect } from "react";
import { getUnreadCounts, markChannelRead } from "@/api/messenger";
import { useAuthStore } from "@/stores/auth-store";
import {
	emptyUnreadSnapshot,
	useMessengerStore,
} from "@/stores/messenger-store";

/** Интервал обновления непрочитанных (60 секунд). */
const UNREAD_POLL_INTERVAL = 60_000;

export const UNREAD_QUERY_KEY = ["messenger", "unread"] as const;

export function unreadQueryKey(userId: string) {
	return [...UNREAD_QUERY_KEY, userId] as const;
}

/**
 * useUnread загружает и синхронизирует счётчики непрочитанных сообщений.
 * Обновляет messenger-store при получении данных.
 */
export function useUnread() {
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const setUnreadSnapshot = useMessengerStore((s) => s.setUnreadSnapshot);

	const query = useQuery({
		queryKey: unreadQueryKey(userId ?? "__no-user__"),
		queryFn: getUnreadCounts,
		enabled: Boolean(userId),
		refetchInterval: UNREAD_POLL_INTERVAL,
		staleTime: UNREAD_POLL_INTERVAL - 5_000,
	});

	// Полная замена карты: каналы, исчезнувшие из Redis HGetAll (MarkRead/HDel), обнуляются.
	useEffect(() => {
		if (!userId) {
			setUnreadSnapshot(emptyUnreadSnapshot());
			return;
		}
		if (query.data && useAuthStore.getState().user?.id === userId) {
			setUnreadSnapshot(query.data);
		}
	}, [query.data, setUnreadSnapshot, userId]);

	return query;
}

/**
 * useMarkRead возвращает функцию для пометки канала как прочитанного.
 * Optimistic resetUnread; при ошибке API — rollback badge (W11-A4-004).
 */
export function useMarkRead() {
	const queryClient = useQueryClient();
	const userId = useAuthStore((state) => state.user?.id ?? null);

	return useCallback(
		async (channelId: string) => {
			if (!userId || useAuthStore.getState().user?.id !== userId) return;
			try {
				await markChannelRead(channelId);
			} catch (error) {
				if (useAuthStore.getState().user?.id === userId) throw error;
				return;
			}
			if (useAuthStore.getState().user?.id !== userId) return;
			await queryClient.invalidateQueries({
				queryKey: unreadQueryKey(userId),
				exact: true,
			});
		},
		[queryClient, userId],
	);
}