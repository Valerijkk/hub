import { createFileRoute } from "@tanstack/react-router";

/**
 * Дефолтная страница мессенджера.
 * Показывается когда канал не выбран.
 */
function MessengerIndex() {
  return (
    <div className="flex h-full flex-1 items-center justify-center">
      <div className="text-center">
        <svg
          className="mx-auto mb-4 h-16 w-16 text-muted-foreground/50"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
          />
        </svg>
        <h2 className="text-lg font-medium text-muted-foreground">
          Выберите канал
        </h2>
        <p className="mt-1 text-sm text-muted-foreground/70">
          Выберите канал слева или создайте новый, чтобы начать общение
        </p>
      </div>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/messenger/")({
  component: MessengerIndex,
});
