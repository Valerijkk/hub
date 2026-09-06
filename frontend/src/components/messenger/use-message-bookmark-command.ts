import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { addBookmark } from "@/api/messenger";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { bookmarksQueryKey } from "./bookmarks-panel";

type BookmarkCommandOptions = {
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
	canBookmark: boolean;
};

type BookmarkScope = BookmarkCommandOptions & {
	actorId: string | undefined;
	liveCanBookmark: boolean;
	channelRevoked: boolean;
};

export function useMessageBookmarkCommand({
	channelId,
	messageId,
	messageChannelId,
	canBookmark,
}: BookmarkCommandOptions): {
	pending: boolean;
	submit: () => Promise<void>;
} {
	const { t } = useTranslation("messenger");
	const queryClient = useQueryClient();
	const actorId = useAuthStore((state) => state.user?.id);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const liveCanBookmark = useMessengerStore(
		(state) =>
			state.channels.find((item) => item.id === channelId)?.capabilities
				?.can_bookmark === true,
	);
	const [pending, setPending] = useState(false);
	const scopeRef = useRef<BookmarkScope | null>(null);
	const requestRef = useRef<object | null>(null);

	useLayoutEffect(() => {
		const scope: BookmarkScope = {
			actorId,
			channelId,
			messageId,
			messageChannelId,
			canBookmark,
			liveCanBookmark,
			channelRevoked,
		};
		scopeRef.current = scope;
		requestRef.current = null;
		setPending(false);

		return () => {
			if (scopeRef.current === scope) {
				scopeRef.current = null;
				requestRef.current = null;
			}
		};
	}, [
		actorId,
		canBookmark,
		channelId,
		channelRevoked,
		liveCanBookmark,
		messageChannelId,
		messageId,
	]);

	const isCurrent = useCallback((request: object, scope: BookmarkScope) => {
		const messenger = useMessengerStore.getState();
		const channel = messenger.channels.find(
			(item) => item.id === scope.channelId,
		);
		return (
			requestRef.current === request &&
			scopeRef.current === scope &&
			Boolean(scope.actorId) &&
			scope.channelId.length > 0 &&
			scope.messageId.length > 0 &&
			scope.messageChannelId === scope.channelId &&
			scope.canBookmark &&
			scope.liveCanBookmark &&
			!scope.channelRevoked &&
			useAuthStore.getState().user?.id === scope.actorId &&
			!messenger.isChannelRevoked(scope.channelId) &&
			channel?.capabilities?.can_bookmark === true
		);
	}, []);

	const submit = useCallback(async (): Promise<void> => {
		const scope = scopeRef.current;
		if (!scope || requestRef.current) return;

		const request = { scope };
		requestRef.current = request;
		if (!isCurrent(request, scope)) {
			if (requestRef.current === request) requestRef.current = null;
			return;
		}

		const capturedUserId = scope.actorId;
		setPending(true);
		try {
			await addBookmark(scope.messageId, scope.channelId);
			if (isCurrent(request, scope) && capturedUserId) {
				await queryClient.invalidateQueries({
					queryKey: bookmarksQueryKey(capturedUserId),
					exact: true,
				});
			}
		} catch (error) {
			if (isCurrent(request, scope)) {
				toastApiError(error, t("messageBookmark.error"));
			}
		} finally {
			if (isCurrent(request, scope)) {
				requestRef.current = null;
				setPending(false);
			}
		}
	}, [isCurrent, queryClient, t]);

	return { pending, submit };
}
