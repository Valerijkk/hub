import { useQueryClient } from "@tanstack/react-query";
import { useCallback, useEffect, useRef } from "react";
import { CHANNELS_QUERY_KEY, useChannels } from "@/hooks/use-channels";
import { useWsEventListener } from "@/hooks/use-websocket";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

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
	const { data: channels } = useChannels();
	const { subscribe, unsubscribe, connected } = useWsStore();
	const socket = useWsStore((state) => state.socket);

	useEffect(() => {
		if (channels) {
			setChannels(channels);
		}
	}, [channels, setChannels]);

	useEffect(() => {
		if (!userId) {
			connectedIdentityRef.current = null;
			return;
		}
		if (!connected || !socket) return;

		const room = `user:${userId}`;
		const previousIdentity = connectedIdentityRef.current;
		const shouldInvalidate =
			previousIdentity !== null &&
			(previousIdentity.userId !== userId ||
				previousIdentity.socket !== socket);
		subscribe(room);
		connectedIdentityRef.current = { userId, socket };

		if (shouldInvalidate) {
			queryClient.invalidateQueries({
				queryKey: CHANNELS_QUERY_KEY,
				exact: true,
			});
		}

		return () => {
			unsubscribe(room);
		};
	}, [connected, userId, socket, subscribe, unsubscribe, queryClient]);

	const handleEvent = useCallback(
		(event: string, _payload: unknown) => {
			if (
				event === "channel_created" ||
				event === "channel_updated" ||
				event === "channel_removed" ||
				event === "new_message_notification"
			) {
				queryClient.invalidateQueries({
					queryKey: CHANNELS_QUERY_KEY,
					exact: true,
				});
			}
		},
		[queryClient],
	);

	useWsEventListener(handleEvent);
}
