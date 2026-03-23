import { LiveKitRoom, RoomAudioRenderer, useParticipants } from "@livekit/components-react";
import { MessageSquare, Users } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import "@livekit/components-styles";

import { cn } from "@/lib/utils";
import { useCallStore } from "@/stores/call-store";
import { CallChat } from "./CallChat";
import { CallControls } from "./CallControls";
import { CallReactionsOverlay } from "./CallReactions";
import { ParticipantList } from "./ParticipantList";
import { ScreenShareView } from "./ScreenShareView";

/** URL сервера LiveKit: через Caddy reverse proxy на HTTPS, напрямую на HTTP. */
const LIVEKIT_URL =
	(import.meta.env.VITE_LIVEKIT_URL as string) ||
	(window.location.protocol === "https:"
		? `wss://${window.location.host}/lk`
		: `ws://${window.location.hostname}:7880`);

/** Пропсы компонента CallRoom. */
interface CallRoomProps {
	/** JWT-токен для подключения к LiveKit. */
	token: string;
	/** Коллбэк при отключении от звонка. */
	onDisconnect: () => void;
}

/**
 * Внутренний контент комнаты — используется внутри LiveKitRoom,
 * чтобы хуки LiveKit (useParticipants) имели доступ к контексту.
 */
function CallRoomContent({ onDisconnect }: { onDisconnect: () => void }) {
	const activeCall = useCallStore((s) => s.activeCall);
	const participants = useParticipants();
	const [elapsedSeconds, setElapsedSeconds] = useState(0);
	const [chatVisible, setChatVisible] = useState(false);
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

	const roomName = activeCall?.room_name ?? "Звонок";

	return (
		<>
			{/* Воспроизведение аудио всех участников */}
			<RoomAudioRenderer />

			{/* Заголовок звонка */}
			<div className="flex items-center justify-between px-4 py-3 border-b border-white/5">
				<div className="flex items-center gap-3">
					{/* Пульсирующий индикатор */}
					<span className="relative flex h-2.5 w-2.5">
						<span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
						<span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500" />
					</span>

					{/* Название комнаты */}
					<span className="text-sm font-semibold text-white/90 truncate max-w-[200px]">
						{roomName}
					</span>

					{/* Бейдж количества участников */}
					<span className="flex items-center gap-1 rounded-full bg-white/10 px-2 py-0.5 text-xs text-white/60">
						<Users className="h-3 w-3" />
						{participants.length}
					</span>
				</div>

				<div className="flex items-center gap-3">
					{/* Таймер */}
					<span className="text-sm font-mono text-white/50 tabular-nums tracking-wider">
						{formatDuration(elapsedSeconds)}
					</span>

					{/* Кнопка чата */}
					{activeCall && (
						<button
							type="button"
							onClick={() => setChatVisible(!chatVisible)}
							className={cn(
								"flex h-8 w-8 items-center justify-center rounded-lg transition-all duration-200",
								chatVisible
									? "bg-blue-500/20 text-blue-400"
									: "bg-white/5 text-white/50 hover:bg-white/10 hover:text-white/70",
							)}
							title={chatVisible ? "Скрыть чат" : "Показать чат"}
						>
							<MessageSquare className="h-4 w-4" />
						</button>
					)}
				</div>
			</div>

			{/* Основная область: участники + демонстрация экрана + чат */}
			<div className="relative flex flex-1 overflow-hidden">
				{/* Оверлей плавающих реакций */}
				<CallReactionsOverlay />

				<div className="flex flex-1 gap-2 overflow-hidden p-3">
					<ParticipantList />
					<ScreenShareView />
				</div>
				{activeCall && chatVisible && (
					<CallChat callId={activeCall.id} />
				)}
			</div>

			{/* Панель управления звонком */}
			<CallControls onDisconnect={onDisconnect} />
		</>
	);
}

/**
 * Главный контейнер звонка.
 * Тёмный фон, таймер длительности, сетка участников,
 * демонстрация экрана и панель управления.
 */
export function CallRoom({ token, onDisconnect }: CallRoomProps) {
	return (
		<LiveKitRoom
			token={token}
			serverUrl={LIVEKIT_URL}
			connect={true}
			audio={false}
			video={false}
			onDisconnected={onDisconnect}
			data-lk-theme="default"
			className="flex h-full w-full flex-col bg-[#0B0F19]"
		>
			<CallRoomContent onDisconnect={onDisconnect} />
		</LiveKitRoom>
	);
}
