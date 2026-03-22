import { LiveKitRoom, RoomAudioRenderer } from "@livekit/components-react";
import { useEffect, useRef, useState } from "react";
import "@livekit/components-styles";

import { useCallStore } from "@/stores/call-store";
import { CallChat } from "./CallChat";
import { CallControls } from "./CallControls";
import { ParticipantList } from "./ParticipantList";
import { ScreenShareView } from "./ScreenShareView";

/** URL сервера LiveKit из переменных окружения. */
const LIVEKIT_URL =
	(import.meta.env.VITE_LIVEKIT_URL as string) ||
	`${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.hostname}:7880`;

/** Пропсы компонента CallRoom. */
interface CallRoomProps {
	/** JWT-токен для подключения к LiveKit. */
	token: string;
	/** Коллбэк при отключении от звонка. */
	onDisconnect: () => void;
}

/**
 * Главный контейнер звонка в стиле Tinkoff Talk.
 * Тёмный фон, таймер длительности, сетка участников,
 * демонстрация экрана и панель управления.
 */
export function CallRoom({ token, onDisconnect }: CallRoomProps) {
	const activeCall = useCallStore((s) => s.activeCall);
	const [elapsedSeconds, setElapsedSeconds] = useState(0);
	const timerRef = useRef<ReturnType<typeof setInterval>>(null);

	/** Таймер длительности звонка. */
	useEffect(() => {
		timerRef.current = setInterval(() => {
			setElapsedSeconds((s) => s + 1);
		}, 1000);
		return () => {
			if (timerRef.current) clearInterval(timerRef.current);
		};
	}, []);

	/** Форматирование времени в MM:SS или HH:MM:SS. */
	const formatDuration = (seconds: number): string => {
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		const s = seconds % 60;
		const pad = (n: number) => n.toString().padStart(2, "0");
		return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
	};

	return (
		<LiveKitRoom
			token={token}
			serverUrl={LIVEKIT_URL}
			connect={true}
			audio={true}
			video={false}
			onDisconnected={onDisconnect}
			data-lk-theme="default"
			className="flex h-full w-full flex-col bg-[#0B0F19]"
		>
			{/* Воспроизведение аудио всех участников */}
			<RoomAudioRenderer />

			{/* Заголовок звонка с таймером */}
			<div className="flex items-center justify-center gap-3 py-3 border-b border-white/10">
				<div className="flex items-center gap-2">
					<span className="relative flex h-2.5 w-2.5">
						<span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
						<span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500" />
					</span>
					<span className="text-sm font-medium text-white/90">
						{activeCall?.channel_id ? "Звонок" : "Звонок"}
					</span>
				</div>
				<span className="text-sm font-mono text-white/60">
					{formatDuration(elapsedSeconds)}
				</span>
			</div>

			{/* Основная область: участники + демонстрация экрана + чат */}
			<div className="flex flex-1 gap-2 overflow-hidden p-3">
				<ParticipantList />
				<ScreenShareView />
				{activeCall && <CallChat callId={activeCall.id} />}
			</div>

			{/* Панель управления звонком */}
			<CallControls onDisconnect={onDisconnect} />
		</LiveKitRoom>
	);
}
