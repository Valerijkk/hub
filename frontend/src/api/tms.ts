import { buildQueryString } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";
import { api } from "./client";
import type { PaginatedResponse } from "./types";

// Реэкспорт для обратной совместимости с legacy-импортами PaginatedResponse.
export type { PaginatedResponse } from "./types";

// ---------------------------------------------------------------------------
// Типы перечислений TMS
// ---------------------------------------------------------------------------

/** Тип тест-кейса. */
export type TestCaseType = "classic" | "bdd";

/** Приоритет тест-кейса. */
export type TestCasePriority =
	| "critical"
	| "high"
	| "medium"
	| "low"
	| "trivial";

/** Статус тест-кейса. */
export type TestCaseStatus = "draft" | "active" | "deprecated";

/** Тип прогона. */
export type TestRunType = "manual" | "automatic" | "mixed";

/** Статус прогона. */
export type TestRunStatus = "created" | "in_progress" | "completed" | "aborted";

/** Статус результата теста. */
export type TestResultStatus =
	| "passed"
	| "failed"
	| "broken"
	| "skipped"
	| "unknown";

// ---------------------------------------------------------------------------
// Интерфейсы ответов
// ---------------------------------------------------------------------------

/** Краткая информация о пользователе. */
export interface UserBriefResponse {
	id: string;
	name: string;
	email: string;
	avatar_url?: string;
}

/** Краткое представление папки. */
export interface SuiteBriefResponse {
	id: string;
	name: string;
}

/** Метка. */
export interface LabelResponse {
	id: string;
	name: string;
	color: string;
}

/** Компонент. */
export interface ComponentResponse {
	id: string;
	name: string;
	description?: string;
	created_at: string;
}

/** Краткое представление задачи. */
export interface IssueBriefResponse {
	id: string;
	issue_key: string;
	title: string;
}

/** Шаг тест-кейса. */
export interface TestStep {
	order: number;
	action: string;
	expected_result?: string;
	keyword?: string;
	/**
	 * Стабильный client-only id для React key при редактировании списка шагов.
	 * Бэкенд игнорирует поле — перед отправкой удаляем через `stripStepUid`,
	 * чтобы не засорять payload лишними данными.
	 */
	_uid?: string;
}

/** Удаляет client-only поле `_uid` из шагов перед отправкой в API. */
export function stripStepUid(
	steps: TestStep[] | undefined,
): TestStep[] | undefined {
	if (!steps) return steps;
	return steps.map(({ _uid: _ignored, ...rest }) => rest);
}

/** Результат выполнения шага. */
export interface StepResult {
	step_order: number;
	status: string;
	comment?: string;
	file_ids?: string[];
}

/** Тест-кейс в ответе API. */
export interface TestCaseResponse {
	id: string;
	case_number: number;
	case_key: string;
	title: string;
	description?: string;
	precondition?: string;
	postcondition?: string;
	case_type: TestCaseType;
	priority: TestCasePriority;
	status: TestCaseStatus;
	steps?: TestStep[];
	full_name?: string;
	class_name?: string;
	method_name?: string;
	external_id?: string;
	automation_id?: string;
	version: number;
	suite?: SuiteBriefResponse;
	created_by?: UserBriefResponse;
	labels?: LabelResponse[];
	components?: ComponentResponse[];
	linked_issues?: IssueBriefResponse[];
	created_at: string;
	updated_at: string;
}

/** Версия тест-кейса. */
export interface TestCaseVersionResponse {
	id: string;
	version_number: number;
	snapshot: Record<string, unknown>;
	change_description?: string;
	created_by?: UserBriefResponse;
	created_at: string;
}

/** Папка/секция тест-кейсов. */
export interface SuiteResponse {
	id: string;
	name: string;
	description?: string;
	sort_order: number;
	parent_id?: string;
	children?: SuiteResponse[];
	created_at: string;
	updated_at: string;
}

/** Краткое представление файла. */
export interface FileBriefResponse {
	id: string;
	file_name: string;
	file_size: number;
	mime_type: string;
}

/** Краткое представление тест-кейса в результате. */
export interface TestCaseBriefDTO {
	id: string;
	case_key: string;
	title: string;
	case_type: TestCaseType;
	priority: TestCasePriority;
}

/** Прогон тестов. */
export interface TestRunResponse {
	id: string;
	title: string;
	description?: string;
	run_type: TestRunType;
	status: TestRunStatus;
	total_count: number;
	passed_count: number;
	failed_count: number;
	skipped_count: number;
	broken_count: number;
	unknown_count: number;
	allure_run_id?: string;
	created_by?: UserBriefResponse;
	started_at?: string;
	completed_at?: string;
	created_at: string;
	updated_at: string;
}

/** Результат теста. */
export interface TestResultResponse {
	id: string;
	status: TestResultStatus;
	comment?: string;
	step_results?: StepResult[];
	duration_ms?: number;
	allure_uuid?: string;
	error_message?: string;
	error_trace?: string;
	is_flaky: boolean;
	match_strategy: string;
	test_case_key?: string;
	executed_by?: UserBriefResponse;
	started_at?: string;
	completed_at?: string;
	created_at: string;
}

/** Детальный результат теста (с информацией о кейсе и файлах). */
export interface TestResultDetailResponse {
	id: string;
	status: TestResultStatus;
	comment?: string;
	step_results?: StepResult[];
	duration_ms?: number;
	is_flaky: boolean;
	test_case?: TestCaseBriefDTO;
	executed_by?: UserBriefResponse;
	files?: FileBriefResponse[];
	started_at?: string;
	completed_at?: string;
	created_at: string;
}

/** Детали прогона с результатами и назначенными. */
export interface TestRunDetailResponse extends TestRunResponse {
	results: TestResultDetailResponse[];
	assignees: UserBriefResponse[];
}

// ---------------------------------------------------------------------------
// Интерфейсы запросов
// ---------------------------------------------------------------------------

/** Создание тест-кейса. */
export interface CreateTestCaseRequest {
	title: string;
	description?: string;
	precondition?: string;
	postcondition?: string;
	case_type?: TestCaseType;
	priority?: TestCasePriority;
	status?: TestCaseStatus;
	steps?: TestStep[];
	automation_id?: string;
	suite_id?: string;
	labels?: string[];
	components?: string[];
}

/** Обновление тест-кейса. */
export interface UpdateTestCaseRequest {
	title?: string;
	description?: string;
	precondition?: string;
	postcondition?: string;
	case_type?: TestCaseType;
	priority?: TestCasePriority;
	status?: TestCaseStatus;
	steps?: TestStep[];
	automation_id?: string;
	suite_id?: string;
	labels?: string[];
	components?: string[];
	version: number;
}

/** Параметры списка тест-кейсов. */
export interface ListTestCasesParams {
	offset?: number;
	limit?: number;
	suite_id?: string;
	status?: TestCaseStatus;
	priority?: TestCasePriority;
	search?: string;
}

/** Создание папки. */
export interface CreateSuiteRequest {
	name: string;
	description?: string;
	sort_order?: number;
	parent_id?: string;
}

/** Обновление папки. */
export interface UpdateSuiteRequest {
	name?: string;
	description?: string;
	sort_order?: number;
}

/** Создание компонента. */
export interface CreateComponentRequest {
	name: string;
	description?: string;
}

/** Создание прогона. */
export interface CreateTestRunRequest {
	title: string;
	description?: string;
	run_type?: TestRunType;
}

/** Параметры списка прогонов. */
export interface ListTestRunsParams {
	offset?: number;
	limit?: number;
	status?: TestRunStatus;
}

/** Обновление результата теста. */
export interface UpdateTestResultRequest {
	status?: TestResultStatus;
	comment?: string;
	step_results?: StepResult[];
	is_flaky?: boolean;
	file_ids?: string[];
}

// ---------------------------------------------------------------------------
// API-функции: Тест-кейсы
// ---------------------------------------------------------------------------

/** Создать тест-кейс в проекте. */
export function createTestCase(projectId: string, data: CreateTestCaseRequest) {
	return api.post<TestCaseResponse>(`/projects/${projectId}/tms/cases`, data);
}

/** Получить список тест-кейсов проекта. */
export function listTestCases(projectId: string, params?: ListTestCasesParams) {
	const qs = buildQueryString(params);
	return api.get<PaginatedResponse<TestCaseResponse>>(
		`/projects/${projectId}/tms/cases${qs}`,
	);
}

/** Получить тест-кейс по ключу. */
export function getTestCase(projectId: string, caseKey: string) {
	return api.get<TestCaseResponse>(
		`/projects/${projectId}/tms/cases/${caseKey}`,
	);
}

/** Обновить тест-кейс. */
export function updateTestCase(
	projectId: string,
	caseKey: string,
	data: UpdateTestCaseRequest,
) {
	return api.put<TestCaseResponse>(
		`/projects/${projectId}/tms/cases/${caseKey}`,
		data,
	);
}

/** Удалить тест-кейс. */
export function deleteTestCase(projectId: string, caseKey: string) {
	return api.delete<void>(`/projects/${projectId}/tms/cases/${caseKey}`);
}

/** Получить список версий тест-кейса. */
export function listVersions(projectId: string, caseKey: string) {
	return api.get<TestCaseVersionResponse[]>(
		`/projects/${projectId}/tms/cases/${caseKey}/versions`,
	);
}

/** Откатить тест-кейс к указанной версии. */
export function rollbackVersion(
	projectId: string,
	caseKey: string,
	versionId: string,
) {
	return api.post<TestCaseResponse>(
		`/projects/${projectId}/tms/cases/${caseKey}/rollback/${versionId}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Папки/сьюты
// ---------------------------------------------------------------------------

/** Создать папку тест-кейсов. */
export function createSuite(projectId: string, data: CreateSuiteRequest) {
	return api.post<SuiteResponse>(`/projects/${projectId}/tms/suites`, data);
}

/** Получить дерево папок. */
export function listSuites(projectId: string) {
	return api.get<SuiteResponse[]>(`/projects/${projectId}/tms/suites`);
}

/** Обновить папку. */
export function updateSuite(
	projectId: string,
	suiteId: string,
	data: UpdateSuiteRequest,
) {
	return api.put<SuiteResponse>(
		`/projects/${projectId}/tms/suites/${suiteId}`,
		data,
	);
}

/** Удалить папку. */
export function deleteSuite(projectId: string, suiteId: string) {
	return api.delete<void>(`/projects/${projectId}/tms/suites/${suiteId}`);
}

// ---------------------------------------------------------------------------
// API-функции: Компоненты
// ---------------------------------------------------------------------------

/** Создать компонент. */
export function createComponent(
	projectId: string,
	data: CreateComponentRequest,
) {
	return api.post<ComponentResponse>(
		`/projects/${projectId}/tms/components`,
		data,
	);
}

/** Получить список компонентов. */
export function listComponents(projectId: string) {
	return api.get<ComponentResponse[]>(`/projects/${projectId}/tms/components`);
}

/** Удалить компонент. */
export function deleteComponent(projectId: string, componentId: string) {
	return api.delete<void>(
		`/projects/${projectId}/tms/components/${componentId}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Прогоны
// ---------------------------------------------------------------------------

/** Создать прогон тестов. */
export function createTestRun(projectId: string, data: CreateTestRunRequest) {
	return api.post<TestRunResponse>(`/projects/${projectId}/tms/runs`, data);
}

/** Получить список прогонов. */
export function listTestRuns(projectId: string, params?: ListTestRunsParams) {
	const qs = buildQueryString(params);
	return api.get<PaginatedResponse<TestRunResponse>>(
		`/projects/${projectId}/tms/runs${qs}`,
	);
}

/** Получить детали прогона с результатами и назначенными. */
export function getTestRunDetail(projectId: string, runId: string) {
	return api.get<TestRunDetailResponse>(
		`/projects/${projectId}/tms/runs/${runId}/detail`,
	);
}

/** Запустить прогон (created -> in_progress). */
export function startRun(projectId: string, runId: string) {
	return api.post<TestRunResponse>(
		`/projects/${projectId}/tms/runs/${runId}/start`,
	);
}

/** Завершить прогон (in_progress -> completed). */
export function completeRun(projectId: string, runId: string) {
	return api.post<TestRunResponse>(
		`/projects/${projectId}/tms/runs/${runId}/complete`,
	);
}

/** Переоткрыть прогон (completed -> in_progress). */
export function reopenRun(projectId: string, runId: string) {
	return api.post<TestRunResponse>(
		`/projects/${projectId}/tms/runs/${runId}/reopen`,
	);
}

/** Добавить тест-кейсы в прогон. */
export function addCasesToRun(
	projectId: string,
	runId: string,
	caseIds: string[],
) {
	return api.post<TestRunDetailResponse>(
		`/projects/${projectId}/tms/runs/${runId}/cases`,
		{ case_ids: caseIds },
	);
}

/** Назначить тестировщиков на прогон. */
export function assignTesters(
	projectId: string,
	runId: string,
	userIds: string[],
) {
	return api.post<TestRunDetailResponse>(
		`/projects/${projectId}/tms/runs/${runId}/assignees`,
		{ user_ids: userIds },
	);
}

// ---------------------------------------------------------------------------
// API-функции: Результаты
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Интерфейсы: Отчёты и покрытие
// ---------------------------------------------------------------------------

/** Покрытие тест-кейса. */
export interface TestCaseCoverage {
	case_key: string;
	case_title: string;
	last_status: string;
}

/** Покрытие задачи тест-кейсами. */
export interface IssueCoverage {
	issue_id: string;
	issue_key: string;
	issue_title: string;
	test_cases: TestCaseCoverage[];
}

/** Отчёт о покрытии. */
export interface CoverageReport {
	issues: IssueCoverage[];
}

/** Точка тренда. */
export interface TrendPoint {
	run_id: string;
	title: string;
	total: number;
	passed: number;
	failed: number;
	skipped: number;
	broken: number;
	pass_rate: number;
	created_at: string;
}

/** Изменение статуса при сравнении прогонов. */
export interface ResultChange {
	case_key: string;
	case_title: string;
	status_run1: string;
	status_run2: string;
}

/** Краткая информация о прогоне. */
export interface TestRunSummary {
	run_id: string;
	title: string;
	total: number;
	passed: number;
	failed: number;
	pass_rate: number;
	created_at: string;
}

/** Сравнение двух прогонов. */
export interface RunComparison {
	run1: TestRunSummary;
	run2: TestRunSummary;
	changes: ResultChange[];
}

/** Результат импорта. */
export interface ImportResult {
	created: number;
	skipped: number;
	errors: number;
	validation_errors?: { row: number; field: string; message: string }[];
}

/** Запрос создания бага. */
export interface CreateBugRequest {
	title?: string;
	description?: string;
	priority?: string;
}

// ---------------------------------------------------------------------------
// API-функции: Привязка к задачам
// ---------------------------------------------------------------------------

/** Привязать тест-кейс к задаче. */
export function linkTestCaseToIssue(
	projectId: string,
	caseKey: string,
	issueId: string,
) {
	return api.post<void>(`/projects/${projectId}/tms/cases/${caseKey}/links`, {
		issue_id: issueId,
	});
}

/** Отвязать тест-кейс от задачи. */
export function unlinkTestCaseFromIssue(
	projectId: string,
	caseKey: string,
	issueId: string,
) {
	return api.delete<void>(
		`/projects/${projectId}/tms/cases/${caseKey}/links/${issueId}`,
	);
}

/** Получить привязанные задачи тест-кейса. */
export function getTestCaseLinks(projectId: string, caseKey: string) {
	return api.get<IssueBriefResponse[]>(
		`/projects/${projectId}/tms/cases/${caseKey}/links`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Создание бага из результата
// ---------------------------------------------------------------------------

/** Создать баг из упавшего результата теста. */
export function createBugFromResult(
	projectId: string,
	runId: string,
	resultId: string,
	data?: CreateBugRequest,
) {
	return api.post<IssueBriefResponse>(
		`/projects/${projectId}/tms/runs/${runId}/results/${resultId}/bug`,
		data ?? {},
	);
}

// ---------------------------------------------------------------------------
// API-функции: Отчёты
// ---------------------------------------------------------------------------

/** Получить отчёт о покрытии задач тестами. */
export function getCoverageReport(projectId: string) {
	return api.get<CoverageReport>(`/projects/${projectId}/tms/reports/coverage`);
}

/** Получить тренды прогонов. */
export function getTestTrends(projectId: string, limit?: number) {
	const qs = limit ? `?limit=${limit}` : "";
	return api.get<TrendPoint[]>(
		`/projects/${projectId}/tms/reports/trends${qs}`,
	);
}

/** Сравнить два прогона. */
export function getRunComparison(
	projectId: string,
	run1: string,
	run2: string,
) {
	return api.get<RunComparison>(
		`/projects/${projectId}/tms/reports/comparison?run1=${run1}&run2=${run2}`,
	);
}

// ---------------------------------------------------------------------------
// API-функции: Импорт/экспорт
// ---------------------------------------------------------------------------

/** Импортировать тест-кейсы из файла (multipart). */
export async function importTestCases(
	projectId: string,
	file: File,
	format?: string,
	mapping?: Record<string, string>,
) {
	const formData = new FormData();
	formData.append("file", file);
	if (format) formData.append("format", format);
	if (mapping) formData.append("mapping", JSON.stringify(mapping));

	// Не через api.post: client.ts делает JSON.stringify(FormData) + application/json.
	const headers: Record<string, string> = {};
	const token = useAuthStore.getState().accessToken;
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	const response = await fetch(`/api/projects/${projectId}/tms/import`, {
		method: "POST",
		headers,
		body: formData,
	});
	if (!response.ok) {
		let message = response.statusText;
		try {
			const err = (await response.json()) as { message?: string; error?: string };
			message = err.message || err.error || message;
		} catch {
			/* ignore */
		}
		throw { code: response.status, message };
	}
	if (response.status === 204) return undefined as unknown as ImportResult;
	return (await response.json()) as ImportResult;
}

/** Экспортировать тест-кейсы (возвращает blob URL). */
export function exportTestCases(projectId: string, format: string = "csv") {
	return api.get<Blob>(`/projects/${projectId}/tms/export?format=${format}`);
}

/** Экспортировать отчёт прогона (возвращает blob URL). */
export function exportTestRunReport(
	projectId: string,
	runId: string,
	format: string = "csv",
) {
	return api.get<Blob>(
		`/projects/${projectId}/tms/runs/${runId}/export?format=${format}`,
	);
}

// ---------------------------------------------------------------------------
// Интерфейсы: Матрица трассировки
// ---------------------------------------------------------------------------

/** Ячейка матрицы трассировки. */
export interface TraceabilityCell {
	case_key: string;
	case_title: string;
	last_status: string;
}

/** Строка матрицы трассировки (задача + привязанные тест-кейсы). */
export interface TraceabilityRow {
	issue_key: string;
	issue_title: string;
	test_cases: TraceabilityCell[];
}

/** Полная матрица трассировки. */
export interface TraceabilityMatrix {
	rows: TraceabilityRow[];
}

/** Элемент отчёта по flaky-тестам. */
export interface FlakyTestItem extends TestCaseResponse {
	status_history: string[];
	flake_rate: number;
	marked_flaky: boolean;
}

/** Отчёт flaky-тестов за последние N прогонов. */
export interface FlakyTestReport {
	items: FlakyTestItem[];
	run_window: number;
}

/** Получить отчёт по flaky-тестам. */
export function getFlakyTests(projectId: string) {
	return api.get<FlakyTestReport>(`/projects/${projectId}/tms/reports/flaky`);
}

/** Получить матрицу трассировки проекта. */
export function getTraceabilityMatrix(projectId: string) {
	return api.get<TraceabilityMatrix>(
		`/projects/${projectId}/tms/reports/traceability`,
	);
}

// ---------------------------------------------------------------------------
// Интерфейсы: Окружения
// ---------------------------------------------------------------------------

/** Тестовое окружение. */
export interface EnvironmentResponse {
	id: string;
	name: string;
	label: string;
	config?: Record<string, unknown>;
	created_at: string;
}

/** Получить список окружений проекта. */
export function listEnvironments(projectId: string) {
	return api.get<EnvironmentResponse[]>(`/projects/${projectId}/environments`);
}

// ---------------------------------------------------------------------------
// Интерфейсы: Сравнение окружений
// ---------------------------------------------------------------------------

/** Элемент сравнения результатов по окружениям. */
export interface EnvironmentComparisonItem {
	case_key: string;
	case_title: string;
	status_env1: string;
	status_env2: string;
}

/** Ответ сравнения окружений. */
export interface EnvironmentComparisonResponse {
	env1: EnvironmentResponse;
	env2: EnvironmentResponse;
	items: EnvironmentComparisonItem[];
}

/** Сравнить результаты двух окружений. */
export function compareEnvironments(
	projectId: string,
	envId1: string,
	envId2: string,
) {
	return api.get<EnvironmentComparisonResponse>(
		`/projects/${projectId}/tms/reports/env-comparison?env1=${envId1}&env2=${envId2}`,
	);
}

/** Обновить результат теста (выполнение с данными шагов). */
export function updateTestResult(
	projectId: string,
	runId: string,
	resultId: string,
	data: UpdateTestResultRequest,
) {
	return api.put<TestResultResponse>(
		`/projects/${projectId}/tms/runs/${runId}/results/${resultId}`,
		data,
	);
}
