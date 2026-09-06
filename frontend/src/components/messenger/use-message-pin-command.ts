import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { togglePin } from "@/api/messenger";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { pinsQueryKey } from "./pinned-messages";

type PinCommandOptions = {
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
	canPin: boolean;
};

type PinScope = PinCommandOptions & {
	actorId: string | undefined;
	liveCanPin: boolean;
	channelRevoked: boolean;
};

interface PinRequest {
	scope: PinScope;
}

export function useMessagePinCommand({
	channelId,
	messageId,
	messageChannelId,
	canPin,
}: PinCommandOptions): {
	pending: boolean;
	submit: () => Promise<void>;
} {
	const { t } = useTranslation("messenger");
	const queryClient = useQueryClient();
	const actorId = useAuthStore((state) => state.user?.id);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const liveCanPin = useMessengerStore(
		(state) =>
			state.channels.find((item) => item.id === channelId)?.capabilities
				?.can_pin === true,
	);
	const [pending, setPending] = useState(false);
	const scopeRef = useRef<PinScope | null>(null);
	const requestRef = useRef<PinRequest | null>(null);

	useLayoutEffect(() => {
		const scope: PinScope = {
			actorId,
			channelId,
			messageId,
			messageChannelId,
			canPin,
			liveCanPin,
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
		canPin,
		channelId,
		channelRevoked,
		liveCanPin,
		messageChannelId,
		messageId,
	]);

	const isCurrent = useCallback((request: PinRequest) => {
		const scope = request.scope;
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
			scope.canPin &&
			scope.liveCanPin &&
			!scope.channelRevoked &&
			useAuthStore.getState().user?.id === scope.actorId &&
			!messenger.isChannelRevoked(scope.channelId) &&
			channel?.capabilities?.can_pin === true
		);
	}, []);

	const submit = useCallback(async (): Promise<void> => {
		const scope = scopeRef.current;
		if (!scope || requestRef.current) return;

		const request: PinRequest = { scope };
		requestRef.current = request;
		if (!isCurrent(request)) {
			if (requestRef.current === request) requestRef.current = null;
			return;
		}

		const capturedUserId = scope.actorId;
		setPending(true);
		try {
			await togglePin(scope.channelId, scope.messageId);
			if (isCurrent(request) && capturedUserId) {
				await queryClient.invalidateQueries({
					queryKey: pinsQueryKey(scope.channelId, capturedUserId),
					exact: true,
				});
			}
		} catch (error) {
			if (isCurrent(request)) {
				toastApiError(error, t("messagePin.error"));
			}
		} finally {
			if (isCurrent(request)) {
				requestRef.current = null;
				setPending(false);
			}
		}
	}, [isCurrent, queryClient, t]);

	return { pending, submit };
}
