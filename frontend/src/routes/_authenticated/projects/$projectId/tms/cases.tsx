import { useState, useCallback } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listTestCases,
  listSuites,
  createTestCase,
  type TestCaseResponse,
  type ListTestCasesParams,
  type SuiteResponse,
  type TestCaseType,
} from "@/api/tms";
import { useTmsStore } from "@/stores/tms-store";
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
 * Форматирует относительное время на русском языке.
 * Например: "2 часа назад", "вчера", "3 дня назад".
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
  if (diffDays === 1) return "вчера";
  if (diffDays < 7) return `${diffDays} дн. назад`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} нед. назад`;
  return date.toLocaleDateString("ru-RU");
}

/**
 * Рекурсивный компонент дерева папок.
 * Отображает папки с возможностью разворачивания, выбора и подсчёта элементов.
 */
function SuiteTreeItem({
  suite,
  selectedId,
  onSelect,
  depth = 0,
}: {
  suite: SuiteResponse;
  selectedId?: string;
  onSelect: (id: string | undefined) => void;
  depth?: number;
}) {
  const [expanded, setExpanded] = useState(depth < 2);
  const hasChildren = suite.children && suite.children.length > 0;

  return (
    <div>
      <button
        className={cn(
          "group flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-all duration-150",
          "hover:bg-muted/80",
          selectedId === suite.id && "bg-primary/10 text-primary font-medium",
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => onSelect(selectedId === suite.id ? undefined : suite.id)}
      >
        {hasChildren ? (
          <span
            className={cn(
              "mr-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded text-[10px] transition-colors",
              "text-muted-foreground hover:bg-muted-foreground/10",
            )}
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
          >
            <svg
              className={cn("h-3 w-3 transition-transform duration-200", expanded && "rotate-90")}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </span>
        ) : (
          <span className="mr-0.5 w-4 shrink-0" />
        )}
        <svg
          className={cn(
            "h-3.5 w-3.5 shrink-0 transition-colors",
            selectedId === suite.id ? "text-primary" : "text-muted-foreground",
          )}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <span className="truncate flex-1">{suite.name}</span>
      </button>
      {expanded && hasChildren && (
        <div className="animate-fade-in">
          {suite.children?.map((child) => (
            <SuiteTreeItem
              key={child.id}
              suite={child}
              selectedId={selectedId}
              onSelect={onSelect}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * Страница списка тест-кейсов — дизайн уровня Allure TestOps.
 * Содержит дерево папок слева и профессиональную таблицу тест-кейсов справа.
 * Поддерживает фильтрацию, поиск, bulk-выбор и создание новых кейсов.
 */
function TestCasesPage() {
  const { projectId } = Route.useParams();
  const queryClient = useQueryClient();
  const { filters, updateFilter, resetFilters } = useTmsStore();
  const [offset, setOffset] = useState(0);
  const [createOpen, setCreateOpen] = useState(false);
  const [newTitle, setNewTitle] = useState("");
  const [newType, setNewType] = useState<TestCaseType>("classic");
  const [newPriority, setNewPriority] = useState("medium");
  const [newSuiteId, setNewSuiteId] = useState("");
  const [newDescription, setNewDescription] = useState("");
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

  // Создание тест-кейса
  const createMutation = useMutation({
    mutationFn: () =>
      createTestCase(projectId, {
        title: newTitle,
        case_type: newType,
        priority: newPriority as TestCaseResponse["priority"],
        suite_id: newSuiteId || undefined,
        description: newDescription || undefined,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-cases", projectId] });
      setCreateOpen(false);
      setNewTitle("");
      setNewDescription("");
      toast.success("Тест-кейс создан");
    },
    onError: () => toast.error("Ошибка создания тест-кейса"),
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
      {/* Дерево папок — боковая панель */}
      <div className="w-60 shrink-0 flex flex-col rounded-xl border bg-card/50 shadow-card overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2.5 border-b bg-muted/30">
          <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
            Тест-сьюты
          </span>
        </div>
        <div className="flex-1 overflow-y-auto p-1.5">
          <button
            className={cn(
              "mb-1 flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-all duration-150",
              "hover:bg-muted/80",
              !filters.suiteId && "bg-primary/10 text-primary font-medium",
            )}
            onClick={() => handleSuiteSelect(undefined)}
          >
            <svg className="h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            <span className="flex-1">Все кейсы</span>
            <span className="text-[10px] text-muted-foreground tabular-nums">{total}</span>
          </button>
          {suites?.map((suite) => (
            <SuiteTreeItem
              key={suite.id}
              suite={suite}
              selectedId={filters.suiteId}
              onSelect={handleSuiteSelect}
            />
          ))}
        </div>
      </div>

      {/* Основная область */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Верхняя панель: поиск + фильтры + действия */}
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <div className="relative">
            <svg className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/60" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <Input
              placeholder="Поиск тест-кейсов..."
              className="w-56 pl-8 h-8 text-sm"
              value={filters.search ?? ""}
              onChange={(e) => {
                updateFilter({ search: e.target.value || undefined });
                setOffset(0);
              }}
            />
          </div>

          <select
            className="rounded-lg border bg-card px-2.5 py-1.5 text-xs h-8"
            value={filters.status ?? ""}
            onChange={(e) => {
              updateFilter({ status: (e.target.value || undefined) as ListTestCasesParams["status"] });
              setOffset(0);
            }}
          >
            <option value="">Все статусы</option>
            <option value="draft">Черновик</option>
            <option value="active">Активный</option>
            <option value="deprecated">Устаревший</option>
          </select>

          <select
            className="rounded-lg border bg-card px-2.5 py-1.5 text-xs h-8"
            value={filters.priority ?? ""}
            onChange={(e) => {
              updateFilter({ priority: (e.target.value || undefined) as ListTestCasesParams["priority"] });
              setOffset(0);
            }}
          >
            <option value="">Все приоритеты</option>
            <option value="critical">Критический</option>
            <option value="high">Высокий</option>
            <option value="medium">Средний</option>
            <option value="low">Низкий</option>
            <option value="trivial">Тривиальный</option>
          </select>

          {hasActiveFilters && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 text-xs"
              onClick={() => {
                resetFilters();
                setOffset(0);
              }}
            >
              <svg className="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
              Сбросить
            </Button>
          )}

          {/* Bulk-панель при выбранных кейсах */}
          {selectedIds.size > 0 && (
            <div className="flex items-center gap-2 ml-2 px-3 py-1 rounded-lg bg-primary/10 border border-primary/20">
              <span className="text-xs font-medium text-primary">
                Выбрано: {selectedIds.size}
              </span>
              <Button variant="ghost" size="sm" className="h-6 text-xs px-2" onClick={() => setSelectedIds(new Set())}>
                Снять
              </Button>
            </div>
          )}

          <div className="ml-auto flex items-center gap-2">
            <span className="text-xs text-muted-foreground tabular-nums">
              {total} {total === 1 ? "кейс" : total < 5 ? "кейса" : "кейсов"}
            </span>
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Button size="sm" className="gradient-tms text-white shadow-sm h-8">
                  <svg className="mr-1.5 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                  </svg>
                  Создать кейс
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Новый тест-кейс</DialogTitle>
                </DialogHeader>
                <div className="space-y-4">
                  <Input
                    placeholder="Название тест-кейса"
                    value={newTitle}
                    onChange={(e) => setNewTitle(e.target.value)}
                  />
                  <div className="grid grid-cols-2 gap-3">
                    <select
                      className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
                      value={newType}
                      onChange={(e) => setNewType(e.target.value as TestCaseType)}
                    >
                      <option value="classic">Classic</option>
                      <option value="bdd">BDD</option>
                    </select>
                    <select
                      className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
                      value={newPriority}
                      onChange={(e) => setNewPriority(e.target.value)}
                    >
                      <option value="critical">Критический</option>
                      <option value="high">Высокий</option>
                      <option value="medium">Средний</option>
                      <option value="low">Низкий</option>
                      <option value="trivial">Тривиальный</option>
                    </select>
                  </div>
                  {suites && suites.length > 0 && (
                    <select
                      className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
                      value={newSuiteId}
                      onChange={(e) => setNewSuiteId(e.target.value)}
                    >
                      <option value="">Без папки</option>
                      {suites.map((s) => (
                        <option key={s.id} value={s.id}>
                          {s.name}
                        </option>
                      ))}
                    </select>
                  )}
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
            <div className="mx-auto mb-5 flex h-20 w-20 items-center justify-center rounded-2xl bg-muted/50 border border-dashed border-muted-foreground/20">
              <svg className="h-10 w-10 text-muted-foreground/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold mb-1">Нет тест-кейсов</h3>
            <p className="text-sm text-muted-foreground mb-5 max-w-xs text-center">
              {hasActiveFilters
                ? "По вашим фильтрам ничего не найдено. Попробуйте изменить параметры поиска."
                : "Создайте первый тест-кейс, чтобы начать управление тестированием"}
            </p>
            {hasActiveFilters ? (
              <Button variant="outline" onClick={() => { resetFilters(); setOffset(0); }}>
                Сбросить фильтры
              </Button>
            ) : (
              <Button className="gradient-tms text-white" onClick={() => setCreateOpen(true)}>
                Создать первый кейс
              </Button>
            )}
          </div>
        ) : (
          <>
            <div className="rounded-xl border shadow-card overflow-hidden flex-1">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/30">
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
                          "border-b transition-colors hover:bg-muted/30 group",
                          isSelected && "bg-primary/5",
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
                Показано {offset + 1}–{Math.min(offset + limit, total)} из {total}
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
