/**
 * Страница админского дашборда.
 * Отображает здоровье сервисов, хранилище, активность пользователей.
 * Доступна только администраторам.
 */
import { createFileRoute, redirect } from "@tanstack/react-router";
import { useEffect } from "react";
import { adminDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { useAuthStore } from "@/stores/auth-store";
import { useDashboardStore } from "@/stores/dashboard-store";

export const Route = createFileRoute("/_authenticated/admin/dashboard")({
	beforeLoad: () => {
		const user = useAuthStore.getState().user;
		if (!user || user.system_role !== "admin") {
			throw redirect({ to: "/dashboard" });
		}
	},
	component: AdminDashboard,
});

function AdminDashboard() {
	const fetchAdmin = useDashboardStore((s) => s.fetchAdmin);
	const loadLayout = useDashboardStore((s) => s.loadLayout);
	const adminData = useDashboardStore((s) => s.adminData);
	const layouts = useDashboardStore((s) => s.layouts);
	const setLayout = useDashboardStore((s) => s.setLayout);
	const loading = useDashboardStore((s) => s.loading);

	useEffect(() => {
		fetchAdmin();
		loadLayout("admin").then(() => {
			const current = useDashboardStore.getState().layouts["admin"];
			if (!current || current.length === 0) {
				setLayout("admin", adminDefault);
			}
		});
	}, [fetchAdmin, loadLayout, setLayout]);

	const items = layouts["admin"] ?? adminDefault;

	return (
		<div>
			<div className="page-header-blur mb-6">
				<h1 className="text-2xl font-bold">Панель администратора</h1>
			</div>

			{/* KPI карточки */}
			{adminData && (
				<div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
					<KpiCard label="Пользователей" value={adminData.total_users} change={+5} />
					<KpiCard label="Проектов" value={adminData.total_projects} change={+2} />
					<KpiCard label="Задач" value={adminData.total_issues} change={-3} />
					<KpiCard
						label="DAU / WAU / MAU"
						value={`${adminData.dau} / ${adminData.wau} / ${adminData.mau}`}
					/>
				</div>
			)}

			{loading && !adminData ? (
				<p className="text-muted-foreground">Загрузка...</p>
			) : (
				<WidgetGrid items={items} dashboardType="admin" data={adminData} />
			)}
		</div>
	);
}

/** KPI-карточка для быстрого обзора метрик. */
function KpiCard({ label, value, change }: { label: string; value: string | number; change?: number }) {
	return (
		<div className="rounded-2xl border border-white/[0.06] bg-card p-5">
			<p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{label}</p>
			<p className="mt-1 text-[28px] font-bold leading-tight">{value}</p>
			{change != null && (
				<p className={`mt-1 text-xs font-medium ${change >= 0 ? "text-emerald-500" : "text-red-500"}`}>
					{change >= 0 ? "\u2191" : "\u2193"} {Math.abs(change)}%
				</p>
			)}
		</div>
	);
}
