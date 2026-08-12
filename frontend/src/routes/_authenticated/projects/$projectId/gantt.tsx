import { createFileRoute } from "@tanstack/react-router";

/**
 * A13-017: stub — Gantt chart in `gantt.lazy.tsx` (heavy deps out of main graph).
 */
export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/gantt",
)({});
