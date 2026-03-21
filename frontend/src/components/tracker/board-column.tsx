import { useEffect, useRef, useState } from "react";
import { Link } from "@tanstack/react-router";
import { dropTargetForElements } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { AnimatePresence, motion } from "motion/react";
import { IssueCard } from "./issue-card";
import { Badge } from "@/components/ui/badge";
import type {
  IssueBriefResponse,
  WorkflowStatusResponse,
} from "@/api/tracker";
import { cn } from "@/lib/utils";

/** Пропсы компонента BoardColumn. */
interface BoardColumnProps {
  /** Статус workflow для данной колонки. */
  status: WorkflowStatusResponse;
  /** Задачи в этой колонке. */
  issues: IssueBriefResponse[];
  /** ID проекта. */
  projectId: string;
  /** Колбэк перемещения задачи в эту колонку. */
  onDrop: (issueKey: string, statusId: string, index: number) => void;
}

/**
 * Колонка Kanban-доски.
 * Является зоной сброса (drop target) для перетаскиваемых карточек.
 * Отображает заголовок со статусом и список карточек задач с анимацией.
 */
export function BoardColumn({
  status,
  issues,
  projectId,
  onDrop,
}: BoardColumnProps) {
  const columnRef = useRef<HTMLDivElement>(null);
  const [isOver, setIsOver] = useState(false);

  useEffect(() => {
    const el = columnRef.current;
    if (!el) return;

    return dropTargetForElements({
      element: el,
      getData: () => ({ statusId: status.id }),
      onDragEnter: () => setIsOver(true),
      onDragLeave: () => setIsOver(false),
      onDrop: ({ source }) => {
        setIsOver(false);
        const issueKey = source.data.issueKey as string;
        const currentStatusId = source.data.currentStatusId as string;
        if (currentStatusId === status.id) return;
        onDrop(issueKey, status.id, issues.length);
      },
    });
  }, [status.id, issues.length, onDrop]);

  /** Лёгкий тон фона колонки на основе цвета статуса */
  const columnBgStyle = status.color
    ? { backgroundColor: `${status.color}08` }
    : undefined;

  return (
    <div
      ref={columnRef}
      className={cn(
        "flex min-w-[300px] max-w-[340px] flex-shrink-0 flex-col rounded-xl border transition-all duration-200",
        isOver
          ? "border-primary/50 ring-2 ring-primary/20 shadow-lg"
          : "border-border/50 bg-muted/20",
      )}
      style={!isOver ? columnBgStyle : undefined}
    >
      {/* Заголовок колонки */}
      <div className="flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-2.5">
          <div
            className="h-2.5 w-2.5 rounded-full ring-2 ring-background"
            style={{ backgroundColor: status.color || "#6b7280" }}
          />
          <span className="text-sm font-semibold">{status.name}</span>
          <Badge variant="secondary" className="h-5 min-w-5 justify-center rounded-full px-1.5 text-[10px] font-bold">
            {issues.length}
          </Badge>
        </div>
      </div>

      {/* Список карточек */}
      <div className="flex-1 space-y-2 overflow-y-auto px-2 pb-2">
        <AnimatePresence mode="popLayout">
          {issues.map((issue) => (
            <motion.div
              key={issue.id}
              layout
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95 }}
              transition={{ duration: 0.15 }}
            >
              <IssueCard issue={issue} projectId={projectId} />
            </motion.div>
          ))}
        </AnimatePresence>

        {/* Пустое состояние — dashed border placeholder */}
        {issues.length === 0 && (
          <div className="flex min-h-[120px] items-center justify-center rounded-lg border-2 border-dashed border-border/50 mx-1">
            <p className="text-xs text-muted-foreground/60">
              Перетащите задачу сюда
            </p>
          </div>
        )}
      </div>

      {/* Кнопка добавления задачи внизу колонки */}
      <div className="px-2 pb-2">
        <Link
          to="/projects/$projectId/issues/new"
          params={{ projectId }}
          className="flex w-full items-center justify-center gap-1.5 rounded-lg border border-dashed border-border/50 px-3 py-2 text-xs text-muted-foreground hover:bg-muted/50 hover:border-border hover:text-foreground transition-colors"
        >
          <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          Добавить задачу
        </Link>
      </div>
    </div>
  );
}
