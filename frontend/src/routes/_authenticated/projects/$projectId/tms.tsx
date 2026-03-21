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
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg gradient-tms text-white">
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
              </svg>
            </div>
            <h1 className="text-2xl font-bold">Управление тестированием</h1>
          </div>
        </div>
        <div className="flex gap-1 border-b">
          <Link
            to="/projects/$projectId/tms/cases"
            params={{ projectId }}
            className={cn(
              "border-b-2 px-4 py-2.5 text-sm font-medium transition-colors",
              !isRunsTab
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            <span className="flex items-center gap-1.5">
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              Тест-кейсы
            </span>
          </Link>
          <Link
            to="/projects/$projectId/tms/runs"
            params={{ projectId }}
            className={cn(
              "border-b-2 px-4 py-2.5 text-sm font-medium transition-colors",
              isRunsTab
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            <span className="flex items-center gap-1.5">
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              Прогоны
            </span>
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
