import { useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";

/** Hooks for centralized React Query cache invalidation. */

export function useInvalidateMessages(channelId: string) {
	const qc = useQueryClient();
	return useCallback(() => {
		void qc.invalidateQueries({
			queryKey: ["messenger", "messages", channelId],
		});
	}, [qc, channelId]);
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
