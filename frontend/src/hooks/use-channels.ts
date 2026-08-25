import { useQuery } from "@tanstack/react-query";
import type { Channel } from "@/api/messenger";
import { getChannels } from "@/api/messenger";

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
 * useChannels fetches the user channel list through TanStack Query.
 * The messenger layout owns realtime synchronization so extra observers do not own WS.
 */
export function useChannels() {
	return useQuery({
		queryKey: CHANNELS_QUERY_KEY,
		queryFn: getChannels,
	});
}
