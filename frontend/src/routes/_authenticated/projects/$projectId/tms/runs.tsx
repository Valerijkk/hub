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

/** Конфигурация статусов прогона. */
const RUN_STATUS_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  created: { bg: "bg-gray-500/10", text: "text-gray-600 dark:text-gray-400", label: "Создан" },
  in_progress: { bg: "bg-blue-500/10", text: "text-blue-700 dark:text-blue-400", label: "Выполняется" },
  completed: { bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400", label: "Завершён" },
  aborted: { bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Прерван" },
};

/** Конфигурация типов прогона. */
const RUN_TYPE_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  manual: { bg: "bg-blue-500/10", text: "text-blue-700 dark:text-blue-400", label: "Ручной" },
  automatic: { bg: "bg-purple-500/10", text: "text-purple-700 dark:text-purple-400", label: "Авто" },
  mixed: { bg: "bg-orange-500/10", text: "text-orange-700 dark:text-orange-400", label: "Смешанный" },
};

/**
 * Страница списка прогонов тестов.
 * Отображает карточки прогонов со статусами, прогресс-барами и статистикой.
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
          className="rounded-lg border bg-card px-3 py-2 text-sm"
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

        <div className="ml-auto flex items-center gap-2">
          <span className="text-xs text-muted-foreground">
            {total} прогонов
          </span>
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild>
              <Button className="gradient-tms text-white shadow-sm">
                <svg className="mr-1.5 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                </svg>
                Создать прогон
              </Button>
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
                  className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
                  value={newType}
                  onChange={(e) => setNewType(e.target.value as TestRunType)}
                >
                  <option value="manual">Ручной</option>
                  <option value="automatic">Автоматический</option>
                  <option value="mixed">Смешанный</option>
                </select>
                <textarea
                  className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
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

      {/* Список прогонов */}
      {isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border p-4 space-y-3">
              <div className="h-5 w-3/4 rounded animate-shimmer" />
              <div className="h-3 w-1/2 rounded animate-shimmer" />
              <div className="h-2 w-full rounded-full animate-shimmer" />
            </div>
          ))}
        </div>
      ) : runs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 animate-fade-in">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-muted">
            <svg className="h-8 w-8 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold mb-1">Нет прогонов</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Создайте первый прогон, чтобы начать тестирование
          </p>
          <Button onClick={() => setCreateOpen(true)}>
            Создать первый прогон
          </Button>
        </div>
      ) : (
        <>
          {/* Карточки прогонов */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {runs.map((run) => {
              const rate = passRate(run);
              const statusCfg = RUN_STATUS_CONFIG[run.status] || { bg: "bg-gray-500/10", text: "text-gray-600 dark:text-gray-400", label: run.status };
              const typeCfg = RUN_TYPE_CONFIG[run.run_type];

              return (
                <Link
                  key={run.id}
                  to="/projects/$projectId/tms/runs/$runId"
                  params={{ projectId, runId: run.id }}
                  className="block"
                >
                  <div className="rounded-xl border bg-card p-4 shadow-card hover-lift transition-all space-y-3">
                    {/* Заголовок */}
                    <div className="flex items-start justify-between gap-2">
                      <h3 className="text-sm font-semibold leading-tight line-clamp-2">{run.title}</h3>
                      <Badge
                        variant="secondary"
                        className={cn("text-[10px] font-semibold shrink-0", statusCfg.bg, statusCfg.text)}
                      >
                        {statusCfg.label}
                      </Badge>
                    </div>

                    {/* Мета: тип + дата */}
                    <div className="flex items-center gap-2">
                      {typeCfg && (
                        <Badge
                          variant="secondary"
                          className={cn("text-[10px] font-medium", typeCfg.bg, typeCfg.text)}
                        >
                          {typeCfg.label}
                        </Badge>
                      )}
                      <span className="text-[11px] text-muted-foreground">
                        {new Date(run.created_at).toLocaleDateString("ru-RU")}
                      </span>
                    </div>

                    {/* Прогресс-бар */}
                    {run.total_count > 0 ? (
                      <div className="space-y-1.5">
                        <div className="flex h-2 overflow-hidden rounded-full bg-muted">
                          {/* Пройденные — зелёный */}
                          {run.passed_count > 0 && (
                            <div
                              className="h-full bg-green-500 transition-all"
                              style={{ width: `${(run.passed_count / run.total_count) * 100}%` }}
                            />
                          )}
                          {/* Проваленные — красный */}
                          {run.failed_count > 0 && (
                            <div
                              className="h-full bg-red-500 transition-all"
                              style={{ width: `${(run.failed_count / run.total_count) * 100}%` }}
                            />
                          )}
                          {/* Сломанные — жёлтый */}
                          {run.broken_count > 0 && (
                            <div
                              className="h-full bg-yellow-500 transition-all"
                              style={{ width: `${(run.broken_count / run.total_count) * 100}%` }}
                            />
                          )}
                          {/* Пропущенные — серый */}
                          {run.skipped_count > 0 && (
                            <div
                              className="h-full bg-gray-400 transition-all"
                              style={{ width: `${(run.skipped_count / run.total_count) * 100}%` }}
                            />
                          )}
                        </div>
                        <div className="flex items-center justify-between text-[10px] text-muted-foreground">
                          <span>{rate}% пройдено</span>
                          <span>{run.passed_count}/{run.total_count}</span>
                        </div>
                      </div>
                    ) : (
                      <p className="text-xs text-muted-foreground/60 italic">
                        Нет кейсов в прогоне
                      </p>
                    )}
                  </div>
                </Link>
              );
            })}
          </div>

          {/* Пагинация */}
          {total > 50 && (
            <div className="mt-6 flex items-center justify-between">
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
