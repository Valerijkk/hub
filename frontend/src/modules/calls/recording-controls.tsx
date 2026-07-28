import { useCallback, useEffect, useRef, useState } from "react";
import * as callsApi from "@/api/calls";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useCallStore } from "@/stores/call-store";

// ---------------------------------------------------------------------------
// RecordingControls — кнопка записи звонка и индикатор записи
// ---------------------------------------------------------------------------

/** Пропсы компонента RecordingControls. */
interface RecordingControlsProps {
	/** ID текущего звонка. */
	callId: string;
}

/**
 * RecordingControls отображает кнопку начала/остановки записи звонка
 * и индикатор записи для всех участников.
 *
 * При начале записи публикуется WS-событие calls.recording.started,
 * которое показывает индикатор "REC" всем участникам звонка.
 */
export function RecordingControls({ callId }: RecordingControlsProps) {
	const { isRecording, recordingId, recordingStartedBy, setRecordingState } =
		useCallStore();
	const userId = useAuthStore((s) => s.user?.id);

	const [isLoading, setIsLoading] = useState(false);
	const [elapsed, setElapsed] = useState(0);
	const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

	// Таймер записи
	useEffect(() => {
		if (isRecording) {
			setElapsed(0);
			timerRef.current = setInterval(() => {
				setElapsed((prev) => prev + 1);
			}, 1000);
		} else {
			if (timerRef.current) {
				clearInterval(timerRef.current);
				timerRef.current = null;
			}
			setElapsed(0);
		}

		return () => {
			if (timerRef.current) {
				clearInterval(timerRef.current);
			}
		};
	}, [isRecording]);

	/** Форматирует секунды в MM:SS. */
	const formatTime = useCallback((seconds: number) => {
		const m = Math.floor(seconds / 60)
			.toString()
			.padStart(2, "0");
		const s = (seconds % 60).toString().padStart(2, "0");
		return `${m}:${s}`;
	}, []);

	/** Начать запись. */
	const handleStart = useCallback(async () => {
		setIsLoading(true);
		try {
			const rec = await callsApi.startRecording(callId);
			setRecordingState(true, rec.id, userId ?? "");
		} catch (err) {
			toastApiError(err, "Не удалось начать запись");
		} finally {
			setIsLoading(false);
		}
	}, [callId, userId, setRecordingState]);

	/** Остановить запись. */
	const handleStop = useCallback(async () => {
		if (!recordingId) return;
		setIsLoading(true);
		try {
			await callsApi.stopRecording(callId, recordingId);
			setRecordingState(false, null, null);
		} catch (err) {
			toastApiError(err, "Не удалось остановить запись");
		} finally {
			setIsLoading(false);
		}
	}, [callId, recordingId, setRecordingState]);

	const isMyRecording = recordingStartedBy === userId;

	return (
		<div className="flex items-center gap-2">
			{/* Индикатор записи -- виден ВСЕМ участникам */}
			{isRecording && (
				<div className="flex items-center gap-1.5 rounded-full bg-red-500/10 px-3 py-1 text-red-500">
					<span className="h-2 w-2 animate-blink rounded-full bg-red-500" />
					<span className="text-xs font-medium">REC</span>
					<span className="text-xs tabular-nums">{formatTime(elapsed)}</span>
				</div>
			)}

			{/* Кнопка записи */}
			{!isRecording ? (
				<button
					type="button"
					onClick={handleStart}
					disabled={isLoading}
					className="flex h-9 w-9 items-center justify-center rounded-full border border-border bg-base text-t2 transition-colors hover:bg-red-500/10 hover:text-red-500 disabled:opacity-50"
					title="Начать запись"
					aria-label="Начать запись звонка"
				>
					<span className="h-3 w-3 rounded-full bg-red-500" />
				</button>
			) : isMyRecording ? (
				<button
					type="button"
					onClick={handleStop}
					disabled={isLoading}
					className="flex h-9 w-9 items-center justify-center rounded-full border border-red-500 bg-red-500/10 text-red-500 transition-colors hover:bg-red-500/20 disabled:opacity-50"
					title="Остановить запись"
					aria-label="Остановить запись звонка"
				>
					<span className="h-3 w-3 rounded-sm bg-red-500" />
				</button>
			) : null}
		</div>
	);
}
