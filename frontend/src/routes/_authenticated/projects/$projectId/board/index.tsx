import { useCallback, useEffect, useMemo } from "react";
import {
  createFileRoute,
  Link,
} from "@tanstack/react-router";
import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { monitorForElements } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { toast } from "sonner";
import {
  getBoardData,
  listBoards,
  moveIssue,
  type BoardDataResponse,
} from "@/api/tracker";
import { useBoardStore } from "@/stores/board-store";
import { BoardColumn } from "@/components/tracker/board-column";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

/**
 * Страница Kanban-доски проекта.
 * Отображает колонки по статусам workflow с drag-and-drop перемещением.
 * Поддерживает фильтрацию, выбор доски и оптимистичные обновления.
 */
function BoardPage() {
  const { projectId } = Route.useParams();
  const queryClient = useQueryClient();
  const {
    activeBoardId,
    setActiveBoard,
    filters,
    setFilters,
    resetFilters,
  } = useBoardStore();

  // Загрузка списка досок
  const { data: boards } = useQuery({
    queryKey: ["boards", projectId],
    queryFn: () => listBoards(projectId),
  });

  // Установка доски по умолчанию
  useEffect(() => {
    if (!activeBoardId && boards && boards.length > 0) {
      const defaultBoard = boards.find((b) => b.is_default) ?? boards[0];
      if (defaultBoard) {
        setActiveBoard(defaultBoard.id);
      }
    }
  }, [boards, activeBoardId, setActiveBoard]);

  // Загрузка данных доски
  const { data: boardData, isLoading } = useQuery({
    queryKey: ["board", projectId, activeBoardId],
    queryFn: () => getBoardData(projectId, activeBoardId!),
    enabled: !!activeBoardId,
  });

  // Мутация перемещения задачи с оптимистичным обновлением
  const moveCardMutation = useMutation({
    mutationFn: (data: {
      issueKey: string;
      statusId: string;
      rank: string;
    }) =>
      moveIssue(projectId, data.issueKey, {
        status_id: data.statusId,
        rank: data.rank,
      }),
    onMutate: async (data) => {
      await queryClient.cancelQueries({
        queryKey: ["board", projectId, activeBoardId],
      });
      const previous = queryClient.getQueryData<BoardDataResponse>([
        "board",
        projectId,
        activeBoardId,
      ]);
      queryClient.setQueryData<BoardDataResponse>(
        ["board", projectId, activeBoardId],
        (old) => {
          if (!old) return old;
          return reorderBoard(old, data);
        },
      );
      return { previous };
    },
    onError: (_err, _data, context) => {
      if (context?.previous) {
        queryClient.setQueryData(
          ["board", projectId, activeBoardId],
          context.previous,
        );
      }
      toast.error("Не удалось переместить задачу");
    },
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: ["board", projectId, activeBoardId],
      });
    },
  });

  // Обработчик drop в колонку
  const handleDrop = useCallback(
    (issueKey: string, statusId: string, _index: number) => {
      moveCardMutation.mutate({
        issueKey,
        statusId,
        rank: "0|aaaaaa:",
      });
    },
    [moveCardMutation],
  );

  // Глобальный монитор drag-and-drop
  useEffect(() => {
    return monitorForElements({
      onDrop: ({ source, location }) => {
        const dest = location.current.dropTargets[0];
        if (!dest) return;
        const issueKey = source.data.issueKey as string;
        const statusId = dest.data.statusId as string;
        const currentStatusId = source.data.currentStatusId as string;
        if (currentStatusId === statusId) return;
        moveCardMutation.mutate({
          issueKey,
          statusId,
          rank: "0|aaaaaa:",
        });
      },
    });
  }, [moveCardMutation]);

  // Фильтрация задач в колонках
  const filteredColumns = useMemo(() => {
    if (!boardData?.columns) return [];
    return boardData.columns.map((col) => ({
      ...col,
      issues: col.issues.filter((issue) => {
        if (filters.search) {
          const q = filters.search.toLowerCase();
          if (
            !issue.title.toLowerCase().includes(q) &&
            !issue.issue_key.toLowerCase().includes(q)
          ) {
            return false;
          }
        }
        if (
          filters.priority?.length &&
          !filters.priority.includes(issue.priority)
        ) {
          return false;
        }
        if (filters.type && issue.issue_type !== filters.type) {
          return false;
        }
        if (
          filters.assignee &&
          issue.assignee?.id !== filters.assignee
        ) {
          return false;
        }
        return true;
      }),
    }));
  }, [boardData?.columns, filters]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Загрузка доски...</p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      {/* Панель фильтров */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        {/* Выбор доски */}
        {boards && boards.length > 1 && (
          <select
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
            value={activeBoardId ?? ""}
            onChange={(e) => setActiveBoard(e.target.value)}
          >
            {boards.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>
        )}

        <Input
          placeholder="Поиск задач..."
          className="w-48"
          value={filters.search ?? ""}
          onChange={(e) => setFilters({ search: e.target.value || undefined })}
        />

        <Button
          variant={filters.myTasks ? "default" : "outline"}
          size="sm"
          onClick={() => setFilters({ myTasks: !filters.myTasks })}
        >
          Мои задачи
        </Button>

        {(filters.search ||
          filters.priority?.length ||
          filters.type ||
          filters.assignee ||
          filters.myTasks) && (
          <Button variant="ghost" size="sm" onClick={resetFilters}>
            Сбросить
          </Button>
        )}

        <div className="ml-auto">
          <Button asChild>
            <Link
              to="/projects/$projectId/issues/new"
              params={{ projectId }}
            >
              Новая задача
            </Link>
          </Button>
        </div>
      </div>

      {/* Колонки доски */}
      {filteredColumns.length > 0 ? (
        <div className="flex flex-1 gap-4 overflow-x-auto pb-4">
          {filteredColumns.map((col) => (
            <BoardColumn
              key={col.status.id}
              status={col.status}
              issues={col.issues}
              projectId={projectId}
              onDrop={handleDrop}
            />
          ))}
        </div>
      ) : (
        <div className="flex flex-1 items-center justify-center">
          <div className="text-center">
            <p className="mb-4 text-lg text-muted-foreground">
              Нет задач на этой доске
            </p>
            <Button asChild>
              <Link
                to="/projects/$projectId/issues/new"
                params={{ projectId }}
              >
                Создать задачу
              </Link>
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * Оптимистично перемещает задачу между колонками доски.
 * Используется для мгновенного отображения результата drag-and-drop.
 */
function reorderBoard(
  board: BoardDataResponse,
  data: { issueKey: string; statusId: string },
): BoardDataResponse {
  const columns = board.columns.map((col) => ({
    ...col,
    issues: [...col.issues],
  }));

  let movedIssue = null;
  for (const col of columns) {
    const idx = col.issues.findIndex((i) => i.issue_key === data.issueKey);
    if (idx !== -1) {
      movedIssue = col.issues.splice(idx, 1)[0];
      break;
    }
  }

  if (movedIssue) {
    const targetCol = columns.find((c) => c.status.id === data.statusId);
    if (targetCol) {
      targetCol.issues.push({ ...movedIssue, status: data.statusId });
    }
  }

  return { ...board, columns };
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/board/",
)({
  component: BoardPage,
});
