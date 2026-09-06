import { useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { canChannel, forwardMessage } from "@/api/messenger";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

type ForwardCommandOptions = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
	canForward: boolean;
};

type ForwardScope = Omit<ForwardCommandOptions, "onOpenChange"> & {
	actorId: string | undefined;
};

type ForwardRequest = {
	readonly scope: ForwardScope;
	readonly targetChannelId: string;
	invalidated: boolean;
	unsubscribers: (() => void)[];
};

function hasAuthority(scope: ForwardScope, targetChannelId: string): boolean {
	const state = useMessengerStore.getState();
	const source = state.channels.find((item) => item.id === scope.channelId);
	const target = state.channels.find((item) => item.id === targetChannelId);
	return (
		scope.open &&
		scope.canForward &&
		Boolean(scope.actorId) &&
		scope.channelId.length > 0 &&
		scope.messageId.length > 0 &&
		targetChannelId.length > 0 &&
		scope.messageChannelId === scope.channelId &&
		useAuthStore.getState().user?.id === scope.actorId &&
		!state.isChannelRevoked(scope.channelId) &&
		!state.isChannelRevoked(targetChannelId) &&
		source !== undefined &&
		target !== undefined &&
		canChannel(source, "can_forward") &&
		canChannel(target, "can_post")
	);
}

function releaseSubscriptions(request: ForwardRequest) {
	for (const unsubscribe of request.unsubscribers.splice(0)) unsubscribe();
}

export function useMessageForwardCommand({
	open,
	onOpenChange,
	channelId,
	messageId,
	messageChannelId,
	canForward,
}: ForwardCommandOptions): {
	pending: boolean;
	submit: (targetChannelId: string) => Promise<boolean>;
} {
	const { t } = useTranslation("messenger");
	const actorId = useAuthStore((state) => state.user?.id);
	const [pending, setPending] = useState(false);
	const scopeRef = useRef<ForwardScope | null>(null);
	const requestRef = useRef<ForwardRequest | null>(null);
	const callbackRef = useRef(onOpenChange);

	useLayoutEffect(() => {
		callbackRef.current = onOpenChange;
	}, [onOpenChange]);

	useLayoutEffect(() => {
		const scope: ForwardScope = {
			actorId,
			open,
			channelId,
			messageId,
			messageChannelId,
			canForward,
		};
		scopeRef.current = scope;
		setPending(false);
		return () => {
			const request = requestRef.current;
			if (request?.scope === scope) {
				request.invalidated = true;
				releaseSubscriptions(request);
				requestRef.current = null;
			}
			if (scopeRef.current === scope) scopeRef.current = null;
		};
	}, [actorId, open, channelId, messageId, messageChannelId, canForward]);

	function ownsRequest(request: ForwardRequest): boolean {
		return scopeRef.current === request.scope && requestRef.current === request;
	}

	function isCurrent(request: ForwardRequest): boolean {
		return (
			ownsRequest(request) &&
			!request.invalidated &&
			hasAuthority(request.scope, request.targetChannelId)
		);
	}

	function release(request: ForwardRequest) {
		releaseSubscriptions(request);
		// Authority loss still needs pending cleanup; detached requests cannot unlock.
		if (ownsRequest(request)) {
			requestRef.current = null;
			setPending(false);
		}
	}

	async function submit(targetChannelId: string): Promise<boolean> {
		const scope = scopeRef.current;
		if (!scope || requestRef.current || !hasAuthority(scope, targetChannelId)) {
			return false;
		}
		const request: ForwardRequest = {
			scope,
			targetChannelId,
			invalidated: false,
			unsubscribers: [],
		};
		requestRef.current = request;
		const observeAuthority = () => {
			// Latch transient store losses even when React batches an A-B-A update.
			if (!hasAuthority(scope, request.targetChannelId)) {
				request.invalidated = true;
			}
		};
		request.unsubscribers.push(
			useMessengerStore.subscribe(observeAuthority),
			useAuthStore.subscribe(observeAuthority),
		);
		setPending(true);
		try {
			await forwardMessage(
				scope.channelId,
				scope.messageId,
				request.targetChannelId,
			);
			if (!isCurrent(request)) return false;
			// A close callback may synchronously submit again in this same scope.
			release(request);
			callbackRef.current(false);
			return true;
		} catch (error) {
			if (isCurrent(request)) {
				toastApiError(error, t("messageForward.error"));
			}
			return false;
		} finally {
			release(request);
		}
	}

	return { pending, submit };
}
