import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  startTimer as apiStartTimer,
  stopTimer as apiStopTimer,
  logTime,
  listTimeLogs,
} from "@/api/tracker";
import { useTimerStore } from "@/stores/timer-store";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

/** Пропсы компонента TimeTracker. */
interface TimeTrackerProps {
  /** ID проекта. */
  projectId: string;
  /** Ключ задачи (например, "HUB-42"). */
  issueKey: string;
  /** Оценочное время в минутах. */
  estimateMinutes?: number;
  /** Уже затраченное время в минутах. */
  timeSpentMinutes?: number;
}

/**
 * Компонент трекинга времени для задачи.
 * Поддерживает Start/Stop таймер, ручное логирование и список записей.
 * Использует useTimerStore для глобального состояния таймера с персистом.
 */
export function TimeTracker({
  projectId,
  issueKey,
  estimateMinutes,
  timeSpentMinutes,
}: TimeTrackerProps) {
  const queryClient = useQueryClient();
  const timer = useTimerStore();

  const isTimerForThisIssue =
    timer.isRunning &&
    timer.activeIssueKey === issueKey &&
    timer.activeProjectId === projectId;

  // Тик таймера каждую секунду
  useEffect(() => {
    if (!isTimerForThisIssue) return;
    const interval = setInterval(() => {
      timer.tick();
    }, 1000);
    return () => clearInterval(interval);
  }, [isTimerForThisIssue, timer]);

  // Логи времени
  const { data: timeLogs } = useQuery({
    queryKey: ["timeLogs", projectId, issueKey],
    queryFn: () => listTimeLogs(projectId, issueKey),
  });

  // Мутации
  const startMutation = useMutation({
    mutationFn: () => apiStartTimer(projectId, issueKey),
    onSuccess: () => {
      timer.startTimer(projectId, issueKey);
      toast.success("Таймер запущен");
    },
    onError: () => toast.error("Не удалось запустить таймер"),
  });

  const stopMutation = useMutation({
    mutationFn: () => apiStopTimer(projectId, issueKey),
    onSuccess: () => {
      timer.stopTimer();
      void queryClient.invalidateQueries({
        queryKey: ["timeLogs", projectId, issueKey],
      });
      void queryClient.invalidateQueries({
        queryKey: ["issue", projectId, issueKey],
      });
      toast.success("Таймер остановлен");
    },
    onError: () => toast.error("Не удалось остановить таймер"),
  });

  // Ручное логирование
  const [manualMinutes, setManualMinutes] = useState("");
  const [manualDescription, setManualDescription] = useState("");

  const logMutation = useMutation({
    mutationFn: () =>
      logTime(projectId, issueKey, {
        minutes: Number(manualMinutes),
        description: manualDescription || undefined,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["timeLogs", projectId, issueKey],
      });
      void queryClient.invalidateQueries({
        queryKey: ["issue", projectId, issueKey],
      });
      setManualMinutes("");
      setManualDescription("");
      toast.success("Время залогировано");
    },
    onError: () => toast.error("Не удалось залогировать время"),
  });

  const handleLogTime = () => {
    if (!manualMinutes || Number(manualMinutes) <= 0) {
      toast.error("Введите количество минут");
      return;
    }
    logMutation.mutate();
  };

  return (
    <div className="space-y-4">
      {/* Сводка времени */}
      <div className="grid grid-cols-2 gap-3">
        <div className="rounded-md border p-3 text-center">
          <p className="text-xs text-muted-foreground">Оценка</p>
          <p className="text-lg font-bold">
            {estimateMinutes ? formatMinutes(estimateMinutes) : "—"}
          </p>
        </div>
        <div className="rounded-md border p-3 text-center">
          <p className="text-xs text-muted-foreground">Потрачено</p>
          <p className="text-lg font-bold">
            {timeSpentMinutes ? formatMinutes(timeSpentMinutes) : "0м"}
          </p>
        </div>
      </div>

      {/* Таймер */}
      <div className="flex items-center gap-3">
        {isTimerForThisIssue ? (
          <>
            <div className="flex-1 text-center text-2xl font-mono font-bold text-primary">
              {formatSeconds(timer.elapsedSeconds)}
            </div>
            <Button
              variant="destructive"
              onClick={() => stopMutation.mutate()}
              disabled={stopMutation.isPending}
            >
              Стоп
            </Button>
          </>
        ) : (
          <Button
            className="w-full"
            onClick={() => startMutation.mutate()}
            disabled={
              startMutation.isPending ||
              (timer.isRunning && !isTimerForThisIssue)
            }
          >
            {timer.isRunning && !isTimerForThisIssue
              ? `Таймер активен на ${timer.activeIssueKey}`
              : "Запустить таймер"}
          </Button>
        )}
      </div>

      {/* Ручное логирование */}
      <div className="space-y-2 border-t pt-3">
        <p className="text-sm font-medium">Залогировать время</p>
        <div className="flex gap-2">
          <div className="flex-1">
            <Label htmlFor="log_minutes" className="sr-only">
              Минуты
            </Label>
            <Input
              id="log_minutes"
              type="number"
              min={1}
              placeholder="Минуты"
              value={manualMinutes}
              onChange={(e) => setManualMinutes(e.target.value)}
            />
          </div>
          <div className="flex-1">
            <Label htmlFor="log_desc" className="sr-only">
              Описание
            </Label>
            <Input
              id="log_desc"
              placeholder="Описание"
              value={manualDescription}
              onChange={(e) => setManualDescription(e.target.value)}
            />
          </div>
          <Button
            size="sm"
            onClick={handleLogTime}
            disabled={logMutation.isPending}
          >
            Добавить
          </Button>
        </div>
      </div>

      {/* Список логов */}
      {timeLogs && timeLogs.length > 0 && (
        <div className="space-y-2 border-t pt-3">
          <p className="text-sm font-medium">История</p>
          {timeLogs.map((log) => (
            <div
              key={log.id}
              className="flex items-center justify-between rounded-md border p-2"
            >
              <div className="flex items-center gap-2">
                <Avatar className="h-5 w-5">
                  <AvatarImage src={log.user.avatar_url} />
                  <AvatarFallback className="text-[10px]">
                    {log.user.name.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="text-sm">{log.user.name}</span>
                {log.description && (
                  <span className="text-xs text-muted-foreground">
                    — {log.description}
                  </span>
                )}
              </div>
              <span className="text-sm font-medium">
                {formatMinutes(log.minutes)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/** Форматирует минуты в строку вида "1ч 30м". */
function formatMinutes(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}м`;
  if (m === 0) return `${h}ч`;
  return `${h}ч ${m}м`;
}

/** Форматирует секунды в строку вида "MM:SS". */
function formatSeconds(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}
