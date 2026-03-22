/**
 * Персональный дашборд пользователя.
 * Отображает виджеты с задачами, непрочитанными сообщениями, звонками, дедлайнами и трекингом времени.
 */
import { createLazyFileRoute } from "@tanstack/react-router";
import { useEffect, useMemo } from "react";
import { personalDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { useAuthStore } from "@/stores/auth-store";
import { useDashboardStore } from "@/stores/dashboard-store";

export const Route = createLazyFileRoute("/_authenticated/dashboard")({
	component: PersonalDashboard,
});

/** Возвращает приветствие на русском в зависимости от текущего часа. */
function getGreeting(): string {
	const hour = new Date().getHours();
	if (hour >= 5 && hour < 12) return "Доброе утро";
	if (hour >= 12 && hour < 17) return "Добрый день";
	if (hour >= 17 && hour < 22) return "Добрый вечер";
	return "Доброй ночи";
}

/** Форматирует сегодняшнюю дату в русской локали. */
function formatTodayDate(): string {
	return new Date().toLocaleDateString("ru-RU", {
		weekday: "long",
		day: "numeric",
		month: "long",
	});
}

function PersonalDashboard() {
	const fetchPersonal = useDashboardStore((s) => s.fetchPersonal);
	const loadLayout = useDashboardStore((s) => s.loadLayout);
	const personalData = useDashboardStore((s) => s.personalData);
	const personalItems = useDashboardStore((s) => s.layouts["personal"]);
	const setLayout = useDashboardStore((s) => s.setLayout);
	const loading = useDashboardStore((s) => s.loading);
	const user = useAuthStore((s) => s.user);

	const greeting = useMemo(() => getGreeting(), []);
	const todayDate = useMemo(() => formatTodayDate(), []);
	const displayName =
		user?.display_name || user?.name || user?.email?.split("@")[0] || "";

	useEffect(() => {
		fetchPersonal();
		void loadLayout("personal").then(() => {
			const current = useDashboardStore.getState().layouts.personal;
			if (!current || current.length === 0) {
				setLayout("personal", personalDefault);
			}
		});
	}, [fetchPersonal, loadLayout, setLayout]);

	const items = personalItems ?? personalDefault;

	return (
		<div className="animate-fade-in">
			{/* Greeting header with gradient accent */}
			<div className="mb-6">
				<div className="flex items-center gap-3 mb-1">
					<div className="flex h-9 w-9 items-center justify-center rounded-lg gradient-dashboard text-white shadow-sm">
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
								d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
							/>
						</svg>
					</div>
					<div>
						<h1 className="text-2xl font-bold tracking-tight greeting-gradient">
							{greeting}
							{displayName ? `, ${displayName}` : ""}
						</h1>
						<p className="text-sm text-muted-foreground capitalize">
							{todayDate}
						</p>
					</div>
				</div>
			</div>

			{loading && !personalData ? (
				/* Professional loading skeletons */
				<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{[1, 2, 3, 4, 5, 6].map((i) => (
						<div
							key={i}
							className="rounded-xl border bg-card p-5 space-y-3"
							style={{ animationDelay: `${i * 80}ms` }}
						>
							<div className="flex items-center gap-3">
								<div className="h-8 w-8 rounded-lg animate-shimmer" />
								<div className="h-4 w-24 rounded animate-shimmer" />
							</div>
							<div className="h-10 w-full rounded animate-shimmer" />
							<div className="h-3 w-3/4 rounded animate-shimmer" />
						</div>
					))}
				</div>
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
