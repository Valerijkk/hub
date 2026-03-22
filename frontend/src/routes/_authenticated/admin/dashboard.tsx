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
			<h1 className="text-2xl font-bold mb-4">Панель администратора</h1>

			{/* KPI карточки */}
			{adminData && (
				<div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
					<KpiCard label="Пользователей" value={adminData.total_users} />
					<KpiCard label="Проектов" value={adminData.total_projects} />
					<KpiCard label="Задач" value={adminData.total_issues} />
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
function KpiCard({ label, value }: { label: string; value: string | number }) {
	return (
		<div className="rounded-lg border bg-card p-4 shadow-sm">
			<p className="text-sm text-muted-foreground">{label}</p>
			<p className="text-2xl font-bold">{value}</p>
		</div>
	);
}
