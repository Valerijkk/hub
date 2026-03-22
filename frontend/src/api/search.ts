import { api } from "./client";

/** Результат глобального поиска. */
export interface SearchResult {
	id: string;
	type:
		| "issue"
		| "test_case"
		| "user"
		| "project"
		| "channel"
		| "file"
		| "message";
	title: string;
	subtitle?: string;
	project_id?: string;
	url: string;
	score: number;
}

/** Ответ глобального поиска. */
export interface SearchResponse {
	results: SearchResult[];
	total: number;
}

/** Сохранённый фильтр трекера. */
export interface SavedFilter {
	id: string;
	name: string;
	filter_config: Record<string, unknown>;
	shared: boolean;
	owner_id: string;
	project_id: string;
	created_at: string;
	updated_at: string;
}

/** Ответ со списком сохранённых фильтров. */
export interface SavedFilterListResponse {
	filters: SavedFilter[];
}

/** Запрос на создание сохранённого фильтра. */
export interface CreateSavedFilterRequest {
	name: string;
	filter_config: Record<string, unknown>;
	project_id: string;
}

/** Глобальный поиск по всем сущностям. */
export function searchGlobal(
	query: string,
	types?: string[],
	limit?: number,
): Promise<SearchResponse> {
	const params = new URLSearchParams({ q: query });
	if (types?.length) params.set("type", types.join(","));
	if (limit) params.set("limit", String(limit));
	return api.get<SearchResponse>(`/search?${params.toString()}`);
}

/** Создание сохранённого фильтра. */
export function createSavedFilter(
	data: CreateSavedFilterRequest,
): Promise<SavedFilter> {
	return api.post<SavedFilter>("/filters", data);
}

/** Получение списка сохранённых фильтров (свои + расшаренные в проекте). */
export function listSavedFilters(
	projectId: string,
): Promise<SavedFilterListResponse> {
	return api.get<SavedFilterListResponse>(`/filters?project_id=${projectId}`);
}

/** Удаление сохранённого фильтра. */
export function deleteSavedFilter(id: string): Promise<void> {
	return api.delete<void>(`/filters/${id}`);
}

/** Переключение расшаривания фильтра. */
export function shareFilter(id: string, shared: boolean): Promise<SavedFilter> {
	return api.patch<SavedFilter>(`/filters/${id}/share`, { shared });
}

// ---------------------------------------------------------------------------
// JQL-поиск
// ---------------------------------------------------------------------------

/** Элемент результата JQL-поиска. */
export interface JQLResultItem {
	key: string;
	title: string;
	type: string;
	priority: string;
	status?: string;
	assignee?: string;
	created_at: string;
}

/** Ответ JQL-поиска. */
export interface JQLSearchResponse {
	query: string;
	total: number;
	results: JQLResultItem[];
}

/** Выполнить JQL-запрос в проекте. */
export function searchJQL(
	projectId: string,
	query: string,
): Promise<JQLSearchResponse> {
	return api.post<JQLSearchResponse>(`/projects/${projectId}/issues/jql`, {
		query,
	});
}
