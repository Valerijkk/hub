import { useCallback, useRef } from "react";
import {
  CheckCircle2,
  MessageSquare,
  Phone,
  TestTube,
  AlertTriangle,
  Clock,
  Link2,
  AtSign,
  ClipboardList,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { useNotificationStore } from "@/stores/notification-store";
import type { NotificationDTO } from "@/api/notifications";

/** Карта иконок по типу уведомления. */
const TYPE_ICONS: Record<string, React.ElementType> = {
  assigned: ClipboardList,
  mentioned: AtSign,
  status_changed: CheckCircle2,
  comment_added: MessageSquare,
  message_received: MessageSquare,
  call_started: Phone,
  run_completed: TestTube,
  deadline_approaching: Clock,
  blocker_added: Link2,
};

/** Русские названия типов уведомлений. */
const TYPE_LABELS: Record<string, string> = {
  assigned: "Назначение",
  mentioned: "Упоминание",
  status_changed: "Смена статуса",
  comment_added: "Комментарий",
  message_received: "Сообщение",
  call_started: "Звонок",
  run_completed: "Прогон завершён",
  deadline_approaching: "Дедлайн",
  blocker_added: "Блокировщик",
};

/**
 * Форматирует относительное время (простая реализация без date-fns).
 */
function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMin < 1) return "только что";
  if (diffMin < 60) return `${diffMin} мин. назад`;
  if (diffHours < 24) return `${diffHours} ч. назад`;
  if (diffDays < 7) return `${diffDays} дн. назад`;
  return date.toLocaleDateString("ru-RU");
}

/**
 * NotificationPanel — содержимое панели уведомлений.
 * Показывает список уведомлений с возможностью пометить прочитанным.
 */
export function NotificationPanel() {
  const notifications = useNotificationStore((s) => s.notifications);
  const loading = useNotificationStore((s) => s.loading);
  const hasMore = useNotificationStore((s) => s.hasMore);
  const markRead = useNotificationStore((s) => s.markRead);
  const markAllRead = useNotificationStore((s) => s.markAllRead);
  const fetchMore = useNotificationStore((s) => s.fetchMore);

  const scrollRef = useRef<HTMLDivElement>(null);

  /** Обработчик прокрутки для подгрузки следующей страницы. */
  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const target = e.currentTarget;
      if (
        target.scrollHeight - target.scrollTop - target.clientHeight < 100 &&
        hasMore &&
        !loading
      ) {
        void fetchMore();
      }
    },
    [hasMore, loading, fetchMore],
  );

  /** Обработчик клика по уведомлению. */
  const handleItemClick = useCallback(
    (notif: NotificationDTO) => {
      if (!notif.is_read) {
        void markRead(notif.id);
      }
      // TODO: навигация к сущности (issue, channel, testrun).
    },
    [markRead],
  );

  return (
    <div className="flex flex-col">
      {/* Заголовок */}
      <div className="flex items-center justify-between px-4 py-3">
        <h3 className="text-sm font-semibold">Уведомления</h3>
        <Button
          variant="ghost"
          size="sm"
          className="h-auto px-2 py-1 text-xs text-muted-foreground"
          onClick={() => void markAllRead()}
        >
          Прочитать все
        </Button>
      </div>
      <Separator />

      {/* Список уведомлений */}
      <ScrollArea className="h-96" ref={scrollRef} onScrollCapture={handleScroll}>
        {notifications.length === 0 && !loading ? (
          <div className="flex h-32 items-center justify-center text-sm text-muted-foreground">
            Нет уведомлений
          </div>
        ) : (
          <div className="flex flex-col">
            {notifications.map((notif) => (
              <NotificationItem
                key={notif.id}
                notification={notif}
                onClick={() => handleItemClick(notif)}
              />
            ))}
            {loading && (
              <div className="flex items-center justify-center py-4 text-xs text-muted-foreground">
                Загрузка...
              </div>
            )}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}

/** Один элемент уведомления. */
function NotificationItem({
  notification,
  onClick,
}: {
  notification: NotificationDTO;
  onClick: () => void;
}) {
  const Icon = TYPE_ICONS[notification.type] ?? AlertTriangle;
  const typeLabel = TYPE_LABELS[notification.type] ?? notification.type;

  return (
    <button
      type="button"
      className={`flex w-full gap-3 px-4 py-3 text-left transition-colors hover:bg-muted/50 ${
        !notification.is_read ? "bg-muted/30" : ""
      }`}
      onClick={onClick}
    >
      {/* Иконка */}
      <div className="mt-0.5 flex-shrink-0">
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>

      {/* Содержимое */}
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm font-medium leading-tight">
            <span className="mr-1 text-xs text-muted-foreground">{typeLabel}</span>
            {notification.title}
          </p>
          {/* Индикатор непрочитанного */}
          {!notification.is_read && (
            <span className="mt-1 h-2 w-2 flex-shrink-0 rounded-full bg-blue-500" />
          )}
        </div>
        {notification.body && (
          <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
            {notification.body}
          </p>
        )}
        <p className="mt-1 text-xs text-muted-foreground">
          {formatRelativeTime(notification.created_at)}
        </p>
      </div>
    </button>
  );
}
