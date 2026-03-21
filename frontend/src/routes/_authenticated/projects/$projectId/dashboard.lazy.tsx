/**
 * Project dashboard page.
 * Displays task charts, burndown, velocity, test trends and team activity.
 */
import { createLazyFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useDashboardStore } from "@/stores/dashboard-store";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import type { LayoutItem } from "@/api/dashboard";
import {
  projectDev,
  projectQA,
  projectManager,
} from "@/components/dashboard/dashboard-templates";

export const Route = createLazyFileRoute(
  "/_authenticated/projects/$projectId/dashboard",
)({
  component: ProjectDashboard,
});

/** Project dashboard templates. */
const TEMPLATES: Record<string, LayoutItem[]> = {
  dev: projectDev,
  qa: projectQA,
  manager: projectManager,
};

function ProjectDashboard() {
  const { projectId } = Route.useParams();
  const fetchProject = useDashboardStore((s) => s.fetchProject);
  const loadLayout = useDashboardStore((s) => s.loadLayout);
  const projectData = useDashboardStore((s) => s.projectData);
  const layouts = useDashboardStore((s) => s.layouts);
  const setLayout = useDashboardStore((s) => s.setLayout);
  const loading = useDashboardStore((s) => s.loading);
  const [template, setTemplate] = useState("dev");

  useEffect(() => {
    fetchProject(projectId);
    loadLayout("project", projectId).then(() => {
      const current = useDashboardStore.getState().layouts["project"];
      if (!current || current.length === 0) {
        setLayout("project", projectDev);
      }
    });
  }, [projectId, fetchProject, loadLayout, setLayout]);

  const items = layouts["project"] ?? TEMPLATES[template] ?? projectDev;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Дашборд проекта</h1>
        <div className="flex gap-1">
          {Object.keys(TEMPLATES).map((key) => (
            <button
              key={key}
              onClick={() => {
                setTemplate(key);
                setLayout("project", TEMPLATES[key] ?? projectDev);
              }}
              className={`px-3 py-1 text-sm rounded-md border ${
                template === key
                  ? "bg-primary text-primary-foreground"
                  : "hover:bg-muted"
              }`}
            >
              {key === "dev" ? "Dev" : key === "qa" ? "QA" : "Manager"}
            </button>
          ))}
        </div>
      </div>

      {loading && !projectData ? (
        <p className="text-muted-foreground">Загрузка...</p>
      ) : (
        <WidgetGrid
          items={items}
          dashboardType="project"
          projectId={projectId}
          data={projectData}
        />
      )}
    </div>
  );
}
