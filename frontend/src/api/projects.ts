import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы запросов (Request DTOs)
// ---------------------------------------------------------------------------

/** Запрос на создание проекта. */
interface CreateProjectRequest {
	name: string;
	key: string;
	description?: string;
}

/** Запрос на обновление проекта. */
interface UpdateProjectRequest {
	name?: string;
	description?: string;
	avatar_url?: string;
}

/** Запрос на добавление участника в проект. */
interface AddMemberRequest {
	user_id: string;
	role: "developer" | "tester" | "viewer";
}

/** Запрос на создание инвайт-ссылки. */
interface CreateInviteRequest {
	role: "developer" | "tester" | "viewer";
	max_uses?: number;
	expires_at?: string;
}

// ---------------------------------------------------------------------------
// Типы ответов (Response DTOs)
// ---------------------------------------------------------------------------

/** Ответ с данными проекта. */
export interface ProjectResponse {
	id: string;
	name: string;
	key: string;
	description?: string;
	avatar_url?: string;
	status: string;
	member_count: number;
	my_role?: string;
	created_at: string;
	updated_at: string;
}

/** Ответ с данными участника проекта. */
export interface MemberResponse {
	user_id: string;
	email: string;
	name: string;
	avatar_url?: string;
	role: string;
	joined_at: string;
}

/** Ответ с данными инвайт-ссылки. */
export interface InviteResponse {
	id: string;
	code: string;
	role: string;
	max_uses: number;
	use_count: number;
	expires_at?: string;
	link: string;
	created_at: string;
}

// ---------------------------------------------------------------------------
// API-функции: Проекты
// ---------------------------------------------------------------------------

/** Endpoint: GET /projects — получить список проектов пользователя. */
export function listProjects() {
	return api.get<ProjectResponse[]>("/projects");
}

/** Endpoint: POST /projects — создать новый проект. */
export function createProject(data: CreateProjectRequest) {
	return api.post<ProjectResponse>("/projects", data);
}

/** Endpoint: GET /projects/:id — получить проект по ID. */
export function getProject(projectId: string) {
	return api.get<ProjectResponse>(`/projects/${projectId}`);
}

/** Endpoint: PUT /projects/:id — обновить проект. */
export function updateProject(projectId: string, data: UpdateProjectRequest) {
	return api.put<ProjectResponse>(`/projects/${projectId}`, data);
}

/** Endpoint: POST /projects/:id/archive — архивировать проект. */
export function archiveProject(projectId: string) {
	return api.post<void>(`/projects/${projectId}/archive`);
}

/** Endpoint: DELETE /projects/:id — удалить проект. */
export function deleteProject(projectId: string) {
	return api.delete<void>(`/projects/${projectId}`);
}

/** Endpoint: POST /projects/:id/restore — восстановить архивированный проект. */
export function restoreProject(projectId: string) {
	return api.post<void>(`/projects/${projectId}/restore`);
}

/** Endpoint: GET /projects/:id/members — получить список участников проекта. */
export function listMembers(projectId: string) {
	return api.get<MemberResponse[]>(`/projects/${projectId}/members`);
}

/** Endpoint: POST /projects/:id/members — добавить участника в проект. */
export function addMember(projectId: string, data: AddMemberRequest) {
	return api.post<MemberResponse>(`/projects/${projectId}/members`, data);
}

/** Endpoint: DELETE /projects/:id/members/:userId — удалить участника из проекта. */
export function removeMember(projectId: string, userId: string) {
	return api.delete<void>(`/projects/${projectId}/members/${userId}`);
}

/** Endpoint: POST /projects/:id/invites — создать инвайт-ссылку для проекта. */
export function createInvite(projectId: string, data: CreateInviteRequest) {
	return api.post<InviteResponse>(`/projects/${projectId}/invites`, data);
}

/** Endpoint: POST /invites/join — присоединиться к проекту по инвайт-коду. */
export function joinByInvite(code: string) {
	return api.post<void>("/invites/join", { code });
}
