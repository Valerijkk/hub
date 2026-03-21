import { Loader2 } from "lucide-react";

import { useCallStore } from "@/stores/call-store";
import { CallRoom } from "./CallRoom";

/**
 * Обёртка звонка, отображается когда пользователь в звонке.
 * Показывает CallRoom с токеном из стора или индикатор загрузки.
 * Позиционируется фиксированно в правом нижнем углу экрана.
 */
export function CallPanel() {
  const { isInCall, isConnecting, token, leaveCall, error, setError } =
    useCallStore();

  /** Не показывать панель если не в звонке, не подключаемся и нет ошибки. */
  if (!isInCall && !isConnecting && !error) return null;

  /** Обработчик отключения. */
  const handleDisconnect = () => {
    void leaveCall();
  };

  return (
    <div className="fixed bottom-0 right-0 z-50 flex h-[500px] w-[420px] flex-col overflow-hidden rounded-tl-xl border bg-background shadow-2xl">
      {/* Ошибка */}
      {error && (
        <div className="flex items-center justify-between bg-destructive/10 px-4 py-2 text-sm text-destructive">
          <span>{error}</span>
          <button
            onClick={() => setError(null)}
            className="ml-2 text-xs underline"
          >
            Закрыть
          </button>
        </div>
      )}

      {/* Загрузка */}
      {isConnecting && (
        <div className="flex flex-1 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <span className="ml-3 text-sm text-muted-foreground">
            Подключение к звонку...
          </span>
        </div>
      )}

      {/* Комната звонка */}
      {isInCall && token && (
        <CallRoom token={token} onDisconnect={handleDisconnect} />
      )}
    </div>
  );
}
