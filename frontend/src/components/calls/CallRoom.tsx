import {
  LiveKitRoom,
  RoomAudioRenderer,
} from "@livekit/components-react";
import "@livekit/components-styles";

import { CallControls } from "./CallControls";
import { ParticipantList } from "./ParticipantList";
import { ScreenShareView } from "./ScreenShareView";
import { CallChat } from "./CallChat";
import { useCallStore } from "@/stores/call-store";

/** URL сервера LiveKit из переменных окружения. */
const LIVEKIT_URL =
  (import.meta.env.VITE_LIVEKIT_URL as string) || "ws://localhost:7880";

/** Пропсы компонента CallRoom. */
interface CallRoomProps {
  /** JWT-токен для подключения к LiveKit. */
  token: string;
  /** Коллбэк при отключении от звонка. */
  onDisconnect: () => void;
}

/**
 * Главный контейнер звонка.
 * Оборачивает LiveKitRoom и отображает список участников,
 * демонстрацию экрана и панель управления.
 */
export function CallRoom({ token, onDisconnect }: CallRoomProps) {
  const activeCall = useCallStore((s) => s.activeCall);

  return (
    <LiveKitRoom
      token={token}
      serverUrl={LIVEKIT_URL}
      connect={true}
      audio={true}
      video={false}
      onDisconnected={onDisconnect}
      data-lk-theme="default"
      className="flex h-full w-full flex-col"
    >
      {/* Воспроизведение аудио всех участников */}
      <RoomAudioRenderer />

      {/* Основная область: участники + демонстрация экрана + чат */}
      <div className="flex flex-1 gap-4 overflow-hidden p-4">
        <ParticipantList />
        <ScreenShareView />
        {activeCall && <CallChat callId={activeCall.id} />}
      </div>

      {/* Панель управления звонком */}
      <CallControls onDisconnect={onDisconnect} />
    </LiveKitRoom>
  );
}
