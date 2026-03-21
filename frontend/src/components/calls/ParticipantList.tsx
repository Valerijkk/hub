import { useParticipants } from "@livekit/components-react";
import { Hand, Mic, MicOff, Monitor } from "lucide-react";

import { Card } from "@/components/ui/card";

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
 * Список участников звонка с индикаторами.
 * Показывает статус микрофона, демонстрации экрана и поднятой руки.
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

  return (
    <div className="flex w-60 shrink-0 flex-col gap-2 overflow-y-auto">
      <h3 className="px-1 text-sm font-medium text-muted-foreground">
        Участники ({participants.length})
      </h3>
      {sorted.map((p) => {
        const meta = parseMeta(p.metadata);
        const name = p.name || p.identity;
        return (
          <Card key={p.identity} className="flex items-center gap-2 px-3 py-2">
            {/* Аватар-заглушка с первой буквой имени */}
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-medium text-primary">
              {name.charAt(0).toUpperCase()}
            </div>

            {/* Имя участника */}
            <span className="min-w-0 flex-1 truncate text-sm">{name}</span>

            {/* Индикаторы */}
            <div className="flex shrink-0 items-center gap-1">
              {meta.hand_raised && (
                <Hand className="h-3.5 w-3.5 text-amber-500" />
              )}
              {p.isScreenShareEnabled && (
                <Monitor className="h-3.5 w-3.5 text-blue-500" />
              )}
              {p.isMicrophoneEnabled ? (
                <Mic className="h-3.5 w-3.5 text-muted-foreground" />
              ) : (
                <MicOff className="h-3.5 w-3.5 text-destructive" />
              )}
            </div>
          </Card>
        );
      })}
    </div>
  );
}
