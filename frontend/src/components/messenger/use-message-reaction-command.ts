import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { toggleReaction } from "@/api/messenger";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

interface ReactionScope {
	actorId: string | undefined;
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
	canReact: boolean;
	liveCanReact: boolean;
	channelRevoked: boolean;
}

interface UseMessageReactionCommandOptions {
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
	canReact: boolean;
}

export function useMessageReactionCommand({
	channelId,
	messageId,
	messageChannelId,
	canReact,
}: UseMessageReactionCommandOptions) {
	const { t } = useTranslation("messenger");
	const actorId = useAuthStore((state) => state.user?.id);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const liveCanReact = useMessengerStore(
		(state) =>
			state.channels.find((item) => item.id === channelId)?.capabilities
				?.can_react === true,
	);
	const [pending, setPending] = useState(false);
	const scopeRef = useRef<ReactionScope | null>(null);
	const requestRef = useRef<object | null>(null);

	useLayoutEffect(() => {
		const scope: ReactionScope = {
			actorId,
			channelId,
			messageId,
			messageChannelId,
			canReact,
			liveCanReact,
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
		canReact,
		channelId,
		channelRevoked,
		liveCanReact,
		messageChannelId,
		messageId,
	]);

	const isCurrentScope = useCallback(
		(request: object, scope: ReactionScope) => {
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
				scope.canReact &&
				scope.liveCanReact &&
				!scope.channelRevoked &&
				useAuthStore.getState().user?.id === scope.actorId &&
				!messenger.isChannelRevoked(scope.channelId) &&
				channel?.capabilities?.can_react === true
			);
		},
		[],
	);

	const submit = useCallback(
		async (emoji: string): Promise<boolean> => {
			const scope = scopeRef.current;
			if (!scope || !scope.channelId || requestRef.current) return false;

			const request = {};
			requestRef.current = request;
			if (!isCurrentScope(request, scope)) {
				if (requestRef.current === request) requestRef.current = null;
				return false;
			}

			setPending(true);
			try {
				await toggleReaction(scope.channelId, scope.messageId, emoji);
				return isCurrentScope(request, scope);
			} catch (error) {
				if (isCurrentScope(request, scope)) {
					toastApiError(error, t("messageReaction.error"));
				}
				return false;
			} finally {
				if (isCurrentScope(request, scope)) {
					requestRef.current = null;
					setPending(false);
				}
			}
		},
		[isCurrentScope, t],
	);

	return { pending, submit };
}
