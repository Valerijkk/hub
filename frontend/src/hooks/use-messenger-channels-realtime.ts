import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef } from "react";
import { channelsQueryKey, useChannels } from "@/hooks/use-channels";
import { unreadQueryKey } from "@/hooks/use-unread";
import { useWsEventListener } from "@/hooks/use-websocket";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

function isRecord(value: unknown): value is Record<string, unknown> {
	return !!value && typeof value === "object" && !Array.isArray(value);
}

export function isStaleMessengerIdentityQuery(
	queryKey: readonly unknown[],
	userId: string,
): boolean {
	if (queryKey[0] !== "messenger") return false;
	switch (queryKey[1]) {
		case "channels":
			return queryKey[2] === "discover"
				? queryKey[3] !== userId
				: queryKey[2] !== userId;
		case "unread":
		case "bookmarks":
			return queryKey[2] !== userId;
		case "messages":
		case "pins":
			return queryKey[3] !== userId;
		case "thread":
		case "thread-messages":
			return queryKey[4] !== userId;
		default:
			return true;
	}
}

export function isRevokedChannelQueryKey(
	queryKey: readonly unknown[],
	channelId: string,
): boolean {
	if (queryKey[0] === "thread-issue-links") return queryKey[2] === channelId;
	if (queryKey[0] !== "messenger") return false;
	switch (queryKey[1]) {
		case "messages":
		case "thread":
		case "thread-messages":
		case "channel":
		case "channel-members":
		case "channel-invites":
		case "pins":
			return queryKey[2] === channelId;
		default:
			return false;
	}
}

/**
 * The one owner of messenger channel realtime synchronization.
 * The messenger layout mounts this hook once; useChannels observers only read Query.
 */
export function useMessengerChannelsRealtime() {
	const queryClient = useQueryClient();
	const userId = useAuthStore((state) => state.user?.id);
	const setChannels = useMessengerStore((state) => state.setChannels);
	const connectedIdentityRef = useRef<{
		userId: string;
		socket: WebSocket;
	} | null>(null);
	const requestedIdentityRef = useRef<{
		userId: string;
		socket: WebSocket;
	} | null>(null);
	const requiresIdentityResetRef = useRef(false);
	const { data: channels } = useChannels();
	const { subscribe, unsubscribe, connected } = useWsStore();
	const socket = useWsStore((state) => state.socket);

	useEffect(() => {
		if (
			channels &&
			userId &&
			useAuthStore.getState().user?.id === userId
		) {
			setChannels(channels);
		}
	}, [channels, setChannels, userId]);

	useEffect(() => {
		let cancelled = false;
		let subscribedRoom: string | null = null;

		if (!userId) {
			requiresIdentityResetRef.current = true;
			connectedIdentityRef.current = null;
			useMessengerStore.getState().reset();
			void (async () => {
				await queryClient.cancelQueries({ queryKey: ["messenger"] });
				if (cancelled) return;
				queryClient.removeQueries({ queryKey: ["messenger"] });
			})();
			return () => {
				cancelled = true;
			};
		}
		if (!connected || !socket) return;

		const room = `user:${userId}`;
		const nextIdentity = { userId, socket };
		const previousIdentity = requestedIdentityRef.current;
		const identityChanged =
			requiresIdentityResetRef.current ||
			(previousIdentity !== null && previousIdentity.userId !== userId);
		const socketChanged =
			previousIdentity !== null && previousIdentity.socket !== socket;
		requestedIdentityRef.current = nextIdentity;

		const activateIdentity = async () => {
			if (identityChanged) {
				requiresIdentityResetRef.current = true;
				connectedIdentityRef.current = null;
				useMessengerStore.getState().reset();
				await queryClient.cancelQueries({
					predicate: (query) =>
						isStaleMessengerIdentityQuery(query.queryKey, userId),
				});
				if (
					cancelled ||
					useAuthStore.getState().user?.id !== userId ||
					!useWsStore.getState().connected ||
					useWsStore.getState().socket !== socket
				) {
					return;
				}
				queryClient.removeQueries({
					predicate: (query) =>
						isStaleMessengerIdentityQuery(query.queryKey, userId),
				});
				if (
					cancelled ||
					useAuthStore.getState().user?.id !== userId ||
					!useWsStore.getState().connected ||
					useWsStore.getState().socket !== socket
				) {
					return;
				}
				requiresIdentityResetRef.current = false;
			}

			subscribe(room);
			subscribedRoom = room;
			connectedIdentityRef.current = nextIdentity;

			if (socketChanged && !identityChanged) {
				void queryClient.invalidateQueries(
					{
						queryKey: channelsQueryKey(userId),
						exact: true,
					},
					{ cancelRefetch: false },
				);
				void queryClient.invalidateQueries(
					{
						queryKey: unreadQueryKey(userId),
						exact: true,
					},
					{ cancelRefetch: false },
				);
			}
		};

		void activateIdentity();

		return () => {
			cancelled = true;
			if (subscribedRoom) {
				unsubscribe(subscribedRoom);
			}
			const activeIdentity = connectedIdentityRef.current;
			if (
				activeIdentity?.userId === userId &&
				activeIdentity.socket === socket
			) {
				connectedIdentityRef.current = null;
			}
		};
	}, [connected, userId, socket, subscribe, unsubscribe, queryClient]);

	const handleEvent = useCallback(
		(event: string, payload: unknown) => {
			const activeIdentity = connectedIdentityRef.current;
			if (
				!userId ||
				useAuthStore.getState().user?.id !== userId ||
				activeIdentity?.userId !== userId ||
				activeIdentity.socket !== useWsStore.getState().socket
			) {
				return;
			}

			if (event === "channel_removed" && isRecord(payload)) {
				const channelId = payload.channel_id;
				if (typeof channelId !== "string" || channelId.length === 0) return;

				useMessengerStore.getState().revokeChannel(channelId);
				void queryClient.cancelQueries({
					predicate: (query) =>
						isRevokedChannelQueryKey(query.queryKey, channelId),
				});
				queryClient.removeQueries({
					predicate: (query) =>
						isRevokedChannelQueryKey(query.queryKey, channelId),
				});
				queryClient.removeQueries({
					queryKey: ["messenger", "bookmarks"],
					exact: true,
				});
				void queryClient.invalidateQueries({
					queryKey: channelsQueryKey(userId),
					exact: true,
				});
				void queryClient.invalidateQueries({
					queryKey: unreadQueryKey(userId),
					exact: true,
				});
				return;
			}

			if (
				event === "channel_created" ||
				event === "channel_updated" ||
				event === "new_message_notification"
			) {
				if (event === "channel_created") {
					if (!isRecord(payload)) return;
					const channelId = payload.channel_id;
					if (typeof channelId !== "string" || channelId.length === 0) return;
					useMessengerStore.getState().restoreChannelAccess(channelId);
				}
				void queryClient.invalidateQueries({
					queryKey: channelsQueryKey(userId),
					exact: true,
				});
				if (event === "new_message_notification") {
					void queryClient.invalidateQueries({
						queryKey: unreadQueryKey(userId),
						exact: true,
					});
				}
			}
		},
		[queryClient, userId],
	);

	useWsEventListener(handleEvent);
}
