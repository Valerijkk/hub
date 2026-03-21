import { useState, useEffect, useCallback } from "react";
import { useSprintStore } from "@/stores/sprint-store";
import * as pokerApi from "@/api/sprint";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/** Значения Фибоначчи для голосования. */
const FIBONACCI_VALUES = [
  "0",
  "1",
  "2",
  "3",
  "5",
  "8",
  "13",
  "21",
  "34",
  "55",
  "89",
  "?",
];

/** Пропсы компонента PlanningPoker. */
interface PlanningPokerProps {
  /** ID сессии Planning Poker. */
  sessionId: string;
  /** ID текущего пользователя. */
  currentUserId: string;
  /** Может ли пользователь раскрыть голоса (создатель сессии). */
  canReveal?: boolean;
}

/**
 * PlanningPoker отображает интерфейс голосования (SCRM-05).
 * Карточки Фибоначчи, скрытые голоса до раскрытия, средний/медиана после раскрытия.
 */
export function PlanningPoker({
  sessionId,
  currentUserId,
  canReveal = false,
}: PlanningPokerProps) {
  const { pokerSession, setPokerSession } = useSprintStore();
  const [selectedValue, setSelectedValue] = useState<string | null>(null);
  const [voting, setVoting] = useState(false);

  const loadSession = useCallback(async () => {
    try {
      const session = await pokerApi.getPokerSession(sessionId);
      setPokerSession(session);
    } catch {
      // сессия истекла
    }
  }, [sessionId, setPokerSession]);

  useEffect(() => {
    loadSession();
    const interval = setInterval(loadSession, 3000);
    return () => clearInterval(interval);
  }, [loadSession]);

  const handleVote = async (value: string) => {
    setVoting(true);
    try {
      await pokerApi.castVote(sessionId, { value });
      setSelectedValue(value);
      loadSession();
    } finally {
      setVoting(false);
    }
  };

  const handleReveal = async () => {
    try {
      const revealed = await pokerApi.revealVotes(sessionId);
      setPokerSession(revealed);
    } catch {
      // ошибка раскрытия
    }
  };

  const computeStats = (votes: Record<string, string>) => {
    const numericVotes = Object.values(votes)
      .filter((v) => v !== "?")
      .map(Number)
      .filter((n) => !isNaN(n));

    if (numericVotes.length === 0) return null;

    const sum = numericVotes.reduce((a, b) => a + b, 0);
    const avg = sum / numericVotes.length;

    const sorted = [...numericVotes].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    const median =
      sorted.length % 2 !== 0
        ? (sorted[mid] ?? 0)
        : ((sorted[mid - 1] ?? 0) + (sorted[mid] ?? 0)) / 2;

    return { avg: Math.round(avg * 10) / 10, median };
  };

  if (!pokerSession) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Сессия не найдена или истекла</p>
        </CardContent>
      </Card>
    );
  }

  const stats =
    pokerSession?.revealed && pokerSession?.votes
      ? computeStats(pokerSession.votes)
      : null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Planning Poker</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Карточки для голосования */}
        {!pokerSession.revealed && (
          <div className="flex flex-wrap gap-2">
            {FIBONACCI_VALUES.map((value) => (
              <button
                key={value}
                onClick={() => handleVote(value)}
                disabled={voting}
                className={`flex h-16 w-12 items-center justify-center rounded-lg border-2 text-lg font-bold transition-all ${
                  selectedValue === value
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-gray-200 bg-white hover:border-primary hover:bg-primary/5"
                }`}
              >
                {value}
              </button>
            ))}
          </div>
        )}

        {/* Статус голосов */}
        <div className="space-y-2">
          {!pokerSession.revealed && pokerSession.voters && (
            <div className="flex flex-wrap gap-2">
              {pokerSession.voters.map((voterId) => (
                <span
                  key={voterId}
                  className="rounded-full bg-green-100 px-3 py-1 text-sm text-green-800"
                >
                  {voterId === currentUserId ? "Вы" : voterId.slice(0, 8)}
                  {" "}
                  Проголосовал
                </span>
              ))}
            </div>
          )}

          {/* Раскрытые голоса */}
          {pokerSession.revealed && pokerSession.votes && (
            <div className="space-y-3">
              <div className="flex flex-wrap gap-3">
                {Object.entries(pokerSession.votes).map(([userId, value]) => (
                  <div
                    key={userId}
                    className="flex flex-col items-center gap-1"
                  >
                    <span className="flex h-14 w-10 items-center justify-center rounded-lg border-2 border-primary bg-primary/10 text-lg font-bold">
                      {value}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {userId === currentUserId ? "Вы" : userId.slice(0, 8)}
                    </span>
                  </div>
                ))}
              </div>

              {stats && (
                <div className="flex gap-4 rounded-lg bg-muted p-3">
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Среднее:{" "}
                    </span>
                    <span className="font-bold">{stats.avg}</span>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Медиана:{" "}
                    </span>
                    <span className="font-bold">{stats.median}</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Кнопка раскрытия */}
        {canReveal && !pokerSession.revealed && (
          <Button onClick={handleReveal} variant="default">
            Раскрыть голоса
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
