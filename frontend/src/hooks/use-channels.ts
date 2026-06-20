import { useQuery } from "@tanstack/react-query";
import { useCallback, useEffect } from "react";
import type { Channel } from "@/api/messenger";
import { getChannels } from "@/api/messenger";
import { useWsEventListener } from "@/hooks/use-websocket";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

/**
 * useChannels загружает список каналов пользователя через TanStack Query
 * и подписывается на WS-обновления (новый канал, обновление, удаление).
 */
export function useChannels() {
	const userId = useAuthStore((s) => s.user?.id);
	const setChannels = useMessengerStore((s) => s.setChannels);
	const addChannel = useMessengerStore((s) => s.addChannel);
	const updateChannel = useMessengerStore((s) => s.updateChannel);
	const incrementUnread = useMessengerStore((s) => s.incrementUnread);

	const { subscribe, unsubscribe, connected } = useWsStore();

	const query = useQuery({
		queryKey: ["messenger", "channels"],
		queryFn: getChannels,
	});

	// Синхронизация данных из query в store.
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

	// Обработка WS-событий каналов.
	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!payload || typeof payload !== "object") return;
			const data = payload as Record<string, unknown>;
			switch (event) {
				case "channel_created": {
					// Ensure required fields exist before adding
					const ch = data as unknown as Channel;
					if (ch && typeof ch.id === "string" && typeof ch.name === "string") {
						addChannel(ch);
					}
					break;
				}
				case "channel_updated":
					if (typeof data.id === "string") {
						// Only pass safe string fields to avoid objects in render
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
						updateChannel(data.id, partial);
					}
					break;
				case "new_message_notification":
					if (typeof data.channel_id === "string") {
						incrementUnread(data.channel_id);
					}
					break;
			}
		},
		[addChannel, updateChannel, incrementUnread],
	);

	useWsEventListener(handleEvent);

	return query;
}
