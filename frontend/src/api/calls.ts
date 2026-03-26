import { api } from "./client";

// ---------------------------------------------------------------------------
// Типы ответов (Response DTOs)
// ---------------------------------------------------------------------------

/** Статус звонка. */
export type CallStatus = "active" | "ended";

/** Ответ API со звонком. */
export interface CallResponse {
	/** ID звонка. */
	id: string;
	/** Имя комнаты LiveKit. */
	room_name: string;
	/** ID канала, к которому привязан звонок. */
	channel_id: string;
	/** Статус звонка. */
	status: CallStatus;
	/** ID создателя звонка. */
	created_by: string;
	/** Количество участников. */
	participant_count: number;
	/** JWT-токен для подключения к LiveKit (присутствует при create/join). */
	token?: string;
	/** Дата создания. */
	created_at: string;
	/** Дата обновления. */
	updated_at: string;
}

/** Участник звонка. */
export interface CallParticipantResponse {
	/** ID записи участника. */
	id: string;
	/** ID звонка. */
	call_id: string;
	/** ID пользователя. */
	user_id: string;
	/** Имя пользователя. */
	user_name: string;
	/** Дата присоединения. */
	joined_at: string;
	/** Дата отключения (null если ещё в звонке). */
	left_at?: string;
}

// ---------------------------------------------------------------------------
// API-функции: Звонки (Calls)
// ---------------------------------------------------------------------------

/**
 * Создать новый звонок в канале.
 * @param channelId - ID канала
 * @returns Данные звонка с токеном для LiveKit
 */
export function createCall(channelId: string) {
	return api.post<CallResponse>("/calls", { channel_id: channelId });
}

/**
 * Присоединиться к существующему звонку.
 * @param callId - ID звонка
 * @returns Данные звонка с токеном для LiveKit
 */
export function joinCall(callId: string) {
	return api.post<CallResponse>(`/calls/${callId}/join`);
}

/**
 * Покинуть звонок.
 * @param callId - ID звонка
 */
export function leaveCall(callId: string) {
	return api.post<void>(`/calls/${callId}/leave`);
}

/**
 * Завершить звонок (только создатель).
 * @param callId - ID звонка
 */
export function endCall(callId: string) {
	return api.post<void>(`/calls/${callId}/end`);
}

/**
 * Получить активный звонок в канале.
 * Возвращает null если активного звонка нет (404).
 * @param channelId - ID канала
 */
export async function getActiveCall(
	channelId: string,
): Promise<CallResponse | null> {
	try {
		// Бэкенд возвращает {"call": null} когда нет активного звонка,
		// или {"id": "...", ...} когда звонок активен
		const raw = await api.get<CallResponse | { call: null }>(
			`/calls/channel/${channelId}`,
		);
		if (!raw || ("call" in raw && raw.call === null)) return null;
		if (!("id" in raw) || !raw.id) return null;
		return raw as CallResponse;
	} catch (err: unknown) {
		const apiErr = err as { code?: number };
		if (apiErr.code === 404) return null;
		return null;
	}
}

/**
 * Поднять или опустить руку в звонке.
 * @param callId - ID звонка
 * @param raised - true = поднять, false = опустить
 */
export function raiseHand(callId: string, raised: boolean) {
	return api.post<void>(`/calls/${callId}/raise-hand`, { raised });
}

// ---------------------------------------------------------------------------
// Эфемерный чат звонка
// ---------------------------------------------------------------------------

/** Сообщение чата звонка. */
export interface CallChatMessage {
	/** ID сообщения. */
	id: string;
	/** ID звонка. */
	call_id: string;
	/** ID отправителя. */
	user_id: string;
	/** Имя отправителя. */
	user_name: string;
	/** Текст сообщения. */
	text: string;
	/** Дата отправки. */
	created_at: string;
}

/**
 * Отправить сообщение в чат звонка.
 * @param callId - ID звонка
 * @param text - текст сообщения
 */
export function sendChatMessage(callId: string, text: string) {
	return api.post<CallChatMessage>(`/calls/${callId}/chat`, { text });
}

/**
 * Получить историю чата звонка.
 * @param callId - ID звонка
 */
export async function listChatMessages(
	callId: string,
): Promise<CallChatMessage[]> {
	const res = await api.get<{ messages: CallChatMessage[] }>(
		`/calls/${callId}/chat`,
	);
	return res.messages ?? [];
}

// ---------------------------------------------------------------------------
// Запись звонков (CALL-04, CALL-05)
// ---------------------------------------------------------------------------

/** Статус записи звонка. */
export type RecordingStatus = "recording" | "processing" | "ready" | "failed";

/** Запись звонка. */
export interface RecordingResponse {
	/** ID записи. */
	id: string;
	/** ID egress-процесса LiveKit. */
	egress_id: string;
	/** Статус записи. */
	status: RecordingStatus;
	/** Путь файла в S3. */
	file_path?: string;
	/** Размер файла в байтах. */
	file_size: number;
	/** Длительность в секундах. */
	duration_seconds: number;
	/** Время начала записи. */
	started_at: string;
	/** Время окончания записи. */
	ended_at?: string;
	/** ID звонка. */
	call_id: string;
	/** ID проекта. */
	project_id: string;
	/** ID пользователя, начавшего запись. */
	started_by: string;
	/** Дата создания. */
	created_at: string;
}

/** Ответ со ссылкой на скачивание записи. */
export interface RecordingDownloadResponse {
	/** Presigned URL для скачивания. */
	url: string;
	/** Время истечения URL. */
	expires_at: string;
}

/**
 * Начать запись звонка.
 * @param callId - ID звонка
 */
export function startRecording(callId: string) {
	return api.post<RecordingResponse>(`/calls/${callId}/recording/start`);
}

/**
 * Остановить запись звонка.
 * @param callId - ID звонка
 * @param recordingId - ID записи
 */
export function stopRecording(callId: string, recordingId: string) {
	return api.post<void>(`/calls/${callId}/recording/stop`, {
		recording_id: recordingId,
	});
}

/**
 * Получить список записей проекта.
 * @param projectId - ID проекта
 * @param offset - смещение для пагинации
 * @param limit - лимит записей
 */
export function listProjectRecordings(
	projectId: string,
	offset = 0,
	limit = 20,
) {
	return api.get<{ recordings: RecordingResponse[]; total: number }>(
		`/projects/${projectId}/recordings?offset=${offset}&limit=${limit}`,
	);
}

/**
 * Получить URL для скачивания записи.
 * @param recordingId - ID записи
 */
export function getRecordingDownloadURL(recordingId: string) {
	return api.get<RecordingDownloadResponse>(
		`/recordings/${recordingId}/download`,
	);
}

// ---------------------------------------------------------------------------
// Запланированные звонки (CALL-02)
// ---------------------------------------------------------------------------

/** Запланированный звонок. */
export interface ScheduledCallResponse {
	/** ID запланированного звонка. */
	id: string;
	/** Название. */
	title: string;
	/** Описание. */
	description?: string;
	/** Время следующего звонка. */
	scheduled_at: string;
	/** Длительность в минутах. */
	duration_minutes: number;
	/** За сколько минут напомнить. */
	reminder_minutes?: number | null;
	/** RFC 5545 RRULE для повторяющихся звонков. */
	rrule?: string;
	/** ID канала. */
	channel_id: string;
	/** ID проекта. */
	project_id: string;
	/** ID создателя. */
	created_by: string;
	/** Активен ли звонок. */
	is_active: boolean;
	/** ID приглашённых участников. */
	participant_ids?: string[];
	/** Дата создания. */
	created_at: string;
	/** Дата обновления. */
	updated_at: string;
}

/** Данные для создания запланированного звонка. */
export interface CreateScheduledCallData {
	/** Название звонка. */
	title: string;
	/** Описание. */
	description?: string;
	/** Время звонка. */
	scheduled_at: string;
	/** Длительность в минутах. */
	duration_minutes?: number;
	/** За сколько минут напомнить (null = без напоминания). */
	reminder_minutes?: number | null;
	/** RRULE для повторяющихся звонков. */
	rrule?: string;
	/** ID канала. */
	channel_id: string;
	/** ID участников звонка. */
	participant_ids?: string[];
}

/** Данные для обновления запланированного звонка. */
export interface UpdateScheduledCallData {
	/** Название. */
	title?: string;
	/** Описание. */
	description?: string;
	/** Время звонка. */
	scheduled_at?: string;
	/** RRULE. */
	rrule?: string;
	/** Активность. */
	is_active?: boolean;
}

/**
 * Создать запланированный звонок.
 * @param projectId - ID проекта
 * @param data - данные звонка
 */
export function createScheduledCall(
	projectId: string,
	data: CreateScheduledCallData,
) {
	return api.post<ScheduledCallResponse>(
		`/projects/${projectId}/scheduled-calls`,
		data,
	);
}

/**
 * Получить запланированные звонки канала.
 * @param channelId - ID канала
 */
export function listScheduledCalls(channelId: string) {
	return api.get<{ scheduled_calls: ScheduledCallResponse[] }>(
		`/channels/${channelId}/scheduled-calls`,
	);
}

/**
 * Обновить запланированный звонок.
 * @param id - ID запланированного звонка
 * @param data - данные для обновления
 */
export function updateScheduledCall(id: string, data: UpdateScheduledCallData) {
	return api.put<ScheduledCallResponse>(`/scheduled-calls/${id}`, data);
}

/**
 * Удалить запланированный звонок.
 * @param id - ID запланированного звонка
 */
export function deleteScheduledCall(id: string) {
	return api.delete<void>(`/scheduled-calls/${id}`);
}

// ---------------------------------------------------------------------------
// Проверка конфликтов расписания
// ---------------------------------------------------------------------------

/** Детали конфликта расписания. */
export interface ConflictDetail {
	/** ID участника с конфликтом. */
	participant_id: string;
	/** Название конфликтующего звонка. */
	call_title: string;
	/** Время конфликтующего звонка. */
	scheduled_at: string;
	/** ID канала конфликтующего звонка. */
	channel_id: string;
}

/** Ответ проверки конфликтов. */
export interface ConflictResponse {
	/** Есть ли конфликт. */
	has_conflict: boolean;
	/** Список конфликтов. */
	conflicts: ConflictDetail[];
}

/**
 * Проверить конфликты расписания для указанного времени и участников.
 * @param time - время начала (ISO 8601)
 * @param duration - длительность в минутах
 * @param participantIds - массив ID участников
 */
export function checkCallConflicts(
	time: string,
	duration: number,
	participantIds: string[],
) {
	const params = new URLSearchParams({
		time,
		duration: String(duration),
		participant_ids: participantIds.join(","),
	});
	return api.get<ConflictResponse>(`/calls/conflicts?${params.toString()}`);
}

/**
 * Получить все запланированные звонки по списку каналов.
 * Используется для календаря, который показывает звонки из всех каналов.
 * @param channelIds - массив ID каналов
 */
export async function listScheduledCallsForChannels(
	channelIds: string[],
): Promise<ScheduledCallResponse[]> {
	if (channelIds.length === 0) return [];

	const results = await Promise.all(
		channelIds.map((id) =>
			listScheduledCalls(id)
				.then((res) => res.scheduled_calls ?? [])
				.catch(() => [] as ScheduledCallResponse[]),
		),
	);

	return results.flat();
}
