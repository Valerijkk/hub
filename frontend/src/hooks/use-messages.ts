import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef } from "react";
import type { Message } from "@/api/messenger";
import {
	buildMessageFromWsEvent,
	getMessages,
	getThreadReplies,
} from "@/api/messenger";
import { useWsEventListener } from "@/hooks/use-websocket";
import { useWsStore } from "@/stores/ws-store";

const PAGE_SIZE = 30;

type MessageCache = {
	pages: Message[][];
	pageParams: unknown[];
};

type ReactionToggledDelta = {
	message_id: string;
	channel_id: string;
	user_id: string;
	emoji: string;
	added: boolean;
};

const channelRoomOwners = new Map<string, number>();

function isRecord(value: unknown): value is Record<string, unknown> {
	return !!value && typeof value === "object" && !Array.isArray(value);
}

function isNonEmptyString(value: unknown): value is string {
	return typeof value === "string" && value.length > 0;
}

function isCachedMessage(value: unknown): value is Message {
	if (!isRecord(value)) return false;
	return (
		isNonEmptyString(value.id) &&
		isNonEmptyString(value.channel_id) &&
		isNonEmptyString(value.sender_id) &&
		typeof value.reply_count === "number" &&
		isNonEmptyString(value.created_at)
	);
}

function isMessageCache(value: unknown): value is MessageCache {
	if (!isRecord(value) || !Array.isArray(value.pages) || !Array.isArray(value.pageParams)) {
		return false;
	}
	return value.pages.every(
		(page) => Array.isArray(page) && page.every(isCachedMessage),
	);
}

function messageEventID(data: Record<string, unknown>): string | null {
	const value = data.message_id ?? data.id;
	return isNonEmptyString(value) ? value : null;
}

function appendMessage(old: unknown, message: Message): unknown {
	if (!isMessageCache(old)) return old;
	if (old.pages.some((page) => page.some((candidate) => candidate.id === message.id))) {
		return old;
	}
	const lastPage = old.pages[old.pages.length - 1] ?? [];
	return {
		...old,
		pages: [...old.pages.slice(0, -1), [...lastPage, message]],
	};
}

function updateMatchingMessage(
	old: unknown,
	messageID: string,
	update: (message: Message) => Message,
): unknown {
	if (!isMessageCache(old)) return old;
	let matched = false;
	const pages = old.pages.map((page) =>
		page.map((message) => {
			if (message.id !== messageID) return message;
			const updated = update(message);
			if (updated === message) return message;
			matched = true;
			return updated;
		}),
	);
	return matched ? { ...old, pages } : old;
}

function removeMatchingMessage(old: unknown, messageID: string): unknown {
	if (!isMessageCache(old)) return old;
	let matched = false;
	const pages = old.pages.map((page) => {
		const next = page.filter((message) => message.id !== messageID);
		if (next.length !== page.length) matched = true;
		return next;
	});
	return matched ? { ...old, pages } : old;
}

function editMatchingMessage(
	old: unknown,
	messageID: string,
	data: Record<string, unknown>,
): unknown {
	return updateMatchingMessage(old, messageID, (message) => ({
		...message,
		content_json:
			data.content_json && typeof data.content_json === "object"
				? (data.content_json as Record<string, unknown>)
				: message.content_json,
		content_text:
			typeof data.content_text === "string" ? data.content_text : message.content_text,
		edited_at:
			typeof data.edited_at === "string" ? data.edited_at : new Date().toISOString(),
	}));
}

function reactionDelta(
	data: Record<string, unknown>,
	channelId: string,
): ReactionToggledDelta | null {
	if (
		!isNonEmptyString(data.message_id) ||
		!isNonEmptyString(data.channel_id) ||
		data.channel_id !== channelId ||
		!isNonEmptyString(data.user_id) ||
		!isNonEmptyString(data.emoji) ||
		typeof data.added !== "boolean"
	) {
		return null;
	}
	return {
		message_id: data.message_id,
		channel_id: data.channel_id,
		user_id: data.user_id,
		emoji: data.emoji,
		added: data.added,
	};
}

function applyReactionDelta(message: Message, delta: ReactionToggledDelta): Message {
	const reactions = message.reactions ?? [];
	const index = reactions.findIndex((reaction) => reaction.emoji === delta.emoji);

	if (delta.added) {
		if (index >= 0 && reactions[index]?.user_ids.includes(delta.user_id)) return message;
		const next =
			index >= 0
				? reactions.map((reaction, reactionIndex) =>
					reactionIndex === index
						? { ...reaction, user_ids: [...reaction.user_ids, delta.user_id] }
						: reaction,
				)
				: [...reactions, { emoji: delta.emoji, user_ids: [delta.user_id] }];
		return { ...message, reactions: next };
	}

	if (index < 0 || !reactions[index]?.user_ids.includes(delta.user_id)) return message;
	const userIDs = reactions[index].user_ids.filter((userID) => userID !== delta.user_id);
	const next =
		userIDs.length === 0
			? reactions.filter((_, reactionIndex) => reactionIndex !== index)
			: reactions.map((reaction, reactionIndex) =>
				reactionIndex === index ? { ...reaction, user_ids: userIDs } : reaction,
			);
	return { ...message, reactions: next };
}

function updateReactionInCache(
	old: unknown,
	delta: ReactionToggledDelta,
): unknown {
	return updateMatchingMessage(old, delta.message_id, (message) =>
		applyReactionDelta(message, delta),
	);
}

function acquireChannelRoom(room: string, subscribe: (room: string) => void): void {
	const owners = channelRoomOwners.get(room) ?? 0;
	channelRoomOwners.set(room, owners + 1);
	if (owners === 0) subscribe(room);
}

function releaseChannelRoom(room: string, unsubscribe: (room: string) => void): void {
	const owners = channelRoomOwners.get(room) ?? 0;
	if (owners <= 1) {
		channelRoomOwners.delete(room);
		if (useWsStore.getState().connected) unsubscribe(room);
		return;
	}
	channelRoomOwners.set(room, owners - 1);
}

function useChannelRoom(
	channelId: string | undefined,
	onReconnect: () => void,
): void {
	const connected = useWsStore((state) => state.connected);
	const subscribe = useWsStore((state) => state.subscribe);
	const unsubscribe = useWsStore((state) => state.unsubscribe);
	const hasConnected = useRef(false);
	const previousConnected = useRef<boolean | null>(null);

	useEffect(() => {
		const wasConnected = previousConnected.current;
		previousConnected.current = connected;
		if (!channelId || !connected) return;

		const room = `channel:${channelId}`;
		acquireChannelRoom(room, subscribe);
		if (hasConnected.current && wasConnected === false) onReconnect();
		hasConnected.current = true;

		return () => {
			releaseChannelRoom(room, unsubscribe);
		};
	}, [channelId, connected, onReconnect, subscribe, unsubscribe]);
}

function canHandleChannelDelta(
	data: Record<string, unknown>,
	channelId: string,
): boolean {
	return isNonEmptyString(data.channel_id) && data.channel_id === channelId;
}

function canHandleThreadDelta(
	data: Record<string, unknown>,
	channelId: string,
	threadId: string,
): boolean {
	return (
		canHandleChannelDelta(data, channelId) &&
		(!isNonEmptyString(data.thread_id) || data.thread_id === threadId)
	);
}

export function useMessages(channelId: string | undefined) {
	const queryClient = useQueryClient();
	const queryKey = useMemo(
		() => ["messenger", "messages", channelId] as const,
		[channelId],
	);
	const invalidateOnReconnect = useCallback(() => {
		void queryClient.invalidateQueries({ queryKey, exact: true });
	}, [queryClient, queryKey]);
	useChannelRoom(channelId, invalidateOnReconnect);

	const query = useInfiniteQuery({
		queryKey,
		queryFn: ({ pageParam }) =>
			getMessages(channelId!, { before: pageParam, limit: PAGE_SIZE }),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (lastPage) => {
			if (lastPage.length < PAGE_SIZE) return undefined;
			return lastPage[0]?.created_at;
		},
		enabled: !!channelId,
	});

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!channelId || !isRecord(payload)) return;
			const data = payload;

			switch (event) {
				case "message_sent": {
					const message = buildMessageFromWsEvent(data);
					if (!message || message.channel_id !== channelId || message.thread_id) return;
					queryClient.setQueryData(queryKey, (old: unknown) => appendMessage(old, message));
					return;
				}
				case "message_edited": {
					if (!canHandleChannelDelta(data, channelId)) return;
					const messageID = messageEventID(data);
					if (!messageID) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						editMatchingMessage(old, messageID, data),
					);
					return;
				}
				case "message_deleted": {
					if (!canHandleChannelDelta(data, channelId)) return;
					const messageID = messageEventID(data);
					if (!messageID) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						removeMatchingMessage(old, messageID),
					);
					return;
				}
				case "reaction_toggled": {
					const delta = reactionDelta(data, channelId);
					if (!delta) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						updateReactionInCache(old, delta),
					);
				}
			}
		},
		[channelId, queryClient, queryKey],
	);

	useWsEventListener(handleEvent);

	return {
		...query,
		messages: query.data?.pages.flat() ?? [],
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}

export function useThreadMessages(
	channelId: string | undefined,
	messageId: string | undefined,
) {
	const queryClient = useQueryClient();
	const queryKey = useMemo(
		() => ["messenger", "thread", channelId, messageId] as const,
		[channelId, messageId],
	);
	const invalidateOnReconnect = useCallback(() => {
		void queryClient.invalidateQueries({ queryKey, exact: true });
	}, [queryClient, queryKey]);
	useChannelRoom(channelId, invalidateOnReconnect);

	const query = useInfiniteQuery({
		queryKey,
		queryFn: ({ pageParam }) =>
			getThreadReplies(channelId!, messageId!, {
				before: pageParam,
				limit: PAGE_SIZE,
			}),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (lastPage) => {
			if (lastPage.length < PAGE_SIZE) return undefined;
			return lastPage[0]?.created_at;
		},
		enabled: !!channelId && !!messageId,
	});

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (!channelId || !messageId || !isRecord(payload)) return;
			const data = payload;

			switch (event) {
				case "message_sent": {
					const message = buildMessageFromWsEvent(data);
					if (
						!message ||
						message.channel_id !== channelId ||
						message.thread_id !== messageId
					) {
						return;
					}
					queryClient.setQueryData(queryKey, (old: unknown) => appendMessage(old, message));
					return;
				}
				case "message_edited": {
					if (!canHandleThreadDelta(data, channelId, messageId)) return;
					const messageID = messageEventID(data);
					if (!messageID) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						editMatchingMessage(old, messageID, data),
					);
					return;
				}
				case "message_deleted": {
					if (!canHandleThreadDelta(data, channelId, messageId)) return;
					const messageID = messageEventID(data);
					if (!messageID) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						removeMatchingMessage(old, messageID),
					);
					return;
				}
				case "reaction_toggled": {
					const delta = reactionDelta(data, channelId);
					if (!delta) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						updateReactionInCache(old, delta),
					);
				}
			}
		},
		[channelId, messageId, queryClient, queryKey],
	);

	useWsEventListener(handleEvent);

	return {
		...query,
		messages: query.data?.pages.flat() ?? [],
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}