import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react";
import { type Message, searchMessages } from "@/api/messenger";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

type SearchScope = "channel" | "global";
type Admission = Readonly<{
	query: string;
	scope: SearchScope;
	effectiveChannelId?: string;
}>;
type Attempt = Readonly<{
	id: string;
	userId: string;
	sessionId: string;
	query: string;
	scope: SearchScope;
	contextChannelId: string | null;
	effectiveChannelId?: string;
	key: readonly unknown[];
}>;
type SessionOptions = { userId: string | null; channelId?: string };
const ROOT_KEY = ["messenger", "search-dialog"] as const;
const EMPTY_RESULTS: readonly Message[] = [];

/** Поиск открытой сессии: данные в Query, действия привязаны к текущей попытке. */
export function useSearchDialogSession({ userId, channelId }: SessionOptions) {
	const queryClient = useQueryClient();
	const [sessionId] = useState(() => crypto.randomUUID());
	const [draft, setDraftState] = useState("");
	const [scope, setScopeState] = useState<SearchScope>("channel");
	const [active, setActive] = useState<Attempt | null>(null);
	const draftRef = useRef(draft);
	const scopeRef = useRef(scope);
	const activeRef = useRef<Attempt | null>(null);
	const inFlightRef = useRef<Attempt | null>(null);
	const mountedRef = useRef(false);
	const authUserId = useAuthStore((state) =>
		state.isAuthenticated ? (state.user?.id ?? null) : null,
	);
	const revokedChannelIds = useMessengerStore((state) => state.revokedChannelIds);
	const idleKey = useMemo(
		() => [...ROOT_KEY, userId ?? "anonymous", sessionId, "idle"] as const,
		[sessionId, userId],
	);

	const hasLiveContext = useCallback((attempt: Attempt) => {
		const auth = useAuthStore.getState();
		if (!auth.isAuthenticated || auth.user?.id !== attempt.userId) return false;
		if (attempt.userId !== userId || attempt.contextChannelId !== (channelId ?? null)) return false;
		return !attempt.contextChannelId ||
			!useMessengerStore.getState().isChannelRevoked(attempt.contextChannelId);
	}, [channelId, userId]);

	const isCurrentAttempt = useCallback((attempt: Attempt) =>
		mountedRef.current &&
		activeRef.current === attempt &&
		draftRef.current.trim() === attempt.query &&
		scopeRef.current === attempt.scope &&
		attempt.effectiveChannelId ===
			(scopeRef.current === "channel" && channelId ? channelId : undefined) &&
		hasLiveContext(attempt),
	[channelId, hasLiveContext]);

	const removeAttempt = useCallback((attempt: Attempt) => {
		// Снимаем только старую попытку; соседняя новая Query не затрагивается.
		void queryClient.cancelQueries({ queryKey: attempt.key, exact: true }).catch(() => undefined);
		queryClient.removeQueries({ queryKey: attempt.key, exact: true });
	}, [queryClient]);

	const discardActive = useCallback(() => {
		const attempt = activeRef.current;
		if (!attempt) return;
		activeRef.current = null;
		if (inFlightRef.current === attempt) inFlightRef.current = null;
		removeAttempt(attempt);
		setActive((current) => (current === attempt ? null : current));
	}, [removeAttempt]);

	const admit = useCallback((query: string, nextScope: SearchScope): Admission | null => {
		const contextChannelId = channelId ?? null;
		const auth = useAuthStore.getState();
		if (
			!mountedRef.current || !userId || !query || !auth.isAuthenticated ||
			auth.user?.id !== userId ||
			(contextChannelId && useMessengerStore.getState().isChannelRevoked(contextChannelId))
		) return null;
		return {
			query,
			scope: nextScope,
			effectiveChannelId: nextScope === "channel" && channelId ? channelId : undefined,
		};
	}, [channelId, userId]);

	useLayoutEffect(() => {
		mountedRef.current = true;
		return () => {
			// Это локальная отмена Query, а не прерывание уже отправленного HTTP.
			mountedRef.current = false;
			activeRef.current = null;
			inFlightRef.current = null;
			const prefix = [...ROOT_KEY, userId ?? "anonymous", sessionId] as const;
			void queryClient.cancelQueries({ queryKey: prefix }).catch(() => undefined);
			queryClient.removeQueries({ queryKey: prefix });
		};
	}, [queryClient, sessionId, userId]);

	useLayoutEffect(() => {
		const attempt = activeRef.current;
		if (attempt && !isCurrentAttempt(attempt)) discardActive();
	}, [authUserId, channelId, discardActive, isCurrentAttempt, revokedChannelIds, userId]);

	const begin = useCallback((admission: Admission) => {
		if (!userId) return;
		const id = crypto.randomUUID();
		const attempt: Attempt = {
			id, userId, sessionId,
			query: admission.query,
			scope: admission.scope,
			contextChannelId: channelId ?? null,
			effectiveChannelId: admission.effectiveChannelId,
			key: [...ROOT_KEY, userId, sessionId, admission.query, admission.scope, admission.effectiveChannelId, id],
		};
		activeRef.current = attempt;
		inFlightRef.current = attempt;
		setActive(attempt);
	}, [channelId, sessionId, userId]);

	const submit = useCallback(() => {
		if (inFlightRef.current) return;
		const admission = admit(draftRef.current.trim(), scopeRef.current);
		if (!admission) return;
		discardActive();
		begin(admission);
	}, [admit, begin, discardActive]);

	const queryState = useQuery<Message[]>({
		queryKey: active?.key ?? idleKey,
		queryFn: () => {
			const attempt = active;
			if (!attempt || !isCurrentAttempt(attempt)) {
				return Promise.reject(new Error("Obsolete search attempt"));
			}
			return searchMessages(attempt.query, attempt.effectiveChannelId).finally(() => {
				if (activeRef.current === attempt && inFlightRef.current === attempt) {
					inFlightRef.current = null;
				}
			});
		},
		enabled: active !== null && isCurrentAttempt(active),
		retry: false,
		refetchOnMount: false,
		refetchOnWindowFocus: false,
		refetchOnReconnect: false,
		refetchInterval: false,
		gcTime: 0,
	});

	const retry = useCallback(() => {
		const attempt = active;
		if (!attempt || inFlightRef.current || !queryState.isError || !isCurrentAttempt(attempt)) return;
		const admission = admit(attempt.query, attempt.scope);
		if (!admission) return;
		discardActive();
		begin(admission);
	}, [admit, active, begin, discardActive, isCurrentAttempt, queryState.isError]);

	const setDraft = useCallback((nextDraft: string) => {
		if (nextDraft === draftRef.current) return;
		draftRef.current = nextDraft;
		setDraftState(nextDraft);
		discardActive();
	}, [discardActive]);
	const setScope = useCallback((nextScope: SearchScope) => {
		if (nextScope === scopeRef.current) return;
		scopeRef.current = nextScope;
		setScopeState(nextScope);
		discardActive();
	}, [discardActive]);

	const current = active !== null && isCurrentAttempt(active);
	const sourceResults = current && queryState.isSuccess
		? (queryState.data ?? EMPTY_RESULTS) : EMPTY_RESULTS;
	const results = sourceResults.filter((message) => !revokedChannelIds.has(message.channel_id));
	const currentSelectionGuard = active && current && queryState.isSuccess
		? (message: Message) => isCurrentAttempt(active) &&
			!useMessengerStore.getState().isChannelRevoked(message.channel_id)
		: null;

	return {
		draft, scope, setDraft, setScope, submit, retry,
		pending: current && queryState.fetchStatus === "fetching",
		successEmpty: current && queryState.isSuccess && sourceResults.length === 0,
		error: current && queryState.isError,
		results, currentSelectionGuard,
	};
}
