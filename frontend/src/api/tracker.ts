import { api } from "./client";
import { buildQueryString } from "@/lib/utils";
import type { PaginatedResponse } from "./types";

// Re-export for backward compatibility.
export type { PaginatedResponse } from "./types";

// ---------------------------------------------------------------------------
// Типы перечислений
// ---------------------------------------------------------------------------

/** Тип задачи в трекере. */
export type IssueType =
	| "bug"
	| "task"
	| "story"
	| "epic"
	| "sub_task"
	| "improvement"
	| "tech_debt"
	| "spike"
	| "release";

/** Приоритет задачи. */
export type IssuePriority = "critical" | "high" | "medium" | "low" | "trivial";

/** Элемент чеклиста задачи. */
export interface ChecklistItem {
	id: string;
	text: string;
	checked: boolean;
}

// ---------------------------------------------------------------------------
// Интерфейсы ответов (Response DTOs)
// ---------------------------------------------------------------------------

/** Краткая информация о пользователе для отображения в карточках и списках. */
export interface UserBriefResponse {
	id: string;
	name: string;
	email: string;
	avatar_url?: string;
}

/** Статус workflow в проекте. */
export interface WorkflowStatusResponse {
	id: string;
	name: string;
	category: "todo" | "in_progress" | "done";
	color: string;
	sort_order: number;
	is_default: boolean;
}

/** Переход workflow между двумя статусами. */
export interface WorkflowTransitionResponse {
	id: string;
	from_status_id: string;
	to_status_id: string;
}

/** Метка, прикреплённая к задаче. */
export interface LabelResponse {
	id: string;
	name: string;
	color: string;
	group_name?: string;
}

/** Группа меток для организации меток. */
export interface LabelGroupResponse {
	id: string;
	name: string;
	labels: LabelResponse[];
}

/** Краткая информация о задаче для Kanban-карточек и ссылок. */
export interface IssueBriefResponse {
	id: string;
	issue_key: string;
	title: string;
	issue_type: IssueType;
	priority: IssuePriority;
	status: string;
	rank?: string;
	story_points?: number;
	checklist?: ChecklistItem[];
	assignee?: UserBriefResponse;
}

/** Полный ответ с детальной информацией о задаче. */
export interface IssueResponse {
	id: string;
	issue_number: number;
	issue_key: string;
	title: string;
	description_json?: Record<string, unknown>;
	description_text?: string;
	issue_type: IssueType;
	priority: IssuePriority;
	rank?: string;
	version: number;
	estimate_minutes?: number;
	story_points?: number;
	time_spent_minutes: number;
	checklist?: ChecklistItem[];
	custom_fields?: Record<string, unknown>;
	due_date?: string;
	status: WorkflowStatusResponse;
	assignee?: UserBriefResponse;
	reporter?: UserBriefResponse;
	parent?: IssueBriefResponse;
	children_count: number;
	labels: LabelResponse[];
	attachments?: {
		file_id: string;
		file_name: string;
		file_size: number;
		mime_type: string;
		url?: string;
	}[];
	created_at: string;
	updated_at: string;
}

/** Связь между двумя задачами. */
export interface IssueLinkResponse {
	id: string;
	link_type: string;
	custom_type?: string;
	issue: IssueBriefResponse;
}

/** Комментарий к задаче. */
export interface CommentResponse {
	id: string;
	body_json: Record<string, unknown>;
	body_text: string;
	is_internal: boolean;
	mentions: string[];
	author: UserBriefResponse;
	reactions: ReactionResponse[];
	created_at: string;
	updated_at: string;
}

/** Реакция на комментарий. */
export interface ReactionResponse {
	emoji: string;
	count: number;
	users: UserBriefResponse[];
	has_reacted: boolean;
}

/** Запись лога активности. */
export interface ActivityLogResponse {
	id: string;
	action: string;
	changes: FieldChange[];
	user: UserBriefResponse;
	created_at: string;
}

/** Одно изменение поля в логе активности. */
export interface FieldChange {
	field: string;
	old_value?: unknown;
	new_value: unknown;
}

/** Запись лога времени. */
export interface TimeLogResponse {
	id: string;
	minutes: number;
	description?: string;
	started_at?: string;
	ended_at?: string;
	user: UserBriefResponse;
	created_at: string;
}

/** Конфигурация доски. */
export interface BoardResponse {
	id: string;
	name: string;
	filters?: Record<string, unknown>;
	is_default: boolean;
	created_at: string;
}

/** Колонка доски (один статус workflow). */
export interface BoardColumnResponse {
	status: WorkflowStatusResponse;
	issues: IssueBriefResponse[];
}

/** Полные данные доски с колонками и задачами. */
export interface BoardDataResponse {
	board: BoardResponse;
	columns: BoardColumnResponse[];
}

/** Определение кастомного поля. */
export interface CustomFieldDefinitionResponse {
	id: string;
	name: string;
	field_type:
		| "text"
		| "number"
		| "date"
		| "checkbox"
		| "select"
		| "multi_select"
		| "user"
		| "url";
	options?: string[];
	required: boolean;
	sort_order: number;
}

// ---------------------------------------------------------------------------
// Интерфейсы запросов (Request DTOs)
// ---------------------------------------------------------------------------

/** Запрос на создание задачи. */
export interface CreateIssueRequest {
	title: string;
	issue_type: IssueType;
	description_json?: Record<string, unknown>;
	priority?: IssuePriority;
	assignee_id?: string;
	parent_id?: string;
	status_id?: string;
	labels?: string[];
	custom_fields?: Record<string, unknown>;
	estimate_minutes?: number;
	story_points?: number;
	checklist?: ChecklistItem[];
	due_date?: string;
}

/** Запрос на обновление задачи. */
export interface UpdateIssueRequest {
	title?: string;
	description_json?: Record<string, unknown>;
	priority?: IssuePriority;
	assignee_id?: string;
	status_id?: string;
	labels?: string[];
	custom_fields?: Record<string, unknown>;
	estimate_minutes?: number;
	story_points?: number;
	checklist?: ChecklistItem[];
	due_date?: string;
	version: number;
}

/** Запрос на перемещение задачи на доске. */
export interface MoveIssueRequest {
	status_id: string;
	rank: string;
}

/** Параметры фильтрации и пагинации списка задач. */
export interface ListIssuesParams {
	cursor?: string;
	limit?: number;
	status?: string[];
	priority?: string[];
	assignee?: string;
	issue_type?: string;
	search?: string;
	order_by?: string;
}

/** Запрос на создание связи между задачами. */
export interface CreateIssueLinkRequest {
	target_issue_id: string;
	link_type: string;
	custom_type?: string;
}

/** Запрос на создание комментария. */
export interface CreateCommentRequest {
	body_json: Record<string, unknown>;
	body_text: string;
	is_internal?: boolean;
	mentions?: string[];
}

/** Запрос на обновление комментария. */
export interface UpdateCommentRequest {
	body_json: Record<string, unknown>;
	body_text: string;
}

/** Запрос на логирование времени. */
export interface LogTimeRequest {
	minutes: number;
	description?: string;
}

/** Запрос на создание метки. */
export interface CreateLabelRequest {
	name: string;
	color: string;
	group_id?: string;
}

/** Запрос на создание группы меток. */
export interface CreateLabelGroupRequest {
	name: string;
}

/** Запрос на создание статуса workflow. */
export interface CreateWorkflowStatusRequest {
	name: string;
	category: "todo" | "in_progress" | "done";
	color?: string;
	sort_order?: number;
}

/** Запрос на создание перехода workflow. */
export interface CreateWorkflowTransitionRequest {
	from_status_id: string;
	to_status_id: string;
}

/** Запрос на создание доски. */
export interface CreateBoardRequest {
	name: string;
	filters?: Record<string, unknown>;
}

/** Запрос на создание кастомного поля. */
export interface CreateCustomFieldRequest {
	name: string;
	field_type: string;
	options?: string[];
	required?: boolean;
}

/** Запрос на добавление/удаление реакции. */
export interface CreateReactionRequest {
	emoji: string;
}

// ---------------------------------------------------------------------------
// API-функции: Задачи (Issues)
// ---------------------------------------------------------------------------

/** Создать новую задачу в проекте. */
export function createIssue(projectId: string, data: CreateIssueRequest) {
	return api.post<IssueResponse>(`/projects/${projectId}/issues`, data);
}

/** Получить список задач проекта с фильтрацией и пагинацией. */
export function listIssues(projectId: string, params?: ListIssuesParams) {
	const qs = buildQueryString(params);
	return api.get<PaginatedResponse<IssueResponse>>(
		`/projects/${projectId}/issues${qs}`,
	);
}

/** Получить детальную информацию о задаче по ключу. */
export function getIssue(projectId: string, issueKey: string) {
	return api.get<IssueResponse>(`/projects/${projectId}/issues/${issueKey}`);
}

/** Обновить задачу (с оптимистичной блокировкой по версии). */
export function updateIssue(
	projectId: string,
	issueKey: string,
	data: UpdateIssueRequest,
) {
	return api.put<IssueResponse>(
		`/projects/${projectId}/issues/${issueKey}`,
		data,
	);
}

/** Удалить задачу. */
export function deleteIssue(projectId: string, issueKey: string) {
	return api.delete<void>(`/projects/${projectId}/issues/${issueKey}`);
}

/** Переместить задачу на доске (изменить статус и позицию). */
export function moveIssue(
	projectId: string,
	issueKey: string,
	data: MoveIssueRequest,
) {
	return api.put<IssueResponse>(
		`/projects/${projectId}/issues/${issueKey}/position`,
		data,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Связи задач (Issue Links)
// ---------------------------------------------------------------------------

/** Создать связь между задачами. */
export function createIssueLink(
	projectId: string,
	issueKey: string,
	data: CreateIssueLinkRequest,
) {
	return api.post<IssueLinkResponse>(
		`/projects/${projectId}/issues/${issueKey}/links`,
		data,
	);
}

/** Получить список связей задачи. */
export function listIssueLinks(projectId: string, issueKey: string) {
	return api.get<IssueLinkResponse[]>(
		`/projects/${projectId}/issues/${issueKey}/links`,
	);
}

/** Удалить связь задачи. */
export function deleteIssueLink(
	projectId: string,
	issueKey: string,
	linkId: string,
) {
	return api.delete<void>(
		`/projects/${projectId}/issues/${issueKey}/links/${linkId}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Комментарии (Comments)
// ---------------------------------------------------------------------------

/** Создать комментарий к задаче. */
export function createComment(
	projectId: string,
	issueKey: string,
	data: CreateCommentRequest,
) {
	return api.post<CommentResponse>(
		`/projects/${projectId}/issues/${issueKey}/comments`,
		data,
	);
}

/** Получить список комментариев задачи. */
export function listComments(projectId: string, issueKey: string) {
	return api.get<CommentResponse[]>(
		`/projects/${projectId}/issues/${issueKey}/comments`,
	);
}

/** Обновить комментарий. */
export function updateComment(
	projectId: string,
	issueKey: string,
	commentId: string,
	data: UpdateCommentRequest,
) {
	return api.put<CommentResponse>(
		`/projects/${projectId}/issues/${issueKey}/comments/${commentId}`,
		data,
	);
}

/** Удалить комментарий. */
export function deleteComment(
	projectId: string,
	issueKey: string,
	commentId: string,
) {
	return api.delete<void>(
		`/projects/${projectId}/issues/${issueKey}/comments/${commentId}`,
	);
}

/** Переключить реакцию на комментарий (добавить/убрать). */
export function toggleReaction(
	projectId: string,
	issueKey: string,
	commentId: string,
	data: CreateReactionRequest,
) {
	return api.post<void>(
		`/projects/${projectId}/issues/${issueKey}/comments/${commentId}/reactions`,
		data,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Активность (Activity)
// ---------------------------------------------------------------------------

/** Получить лог активности по задаче. */
export function listActivities(projectId: string, issueKey: string) {
	return api.get<ActivityLogResponse[]>(
		`/projects/${projectId}/issues/${issueKey}/activity`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Таймтрекинг (Time Tracking)
// ---------------------------------------------------------------------------

/** Залогировать потраченное время на задачу. */
export function logTime(
	projectId: string,
	issueKey: string,
	data: LogTimeRequest,
) {
	return api.post<TimeLogResponse>(
		`/projects/${projectId}/issues/${issueKey}/time`,
		data,
	);
}

/** Запустить таймер на задаче. */
export function startTimer(projectId: string, issueKey: string) {
	return api.post<TimeLogResponse>(
		`/projects/${projectId}/issues/${issueKey}/timer/start`,
	);
}

/** Остановить таймер на задаче. */
export function stopTimer(projectId: string, issueKey: string) {
	return api.post<TimeLogResponse>(
		`/projects/${projectId}/issues/${issueKey}/timer/stop`,
	);
}

/** Получить список записей времени по задаче. */
export function listTimeLogs(projectId: string, issueKey: string) {
	return api.get<TimeLogResponse[]>(
		`/projects/${projectId}/issues/${issueKey}/time-logs`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Доски (Boards)
// ---------------------------------------------------------------------------

/** Получить список досок проекта. */
export function listBoards(projectId: string) {
	return api.get<BoardResponse[]>(`/projects/${projectId}/boards`);
}

/** Создать новую доску в проекте. */
export function createBoard(projectId: string, data: CreateBoardRequest) {
	return api.post<BoardResponse>(`/projects/${projectId}/boards`, data);
}

/** Получить полные данные доски (колонки + задачи). */
export function getBoardData(projectId: string, boardId: string) {
	return api.get<BoardDataResponse>(`/projects/${projectId}/boards/${boardId}`);
}

// ---------------------------------------------------------------------------
// API-функции: Workflow
// ---------------------------------------------------------------------------

/** Получить список статусов workflow проекта. */
export function listWorkflowStatuses(projectId: string) {
	return api.get<WorkflowStatusResponse[]>(
		`/projects/${projectId}/workflow/statuses`,
	);
}

/** Создать новый статус workflow. */
export function createWorkflowStatus(
	projectId: string,
	data: CreateWorkflowStatusRequest,
) {
	return api.post<WorkflowStatusResponse>(
		`/projects/${projectId}/workflow/statuses`,
		data,
	);
}

/** Получить список переходов workflow проекта. */
export function listWorkflowTransitions(projectId: string) {
	return api.get<WorkflowTransitionResponse[]>(
		`/projects/${projectId}/workflow/transitions`,
	);
}

/** Создать новый переход workflow. */
export function createWorkflowTransition(
	projectId: string,
	data: CreateWorkflowTransitionRequest,
) {
	return api.post<WorkflowTransitionResponse>(
		`/projects/${projectId}/workflow/transitions`,
		data,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Метки (Labels)
// ---------------------------------------------------------------------------

/** Получить список меток проекта. */
export function listLabels(projectId: string) {
	return api.get<LabelResponse[]>(`/projects/${projectId}/labels`);
}

/** Создать новую метку. */
export function createLabel(projectId: string, data: CreateLabelRequest) {
	return api.post<LabelResponse>(`/projects/${projectId}/labels`, data);
}

/** Удалить метку. */
export function deleteLabel(projectId: string, labelId: string) {
	return api.delete<void>(`/projects/${projectId}/labels/${labelId}`);
}

/** Получить список групп меток с метками. */
export function listLabelGroups(projectId: string) {
	return api.get<LabelGroupResponse[]>(`/projects/${projectId}/labels/groups`);
}

/** Создать новую группу меток. */
export function createLabelGroup(
	projectId: string,
	data: CreateLabelGroupRequest,
) {
	return api.post<LabelGroupResponse>(
		`/projects/${projectId}/labels/groups`,
		data,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Кастомные поля (Custom Fields)
// ---------------------------------------------------------------------------

/** Получить список определений кастомных полей проекта. */
export function listCustomFields(projectId: string) {
	return api.get<CustomFieldDefinitionResponse[]>(
		`/projects/${projectId}/custom-fields`,
	);
}

/** Создать новое определение кастомного поля. */
export function createCustomField(
	projectId: string,
	data: CreateCustomFieldRequest,
) {
	return api.post<CustomFieldDefinitionResponse>(
		`/projects/${projectId}/custom-fields`,
		data,
	);
}

/** Удалить определение кастомного поля. */
export function deleteCustomField(projectId: string, fieldId: string) {
	return api.delete<void>(`/projects/${projectId}/custom-fields/${fieldId}`);
}
