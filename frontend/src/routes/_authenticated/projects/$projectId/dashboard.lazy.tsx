/**
 * Страница проектного дашборда.
 * Burndown, velocity, тренды тестов, активность команды.
 */
import { createLazyFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import type { LayoutItem } from "@/api/dashboard";
import {
	projectDev,
	projectManager,
	projectQA,
} from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { useDashboardStore } from "@/stores/dashboard-store";

export const Route = createLazyFileRoute(
	"/_authenticated/projects/$projectId/dashboard",
)({
	component: ProjectDashboardWithBoundary,
});

/** Обёртка для изоляции рантайм-ошибок дашборда проекта (FE-H4). */
function ProjectDashboardWithBoundary() {
	return (
		<RouteErrorBoundary module="Дашборд проекта">
			<ProjectDashboard />
		</RouteErrorBoundary>
	);
}

/** Шаблоны раскладки проектного дашборда. */
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
			const current = useDashboardStore.getState().layouts.project;
			if (!current || current.length === 0) {
				setLayout("project", projectDev);
			}
		});
	}, [projectId, fetchProject, loadLayout, setLayout]);

	const items = layouts.project ?? TEMPLATES[template] ?? projectDev;

	return (
		<div className="animate-[eu_450ms_var(--spring)_both]">
			<div className="flex items-center justify-between mb-4">
				<h1 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1">
					Дашборд проекта
				</h1>
				<div className="flex gap-1">
					{Object.keys(TEMPLATES).map((key) => (
						<button
							key={key}
							type="button"
							onClick={() => {
								setTemplate(key);
								setLayout("project", TEMPLATES[key] ?? projectDev);
							}}
							className={`px-3 py-1 text-[13px] font-semibold rounded-[var(--r2)] border transition-colors duration-200 ${
								template === key
									? "bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white border-blue/30"
									: "border-b1 text-t2 hover:bg-[var(--hover)] hover:text-t1"
							}`}
						>
							{key === "dev" ? "Dev" : key === "qa" ? "QA" : "Manager"}
						</button>
					))}
				</div>
			</div>

			{loading && !projectData ? (
				<p className="text-t2 text-[13px]">Загрузка...</p>
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
