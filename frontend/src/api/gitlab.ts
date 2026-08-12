import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы ответов
// ---------------------------------------------------------------------------

/** Ответ с данными интеграции GitLab. */
export interface GitLabIntegrationResponse {
	id: string;
	gitlab_url: string;
	project_path: string;
	created_at: string;
}

/** Ответ с данными пайплайна GitLab CI/CD. */
export interface GitLabPipelineResponse {
	id: number;
	status:
		| "created"
		| "waiting_for_resource"
		| "preparing"
		| "pending"
		| "running"
		| "success"
		| "failed"
		| "canceled"
		| "skipped"
		| "manual"
		| "scheduled";
	ref: string;
	sha: string;
	web_url: string;
	source: string;
	created_at: string;
	updated_at: string;
}

/** Ответ при создании ветки. */
export interface GitLabCreateBranchResponse {
	branch_name: string;
}

// ---------------------------------------------------------------------------
// Типы запросов
// ---------------------------------------------------------------------------

/** Запрос на подключение GitLab-репозитория. */
export interface ConnectGitLabRequest {
	gitlab_url?: string;
	project_path: string;
	access_token: string;
	webhook_secret?: string;
}

/** Запрос на создание ветки из задачи. */
export interface GitLabCreateBranchRequest {
	integration_id: string;
	branch_name?: string;
	ref?: string;
}

// ---------------------------------------------------------------------------
// API-функции: Интеграции
// ---------------------------------------------------------------------------

/** Подключить GitLab-репозиторий к проекту. */
export function connectGitLab(projectId: string, data: ConnectGitLabRequest) {
	return api.post<GitLabIntegrationResponse>(
		`/projects/${projectId}/gitlab/integrations`,
		data,
	);
}

/** Получить список GitLab-интеграций проекта. */
export function listGitLabIntegrations(projectId: string) {
	return api.get<GitLabIntegrationResponse[]>(
		`/projects/${projectId}/gitlab/integrations`,
	);
}

/** Отключить GitLab-интеграцию. */
export function disconnectGitLab(projectId: string, integrationId: string) {
	return api.delete<void>(
		`/projects/${projectId}/gitlab/integrations/${integrationId}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Пайплайны
// ---------------------------------------------------------------------------

/** Получить последние пайплайны GitLab CI/CD для проекта. */
export function listGitLabPipelines(projectId: string) {
	return api.get<GitLabPipelineResponse[]>(
		`/projects/${projectId}/gitlab/pipelines`,
	);
}

/** Запрос на запуск пайплайна GitLab CI/CD. */
export interface TriggerPipelineRequest {
	integration_id: string;
	ref?: string;
	variables?: Record<string, string>;
}

/** Ответ при запуске пайплайна. */
export interface TriggerPipelineResponse {
	pipeline_id: number;
	web_url: string;
	status: string;
	ref: string;
}

/** Запустить пайплайн GitLab CI/CD с переменными. */
export function triggerPipeline(
	projectId: string,
	data: TriggerPipelineRequest,
) {
	return api.post<TriggerPipelineResponse>(
		`/projects/${projectId}/gitlab/pipelines/trigger`,
		data,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Ветки
// ---------------------------------------------------------------------------

/** Создать ветку в GitLab из задачи. */
export function createGitLabBranch(
	projectId: string,
	issueKey: string,
	data: GitLabCreateBranchRequest,
) {
	return api.post<GitLabCreateBranchResponse>(
		`/projects/${projectId}/issues/${issueKey}/gitlab/create-branch`,
		data,
	);
}
