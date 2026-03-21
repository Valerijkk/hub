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

/** Конфигурация статусов результата — цвета и лейблы. */
const RESULT_STATUS_CONFIG: Record<string, { bg: string; text: string; label: string; dotColor: string }> = {
  passed: { bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400", label: "Passed", dotColor: "bg-green-500" },
  failed: { bg: "bg-red-500/10", text: "text-red-700 dark:text-red-400", label: "Failed", dotColor: "bg-red-500" },
  broken: { bg: "bg-orange-500/10", text: "text-orange-700 dark:text-orange-400", label: "Broken", dotColor: "bg-orange-500" },
  skipped: { bg: "bg-gray-200 dark:bg-gray-700/50", text: "text-gray-600 dark:text-gray-400", label: "Skipped", dotColor: "bg-gray-400" },
  unknown: { bg: "bg-gray-100 dark:bg-gray-800/50", text: "text-gray-500", label: "Unknown", dotColor: "bg-gray-300" },
};

/** Конфигурация статусов прогона. */
const RUN_STATUS_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  created: { bg: "bg-gray-500/10", text: "text-gray-600 dark:text-gray-400", label: "Создан" },
  in_progress: { bg: "bg-blue-500/10", text: "text-blue-700 dark:text-blue-400", label: "Выполняется" },
  completed: { bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400", label: "Завершён" },
  aborted: { bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Прерван" },
};

/**
 * Форматирует длительность в человекочитаемый формат.
 */
function formatDuration(ms?: number): string {
  if (!ms) return "--";
  if (ms < 1000) return `${ms}мс`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}с`;
  return `${Math.floor(ms / 60000)}м ${Math.floor((ms % 60000) / 1000)}с`;
}

/**
 * Панель выполнения тест-кейса — профессиональный дизайн.
 * Отображает информацию о кейсе, кнопки статусов, комментарий.
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

  /** Конфигурация кнопок статусов выполнения. */
  const statusButtons: { status: TestResultStatus; label: string; color: string; activeColor: string }[] = [
    { status: "passed", label: "Passed", color: "hover:bg-green-500/10 hover:text-green-700 hover:border-green-500/30", activeColor: "bg-green-500/15 text-green-700 dark:text-green-400 border-green-500/30" },
    { status: "failed", label: "Failed", color: "hover:bg-red-500/10 hover:text-red-700 hover:border-red-500/30", activeColor: "bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/30" },
    { status: "broken", label: "Broken", color: "hover:bg-orange-500/10 hover:text-orange-700 hover:border-orange-500/30", activeColor: "bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/30" },
    { status: "skipped", label: "Skip", color: "hover:bg-gray-500/10 hover:text-gray-700 hover:border-gray-500/30", activeColor: "bg-gray-500/15 text-gray-700 dark:text-gray-400 border-gray-500/30" },
  ];

  return (
    <div className="space-y-4">
      {/* Информация о тест-кейсе */}
      {tc && (
        <div className="rounded-lg border bg-muted/20 p-3">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-mono text-xs font-bold text-primary">{tc.case_key}</span>
            <Badge variant="outline" className="text-[10px]">
              {tc.case_type === "bdd" ? "BDD" : "Classic"}
            </Badge>
          </div>
          <div className="text-sm font-medium leading-tight">{tc.title}</div>
        </div>
      )}

      {/* Кнопки статусов */}
      <div>
        <div className="mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Результат</div>
        <div className="grid grid-cols-4 gap-1.5">
          {statusButtons.map((btn) => (
            <button
              key={btn.status}
              className={cn(
                "rounded-lg border px-3 py-2 text-xs font-semibold transition-all duration-150",
                overallStatus === btn.status
                  ? btn.activeColor
                  : cn("border-border text-muted-foreground", btn.color),
              )}
              onClick={() => setOverallStatus(btn.status)}
            >
              {btn.label}
            </button>
          ))}
        </div>
      </div>

      {/* Комментарий */}
      <div>
        <div className="mb-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">Комментарий</div>
        <textarea
          className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
          rows={3}
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          placeholder="Комментарий к результату..."
        />
      </div>

      {/* Кнопки управления */}
      <div className="flex gap-2">
        <Button
          className="flex-1 gradient-tms text-white"
          disabled={submitMutation.isPending}
          onClick={() => submitMutation.mutate()}
        >
          {submitMutation.isPending ? "Сохранение..." : "Сохранить"}
        </Button>
        <Button variant="outline" onClick={onClose} className="shrink-0">
          Закрыть
        </Button>
      </div>
    </div>
  );
}

/**
 * Страница деталей прогона тестов — дизайн уровня Allure TestOps.
 * Заголовок с прогресс-баром, 4 цветных карточки статистики,
 * дерево результатов слева и панель выполнения справа.
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
      void queryClient.invalidateQueries({ queryKey: ["tms-run-detail", projectId, runId] });
      toast.success("Прогон запущен");
    },
    onError: () => toast.error("Ошибка запуска прогона"),
  });

  const completeMutation = useMutation({
    mutationFn: () => completeRun(projectId, runId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-run-detail", projectId, runId] });
      toast.success("Прогон завершён");
    },
    onError: () => toast.error("Ошибка завершения прогона"),
  });

  const reopenMutation = useMutation({
    mutationFn: () => reopenRun(projectId, runId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-run-detail", projectId, runId] });
      toast.success("Прогон переоткрыт");
    },
    onError: () => toast.error("Ошибка переоткрытия прогона"),
  });

  const addCasesMutation = useMutation({
    mutationFn: () => addCasesToRun(projectId, runId, selectedCaseIds),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-run-detail", projectId, runId] });
      setAddCasesOpen(false);
      setSelectedCaseIds([]);
      toast.success("Тест-кейсы добавлены");
    },
    onError: () => toast.error("Ошибка добавления тест-кейсов"),
  });

  /** Переходит к следующему невыполненному результату. */
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

  /** Переключает выбор тест-кейса для добавления. */
  const toggleCaseSelection = useCallback((caseId: string) => {
    setSelectedCaseIds((prev) =>
      prev.includes(caseId)
        ? prev.filter((id) => id !== caseId)
        : [...prev, caseId],
    );
  }, []);

  if (isLoading || !run) {
    return (
      <div className="mx-auto max-w-6xl animate-fade-in">
        <div className="mb-4 flex items-center gap-2">
          <div className="h-4 w-16 rounded animate-shimmer" />
          <div className="h-4 w-4 rounded animate-shimmer" />
          <div className="h-4 w-40 rounded animate-shimmer" />
        </div>
        <div className="h-8 w-64 rounded animate-shimmer mb-4" />
        <div className="grid grid-cols-4 gap-3 mb-6">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-20 rounded-xl animate-shimmer" />
          ))}
        </div>
        <div className="h-64 w-full rounded-xl animate-shimmer" />
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

  const statusCfg = (RUN_STATUS_CONFIG[run.status] ?? RUN_STATUS_CONFIG.created)!;

  /** Конфигурация карточек статистики. */
  const statCards = [
    { label: "Passed", count: run.passed_count, color: "text-green-600 dark:text-green-400", bgColor: "bg-green-500/10", borderColor: "border-green-500/20" },
    { label: "Failed", count: run.failed_count, color: "text-red-600 dark:text-red-400", bgColor: "bg-red-500/10", borderColor: "border-red-500/20" },
    { label: "Broken", count: run.broken_count, color: "text-orange-600 dark:text-orange-400", bgColor: "bg-orange-500/10", borderColor: "border-orange-500/20" },
    { label: "Skipped", count: run.skipped_count, color: "text-gray-500", bgColor: "bg-gray-500/10", borderColor: "border-gray-500/20" },
  ];

  return (
    <div className="mx-auto max-w-6xl animate-fade-in">
      {/* Навигация (breadcrumb) */}
      <div className="mb-4 flex items-center gap-1.5 text-sm">
        <Link
          to="/projects/$projectId/tms/runs"
          params={{ projectId }}
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          Прогоны
        </Link>
        <svg className="h-3.5 w-3.5 text-muted-foreground/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
        </svg>
        <span className="font-medium line-clamp-1">{run.title}</span>
      </div>

      {/* Заголовок и действия */}
      <div className="mb-6 flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-3 mb-2">
            <h2 className="text-xl font-semibold leading-tight">{run.title}</h2>
            <Badge variant="secondary" className={cn("text-[10px] font-semibold shrink-0", statusCfg.bg, statusCfg.text)}>
              {statusCfg.label}
            </Badge>
          </div>
          {run.description && (
            <p className="text-sm text-muted-foreground mb-2">{run.description}</p>
          )}
          {/* Прогресс-бар в заголовке */}
          {run.total_count > 0 && (
            <div className="flex items-center gap-3">
              <div className="h-2 w-48 overflow-hidden rounded-full bg-muted">
                {run.passed_count > 0 && (
                  <div className="h-full bg-green-500 float-left transition-all duration-500"
                    style={{ width: `${(run.passed_count / run.total_count) * 100}%` }} />
                )}
                {run.failed_count > 0 && (
                  <div className="h-full bg-red-500 float-left transition-all duration-500"
                    style={{ width: `${(run.failed_count / run.total_count) * 100}%` }} />
                )}
                {run.broken_count > 0 && (
                  <div className="h-full bg-orange-500 float-left transition-all duration-500"
                    style={{ width: `${(run.broken_count / run.total_count) * 100}%` }} />
                )}
                {run.skipped_count > 0 && (
                  <div className="h-full bg-gray-400 float-left transition-all duration-500"
                    style={{ width: `${(run.skipped_count / run.total_count) * 100}%` }} />
                )}
              </div>
              <span className="text-xs text-muted-foreground tabular-nums font-medium">{passRate}% passed</span>
            </div>
          )}
          {/* Назначенные */}
          {run.assignees.length > 0 && (
            <div className="mt-2 flex items-center gap-1.5">
              <span className="text-[10px] text-muted-foreground uppercase tracking-wider font-medium">Назначены:</span>
              {run.assignees.map((u) => (
                <Badge key={u.id} variant="outline" className="text-[10px] font-medium">
                  {u.name}
                </Badge>
              ))}
            </div>
          )}
        </div>

        <div className="flex gap-2 shrink-0">
          {run.status === "created" && (
            <>
              <Button
                variant="outline"
                size="sm"
                className="h-8"
                onClick={() => setAddCasesOpen(true)}
              >
                <svg className="mr-1 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                </svg>
                Добавить кейсы
              </Button>
              <Button
                size="sm"
                className="h-8 gradient-tms text-white"
                disabled={startMutation.isPending || run.results.length === 0}
                onClick={() => startMutation.mutate()}
              >
                <svg className="mr-1 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                </svg>
                Запустить
              </Button>
            </>
          )}
          {run.status === "in_progress" && (
            <Button
              size="sm"
              className="h-8"
              disabled={completeMutation.isPending}
              onClick={() => completeMutation.mutate()}
            >
              <svg className="mr-1 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
              Завершить
            </Button>
          )}
          {(run.status === "completed" || run.status === "aborted") && (
            <Button
              variant="outline"
              size="sm"
              className="h-8"
              disabled={reopenMutation.isPending}
              onClick={() => reopenMutation.mutate()}
            >
              Переоткрыть
            </Button>
          )}
        </div>
      </div>

      {/* Карточки статистики — 4 больших числа */}
      <div className="mb-6 grid grid-cols-4 gap-3">
        {statCards.map((card) => (
          <div
            key={card.label}
            className={cn(
              "rounded-xl border p-4 text-center transition-all",
              card.bgColor, card.borderColor,
            )}
          >
            <div className={cn("text-2xl font-bold tabular-nums", card.color)}>
              {card.count}
            </div>
            <div className="text-[11px] font-medium text-muted-foreground mt-0.5">
              {card.label}
            </div>
          </div>
        ))}
      </div>

      {/* Основной контент: таблица + панель выполнения */}
      <div className="flex gap-4">
        {/* Таблица результатов */}
        <div className="flex-1 min-w-0">
          {run.results.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center border rounded-xl border-dashed">
              <svg className="h-12 w-12 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              <p className="text-sm text-muted-foreground mb-3">Нет результатов. Добавьте тест-кейсы в прогон.</p>
              {run.status === "created" && (
                <Button variant="outline" size="sm" onClick={() => setAddCasesOpen(true)}>
                  Добавить тест-кейсы
                </Button>
              )}
            </div>
          ) : (
            <div className="rounded-xl border shadow-card overflow-hidden">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/30">
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Тест-кейс
                    </th>
                    <th className="w-24 px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Статус
                    </th>
                    <th className="w-24 px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Время
                    </th>
                    <th className="w-28 px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Исполнитель
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {run.results.map((r) => {
                    const resultCfg = (RESULT_STATUS_CONFIG[r.status] ?? RESULT_STATUS_CONFIG.unknown)!;
                    const isExecuting = executingResultId === r.id;
                    const isClickable = run.status === "in_progress";

                    return (
                      <tr
                        key={r.id}
                        className={cn(
                          "border-b transition-all duration-150 last:border-0",
                          isClickable && "cursor-pointer hover:bg-muted/30",
                          isExecuting && "bg-primary/5 border-l-2 border-l-primary",
                        )}
                        onClick={() => {
                          if (isClickable) setExecutingResultId(r.id);
                        }}
                      >
                        <td className="px-3 py-2.5">
                          <div className="flex items-center gap-2">
                            <span className={cn("h-2 w-2 rounded-full shrink-0", resultCfg.dotColor)} />
                            <div className="min-w-0">
                              <span className="font-mono text-xs font-bold text-primary">
                                {r.test_case?.case_key ?? "N/A"}
                              </span>
                              <div className="text-xs text-muted-foreground truncate max-w-xs">
                                {r.test_case?.title}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="px-3 py-2.5">
                          <Badge
                            variant="secondary"
                            className={cn("text-[10px] font-semibold", resultCfg.bg, resultCfg.text)}
                          >
                            {resultCfg.label}
                          </Badge>
                        </td>
                        <td className="px-3 py-2.5">
                          <span className="text-xs text-muted-foreground tabular-nums">
                            {formatDuration(r.duration_ms)}
                          </span>
                        </td>
                        <td className="px-3 py-2.5">
                          <span className="text-xs text-muted-foreground">
                            {r.executed_by?.name ?? "--"}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Панель выполнения — фиксированная справа */}
        {executingResult && run.status === "in_progress" && (
          <div className="w-80 shrink-0">
            <Card className="shadow-card sticky top-4">
              <CardHeader className="pb-2 px-4 pt-4">
                <CardTitle className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Выполнение теста
                </CardTitle>
              </CardHeader>
              <CardContent className="px-4 pb-4">
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
          <div className="space-y-1.5">
            {casesData?.items.map((tc) => {
              const existingIds = new Set(
                run.results
                  .map((r) => r.test_case?.id)
                  .filter(Boolean),
              );
              const alreadyAdded = existingIds.has(tc.id);
              const isSelected = selectedCaseIds.includes(tc.id);
              return (
                <label
                  key={tc.id}
                  className={cn(
                    "flex items-center gap-3 rounded-lg border p-2.5 text-sm transition-colors cursor-pointer",
                    alreadyAdded && "opacity-40 cursor-not-allowed",
                    isSelected && !alreadyAdded && "bg-primary/5 border-primary/30",
                    !alreadyAdded && !isSelected && "hover:bg-muted/50",
                  )}
                >
                  <input
                    type="checkbox"
                    checked={isSelected}
                    disabled={alreadyAdded}
                    onChange={() => toggleCaseSelection(tc.id)}
                    className="rounded"
                  />
                  <div className="min-w-0 flex-1">
                    <span className="font-mono text-xs font-bold text-primary">{tc.case_key}</span>
                    <span className="ml-2 text-sm">{tc.title}</span>
                    {alreadyAdded && (
                      <span className="ml-2 text-[10px] text-muted-foreground">(уже добавлен)</span>
                    )}
                  </div>
                </label>
              );
            })}
          </div>
          <Button
            className="mt-4 w-full gradient-tms text-white"
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
