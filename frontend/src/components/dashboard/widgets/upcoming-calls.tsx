/**
 * Виджет "Предстоящие звонки" -- список запланированных звонков.
 */
import type { PersonalData, CallSummary } from "@/api/dashboard";
import { Button } from "@/components/ui/button";

interface Props {
  data?: unknown;
}

/** Виджет предстоящих звонков. */
export function UpcomingCallsWidget({ data }: Props) {
  const personalData = data as PersonalData | undefined;
  const calls = personalData?.upcoming_calls ?? [];

  if (calls.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">Нет предстоящих звонков</p>
    );
  }

  return (
    <div className="space-y-2">
      {calls.map((call: CallSummary) => (
        <div
          key={call.id}
          className="flex items-center justify-between gap-2 rounded-md border p-2 text-sm"
        >
          <div className="min-w-0">
            <p className="truncate font-medium">{call.title}</p>
            <p className="text-xs text-muted-foreground">
              {call.channel_name} &middot;{" "}
              {new Date(call.scheduled_at).toLocaleString("ru-RU", {
                hour: "2-digit",
                minute: "2-digit",
                day: "numeric",
                month: "short",
              })}
            </p>
          </div>
          <Button variant="outline" size="sm">
            Войти
          </Button>
        </div>
      ))}
    </div>
  );
}
