import { useParticipants, useIsSpeaking } from "@livekit/components-react";
import type { Participant } from "livekit-client";
import { Hand, MicOff, Monitor } from "lucide-react";

import { cn } from "@/lib/utils";

/** Метаданные участника из LiveKit. */
interface ParticipantMeta {
  /** Рука поднята. */
  hand_raised?: boolean;
}

/**
 * Безопасный парсинг метаданных участника.
 * @param metadata - строка JSON из LiveKit
 */
function parseMeta(metadata?: string): ParticipantMeta {
  if (!metadata) return {};
  try {
    return JSON.parse(metadata) as ParticipantMeta;
  } catch {
    return {};
  }
}

/**
 * Карточка одного участника звонка.
 * Показывает аватар, имя, индикатор говорения (пульсирующая рамка),
 * статус микрофона и поднятую руку.
 */
function ParticipantCard({ participant }: { participant: Participant }) {
  const isSpeaking = useIsSpeaking(participant);
  const meta = parseMeta(participant.metadata);
  const name = participant.name || participant.identity;

  return (
    <div
      className={cn(
        "relative flex flex-col items-center justify-center gap-2 rounded-xl p-4 transition-all duration-300",
        "bg-white/5 border border-white/10",
        isSpeaking && "ring-2 ring-green-500 shadow-[0_0_20px_rgba(34,197,94,0.15)]",
      )}
    >
      {/* Аватар с индикатором говорения */}
      <div className="relative">
        <div
          className={cn(
            "flex h-16 w-16 items-center justify-center rounded-full text-xl font-semibold transition-all duration-300",
            isSpeaking
              ? "bg-green-500/20 text-green-400 ring-2 ring-green-500/50 animate-pulse-subtle"
              : "bg-white/10 text-white/80",
          )}
        >
          {name.charAt(0).toUpperCase()}
        </div>

        {/* Индикатор микрофона на аватаре */}
        {!participant.isMicrophoneEnabled && (
          <span className="absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 shadow-md">
            <MicOff className="h-3 w-3 text-white" />
          </span>
        )}

        {/* Поднятая рука */}
        {meta.hand_raised && (
          <span className="absolute -top-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full bg-amber-500 shadow-md text-xs animate-bounce">
            <Hand className="h-3 w-3 text-white" />
          </span>
        )}
      </div>

      {/* Имя участника */}
      <span className="text-xs font-medium text-white/80 truncate max-w-full text-center">
        {name}
      </span>

      {/* Индикаторы */}
      <div className="flex items-center gap-1.5">
        {participant.isScreenShareEnabled && (
          <Monitor className="h-3.5 w-3.5 text-blue-400" />
        )}
        {isSpeaking && (
          <span className="text-[10px] text-green-400 font-medium">
            говорит
          </span>
        )}
      </div>
    </div>
  );
}

/**
 * Сетка участников звонка в стиле Tinkoff Talk.
 * Адаптивная сетка: 2 колонки для 2-4 человек, 3 колонки для 5+.
 * Участники с поднятой рукой отображаются первыми.
 */
export function ParticipantList() {
  const participants = useParticipants();

  /** Сортировка: поднявшие руку сверху, затем по имени. */
  const sorted = [...participants].sort((a, b) => {
    const aMeta = parseMeta(a.metadata);
    const bMeta = parseMeta(b.metadata);
    if (aMeta.hand_raised && !bMeta.hand_raised) return -1;
    if (!aMeta.hand_raised && bMeta.hand_raised) return 1;
    return (a.name ?? a.identity).localeCompare(b.name ?? b.identity);
  });

  const gridCols =
    participants.length <= 1
      ? "grid-cols-1"
      : participants.length <= 4
        ? "grid-cols-2"
        : "grid-cols-3";

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="flex items-center justify-between px-2 py-1 mb-2">
        <h3 className="text-xs font-medium text-white/50 uppercase tracking-wider">
          Участники ({participants.length})
        </h3>
      </div>
      <div className={cn("grid gap-2", gridCols)}>
        {sorted.map((p) => (
          <ParticipantCard key={p.identity} participant={p} />
        ))}
      </div>
    </div>
  );
}
