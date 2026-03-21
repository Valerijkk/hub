import { useState } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  listWorkflowStatuses,
  createWorkflowStatus,
  listWorkflowTransitions,
  createWorkflowTransition,
  type WorkflowStatusResponse,
} from "@/api/tracker";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

/** Предустановленные цвета для статусов. */
const PRESET_COLORS = [
  "#6b7280",
  "#3b82f6",
  "#10b981",
  "#f59e0b",
  "#ef4444",
  "#8b5cf6",
  "#ec4899",
  "#14b8a6",
];

/** Категории статусов. */
const CATEGORIES = [
  { value: "todo" as const, label: "К выполнению" },
  { value: "in_progress" as const, label: "В работе" },
  { value: "done" as const, label: "Готово" },
];

/**
 * Страница настроек workflow проекта.
 * Управление статусами (создание, просмотр) и переходами между статусами.
 * Позволяет настраивать цвета, категории и порядок сортировки статусов.
 */
function WorkflowSettingsPage() {
  const { projectId } = Route.useParams();
  const queryClient = useQueryClient();

  // Статусы
  const { data: statuses, isLoading: statusesLoading } = useQuery({
    queryKey: ["workflow-statuses", projectId],
    queryFn: () => listWorkflowStatuses(projectId),
  });

  // Переходы
  const { data: transitions, isLoading: transitionsLoading } = useQuery({
    queryKey: ["workflow-transitions", projectId],
    queryFn: () => listWorkflowTransitions(projectId),
  });

  // Форма нового статуса
  const [statusName, setStatusName] = useState("");
  const [statusCategory, setStatusCategory] = useState<
    "todo" | "in_progress" | "done"
  >("todo");
  const [statusColor, setStatusColor] = useState(PRESET_COLORS[0]!);
  const [statusOrder, setStatusOrder] = useState("0");

  const createStatusMutation = useMutation({
    mutationFn: () =>
      createWorkflowStatus(projectId, {
        name: statusName.trim(),
        category: statusCategory,
        color: statusColor,
        sort_order: Number(statusOrder),
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["workflow-statuses", projectId],
      });
      setStatusName("");
      setStatusOrder("0");
      toast.success("Статус создан");
    },
    onError: () => toast.error("Не удалось создать статус"),
  });

  // Форма нового перехода
  const [fromStatusId, setFromStatusId] = useState("");
  const [toStatusId, setToStatusId] = useState("");

  const createTransitionMutation = useMutation({
    mutationFn: () =>
      createWorkflowTransition(projectId, {
        from_status_id: fromStatusId,
        to_status_id: toStatusId,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["workflow-transitions", projectId],
      });
      setFromStatusId("");
      setToStatusId("");
      toast.success("Переход создан");
    },
    onError: () => toast.error("Не удалось создать переход"),
  });

  const handleCreateStatus = (e: React.FormEvent) => {
    e.preventDefault();
    if (!statusName.trim()) {
      toast.error("Введите название статуса");
      return;
    }
    createStatusMutation.mutate();
  };

  const handleCreateTransition = (e: React.FormEvent) => {
    e.preventDefault();
    if (!fromStatusId || !toStatusId) {
      toast.error("Выберите оба статуса");
      return;
    }
    if (fromStatusId === toStatusId) {
      toast.error("Статусы должны отличаться");
      return;
    }
    createTransitionMutation.mutate();
  };

  /** Находит статус по ID. */
  const getStatusName = (id: string): string =>
    statuses?.find((s) => s.id === id)?.name ?? id;

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Настройки Workflow</h1>
        <Button variant="outline" asChild>
          <Link to="/projects/$projectId" params={{ projectId }}>
            Назад
          </Link>
        </Button>
      </div>

      {/* Секция статусов */}
      <Card>
        <CardHeader>
          <CardTitle>Статусы</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Список существующих */}
          {statusesLoading ? (
            <p className="text-sm text-muted-foreground">Загрузка...</p>
          ) : (
            <div className="space-y-2">
              {statuses?.map((status) => (
                <StatusRow key={status.id} status={status} />
              ))}
              {statuses?.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  Нет статусов. Создайте первый!
                </p>
              )}
            </div>
          )}

          <Separator />

          {/* Форма добавления */}
          <form onSubmit={handleCreateStatus} className="space-y-3">
            <p className="text-sm font-medium">Добавить статус</p>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <Label htmlFor="status_name">Название</Label>
                <Input
                  id="status_name"
                  value={statusName}
                  onChange={(e) => setStatusName(e.target.value)}
                  placeholder="Название статуса"
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="status_category">Категория</Label>
                <select
                  id="status_category"
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                  value={statusCategory}
                  onChange={(e) =>
                    setStatusCategory(
                      e.target.value as "todo" | "in_progress" | "done",
                    )
                  }
                >
                  {CATEGORIES.map((c) => (
                    <option key={c.value} value={c.value}>
                      {c.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <Label>Цвет</Label>
                <div className="flex gap-1">
                  {PRESET_COLORS.map((color) => (
                    <button
                      key={color}
                      type="button"
                      className={`h-6 w-6 rounded-full border-2 ${statusColor === color ? "border-foreground" : "border-transparent"}`}
                      style={{ backgroundColor: color }}
                      onClick={() => setStatusColor(color)}
                    />
                  ))}
                </div>
              </div>
              <div className="space-y-1">
                <Label htmlFor="sort_order">Порядок</Label>
                <Input
                  id="sort_order"
                  type="number"
                  value={statusOrder}
                  onChange={(e) => setStatusOrder(e.target.value)}
                />
              </div>
            </div>
            <Button
              type="submit"
              size="sm"
              disabled={createStatusMutation.isPending}
            >
              {createStatusMutation.isPending
                ? "Создание..."
                : "Создать статус"}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Секция переходов */}
      <Card>
        <CardHeader>
          <CardTitle>Переходы</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {transitionsLoading ? (
            <p className="text-sm text-muted-foreground">Загрузка...</p>
          ) : (
            <div className="space-y-2">
              {transitions?.map((t) => (
                <div
                  key={t.id}
                  className="flex items-center gap-2 rounded-md border p-2 text-sm"
                >
                  <Badge variant="outline">{getStatusName(t.from_status_id)}</Badge>
                  <span className="text-muted-foreground">&rarr;</span>
                  <Badge variant="outline">{getStatusName(t.to_status_id)}</Badge>
                </div>
              ))}
              {transitions?.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  Нет переходов
                </p>
              )}
            </div>
          )}

          <Separator />

          <form onSubmit={handleCreateTransition} className="space-y-3">
            <p className="text-sm font-medium">Добавить переход</p>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <Label htmlFor="from_status">Из статуса</Label>
                <select
                  id="from_status"
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                  value={fromStatusId}
                  onChange={(e) => setFromStatusId(e.target.value)}
                >
                  <option value="">Выберите...</option>
                  {statuses?.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1">
                <Label htmlFor="to_status">В статус</Label>
                <select
                  id="to_status"
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                  value={toStatusId}
                  onChange={(e) => setToStatusId(e.target.value)}
                >
                  <option value="">Выберите...</option>
                  {statuses?.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <Button
              type="submit"
              size="sm"
              disabled={createTransitionMutation.isPending}
            >
              {createTransitionMutation.isPending
                ? "Создание..."
                : "Создать переход"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

/** Пропсы компонента StatusRow. */
interface StatusRowProps {
  /** Данные статуса workflow. */
  status: WorkflowStatusResponse;
}

/** Строка статуса в списке. */
function StatusRow({ status }: StatusRowProps) {
  const categoryLabel =
    CATEGORIES.find((c) => c.value === status.category)?.label ??
    status.category;

  return (
    <div className="flex items-center justify-between rounded-md border p-3">
      <div className="flex items-center gap-3">
        <div
          className="h-4 w-4 rounded-full"
          style={{ backgroundColor: status.color || "#6b7280" }}
        />
        <span className="text-sm font-medium">{status.name}</span>
        <Badge variant="secondary" className="text-xs">
          {categoryLabel}
        </Badge>
        {status.is_default && (
          <Badge variant="outline" className="text-xs">
            По умолчанию
          </Badge>
        )}
      </div>
      <span className="text-xs text-muted-foreground">
        Порядок: {status.sort_order}
      </span>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/settings/workflow",
)({
  component: WorkflowSettingsPage,
});
