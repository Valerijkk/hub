import { createFileRoute } from "@tanstack/react-router";

/**
 * Стаб маршрута /projects/$projectId/tms/cases_/$caseKey. Реальный компонент
 * (~842 LOC) — в `cases_.$caseKey.lazy.tsx`. Code-split TanStack Router
 * исключает страницу детального просмотра тест-кейса из main-бандла.
 */
export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/tms/cases_/$caseKey",
)({});
