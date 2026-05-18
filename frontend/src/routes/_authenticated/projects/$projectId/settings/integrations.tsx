import { createFileRoute } from "@tanstack/react-router";

/**
 * Стаб маршрута настроек интеграций проекта. Реальный компонент (~528 LOC) лежит
 * в `integrations.lazy.tsx` — code-split TanStack Router исключает страницу из
 * main-бандла.
 */
export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/settings/integrations",
)({});
