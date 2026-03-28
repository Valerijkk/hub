import { useLocalParticipant, useRoomContext } from "@livekit/components-react";
import { Track } from "livekit-client";
import {
	Camera,
	CameraOff,
	Hand,
	Mic,
	MicOff,
	Monitor,
	MonitorOff,
	PhoneOff,
	Settings,
	Smile,
} from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";
import { useCallStore } from "@/stores/call-store";
import { CallReactionBar } from "./CallReactions";
import { CallSettingsModal } from "./CallSettingsModal";

/** Пропсы компонента CallControls. */
interface CallControlsProps {
	/** Коллбэк при отключении от звонка. */
	onDisconnect: () => void;
}

/**
 * Панель управления звонком.
 * Центрированная нижняя панель со стеклянным эффектом,
 * большие круглые кнопки с подписями и разделителями.
 */
export function CallControls({ onDisconnect }: CallControlsProps) {
	const { localParticipant } = useLocalParticipant();
	const room = useRoomContext();
	const { activeCall, handRaised, toggleHand, leaveCall, endCall } =
		useCallStore();
	const currentUserId = useAuthStore((s) => s.user?.id);
	const [cameraDialogOpen, setCameraDialogOpen] = useState(false);
	const [settingsOpen, setSettingsOpen] = useState(false);
	const [reactionsOpen, setReactionsOpen] = useState(false);

	/** Микрофон включён. */
	const isMicEnabled = localParticipant.isMicrophoneEnabled;

	/** Демонстрация экрана активна. */
	const isScreenSharing = localParticipant
		.getTrackPublications()
		.some(
			(pub) =>
				pub.source === Track.Source.ScreenShare && pub.track !== undefined,
		);

	/** Камера включена. */
	const isCameraEnabled = localParticipant.isCameraEnabled;

	/** Текущий пользователь -- создатель звонка. */
	const isCreator = activeCall?.created_by === currentUserId;

	/** Переключить микрофон. */
	const handleMicToggle = async () => {
		await localParticipant.setMicrophoneEnabled(!isMicEnabled);
	};

	/** Переключить демонстрацию экрана. */
	const handleScreenShareToggle = async () => {
		await room.localParticipant.setScreenShareEnabled(!isScreenSharing);
	};

	/** Переключить камеру (с подтверждением при включении). */
	const handleCameraToggle = async () => {
		if (isCameraEnabled) {
			await localParticipant.setCameraEnabled(false);
		} else {
			setCameraDialogOpen(true);
		}
	};

	/** Подтвердить включение камеры. */
	const handleCameraConfirm = async () => {
		setCameraDialogOpen(false);
		await localParticipant.setCameraEnabled(true);
	};

	/** Покинуть звонок. */
	const handleLeave = async () => {
		await leaveCall();
		onDisconnect();
	};

	/** Завершить звонок для всех. */
	const handleEnd = async () => {
		await endCall();
		onDisconnect();
	};

	return (
		<TooltipProvider>
			<div className="relative flex items-center justify-center gap-2 border-t border-white/5 bg-black/50 backdrop-blur-xl px-4 py-3">
				{/* Панель реакций — всплывает над кнопкой */}
				{reactionsOpen && (
					<CallReactionBar onClose={() => setReactionsOpen(false)} />
				)}

				{/* Группа: Микрофон + Камера */}
				<div className="flex items-center gap-2">
					{/* Микрофон */}
					<Tooltip>
						<TooltipTrigger asChild>
							<div className="flex flex-col items-center gap-1">
								<button
									type="button"
									onClick={handleMicToggle}
									aria-label={
										isMicEnabled ? "Выключить микрофон" : "Включить микрофон"
									}
									className={cn(
										"relative flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
										"hover:scale-105 active:scale-95",
										isMicEnabled
											? "bg-white/10 text-white hover:bg-white/15"
											: "bg-red-500/15 text-red-500 hover:bg-red-500/25",
									)}
								>
									{isMicEnabled ? (
										<Mic className="h-5 w-5" />
									) : (
										<MicOff className="h-5 w-5" />
									)}
								</button>
								<span className="text-[10px] text-white/40">Микрофон</span>
							</div>
						</TooltipTrigger>
						<TooltipContent>
							{isMicEnabled ? "Выключить микрофон" : "Включить микрофон"}
						</TooltipContent>
					</Tooltip>

					{/* Камера */}
					<Tooltip>
						<TooltipTrigger asChild>
							<div className="flex flex-col items-center gap-1">
								<button
									type="button"
									onClick={handleCameraToggle}
									aria-label={
										isCameraEnabled ? "Выключить камеру" : "Включить камеру"
									}
									className={cn(
										"relative flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
										"hover:scale-105 active:scale-95",
										isCameraEnabled
											? "bg-white/10 text-white hover:bg-white/15"
											: "bg-red-500/15 text-red-500 hover:bg-red-500/25",
									)}
								>
									{isCameraEnabled ? (
										<Camera className="h-5 w-5" />
									) : (
										<CameraOff className="h-5 w-5" />
									)}
								</button>
								<span className="text-[10px] text-white/40">Камера</span>
							</div>
						</TooltipTrigger>
						<TooltipContent>
							{isCameraEnabled ? "Выключить камеру" : "Включить камеру"}
						</TooltipContent>
					</Tooltip>
				</div>

				{/* Разделитель */}
				<div className="w-px h-10 bg-white/10 mx-1" />

				{/* Группа: Экран + Рука + Реакции + Настройки */}
				<div className="flex items-center gap-2">
					{/* Демонстрация экрана */}
					<Tooltip>
						<TooltipTrigger asChild>
							<div className="flex flex-col items-center gap-1">
								<button
									type="button"
									onClick={handleScreenShareToggle}
									aria-label={
										isScreenSharing
											? "Остановить демонстрацию"
											: "Демонстрация экрана"
									}
									className={cn(
										"flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
										"hover:scale-105 active:scale-95",
										isScreenSharing
											? "bg-white/10 text-blue-400 hover:bg-white/15"
											: "bg-white/10 text-white/60 hover:bg-white/15",
									)}
								>
									{isScreenSharing ? (
										<MonitorOff className="h-5 w-5" />
									) : (
										<Monitor className="h-5 w-5" />
									)}
								</button>
								<span className="text-[10px] text-white/40">Экран</span>
							</div>
						</TooltipTrigger>
						<TooltipContent>
							{isScreenSharing
								? "Остановить демонстрацию"
								: "Демонстрация экрана"}
						</TooltipContent>
					</Tooltip>

					{/* Поднять руку */}
					<Tooltip>
						<TooltipTrigger asChild>
							<div className="flex flex-col items-center gap-1">
								<button
									type="button"
									onClick={toggleHand}
									aria-label={handRaised ? "Опустить руку" : "Поднять руку"}
									className={cn(
										"flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
										"hover:scale-105 active:scale-95",
										handRaised
											? "bg-white/10 text-amber-400 hover:bg-white/15"
											: "bg-white/10 text-white/60 hover:bg-white/15",
									)}
								>
									<Hand className="h-5 w-5" />
								</button>
								<span className="text-[10px] text-white/40">Рука</span>
							</div>
						</TooltipTrigger>
						<TooltipContent>
							{handRaised ? "Опустить руку" : "Поднять руку"}
						</TooltipContent>
					</Tooltip>

					{/* Реакции */}
					<Tooltip>
						<TooltipTrigger asChild>
							<div className="flex flex-col items-center gap-1">
								<button
									type="button"
									onClick={() => setReactionsOpen(!reactionsOpen)}
									aria-label="Реакции"
									className={cn(
										"flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
										"hover:scale-105 active:scale-95",
										reactionsOpen
											? "bg-white/10 text-purple-400 hover:bg-white/15"
											: "bg-white/10 text-white/60 hover:bg-white/15",
									)}
								>
									<Smile className="h-5 w-5" />
								</button>
								<span className="text-[10px] text-white/40">Реакции</span>
							</div>
						</TooltipTrigger>
						<TooltipContent>Реакции</TooltipContent>
					</Tooltip>

					{/* Настройки устройств */}
					<Tooltip>
						<TooltipTrigger asChild>
							<div className="flex flex-col items-center gap-1">
								<button
									type="button"
									onClick={() => setSettingsOpen(true)}
									aria-label="Настройки устройств"
									className={cn(
										"flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
										"bg-white/10 text-white/60 hover:bg-white/15",
										"hover:scale-105 active:scale-95",
									)}
								>
									<Settings className="h-5 w-5" />
								</button>
								<span className="text-[10px] text-white/40">
									Настройки
								</span>
							</div>
						</TooltipTrigger>
						<TooltipContent>Настройки устройств</TooltipContent>
					</Tooltip>
				</div>

				{/* Разделитель */}
				<div className="w-px h-10 bg-white/10 mx-1" />

				{/* Группа: Выход */}
				<div className="flex items-center gap-2">
					{/* Покинуть звонок — pill shape */}
					<Tooltip>
						<TooltipTrigger asChild>
							<button
								type="button"
								onClick={handleLeave}
								aria-label="Покинуть звонок"
								className={cn(
									"flex h-12 min-w-[72px] items-center justify-center gap-2 rounded-full px-5",
									"bg-red-500 text-white hover:bg-red-600",
									"transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 hover:scale-105 active:scale-95",
									"shadow-lg shadow-red-500/25",
									"text-sm font-semibold",
								)}
							>
								<PhoneOff className="h-5 w-5" />
								<span>Выйти</span>
							</button>
						</TooltipTrigger>
						<TooltipContent>Покинуть звонок</TooltipContent>
					</Tooltip>

					{/* Завершить звонок (только создатель) */}
					{isCreator && (
						<Tooltip>
							<TooltipTrigger asChild>
								<button
									type="button"
									onClick={handleEnd}
									aria-label="Завершить звонок для всех"
									className={cn(
										"flex h-10 items-center gap-2 rounded-full px-4",
										"bg-red-600/80 text-white/90 hover:bg-red-700",
										"transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200 hover:scale-105 active:scale-95",
										"text-xs font-medium",
									)}
								>
									Завершить
								</button>
							</TooltipTrigger>
							<TooltipContent>Завершить звонок для всех</TooltipContent>
						</Tooltip>
					)}
				</div>
			</div>

			{/* Диалог подтверждения включения камеры */}
			<Dialog open={cameraDialogOpen} onOpenChange={setCameraDialogOpen}>
				<DialogContent className="border-white/10 bg-[#1a1f2e] text-white sm:max-w-md">
					<DialogHeader>
						<DialogTitle>Включить камеру?</DialogTitle>
						<DialogDescription className="text-white/60">
							Все участники звонка увидят вас.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter className="gap-2 sm:gap-0">
						<Button
							variant="ghost"
							className="text-white/70 hover:text-white hover:bg-white/10"
							onClick={() => setCameraDialogOpen(false)}
						>
							Отмена
						</Button>
						<Button
							className="bg-blue-600 text-white hover:bg-blue-700"
							onClick={handleCameraConfirm}
						>
							Включить
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Модальное окно настроек устройств */}
			<CallSettingsModal open={settingsOpen} onOpenChange={setSettingsOpen} />
		</TooltipProvider>
	);
}
