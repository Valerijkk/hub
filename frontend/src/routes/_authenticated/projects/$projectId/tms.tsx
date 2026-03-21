import { createFileRoute, Link, Outlet, useLocation } from "@tanstack/react-router";
import { cn } from "@/lib/utils";

/**
 * Лейаут-маршрут для модуля TMS (управление тестированием).
 * Отображает табы (Тест-кейсы / Прогоны) и рендерит дочерние маршруты.
 */
function TmsLayout() {
  const { projectId } = Route.useParams();
  const location = useLocation();
  const isRunsTab = location.pathname.includes("/tms/runs");

  return (
    <div className="flex h-full flex-col">
      {/* Заголовок и табы */}
      <div className="mb-6">
        <h1 className="mb-4 text-2xl font-bold">Управление тестированием</h1>
        <div className="flex gap-1 border-b">
          <Link
            to="/projects/$projectId/tms/cases"
            params={{ projectId }}
            className={cn(
              "border-b-2 px-4 py-2 text-sm font-medium transition-colors",
              !isRunsTab
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            Тест-кейсы
          </Link>
          <Link
            to="/projects/$projectId/tms/runs"
            params={{ projectId }}
            className={cn(
              "border-b-2 px-4 py-2 text-sm font-medium transition-colors",
              isRunsTab
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            Прогоны
          </Link>
        </div>
      </div>

      {/* Дочерний маршрут */}
      <Outlet />
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms",
)({
  component: TmsLayout,
});
