import { Link, createFileRoute, Outlet, useParams, useMatchRoute } from "@tanstack/react-router";
import { cn } from "@/lib/utils";

const PROJECT_TABS: { to: string; label: string; exact?: boolean }[] = [
	{ to: "/projects/$projectId", label: "Обзор", exact: true },
	{ to: "/projects/$projectId/board", label: "Доска" },
	{ to: "/projects/$projectId/issues", label: "Задачи" },
	{ to: "/projects/$projectId/backlog", label: "Бэклог" },
	{ to: "/projects/$projectId/sprints", label: "Спринты" },
	{ to: "/projects/$projectId/tms/cases", label: "TMS" },
	{ to: "/projects/$projectId/dashboard", label: "Дашборд" },
	{ to: "/projects/$projectId/settings/workflow", label: "Настройки" },
];

function ProjectLayout() {
	const { projectId } = useParams({ from: "/_authenticated/projects/$projectId" });
	const matchRoute = useMatchRoute();

	return (
		<div className="flex flex-col h-full">
			{/* Горизонтальная навигация проекта */}
			<div className="shrink-0 border-b border-white/[0.06]">
				<nav className="flex gap-0.5 px-4 overflow-x-auto scrollbar-thin">
					{PROJECT_TABS.map((tab) => {
						const isActive = tab.exact
							? matchRoute({ to: tab.to, params: { projectId } })
							: matchRoute({ to: tab.to, params: { projectId }, fuzzy: true });

						return (
							<Link
								key={tab.to}
								to={tab.to}
								params={{ projectId }}
								className={cn(
									"shrink-0 px-3 py-2.5 text-[13px] font-medium border-b-2 transition-colors",
									isActive
										? "border-primary text-white"
										: "border-transparent text-t2 hover:text-white",
								)}
							>
								{tab.label}
							</Link>
						);
					})}
				</nav>
			</div>

			{/* Контент */}
			<div className="flex-1 overflow-y-auto">
				<Outlet />
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/projects/$projectId")({
	component: ProjectLayout,
});
