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
	type LucideIcon,
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

/** Round control button with tooltip and label. */
function ControlButton({
	icon: Icon,
	label,
	active,
	activeColor,
	danger,
	onClick,
}: {
	icon: LucideIcon;
	label: string;
	active?: boolean;
	activeColor?: string;
	danger?: boolean;
	onClick: () => void;
}) {
	return (
		<Tooltip>
			<TooltipTrigger asChild>
				<div className="flex flex-col items-center gap-1">
					<button
						type="button"
						onClick={onClick}
						aria-label={label}
						className={cn(
							"relative flex h-12 w-12 items-center justify-center rounded-full transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-200",
							"hover:scale-105 active:scale-95",
							danger
								? "bg-red-500/15 text-red-500 hover:bg-red-500/25"
								: active && activeColor
									? `bg-white/10 ${activeColor} hover:bg-white/15`
									: active
										? "bg-white/10 text-white hover:bg-white/15"
										: "bg-white/10 text-white/60 hover:bg-white/15",
						)}
					>
						<Icon className="h-5 w-5" />
					</button>
					<span className="text-[10px] text-white/40">{label}</span>
				</div>
			</TooltipTrigger>
			<TooltipContent>{label}</TooltipContent>
		</Tooltip>
	);
}

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

	/** Переключить демонстрацию экрана с высоким качеством. */
	const handleScreenShareToggle = async () => {
		if (!isScreenSharing) {
			await room.localParticipant.setScreenShareEnabled(true, {
				resolution: { width: 1920, height: 1080, frameRate: 15 },
				contentHint: "detail",
			});
		} else {
			await room.localParticipant.setScreenShareEnabled(false);
		}
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
					<ControlButton
						icon={isMicEnabled ? Mic : MicOff}
						label={isMicEnabled ? "Выключить микрофон" : "Включить микрофон"}
						active={isMicEnabled}
						danger={!isMicEnabled}
						onClick={handleMicToggle}
					/>
					<ControlButton
						icon={isCameraEnabled ? Camera : CameraOff}
						label={isCameraEnabled ? "Выключить камеру" : "Включить камеру"}
						active={isCameraEnabled}
						danger={!isCameraEnabled}
						onClick={handleCameraToggle}
					/>
				</div>

				{/* Разделитель */}
				<div className="w-px h-10 bg-white/10 mx-1" />

				{/* Группа: Экран + Рука + Реакции + Настройки */}
				<div className="flex items-center gap-2">
					<ControlButton
						icon={isScreenSharing ? MonitorOff : Monitor}
						label={isScreenSharing ? "Остановить демонстрацию" : "Демонстрация экрана"}
						active={isScreenSharing}
						activeColor="text-blue-400"
						onClick={handleScreenShareToggle}
					/>
					<ControlButton
						icon={Hand}
						label={handRaised ? "Опустить руку" : "Поднять руку"}
						active={handRaised}
						activeColor="text-amber-400"
						onClick={toggleHand}
					/>
					<ControlButton
						icon={Smile}
						label="Реакции"
						active={reactionsOpen}
						activeColor="text-purple-400"
						onClick={() => setReactionsOpen(!reactionsOpen)}
					/>
					<ControlButton
						icon={Settings}
						label="Настройки"
						onClick={() => setSettingsOpen(true)}
					/>
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
