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
const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-500 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-500 text-black",
  low: "bg-blue-500 text-white",
  trivial: "bg-gray-400 text-white",
};

/** Бейдж типа тест-кейса. */
const TYPE_BADGE: Record<string, { label: string; class: string }> = {
  classic: { label: "Classic", class: "bg-blue-100 text-blue-800" },
  bdd: { label: "BDD", class: "bg-green-100 text-green-800" },
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
          "flex w-full items-center gap-1 rounded px-2 py-1 text-left text-sm hover:bg-muted",
          selectedId === suite.id && "bg-muted font-medium",
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => onSelect(selectedId === suite.id ? undefined : suite.id)}
      >
        {hasChildren && (
          <span
            className="mr-1 cursor-pointer text-xs text-muted-foreground"
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
          >
            {expanded ? "\u25BC" : "\u25B6"}
          </span>
        )}
        {!hasChildren && <span className="mr-1 w-3" />}
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
      <div className="w-56 shrink-0 rounded-md border p-2">
        <div className="mb-2 text-xs font-medium uppercase text-muted-foreground">
          Папки
        </div>
        <button
          className={cn(
            "mb-1 flex w-full items-center rounded px-2 py-1 text-left text-sm hover:bg-muted",
            !filters.suiteId && "bg-muted font-medium",
          )}
          onClick={() => handleSuiteSelect(undefined)}
        >
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
          <Input
            placeholder="Поиск тест-кейсов..."
            className="w-48"
            value={filters.search ?? ""}
            onChange={(e) => {
              updateFilter({ search: e.target.value || undefined });
              setOffset(0);
            }}
          />
          <select
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
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
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
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

          <div className="ml-auto">
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Button>Создать кейс</Button>
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
                    className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                    value={newType}
                    onChange={(e) => setNewType(e.target.value as TestCaseType)}
                  >
                    <option value="classic">Classic</option>
                    <option value="bdd">BDD</option>
                  </select>
                  <select
                    className="w-full rounded-md border bg-background px-3 py-2 text-sm"
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
                      className="w-full rounded-md border bg-background px-3 py-2 text-sm"
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

        {/* Таблица тест-кейсов */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <p className="text-muted-foreground">Загрузка...</p>
          </div>
        ) : cases.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <p className="mb-4 text-lg text-muted-foreground">
              Тест-кейсы не найдены
            </p>
          </div>
        ) : (
          <>
            <div className="rounded-md border">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Ключ
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Название
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Тип
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Приоритет
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Статус
                    </th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
                      Обновлён
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {cases.map((tc) => (
                    <tr
                      key={tc.id}
                      className="border-b transition-colors hover:bg-muted/50"
                    >
                      <td className="px-4 py-3">
                        <Link
                          to="/projects/$projectId/tms/cases/$caseKey"
                          params={{ projectId, caseKey: tc.case_key }}
                          className="text-sm font-bold text-primary hover:underline"
                        >
                          {tc.case_key}
                        </Link>
                      </td>
                      <td className="max-w-md truncate px-4 py-3 text-sm">
                        <Link
                          to="/projects/$projectId/tms/cases/$caseKey"
                          params={{ projectId, caseKey: tc.case_key }}
                          className="hover:underline"
                        >
                          {tc.title}
                        </Link>
                      </td>
                      <td className="px-4 py-3">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-xs",
                            TYPE_BADGE[tc.case_type]?.class,
                          )}
                        >
                          {TYPE_BADGE[tc.case_type]?.label ?? tc.case_type}
                        </Badge>
                      </td>
                      <td className="px-4 py-3">
                        <Badge
                          variant="secondary"
                          className={cn(
                            "text-xs",
                            PRIORITY_COLORS[tc.priority],
                          )}
                        >
                          {tc.priority}
                        </Badge>
                      </td>
                      <td className="px-4 py-3">
                        <Badge variant="outline" className="text-xs">
                          {tc.status}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground">
                        {new Date(tc.updated_at).toLocaleDateString("ru-RU")}
                      </td>
                    </tr>
                  ))}
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
