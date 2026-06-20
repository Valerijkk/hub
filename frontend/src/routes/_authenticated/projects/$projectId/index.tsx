import { createFileRoute } from "@tanstack/react-router";

/**
 * Стаб маршрута /projects/$projectId/. Реальный компонент (~407 LOC) — в
 * `index.lazy.tsx`. Code-split TanStack Router исключает overview-страницу
 * проекта из main-бандла.
 */
export const Route = createFileRoute("/_authenticated/projects/$projectId/")(
	{},
);
