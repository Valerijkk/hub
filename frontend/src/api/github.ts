import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы ответов
// ---------------------------------------------------------------------------

/** Ответ с данными интеграции GitHub. */
export interface IntegrationResponse {
	id: string;
	repo_owner: string;
	repo_name: string;
	merge_status_id?: string;
	created_at: string;
}

/** Ответ с данными связи PR/коммита с задачей. */
export interface GitHubLinkResponse {
	id: string;
	link_type: "commit" | "pull_request";
	external_id: string;
	title: string;
	url: string;
	status: "open" | "merged" | "closed" | "none";
	author?: string;
	branch?: string;
	created_at: string;
	updated_at: string;
}

/** Ответ при создании ветки. */
export interface CreateBranchResponse {
	branch_name: string;
}

// ---------------------------------------------------------------------------
// Типы запросов
// ---------------------------------------------------------------------------

/** Запрос на создание интеграции с GitHub-репозиторием. */
export interface CreateIntegrationRequest {
	repo_owner: string;
	repo_name: string;
	access_token: string;
	webhook_secret: string;
	merge_status_id?: string;
}

/** Запрос на создание ветки из задачи. */
export interface CreateBranchRequest {
	integration_id: string;
	branch_name?: string;
}

// ---------------------------------------------------------------------------
// API-функции: Интеграции
// ---------------------------------------------------------------------------

/** Создать интеграцию с GitHub-репозиторием. */
export function createIntegration(
	projectId: string,
	data: CreateIntegrationRequest,
) {
	return api.post<IntegrationResponse>(
		`/projects/${projectId}/github/integrations`,
		data,
	);
}

/** Получить список интеграций проекта. */
export function listIntegrations(projectId: string) {
	return api.get<IntegrationResponse[]>(
		`/projects/${projectId}/github/integrations`,
	);
}

/** Удалить интеграцию. */
export function deleteIntegration(projectId: string, integrationId: string) {
	return api.delete<void>(
		`/projects/${projectId}/github/integrations/${integrationId}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: GitHub ссылки
// ---------------------------------------------------------------------------

/** Получить список GitHub-ссылок для задачи. */
export function listGitHubLinks(issueKey: string) {
	return api.get<GitHubLinkResponse[]>(`/issues/${issueKey}/github-links`);
}

/** Создать ветку из задачи (GHUB-05). */
export function createBranch(
	projectId: string,
	issueKey: string,
	data: CreateBranchRequest,
) {
	return api.post<CreateBranchResponse>(
		`/projects/${projectId}/issues/${issueKey}/github/create-branch`,
		data,
	);
}

// ---------------------------------------------------------------------------
// Типы: CI/CD Pipeline
// ---------------------------------------------------------------------------

/** Ответ с данными запуска workflow (CI/CD пайплайна). */
export interface WorkflowRunResponse {
	id: number;
	name: string;
	status: "queued" | "in_progress" | "completed";
	conclusion:
		| "success"
		| "failure"
		| "cancelled"
		| "skipped"
		| "timed_out"
		| "action_required"
		| "";
	branch: string;
	event: string;
	html_url: string;
	created_at: string;
	updated_at: string;
	head_sha: string;
	run_number: number;
}

/** Джоб внутри workflow run. */
export interface WorkflowJobResponse {
	id: number;
	name: string;
	status: string;
	conclusion: string;
	started_at: string;
	completed_at?: string;
}

/** Детальный ответ запуска workflow с джобами. */
export interface WorkflowRunDetailResponse extends WorkflowRunResponse {
	jobs: WorkflowJobResponse[];
}

// ---------------------------------------------------------------------------
// API-функции: CI/CD Pipeline
// ---------------------------------------------------------------------------

/** Получить последние запуски workflow для проекта. */
export function listWorkflowRuns(projectId: string) {
	return api.get<WorkflowRunResponse[]>(
		`/projects/${projectId}/github/workflows`,
	);
}

/** Получить детали запуска workflow. */
export function getWorkflowRunDetail(projectId: string, runId: number) {
	return api.get<WorkflowRunDetailResponse>(
		`/projects/${projectId}/github/workflows/${runId}`,
	);
}
