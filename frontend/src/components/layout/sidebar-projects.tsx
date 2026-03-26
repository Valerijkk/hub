/**
 * Секция проектов сайдбара -- Slack-style список с expand/collapse подменю.
 */

import { useQuery } from "@tanstack/react-query";
import { useMatchRoute } from "@tanstack/react-router";
import { useState } from "react";
import { listProjects } from "@/api/projects";

import { cn } from "@/lib/utils";
import { SidebarLink } from "./sidebar-link";

/** Цвета точек проектов. */
const PROJECT_DOT_COLORS = ["#3B82F6", "#8B5CF6", "#10B981", "#F59E0B"] as const;

/** Подменю модулей проекта (Доска, Задачи, TMS, Бэклог, Спринты). */
interface ProjectSubNavItem {
	to: string;
	label: string;
	icon: React.ReactNode;
}

/** Элементы подменю проекта. */
function getProjectSubNav(projectId: string): ProjectSubNavItem[] {
	return [
		{
			to: `/projects/${projectId}/board`,
			label: "Доска",
			icon: (
				<svg
					className="text-amber-400"
					width="16"
					height="16"
					fill="none"
					viewBox="0 0 16 16"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<rect x="1" y="2" width="4" height="12" rx="0.5" />
					<rect x="6" y="2" width="4" height="8" rx="0.5" />
					<rect x="11" y="2" width="4" height="5" rx="0.5" />
				</svg>
			),
		},
		{
			to: `/projects/${projectId}/issues`,
			label: "Задачи",
			icon: (
				<svg
					className="text-blue-400"
					width="16"
					height="16"
					fill="none"
					viewBox="0 0 16 16"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<circle cx="8" cy="8" r="6" />
					<path d="M5.5 8l2 2 3.5-4" />
				</svg>
			),
		},
		{
			to: `/projects/${projectId}/tms`,
			label: "TMS",
			icon: (
				<svg
					className="text-green-400"
					width="16"
					height="16"
					fill="none"
					viewBox="0 0 16 16"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M2 3h12M2 6.5h8M2 10h10M2 13.5h6" />
				</svg>
			),
		},
		{
			to: `/projects/${projectId}/backlog`,
			label: "Бэклог",
			icon: (
				<svg
					className="text-slate-400"
					width="16"
					height="16"
					fill="none"
					viewBox="0 0 16 16"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M3 4h10M3 8h10M3 12h10" />
				</svg>
			),
		},
		{
			to: `/projects/${projectId}/sprints`,
			label: "Спринты",
			icon: (
				<svg
					className="text-purple-400"
					width="16"
					height="16"
					fill="none"
					viewBox="0 0 16 16"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M3 13c2-4 4-8 10-10" />
					<path d="M10 3h3v3" />
				</svg>
			),
		},
	];
}

/** Секция проектов в сайдбаре с раскрывающимися подменю модулей. */
export function SidebarProjects({
	collapsed,
	t,
}: {
	collapsed: boolean;
	t: (key: string, options?: Record<string, string>) => string;
}) {
	const matchRoute = useMatchRoute();

	/** ID раскрытого проекта в боковом меню. */
	const [expandedProjectId, setExpandedProjectId] = useState<string | null>(
		null,
	);

	const { data: projects } = useQuery({
		queryKey: ["projects"],
		queryFn: listProjects,
	});

	/** Переключение раскрытия подменю проекта. */
	const toggleProject = (projectId: string) => {
		setExpandedProjectId((prev) => (prev === projectId ? null : projectId));
	};

	return (
		<div className="flex-1 overflow-y-auto px-3 py-2">
			{/* Заголовок секции */}
			<div className="mb-1.5 flex items-center justify-between px-2">
				{!collapsed && (
					<span className="text-[11px] font-semibold uppercase tracking-wider text-[#D1D2D3]/40">
						{t("nav.projects")}
					</span>
				)}
			</div>

			<div className="space-y-0.5">
				{projects?.map((project, index) => {
					const isProjectActive = matchRoute({
						to: "/projects/$projectId",
						fuzzy: true,
						params: { projectId: project.id },
					});
					const isExpanded =
						expandedProjectId === project.id || isProjectActive;

					return (
						<div key={project.id}>
							{/* Основная ссылка проекта с кнопкой раскрытия */}
							<div className="flex items-center">
								{/* Кнопка раскрытия подменю */}
								{!collapsed && (
									<button
										type="button"
										onClick={() =>
											toggleProject(project.id)
										}
										className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-[#D1D2D3]/40 hover:text-white transition-colors duration-150"
										aria-label={
											isExpanded
												? "Свернуть модули"
												: "Развернуть модули"
										}
									>
										<svg
											width="10"
											height="10"
											viewBox="0 0 10 10"
											fill="none"
											stroke="currentColor"
											strokeWidth="1.5"
											strokeLinecap="round"
											strokeLinejoin="round"
											className={cn(
												"transition-transform duration-150",
												isExpanded && "rotate-90",
											)}
										>
											<path d="M3.5 1.5l4 3.5-4 3.5" />
										</svg>
									</button>
								)}
								<div className="flex-1">
									<SidebarLink
										to={`/projects/${project.id}`}
										label={project.name}
										collapsed={collapsed}
										icon={
											<div
												className="w-[7px] h-[7px] rounded-full shrink-0"
												style={{ background: PROJECT_DOT_COLORS[index % PROJECT_DOT_COLORS.length] }}
											/>
										}
									/>
								</div>
							</div>

							{/* Подменю модулей проекта */}
							{isExpanded && !collapsed && (
								<div className="space-y-0.5 mt-0.5 mb-1">
									{getProjectSubNav(project.id).map(
										(item) => (
											<SidebarLink
												key={item.to}
												to={item.to}
												label={item.label}
												collapsed={collapsed}
												params={{
													projectId: project.id,
												}}
												indent
												icon={item.icon}
											/>
										),
									)}
								</div>
							)}
						</div>
					);
				})}
			</div>

			{projects?.length === 0 && !collapsed && (
				<p className="px-3 py-2 text-xs text-[#D1D2D3]/30">
					{t("sidebar.noProjects")}
				</p>
			)}
		</div>
	);
}
