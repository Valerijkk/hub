import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Message } from "@/api/messenger";
import {
	buildMessageFromWsEvent,
	getMessages,
	getThreadReplies,
} from "@/api/messenger";
import { useWsEventListener } from "@/hooks/use-websocket";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

const PAGE_SIZE = 30;
const MAX_PENDING_THREAD_MESSAGES = 100;
const MAX_PENDING_THREAD_MUTATIONS = 100;

export const MESSAGES_QUERY_KEY = ["messenger", "messages"] as const;

export function messagesQueryKey(
	channelId: string | undefined,
	userId: string,
) {
	return [...MESSAGES_QUERY_KEY, channelId, userId] as const;
}

export const THREAD_MESSAGES_QUERY_KEY = [
	"messenger",
	"thread-messages",
] as const;
export const THREAD_DETAIL_QUERY_KEY = ["messenger", "thread"] as const;

export function threadMessagesQueryKey(
	channelId: string | undefined,
	messageId: string | undefined,
	userId: string,
) {
	return [...THREAD_MESSAGES_QUERY_KEY, channelId, messageId, userId] as const;
}

export function threadDetailQueryKey(
	channelId: string,
	messageId: string,
	userId: string,
) {
	return [...THREAD_DETAIL_QUERY_KEY, channelId, messageId, userId] as const;
}

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

type PendingThreadMutation =
	| { kind: "edit"; messageID: string; data: Record<string, unknown> }
	| { kind: "delete"; messageID: string }
	| { kind: "reaction"; delta: ReactionToggledDelta };

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

function queuePendingThreadMessage(
	pendingMessages: Map<string, Message>,
	message: Message,
): boolean {
	let evictedMessage = false;
	if (pendingMessages.has(message.id)) {
		pendingMessages.delete(message.id);
	} else if (pendingMessages.size >= MAX_PENDING_THREAD_MESSAGES) {
		const oldestMessageID = pendingMessages.keys().next().value;
		if (typeof oldestMessageID === "string") {
			pendingMessages.delete(oldestMessageID);
			evictedMessage = true;
		}
	}
	pendingMessages.set(message.id, message);
	return evictedMessage;
}

function queuePendingThreadMutation(
	pendingMutations: PendingThreadMutation[],
	mutation: PendingThreadMutation,
): boolean {
	const overflowed = pendingMutations.length >= MAX_PENDING_THREAD_MUTATIONS;
	if (overflowed) pendingMutations.shift();
	pendingMutations.push(mutation);
	return overflowed;
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

function applyPendingThreadMutations(
	cache: MessageCache | undefined,
	pendingMutations: readonly PendingThreadMutation[],
): MessageCache | undefined {
	let next: unknown = cache;
	for (const mutation of pendingMutations) {
		switch (mutation.kind) {
			case "edit":
				next = editMatchingMessage(next, mutation.messageID, mutation.data);
				break;
			case "delete":
				next = removeMatchingMessage(next, mutation.messageID);
				break;
			case "reaction":
				next = updateReactionInCache(next, mutation.delta);
				break;
		}
	}
	return isMessageCache(next) ? next : undefined;
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
	isRevoked: boolean,
): void {
	const connected = useWsStore((state) => state.connected);
	const subscribe = useWsStore((state) => state.subscribe);
	const unsubscribe = useWsStore((state) => state.unsubscribe);
	const hasConnected = useRef(false);
	const previousConnected = useRef<boolean | null>(null);

	useEffect(() => {
		const wasConnected = previousConnected.current;
		previousConnected.current = connected;
		if (!channelId || !connected || isRevoked) return;

		const room = `channel:${channelId}`;
		acquireChannelRoom(room, subscribe);
		if (hasConnected.current && wasConnected === false) onReconnect();
		hasConnected.current = true;

		return () => {
			releaseChannelRoom(room, unsubscribe);
		};
	}, [channelId, connected, isRevoked, onReconnect, subscribe, unsubscribe]);
}

function canHandleChannelDelta(
	data: Record<string, unknown>,
	channelId: string,
): boolean {
	return isNonEmptyString(data.channel_id) && data.channel_id === channelId;
}

export function getMessageCacheTargets(
	message: Message,
	channelId: string,
	rootId?: string,
): { timeline: boolean; thread: boolean } {
	if (message.channel_id !== channelId) {
		return { timeline: false, thread: false };
	}
	const isReply = Boolean(message.thread_id);
	return {
		timeline: !isReply || message.also_send_to_channel === true,
		thread: Boolean(rootId) && message.thread_id === rootId,
	};
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
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const isRevoked = useMessengerStore((state) =>
		channelId ? state.revokedChannelIds.has(channelId) : false,
	);
	const queryKey = useMemo(
		() => messagesQueryKey(channelId, userId ?? "__no-user__"),
		[channelId, userId],
	);
	const invalidateOnReconnect = useCallback(() => {
		if (!userId || useAuthStore.getState().user?.id !== userId) return;
		void queryClient.invalidateQueries({ queryKey, exact: true });
	}, [queryClient, queryKey, userId]);
	useChannelRoom(
		userId ? channelId : undefined,
		invalidateOnReconnect,
		isRevoked || !userId,
	);

	const query = useInfiniteQuery({
		queryKey,
		queryFn: ({ pageParam }) =>
			getMessages(channelId!, { before: pageParam, limit: PAGE_SIZE }),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (lastPage) => {
			if (lastPage.length < PAGE_SIZE) return undefined;
			return lastPage[0]?.created_at;
		},
		enabled: !!channelId && !!userId && !isRevoked,
	});

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (
				!channelId ||
				!userId ||
				useAuthStore.getState().user?.id !== userId ||
				useMessengerStore.getState().isChannelRevoked(channelId) ||
				!isRecord(payload)
			) {
				return;
			}
			const data = payload;

			switch (event) {
				case "message_sent": {
					const message = buildMessageFromWsEvent(data);
					if (!message || !getMessageCacheTargets(message, channelId).timeline) return;
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
		[channelId, queryClient, queryKey, userId],
	);

	useWsEventListener(handleEvent);

	return {
		...query,
		data: userId && !isRevoked ? query.data : undefined,
		messages: userId && !isRevoked ? (query.data?.pages.flat() ?? []) : [],
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}

export function useThreadMessages(
	channelId: string | undefined,
	messageId: string | undefined,
) {
	const queryClient = useQueryClient();
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const isRevoked = useMessengerStore((state) =>
		channelId ? state.revokedChannelIds.has(channelId) : false,
	);
	const queryKey = useMemo(
		() =>
			threadMessagesQueryKey(
				channelId,
				messageId,
				userId ?? "__no-user__",
			),
		[channelId, messageId, userId],
	);
	const pendingThreadMessagesRef = useRef(new Map<string, Message>());
	const pendingThreadMutationsRef = useRef<PendingThreadMutation[]>([]);
	const pendingThreadScopeRef = useRef<string | null>(null);
	const pendingThreadOverflowScopeRef = useRef<string | null>(null);
	const pendingThreadMutationOverflowScopeRef = useRef<string | null>(null);
	const needsCanonicalThreadReconcileScopeRef = useRef<string | null>(null);
	const canonicalThreadRecoveryScopeRef = useRef<string | null>(null);
	const [mutationOverlayVersion, setMutationOverlayVersion] = useState(0);
	const threadScope =
		userId && channelId && messageId && !isRevoked
			? JSON.stringify([channelId, messageId, userId])
			: null;
	const invalidateOnReconnect = useCallback(() => {
		if (!userId || useAuthStore.getState().user?.id !== userId) return;
		void queryClient.invalidateQueries({ queryKey, exact: true });
	}, [queryClient, queryKey, userId]);
	useChannelRoom(
		userId ? channelId : undefined,
		invalidateOnReconnect,
		isRevoked || !userId,
	);

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
		enabled: !!channelId && !!messageId && !!userId && !isRevoked,
	});

	useEffect(() => {
		if (pendingThreadScopeRef.current === threadScope) return;
		pendingThreadMessagesRef.current.clear();
		pendingThreadMutationsRef.current = [];
		pendingThreadScopeRef.current = threadScope;
		pendingThreadOverflowScopeRef.current = null;
		pendingThreadMutationOverflowScopeRef.current = null;
		needsCanonicalThreadReconcileScopeRef.current = null;
		canonicalThreadRecoveryScopeRef.current = null;
		setMutationOverlayVersion((version) => version + 1);
	}, [threadScope]);

	useEffect(
		() => () => {
			pendingThreadMessagesRef.current.clear();
			pendingThreadMutationsRef.current = [];
			pendingThreadScopeRef.current = null;
			pendingThreadOverflowScopeRef.current = null;
			pendingThreadMutationOverflowScopeRef.current = null;
			needsCanonicalThreadReconcileScopeRef.current = null;
			canonicalThreadRecoveryScopeRef.current = null;
		},
		[],
	);

	useEffect(() => {
		if (
			!threadScope ||
			pendingThreadScopeRef.current !== threadScope ||
			!query.data ||
			query.isFetching
		) {
			return;
		}
		const pendingMessages = [...pendingThreadMessagesRef.current.values()];
		const requiresRecovery =
			pendingThreadOverflowScopeRef.current === threadScope ||
			pendingThreadMutationOverflowScopeRef.current === threadScope ||
			needsCanonicalThreadReconcileScopeRef.current === threadScope;
		const recoveryInFlight = canonicalThreadRecoveryScopeRef.current === threadScope;
		if (pendingMessages.length === 0 && !requiresRecovery && !recoveryInFlight) {
			return;
		}
		if (pendingMessages.length > 0) {
			queryClient.setQueryData(queryKey, (old: unknown) => {
				let next = old;
				for (const message of pendingMessages) {
					next = appendMessage(next, message);
				}
				return next;
			});
			pendingThreadMessagesRef.current.clear();
		}
		if (!recoveryInFlight) {
			if (!requiresRecovery) return;
			pendingThreadOverflowScopeRef.current = null;
			needsCanonicalThreadReconcileScopeRef.current = null;
			canonicalThreadRecoveryScopeRef.current = threadScope;
			void queryClient.invalidateQueries({ queryKey, exact: true });
			return;
		}
		if (query.isError) return;
		const recoveryRearmed =
			pendingThreadOverflowScopeRef.current === threadScope ||
			needsCanonicalThreadReconcileScopeRef.current === threadScope;
		if (recoveryRearmed) {
			pendingThreadOverflowScopeRef.current = null;
			needsCanonicalThreadReconcileScopeRef.current = null;
			void queryClient.invalidateQueries({ queryKey, exact: true });
			return;
		}
		pendingThreadMutationsRef.current = [];
		pendingThreadMutationOverflowScopeRef.current = null;
		canonicalThreadRecoveryScopeRef.current = null;
		setMutationOverlayVersion((version) => version + 1);
	}, [
		query.data,
		query.isError,
		query.isFetching,
		queryClient,
		queryKey,
		threadScope,
	]);

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			if (
				!channelId ||
				!messageId ||
				!userId ||
				useAuthStore.getState().user?.id !== userId ||
				useMessengerStore.getState().isChannelRevoked(channelId) ||
				!isRecord(payload)
			) {
				return;
			}
			const data = payload;
			const ensureThreadScope = () => {
				if (!threadScope) return false;
				if (pendingThreadScopeRef.current !== threadScope) {
					pendingThreadMessagesRef.current.clear();
					pendingThreadMutationsRef.current = [];
					pendingThreadScopeRef.current = threadScope;
					pendingThreadOverflowScopeRef.current = null;
					pendingThreadMutationOverflowScopeRef.current = null;
					needsCanonicalThreadReconcileScopeRef.current = null;
					canonicalThreadRecoveryScopeRef.current = null;
					setMutationOverlayVersion((version) => version + 1);
				}
				return true;
			};
			const markNeedsCanonicalReconcile = () => {
				if (
					queryClient.isFetching({ queryKey, exact: true }) === 0 ||
					!ensureThreadScope()
				) {
					return false;
				}
				needsCanonicalThreadReconcileScopeRef.current = threadScope;
				return true;
			};
			const retainPendingThreadMutation = (mutation: PendingThreadMutation) => {
				if (!markNeedsCanonicalReconcile()) return;
				if (queuePendingThreadMutation(pendingThreadMutationsRef.current, mutation)) {
					pendingThreadMutationOverflowScopeRef.current = threadScope;
				}
				setMutationOverlayVersion((version) => version + 1);
			};

			switch (event) {
				case "message_sent": {
					const message = buildMessageFromWsEvent(data);
					if (
						!message ||
						!getMessageCacheTargets(message, channelId, messageId).thread
					) {
						return;
					}
					const isFetching =
						queryClient.isFetching({ queryKey, exact: true }) > 0;
					if (isMessageCache(queryClient.getQueryData(queryKey))) {
						queryClient.setQueryData(queryKey, (old: unknown) =>
							appendMessage(old, message),
						);
						if (!isFetching) return;
					}
					if (!ensureThreadScope()) return;
					if (queuePendingThreadMessage(pendingThreadMessagesRef.current, message)) {
						pendingThreadOverflowScopeRef.current = threadScope;
					}
					return;
				}
				case "message_edited": {
					if (!canHandleThreadDelta(data, channelId, messageId)) return;
					const messageID = messageEventID(data);
					if (!messageID) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						editMatchingMessage(old, messageID, data),
					);
					retainPendingThreadMutation({
						kind: "edit",
						messageID,
						data,
					});
					return;
				}
				case "message_deleted": {
					if (!canHandleThreadDelta(data, channelId, messageId)) return;
					const messageID = messageEventID(data);
					if (!messageID) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						removeMatchingMessage(old, messageID),
					);
					retainPendingThreadMutation({ kind: "delete", messageID });
					return;
				}
				case "reaction_toggled": {
					const delta = reactionDelta(data, channelId);
					if (!delta) return;
					queryClient.setQueryData(queryKey, (old: unknown) =>
						updateReactionInCache(old, delta),
					);
					retainPendingThreadMutation({ kind: "reaction", delta });
					return;
				}
			}
		},
		[channelId, messageId, queryClient, queryKey, threadScope, userId],
	);

	useWsEventListener(handleEvent);

	const publicThreadData = useMemo(() => {
		if (!userId || isRevoked) return undefined;
		if (pendingThreadScopeRef.current !== threadScope) return query.data;
		if (pendingThreadMutationOverflowScopeRef.current === threadScope) {
			return undefined;
		}
		return applyPendingThreadMutations(
			query.data,
			pendingThreadMutationsRef.current,
		);
	}, [isRevoked, mutationOverlayVersion, query.data, threadScope, userId]);

	return {
		...query,
		data: publicThreadData,
		messages: publicThreadData?.pages.flat() ?? [],
		loadMore: query.fetchNextPage,
		hasMore: query.hasNextPage,
	};
}