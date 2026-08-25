import { queryClient } from "@/lib/query-client";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

/**
 * Removes all client state that belongs to the current authenticated identity.
 *
 * Keep this synchronous: a subsequent login in the same tab must never render
 * the previous identity's cache, messenger state, or WebSocket listeners.
 */
export function resetAuthenticatedClientState(): void {
	useWsStore.getState().resetSession();
	queryClient.clear();
	useMessengerStore.getState().reset();
}