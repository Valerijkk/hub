import { useCallback, useEffect, useRef, useState } from "react";
import { uploadFile } from "@/api/files";
import { Button } from "@/components/ui/button";

/** Пропсы компонента VoiceRecorder. */
interface VoiceRecorderProps {
	/** Колбэк при успешной отправке голосового сообщения. */
	onSend: (voiceUrl: string, durationMs: number) => void;
	/** Колбэк при отмене записи. */
	onCancel: () => void;
}

/** Максимальная длительность записи (5 минут в мс). */
const MAX_DURATION_MS = 5 * 60 * 1000;

/**
 * VoiceRecorder -- компонент записи голосового сообщения.
 * Использует MediaRecorder API с кодеком opus.
 * Максимальная длительность -- 5 минут.
 */
export function VoiceRecorder({ onSend, onCancel }: VoiceRecorderProps) {
	const [isRecording, setIsRecording] = useState(false);
	const [isUploading, setIsUploading] = useState(false);
	const [duration, setDuration] = useState(0);
	const [error, setError] = useState<string | null>(null);

	const mediaRecorderRef = useRef<MediaRecorder | null>(null);
	const chunksRef = useRef<Blob[]>([]);
	const startTimeRef = useRef<number>(0);
	const timerRef = useRef<ReturnType<typeof setInterval> | undefined>(
		undefined,
	);

	/** Начать запись. */
	const startRecording = useCallback(async () => {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			const mimeType = MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
				? "audio/webm;codecs=opus"
				: "audio/webm";

			const recorder = new MediaRecorder(stream, { mimeType });
			chunksRef.current = [];

			recorder.ondataavailable = (e) => {
				if (e.data.size > 0) {
					chunksRef.current.push(e.data);
				}
			};

			recorder.onstop = () => {
				stream.getTracks().forEach((t) => t.stop());
				clearInterval(timerRef.current);
			};

			recorder.start(100);
			mediaRecorderRef.current = recorder;
			startTimeRef.current = Date.now();
			setIsRecording(true);
			setError(null);

			// Таймер обновления длительности.
			timerRef.current = setInterval(() => {
				const elapsed = Date.now() - startTimeRef.current;
				setDuration(elapsed);
				if (elapsed >= MAX_DURATION_MS) {
					recorder.stop();
					setIsRecording(false);
				}
			}, 200);
		} catch {
			setError("Микрофон недоступен. Проверьте разрешения браузера.");
		}
	}, []);

	/** Остановить запись и загрузить. */
	const stopAndSend = useCallback(async () => {
		const recorder = mediaRecorderRef.current;
		if (!recorder || recorder.state !== "recording") return;

		const durationMs = Date.now() - startTimeRef.current;
		recorder.stop();
		setIsRecording(false);

		// Ждём последний chunk.
		await new Promise<void>((resolve) => {
			recorder.onstop = () => {
				recorder.stream.getTracks().forEach((t) => t.stop());
				clearInterval(timerRef.current);
				resolve();
			};
		});

		const blob = new Blob(chunksRef.current, { type: "audio/webm" });
		const file = new File([blob], `voice-${Date.now()}.webm`, {
			type: "audio/webm",
		});

		setIsUploading(true);
		try {
			const fileId = await uploadFile(file, "messenger", "voice", "temp");
			// URL для файла.
			const voiceUrl = `/api/files/${fileId}/url`;
			onSend(voiceUrl, durationMs);
		} catch {
			setError("Ошибка загрузки записи.");
		} finally {
			setIsUploading(false);
		}
	}, [onSend]);

	/** Отмена записи. */
	const cancelRecording = useCallback(() => {
		const recorder = mediaRecorderRef.current;
		if (recorder && recorder.state === "recording") {
			recorder.stop();
			// MEDIUM: явно освобождаем MediaStream — иначе индикатор
			// микрофона горит до перезагрузки страницы.
			recorder.stream.getTracks().forEach((t) => t.stop());
		}
		clearInterval(timerRef.current);
		setIsRecording(false);
		onCancel();
	}, [onCancel]);

	// Очистка при размонтировании: останавливаем recorder, освобождаем
	// MediaStream-треки и чистим таймер. Без stream.stop индикатор микрофона
	// продолжает гореть после unmount компонента.
	useEffect(() => {
		return () => {
			clearInterval(timerRef.current);
			const recorder = mediaRecorderRef.current;
			if (recorder) {
				if (recorder.state === "recording") {
					recorder.stop();
				}
				recorder.stream.getTracks().forEach((t) => t.stop());
			}
		};
	}, []);

	/** Форматирование длительности. */
	const formatDuration = (ms: number) => {
		const seconds = Math.floor(ms / 1000);
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		return `${m}:${s.toString().padStart(2, "0")}`;
	};

	if (error) {
		return (
			<div className="flex items-center gap-2 rounded-md border border-destructive/50 p-3">
				<span className="text-sm text-destructive">{error}</span>
				<Button variant="ghost" size="sm" onClick={onCancel}>
					Закрыть
				</Button>
			</div>
		);
	}

	return (
		<div className="flex items-center gap-3 rounded-md border p-3">
			{!isRecording && !isUploading ? (
				<>
					<Button size="sm" variant="destructive" onClick={startRecording}>
						Начать запись
					</Button>
					<Button variant="ghost" size="sm" onClick={cancelRecording}>
						Отмена
					</Button>
				</>
			) : isUploading ? (
				<span className="text-sm text-t2">Загрузка...</span>
			) : (
				<>
					{/* Индикатор записи */}
					<div className="flex items-center gap-2">
						<span className="h-3 w-3 animate-pulse rounded-full bg-red-500" />
						<span className="text-sm font-medium">
							{formatDuration(duration)}
						</span>
					</div>
					<Button
						size="sm"
						onClick={() => void stopAndSend()}
						className="ml-auto"
					>
						Отправить
					</Button>
					<Button variant="ghost" size="sm" onClick={cancelRecording}>
						Отмена
					</Button>
				</>
			)}
		</div>
	);
}
