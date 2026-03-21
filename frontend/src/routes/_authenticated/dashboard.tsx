/**
 * Страница персонального дашборда.
 * Отображает виджеты с задачами, непрочитанными, звонками, дедлайнами и тайм-трекингом.
 */
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";
import { useDashboardStore } from "@/stores/dashboard-store";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { personalDefault } from "@/components/dashboard/dashboard-templates";

export const Route = createFileRoute("/_authenticated/dashboard")({
  component: PersonalDashboard,
});

function PersonalDashboard() {
  const fetchPersonal = useDashboardStore((s) => s.fetchPersonal);
  const loadLayout = useDashboardStore((s) => s.loadLayout);
  const personalData = useDashboardStore((s) => s.personalData);
  const layouts = useDashboardStore((s) => s.layouts);
  const setLayout = useDashboardStore((s) => s.setLayout);
  const loading = useDashboardStore((s) => s.loading);

  useEffect(() => {
    fetchPersonal();
    loadLayout("personal").then(() => {
      // Если раскладка пустая -- применяем шаблон по умолчанию
      const current = useDashboardStore.getState().layouts["personal"];
      if (!current || current.length === 0) {
        setLayout("personal", personalDefault);
      }
    });
  }, [fetchPersonal, loadLayout, setLayout]);

  const items = layouts["personal"] ?? personalDefault;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Мой дашборд</h1>
      {loading && !personalData ? (
        <p className="text-muted-foreground">Загрузка...</p>
      ) : (
        <WidgetGrid
          items={items}
          dashboardType="personal"
          data={personalData}
        />
      )}
    </div>
  );
}
