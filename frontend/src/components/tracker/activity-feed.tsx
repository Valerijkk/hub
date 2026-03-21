import { useQuery } from "@tanstack/react-query";
import {
  listActivities,
  type ActivityLogResponse,
} from "@/api/tracker";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";

/** Пропсы компонента ActivityFeed. */
interface ActivityFeedProps {
  /** ID проекта. */
  projectId: string;
  /** Ключ задачи (например, "HUB-42"). */
  issueKey: string;
}

/**
 * Лента активности задачи.
 * Отображает хронологический список изменений с аватарами и diff-ами полей.
 * Поддерживает цветовое кодирование для статусов.
 */
export function ActivityFeed({ projectId, issueKey }: ActivityFeedProps) {
  const { data: activities, isLoading } = useQuery({
    queryKey: ["activities", projectId, issueKey],
    queryFn: () => listActivities(projectId, issueKey),
  });

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">Загрузка активности...</p>;
  }

  if (!activities || activities.length === 0) {
    return <p className="text-sm text-muted-foreground">Нет активности</p>;
  }

  return (
    <div className="space-y-3">
      {activities.map((activity) => (
        <ActivityItem key={activity.id} activity={activity} />
      ))}
    </div>
  );
}

/** Пропсы элемента активности. */
interface ActivityItemProps {
  /** Данные записи лога активности. */
  activity: ActivityLogResponse;
}

/** Элемент лога активности с описанием изменений. */
function ActivityItem({ activity }: ActivityItemProps) {
  const timeAgo = getTimeAgo(activity.created_at);

  return (
    <div className="flex gap-3 rounded-md border p-3">
      <Avatar className="h-6 w-6 flex-shrink-0">
        <AvatarImage src={activity.user.avatar_url} />
        <AvatarFallback className="text-[10px]">
          {activity.user.name.charAt(0).toUpperCase()}
        </AvatarFallback>
      </Avatar>

      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{activity.user.name}</span>
          <span className="text-xs text-muted-foreground">
            {getActionLabel(activity.action)}
          </span>
          <span className="text-xs text-muted-foreground">{timeAgo}</span>
        </div>

        {/* Изменения полей */}
        {(activity.changes ?? []).length > 0 && (
          <div className="mt-1 space-y-1">
            {(activity.changes ?? []).map((change, idx) => (
              <div key={idx} className="flex items-center gap-2 text-xs">
                <span className="font-medium text-muted-foreground">
                  {getFieldLabel(change.field)}:
                </span>
                {change.field === "status" ? (
                  <>
                    <Badge variant="outline" className="text-xs">
                      {String(change.old_value ?? "—")}
                    </Badge>
                    <span className="text-muted-foreground">&rarr;</span>
                    <Badge variant="default" className="text-xs">
                      {String(change.new_value)}
                    </Badge>
                  </>
                ) : (
                  <span>
                    <span className="text-muted-foreground line-through">
                      {String(change.old_value ?? "—")}
                    </span>
                    {" \u2192 "}
                    {String(change.new_value)}
                  </span>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

/** Получить русский ярлык действия. */
function getActionLabel(action: string): string {
  const labels: Record<string, string> = {
    created: "создал задачу",
    updated: "обновил задачу",
    status_changed: "изменил статус",
    commented: "добавил комментарий",
    assigned: "назначил исполнителя",
    priority_changed: "изменил приоритет",
    deleted: "удалил",
  };
  return labels[action] ?? action;
}

/** Получить русский ярлык поля. */
function getFieldLabel(field: string): string {
  const labels: Record<string, string> = {
    status: "Статус",
    priority: "Приоритет",
    assignee: "Исполнитель",
    title: "Название",
    description: "Описание",
    estimate_minutes: "Оценка",
    due_date: "Срок",
    labels: "Метки",
  };
  return labels[field] ?? field;
}

/** Вычисляет относительное время. */
function getTimeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = Math.floor((now - then) / 1000);

  if (diff < 60) return "только что";
  if (diff < 3600) return `${Math.floor(diff / 60)} мин назад`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} ч назад`;
  if (diff < 2592000) return `${Math.floor(diff / 86400)} дн назад`;
  return new Date(dateStr).toLocaleDateString("ru-RU");
}
