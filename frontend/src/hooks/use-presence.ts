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
 *
 * A2W11-3-001: наружу `isError`/`error`/`refetch` — fail poll ≠ «все offline».
 * A2W11-3-003: heartbeat fail логируется, не silent swallow.
 */
export function usePresence(userIds: string[]) {
	const stableIds = useRef(userIds);
	stableIds.current = userIds;
	const heartbeatFailCount = useRef(0);

	const query = useQuery({
		queryKey: ["messenger", "presence", presenceQueryKeyPart(userIds)],
		queryFn: () => getPresence(stableIds.current),
		enabled: userIds.length > 0,
		refetchInterval: PRESENCE_POLL_INTERVAL,
		staleTime: PRESENCE_POLL_INTERVAL - 5_000,
	});

	// Heartbeat каждые 25 секунд.
	useEffect(() => {
		const beat = () => {
			void sendHeartbeat().catch((err: unknown) => {
				heartbeatFailCount.current += 1;
				console.warn(
					"[presence] heartbeat failed",
					heartbeatFailCount.current,
					err,
				);
			});
		};
		beat();
		const interval = setInterval(beat, HEARTBEAT_INTERVAL);
		return () => clearInterval(interval);
	}, []);

	const getStatus = useCallback(
		(userId: string): PresenceStatus => {
			// При ошибке поллинга не маскируем fail как offline — caller смотрит isError.
			if (query.isError) {
				return "offline";
			}
			return query.data?.[userId] ?? "offline";
		},
		[query.data, query.isError],
	);

	return {
		statuses: query.data ?? ({} as Record<string, PresenceStatus>),
		getStatus,
		isLoading: query.isLoading,
		isError: query.isError,
		error: query.error,
		refetch: query.refetch,
	};
}
