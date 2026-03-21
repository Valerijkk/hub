import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { formatRelativeTime } from "@/lib/format-time";
import { api } from "@/api/client";

interface CallRecord {
  id: string;
  channel_id: string;
  channel_name: string;
  status: "active" | "ended";
  started_at: string;
  ended_at?: string;
  participant_count: number;
}

/**
 * Страница "Созвоны" — список активных и недавних звонков.
 * Аналог раздела звонков в Tinkoff Talk / Telegram.
 */
function CallsPage() {
  const { t } = useTranslation("calls");

  const { data: activeCalls } = useQuery({
    queryKey: ["calls", "active"],
    queryFn: () => api.get<CallRecord[]>("/calls/active"),
    refetchInterval: 5000,
  });

  const { data: scheduledCalls } = useQuery({
    queryKey: ["calls", "scheduled"],
    queryFn: () => api.get<CallRecord[]>("/calls/scheduled"),
  });

  return (
    <div className="mx-auto max-w-4xl p-6 space-y-8">
      {/* Заголовок */}
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-violet-500/10 text-violet-500">
          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
          </svg>
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {t("title", { defaultValue: "Созвоны" })}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle", { defaultValue: "Активные звонки и запланированные встречи" })}
          </p>
        </div>
      </div>

      {/* Активные звонки */}
      <section>
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-3">
          {t("active", { defaultValue: "Активные" })}
        </h2>
        {activeCalls && activeCalls.length > 0 ? (
          <div className="space-y-2">
            {activeCalls.map((call) => (
              <div
                key={call.id}
                className="flex items-center justify-between rounded-xl border bg-card p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-center gap-3">
                  <div className="relative">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-green-500/10 text-green-500">
                      <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                    </div>
                    <span className="absolute -right-0.5 -top-0.5 h-3 w-3 rounded-full bg-green-500 ring-2 ring-card animate-pulse" />
                  </div>
                  <div>
                    <p className="font-medium">{call.channel_name}</p>
                    <p className="text-xs text-muted-foreground">
                      {call.participant_count} {t("participants", { defaultValue: "участников" })} · {formatRelativeTime(call.started_at)}
                    </p>
                  </div>
                </div>
                <button className="rounded-lg bg-green-500 px-4 py-2 text-sm font-medium text-white hover:bg-green-600 transition-colors">
                  {t("join", { defaultValue: "Присоединиться" })}
                </button>
              </div>
            ))}
          </div>
        ) : (
          <div className="rounded-xl border border-dashed p-8 text-center">
            <p className="text-sm text-muted-foreground">
              {t("noActive", { defaultValue: "Нет активных звонков" })}
            </p>
          </div>
        )}
      </section>

      {/* Запланированные */}
      <section>
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-3">
          {t("scheduled", { defaultValue: "Запланированные" })}
        </h2>
        {scheduledCalls && scheduledCalls.length > 0 ? (
          <div className="space-y-2">
            {scheduledCalls.map((call) => (
              <div
                key={call.id}
                className="flex items-center justify-between rounded-xl border bg-card p-4"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-violet-500/10 text-violet-500">
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium">{call.channel_name}</p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(call.started_at).toLocaleString("ru-RU", {
                        day: "numeric",
                        month: "short",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="rounded-xl border border-dashed p-8 text-center">
            <p className="text-sm text-muted-foreground">
              {t("noScheduled", { defaultValue: "Нет запланированных звонков" })}
            </p>
          </div>
        )}
      </section>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/calls")({
  component: CallsPage,
});
