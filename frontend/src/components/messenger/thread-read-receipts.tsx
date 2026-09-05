import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { RefObject } from "react";
import type { Channel, Message, ThreadResponse } from "@/api/messenger";
import { markThreadRead } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { channelsQueryKey, useChannels } from "@/hooks/use-channels";
import {
	threadDetailQueryKey,
	threadMessagesQueryKey,
} from "@/hooks/use-messages";
import { unreadQueryKey } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

type Candidate = {
	id: string;
	createdAt: number;
	marker: HTMLElement;
	generation: number;
};

interface ThreadReadReceiptsProps {
	channelId: string;
	rootMessageId: string;
	userId: string;
	replies: Message[];
	containerRef: RefObject<HTMLDivElement | null>;
	detail: ThreadResponse;
	detailReady: boolean;
	repliesReady: boolean;
}

function compareCandidates(left: Candidate, right: Candidate): number {
	if (left.createdAt !== right.createdAt) {
		return left.createdAt > right.createdAt ? 1 : -1;
	}
	if (left.id === right.id) return 0;
	return left.id > right.id ? 1 : -1;
}

function isEligibleReply(
	reply: Message | undefined,
	channelId: string,
	rootMessageId: string,
): reply is Message {
	return Boolean(
		reply &&
			reply.id &&
			reply.channel_id === channelId &&
			reply.thread_id === rootMessageId &&
			!reply.deleted_at &&
			Number.isFinite(Date.parse(reply.created_at)),
	);
}

function isVisibleAndFocused(): boolean {
	return (
		typeof document !== "undefined" &&
		document.visibilityState === "visible" &&
		document.hasFocus()
	);
}

export function ThreadReadReceipts({
	channelId,
	rootMessageId,
	userId,
	replies,
	containerRef,
	detail,
	detailReady,
	repliesReady,
}: ThreadReadReceiptsProps) {
	const { t } = useTranslation("messenger");
	const queryClient = useQueryClient();
	const channelsQuery = useChannels();
	const currentUserId = useAuthStore((state) => state.user?.id ?? null);
	const isRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const [receiptError, setReceiptError] = useState(false);
	const [observationEpoch, setObservationEpoch] = useState(0);
	const scopeKey = JSON.stringify([channelId, rootMessageId, userId]);
	const lifecycleRef = useRef(0);
	const observerGenerationRef = useRef(0);
	const observerRef = useRef<IntersectionObserver | null>(null);
	const pendingRef = useRef<Candidate | null>(null);
	const inFlightRef = useRef<object | null>(null);
	const watermarkRef = useRef<Candidate | null>(null);
	const receiptErrorRef = useRef(false);
	const repliesRef = useRef(replies);
	const detailRef = useRef(detail);
	const detailReadyRef = useRef(detailReady);
	const repliesReadyRef = useRef(repliesReady);

	function disconnectAndClear(): void {
		pendingRef.current = null;
		observerGenerationRef.current += 1;
		observerRef.current?.disconnect();
		observerRef.current = null;
	}

	function isCurrentScope(lifecycle: number): boolean {
		return (
			lifecycleRef.current === lifecycle &&
			useAuthStore.getState().user?.id === userId &&
			!useMessengerStore.getState().revokedChannelIds.has(channelId)
		);
	}

	function currentCandidate(
		marker: HTMLElement,
		generation: number,
	): Candidate | null {
		const container = containerRef.current;
		if (
			generation !== observerGenerationRef.current ||
			!container ||
			!marker.isConnected ||
			!container.contains(marker) ||
			marker.dataset.threadReadMarker !== ""
		) {
			return null;
		}

		const id = marker.dataset.messageId;
		const reply = repliesRef.current.find((candidate) => candidate.id === id);
		if (!id || !isEligibleReply(reply, channelId, rootMessageId)) return null;

		return {
			id: reply.id,
			createdAt: Date.parse(reply.created_at),
			marker,
			generation,
		};
	}

	function canDispatch(
		candidate: Candidate,
		lifecycle: number,
	): Candidate | null {
		if (
			!isCurrentScope(lifecycle) ||
			!isVisibleAndFocused() ||
			!detailReadyRef.current ||
			!repliesReadyRef.current
		) {
			return null;
		}

		const current = currentCandidate(candidate.marker, candidate.generation);
		if (
			!current ||
			current.id !== candidate.id ||
			current.createdAt !== candidate.createdAt
		) {
			return null;
		}

		const channelsState = queryClient.getQueryState<Channel[]>(
			channelsQueryKey(userId),
		);
		const detailState = queryClient.getQueryState<ThreadResponse>(
			threadDetailQueryKey(channelId, rootMessageId, userId),
		);
		const repliesState = queryClient.getQueryState(
			threadMessagesQueryKey(channelId, rootMessageId, userId),
		);
		if (
			channelsState?.status !== "success" ||
			channelsState.fetchStatus !== "idle" ||
			channelsState.error != null ||
			!Array.isArray(channelsState.data) ||
			detailState?.status !== "success" ||
			detailState.fetchStatus !== "idle" ||
			detailState.error != null ||
			!detailState.data ||
			repliesState?.status !== "success" ||
			repliesState.fetchStatus !== "idle" ||
			repliesState.error != null ||
			!repliesState.data ||
			detailState.data !== detailRef.current
		) {
			return null;
		}

		const channel = channelsState.data.find(
			(candidateChannel) => candidateChannel.id === channelId,
		);
		const capabilities = channel?.capabilities;
		const parent = detailState.data.parent;
		if (
			!channel ||
			capabilities?.is_member !== true ||
			capabilities.can_read !== true ||
			capabilities.can_view_unread !== true ||
			(parent.id !== rootMessageId ||
				parent.channel_id !== channelId ||
				parent.deleted_at ||
				parent.thread_id) ||
			(detailState.data.follow_mode !== "auto" &&
				detailState.data.follow_mode !== "manual_follow")
		) {
			return null;
		}

		return current;
	}

	function pump(lifecycle: number): void {
		if (inFlightRef.current || receiptErrorRef.current) return;
		const pending = pendingRef.current;
		if (!pending) return;

		const candidate = canDispatch(pending, lifecycle);
		pendingRef.current = null;
		if (!candidate) return;
		if (
			watermarkRef.current &&
			compareCandidates(candidate, watermarkRef.current) <= 0
		) {
			return;
		}

		const request = {};
		inFlightRef.current = request;
		void (async () => {
			try {
				await markThreadRead(channelId, rootMessageId, candidate.id);
				if (!isCurrentScope(lifecycle)) return;

				watermarkRef.current = candidate;
				try {
					await queryClient.invalidateQueries({
						queryKey: unreadQueryKey(userId),
						exact: true,
					});
				} catch {
					// Transport завершён; восстановление unread остаётся авторитетным.
				}
			} catch {
				if (!isCurrentScope(lifecycle)) return;
				pendingRef.current = null;
				receiptErrorRef.current = true;
				setReceiptError(true);
			} finally {
				if (inFlightRef.current === request) {
					inFlightRef.current = null;
				}
				if (isCurrentScope(lifecycle) && !receiptErrorRef.current) {
					pump(lifecycle);
				}
			}
		})();
	}

	useLayoutEffect(() => {
		repliesRef.current = replies;
		detailRef.current = detail;
		detailReadyRef.current = detailReady;
		repliesReadyRef.current = repliesReady;
	}, [detail, detailReady, replies, repliesReady]);

	useLayoutEffect(() => {
		lifecycleRef.current += 1;
		pendingRef.current = null;
		inFlightRef.current = null;
		watermarkRef.current = null;
		receiptErrorRef.current = false;
		setReceiptError(false);
		setObservationEpoch((epoch) => epoch + 1);

		return () => {
			lifecycleRef.current += 1;
			disconnectAndClear();
			inFlightRef.current = null;
		};
	}, [scopeKey]);

	useEffect(() => {
		const reobserve = () => {
			disconnectAndClear();
			if (isVisibleAndFocused() && !receiptErrorRef.current) {
				setObservationEpoch((epoch) => epoch + 1);
			}
		};

		const onVisibilityChange = () => reobserve();
		window.addEventListener("blur", reobserve);
		window.addEventListener("focus", reobserve);
		document.addEventListener("visibilitychange", onVisibilityChange);
		return () => {
			window.removeEventListener("blur", reobserve);
			window.removeEventListener("focus", reobserve);
			document.removeEventListener("visibilitychange", onVisibilityChange);
		};
	}, [scopeKey]);

	const resourcesReady =
		currentUserId === userId &&
		!isRevoked &&
		detailReady &&
		repliesReady &&
		channelsQuery.isSuccess &&
		!channelsQuery.isError &&
		!channelsQuery.isFetching;
	const observationChannel = channelsQuery.data?.find(
		(candidateChannel) => candidateChannel.id === channelId,
	);
	const observationCapabilities = observationChannel?.capabilities;
	const observationParent = detail.parent;
	const observationEligible =
		resourcesReady &&
		observationCapabilities?.is_member === true &&
		observationCapabilities.can_read === true &&
		observationCapabilities.can_view_unread === true &&
		observationParent.id === rootMessageId &&
		observationParent.channel_id === channelId &&
		!observationParent.deleted_at &&
		!observationParent.thread_id &&
		(detail.follow_mode === "auto" || detail.follow_mode === "manual_follow");

	useEffect(() => {
		disconnectAndClear();
		const container = containerRef.current;
		if (
			!observationEligible ||
			receiptErrorRef.current ||
			!container ||
			!isVisibleAndFocused() ||
			typeof IntersectionObserver === "undefined"
		) {
			return;
		}

		const lifecycle = lifecycleRef.current;
		const generation = observerGenerationRef.current + 1;
		observerGenerationRef.current = generation;
		const observer = new IntersectionObserver(
			(entries) => {
				if (
					generation !== observerGenerationRef.current ||
					!isCurrentScope(lifecycle) ||
					receiptErrorRef.current ||
					!isVisibleAndFocused()
				) {
					return;
				}

				let greatest: Candidate | null = null;
				for (const entry of entries) {
					if (!entry.isIntersecting || entry.intersectionRatio !== 1) continue;
					if (!(entry.target instanceof HTMLElement)) continue;
					const candidate = currentCandidate(entry.target, generation);
					if (
						candidate &&
						(!greatest || compareCandidates(candidate, greatest) > 0)
					) {
						greatest = candidate;
					}
				}

				if (!greatest) return;
				if (
					watermarkRef.current &&
					compareCandidates(greatest, watermarkRef.current) <= 0
				) {
					return;
				}
				if (
					!pendingRef.current ||
					compareCandidates(greatest, pendingRef.current) > 0
				) {
					pendingRef.current = greatest;
				}
				pump(lifecycle);
			},
			{ root: null, rootMargin: "0px", threshold: 1 },
		);
		observerRef.current = observer;
		container
			.querySelectorAll<HTMLElement>(
				'[data-thread-read-marker=""][data-message-id]',
			)
			.forEach((marker) => observer.observe(marker));

		return () => {
			observer.disconnect();
			if (observerRef.current === observer) observerRef.current = null;
			if (observerGenerationRef.current === generation) {
				observerGenerationRef.current += 1;
			}
			pendingRef.current = null;
		};
	}, [
		containerRef,
		observationEpoch,
		receiptError,
		replies,
		observationEligible,
		scopeKey,
	]);

	function handleRetry(): void {
		if (!isCurrentScope(lifecycleRef.current)) return;
		pendingRef.current = null;
		receiptErrorRef.current = false;
		setReceiptError(false);
		setObservationEpoch((epoch) => epoch + 1);
	}

	if (!receiptError) return null;

	return (
		<div className="flex items-center gap-2 px-4 py-2 text-sm text-destructive" role="alert">
			<span>{t("threadRead.error")}</span>
			<Button type="button" variant="outline" size="sm" onClick={handleRetry}>
				{t("threadRead.retry")}
			</Button>
		</div>
	);
}
