import { useState, useCallback } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getTestRunDetail,
  startRun,
  completeRun,
  reopenRun,
  addCasesToRun,
  updateTestResult,
  listTestCases,
  type TestResultDetailResponse,
  type StepResult,
  type UpdateTestResultRequest,
  type TestResultStatus,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

/** Цвета статусов результата. */
const RESULT_STATUS_COLORS: Record<string, string> = {
  passed: "bg-green-100 text-green-800",
  failed: "bg-red-100 text-red-800",
  broken: "bg-orange-100 text-orange-800",
  skipped: "bg-gray-200 text-gray-700",
  unknown: "bg-gray-100 text-gray-500",
};

/** Цвета статусов прогона. */
const RUN_STATUS_COLORS: Record<string, string> = {
  created: "bg-gray-200 text-gray-700",
  in_progress: "bg-blue-100 text-blue-800",
  completed: "bg-green-100 text-green-800",
  aborted: "bg-red-100 text-red-800",
};

/** Лейблы статусов прогона. */
const RUN_STATUS_LABELS: Record<string, string> = {
  created: "Создан",
  in_progress: "Выполняется",
  completed: "Завершён",
  aborted: "Прерван",
};

/**
 * Панель выполнения тест-кейса.
 * Отображает шаги с кнопками Pass/Fail/Skip, поле комментария и отправку результата.
 */
function ExecutionPanel({
  result,
  projectId,
  runId,
  onClose,
  onNext,
}: {
  result: TestResultDetailResponse;
  projectId: string;
  runId: string;
  onClose: () => void;
  onNext: () => void;
}) {
  const queryClient = useQueryClient();
  const tc = result.test_case;
  const [comment, setComment] = useState(result.comment ?? "");
  const [overallStatus, setOverallStatus] = useState<TestResultStatus>(
    result.status === "unknown" ? "passed" : result.status,
  );
  const [stepResults] = useState<StepResult[]>(() => {
    if (result.step_results && result.step_results.length > 0) {
      return result.step_results;
    }
    return [];
  });

  const submitMutation = useMutation({
    mutationFn: () => {
      const data: UpdateTestResultRequest = {
        status: overallStatus,
        comment: comment || undefined,
        step_results: stepResults.length > 0 ? stepResults : undefined,
      };
      return updateTestResult(projectId, runId, result.id, data);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["tms-run-detail", projectId, runId],
      });
      toast.success("Результат сохранён");
      onNext();
    },
    onError: () => toast.error("Ошибка сохранения результата"),
  });

  return (
    <div className="space-y-4">
      {/* Информация о тест-кейсе */}
      {tc && (
        <div className="rounded border p-3">
          <div className="mb-1 text-sm font-medium">
            {tc.case_key}: {tc.title}
          </div>
          <div className="flex gap-2">
            <Badge variant="outline" className="text-xs">
              {tc.case_type === "bdd" ? "BDD" : "Classic"}
            </Badge>
            <Badge variant="outline" className="text-xs">
              {tc.priority}
            </Badge>
          </div>
        </div>
      )}

      {/* Общий статус */}
      <div>
        <div className="mb-2 text-sm font-medium">Общий результат</div>
        <div className="flex gap-2">
          {(["passed", "failed", "broken", "skipped"] as TestResultStatus[]).map(
            (s) => (
              <Button
                key={s}
                size="sm"
                variant={overallStatus === s ? "default" : "outline"}
                className={cn(
                  overallStatus === s && RESULT_STATUS_COLORS[s],
                )}
                onClick={() => setOverallStatus(s)}
              >
                {s === "passed"
                  ? "Passed"
                  : s === "failed"
                    ? "Failed"
                    : s === "broken"
                      ? "Broken"
                      : "Skipped"}
              </Button>
            ),
          )}
        </div>
      </div>

      {/* Комментарий */}
      <div>
        <div className="mb-1 text-sm font-medium">Комментарий</div>
        <textarea
          className="w-full rounded-md border bg-background px-3 py-2 text-sm"
          rows={3}
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          placeholder="Комментарий к результату..."
        />
      </div>

      {/* Кнопки управления */}
      <div className="flex gap-2">
        <Button
          className="flex-1"
          disabled={submitMutation.isPending}
          onClick={() => submitMutation.mutate()}
        >
          {submitMutation.isPending ? "Сохранение..." : "Сохранить результат"}
        </Button>
        <Button variant="outline" onClick={onClose}>
          Закрыть
        </Button>
      </div>
    </div>
  );
}

/**
 * Страница деталей прогона тестов.
 * Отображает заголовок, статистику, действия и таблицу результатов.
 * Клик по строке результата открывает панель выполнения.
 */
function TestRunDetailPage() {
  const { projectId, runId } = Route.useParams();
  const queryClient = useQueryClient();
  const [executingResultId, setExecutingResultId] = useState<string | null>(null);
  const [addCasesOpen, setAddCasesOpen] = useState(false);
  const [selectedCaseIds, setSelectedCaseIds] = useState<string[]>([]);

  // Загрузка деталей прогона
  const { data: run, isLoading } = useQuery({
    queryKey: ["tms-run-detail", projectId, runId],
    queryFn: () => getTestRunDetail(projectId, runId),
  });

  // Загрузка тест-кейсов для диалога добавления
  const { data: casesData } = useQuery({
    queryKey: ["tms-cases-for-run", projectId],
    queryFn: () => listTestCases(projectId, { limit: 100 }),
    enabled: addCasesOpen,
  });

  // Мутации управления прогоном
  const startMutation = useMutation({
    mutationFn: () => startRun(projectId, runId),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["tms-run-detail", projectId, runId],
      });
      toast.success("Прогон запущен");
    },
    onError: () => toast.error("Ошибка запуска прогона"),
  });

  const completeMutation = useMutation({
    mutationFn: () => completeRun(projectId, runId),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["tms-run-detail", projectId, runId],
      });
      toast.success("Прогон завершён");
    },
    onError: () => toast.error("Ошибка завершения прогона"),
  });

  const reopenMutation = useMutation({
    mutationFn: () => reopenRun(projectId, runId),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["tms-run-detail", projectId, runId],
      });
      toast.success("Прогон переоткрыт");
    },
    onError: () => toast.error("Ошибка переоткрытия прогона"),
  });

  const addCasesMutation = useMutation({
    mutationFn: () => addCasesToRun(projectId, runId, selectedCaseIds),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["tms-run-detail", projectId, runId],
      });
      setAddCasesOpen(false);
      setSelectedCaseIds([]);
      toast.success("Тест-кейсы добавлены");
    },
    onError: () => toast.error("Ошибка добавления тест-кейсов"),
  });

  /**
   * Переходит к следующему невыполненному результату.
   */
  const handleNext = useCallback(() => {
    if (!run) {
      setExecutingResultId(null);
      return;
    }
    const nextUnknown = run.results.find(
      (r) => r.status === "unknown" && r.id !== executingResultId,
    );
    if (nextUnknown) {
      setExecutingResultId(nextUnknown.id);
    } else {
      setExecutingResultId(null);
    }
  }, [run, executingResultId]);

  /**
   * Переключает выбор тест-кейса для добавления.
   */
  const toggleCaseSelection = useCallback((caseId: string) => {
    setSelectedCaseIds((prev) =>
      prev.includes(caseId)
        ? prev.filter((id) => id !== caseId)
        : [...prev, caseId],
    );
  }, []);

  if (isLoading || !run) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    );
  }

  const executingResult = executingResultId
    ? run.results.find((r) => r.id === executingResultId)
    : null;

  const passRate =
    run.total_count > 0
      ? Math.round((run.passed_count / run.total_count) * 100)
      : 0;

  return (
    <div className="mx-auto max-w-5xl">
      {/* Навигация */}
      <div className="mb-4">
        <Link
          to="/projects/$projectId/tms/runs"
          params={{ projectId }}
          className="text-sm text-muted-foreground hover:underline"
        >
          Прогоны
        </Link>
        <span className="mx-2 text-muted-foreground">/</span>
        <span className="text-sm font-medium">{run.title}</span>
      </div>

      {/* Заголовок и действия */}
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h2 className="text-xl font-semibold">{run.title}</h2>
          {run.description && (
            <p className="mt-1 text-sm text-muted-foreground">{run.description}</p>
          )}
          <div className="mt-2 flex items-center gap-3">
            <Badge
              variant="secondary"
              className={cn("text-xs", RUN_STATUS_COLORS[run.status])}
            >
              {RUN_STATUS_LABELS[run.status] ?? run.status}
            </Badge>
            {run.total_count > 0 && (
              <div className="flex items-center gap-2">
                <div className="h-2 w-32 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-green-500 transition-all"
                    style={{ width: `${passRate}%` }}
                  />
                </div>
                <span className="text-xs text-muted-foreground">{passRate}%</span>
              </div>
            )}
          </div>
        </div>

        <div className="flex gap-2">
          {run.status === "created" && (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAddCasesOpen(true)}
              >
                Добавить кейсы
              </Button>
              <Button
                size="sm"
                disabled={startMutation.isPending}
                onClick={() => startMutation.mutate()}
              >
                Запустить
              </Button>
            </>
          )}
          {run.status === "in_progress" && (
            <Button
              size="sm"
              disabled={completeMutation.isPending}
              onClick={() => completeMutation.mutate()}
            >
              Завершить
            </Button>
          )}
          {(run.status === "completed" || run.status === "aborted") && (
            <Button
              variant="outline"
              size="sm"
              disabled={reopenMutation.isPending}
              onClick={() => reopenMutation.mutate()}
            >
              Переоткрыть
            </Button>
          )}
        </div>
      </div>

      {/* Статистика */}
      <div className="mb-6 grid grid-cols-6 gap-3">
        <Card>
          <CardContent className="p-3 text-center">
            <div className="text-lg font-bold">{run.total_count}</div>
            <div className="text-xs text-muted-foreground">Всего</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3 text-center">
            <div className="text-lg font-bold text-green-600">{run.passed_count}</div>
            <div className="text-xs text-muted-foreground">Passed</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3 text-center">
            <div className="text-lg font-bold text-red-600">{run.failed_count}</div>
            <div className="text-xs text-muted-foreground">Failed</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3 text-center">
            <div className="text-lg font-bold text-orange-600">{run.broken_count}</div>
            <div className="text-xs text-muted-foreground">Broken</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3 text-center">
            <div className="text-lg font-bold text-gray-500">{run.skipped_count}</div>
            <div className="text-xs text-muted-foreground">Skipped</div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-3 text-center">
            <div className="text-lg font-bold text-gray-400">{run.unknown_count}</div>
            <div className="text-xs text-muted-foreground">Unknown</div>
          </CardContent>
        </Card>
      </div>

      {/* Назначенные */}
      {run.assignees.length > 0 && (
        <div className="mb-4">
          <span className="text-xs text-muted-foreground">Назначены: </span>
          {run.assignees.map((u) => (
            <Badge key={u.id} variant="outline" className="mr-1 text-xs">
              {u.name}
            </Badge>
          ))}
        </div>
      )}

      <div className="flex gap-6">
        {/* Таблица результатов */}
        <div className="flex-1">
          {run.results.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              Нет результатов. Добавьте тест-кейсы в прогон.
            </div>
          ) : (
            <div className="rounded-md border">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Тест-кейс
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Статус
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Исполнитель
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {run.results.map((r) => (
                    <tr
                      key={r.id}
                      className={cn(
                        "cursor-pointer border-b transition-colors hover:bg-muted/50",
                        executingResultId === r.id && "bg-blue-50",
                      )}
                      onClick={() => {
                        if (run.status === "in_progress") {
                          setExecutingResultId(r.id);
                        }
                      }}
                    >
                      <td className="px-4 py-3">
                        <div className="text-sm font-medium">
                          {r.test_case?.case_key ?? "N/A"}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {r.test_case?.title}
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-xs",
                            RESULT_STATUS_COLORS[r.status],
                          )}
                        >
                          {r.status}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {r.executed_by?.name ?? "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Панель выполнения */}
        {executingResult && run.status === "in_progress" && (
          <div className="w-96 shrink-0">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Выполнение теста</CardTitle>
              </CardHeader>
              <CardContent>
                <ExecutionPanel
                  result={executingResult}
                  projectId={projectId}
                  runId={runId}
                  onClose={() => setExecutingResultId(null)}
                  onNext={handleNext}
                />
              </CardContent>
            </Card>
          </div>
        )}
      </div>

      {/* Диалог добавления тест-кейсов */}
      <Dialog open={addCasesOpen} onOpenChange={setAddCasesOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Добавить тест-кейсы в прогон</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            {casesData?.items.map((tc) => {
              const existingIds = new Set(
                run.results
                  .map((r) => r.test_case?.id)
                  .filter(Boolean),
              );
              const alreadyAdded = existingIds.has(tc.id);
              return (
                <label
                  key={tc.id}
                  className={cn(
                    "flex items-center gap-3 rounded border p-2 text-sm",
                    alreadyAdded && "opacity-50",
                  )}
                >
                  <input
                    type="checkbox"
                    checked={selectedCaseIds.includes(tc.id)}
                    disabled={alreadyAdded}
                    onChange={() => toggleCaseSelection(tc.id)}
                    className="rounded"
                  />
                  <div>
                    <span className="font-medium">{tc.case_key}</span>
                    <span className="ml-2">{tc.title}</span>
                    {alreadyAdded && (
                      <span className="ml-2 text-xs text-muted-foreground">
                        (уже добавлен)
                      </span>
                    )}
                  </div>
                </label>
              );
            })}
          </div>
          <Button
            className="mt-4 w-full"
            disabled={selectedCaseIds.length === 0 || addCasesMutation.isPending}
            onClick={() => addCasesMutation.mutate()}
          >
            {addCasesMutation.isPending
              ? "Добавление..."
              : `Добавить (${selectedCaseIds.length})`}
          </Button>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/runs_/$runId",
)({
  component: TestRunDetailPage,
});
