import type { AdminStats, PersonalData, ProjectStats } from "@/api/dashboard";

/** Данные виджетов дашборда: персональные / проект / admin. */
export type DashboardWidgetData = PersonalData | ProjectStats | AdminStats;

export function isPersonalData(
	data: DashboardWidgetData | undefined,
): data is PersonalData {
	return !!data && "my_tasks" in data;
}

export function isProjectStats(
	data: DashboardWidgetData | undefined,
): data is ProjectStats {
	return !!data && "burndown_data" in data;
}

export function isAdminStats(
	data: DashboardWidgetData | undefined,
): data is AdminStats {
	return !!data && "total_users" in data;
}
