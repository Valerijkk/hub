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
const PRIORITY_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  critical: { bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Критический" },
  high: { bg: "bg-orange-500/10", text: "text-orange-600 dark:text-orange-400", label: "Высокий" },
  medium: { bg: "bg-yellow-500/10", text: "text-yellow-700 dark:text-yellow-400", label: "Средний" },
  low: { bg: "bg-blue-500/10", text: "text-blue-600 dark:text-blue-400", label: "Низкий" },
  trivial: { bg: "bg-gray-500/10", text: "text-gray-500", label: "Тривиальный" },
};

/** Бейдж типа тест-кейса. */
const TYPE_BADGE: Record<string, { label: string; bg: string; text: string }> = {
  classic: { label: "Classic", bg: "bg-blue-500/10", text: "text-blue-700 dark:text-blue-400" },
  bdd: { label: "BDD", bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400" },
};

/** Лейблы статусов тест-кейсов. */
const STATUS_LABELS: Record<string, string> = {
  draft: "Черновик",
  active: "Активный",
  deprecated: "Устаревший",
};

/**
 * Рекурсивный компонент дерева папок.
 * Отображает папки с возможностью разворачивания и выбора для фильтрации.
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
  const [expanded, setExpanded] = useState(true);
  const hasChildren = suite.children && suite.children.length > 0;

  return (
    <div>
      <button
        className={cn(
          "flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-white/5",
          selectedId === suite.id && "bg-primary/10 text-primary font-medium",
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => onSelect(selectedId === suite.id ? undefined : suite.id)}
      >
        {hasChildren && (
          <span
            className="mr-0.5 cursor-pointer text-[10px] text-muted-foreground"
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
          >
            {expanded ? "\u25BC" : "\u25B6"}
          </span>
        )}
        {!hasChildren && <span className="mr-0.5 w-3" />}
        <svg className="h-3.5 w-3.5 shrink-0 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <span className="truncate">{suite.name}</span>
      </button>
      {expanded && hasChildren && (
        <div>
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
 * Страница списка тест-кейсов.
 * Содержит дерево папок слева и таблицу тест-кейсов справа.
 * Поддерживает фильтрацию, поиск и создание новых кейсов.
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

  const params: ListTestCasesParams = {
    offset,
    limit: 50,
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

  return (
    <div className="flex flex-1 gap-4">
      {/* Дерево папок */}
      <div className="w-56 shrink-0 rounded-xl border bg-card/50 p-2">
        <div className="mb-2 px-2 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
          Папки
        </div>
        <button
          className={cn(
            "mb-1 flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-muted",
            !filters.suiteId && "bg-primary/10 text-primary font-medium",
          )}
          onClick={() => handleSuiteSelect(undefined)}
        >
          <svg className="h-3.5 w-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          Все кейсы
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

      {/* Основная область */}
      <div className="flex-1">
        {/* Верхняя панель */}
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <div className="relative">
            <svg className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <Input
              placeholder="Поиск тест-кейсов..."
              className="w-52 pl-9"
              value={filters.search ?? ""}
              onChange={(e) => {
                updateFilter({ search: e.target.value || undefined });
                setOffset(0);
              }}
            />
          </div>
          <select
            className="rounded-lg border bg-card px-3 py-2 text-sm"
            value={filters.status ?? ""}
            onChange={(e) => {
              updateFilter({
                status: (e.target.value || undefined) as ListTestCasesParams["status"],
              });
              setOffset(0);
            }}
          >
            <option value="">Все статусы</option>
            <option value="draft">Черновик</option>
            <option value="active">Активный</option>
            <option value="deprecated">Устаревший</option>
          </select>
          <select
            className="rounded-lg border bg-card px-3 py-2 text-sm"
            value={filters.priority ?? ""}
            onChange={(e) => {
              updateFilter({
                priority: (e.target.value || undefined) as ListTestCasesParams["priority"],
              });
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

          {(filters.search || filters.status || filters.priority || filters.suiteId) && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                resetFilters();
                setOffset(0);
              }}
            >
              Сбросить
            </Button>
          )}

          <div className="ml-auto flex items-center gap-2">
            <span className="text-xs text-muted-foreground">
              {total} кейсов
            </span>
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Button className="gradient-tms text-white shadow-sm">
                  <svg className="mr-1.5 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
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
                    placeholder="Название"
                    value={newTitle}
                    onChange={(e) => setNewTitle(e.target.value)}
                  />
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
          <div className="space-y-2">
            {/* Скелетон таблицы */}
            <div className="rounded-xl border">
              <div className="border-b bg-muted/30 px-4 py-3">
                <div className="h-4 w-full rounded animate-shimmer" />
              </div>
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="border-b px-4 py-4">
                  <div className="h-4 w-3/4 rounded animate-shimmer" />
                </div>
              ))}
            </div>
          </div>
        ) : cases.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 animate-fade-in">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-muted">
              <svg className="h-8 w-8 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold mb-1">Нет тест-кейсов</h3>
            <p className="text-sm text-muted-foreground mb-4">
              Создайте первый тест-кейс, чтобы начать тестирование
            </p>
            <Button onClick={() => setCreateOpen(true)}>
              Создать первый кейс
            </Button>
          </div>
        ) : (
          <>
            <div className="rounded-xl border shadow-card overflow-hidden">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/30">
                    <th className="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Ключ
                    </th>
                    <th className="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Название
                    </th>
                    <th className="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Тип
                    </th>
                    <th className="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Приоритет
                    </th>
                    <th className="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Статус
                    </th>
                    <th className="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                      Обновлён
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {cases.map((tc) => {
                    const priority = PRIORITY_CONFIG[tc.priority] ?? PRIORITY_CONFIG.medium;
                    const typeBadge = TYPE_BADGE[tc.case_type];
                    return (
                      <tr
                        key={tc.id}
                        className="border-b transition-colors hover:bg-muted/30 group"
                      >
                        <td className="px-4 py-3">
                          <Link
                            to="/projects/$projectId/tms/cases/$caseKey"
                            params={{ projectId, caseKey: tc.case_key }}
                            className="font-mono text-xs font-bold text-primary hover:underline"
                          >
                            {tc.case_key}
                          </Link>
                        </td>
                        <td className="max-w-md px-4 py-3 text-sm">
                          <Link
                            to="/projects/$projectId/tms/cases/$caseKey"
                            params={{ projectId, caseKey: tc.case_key }}
                            className="font-medium hover:underline line-clamp-1"
                          >
                            {tc.title}
                          </Link>
                        </td>
                        <td className="px-4 py-3">
                          {typeBadge && (
                            <Badge
                              variant="secondary"
                              className={cn("text-[10px] font-semibold", typeBadge.bg, typeBadge.text)}
                            >
                              {typeBadge.label}
                            </Badge>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          <Badge
                            variant="secondary"
                            className={cn("text-[10px] font-semibold", priority.bg, priority.text)}
                          >
                            {priority.label}
                          </Badge>
                        </td>
                        <td className="px-4 py-3">
                          <Badge variant="outline" className="text-[10px] font-medium">
                            {STATUS_LABELS[tc.status] ?? tc.status}
                          </Badge>
                        </td>
                        <td className="px-4 py-3 text-xs text-muted-foreground">
                          {new Date(tc.updated_at).toLocaleDateString("ru-RU")}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Пагинация */}
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
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/cases",
)({
  component: TestCasesPage,
});
