/**
 * Таблица результатов прогона тестов с фильтрацией по статусу.
 */
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

/** Конфигурация статусов результата -- цвета, лейблы и цвет точки-индикатора. */
export const RESULT_STATUS_CONFIG: Record<string, { bg: string; text: string; label: string; dotColor: string }> = {
  passed: { bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400", label: "Passed", dotColor: "bg-green-500" },
  failed: { bg: "bg-red-500/10", text: "text-red-700 dark:text-red-400", label: "Failed", dotColor: "bg-red-500" },
  broken: { bg: "bg-orange-500/10", text: "text-orange-700 dark:text-orange-400", label: "Broken", dotColor: "bg-orange-500" },
  skipped: { bg: "bg-gray-200 dark:bg-gray-700/50", text: "text-gray-600 dark:text-gray-400", label: "Skipped", dotColor: "bg-gray-400" },
  unknown: { bg: "bg-gray-100 dark:bg-gray-800/50", text: "text-gray-500", label: "Unknown", dotColor: "bg-gray-300" },
};

/**
 * Форматирует длительность в человекочитаемый формат.
 */
export function formatDuration(ms?: number): string {
  if (!ms) return "--";
  if (ms < 1000) return `${ms}мс`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}с`;
  return `${Math.floor(ms / 60000)}м ${Math.floor((ms % 60000) / 1000)}с`;
}

/** Строка результата прогона для таблицы. */
interface ResultRow {
  id: string;
  status: string;
  duration_ms?: number;
  test_case?: { case_key: string; title: string } | null;
  executed_by?: { name: string } | null;
}

/** Пропсы таблицы результатов прогона. */
interface RunResultsTableProps {
  results: ResultRow[];
  runStatus: string;
  executingResultId: string | null;
  onSelectResult: (id: string) => void;
  onAddCases: () => void;
}

/** Таблица результатов прогона тестов с фильтрацией по статусу и подсветкой текущего. */
export function RunResultsTable({
  results,
  runStatus,
  executingResultId,
  onSelectResult,
  onAddCases,
}: RunResultsTableProps) {
  const [statusFilter, setStatusFilter] = useState("");

  if (results.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center border rounded-xl border-dashed">
        <svg className="h-12 w-12 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <p className="text-sm text-muted-foreground mb-3">Нет результатов. Добавьте тест-кейсы в прогон.</p>
        {runStatus === "created" && (
          <Button variant="outline" size="sm" onClick={onAddCases}>
            Добавить тест-кейсы
          </Button>
        )}
      </div>
    );
  }

  const filteredResults = statusFilter
    ? results.filter((r) => r.status === statusFilter)
    : results;

  return (
    <>
      {/* Фильтр по статусу результатов */}
      <div className="mb-3 flex items-center gap-2">
        <select
          className="rounded-lg border bg-card px-2.5 py-1.5 text-xs h-8 shadow-sm cursor-pointer hover:border-primary/40 transition-colors focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/40"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
        >
          <option value="">Все результаты</option>
          <option value="passed">Passed</option>
          <option value="failed">Failed</option>
          <option value="broken">Broken</option>
          <option value="skipped">Skipped</option>
          <option value="unknown">Unknown</option>
        </select>
        {statusFilter && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 text-xs"
            onClick={() => setStatusFilter("")}
          >
            <svg className="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
            Сбросить
          </Button>
        )}
        <span className="ml-auto text-[11px] text-muted-foreground tabular-nums">
          {statusFilter
            ? `${filteredResults.length} из ${results.length}`
            : `${results.length} результатов`}
        </span>
      </div>

      <div className="rounded-xl border shadow-card overflow-hidden">
        {filteredResults.length === 0 && statusFilter ? (
          <div className="flex flex-col items-center justify-center py-14 text-center">
            <svg className="h-10 w-10 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
            </svg>
            <p className="text-sm font-medium text-muted-foreground mb-1">Нет результатов с выбранным статусом</p>
            <p className="text-xs text-muted-foreground/60 mb-3">
              Статус "{(RESULT_STATUS_CONFIG[statusFilter] ?? { label: statusFilter }).label}" не найден среди результатов
            </p>
            <Button variant="outline" size="sm" onClick={() => setStatusFilter("")}>
              Показать все результаты
            </Button>
          </div>
        ) : (
          <table className="w-full table-striped">
            <thead>
              <tr className="border-b bg-muted/40">
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
              {filteredResults.map((r) => {
                const resultCfg = (RESULT_STATUS_CONFIG[r.status] ?? RESULT_STATUS_CONFIG.unknown)!;
                const isExecuting = executingResultId === r.id;
                const isClickable = runStatus === "in_progress";

                return (
                  <tr
                    key={r.id}
                    className={cn(
                      "border-b transition-all duration-150 last:border-0",
                      isClickable && "cursor-pointer hover:bg-muted/30",
                      isExecuting && "bg-primary/5 border-l-2 border-l-primary",
                    )}
                    onClick={() => {
                      if (isClickable) onSelectResult(r.id);
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
        )}
      </div>
    </>
  );
}
