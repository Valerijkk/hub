import { useState } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listTestRuns,
  createTestRun,
  type ListTestRunsParams,
  type TestRunType,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

/** Цвета статусов прогона. */
const RUN_STATUS_COLORS: Record<string, string> = {
  created: "bg-gray-200 text-gray-700",
  in_progress: "bg-blue-100 text-blue-800",
  completed: "bg-green-100 text-green-800",
  aborted: "bg-red-100 text-red-800",
};

/** Лейблы статусов. */
const RUN_STATUS_LABELS: Record<string, string> = {
  created: "Создан",
  in_progress: "Выполняется",
  completed: "Завершён",
  aborted: "Прерван",
};

/** Бейдж типа прогона. */
const RUN_TYPE_LABELS: Record<string, { label: string; class: string }> = {
  manual: { label: "Ручной", class: "bg-blue-100 text-blue-800" },
  automatic: { label: "Авто", class: "bg-purple-100 text-purple-800" },
  mixed: { label: "Смешанный", class: "bg-orange-100 text-orange-800" },
};

/**
 * Страница списка прогонов тестов.
 * Отображает таблицу прогонов со статусами и статистикой.
 */
function TestRunsPage() {
  const { projectId } = Route.useParams();
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState("");
  const [offset, setOffset] = useState(0);
  const [createOpen, setCreateOpen] = useState(false);
  const [newTitle, setNewTitle] = useState("");
  const [newType, setNewType] = useState<TestRunType>("manual");
  const [newDescription, setNewDescription] = useState("");

  const params: ListTestRunsParams = {
    offset,
    limit: 50,
    status: (statusFilter || undefined) as ListTestRunsParams["status"],
  };

  // Загрузка прогонов
  const { data, isLoading } = useQuery({
    queryKey: ["tms-runs", projectId, params],
    queryFn: () => listTestRuns(projectId, params),
  });

  // Создание прогона
  const createMutation = useMutation({
    mutationFn: () =>
      createTestRun(projectId, {
        title: newTitle,
        run_type: newType,
        description: newDescription || undefined,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-runs", projectId] });
      setCreateOpen(false);
      setNewTitle("");
      setNewDescription("");
      toast.success("Прогон создан");
    },
    onError: () => toast.error("Ошибка создания прогона"),
  });

  const runs = data?.items ?? [];
  const total = data?.total ?? 0;

  /**
   * Рассчитывает процент пройденных тестов.
   */
  function passRate(r: (typeof runs)[number]): number {
    if (r.total_count === 0) return 0;
    return Math.round((r.passed_count / r.total_count) * 100);
  }

  return (
    <div>
      {/* Верхняя панель */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <select
          className="rounded-md border bg-background px-3 py-1.5 text-sm"
          value={statusFilter}
          onChange={(e) => {
            setStatusFilter(e.target.value);
            setOffset(0);
          }}
        >
          <option value="">Все статусы</option>
          <option value="created">Создан</option>
          <option value="in_progress">Выполняется</option>
          <option value="completed">Завершён</option>
          <option value="aborted">Прерван</option>
        </select>

        {statusFilter && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setStatusFilter("");
              setOffset(0);
            }}
          >
            Сбросить
          </Button>
        )}

        <div className="ml-auto">
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild>
              <Button>Создать прогон</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Новый прогон тестов</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <Input
                  placeholder="Название прогона"
                  value={newTitle}
                  onChange={(e) => setNewTitle(e.target.value)}
                />
                <select
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                  value={newType}
                  onChange={(e) => setNewType(e.target.value as TestRunType)}
                >
                  <option value="manual">Ручной</option>
                  <option value="automatic">Автоматический</option>
                  <option value="mixed">Смешанный</option>
                </select>
                <textarea
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                  placeholder="Описание (необязательно)"
                  rows={3}
                  value={newDescription}
                  onChange={(e) => setNewDescription(e.target.value)}
                />
                <Button
                  className="w-full"
                  disabled={!newTitle.trim() || createMutation.isPending}
                  onClick={() => createMutation.mutate()}
                >
                  {createMutation.isPending ? "Создание..." : "Создать"}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Таблица прогонов */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <p className="text-muted-foreground">Загрузка...</p>
        </div>
      ) : runs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12">
          <p className="mb-4 text-lg text-muted-foreground">
            Прогоны не найдены
          </p>
        </div>
      ) : (
        <>
          <div className="rounded-md border">
            <table className="w-full">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                    Название
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                    Тип
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                    Статус
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                    Прогресс
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                    Создан
                  </th>
                </tr>
              </thead>
              <tbody>
                {runs.map((run) => {
                  const rate = passRate(run);
                  const typeBadge = RUN_TYPE_LABELS[run.run_type];
                  return (
                    <tr
                      key={run.id}
                      className="border-b transition-colors hover:bg-muted/50"
                    >
                      <td className="px-4 py-3">
                        <Link
                          to="/projects/$projectId/tms/runs/$runId"
                          params={{ projectId, runId: run.id }}
                          className="text-sm font-medium text-primary hover:underline"
                        >
                          {run.title}
                        </Link>
                      </td>
                      <td className="px-4 py-3">
                        {typeBadge && (
                          <Badge
                            variant="secondary"
                            className={cn("text-xs", typeBadge.class)}
                          >
                            {typeBadge.label}
                          </Badge>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-xs",
                            RUN_STATUS_COLORS[run.status],
                          )}
                        >
                          {RUN_STATUS_LABELS[run.status] ?? run.status}
                        </Badge>
                      </td>
                      <td className="px-4 py-3">
                        {run.total_count > 0 ? (
                          <div className="flex items-center gap-2">
                            <div className="h-2 w-24 overflow-hidden rounded-full bg-muted">
                              <div
                                className="h-full rounded-full bg-green-500 transition-all"
                                style={{ width: `${rate}%` }}
                              />
                            </div>
                            <span className="text-xs text-muted-foreground">
                              {rate}% ({run.passed_count}/{run.total_count})
                            </span>
                          </div>
                        ) : (
                          <span className="text-xs text-muted-foreground">
                            Нет кейсов
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {new Date(run.created_at).toLocaleDateString("ru-RU")}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {total > 50 && (
            <div className="mt-4 flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                {offset + 1}-{Math.min(offset + 50, total)} из {total}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - 50))}
                >
                  Назад
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={offset + 50 >= total}
                  onClick={() => setOffset(offset + 50)}
                >
                  Далее
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/runs",
)({
  component: TestRunsPage,
});
