import { useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { Channel } from "@/api/messenger";
import { discoverChannels, getChannels } from "@/api/messenger";
import type { ApiError } from "@/api/client";
import { useAuthStore } from "@/stores/auth-store";

/** Базовый namespace списка каналов. Конкретный кэш всегда привязан к identity. */
export const CHANNELS_QUERY_KEY = ["messenger", "channels"] as const;
export const DISCOVERABLE_CHANNELS_QUERY_KEY = [
	"messenger",
	"channels",
	"discover",
] as const;

export function channelsQueryKey(userId: string) {
	return [...CHANNELS_QUERY_KEY, userId] as const;
}

export function discoverableChannelsQueryKey(userId: string) {
	return [...DISCOVERABLE_CHANNELS_QUERY_KEY, userId] as const;
}

function isAuthorizationError(error: unknown): error is ApiError {
	if (!error || typeof error !== "object" || !("code" in error)) return false;
	const { code } = error as ApiError;
	return code === 401 || code === 403;
}

/**
 * Вставляет канал в список (в начало), без дубликатов по id.
 * A10-001: источник правды для UI — React Query, не только zustand.
 */
export function upsertChannelInList(
	channels: Channel[] | undefined,
	channel: Channel,
): Channel[] {
	if (!channels) return [channel];
	if (channels.some((c) => c.id === channel.id)) {
		return channels.map((c) =>
			c.id === channel.id ? { ...c, ...channel } : c,
		);
	}
	return [channel, ...channels];
}

/**
 * Патчит канал в списке по id. Неизвестный id — список без изменений.
 */
export function patchChannelInList(
	channels: Channel[] | undefined,
	channelId: string,
	partial: Partial<Channel>,
): Channel[] | undefined {
	if (!channels) return channels;
	return channels.map((ch) =>
		ch.id === channelId ? { ...ch, ...partial } : ch,
	);
}

/**
 * useChannels fetches the current user's channel list through TanStack Query.
 * The messenger layout owns realtime synchronization so extra observers do not own WS.
 */
export function useChannels() {
	const userId = useAuthStore((state) => state.user?.id ?? null);
	return useQuery({
		queryKey: channelsQueryKey(userId ?? "__no-user__"),
		queryFn: getChannels,
		enabled: Boolean(userId),
	});
}

/**
 * Fetches the authenticated user browse seed of discoverable public channels.
 * Authorization failures remove this exact identity-scoped cache and stop automatic refetches.
 */
export function useDiscoverableChannels(enabled: boolean) {
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const queryClient = useQueryClient();
	const [authorizationDeniedFor, setAuthorizationDeniedFor] = useState<string | null>(
		null,
	);
	const handledAuthorizationErrorRef = useRef<string | null>(null);
	const queryKey = discoverableChannelsQueryKey(userId ?? "__no-user__");
	const isAuthorizationDenied = authorizationDeniedFor === userId;

	useEffect(() => {
		setAuthorizationDeniedFor(null);
		handledAuthorizationErrorRef.current = null;
	}, [userId]);

	const query = useQuery({
		queryKey,
		queryFn: discoverChannels,
		enabled: Boolean(userId) && enabled && !isAuthorizationDenied,
		staleTime: 0,
		refetchOnWindowFocus: true,
		retry: (failureCount, error) =>
			!isAuthorizationError(error) && failureCount < 3,
	});

	useEffect(() => {
		if (
			!userId ||
			!isAuthorizationError(query.error) ||
			handledAuthorizationErrorRef.current === userId
		) {
			return;
		}

		handledAuthorizationErrorRef.current = userId;
		setAuthorizationDeniedFor(userId);
		void (async () => {
			await queryClient.cancelQueries({ queryKey, exact: true });
			queryClient.removeQueries({ queryKey, exact: true });
		})();
	}, [query.error, queryClient, queryKey, userId]);

	return query.isError ? { ...query, data: undefined } : query;
}
