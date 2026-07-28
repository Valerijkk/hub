/**
 * Страница админского дашборда с табами.
 * Дашборд | Пользователи | Журнал аудита | Лимиты
 * Доступна только администраторам.
 */
import { createFileRoute, redirect } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { AISettingsPanel } from "@/components/admin/ai-settings";
import { AuditLogViewer } from "@/components/admin/audit-log";
import { MailSettingsPanel } from "@/components/admin/mail-settings";
import { OnecSettingsPanel } from "@/components/admin/onec-settings";
import { OperationMetricsPanel } from "@/components/admin/operation-metrics";
import { ResourceLimits } from "@/components/admin/resource-limits";
import { UserManagement } from "@/components/admin/user-management";
import { adminDefault } from "@/components/dashboard/dashboard-templates";
import { WidgetGrid } from "@/components/dashboard/widget-grid";
import { Tabs, TabsContent } from "@/components/ui/tabs";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { useAuthStore } from "@/stores/auth-store";
import { useDashboardStore } from "@/stores/dashboard-store";

export const Route = createFileRoute("/_authenticated/admin/dashboard")({
	beforeLoad: () => {
		const user = useAuthStore.getState().user;
		if (!user || !isSystemAdmin(user.system_role)) {
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
	const error = useDashboardStore((s) => s.error);
	const [activeTab, setActiveTab] = useState("dashboard");

	useEffect(() => {
		fetchAdmin();
		loadLayout("admin").then(() => {
			const current = useDashboardStore.getState().layouts.admin;
			if (!current || current.length === 0) {
				setLayout("admin", adminDefault);
			}
		});
	}, [fetchAdmin, loadLayout, setLayout]);

	const items = layouts.admin ?? adminDefault;

	return (
		<div className="flex-1 overflow-y-auto p-6 animate-[eu_450ms_var(--spring)]">
			{/* Панель вкладок */}
			<Tabs value={activeTab} onValueChange={setActiveTab}>
				<div className="flex gap-[3px] mb-5 border-b border-b1 pb-[2px]">
					{[
						{
							value: "dashboard",
							label: "\u0414\u0430\u0448\u0431\u043E\u0440\u0434",
						},
						{
							value: "users",
							label:
								"\u041F\u043E\u043B\u044C\u0437\u043E\u0432\u0430\u0442\u0435\u043B\u0438",
						},
						{ value: "audit", label: "\u0410\u0443\u0434\u0438\u0442" },
						{
							value: "metrics",
							label: "\u041C\u0435\u0442\u0440\u0438\u043A\u0438",
						},
						{ value: "limits", label: "\u041B\u0438\u043C\u0438\u0442\u044B" },
						{ value: "mail", label: "\u041F\u043E\u0447\u0442\u0430" },
						{ value: "ai", label: "AI" },
						{ value: "onec", label: "1C" },
					].map((tab) => (
						<button
							key={tab.value}
							type="button"
							onClick={() => setActiveTab(tab.value)}
							className={`px-[14px] py-[7px] text-[13px] font-medium text-[var(--t2)] rounded-t-[var(--r2)] transition-colors duration-200 cursor-pointer relative hover:text-t1 ${
								activeTab === tab.value
									? "text-t1 font-semibold after:absolute after:bottom-[-2px] after:left-1/4 after:right-1/4 after:h-[2px] after:bg-red after:rounded-[1px] after:shadow-[0_0_6px_var(--red)]"
									: ""
							}`}
						>
							{tab.label}
						</button>
					))}
				</div>

				{/* Вкладка дашборда */}
				<TabsContent value="dashboard">
					{adminData && (
						<div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-7">
							<KpiCard
								label={
									"\u041F\u043E\u043B\u044C\u0437\u043E\u0432\u0430\u0442\u0435\u043B\u0435\u0439"
								}
								value={adminData.total_users}
								color="green"
							/>
							<KpiCard
								label={"\u041F\u0440\u043E\u0435\u043A\u0442\u043E\u0432"}
								value={adminData.total_projects}
								color="blue"
							/>
							<KpiCard
								label={"\u0417\u0430\u0434\u0430\u0447"}
								value={adminData.total_issues}
								color="purple"
							/>
							<KpiCard
								label="DAU / WAU / MAU"
								value={`${adminData.dau} / ${adminData.wau} / ${adminData.mau}`}
							/>
						</div>
					)}
					{loading && !adminData ? (
						<p className="text-[var(--t2)] text-[13px]">
							{"\u0417\u0430\u0433\u0440\u0443\u0437\u043A\u0430..."}
						</p>
					) : error && !adminData ? (
						<div
							role="alert"
							className="flex flex-col items-center justify-center py-16 gap-3 text-t2"
						>
							<p className="text-sm font-medium">{error}</p>
							<button
								type="button"
								className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
								onClick={() => void fetchAdmin()}
							>
								Попробовать снова
							</button>
						</div>
					) : (
						<WidgetGrid items={items} dashboardType="admin" data={adminData} />
					)}
				</TabsContent>

				{/* Вкладка пользователей */}
				<TabsContent value="users">
					<UserManagement />
				</TabsContent>

				{/* Вкладка audit log */}
				<TabsContent value="audit">
					<AuditLogViewer />
				</TabsContent>

				<TabsContent value="metrics">
					<OperationMetricsPanel />
				</TabsContent>

				{/* Вкладка лимитов */}
				<TabsContent value="limits">
					<ResourceLimits />
				</TabsContent>

				{/* Вкладка почты */}
				<TabsContent value="mail">
					<MailSettingsPanel />
				</TabsContent>

				<TabsContent value="ai">
					<AISettingsPanel />
				</TabsContent>

				<TabsContent value="onec">
					<OnecSettingsPanel />
				</TabsContent>
			</Tabs>
		</div>
	);
}

/** KPI-карточка для быстрого обзора метрик. */
function KpiCard({
	label,
	value,
	change,
	color,
}: {
	label: string;
	value: string | number;
	change?: number;
	color?: string;
}) {
	return (
		<div className="admin-kpi bg-gradient-to-b from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r4)] p-5 transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-250 hover:border-b2 hover:shadow-[var(--sh2)] hover:-translate-y-[2px]">
			<p className="text-[11px] font-bold uppercase tracking-[0.06em] text-[var(--t2)] mb-2">
				{label}
			</p>
			<div className="flex items-end gap-[10px]">
				<span className="text-[30px] font-black tracking-[-0.03em]">
					{value}
				</span>
				{change != null && (
					<span
						className={`text-[12px] font-bold flex items-center gap-[3px] mb-1 ${change >= 0 ? "text-green" : "text-red"}`}
					>
						{change >= 0 ? "\u2191" : "\u2193"} {change >= 0 ? "+" : ""}
						{Math.abs(change)}
					</span>
				)}
				{color && (
					<div className="ml-auto">
						<svg aria-hidden="true" className="w-20 h-8" viewBox="0 0 80 32">
							<polyline
								points="0,28 12,24 24,26 36,20 48,18 60,15 72,10 80,8"
								fill="none"
								stroke={`var(--${color})`}
								strokeWidth="2"
								strokeLinecap="round"
							/>
						</svg>
					</div>
				)}
			</div>
		</div>
	);
}
