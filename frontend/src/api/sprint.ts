import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

/** Статус спринта. */
export type SprintStatus = "planning" | "active" | "completed";

/** Ответ спринта. */
export interface SprintResponse {
  id: string;
  name: string;
  goal?: string;
  status: SprintStatus;
  start_date: string;
  end_date: string;
  retro_notes?: Record<string, unknown>;
  issue_count: number;
  created_at: string;
  updated_at: string;
}

/** Точка данных burndown-графика. */
export interface BurndownPoint {
  date: string;
  remaining: number;
  ideal: number;
}

/** Точка данных velocity-графика. */
export interface VelocityPoint {
  sprint_name: string;
  committed: number;
  completed: number;
}

/** Сессия Planning Poker. */
export interface PokerSessionResponse {
  session_id: string;
  issue_id: string;
  revealed: boolean;
  votes?: Record<string, string>;
  voters?: string[];
}

// ---------------------------------------------------------------------------
// Запросы
// ---------------------------------------------------------------------------

/** Запрос на создание спринта. */
export interface CreateSprintRequest {
  name: string;
  goal?: string;
  start_date: string;
  end_date: string;
}

/** Запрос на обновление спринта. */
export interface UpdateSprintRequest {
  name?: string;
  goal?: string;
  start_date?: string;
  end_date?: string;
}

/** Заметки ретроспективы. */
export interface RetroNotesRequest {
  went_well: string[];
  to_improve: string[];
  actions: string[];
}

/** Запрос на голосование в Planning Poker. */
export interface VoteRequest {
  value: string;
}

/** Запрос на запуск сессии Planning Poker. */
export interface StartPokerRequest {
  issue_id: string;
}

/** Запрос на добавление задач в спринт. */
export interface MoveIssueToSprintRequest {
  issue_ids: string[];
}

// ---------------------------------------------------------------------------
// API-функции: Спринты
// ---------------------------------------------------------------------------

/** Создать новый спринт. */
export function createSprint(projectId: string, data: CreateSprintRequest) {
  return api.post<SprintResponse>(`/projects/${projectId}/sprints`, data);
}

/** Получить список спринтов проекта. */
export function listSprints(projectId: string) {
  return api.get<SprintResponse[]>(`/projects/${projectId}/sprints`);
}

/** Получить спринт по ID. */
export function getSprint(projectId: string, sprintId: string) {
  return api.get<SprintResponse>(
    `/projects/${projectId}/sprints/${sprintId}`,
  );
}

/** Обновить спринт. */
export function updateSprint(
  projectId: string,
  sprintId: string,
  data: UpdateSprintRequest,
) {
  return api.put<SprintResponse>(
    `/projects/${projectId}/sprints/${sprintId}`,
    data,
  );
}

/** Запустить спринт. */
export function startSprint(projectId: string, sprintId: string) {
  return api.post<SprintResponse>(
    `/projects/${projectId}/sprints/${sprintId}/start`,
  );
}

/** Завершить спринт. */
export function completeSprint(projectId: string, sprintId: string) {
  return api.post<SprintResponse>(
    `/projects/${projectId}/sprints/${sprintId}/complete`,
  );
}

/** Обновить заметки ретроспективы. */
export function updateRetroNotes(
  projectId: string,
  sprintId: string,
  data: RetroNotesRequest,
) {
  return api.put<SprintResponse>(
    `/projects/${projectId}/sprints/${sprintId}/retro`,
    data,
  );
}

/** Получить данные burndown-графика. */
export function getBurndown(projectId: string, sprintId: string) {
  return api.get<BurndownPoint[]>(
    `/projects/${projectId}/sprints/${sprintId}/burndown`,
  );
}

/** Получить данные velocity-графика проекта. */
export function getVelocity(projectId: string) {
  return api.get<VelocityPoint[]>(`/projects/${projectId}/velocity`);
}

/** Добавить задачи в спринт. */
export function addIssuesToSprint(
  projectId: string,
  sprintId: string,
  data: MoveIssueToSprintRequest,
) {
  return api.post<void>(
    `/projects/${projectId}/sprints/${sprintId}/issues`,
    data,
  );
}

// ---------------------------------------------------------------------------
// API-функции: Planning Poker
// ---------------------------------------------------------------------------

/** Начать сессию Planning Poker. */
export function startPokerSession(
  projectId: string,
  sprintId: string,
  data: StartPokerRequest,
) {
  return api.post<PokerSessionResponse>(
    `/projects/${projectId}/sprints/${sprintId}/poker`,
    data,
  );
}

/** Отдать голос в Planning Poker. */
export function castVote(sessionId: string, data: VoteRequest) {
  return api.post<void>(`/sprints/poker/${sessionId}/vote`, data);
}

/** Раскрыть голоса. */
export function revealVotes(sessionId: string) {
  return api.post<PokerSessionResponse>(
    `/sprints/poker/${sessionId}/reveal`,
  );
}

/** Получить состояние сессии Planning Poker. */
export function getPokerSession(sessionId: string) {
  return api.get<PokerSessionResponse>(`/sprints/poker/${sessionId}`);
}
