/**
 * Персональный дашборд пользователя.
 * Виджеты: задачи, непрочитанные, звонки, дедлайны, таймер.
 */

import { createLazyFileRoute, useNavigate } from "@tanstack/react-router";
import {
	Bell,
	CalendarPlus,
	MessageSquarePlus,
	Plus,
	Search,
} from "lucide-react";
import { useEffect, useMemo } from "react";
import { personalDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { useAuthStore } from "@/stores/auth-store";
import { useDashboardStore } from "@/stores/dashboard-store";

/** Обёртка для изоляции рантайм-ошибок раздела «Главная» (FE-H4). */
function PersonalDashboardWithBoundary() {
	return (
		<RouteErrorBoundary module="Главная">
			<PersonalDashboard />
		</RouteErrorBoundary>
	);
}

export const Route = createLazyFileRoute("/_authenticated/dashboard")({
	component: PersonalDashboardWithBoundary,
});

/** Приветствие на русском в зависимости от часа. */
function getGreeting(): string {
	const hour = new Date().getHours();
	if (hour >= 5 && hour < 12) return "Доброе утро";
	if (hour >= 12 && hour < 17) return "Добрый день";
	if (hour >= 17 && hour < 22) return "Добрый вечер";
	return "Доброй ночи";
}

/** Сегодняшняя дата в русской локали. */
function formatTodayDate(): string {
	return new Date().toLocaleDateString("ru-RU", {
		weekday: "long",
		day: "numeric",
		month: "long",
	});
}

/** Текущее время для шапки дашборда. */
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
	const personalItems = useDashboardStore((s) => s.layouts.personal);
	const setLayout = useDashboardStore((s) => s.setLayout);
	const loading = useDashboardStore((s) => s.loading);
	const error = useDashboardStore((s) => s.error);
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
			{/* Липкий заголовок страницы */}
			<div className="page-header-blur flex items-center h-[49px] px-5 shrink-0">
				<span className="text-[15px] font-semibold">Главная</span>
				<div className="ml-auto flex items-center gap-2">
					<button
						type="button"
						className="w-[34px] h-[34px] rounded-lg bg-transparent text-t2 hover:bg-white/[0.07] hover:text-white flex items-center justify-center transition-colors"
						aria-label="Поиск"
					>
						<Search className="h-4 w-4" />
					</button>
					<button
						type="button"
						className="w-[34px] h-[34px] rounded-lg bg-transparent text-t2 hover:bg-white/[0.07] hover:text-white flex items-center justify-center transition-colors"
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

			{/* Прокручиваемый контент */}
			<div className="flex-1 overflow-y-auto p-6">
				{/* Приветствие */}
				<div className="mb-2.5 dash-enter">
					<h1 className="text-[34px] font-black tracking-[-0.04em] leading-tight dash-greeting-gradient dash-enter">
						{greeting}
						{displayName ? `, ${displayName}` : ""} <span>👋</span>
					</h1>
					<p className="text-sm text-[var(--t2)] mt-1.5 dash-enter">
						{todayDate}
						<span className="mx-1.5 text-[var(--t4)]">&middot;</span>
						{currentTime}
					</p>
				</div>

				{/* Быстрые действия */}
				<div className="flex gap-2.5 mb-7 flex-wrap dash-enter dash-enter-2">
					<button
						type="button"
						className="dash-qa-btn flex items-center gap-2.5 px-[18px] py-[11px] rounded-[var(--r3)] text-[13px] font-semibold text-[var(--t2)] cursor-pointer transition-colors"
						onClick={() => navigate({ to: "/projects" })}
					>
						<div
							className="w-[34px] h-[34px] rounded-[var(--r2)] flex items-center justify-center"
							style={{ background: "rgba(91,156,245,.06)" }}
						>
							<Plus
								className="h-[18px] w-[18px]"
								style={{ color: "var(--blue)" }}
							/>
						</div>
						Новая задача
					</button>
					<button
						type="button"
						className="dash-qa-btn flex items-center gap-2.5 px-[18px] py-[11px] rounded-[var(--r3)] text-[13px] font-semibold text-[var(--t2)] cursor-pointer transition-colors"
						onClick={() => navigate({ to: "/messenger" })}
					>
						<div
							className="w-[34px] h-[34px] rounded-[var(--r2)] flex items-center justify-center"
							style={{ background: "rgba(52,211,153,.06)" }}
						>
							<MessageSquarePlus
								className="h-[18px] w-[18px]"
								style={{ color: "var(--green)" }}
							/>
						</div>
						Написать
					</button>
					<button
						type="button"
						className="dash-qa-btn flex items-center gap-2.5 px-[18px] py-[11px] rounded-[var(--r3)] text-[13px] font-semibold text-[var(--t2)] cursor-pointer transition-colors"
						onClick={() => navigate({ to: "/calls" })}
					>
						<div
							className="w-[34px] h-[34px] rounded-[var(--r2)] flex items-center justify-center"
							style={{ background: "rgba(167,139,250,.06)" }}
						>
							<CalendarPlus
								className="h-[18px] w-[18px]"
								style={{ color: "var(--purple)" }}
							/>
						</div>
						Событие
					</button>
					<button
						type="button"
						className="dash-qa-btn flex items-center gap-2.5 px-[18px] py-[11px] rounded-[var(--r3)] text-[13px] font-semibold text-[var(--t2)] cursor-pointer transition-colors"
					>
						<div
							className="w-[34px] h-[34px] rounded-[var(--r2)] flex items-center justify-center"
							style={{ background: "rgba(255,255,255,.03)" }}
						>
							<Search
								className="h-[18px] w-[18px]"
								style={{ color: "var(--t2)" }}
							/>
						</div>
						Поиск
					</button>
				</div>

				{/* Сетка виджетов */}
				{loading && !personalData ? (
					<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 dash-enter dash-enter-3">
						{[1, 2, 3, 4, 5, 6].map((i) => (
							<div
								key={`dash-sk-${i}`}
								className="dash-widget p-5 space-y-3"
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
				) : error && !personalData ? (
					<div className="flex flex-col items-center justify-center py-20 gap-3 text-t2 dash-enter dash-enter-3">
						<p className="text-sm font-medium">{error}</p>
						<button
							type="button"
							className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
							onClick={() => fetchPersonal()}
						>
							Попробовать снова
						</button>
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
