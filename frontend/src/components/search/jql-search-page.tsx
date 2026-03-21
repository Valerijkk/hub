import { useState, useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { searchJQL } from "@/api/search";
import {
  listSavedFilters,
  createSavedFilter,
  type SavedFilter,
} from "@/api/search";
import { JQLEditor } from "./jql-editor";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

// ---------------------------------------------------------------------------
// Конфигурация отображения
// ---------------------------------------------------------------------------

/** Цвета бейджей приоритетов задач. */
const PRIORITY_COLORS: Record<string, string> = {
  low: "bg-blue-500/10 text-blue-700 dark:text-blue-400",
  medium: "bg-yellow-500/10 text-yellow-700 dark:text-yellow-400",
  high: "bg-orange-500/10 text-orange-700 dark:text-orange-400",
  critical: "bg-red-500/10 text-red-600 dark:text-red-400",
};

/** Цвета бейджей типов задач. */
const TYPE_COLORS: Record<string, string> = {
  task: "bg-blue-500/10 text-blue-700 dark:text-blue-400",
  bug: "bg-red-500/10 text-red-600 dark:text-red-400",
  story: "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400",
  epic: "bg-purple-500/10 text-purple-700 dark:text-purple-400",
};

// ---------------------------------------------------------------------------
// Компонент
// ---------------------------------------------------------------------------

interface JQLSearchPageProps {
  /** ID проекта для выполнения JQL-запроса. */
  projectId: string;
}

/**
 * Страница JQL-поиска задач.
 * Содержит JQL-редактор с автодополнением, таблицу результатов
 * и возможность сохранения запросов как фильтров.
 */
export function JQLSearchPage({ projectId }: JQLSearchPageProps) {
  const [query, setQuery] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [saveName, setSaveName] = useState("");
  const [showSaveDialog, setShowSaveDialog] = useState(false);

  /** Выполнение JQL-запроса. */
  const { data, isLoading, error } = useQuery({
    queryKey: ["jql-search", projectId, submittedQuery],
    queryFn: () => searchJQL(projectId, submittedQuery),
    enabled: !!submittedQuery,
  });

  /** Загрузка сохранённых фильтров. */
  const { data: filtersData, refetch: refetchFilters } = useQuery({
    queryKey: ["saved-filters-jql", projectId],
    queryFn: () => listSavedFilters(projectId),
  });

  /** Отправка запроса. */
  const handleSubmit = useCallback(
    (q: string) => {
      const trimmed = q.trim();
      if (trimmed) {
        setSubmittedQuery(trimmed);
      }
    },
    [],
  );

  /** Сохранение текущего запроса как фильтра. */
  const handleSaveFilter = useCallback(async () => {
    if (!saveName.trim() || !submittedQuery) return;
    try {
      await createSavedFilter({
        name: saveName.trim(),
        filter_config: { jql: submittedQuery },
        project_id: projectId,
      });
      setSaveName("");
      setShowSaveDialog(false);
      void refetchFilters();
      toast.success("Фильтр сохранён");
    } catch {
      toast.error("Ошибка сохранения фильтра");
    }
  }, [saveName, submittedQuery, projectId, refetchFilters]);

  /** Применение сохранённого фильтра. */
  const handleApplyFilter = useCallback(
    (filter: SavedFilter) => {
      const jql =
        (filter.filter_config as Record<string, string>)?.jql ?? "";
      if (jql) {
        setQuery(jql);
        setSubmittedQuery(jql);
      }
    },
    [],
  );

  const results = data?.results ?? [];
  const savedFilters = filtersData?.filters ?? [];

  return (
    <div className="space-y-4">
      {/* Заголовок */}
      <div className="flex items-center gap-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <svg
            className="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
            />
          </svg>
        </div>
        <div>
          <h2 className="text-lg font-semibold">JQL-поиск</h2>
          <p className="text-xs text-muted-foreground">
            Поиск задач с помощью JQL-запросов
          </p>
        </div>
      </div>

      {/* Сохранённые фильтры */}
      {savedFilters.length > 0 && (
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs text-muted-foreground">Фильтры:</span>
          {savedFilters
            .filter((f) => f.filter_config && "jql" in f.filter_config)
            .map((filter) => (
              <button
                key={filter.id}
                type="button"
                className="rounded-md border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted/80 hover:text-primary transition-colors"
                onClick={() => handleApplyFilter(filter)}
              >
                {filter.name}
              </button>
            ))}
        </div>
      )}

      {/* Строка поиска */}
      <div className="flex gap-2">
        <JQLEditor
          value={query}
          onChange={setQuery}
          onSubmit={handleSubmit}
          className="flex-1"
        />
        <Button
          onClick={() => handleSubmit(query)}
          disabled={!query.trim() || isLoading}
          className="shrink-0"
        >
          {isLoading ? "Поиск..." : "Найти"}
        </Button>
        {submittedQuery && (
          <Button
            variant="outline"
            size="sm"
            className="shrink-0"
            onClick={() => setShowSaveDialog(!showSaveDialog)}
          >
            <svg
              className="mr-1 h-3.5 w-3.5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
              />
            </svg>
            Сохранить
          </Button>
        )}
      </div>

      {/* Диалог сохранения */}
      {showSaveDialog && (
        <div className="flex items-center gap-2 rounded-lg border bg-card p-3">
          <input
            type="text"
            placeholder="Название фильтра"
            className="flex-1 rounded-md border bg-background px-3 py-1.5 text-sm"
            value={saveName}
            onChange={(e) => setSaveName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") void handleSaveFilter();
            }}
          />
          <Button
            size="sm"
            disabled={!saveName.trim()}
            onClick={() => void handleSaveFilter()}
          >
            Сохранить
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setShowSaveDialog(false)}
          >
            Отмена
          </Button>
        </div>
      )}

      {/* Ошибка */}
      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950/30 p-3 text-sm text-red-700 dark:text-red-400">
          {(error as { message?: string }).message ?? "Ошибка выполнения запроса"}
        </div>
      )}

      {/* Информация о запросе */}
      {submittedQuery && data && (
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <span>
            Запрос: <code className="rounded bg-muted px-1.5 py-0.5 font-mono">{data.query}</code>
          </span>
          <span className="tabular-nums">
            Найдено: {data.total}
          </span>
        </div>
      )}

      {/* Таблица результатов */}
      {submittedQuery && !isLoading && results.length > 0 && (
        <div className="rounded-xl border shadow-card overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b bg-muted/40">
                <th className="px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Ключ
                </th>
                <th className="px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Название
                </th>
                <th className="px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Тип
                </th>
                <th className="px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Приоритет
                </th>
                <th className="px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Статус
                </th>
                <th className="px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  Исполнитель
                </th>
              </tr>
            </thead>
            <tbody>
              {results.map((item) => (
                <tr
                  key={item.key}
                  className="border-b transition-colors hover:bg-primary/[0.03]"
                >
                  <td className="px-4 py-2.5">
                    <span className="font-mono text-xs font-bold text-primary">
                      {item.key}
                    </span>
                  </td>
                  <td className="max-w-md px-4 py-2.5">
                    <span className="text-sm font-medium line-clamp-1">
                      {item.title}
                    </span>
                  </td>
                  <td className="px-4 py-2.5">
                    <Badge
                      variant="secondary"
                      className={cn(
                        "text-[10px] font-semibold px-2 py-0.5",
                        TYPE_COLORS[item.type] ?? "bg-gray-500/10 text-gray-600",
                      )}
                    >
                      {item.type}
                    </Badge>
                  </td>
                  <td className="px-4 py-2.5">
                    <Badge
                      variant="secondary"
                      className={cn(
                        "text-[10px] font-semibold px-2 py-0.5",
                        PRIORITY_COLORS[item.priority] ?? "bg-gray-500/10 text-gray-600",
                      )}
                    >
                      {item.priority}
                    </Badge>
                  </td>
                  <td className="px-4 py-2.5">
                    {item.status ? (
                      <span className="text-xs text-muted-foreground">
                        {item.status}
                      </span>
                    ) : (
                      <span className="text-xs text-muted-foreground/40">--</span>
                    )}
                  </td>
                  <td className="px-4 py-2.5">
                    {item.assignee ? (
                      <span className="text-xs text-muted-foreground">
                        {item.assignee}
                      </span>
                    ) : (
                      <span className="text-xs text-muted-foreground/40">--</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Пустой результат */}
      {submittedQuery && !isLoading && results.length === 0 && !error && (
        <div className="flex flex-col items-center justify-center py-16 animate-fade-in">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-muted/60 to-muted/30 border border-dashed border-muted-foreground/20">
            <svg
              className="h-8 w-8 text-muted-foreground/30"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>
          <h3 className="text-base font-semibold mb-1">
            Ничего не найдено
          </h3>
          <p className="text-sm text-muted-foreground">
            Попробуйте изменить JQL-запрос
          </p>
        </div>
      )}

      {/* Начальное состояние */}
      {!submittedQuery && (
        <div className="rounded-lg border bg-card/50 p-6 text-center">
          <p className="text-sm text-muted-foreground mb-3">
            Введите JQL-запрос для поиска задач
          </p>
          <div className="space-y-1.5 text-xs text-muted-foreground/80 font-mono">
            <p>status = "In Progress" AND priority = high</p>
            <p>assignee = currentUser() ORDER BY created DESC</p>
            <p>type IN (bug, task) AND created {">"}= 2024-01-01</p>
          </div>
        </div>
      )}
    </div>
  );
}
