import { createFileRoute } from "@tanstack/react-router";

/**
 * Дефолтная страница мессенджера.
 * Показывается когда канал не выбран — красивое пустое состояние.
 */
function MessengerIndex() {
  return (
    <div className="flex h-full flex-1 items-center justify-center bg-background">
      <div className="text-center max-w-sm animate-fade-in">
        {/* Иконка */}
        <div className="mx-auto mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500/10 to-teal-500/10">
          <svg
            className="h-10 w-10 text-emerald-500/60"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={1.5}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
            />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-foreground mb-2">
          Выберите канал
        </h2>
        <p className="text-sm text-muted-foreground leading-relaxed">
          Выберите канал слева или создайте новый, чтобы начать общение с командой
        </p>
      </div>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/messenger/")({
  component: MessengerIndex,
});
