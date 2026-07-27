import { useQuery } from "@tanstack/react-query";
import { useCallback, useEffect, useRef } from "react";
import {
	getPresence,
	type PresenceStatus,
	sendHeartbeat,
} from "@/api/messenger";

/** Интервал поллинга присутствия (30 секунд). */
const PRESENCE_POLL_INTERVAL = 30_000;

/** Интервал heartbeat (25 секунд, чуть меньше поллинга). */
const HEARTBEAT_INTERVAL = 25_000;

/**
 * Стабильный ключ presence-query без мутации входного массива.
 * A10-004: Array.prototype.sort() мутирует in-place → ломает props/caller state.
 */
export function presenceQueryKeyPart(userIds: string[]): string {
	return [...userIds].sort().join(",");
}

/**
 * usePresence загружает статусы присутствия для списка пользователей
 * с поллингом каждые 30 секунд и отправляет heartbeat каждые 25 секунд.
 */
export function usePresence(userIds: string[]) {
	const stableIds = useRef(userIds);
	stableIds.current = userIds;

	const query = useQuery({
		queryKey: ["messenger", "presence", presenceQueryKeyPart(userIds)],
		queryFn: () => getPresence(stableIds.current),
		enabled: userIds.length > 0,
		refetchInterval: PRESENCE_POLL_INTERVAL,
		staleTime: PRESENCE_POLL_INTERVAL - 5_000,
	});

	// Heartbeat каждые 25 секунд.
	useEffect(() => {
		sendHeartbeat().catch(() => {});
		const interval = setInterval(() => {
			sendHeartbeat().catch(() => {});
		}, HEARTBEAT_INTERVAL);
		return () => clearInterval(interval);
	}, []);

	const getStatus = useCallback(
		(userId: string): PresenceStatus => {
			return query.data?.[userId] ?? "offline";
		},
		[query.data],
	);

	return {
		statuses: query.data ?? ({} as Record<string, PresenceStatus>),
		getStatus,
		isLoading: query.isLoading,
	};
}
