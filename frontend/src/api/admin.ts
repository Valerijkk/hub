import { api } from "./client";

// --- Типы ответов ---

/** Запись журнала аудита. */
export interface AuditLogEntry {
	id: string;
	action: string;
	entity_type: string;
	entity_id?: string;
	user_id?: string;
	project_id?: string;
	details?: Record<string, unknown>;
	ip_address?: string;
	user_agent?: string;
	created_at: string;
}

/** Пагинированный ответ журнала аудита. */
export interface AuditLogListResponse {
	items: AuditLogEntry[];
	total: number;
	page: number;
	per_page: number;
}

/** Фильтры для запроса журнала аудита. */
export interface AuditLogFilters {
	action?: string;
	entity_type?: string;
	entity_id?: string;
	user_id?: string;
	project_id?: string;
	ip_address?: string;
	slow?: boolean;
	date_from?: string;
	date_to?: string;
	page?: number;
	per_page?: number;
}

/** Лимит ресурса. */
export interface ResourceLimit {
	key: string;
	value: string;
}

/** Элемент корзины. */
export interface TrashItem {
	id: string;
	name: string;
	entity_type: string;
	deleted_at: string;
	days_remaining: number;
}

/** Пагинированный ответ корзины. */
export interface TrashListResponse {
	items: TrashItem[];
	total: number;
	page: number;
	per_page: number;
}

/** Ответ на операцию управления. */
export interface ManagementResponse {
	success: boolean;
	message: string;
}

/** Агрегированная метрика observability.Timed (OBS-01). */
export interface OperationMetric {
	operation: string;
	count: number;
	error_count: number;
	total_ms: number;
	max_ms: number;
	avg_ms: number;
	last_seen_utc?: string;
}

// --- API клиент ---

/** Получить список записей журнала аудита с фильтрами. */
export function getAuditLogs(filters: AuditLogFilters = {}) {
	const params = new URLSearchParams();
	if (filters.action) params.set("action", filters.action);
	if (filters.entity_type) params.set("entity_type", filters.entity_type);
	if (filters.entity_id) params.set("entity_id", filters.entity_id);
	if (filters.user_id) params.set("user_id", filters.user_id);
	if (filters.project_id) params.set("project_id", filters.project_id);
	if (filters.ip_address) params.set("ip_address", filters.ip_address);
	if (filters.slow) params.set("slow", "1");
	if (filters.date_from) params.set("date_from", filters.date_from);
	if (filters.date_to) params.set("date_to", filters.date_to);
	if (filters.page) params.set("page", String(filters.page));
	if (filters.per_page) params.set("per_page", String(filters.per_page));
	const qs = params.toString();
	return api.get<AuditLogListResponse>(`/admin/audit${qs ? `?${qs}` : ""}`);
}

/** Получить in-memory метрики Timed()-операций (admin-only). */
export function getOperationMetrics() {
	return api.get<OperationMetric[]>("/admin/observability/ops");
}

/** Получить текущие лимиты ресурсов. */
export function getLimits() {
	return api.get<ResourceLimit[]>("/admin/limits");
}

/** Обновить лимит ресурса. */
export function updateLimit(key: string, value: string) {
	return api.put<{ success: boolean }>(`/admin/limits/${key}`, { value });
}

/** Получить список элементов корзины. */
export function getTrash(type: string, projectId?: string, page?: number) {
	const params = new URLSearchParams({ type });
	if (projectId) params.set("project_id", projectId);
	if (page) params.set("page", String(page));
	return api.get<TrashListResponse>(`/admin/trash?${params.toString()}`);
}

/** Восстановить элемент из корзины. */
export function restoreItem(type: string, id: string) {
	return api.post<ManagementResponse>(`/admin/trash/${type}/${id}/restore`);
}

/** Деактивировать пользователя. */
export function deactivateUser(id: string) {
	return api.post<ManagementResponse>(`/admin/users/${id}/deactivate`);
}

/** Заблокировать пользователя с указанием причины. */
export function banUser(id: string, reason: string) {
	return api.post<ManagementResponse>(`/admin/users/${id}/ban`, { reason });
}

/** Переназначить задачи пользователя. */
export function reassignTasks(fromId: string, toId: string) {
	return api.post<{ success: boolean; reassigned_count: number }>(
		`/admin/users/${fromId}/reassign`,
		{ to_user_id: toId },
	);
}

/** Удалить данные пользователя (GDPR). */
export function gdprDeleteUser(id: string) {
	return api.delete<ManagementResponse>(`/admin/users/${id}/gdpr`);
}
