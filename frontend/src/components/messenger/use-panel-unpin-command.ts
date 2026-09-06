import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { togglePin, type PinnedMessage } from "@/api/messenger";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { pinsQueryKey } from "./pinned-messages";

type PanelUnpinOptions = {
	open: boolean;
	channelId: string;
	canListPins: boolean;
	canPin: boolean;
};

type PanelUnpinScope = PanelUnpinOptions & {
	actorId: string | undefined;
};

type PanelUnpinRequest = {
	scope: PanelUnpinScope;
	actorId: string;
	pinId: string;
	messageId: string;
	channelId: string;
	invalidated: boolean;
	releaseListeners: () => void;
};

function hasLiveAuthority(scope: PanelUnpinScope): boolean {
	const messenger = useMessengerStore.getState();
	const channel = messenger.channels.find((item) => item.id === scope.channelId);
	const capabilities = channel?.capabilities;
	return (
		Boolean(scope.actorId) &&
		scope.open &&
		scope.channelId.length > 0 &&
		scope.canListPins &&
		scope.canPin &&
		useAuthStore.getState().user?.id === scope.actorId &&
		!messenger.isChannelRevoked(scope.channelId) &&
		capabilities?.can_list_pins === true &&
		capabilities?.can_pin === true
	);
}

export function usePanelUnpinCommand({
	open,
	channelId,
	canListPins,
	canPin,
}: PanelUnpinOptions): {
	pending: boolean;
	submit: (pin: PinnedMessage) => Promise<void>;
} {
	const { t } = useTranslation("messenger");
	const queryClient = useQueryClient();
	const actorId = useAuthStore((state) => state.user?.id);
	const [pending, setPending] = useState(false);
	const scopeRef = useRef<PanelUnpinScope | null>(null);
	const requestRef = useRef<PanelUnpinRequest | null>(null);

	useLayoutEffect(() => {
		const scope: PanelUnpinScope = {
			actorId,
			open,
			channelId,
			canListPins,
			canPin,
		};
		scopeRef.current = scope;
		requestRef.current = null;
		setPending(false);

		return () => {
			if (scopeRef.current !== scope) return;
			scopeRef.current = null;
			const request = requestRef.current;
			if (request?.scope === scope) {
				request.invalidated = true;
				request.releaseListeners();
				requestRef.current = null;
			}
		};
	}, [actorId, canListPins, canPin, channelId, open]);

	const owns = useCallback((request: PanelUnpinRequest) => {
		return requestRef.current === request && scopeRef.current === request.scope;
	}, []);

	const isCurrent = useCallback((request: PanelUnpinRequest) => {
		return owns(request) && !request.invalidated && hasLiveAuthority(request.scope);
	}, [owns]);

	const submit = useCallback(async (pin: PinnedMessage): Promise<void> => {
		const scope = scopeRef.current;
		if (!scope || requestRef.current) return;

		const capturedActorId = scope.actorId;
		const capturedChannelId = scope.channelId;
		const capturedPinId = pin.id;
		const capturedMessageId = pin.message_id;
		const capturedPinChannelId = pin.channel_id;
		if (
			!capturedActorId ||
			!capturedChannelId ||
			!capturedPinId ||
			!capturedMessageId ||
			!capturedPinChannelId ||
			capturedPinChannelId !== capturedChannelId ||
			!hasLiveAuthority(scope)
		) return;

		const request: PanelUnpinRequest = {
			scope,
			actorId: capturedActorId,
			pinId: capturedPinId,
			messageId: capturedMessageId,
			channelId: capturedChannelId,
			invalidated: false,
			releaseListeners: () => undefined,
		};
		requestRef.current = request;
		setPending(true);

		const unsubscribeAuth = useAuthStore.subscribe((state) => {
			if (state.user?.id !== request.actorId) request.invalidated = true;
		});
		const unsubscribeMessenger = useMessengerStore.subscribe((state) => {
			const channel = state.channels.find((item) => item.id === request.channelId);
			const capabilities = channel?.capabilities;
			if (
				state.isChannelRevoked(request.channelId) ||
				capabilities?.can_list_pins !== true ||
				capabilities?.can_pin !== true
			) request.invalidated = true;
		});
		let released = false;
		request.releaseListeners = () => {
			if (released) return;
			released = true;
			unsubscribeAuth();
			unsubscribeMessenger();
		};

		try {
			await togglePin(request.channelId, request.messageId);
			if (isCurrent(request)) {
				await queryClient.invalidateQueries({
					queryKey: pinsQueryKey(request.channelId, request.actorId),
					exact: true,
				});
			}
		} catch (error) {
			if (isCurrent(request)) toastApiError(error, t("messagePin.error"));
		} finally {
			request.releaseListeners();
			if (owns(request)) {
				requestRef.current = null;
				setPending(false);
			}
		}
	}, [isCurrent, owns, queryClient, t]);

	return { pending, submit };
}
