import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef } from "react";
import type { Message } from "@/api/messenger";
import { getMessages, getThreadReplies } from "@/api/messenger";
import { useWsStore } from "@/stores/ws-store";

/** Количество сообщений за один запрос. */
const PAGE_SIZE = 30;

/**
 * useMessages загружает сообщения канала с cursor-пагинацией (вверх)
 * и подписывается на WS room channel:{channelId} для real-time обновлений.
 */
export function useMessages(channelId: string | undefined) {
	const queryClient = useQueryClient();
	const { subscribe, unsubscribe, addListener, removeListener, connected } =
		useWsStore();

	const query = useInfiniteQuery({
		queryKey: ["messenger", "messages", channelId],
		queryFn: ({ pageParam }) =>
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
	const handlerRef = useRef<((event: string, payload: unknown) => void) | null>(
		null,
	);

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
					// WS event is MessageSentEvent (partial), not full Message.
					// Build a safe Message from available fields.
					const evt = data as Record<string, unknown>;
					const message: Message = {
						id: (evt.message_id ?? evt.id ?? "") as string,
						channel_id: (evt.channel_id ?? channelId) as string,
						sender_id: (evt.sender_id ?? "") as string,
						message_type: (evt.message_type as string) ?? undefined,
						poll_id: (evt.poll_id as string) ?? undefined,
						content_text: (evt.content_text as string) ?? undefined,
						content_json: (evt.content_json as Record<string, unknown>) ?? undefined,
						reply_count: (evt.reply_count as number) ?? 0,
						created_at: (evt.created_at as string) ?? new Date().toISOString(),
						thread_id: (evt.thread_id as string) ?? undefined,
						reactions: (evt.reactions as Message["reactions"]) ?? undefined,
						attachments: (evt.attachments as Message["attachments"]) ?? undefined,
						forwarded_from_id: (evt.forwarded_from_id as string) ?? undefined,
						edited_at: (evt.edited_at as string) ?? undefined,
						deleted_at: (evt.deleted_at as string) ?? undefined,
						voice_url: (evt.voice_url as string) ?? undefined,
						voice_duration_ms: (evt.voice_duration_ms as number) ?? undefined,
						mentions: (evt.mentions as string[]) ?? undefined,
						mention_groups: (evt.mention_groups as string[]) ?? undefined,
					};
					if (!message.id) break; // Skip malformed events
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

	useEffect(() => {
		if (handlerRef.current) {
			removeListener(handlerRef.current);
		}
		handlerRef.current = handleEvent;
		addListener(handleEvent);
		return () => {
			removeListener(handleEvent);
		};
	}, [handleEvent, addListener, removeListener]);

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
	const { addListener, removeListener } = useWsStore();

	const query = useInfiniteQuery({
		queryKey: ["messenger", "thread", channelId, messageId],
		queryFn: ({ pageParam }) =>
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

	// Обработка WS-событий для треда.
	const handlerRef = useRef<((event: string, payload: unknown) => void) | null>(
		null,
	);

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!channelId || !messageId) return;
			const data = payload as Record<string, unknown>;
			const threadId = data.thread_id as string | undefined;
			if (threadId !== messageId) return;

			const queryKey = ["messenger", "thread", channelId, messageId];

			if (event === "message_sent") {
				// Build safe Message from WS event
				const evt = data;
				const message: Message = {
					id: (evt.message_id ?? evt.id ?? "") as string,
					channel_id: (evt.channel_id ?? channelId) as string,
					sender_id: (evt.sender_id ?? "") as string,
					content_text: (evt.content_text as string) ?? undefined,
					content_json: (evt.content_json as Record<string, unknown>) ?? undefined,
					reply_count: (evt.reply_count as number) ?? 0,
					created_at: (evt.created_at as string) ?? new Date().toISOString(),
					thread_id: (evt.thread_id as string) ?? undefined,
					reactions: (evt.reactions as Message["reactions"]) ?? undefined,
					attachments: (evt.attachments as Message["attachments"]) ?? undefined,
					forwarded_from_id: (evt.forwarded_from_id as string) ?? undefined,
					edited_at: (evt.edited_at as string) ?? undefined,
					deleted_at: (evt.deleted_at as string) ?? undefined,
					voice_url: (evt.voice_url as string) ?? undefined,
					voice_duration_ms: (evt.voice_duration_ms as number) ?? undefined,
					mentions: (evt.mentions as string[]) ?? undefined,
					mention_groups: (evt.mention_groups as string[]) ?? undefined,
				};
				if (!message.id) return;
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

	useEffect(() => {
		if (handlerRef.current) {
			removeListener(handlerRef.current);
		}
		handlerRef.current = handleEvent;
		addListener(handleEvent);
		return () => {
			removeListener(handleEvent);
		};
	}, [handleEvent, addListener, removeListener]);

	const messages = query.data?.pages.flat() ?? [];

	return {
		...query,
		messages,
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}
