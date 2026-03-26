/**
 * Страница админского дашборда с табами.
 * Дашборд | Пользователи | Журнал аудита | Лимиты
 * Доступна только администраторам.
 */
import { createFileRoute, redirect } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { adminDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { AuditLogViewer } from "@/components/admin/audit-log";
import { ResourceLimits } from "@/components/admin/resource-limits";
import { UserManagement } from "@/components/admin/user-management";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
	const [activeTab, setActiveTab] = useState("dashboard");

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
			{/* Page header */}
			<div className="page-header-blur mb-6 flex items-center justify-between">
				<h1 className="text-2xl font-bold">
					{"\u0410\u0434\u043C\u0438\u043D\u0438\u0441\u0442\u0440\u0438\u0440\u043E\u0432\u0430\u043D\u0438\u0435"}
				</h1>
				<div className="flex items-center gap-2">
					<Button variant="ghost" size="sm" className="gap-1.5 text-muted-foreground">
						<span>{"\u{1F4E5}"}</span>
						{"\u042D\u043A\u0441\u043F\u043E\u0440\u0442 CSV"}
					</Button>
					<Button
						size="sm"
						className="btn-glow gap-1.5 bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-500 hover:to-blue-400 text-white shadow-sm shadow-blue-500/20"
					>
						+ {"\u041F\u0440\u0438\u0433\u043B\u0430\u0441\u0438\u0442\u044C"}
					</Button>
				</div>
			</div>

			{/* Tab bar */}
			<Tabs value={activeTab} onValueChange={setActiveTab}>
				<TabsList className="mb-6">
					<TabsTrigger value="dashboard">
						{"\u0414\u0430\u0448\u0431\u043E\u0440\u0434"}
					</TabsTrigger>
					<TabsTrigger value="users">
						{"\u041F\u043E\u043B\u044C\u0437\u043E\u0432\u0430\u0442\u0435\u043B\u0438"}
					</TabsTrigger>
					<TabsTrigger value="audit">
						{"\u0416\u0443\u0440\u043D\u0430\u043B \u0430\u0443\u0434\u0438\u0442\u0430"}
					</TabsTrigger>
					<TabsTrigger value="limits">
						{"\u041B\u0438\u043C\u0438\u0442\u044B"}
					</TabsTrigger>
				</TabsList>

				{/* Dashboard tab */}
				<TabsContent value="dashboard">
					{adminData && (
						<div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
							<KpiCard label={"\u041F\u043E\u043B\u044C\u0437\u043E\u0432\u0430\u0442\u0435\u043B\u0435\u0439"} value={adminData.total_users} change={+5} />
							<KpiCard label={"\u041F\u0440\u043E\u0435\u043A\u0442\u043E\u0432"} value={adminData.total_projects} change={+2} />
							<KpiCard label={"\u0417\u0430\u0434\u0430\u0447"} value={adminData.total_issues} change={-3} />
							<KpiCard
								label="DAU / WAU / MAU"
								value={`${adminData.dau} / ${adminData.wau} / ${adminData.mau}`}
							/>
						</div>
					)}
					{loading && !adminData ? (
						<p className="text-muted-foreground">{"\u0417\u0430\u0433\u0440\u0443\u0437\u043A\u0430..."}</p>
					) : (
						<WidgetGrid items={items} dashboardType="admin" data={adminData} />
					)}
				</TabsContent>

				{/* Users tab */}
				<TabsContent value="users">
					<UserManagement />
				</TabsContent>

				{/* Audit log tab */}
				<TabsContent value="audit">
					<AuditLogViewer />
				</TabsContent>

				{/* Limits tab */}
				<TabsContent value="limits">
					<ResourceLimits />
				</TabsContent>
			</Tabs>
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
