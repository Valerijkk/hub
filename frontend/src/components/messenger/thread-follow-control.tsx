import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { Channel, ThreadResponse } from "@/api/messenger";
import { followThread, unfollowThread } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { channelsQueryKey, useChannels } from "@/hooks/use-channels";
import { threadDetailQueryKey } from "@/hooks/use-messages";
import { unreadQueryKey } from "@/hooks/use-unread";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

type ThreadFollowAction = "follow" | "unfollow";

type CurrentResources = {
	channel: Channel;
	thread: ThreadResponse;
};

interface ThreadFollowControlProps {
	channelId: string;
	rootMessageId: string;
	userId: string;
	thread: ThreadResponse;
	detailIsSuccess: boolean;
	detailIsError: boolean;
	detailIsFetching: boolean;
}

function followAction(
	followMode: string | undefined,
	channel: Channel,
): ThreadFollowAction | null {
	if (
		(followMode === undefined || followMode === "manual_unfollow") &&
		channel.capabilities?.can_follow === true
	) {
		return "follow";
	}
	if (
		(followMode === "auto" || followMode === "manual_follow") &&
		channel.capabilities?.can_unfollow === true
	) {
		return "unfollow";
	}
	return null;
}

export function ThreadFollowControl({
	channelId,
	rootMessageId,
	userId,
	thread,
	detailIsSuccess,
	detailIsError,
	detailIsFetching,
}: ThreadFollowControlProps) {
	const { t } = useTranslation("messenger");
	const queryClient = useQueryClient();
	const currentUserId = useAuthStore((state) => state.user?.id ?? null);
	const isRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const channelsQuery = useChannels();
	const [isPending, setIsPending] = useState(false);
	const lifecycleRef = useRef(0);
	const activeLifecycleRef = useRef<number | null>(null);
	const scopeKey = JSON.stringify([channelId, rootMessageId, userId]);
	const detailQueryKey = threadDetailQueryKey(channelId, rootMessageId, userId);

	useEffect(() => {
		lifecycleRef.current += 1;
		setIsPending(false);
		return () => {
			lifecycleRef.current += 1;
		};
	}, [scopeKey]);

	function isCurrentScope(lifecycle: number): boolean {
		return (
			lifecycleRef.current === lifecycle &&
			useAuthStore.getState().user?.id === userId &&
			!useMessengerStore.getState().revokedChannelIds.has(channelId)
		);
	}

	function currentResources(lifecycle: number): CurrentResources | null {
		if (!isCurrentScope(lifecycle)) return null;

		const channelsState = queryClient.getQueryState<Channel[]>(
			channelsQueryKey(userId),
		);
		const detailState = queryClient.getQueryState<ThreadResponse>(
			detailQueryKey,
		);
		if (
			channelsState?.status !== "success" ||
			channelsState.fetchStatus !== "idle" ||
			!Array.isArray(channelsState.data) ||
			detailState?.status !== "success" ||
			detailState.fetchStatus !== "idle" ||
			!detailState.data
		) {
			return null;
		}

		const channel = channelsState.data.find(
			(candidate) => candidate.id === channelId,
		);
		return channel ? { channel, thread: detailState.data } : null;
	}

	function actionNow(lifecycle: number): ThreadFollowAction | null {
		const resources = currentResources(lifecycle);
		return resources
			? followAction(resources.thread.follow_mode, resources.channel)
			: null;
	}

	const channel =
		channelsQuery.isSuccess && !channelsQuery.isError
			? channelsQuery.data?.find((candidate) => candidate.id === channelId)
			: undefined;
	const action =
		currentUserId === userId &&
		!isRevoked &&
		detailIsSuccess &&
		!detailIsError &&
		channel
			? followAction(thread.follow_mode, channel)
			: null;
	const isBusy = isPending || detailIsFetching || channelsQuery.isFetching;

	function handleClick(): void {
		const lifecycle = lifecycleRef.current;
		if (activeLifecycleRef.current === lifecycle) return;

		const currentAction = actionNow(lifecycle);
		if (!currentAction) return;

		activeLifecycleRef.current = lifecycle;
		setIsPending(true);

		void (async () => {
			try {
				if (currentAction === "follow") {
					await followThread(channelId, rootMessageId);
				} else {
					await unfollowThread(channelId, rootMessageId);
				}

				if (!isCurrentScope(lifecycle)) return;

				await queryClient.invalidateQueries({
					queryKey: detailQueryKey,
					exact: true,
				});
				if (!isCurrentScope(lifecycle)) return;

				await queryClient.invalidateQueries({
					queryKey: unreadQueryKey(userId),
					exact: true,
				});
			} catch (error) {
				if (isCurrentScope(lifecycle)) {
					toastApiError(error, t("threadFollow.error"));
				}
			} finally {
				if (activeLifecycleRef.current === lifecycle) {
					activeLifecycleRef.current = null;
				}
				if (isCurrentScope(lifecycle)) {
					setIsPending(false);
				}
			}
		})();
	}

	if (!action) return null;

	return (
		<>
			<Button
				type="button"
				variant="outline"
				size="sm"
				className="h-7 px-2 text-xs"
				disabled={isBusy}
				aria-busy={isBusy}
				onClick={handleClick}
			>
				{action === "follow"
					? t("threadFollow.follow")
					: t("threadFollow.unfollow")}
			</Button>
			<span className="sr-only" role="status" aria-live="polite">
				{isBusy ? t("threadFollow.updating") : ""}
			</span>
		</>
	);
}
