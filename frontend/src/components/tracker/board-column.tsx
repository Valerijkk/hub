import { useEffect, useRef, useState } from "react";
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

  return (
    <div
      ref={columnRef}
      className={cn(
        "flex min-w-[280px] max-w-[320px] flex-shrink-0 flex-col rounded-lg border bg-muted/30 transition-colors",
        isOver && "border-primary bg-primary/5",
      )}
    >
      {/* Заголовок колонки */}
      <div className="flex items-center justify-between border-b p-3">
        <div className="flex items-center gap-2">
          <div
            className="h-3 w-3 rounded-full"
            style={{ backgroundColor: status.color || "#6b7280" }}
          />
          <span className="text-sm font-semibold">{status.name}</span>
        </div>
        <Badge variant="secondary" className="text-xs">
          {issues.length}
        </Badge>
      </div>

      {/* Список карточек */}
      <div className="flex-1 space-y-2 overflow-y-auto p-2">
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

        {issues.length === 0 && (
          <p className="py-8 text-center text-xs text-muted-foreground">
            Нет задач
          </p>
        )}
      </div>
    </div>
  );
}
