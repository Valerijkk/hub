/**
 * Страница проектного дашборда.
 * Burndown, velocity, тренды тестов, активность команды.
 * A11-012: list/apply шаблонов через API `/dashboard/templates`.
 */
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createLazyFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import {
	applyDashboardTemplate,
	type LayoutItem,
	listDashboardTemplates,
} from "@/api/dashboard";
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

/** Локальный fallback, если API шаблонов недоступен. */
const LOCAL_TEMPLATES: Record<string, LayoutItem[]> = {
	dev: projectDev,
	qa: projectQA,
	manager: projectManager,
};

const TEMPLATE_LABELS: Record<string, string> = {
	dev: "Dev",
	qa: "QA",
	manager: "Manager",
};

function ProjectDashboard() {
	const { projectId } = Route.useParams();
	const queryClient = useQueryClient();
	const fetchProject = useDashboardStore((s) => s.fetchProject);
	const loadLayout = useDashboardStore((s) => s.loadLayout);
	const projectData = useDashboardStore((s) => s.projectData);
	const layouts = useDashboardStore((s) => s.layouts);
	const setLayout = useDashboardStore((s) => s.setLayout);
	const loading = useDashboardStore((s) => s.loading);
	const error = useDashboardStore((s) => s.error);
	const [template, setTemplate] = useState("dev");

	const { data: remoteTemplates } = useQuery({
		queryKey: ["dashboard", "templates"],
		queryFn: listDashboardTemplates,
		staleTime: 60_000,
	});

	const templateKeys =
		remoteTemplates && Object.keys(remoteTemplates).length > 0
			? Object.keys(remoteTemplates)
			: Object.keys(LOCAL_TEMPLATES);

	const applyMutation = useMutation({
		mutationFn: (key: string) => applyDashboardTemplate(projectId, key),
		onSuccess: async (_data, key) => {
			setTemplate(key);
			const fromApi = remoteTemplates?.[key];
			const layout = fromApi ?? LOCAL_TEMPLATES[key] ?? projectDev;
			setLayout("project", layout);
			await loadLayout("project", projectId);
			void queryClient.invalidateQueries({
				queryKey: ["dashboard", "layout", projectId],
			});
			toast.success("Шаблон применён");
		},
		onError: () => toast.error("Не удалось применить шаблон"),
	});

	useEffect(() => {
		fetchProject(projectId);
		loadLayout("project", projectId).then(() => {
			const current = useDashboardStore.getState().layouts.project;
			if (!current || current.length === 0) {
				setLayout("project", projectDev);
			}
		});
	}, [projectId, fetchProject, loadLayout, setLayout]);

	const items =
		layouts.project ??
		remoteTemplates?.[template] ??
		LOCAL_TEMPLATES[template] ??
		projectDev;

	return (
		<div className="animate-[eu_450ms_var(--spring)_both]">
			<div className="flex items-center justify-between mb-4">
				<h1 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1">
					Дашборд проекта
				</h1>
				<div className="flex gap-1">
					{templateKeys.map((key) => (
						<button
							key={key}
							type="button"
							disabled={applyMutation.isPending}
							onClick={() => applyMutation.mutate(key)}
							className={`px-3 py-1 text-[13px] font-semibold rounded-[var(--r2)] border transition-colors duration-200 ${
								template === key
									? "bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white border-blue/30"
									: "border-b1 text-t2 hover:bg-[var(--hover)] hover:text-t1"
							}`}
						>
							{TEMPLATE_LABELS[key] ?? key}
						</button>
					))}
				</div>
			</div>

			{loading && !projectData ? (
				<p className="text-t2 text-[13px]">Загрузка...</p>
			) : error && !projectData ? (
				<div
					role="alert"
					className="flex flex-col items-center justify-center py-16 gap-3 text-t2"
				>
					<p className="text-sm font-medium">{error}</p>
					<button
						type="button"
						className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
						onClick={() => void fetchProject(projectId)}
					>
						Попробовать снова
					</button>
				</div>
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
