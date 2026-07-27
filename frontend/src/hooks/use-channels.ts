import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect } from "react";
import type { Channel } from "@/api/messenger";
import { getChannels } from "@/api/messenger";
import { useWsEventListener } from "@/hooks/use-websocket";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

/** Ключ кэша списка каналов (единый для Query и WS). */
export const CHANNELS_QUERY_KEY = ["messenger", "channels"] as const;

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
 * useChannels загружает список каналов пользователя через TanStack Query
 * и подписывается на WS-обновления (новый канал, обновление, удаление).
 */
export function useChannels() {
	const queryClient = useQueryClient();
	const userId = useAuthStore((s) => s.user?.id);
	const setChannels = useMessengerStore((s) => s.setChannels);
	const incrementUnread = useMessengerStore((s) => s.incrementUnread);

	const { subscribe, unsubscribe, connected } = useWsStore();

	const query = useQuery({
		queryKey: CHANNELS_QUERY_KEY,
		queryFn: getChannels,
	});

	// Синхронизация данных из query в store (unread / side panels).
	useEffect(() => {
		if (query.data) {
			setChannels(query.data);
		}
	}, [query.data, setChannels]);

	// Подписка на WS room для обновлений каналов.
	useEffect(() => {
		if (!connected || !userId) return;
		const room = `user:${userId}:channels`;
		subscribe(room);
		return () => {
			unsubscribe(room);
		};
	}, [connected, userId, subscribe, unsubscribe]);

	// Обработка WS-событий каналов — пишем в Query (UI читает query.data).
	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!payload || typeof payload !== "object") return;
			const data = payload as Record<string, unknown>;
			switch (event) {
				case "channel_created": {
					// Проверяем обязательные поля — WS-payload может быть неполным.
					const ch = data as unknown as Channel;
					if (ch && typeof ch.id === "string" && typeof ch.name === "string") {
						queryClient.setQueryData<Channel[]>(CHANNELS_QUERY_KEY, (old) =>
							upsertChannelInList(old, ch),
						);
					}
					break;
				}
				case "channel_updated":
					if (typeof data.id === "string") {
						// В partial только строковые поля — объекты в render ломают React.
						const partial: Partial<Channel> = {};
						if (typeof data.name === "string") partial.name = data.name;
						if (typeof data.description === "string")
							partial.description = data.description;
						if (typeof data.channel_type === "string")
							partial.channel_type =
								data.channel_type as Channel["channel_type"];
						if (typeof data.is_readonly === "boolean")
							partial.is_readonly = data.is_readonly;
						if (typeof data.updated_at === "string")
							partial.updated_at = data.updated_at;
						queryClient.setQueryData<Channel[]>(CHANNELS_QUERY_KEY, (old) =>
							patchChannelInList(old, data.id as string, partial),
						);
					}
					break;
				case "new_message_notification":
					if (typeof data.channel_id === "string") {
						incrementUnread(data.channel_id);
					}
					break;
			}
		},
		[queryClient, incrementUnread],
	);

	useWsEventListener(handleEvent);

	return query;
}
