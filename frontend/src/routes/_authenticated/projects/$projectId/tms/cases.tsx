import { useState, useCallback } from "react";
import { formatRelativeTime } from "@/lib/format-time";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import {
  listTestCases,
  listSuites,
  type ListTestCasesParams,
} from "@/api/tms";
import { useTmsStore } from "@/stores/tms-store";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { CaseTreePanel } from "@/modules/tms/components/case-tree";
import { CaseFilters } from "@/modules/tms/components/case-filters";
import { CreateCaseDialog } from "@/modules/tms/components/create-case-dialog";

/** Цвета приоритетов тест-кейсов. */
const PRIORITY_CONFIG: Record<string, { dot: string; bg: string; text: string; label: string }> = {
  critical: { dot: "bg-red-500", bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Критический" },
  high: { dot: "bg-orange-500", bg: "bg-orange-500/10", text: "text-orange-600 dark:text-orange-400", label: "Высокий" },
  medium: { dot: "bg-yellow-500", bg: "bg-yellow-500/10", text: "text-yellow-700 dark:text-yellow-400", label: "Средний" },
  low: { dot: "bg-blue-500", bg: "bg-blue-500/10", text: "text-blue-600 dark:text-blue-400", label: "Низкий" },
  trivial: { dot: "bg-gray-400", bg: "bg-gray-500/10", text: "text-gray-500", label: "Тривиальный" },
};

/** Бейдж типа тест-кейса. */
const TYPE_BADGE: Record<string, { label: string; bg: string; text: string }> = {
  classic: { label: "Classic", bg: "bg-blue-500/10", text: "text-blue-700 dark:text-blue-400" },
  bdd: { label: "BDD", bg: "bg-emerald-500/10", text: "text-emerald-700 dark:text-emerald-400" },
};

/** Конфигурация статусов тест-кейсов. */
const STATUS_CONFIG: Record<string, { label: string; bg: string; text: string }> = {
  draft: { label: "Черновик", bg: "bg-gray-500/10", text: "text-gray-600 dark:text-gray-400" },
  active: { label: "Активный", bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400" },
  deprecated: { label: "Устаревший", bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400" },
};

/**
 * Страница списка тест-кейсов -- дизайн уровня Allure TestOps.
 */
function TestCasesPage() {
  const { projectId } = Route.useParams();
  const { filters, updateFilter, resetFilters } = useTmsStore();
  const [offset, setOffset] = useState(0);
  const [createOpen, setCreateOpen] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const limit = 25;

  const params: ListTestCasesParams = {
    offset,
    limit,
    suite_id: filters.suiteId,
    status: filters.status,
    priority: filters.priority,
    search: filters.search,
  };

  // Загрузка тест-кейсов
  const { data, isLoading } = useQuery({
    queryKey: ["tms-cases", projectId, params],
    queryFn: () => listTestCases(projectId, params),
  });

  // Загрузка дерева папок
  const { data: suites } = useQuery({
    queryKey: ["tms-suites", projectId],
    queryFn: () => listSuites(projectId),
  });

  const handleSuiteSelect = useCallback(
    (id: string | undefined) => {
      updateFilter({ suiteId: id });
      setOffset(0);
    },
    [updateFilter],
  );

  const cases = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / limit);
  const currentPage = Math.floor(offset / limit) + 1;

  /** Переключает выбор одного кейса. */
  const toggleSelect = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  /** Переключает выбор всех кейсов на текущей странице. */
  const toggleSelectAll = useCallback(() => {
    if (selectedIds.size === cases.length && cases.length > 0) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(cases.map((c) => c.id)));
    }
  }, [cases, selectedIds.size]);

  /** Проверяет, активны ли какие-либо фильтры. */
  const hasActiveFilters = !!(filters.search || filters.status || filters.priority || filters.suiteId);

  return (
    <div className="flex flex-1 gap-4 min-h-0">
      {/* Дерево папок -- боковая панель */}
      <CaseTreePanel
        suites={suites}
        selectedSuiteId={filters.suiteId}
        onSelectSuite={handleSuiteSelect}
        total={total}
      />

      {/* Основная область */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Верхняя панель: поиск + фильтры + действия */}
        <CaseFilters
          filters={filters}
          updateFilter={updateFilter}
          resetFilters={resetFilters}
          resetOffset={() => setOffset(0)}
          selectedCount={selectedIds.size}
          onClearSelection={() => setSelectedIds(new Set())}
          total={total}
          onCreateClick={() => setCreateOpen(true)}
          createButton={
            <CreateCaseDialog
              projectId={projectId}
              suites={suites}
              open={createOpen}
              onOpenChange={setCreateOpen}
              trigger={
                <Button size="sm" className="gradient-tms text-white shadow-sm h-8">
                  <svg className="mr-1.5 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                  </svg>
                  Создать кейс
                </Button>
              }
            />
          }
        />

        {/* Таблица тест-кейсов */}
        {isLoading ? (
          <div className="rounded-xl border shadow-card overflow-hidden">
            <div className="border-b bg-muted/30 px-4 py-3">
              <div className="h-4 w-full rounded animate-shimmer" />
            </div>
            {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
              <div key={i} className="border-b px-4 py-3.5 flex items-center gap-4">
                <div className="h-4 w-4 rounded animate-shimmer" />
                <div className="h-3.5 w-16 rounded animate-shimmer" />
                <div className="h-3.5 w-48 rounded animate-shimmer flex-1" />
                <div className="h-5 w-14 rounded-full animate-shimmer" />
                <div className="h-5 w-20 rounded-full animate-shimmer" />
                <div className="h-5 w-16 rounded-full animate-shimmer" />
                <div className="h-3.5 w-20 rounded animate-shimmer" />
              </div>
            ))}
          </div>
        ) : cases.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 animate-fade-in">
            <div className="mx-auto mb-5 flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-muted/60 to-muted/30 border border-dashed border-muted-foreground/20 shadow-sm">
              {filters.suiteId && !filters.search && !filters.status && !filters.priority ? (
                <svg className="h-10 w-10 text-muted-foreground/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
              ) : (
                <svg className="h-10 w-10 text-muted-foreground/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                </svg>
              )}
            </div>
            <h3 className="text-lg font-semibold mb-2">
              {filters.suiteId && !filters.search && !filters.status && !filters.priority
                ? "В этой папке пока нет тест-кейсов"
                : "Нет тест-кейсов"}
            </h3>
            <p className="text-sm text-muted-foreground mb-5 max-w-sm text-center leading-relaxed">
              {filters.suiteId && !filters.search && !filters.status && !filters.priority
                ? "Создайте новый тест-кейс в выбранной папке или перенесите существующие"
                : hasActiveFilters
                  ? "По вашим фильтрам ничего не найдено. Попробуйте изменить параметры поиска."
                  : "Создайте первый тест-кейс, чтобы начать управление тестированием"}
            </p>
            {hasActiveFilters ? (
              <div className="flex gap-2">
                <Button variant="outline" onClick={() => { resetFilters(); setOffset(0); }}>
                  Сбросить фильтры
                </Button>
                {filters.suiteId && (
                  <Button className="gradient-tms text-white" onClick={() => setCreateOpen(true)}>
                    Создать кейс
                  </Button>
                )}
              </div>
            ) : (
              <Button className="gradient-tms text-white" onClick={() => setCreateOpen(true)}>
                Создать первый кейс
              </Button>
            )}
          </div>
        ) : (
          <>
            <div className="rounded-xl border shadow-card overflow-hidden flex-1">
              <table className="w-full table-striped">
                <thead>
                  <tr className="border-b bg-muted/40">
                    <th className="w-10 px-3 py-2.5">
                      <input
                        type="checkbox"
                        className="rounded border-muted-foreground/30"
                        checked={selectedIds.size === cases.length && cases.length > 0}
                        onChange={toggleSelectAll}
                      />
                    </th>
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      ID
                    </th>
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Название
                    </th>
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Тип
                    </th>
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Приоритет
                    </th>
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Статус
                    </th>
                    <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Авто
                    </th>
                    <th className="px-3 py-2.5 text-right text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Обновлено
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {cases.map((tc) => {
                    const priority = (PRIORITY_CONFIG[tc.priority] ?? PRIORITY_CONFIG.medium)!;
                    const typeBadge = TYPE_BADGE[tc.case_type];
                    const statusCfg = (STATUS_CONFIG[tc.status] ?? STATUS_CONFIG.draft)!;
                    const isSelected = selectedIds.has(tc.id);
                    const hasAutotest = !!(tc.full_name || tc.class_name || tc.external_id);

                    return (
                      <tr
                        key={tc.id}
                        className={cn(
                          "border-b transition-colors duration-150 hover:bg-primary/[0.03] group",
                          isSelected && "bg-primary/5 hover:bg-primary/[0.07]",
                        )}
                      >
                        <td className="w-10 px-3 py-2.5">
                          <input
                            type="checkbox"
                            className="rounded border-muted-foreground/30"
                            checked={isSelected}
                            onChange={() => toggleSelect(tc.id)}
                          />
                        </td>
                        <td className="px-3 py-2.5">
                          <Link
                            to="/projects/$projectId/tms/cases/$caseKey"
                            params={{ projectId, caseKey: tc.case_key }}
                            className="font-mono text-xs font-bold text-primary hover:underline whitespace-nowrap"
                          >
                            {tc.case_key}
                          </Link>
                        </td>
                        <td className="max-w-md px-3 py-2.5">
                          <Link
                            to="/projects/$projectId/tms/cases/$caseKey"
                            params={{ projectId, caseKey: tc.case_key }}
                            className="text-sm font-medium hover:underline line-clamp-1 hover:text-primary transition-colors"
                          >
                            {tc.title}
                          </Link>
                        </td>
                        <td className="px-3 py-2.5">
                          {typeBadge && (
                            <Badge
                              variant="secondary"
                              className={cn("text-[10px] font-semibold px-2 py-0.5", typeBadge.bg, typeBadge.text)}
                            >
                              {typeBadge.label}
                            </Badge>
                          )}
                        </td>
                        <td className="px-3 py-2.5">
                          <div className="flex items-center gap-1.5">
                            <span className={cn("h-2 w-2 rounded-full shrink-0", priority.dot)} />
                            <span className={cn("text-xs font-medium", priority.text)}>
                              {priority.label}
                            </span>
                          </div>
                        </td>
                        <td className="px-3 py-2.5">
                          <Badge
                            variant="secondary"
                            className={cn("text-[10px] font-medium px-2 py-0.5", statusCfg.bg, statusCfg.text)}
                          >
                            {statusCfg.label}
                          </Badge>
                        </td>
                        <td className="px-3 py-2.5">
                          {hasAutotest ? (
                            <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400" title="Автоматизирован">
                              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                              </svg>
                            </span>
                          ) : (
                            <span className="flex items-center gap-1 text-muted-foreground/40" title="Ручной">
                              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11" />
                              </svg>
                            </span>
                          )}
                        </td>
                        <td className="px-3 py-2.5 text-right">
                          <span className="text-[11px] text-muted-foreground whitespace-nowrap">
                            {formatRelativeTime(tc.updated_at)}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Пагинация */}
            <div className="mt-3 flex items-center justify-between px-1">
              <span className="text-xs text-muted-foreground tabular-nums">
                Показано {offset + 1}&ndash;{Math.min(offset + limit, total)} из {total}
              </span>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  className="h-7 w-7 p-0"
                  disabled={offset === 0}
                  onClick={() => setOffset(0)}
                >
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
                  </svg>
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="h-7 w-7 p-0"
                  disabled={offset === 0}
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                >
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
                  </svg>
                </Button>
                <span className="px-2 text-xs text-muted-foreground tabular-nums">
                  {currentPage} / {totalPages || 1}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  className="h-7 w-7 p-0"
                  disabled={offset + limit >= total}
                  onClick={() => setOffset(offset + limit)}
                >
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="h-7 w-7 p-0"
                  disabled={offset + limit >= total}
                  onClick={() => setOffset(Math.max(0, (totalPages - 1) * limit))}
                >
                  <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
                  </svg>
                </Button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/cases",
)({
  component: TestCasesPage,
});
