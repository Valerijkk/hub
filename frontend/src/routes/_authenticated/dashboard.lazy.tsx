/**
 * Personal user dashboard.
 * Displays widgets with tasks, unread messages, calls, deadlines and time tracking.
 */
import {
	Bell,
	CalendarPlus,
	MessageSquarePlus,
	Plus,
	Search,
} from "lucide-react";
import { createLazyFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useMemo } from "react";
import { personalDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { useAuthStore } from "@/stores/auth-store";
import { useDashboardStore } from "@/stores/dashboard-store";

export const Route = createLazyFileRoute("/_authenticated/dashboard")({
	component: PersonalDashboard,
});

/** Returns greeting in Russian based on current hour. */
function getGreeting(): string {
	const hour = new Date().getHours();
	if (hour >= 5 && hour < 12) return "Доброе утро";
	if (hour >= 12 && hour < 17) return "Добрый день";
	if (hour >= 17 && hour < 22) return "Добрый вечер";
	return "Доброй ночи";
}

/** Formats today's date in Russian locale. */
function formatTodayDate(): string {
	return new Date().toLocaleDateString("ru-RU", {
		weekday: "long",
		day: "numeric",
		month: "long",
	});
}

/** Formats current time. */
function formatTime(): string {
	return new Date().toLocaleTimeString("ru-RU", {
		hour: "2-digit",
		minute: "2-digit",
	});
}

function PersonalDashboard() {
	const navigate = useNavigate();
	const fetchPersonal = useDashboardStore((s) => s.fetchPersonal);
	const loadLayout = useDashboardStore((s) => s.loadLayout);
	const personalData = useDashboardStore((s) => s.personalData);
	const personalItems = useDashboardStore((s) => s.layouts["personal"]);
	const setLayout = useDashboardStore((s) => s.setLayout);
	const loading = useDashboardStore((s) => s.loading);
	const user = useAuthStore((s) => s.user);

	const greeting = useMemo(() => getGreeting(), []);
	const todayDate = useMemo(() => formatTodayDate(), []);
	const currentTime = useMemo(() => formatTime(), []);
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
		<div className="flex flex-col h-full animate-fade-in">
			{/* Sticky page header */}
			<div className="page-header-blur flex items-center h-[49px] px-5 shrink-0">
				<span className="text-[15px] font-semibold">Главная</span>
				<div className="ml-auto flex items-center gap-2">
					<button
						type="button"
						className="w-[34px] h-[34px] rounded-lg bg-transparent text-muted-foreground hover:bg-white/[0.07] hover:text-white flex items-center justify-center transition-all"
						title="Поиск (Ctrl+K)"
					>
						<Search className="h-4 w-4" />
					</button>
					<button
						type="button"
						className="w-[34px] h-[34px] rounded-lg bg-transparent text-muted-foreground hover:bg-white/[0.07] hover:text-white flex items-center justify-center transition-all"
						title="Уведомления"
					>
						<Bell className="h-4 w-4" />
					</button>
					<button
						type="button"
						className="btn-glow bg-gradient-to-br from-blue-500 to-blue-600 text-white text-[13px] font-medium px-3 py-1.5 rounded-lg flex items-center gap-1.5"
						onClick={() => navigate({ to: "/projects" })}
					>
						<Plus className="h-3.5 w-3.5" />
						Новая задача
					</button>
				</div>
			</div>

			{/* Scrollable content */}
			<div className="flex-1 overflow-y-auto p-6">
				{/* Greeting */}
				<div className="mb-6">
					<h1 className="text-[28px] font-bold tracking-[-0.015em] greeting-gradient">
						{greeting}
						{displayName ? `, ${displayName}` : ""} <span>👋</span>
					</h1>
					<p className="text-sm text-muted-foreground capitalize mt-1">
						{todayDate}
						<span className="mx-1.5 text-border">|</span>
						{currentTime}
					</p>
				</div>

				{/* Quick actions */}
				<div className="flex gap-2 mb-6 flex-wrap">
					<button
						type="button"
						className="flex items-center gap-2 px-3.5 py-2 bg-card border border-white/[0.06] rounded-lg text-[13px] text-muted-foreground hover:text-white hover:bg-white/[0.07] hover:-translate-y-[2px] hover:shadow-lg transition-all duration-200"
						onClick={() => navigate({ to: "/projects" })}
					>
						<Plus className="h-3.5 w-3.5" />
						Новая задача
					</button>
					<button
						type="button"
						className="flex items-center gap-2 px-3.5 py-2 bg-card border border-white/[0.06] rounded-lg text-[13px] text-muted-foreground hover:text-white hover:bg-white/[0.07] hover:-translate-y-[2px] hover:shadow-lg transition-all duration-200"
						onClick={() => navigate({ to: "/messenger" })}
					>
						<MessageSquarePlus className="h-3.5 w-3.5" />
						Написать
					</button>
					<button
						type="button"
						className="flex items-center gap-2 px-3.5 py-2 bg-card border border-white/[0.06] rounded-lg text-[13px] text-muted-foreground hover:text-white hover:bg-white/[0.07] hover:-translate-y-[2px] hover:shadow-lg transition-all duration-200"
						onClick={() => navigate({ to: "/calls" })}
					>
						<CalendarPlus className="h-3.5 w-3.5" />
						Событие
					</button>
					<button
						type="button"
						className="flex items-center gap-2 px-3.5 py-2 bg-card border border-white/[0.06] rounded-lg text-[13px] text-muted-foreground hover:text-white hover:bg-white/[0.07] hover:-translate-y-[2px] hover:shadow-lg transition-all duration-200"
					>
						<Search className="h-3.5 w-3.5" />
						Поиск
					</button>
				</div>

				{/* Widget grid */}
				{loading && !personalData ? (
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
						{[1, 2, 3, 4, 5, 6].map((i) => (
							<div
								key={i}
								className="rounded-2xl border border-white/[0.06] bg-card p-5 space-y-3 shadow-sm"
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
		</div>
	);
}
