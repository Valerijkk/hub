import {
	createFileRoute,
	Link,
	Outlet,
	redirect,
	useLocation,
} from "@tanstack/react-router";
import { cn } from "@/lib/utils";

/**
 * Лейаут-маршрут для модуля TMS (управление тестированием).
 * Отображает табы навигации и рендерит дочерние маршруты.
 * Дизайн уровня Allure TestOps — чистый, профессиональный.
 */
function TmsLayout() {
	const { projectId } = Route.useParams();
	const location = useLocation();

	/** Определяет активный таб по текущему пути. */
	const getActiveTab = () => {
		if (location.pathname.includes("/tms/runs")) return "runs";
		if (location.pathname.includes("/tms/traceability")) return "traceability";
		if (location.pathname.includes("/tms/environments")) return "environments";
		return "cases";
	};

	const activeTab = getActiveTab();

	/** Конфигурация табов навигации. */
	const tabs = [
		{
			id: "cases",
			label: "Тест-кейсы",
			to: "/projects/$projectId/tms/cases" as const,
			icon: (
				<svg
					className="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
					/>
				</svg>
			),
		},
		{
			id: "runs",
			label: "Прогоны",
			to: "/projects/$projectId/tms/runs" as const,
			icon: (
				<svg
					className="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M13 10V3L4 14h7v7l9-11h-7z"
					/>
				</svg>
			),
		},
		{
			id: "traceability",
			label: "Трассировка",
			to: "/projects/$projectId/tms/traceability" as const,
			icon: (
				<svg
					className="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"
					/>
				</svg>
			),
		},
		{
			id: "environments",
			label: "Окружения",
			to: "/projects/$projectId/tms/environments" as const,
			icon: (
				<svg
					className="h-4 w-4"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2z"
					/>
				</svg>
			),
		},
	];

	return (
		<div className="flex h-full flex-col">
			{/* Заголовок модуля */}
			<div className="mb-6">
				<div className="flex items-center justify-between mb-4">
					<div className="flex items-center gap-3">
						<div className="flex h-9 w-9 items-center justify-center rounded-lg gradient-tms text-white shadow-sm">
							<svg
								className="h-5 w-5"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={2}
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
								/>
							</svg>
						</div>
						<div>
							<h1 className="text-2xl font-bold tracking-tight">
								Управление тестированием
							</h1>
							<p className="text-xs text-muted-foreground mt-0.5">
								Тест-кейсы, прогоны и отчёты
							</p>
						</div>
					</div>
				</div>

				{/* Навигация табов */}
				<div className="flex gap-1 border-b">
					{tabs.map((tab) => (
						<Link
							key={tab.id}
							to={tab.to}
							params={{ projectId }}
							className={cn(
								"relative border-b-2 px-4 py-2.5 text-sm font-medium transition-all duration-200",
								activeTab === tab.id
									? "border-primary text-primary"
									: "border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground/30",
							)}
						>
							<span className="flex items-center gap-1.5">
								{tab.icon}
								{tab.label}
							</span>
						</Link>
					))}
				</div>
			</div>

			{/* Дочерний маршрут */}
			<div className="flex-1 min-h-0">
				<Outlet />
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/projects/$projectId/tms")(
	{
		component: TmsLayout,
		beforeLoad: ({ location, params }) => {
			if (location.pathname.endsWith("/tms") || location.pathname.endsWith("/tms/")) {
				throw redirect({ to: "/projects/$projectId/tms/cases", params: { projectId: params.projectId } });
			}
		},
	},
);
