import {
  useLocalParticipant,
  useRoomContext,
} from "@livekit/components-react";
import { Track } from "livekit-client";
import { Hand, Mic, MicOff, Monitor, MonitorOff, PhoneOff } from "lucide-react";

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useAuthStore } from "@/stores/auth-store";
import { useCallStore } from "@/stores/call-store";
import { cn } from "@/lib/utils";

/** Пропсы компонента CallControls. */
interface CallControlsProps {
  /** Коллбэк при отключении от звонка. */
  onDisconnect: () => void;
}

/**
 * Панель управления звонком в стиле Tinkoff Talk.
 * Центрированная нижняя панель со стеклянным эффектом,
 * большие круглые кнопки с активным/неактивным состоянием.
 */
export function CallControls({ onDisconnect }: CallControlsProps) {
  const { localParticipant } = useLocalParticipant();
  const room = useRoomContext();
  const { activeCall, handRaised, toggleHand, leaveCall, endCall } =
    useCallStore();
  const currentUserId = useAuthStore((s) => s.user?.id);

  /** Микрофон включён. */
  const isMicEnabled = localParticipant.isMicrophoneEnabled;

  /** Демонстрация экрана активна. */
  const isScreenSharing = localParticipant
    .getTrackPublications()
    .some(
      (pub) =>
        pub.source === Track.Source.ScreenShare && pub.track !== undefined,
    );

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
      <div className="flex items-center justify-center gap-3 border-t border-white/10 bg-black/40 backdrop-blur-xl p-4">
        {/* Микрофон */}
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={handleMicToggle}
              aria-label={isMicEnabled ? "Выключить микрофон" : "Включить микрофон"}
              className={cn(
                "relative flex h-12 w-12 items-center justify-center rounded-full transition-all duration-200",
                isMicEnabled
                  ? "bg-white/10 text-white hover:bg-white/20 ring-2 ring-green-500/50"
                  : "bg-red-500/20 text-red-400 hover:bg-red-500/30 ring-2 ring-red-500/50",
              )}
            >
              {isMicEnabled ? (
                <Mic className="h-5 w-5" />
              ) : (
                <MicOff className="h-5 w-5" />
              )}
            </button>
          </TooltipTrigger>
          <TooltipContent>
            {isMicEnabled ? "Выключить микрофон" : "Включить микрофон"}
          </TooltipContent>
        </Tooltip>

        {/* Демонстрация экрана */}
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={handleScreenShareToggle}
              aria-label={isScreenSharing ? "Остановить демонстрацию" : "Демонстрация экрана"}
              className={cn(
                "flex h-12 w-12 items-center justify-center rounded-full transition-all duration-200",
                isScreenSharing
                  ? "bg-blue-500/20 text-blue-400 hover:bg-blue-500/30 ring-2 ring-blue-500/50"
                  : "bg-white/10 text-white/70 hover:bg-white/20",
              )}
            >
              {isScreenSharing ? (
                <MonitorOff className="h-5 w-5" />
              ) : (
                <Monitor className="h-5 w-5" />
              )}
            </button>
          </TooltipTrigger>
          <TooltipContent>
            {isScreenSharing ? "Остановить демонстрацию" : "Демонстрация экрана"}
          </TooltipContent>
        </Tooltip>

        {/* Поднять руку */}
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={toggleHand}
              aria-label={handRaised ? "Опустить руку" : "Поднять руку"}
              className={cn(
                "flex h-12 w-12 items-center justify-center rounded-full transition-all duration-200",
                handRaised
                  ? "bg-amber-500/20 text-amber-400 hover:bg-amber-500/30 ring-2 ring-amber-500/50"
                  : "bg-white/10 text-white/70 hover:bg-white/20",
              )}
            >
              <Hand className="h-5 w-5" />
            </button>
          </TooltipTrigger>
          <TooltipContent>
            {handRaised ? "Опустить руку" : "Поднять руку"}
          </TooltipContent>
        </Tooltip>

        {/* Разделитель */}
        <div className="w-px h-8 bg-white/10 mx-1" />

        {/* Покинуть звонок */}
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={handleLeave}
              aria-label="Покинуть звонок"
              className="flex h-12 w-12 items-center justify-center rounded-full bg-red-500 text-white hover:bg-red-600 transition-all duration-200 shadow-lg shadow-red-500/20"
            >
              <PhoneOff className="h-5 w-5" />
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
                className="flex h-10 items-center gap-2 rounded-full bg-red-600 px-4 text-sm font-medium text-white hover:bg-red-700 transition-all duration-200 shadow-lg shadow-red-600/20"
              >
                Завершить
              </button>
            </TooltipTrigger>
            <TooltipContent>Завершить звонок для всех</TooltipContent>
          </Tooltip>
        )}
      </div>
    </TooltipProvider>
  );
}
