import { LiveKitRoom, RoomAudioRenderer, useLocalParticipant, useParticipants } from "@livekit/components-react";
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
/** Экспортируется для переиспользования в гостевом звонке (/call/$code). */
export function CallRoomContent({ onDisconnect }: { onDisconnect: () => void }) {
	const activeCall = useCallStore((s) => s.activeCall);
	const participants = useParticipants();
	const { localParticipant } = useLocalParticipant();
	const [elapsedSeconds, setElapsedSeconds] = useState(0);
	const [chatVisible, setChatVisible] = useState(false);
	const timerRef = useRef<ReturnType<typeof setInterval>>(null);
	const didMuteRef = useRef(false);

	/** Автоматически мьютим микрофон после подключения (по умолчанию OFF). */
	useEffect(() => {
		if (!didMuteRef.current && localParticipant.isMicrophoneEnabled) {
			didMuteRef.current = true;
			localParticipant.setMicrophoneEnabled(false);
		}
	}, [localParticipant, localParticipant.isMicrophoneEnabled]);

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

					{/* Кнопка копирования ссылки (для instant calls) */}
					{activeCall?.invite_code && (
						<button
							type="button"
							onClick={() => {
								const url = `${window.location.origin}/call/${activeCall.invite_code}`;
								navigator.clipboard.writeText(url).catch(() => {
									window.prompt("Скопируйте ссылку:", url);
								});
							}}
							className="flex items-center gap-1.5 rounded-full bg-blue-500/20 px-3 py-1 text-xs text-blue-400 hover:bg-blue-500/30 transition-colors"
							title="Скопировать ссылку приглашения"
						>
							<svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
								<path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
							</svg>
							Ссылка
						</button>
					)}
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
								"flex h-8 w-8 items-center justify-center rounded-lg transition-colors duration-200",
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
					<div className="flex flex-1 flex-col gap-2">
						<ParticipantList />

						{/* Блок приглашения — Телемост стиль (только когда 1 участник + invite) */}
						{activeCall?.invite_code && participants.length <= 1 && (
							<InviteBlock inviteCode={activeCall.invite_code} />
						)}
					</div>
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

/** Блок приглашения в стиле Телемост — когда ты один в звонке. */
function InviteBlock({ inviteCode }: { inviteCode: string }) {
	const [copied, setCopied] = useState(false);
	const url = `${window.location.origin}/call/${inviteCode}`;

	const copyLink = () => {
		navigator.clipboard.writeText(url).catch(() => {
			window.prompt("Скопируйте ссылку:", url);
		});
		setCopied(true);
		setTimeout(() => setCopied(false), 3000);
	};

	const shareViaTelegram = () => {
		window.open(`https://t.me/share/url?url=${encodeURIComponent(url)}&text=${encodeURIComponent("Присоединяйся к звонку!")}`, "_blank");
	};

	const shareViaWhatsApp = () => {
		window.open(`https://wa.me/?text=${encodeURIComponent("Присоединяйся к звонку! " + url)}`, "_blank");
	};

	return (
		<div className="flex flex-col items-center justify-center rounded-2xl bg-[hsl(220,11%,12%)] border border-white/[0.06] p-8 gap-5">
			<h3 className="text-[16px] font-semibold text-white text-center leading-snug">
				Чтобы пригласить других участников,<br />
				отправьте им ссылку на встречу
			</h3>

			{/* Главная кнопка копирования */}
			<button
				type="button"
				onClick={copyLink}
				className={cn(
					"flex items-center gap-2 px-6 py-2.5 rounded-lg text-sm font-medium transition-colors",
					copied
						? "bg-emerald-500 text-white"
						: "bg-emerald-500/90 text-white hover:bg-emerald-500",
				)}
			>
				{copied ? (
					<>
						<svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
							<path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
						</svg>
						Ссылка скопирована
					</>
				) : (
					<>
						<svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
							<path strokeLinecap="round" strokeLinejoin="round" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
						</svg>
						Скопировать ссылку
					</>
				)}
			</button>

			{/* Кнопки шаринга */}
			<div className="flex gap-3">
				<button
					type="button"
					onClick={shareViaTelegram}
					className="flex h-10 w-10 items-center justify-center rounded-lg bg-white/[0.07] text-white/60 hover:bg-white/[0.12] hover:text-white transition-colors"
					title="Telegram"
				>
					<svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
						<path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
					</svg>
				</button>
				<button
					type="button"
					onClick={shareViaWhatsApp}
					className="flex h-10 w-10 items-center justify-center rounded-lg bg-white/[0.07] text-white/60 hover:bg-white/[0.12] hover:text-white transition-colors"
					title="WhatsApp"
				>
					<svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
						<path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
					</svg>
				</button>
				<button
					type="button"
					onClick={copyLink}
					className="flex h-10 w-10 items-center justify-center rounded-lg bg-white/[0.07] text-white/60 hover:bg-white/[0.12] hover:text-white transition-colors"
					title="Скопировать ссылку"
				>
					<svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
						<path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
					</svg>
				</button>
			</div>

			{/* Номер встречи */}
			<div className="flex items-center gap-2 text-xs text-white/40">
				<span>Встреча: {inviteCode}</span>
				<button
					type="button"
					onClick={copyLink}
					className="hover:text-white/60 transition-colors"
					title="Скопировать"
				>
					<svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
						<path strokeLinecap="round" strokeLinejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
					</svg>
				</button>
			</div>
		</div>
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
			audio={true}
			video={false}
			onDisconnected={onDisconnect}
			data-lk-theme="default"
			className="flex h-full w-full flex-col bg-[#0B0F19]"
		>
			<CallRoomContent onDisconnect={onDisconnect} />
		</LiveKitRoom>
	);
}
