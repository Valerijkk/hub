import { useState } from "react";
import {
  createFileRoute,
  Link,
  useNavigate,
} from "@tanstack/react-router";
import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import type { Value } from "@udecode/plate";
import {
  getIssue,
  updateIssue,
  deleteIssue,
  listWorkflowStatuses,
} from "@/api/tracker";
import { PlateEditor } from "@/components/tracker/plate-editor";
import { CommentList } from "@/components/tracker/comment-list";
import { ActivityFeed } from "@/components/tracker/activity-feed";
import { TimeTracker } from "@/components/tracker/time-tracker";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

/** Конфигурация иконок типов задач. */
const TYPE_ICONS: Record<string, string> = {
  bug: "\u{1F41B}",
  task: "\u{2705}",
  story: "\u{1F4D6}",
  epic: "\u{26A1}",
  sub_task: "\u{1F4CB}",
  improvement: "\u{1F680}",
  tech_debt: "\u{1F527}",
  spike: "\u{1F50D}",
  release: "\u{1F3F7}\uFE0F",
};

/** Конфигурация цветов приоритетов. */
const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-500 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-500 text-black",
  low: "bg-blue-500 text-white",
  trivial: "bg-gray-400 text-white",
};

/**
 * Страница детальной информации о задаче.
 * Двухколоночная раскладка с inline-редактированием, комментариями,
 * лентой активности и трекером времени.
 */
function IssueDetailPage() {
  const { projectId, issueKey } = Route.useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const [editTitle, setEditTitle] = useState("");
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [editDescription, setEditDescription] = useState<Value | undefined>();

  const { data: issue, isLoading } = useQuery({
    queryKey: ["issue", projectId, issueKey],
    queryFn: () => getIssue(projectId, issueKey),
  });

  const { data: statuses } = useQuery({
    queryKey: ["workflow-statuses", projectId],
    queryFn: () => listWorkflowStatuses(projectId),
  });

  const updateMutation = useMutation({
    mutationFn: (fields: Record<string, unknown>) => {
      if (!issue) throw new Error("No issue data");
      return updateIssue(projectId, issueKey, {
        version: issue.version,
        ...fields,
      } as Parameters<typeof updateIssue>[2]);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["issue", projectId, issueKey],
      });
      void queryClient.invalidateQueries({
        queryKey: ["board", projectId],
      });
      toast.success("Задача обновлена");
    },
    onError: () => toast.error("Не удалось обновить задачу"),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteIssue(projectId, issueKey),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["issues", projectId],
      });
      void queryClient.invalidateQueries({
        queryKey: ["board", projectId],
      });
      toast.success("Задача удалена");
      void navigate({
        to: "/projects/$projectId/issues",
        params: { projectId },
      });
    },
    onError: () => toast.error("Не удалось удалить задачу"),
  });

  const saveTitle = () => {
    if (editTitle.trim() && editTitle !== issue?.title) {
      updateMutation.mutate({ title: editTitle.trim() });
    }
    setIsEditingTitle(false);
  };

  const saveDescription = () => {
    updateMutation.mutate({
      description_json: editDescription as Record<string, unknown> | undefined,
    });
    setIsEditingDescription(false);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    );
  }

  if (!issue) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Задача не найдена</p>
      </div>
    );
  }

  return (
    <div>
      {/* Хлебные крошки */}
      <nav className="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
        <Link
          to="/projects/$projectId"
          params={{ projectId }}
          className="hover:text-foreground"
        >
          Проект
        </Link>
        <span>/</span>
        <Link
          to="/projects/$projectId/issues"
          params={{ projectId }}
          className="hover:text-foreground"
        >
          Задачи
        </Link>
        <span>/</span>
        <span className="text-foreground">{issue.issue_key}</span>
      </nav>

      {/* Заголовок с типом и ключом */}
      <div className="mb-6 flex items-start justify-between">
        <div className="flex items-center gap-3">
          <span className="text-2xl">
            {TYPE_ICONS[issue.issue_type] ?? "\u{1F4CB}"}
          </span>
          <span className="text-lg font-bold text-muted-foreground">
            {issue.issue_key}
          </span>
        </div>
        <Button
          variant="destructive"
          size="sm"
          onClick={() => {
            if (window.confirm("Удалить задачу?")) {
              deleteMutation.mutate();
            }
          }}
          disabled={deleteMutation.isPending}
        >
          Удалить
        </Button>
      </div>

      {/* Двухколоночная раскладка */}
      <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
        {/* Левая колонка: основное */}
        <div className="space-y-6">
          {/* Название */}
          {isEditingTitle ? (
            <div className="flex gap-2">
              <Input
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") saveTitle();
                  if (e.key === "Escape") setIsEditingTitle(false);
                }}
                autoFocus
                className="text-xl font-bold"
              />
              <Button size="sm" onClick={saveTitle}>
                Сохранить
              </Button>
            </div>
          ) : (
            <h1
              className="cursor-pointer text-xl font-bold hover:bg-muted/50"
              onClick={() => {
                setEditTitle(issue.title);
                setIsEditingTitle(true);
              }}
            >
              {issue.title}
            </h1>
          )}

          {/* Описание */}
          <div>
            <Label className="mb-2 block text-sm font-medium">Описание</Label>
            {isEditingDescription ? (
              <div className="space-y-2">
                <PlateEditor
                  value={editDescription}
                  onChange={setEditDescription}
                  placeholder="Описание задачи..."
                />
                <div className="flex gap-2">
                  <Button size="sm" onClick={saveDescription}>
                    Сохранить
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setIsEditingDescription(false)}
                  >
                    Отмена
                  </Button>
                </div>
              </div>
            ) : (
              <div
                className="cursor-pointer rounded-md p-2 hover:bg-muted/50"
                onClick={() => {
                  setEditDescription(
                    issue.description_json as Value | undefined,
                  );
                  setIsEditingDescription(true);
                }}
              >
                {issue.description_json ? (
                  <PlateEditor
                    value={issue.description_json as unknown as Value}
                    readOnly
                    className="border-none p-0"
                  />
                ) : (
                  <p className="text-sm text-muted-foreground">
                    Нажмите, чтобы добавить описание
                  </p>
                )}
              </div>
            )}
          </div>

          <Separator />

          {/* Вкладки: Комментарии | Активность | Время */}
          <Tabs defaultValue="comments">
            <TabsList>
              <TabsTrigger value="comments">Комментарии</TabsTrigger>
              <TabsTrigger value="activity">Активность</TabsTrigger>
              <TabsTrigger value="time">Время</TabsTrigger>
            </TabsList>

            <TabsContent value="comments" className="mt-4">
              <CommentList projectId={projectId} issueKey={issueKey} />
            </TabsContent>

            <TabsContent value="activity" className="mt-4">
              <ActivityFeed projectId={projectId} issueKey={issueKey} />
            </TabsContent>

            <TabsContent value="time" className="mt-4">
              <TimeTracker
                projectId={projectId}
                issueKey={issueKey}
                estimateMinutes={issue.estimate_minutes}
                timeSpentMinutes={issue.time_spent_minutes}
              />
            </TabsContent>
          </Tabs>
        </div>

        {/* Правая колонка: свойства */}
        <div className="space-y-4">
          {/* Статус */}
          <div className="space-y-1">
            <Label className="text-xs text-muted-foreground">Статус</Label>
            <select
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              value={issue.status.id}
              onChange={(e) =>
                updateMutation.mutate({ status_id: e.target.value })
              }
            >
              {statuses?.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
              {!statuses && (
                <option value={issue.status.id}>{issue.status.name}</option>
              )}
            </select>
          </div>

          {/* Приоритет */}
          <div className="space-y-1">
            <Label className="text-xs text-muted-foreground">Приоритет</Label>
            <select
              className="w-full rounded-md border bg-background px-3 py-2 text-sm"
              value={issue.priority}
              onChange={(e) =>
                updateMutation.mutate({ priority: e.target.value })
              }
            >
              <option value="critical">Критический</option>
              <option value="high">Высокий</option>
              <option value="medium">Средний</option>
              <option value="low">Низкий</option>
              <option value="trivial">Тривиальный</option>
            </select>
            <Badge
              variant="secondary"
              className={cn("text-xs", PRIORITY_COLORS[issue.priority])}
            >
              {issue.priority}
            </Badge>
          </div>

          <Separator />

          {/* Исполнитель */}
          <div className="space-y-1">
            <Label className="text-xs text-muted-foreground">Исполнитель</Label>
            {issue.assignee ? (
              <div className="flex items-center gap-2">
                <Avatar className="h-6 w-6">
                  <AvatarImage src={issue.assignee.avatar_url} />
                  <AvatarFallback className="text-[10px]">
                    {issue.assignee.name.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="text-sm">{issue.assignee.name}</span>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Не назначен</p>
            )}
          </div>

          {/* Автор */}
          <div className="space-y-1">
            <Label className="text-xs text-muted-foreground">Автор</Label>
            {issue.reporter ? (
              <div className="flex items-center gap-2">
                <Avatar className="h-6 w-6">
                  <AvatarImage src={issue.reporter.avatar_url} />
                  <AvatarFallback className="text-[10px]">
                    {issue.reporter.name.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="text-sm">{issue.reporter.name}</span>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">—</p>
            )}
          </div>

          <Separator />

          {/* Метки */}
          <div className="space-y-1">
            <Label className="text-xs text-muted-foreground">Метки</Label>
            <div className="flex flex-wrap gap-1">
              {issue.labels.length > 0 ? (
                issue.labels.map((label) => (
                  <Badge
                    key={label.id}
                    variant="outline"
                    className="text-xs"
                    style={{ borderColor: label.color, color: label.color }}
                  >
                    {label.name}
                  </Badge>
                ))
              ) : (
                <p className="text-sm text-muted-foreground">Нет меток</p>
              )}
            </div>
          </div>

          <Separator />

          {/* Оценка и срок */}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Оценка</Label>
              <p className="text-sm">
                {issue.estimate_minutes
                  ? `${Math.floor(issue.estimate_minutes / 60)}ч ${issue.estimate_minutes % 60}м`
                  : "—"}
              </p>
            </div>
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Потрачено</Label>
              <p className="text-sm">
                {issue.time_spent_minutes
                  ? `${Math.floor(issue.time_spent_minutes / 60)}ч ${issue.time_spent_minutes % 60}м`
                  : "0м"}
              </p>
            </div>
          </div>

          {issue.due_date && (
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Срок</Label>
              <p className="text-sm">
                {new Date(issue.due_date).toLocaleDateString("ru-RU")}
              </p>
            </div>
          )}

          {/* Родительская задача */}
          {issue.parent && (
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">
                Родительская задача
              </Label>
              <Link
                to="/projects/$projectId/issues/$issueKey"
                params={{ projectId, issueKey: issue.parent.issue_key }}
                className="text-sm text-primary hover:underline"
              >
                {issue.parent.issue_key}: {issue.parent.title}
              </Link>
            </div>
          )}

          {/* Дочерние задачи */}
          {issue.children_count > 0 && (
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">
                Подзадачи ({issue.children_count})
              </Label>
            </div>
          )}

          <Separator />

          {/* Мета */}
          <div className="space-y-1 text-xs text-muted-foreground">
            <p>
              Создана:{" "}
              {new Date(issue.created_at).toLocaleDateString("ru-RU")}
            </p>
            <p>
              Обновлена:{" "}
              {new Date(issue.updated_at).toLocaleDateString("ru-RU")}
            </p>
            <p>Версия: {issue.version}</p>
          </div>
        </div>
      </div>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/issues/$issueKey",
)({
  component: IssueDetailPage,
});
