import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

/**
 * Страница "Созвоны" — hub для звонков.
 * Показывает информацию о звонках и инструкцию по использованию.
 * Активные звонки начинаются из каналов мессенджера (кнопка звонка в шапке чата).
 */
function CallsPage() {
  const { t } = useTranslation("calls");

  return (
    <div className="mx-auto max-w-4xl p-6 space-y-8 animate-fade-in">
      {/* Заголовок */}
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-xl gradient-calls text-white shadow-sm">
          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
          </svg>
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {t("title", { defaultValue: "Созвоны" })}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle", { defaultValue: "Голосовые звонки и демонстрация экрана" })}
          </p>
        </div>
      </div>

      {/* Как начать звонок */}
      <section className="rounded-xl border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Как начать звонок</h2>
        <div className="grid gap-4 md:grid-cols-3">
          <div className="flex gap-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-500 text-sm font-bold">1</div>
            <div>
              <p className="font-medium text-sm">Откройте канал</p>
              <p className="text-xs text-muted-foreground mt-0.5">Перейдите в мессенджер и выберите канал или DM</p>
            </div>
          </div>
          <div className="flex gap-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-green-500/10 text-green-500 text-sm font-bold">2</div>
            <div>
              <p className="font-medium text-sm">Нажмите кнопку звонка</p>
              <p className="text-xs text-muted-foreground mt-0.5">Иконка телефона в правом верхнем углу чата</p>
            </div>
          </div>
          <div className="flex gap-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-500/10 text-violet-500 text-sm font-bold">3</div>
            <div>
              <p className="font-medium text-sm">Общайтесь!</p>
              <p className="text-xs text-muted-foreground mt-0.5">Голос, screen share, до 50 участников</p>
            </div>
          </div>
        </div>
      </section>

      {/* Возможности */}
      <section>
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground mb-3">Возможности</h2>
        <div className="grid gap-3 md:grid-cols-2">
          {[
            { icon: "🎙️", title: "Аудиозвонки", desc: "Голосовое общение до 50 участников" },
            { icon: "🖥️", title: "Screen Share", desc: "Демонстрация экрана или окна" },
            { icon: "🔴", title: "Запись", desc: "Запись звонка по кнопке с уведомлением" },
            { icon: "📅", title: "Планирование", desc: "Запланированные и повторяющиеся звонки" },
            { icon: "✋", title: "Поднять руку", desc: "Сигнал спикеру без прерывания" },
            { icon: "💬", title: "Чат звонка", desc: "Эфемерный чат — исчезает после звонка" },
          ].map((item) => (
            <div key={item.title} className="flex gap-3 rounded-xl border p-4 hover:bg-muted/50 transition-colors">
              <span className="text-2xl">{item.icon}</span>
              <div>
                <p className="font-medium text-sm">{item.title}</p>
                <p className="text-xs text-muted-foreground">{item.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/calls")({
  component: CallsPage,
});
