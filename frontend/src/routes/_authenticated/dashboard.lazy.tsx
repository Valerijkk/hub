/**
 * Personal user dashboard.
 * Displays widgets with tasks, unread messages, calls, deadlines and time tracking.
 */
import {
	CalendarPlus,
	MessageSquarePlus,
	Plus,
	Search,
} from "lucide-react";
import { createLazyFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useMemo } from "react";
import { personalDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { Button } from "@/components/ui/button";
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
		<div className="animate-fade-in">
			{/* Greeting header */}
			<div className="mb-8">
				<div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
					<div className="flex items-center gap-3">
						<div className="flex h-10 w-10 items-center justify-center rounded-xl gradient-dashboard text-white shadow-sm">
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
							<h1 className="text-[28px] font-bold tracking-[-0.015em] greeting-gradient">
								{greeting}
								{displayName ? `, ${displayName}` : ""}
							</h1>
							<p className="text-sm text-muted-foreground capitalize">
								{todayDate}
								<span className="mx-1.5 text-border">|</span>
								{currentTime}
							</p>
						</div>
					</div>

					{/* Quick action buttons */}
					<div className="flex items-center gap-2">
						<Button
							variant="outline"
							size="sm"
							className="gap-1.5 rounded-lg border bg-card transition-all duration-200 hover:-translate-y-[2px] hover:shadow-lg"
							onClick={() => navigate({ to: "/projects" })}
						>
							<Plus className="h-3.5 w-3.5" />
							Новая задача
						</Button>
						<Button
							variant="outline"
							size="sm"
							className="gap-1.5 rounded-lg border bg-card transition-all duration-200 hover:-translate-y-[2px] hover:shadow-lg"
							onClick={() => navigate({ to: "/messenger" })}
						>
							<MessageSquarePlus className="h-3.5 w-3.5" />
							Написать
						</Button>
						<Button
							variant="outline"
							size="sm"
							className="gap-1.5 rounded-lg border bg-card transition-all duration-200 hover:-translate-y-[2px] hover:shadow-lg"
							onClick={() => navigate({ to: "/calls" })}
						>
							<CalendarPlus className="h-3.5 w-3.5" />
							Событие
						</Button>
						<Button
							variant="ghost"
							size="sm"
							className="rounded-lg transition-all duration-200 hover:-translate-y-[2px] hover:shadow-lg"
						>
							<Search className="h-4 w-4" />
						</Button>
					</div>
				</div>
			</div>

			{loading && !personalData ? (
				/* Loading skeletons */
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
	);
}
