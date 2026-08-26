import { useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import { messagesQueryKey, threadMessagesQueryKey } from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";

/** Хуки централизованной инвалидации кэша React Query после мутаций. */

export function useInvalidateMessages(channelId: string) {
	const qc = useQueryClient();
	const userId = useAuthStore((state) => state.user?.id ?? null);
	return useCallback((threadId?: string) => {
		if (!userId || useAuthStore.getState().user?.id !== userId) return;
		void qc.invalidateQueries({
			queryKey: messagesQueryKey(channelId, userId),
			exact: true,
		});
		if (threadId) {
			void qc.invalidateQueries({
				queryKey: threadMessagesQueryKey(channelId, threadId, userId),
				exact: true,
			});
		}
	}, [qc, channelId, userId]);
}

export function useInvalidateChannels() {
	const qc = useQueryClient();
	return useCallback(() => {
		void qc.invalidateQueries({ queryKey: ["messenger", "channels"] });
	}, [qc]);
}

export function useInvalidateProjects() {
	const qc = useQueryClient();
	return useCallback(() => {
		void qc.invalidateQueries({ queryKey: ["projects"] });
	}, [qc]);
}
