import { create } from "zustand";

import type { CallResponse } from "@/api/calls";
import * as callsApi from "@/api/calls";

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
	/** Идёт запись звонка. */
	isRecording: boolean;
	/** ID текущей записи. */
	recordingId: string | null;
	/** ID пользователя, начавшего запись. */
	recordingStartedBy: string | null;

	/** Начать новый звонок в канале. */
	startCall: (channelId: string) => Promise<void>;
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
			} catch {
				// Игнорируем ошибку при выходе
			}
		}
		set(initialState);
	},

	endCall: async () => {
		const { activeCall } = get();
		if (activeCall) {
			try {
				await callsApi.endCall(activeCall.id);
			} catch {
				// Игнорируем ошибку при завершении
			}
		}
		set(initialState);
	},

	setError: (msg) => set({ error: msg }),

	checkActiveCall: async (channelId) => {
		try {
			return await callsApi.getActiveCall(channelId);
		} catch {
			return null;
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

	setRecordingState: (isRecording, recordingId, startedBy) =>
		set({
			isRecording,
			recordingId,
			recordingStartedBy: startedBy,
		}),

	reset: () => set(initialState),
}));
