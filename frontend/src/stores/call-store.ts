import { create } from "zustand";

import type { CallResponse } from "@/api/calls";
import * as callsApi from "@/api/calls";

/** Извлечь сообщение из Error / ApiError. */
function errorMessage(err: unknown, fallback: string): string {
	if (err instanceof Error && err.message.trim()) return err.message;
	if (err && typeof err === "object" && "message" in err) {
		const msg = (err as { message: unknown }).message;
		if (typeof msg === "string" && msg.trim()) return msg;
	}
	return fallback;
}

/** Состояние стора звонков. */
interface CallState {
	/** Активный звонок (null если не в звонке). */
	activeCall: CallResponse | null;
	/** JWT-токен для подключения к LiveKit. */
	token: string | null;
	/** Пользователь находится в звонке. */
	isInCall: boolean;
	/** Идёт подключение к звонку. */
	isConnecting: boolean;
	/** Сообщение об ошибке. */
	error: string | null;
	/** Рука поднята. */
	handRaised: boolean;
	/** Логирование чата включено (сообщения сохраняются в БД). */
	chatLoggingEnabled: boolean;
	/** Идёт запись звонка. */
	isRecording: boolean;
	/** ID текущей записи. */
	recordingId: string | null;
	/** ID пользователя, начавшего запись. */
	recordingStartedBy: string | null;

	/** Начать новый звонок в канале. */
	startCall: (channelId: string) => Promise<void>;
	/** Создать быстрый звонок без канала. */
	startInstantCall: (title?: string) => Promise<string | null>;
	/** Присоединиться по invite-коду. */
	joinByCode: (code: string) => Promise<void>;
	/** Присоединиться к существующему звонку. */
	joinCall: (callId: string) => Promise<void>;
	/** Покинуть текущий звонок. */
	leaveCall: () => Promise<void>;
	/** Завершить звонок (только создатель). */
	endCall: () => Promise<void>;
	/** Установить ошибку. */
	setError: (msg: string | null) => void;
	/** Проверить наличие активного звонка в канале. */
	checkActiveCall: (channelId: string) => Promise<CallResponse | null>;
	/** Поднять/опустить руку. */
	toggleHand: () => Promise<void>;
	/** Переключить логирование чата. */
	toggleChatLogging: () => void;
	/** Установить состояние записи. */
	setRecordingState: (
		isRecording: boolean,
		recordingId: string | null,
		startedBy: string | null,
	) => void;
	/** Сбросить состояние звонка. */
	reset: () => void;
}

/** Начальное состояние. */
const initialState = {
	activeCall: null,
	token: null,
	isInCall: false,
	isConnecting: false,
	error: null,
	handRaised: false,
	chatLoggingEnabled: true,
	isRecording: false,
	recordingId: null,
	recordingStartedBy: null,
};

/** Стор управления состоянием голосовых звонков. */
export const useCallStore = create<CallState>()((set, get) => ({
	...initialState,

	startCall: async (channelId) => {
		set({ isConnecting: true, error: null });
		try {
			const call = await callsApi.createCall(channelId);
			set({
				activeCall: call,
				token: call.token ?? null,
				isInCall: true,
				isConnecting: false,
			});
		} catch (err: unknown) {
			const msg =
				err instanceof Error ? err.message : "Не удалось начать звонок";
			set({ isConnecting: false, error: msg });
		}
	},

	startInstantCall: async (title?: string) => {
		set({ isConnecting: true, error: null });
		try {
			const call = await callsApi.createInstantCall(title);
			set({
				activeCall: call,
				token: call.token ?? null,
				isInCall: true,
				isConnecting: false,
			});
			return call.invite_code ?? null;
		} catch (err: unknown) {
			const msg =
				err instanceof Error ? err.message : "Не удалось создать звонок";
			set({ isConnecting: false, error: msg });
			return null;
		}
	},

	joinByCode: async (code: string) => {
		set({ isConnecting: true, error: null });
		try {
			const call = await callsApi.joinByInviteCode(code);
			set({
				activeCall: call,
				token: call.token ?? null,
				isInCall: true,
				isConnecting: false,
			});
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : "Звонок не найден";
			set({ isConnecting: false, error: msg });
		}
	},

	joinCall: async (callId) => {
		set({ isConnecting: true, error: null });
		try {
			const call = await callsApi.joinCall(callId);
			set({
				activeCall: call,
				token: call.token ?? null,
				isInCall: true,
				isConnecting: false,
			});
		} catch (err: unknown) {
			const msg =
				err instanceof Error
					? err.message
					: "Не удалось присоединиться к звонку";
			set({ isConnecting: false, error: msg });
		}
	},

	leaveCall: async () => {
		const { activeCall } = get();
		if (activeCall) {
			try {
				await callsApi.leaveCall(activeCall.id);
			} catch (err: unknown) {
				set({ error: errorMessage(err, "Не удалось покинуть звонок") });
				return;
			}
		}
		set(initialState);
	},

	endCall: async () => {
		const { activeCall } = get();
		if (activeCall) {
			try {
				await callsApi.endCall(activeCall.id);
			} catch (err: unknown) {
				set({ error: errorMessage(err, "Не удалось завершить звонок") });
				return;
			}
		}
		set(initialState);
	},

	setError: (msg) => set({ error: msg }),

	checkActiveCall: async (channelId) => {
		try {
			return await callsApi.getActiveCall(channelId);
		} catch (err: unknown) {
			set({
				error: errorMessage(err, "Не удалось проверить активный звонок"),
			});
			throw err;
		}
	},

	toggleHand: async () => {
		const { activeCall, handRaised } = get();
		if (!activeCall) return;
		try {
			await callsApi.raiseHand(activeCall.id, !handRaised);
			set({ handRaised: !handRaised });
		} catch {
			// Игнорируем ошибку
		}
	},

	toggleChatLogging: () =>
		set((s) => ({ chatLoggingEnabled: !s.chatLoggingEnabled })),

	setRecordingState: (isRecording, recordingId, startedBy) =>
		set({
			isRecording,
			recordingId,
			recordingStartedBy: startedBy,
		}),

	reset: () => set(initialState),
}));
