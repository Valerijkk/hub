import { createFileRoute, Outlet } from "@tanstack/react-router";

/**
 * Лейаут-маршрут для проекта.
 * Оборачивает дочерние маршруты (доска, задачи, настройки).
 */
function ProjectLayout() {
	return <Outlet />;
}

export const Route = createFileRoute("/_authenticated/projects/$projectId")({
	component: ProjectLayout,
});
