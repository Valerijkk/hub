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
  creator_id: string;
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
    return await api.get<CallResponse>(`/calls/channel/${channelId}`);
  } catch (err: unknown) {
    const apiErr = err as { code?: number };
    if (apiErr.code === 404) return null;
    throw err;
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
