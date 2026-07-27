/**
 * API-клиент для модуля дашбордов.
 * Предоставляет методы для получения данных и управления раскладками виджетов.
 */
import { api } from "./client";

/** Элемент раскладки виджета. */
export interface LayoutItem {
	i: string;
	x: number;
	y: number;
	w: number;
	h: number;
	widget_type: string;
	config?: Record<string, unknown>;
}

/** Раскладка дашборда. */
export interface LayoutDTO {
	dashboard_type: string;
	project_id?: string;
	items: LayoutItem[];
}

/** Краткая информация о задаче. */
export interface TaskSummary {
	id: string;
	key: string;
	title: string;
	status: string;
	priority: number;
	due_date?: string;
}

/** Краткая информация о звонке. */
export interface CallSummary {
	id: string;
	title: string;
	scheduled_at: string;
	channel_id?: string;
	channel_name: string;
}

/** Краткая информация о прогоне тестов. */
export interface RunSummary {
	id: string;
	name: string;
	status: string;
	passed: number;
	failed: number;
	skipped: number;
	created_at: string;
}

/** Данные персонального дашборда. */
export interface PersonalData {
	my_tasks: TaskSummary[];
	unread_count: number;
	upcoming_calls: CallSummary[];
	recent_runs: RunSummary[];
	deadlines: TaskSummary[];
	time_today: number;
}

/** Точка данных для графиков. */
export interface DataPoint {
	label: string;
	value: number;
}

/** Запись активности. */
export interface ActivityEntry {
	user_name: string;
	action: string;
	target: string;
	created_at: string;
}

/** Статистика проекта. */
export interface ProjectStats {
	tasks_by_status: Record<string, number>;
	burndown_data: DataPoint[];
	velocity_data: DataPoint[];
	bug_count: number;
	team_activity: ActivityEntry[];
	test_trends: DataPoint[];
	coverage: number;
	flaky_count: number;
	unassigned: number;
	overdue: number;
}

/** Рейтинг проекта. */
export interface ProjectRank {
	id: string;
	name: string;
	issue_count: number;
}

/** Статистика администратора. */
export interface AdminStats {
	total_users: number;
	active_users: number;
	total_projects: number;
	total_issues: number;
	dau: number;
	wau: number;
	mau: number;
	storage_used: number;
	service_health: Record<string, string>;
	top_projects: ProjectRank[];
}

/** Получить данные персонального дашборда. */
export function getPersonalDashboard() {
	return api.get<PersonalData>("/dashboard/personal");
}

/** Получить статистику проекта. */
export function getProjectDashboard(projectId: string) {
	return api.get<ProjectStats>(`/dashboard/project/${projectId}`);
}

/** Получить статистику администратора. */
export function getAdminDashboard() {
	return api.get<AdminStats>("/dashboard/admin");
}

/** Получить раскладку дашборда. */
export function getLayout(type: string, projectId?: string) {
	const params = new URLSearchParams({ type });
	if (projectId) params.set("project_id", projectId);
	return api.get<LayoutDTO>(`/dashboard/layout?${params.toString()}`);
}

/** Сохранить раскладку дашборда. */
export function saveLayout(
	dashboardType: string,
	items: LayoutItem[],
	projectId?: string,
) {
	return api.put<void>("/dashboard/layout", {
		dashboard_type: dashboardType,
		project_id: projectId,
		items,
	});
}

/** Список пресетов шаблонов раскладки (GET /dashboard/templates). */
export function listDashboardTemplates() {
	return api.get<Record<string, LayoutItem[]>>("/dashboard/templates");
}

/** Применить шаблон к проектному дашборду (POST /dashboard/templates/apply). */
export function applyDashboardTemplate(projectId: string, template: string) {
	return api.post<void>("/dashboard/templates/apply", {
		project_id: projectId,
		template,
	});
}
