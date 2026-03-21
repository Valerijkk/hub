import {
  useLocalParticipant,
  useRoomContext,
} from "@livekit/components-react";
import { Track } from "livekit-client";
import { Hand, Mic, MicOff, Monitor, MonitorOff, PhoneOff } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useAuthStore } from "@/stores/auth-store";
import { useCallStore } from "@/stores/call-store";

/** Пропсы компонента CallControls. */
interface CallControlsProps {
  /** Коллбэк при отключении от звонка. */
  onDisconnect: () => void;
}

/**
 * Панель управления звонком.
 * Кнопки: микрофон, демонстрация экрана, поднять руку, покинуть, завершить.
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

  /** Текущий пользователь — создатель звонка. */
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
      <div className="flex items-center justify-center gap-2 border-t bg-background p-3">
        {/* Микрофон */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant={isMicEnabled ? "secondary" : "outline"}
              size="icon"
              onClick={handleMicToggle}
              aria-label={isMicEnabled ? "Выключить микрофон" : "Включить микрофон"}
            >
              {isMicEnabled ? <Mic className="h-4 w-4" /> : <MicOff className="h-4 w-4" />}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {isMicEnabled ? "Выключить микрофон" : "Включить микрофон"}
          </TooltipContent>
        </Tooltip>

        {/* Демонстрация экрана */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant={isScreenSharing ? "secondary" : "outline"}
              size="icon"
              onClick={handleScreenShareToggle}
              aria-label={isScreenSharing ? "Остановить демонстрацию" : "Демонстрация экрана"}
            >
              {isScreenSharing ? (
                <MonitorOff className="h-4 w-4" />
              ) : (
                <Monitor className="h-4 w-4" />
              )}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {isScreenSharing ? "Остановить демонстрацию" : "Демонстрация экрана"}
          </TooltipContent>
        </Tooltip>

        {/* Поднять руку */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant={handRaised ? "default" : "outline"}
              size="icon"
              onClick={toggleHand}
              aria-label={handRaised ? "Опустить руку" : "Поднять руку"}
            >
              <Hand className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {handRaised ? "Опустить руку" : "Поднять руку"}
          </TooltipContent>
        </Tooltip>

        {/* Покинуть звонок */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="destructive"
              size="icon"
              onClick={handleLeave}
              aria-label="Покинуть звонок"
            >
              <PhoneOff className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Покинуть звонок</TooltipContent>
        </Tooltip>

        {/* Завершить звонок (только создатель) */}
        {isCreator && (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="destructive"
                onClick={handleEnd}
                className="ml-2"
                aria-label="Завершить звонок для всех"
              >
                Завершить
              </Button>
            </TooltipTrigger>
            <TooltipContent>Завершить звонок для всех</TooltipContent>
          </Tooltip>
        )}
      </div>
    </TooltipProvider>
  );
}
