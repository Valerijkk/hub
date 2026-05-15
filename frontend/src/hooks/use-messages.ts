import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect } from "react";
import type { Message } from "@/api/messenger";
import { buildMessageFromWsEvent, getMessages, getThreadReplies } from "@/api/messenger";
import { useWsStore } from "@/stores/ws-store";
import { useWsEventListener } from "@/hooks/use-websocket";

/** Количество сообщений за один запрос. */
const PAGE_SIZE = 30;

/**
 * useMessages загружает сообщения канала с cursor-пагинацией (вверх)
 * и подписывается на WS room channel:{channelId} для real-time обновлений.
 */
export function useMessages(channelId: string | undefined) {
	const queryClient = useQueryClient();
	const { subscribe, unsubscribe, connected } = useWsStore();

	const query = useInfiniteQuery({
		queryKey: ["messenger", "messages", channelId],
		queryFn: ({ pageParam }) =>
			// biome-ignore lint/style/noNonNullAssertion: enabled flag guards channelId being defined
			getMessages(channelId!, { before: pageParam, limit: PAGE_SIZE }),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (lastPage) => {
			if (lastPage.length < PAGE_SIZE) return undefined;
			// Backend expects RFC3339 timestamp for cursor-based pagination.
			return lastPage[0]?.created_at;
		},
		enabled: !!channelId,
	});

	// Подписка на WS room канала.
	useEffect(() => {
		if (!connected || !channelId) return;
		const room = `channel:${channelId}`;
		subscribe(room);
		return () => {
			unsubscribe(room);
		};
	}, [connected, channelId, subscribe, unsubscribe]);

	// Обработка WS-событий сообщений.
	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!channelId) return;
			const data = payload as Record<string, unknown>;
			const msgChannelId = data.channel_id as string | undefined;

			// Игнорируем события для других каналов.
			if (msgChannelId && msgChannelId !== channelId) return;

			const queryKey = ["messenger", "messages", channelId];

			switch (event) {
				case "message_sent": {
					const message = buildMessageFromWsEvent(data, channelId);
					if (!message) break;
					queryClient.setQueryData(queryKey, (old: unknown) => {
						if (!old) return old;
						const typed = old as { pages: Message[][]; pageParams: unknown[] };
						// Deduplicate: skip if message already exists
						const exists = typed.pages.some((page) => page.some((m) => m.id === message.id));
						if (exists) return old;
						const lastPage = typed.pages[typed.pages.length - 1] ?? [];
						return {
							...typed,
							pages: [...typed.pages.slice(0, -1), [...lastPage, message]],
						};
					});
					break;
				}
				case "message_edited": {
					const evt = data as Record<string, unknown>;
					const editedId = (evt.message_id ?? evt.id) as string;
					if (!editedId) break;
					queryClient.setQueryData(queryKey, (old: unknown) => {
						if (!old) return old;
						const typed = old as { pages: Message[][]; pageParams: unknown[] };
						return {
							...typed,
							pages: typed.pages.map((page) =>
								page.map((msg) =>
									msg.id === editedId
										? {
												...msg,
												content_json: (evt.content_json as Record<string, unknown>) ?? msg.content_json,
												content_text: (evt.content_text as string) ?? msg.content_text,
												edited_at: (evt.edited_at as string) ?? new Date().toISOString(),
											}
										: msg,
								),
							),
						};
					});
					break;
				}
				case "message_deleted": {
					const deletedEvt = data as Record<string, unknown>;
					const deletedId = (deletedEvt.message_id ?? deletedEvt.id) as string;
					if (!deletedId) break;
					queryClient.setQueryData(queryKey, (old: unknown) => {
						if (!old) return old;
						const typed = old as { pages: Message[][]; pageParams: unknown[] };
						return {
							...typed,
							pages: typed.pages.map((page) =>
								page.filter((msg) => msg.id !== deletedId),
							),
						};
					});
					break;
				}
				case "reaction_toggled": {
					const reactEvt = data as Record<string, unknown>;
					const msgId = (reactEvt.message_id ?? reactEvt.id) as string;
					const reactions = reactEvt.reactions as Message["reactions"];
					queryClient.setQueryData(queryKey, (old: unknown) => {
						if (!old) return old;
						const typed = old as { pages: Message[][]; pageParams: unknown[] };
						return {
							...typed,
							pages: typed.pages.map((page) =>
								page.map((msg) =>
									msg.id === msgId ? { ...msg, reactions } : msg,
								),
							),
						};
					});
					break;
				}
			}
		},
		[channelId, queryClient],
	);

	useWsEventListener(handleEvent);

	/** Все сообщения из всех страниц (отсортированные по времени). */
	const messages = query.data?.pages.flat() ?? [];

	return {
		...query,
		messages,
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}

/**
 * useThreadMessages загружает ответы в треде с cursor-пагинацией (вверх)
 * и подписывается на WS room для обновлений.
 */
export function useThreadMessages(
	channelId: string | undefined,
	messageId: string | undefined,
) {
	const queryClient = useQueryClient();

	const query = useInfiniteQuery({
		queryKey: ["messenger", "thread", channelId, messageId],
		queryFn: ({ pageParam }) =>
			// biome-ignore lint/style/noNonNullAssertion: enabled flag guards channelId/messageId being defined
			getThreadReplies(channelId!, messageId!, {
				before: pageParam,
				limit: PAGE_SIZE,
			}),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (lastPage) => {
			if (lastPage.length < PAGE_SIZE) return undefined;
			// Backend expects RFC3339 timestamp for cursor-based pagination.
			return lastPage[0]?.created_at;
		},
		enabled: !!channelId && !!messageId,
	});

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!channelId || !messageId) return;
			const data = payload as Record<string, unknown>;
			const threadId = data.thread_id as string | undefined;
			if (threadId !== messageId) return;

			const queryKey = ["messenger", "thread", channelId, messageId];

			if (event === "message_sent") {
				const message = buildMessageFromWsEvent(data, channelId);
				if (!message) return;
				queryClient.setQueryData(queryKey, (old: unknown) => {
					if (!old) return old;
					const typed = old as { pages: Message[][]; pageParams: unknown[] };
					const exists = typed.pages.some((page) => page.some((m) => m.id === message.id));
					if (exists) return old;
					const lastPage = typed.pages[typed.pages.length - 1] ?? [];
					return {
						...typed,
						pages: [...typed.pages.slice(0, -1), [...lastPage, message]],
					};
				});
			}
		},
		[channelId, messageId, queryClient],
	);

	useWsEventListener(handleEvent);

	const messages = query.data?.pages.flat() ?? [];

	return {
		...query,
		messages,
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}
