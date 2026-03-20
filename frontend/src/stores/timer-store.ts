import { create } from "zustand";
import { persist } from "zustand/middleware";

/** Состояние стора таймера для трекинга времени. */
interface TimerState {
  /** Запущен ли таймер. */
  isRunning: boolean;
  /** Ключ задачи, на которой запущен таймер (например, "HUB-42"). */
  activeIssueKey: string | null;
  /** ID проекта для активного таймера. */
  activeProjectId: string | null;
  /** Временная метка запуска таймера (ISO-строка). */
  startedAt: string | null;
  /** Прошедшее время в секундах (обновляется интервалом). */
  elapsedSeconds: number;

  /** Запустить таймер на задаче. */
  startTimer: (projectId: string, issueKey: string) => void;
  /** Остановить таймер и сбросить состояние. */
  stopTimer: () => void;
  /** Обновить elapsed на основе startedAt (вызывается по интервалу). */
  tick: () => void;
}

/** Стор управления таймером трекинга времени с персистом. */
export const useTimerStore = create<TimerState>()(
  persist(
    (set, get) => ({
      isRunning: false,
      activeIssueKey: null,
      activeProjectId: null,
      startedAt: null,
      elapsedSeconds: 0,

      startTimer: (projectId, issueKey) =>
        set({
          isRunning: true,
          activeIssueKey: issueKey,
          activeProjectId: projectId,
          startedAt: new Date().toISOString(),
          elapsedSeconds: 0,
        }),

      stopTimer: () =>
        set({
          isRunning: false,
          activeIssueKey: null,
          activeProjectId: null,
          startedAt: null,
          elapsedSeconds: 0,
        }),

      tick: () => {
        const { startedAt, isRunning } = get();
        if (!isRunning || !startedAt) return;
        const elapsed = Math.floor(
          (Date.now() - new Date(startedAt).getTime()) / 1000,
        );
        set({ elapsedSeconds: elapsed });
      },
    }),
    { name: "hub-timer" },
  ),
);
