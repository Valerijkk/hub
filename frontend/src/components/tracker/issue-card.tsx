import { useEffect, useRef, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { draggable } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { IssueBriefResponse } from "@/api/tracker";
import { cn } from "@/lib/utils";

/** Конфигурация цветов и меток приоритетов задач. */
const PRIORITY_CONFIG: Record<string, { dot: string; bg: string; text: string; label: string; borderColor: string }> = {
  critical: { dot: "bg-red-500", bg: "bg-red-500/15", text: "text-red-700 dark:text-red-400", label: "Критический", borderColor: "#ef4444" },
  high: { dot: "bg-orange-500", bg: "bg-orange-500/15", text: "text-orange-700 dark:text-orange-400", label: "Высокий", borderColor: "#f97316" },
  medium: { dot: "bg-yellow-500", bg: "bg-yellow-500/15", text: "text-yellow-700 dark:text-yellow-400", label: "Средний", borderColor: "#eab308" },
  low: { dot: "bg-blue-500", bg: "bg-blue-500/15", text: "text-blue-700 dark:text-blue-400", label: "Низкий", borderColor: "#3b82f6" },
  trivial: { dot: "bg-gray-400", bg: "bg-gray-400/15", text: "text-gray-600 dark:text-gray-400", label: "Тривиальный", borderColor: "#9ca3af" },
};

/** Конфигурация типов задач — SVG-иконки с цветами для лучшей читаемости. */
const TYPE_CONFIG: Record<string, { icon: string; color: string }> = {
  bug: { icon: "\u{1F41B}", color: "text-red-500" },
  task: { icon: "\u{2705}", color: "text-blue-500" },
  story: { icon: "\u{1F4D6}", color: "text-green-500" },
  epic: { icon: "\u{26A1}", color: "text-purple-500" },
  sub_task: { icon: "\u{1F4CB}", color: "text-gray-500" },
  improvement: { icon: "\u{1F680}", color: "text-cyan-500" },
  tech_debt: { icon: "\u{1F527}", color: "text-amber-500" },
  spike: { icon: "\u{1F50D}", color: "text-indigo-500" },
  release: { icon: "\u{1F3F7}\uFE0F", color: "text-pink-500" },
};

/** Пропсы компонента IssueCard. */
interface IssueCardProps {
  /** Краткая информация о задаче. */
  issue: IssueBriefResponse;
  /** ID проекта. */
  projectId: string;
}

/**
 * Карточка задачи для Kanban-доски.
 * Поддерживает drag-and-drop перетаскивание между колонками.
 * Кликабельна — переход на страницу детальной информации о задаче.
 * Отображает тип, ключ, название (до 2 строк), приоритет и аватар исполнителя.
 *
 * Навигация реализована через useNavigate + отслеживание координат указателя,
 * чтобы избежать конфликта между pragmatic-drag-and-drop и обычным кликом.
 * Если указатель сместился более чем на 5 px — считается перетаскиванием.
 */
export function IssueCard({ issue, projectId }: IssueCardProps) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const navigate = useNavigate();

  /** Координаты указателя при нажатии — для отличия клика от перетаскивания. */
  const pointerDownPos = useRef<{ x: number; y: number } | null>(null);

  useEffect(() => {
    const el = cardRef.current;
    if (!el) return;

    return draggable({
      element: el,
      getInitialData: () => ({
        issueKey: issue.issue_key,
        issueId: issue.id,
        currentStatusId: issue.status,
        rank: issue.rank ?? "",
      }),
      onDragStart: () => setIsDragging(true),
      onDrop: () => setIsDragging(false),
    });
  }, [issue.issue_key, issue.id, issue.status, issue.rank]);

  const priority = PRIORITY_CONFIG[issue.priority] ?? PRIORITY_CONFIG.medium!;
  const typeConfig = TYPE_CONFIG[issue.issue_type] ?? { icon: "\u{1F4CB}", color: "text-gray-500" };

  /**
   * Запоминаем начальную позицию указателя при mousedown.
   * Используется для определения: был ли это клик или начало drag.
   */
  const handlePointerDown = (e: React.PointerEvent) => {
    pointerDownPos.current = { x: e.clientX, y: e.clientY };
  };

  /**
   * При отпускании указателя — если смещение <= 5px и не было drag,
   * выполняем навигацию на страницу задачи.
   */
  const handleClick = (e: React.MouseEvent) => {
    if (isDragging) return;

    const start = pointerDownPos.current;
    if (start) {
      const dx = Math.abs(e.clientX - start.x);
      const dy = Math.abs(e.clientY - start.y);
      if (dx > 5 || dy > 5) return;
    }

    void navigate({
      to: "/projects/$projectId/issues/$issueKey",
      params: { projectId, issueKey: issue.issue_key },
    });
  };

  return (
    <div
      ref={cardRef}
      role="button"
      tabIndex={0}
      className={cn(
        "cursor-grab active:cursor-grabbing select-none",
        isDragging && "opacity-30",
      )}
      onPointerDown={handlePointerDown}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          void navigate({
            to: "/projects/$projectId/issues/$issueKey",
            params: { projectId, issueKey: issue.issue_key },
          });
        }
      }}
    >
      <div className={cn(
        "rounded-lg border bg-card p-3 transition-all duration-200",
        "border-border/50 shadow-sm",
        "border-l-[3px]",
        "hover:shadow-md hover:border-primary/30 hover:-translate-y-[2px]",
        "focus-visible:ring-2 focus-visible:ring-primary/30",
      )}
      style={{ borderLeftColor: priority.borderColor }}
      >
        {/* Верхняя строка: тип + ключ */}
        <div className="flex items-center gap-1.5 mb-1.5">
          <span className={cn("text-sm leading-none", typeConfig.color)} title={issue.issue_type}>
            {typeConfig.icon}
          </span>
          <span className="font-mono text-[11px] font-semibold text-muted-foreground/80 tracking-wide">
            {issue.issue_key}
          </span>
        </div>

        {/* Название — до 2 строк */}
        <p className="text-[13px] font-medium leading-snug line-clamp-2 mb-2.5 text-foreground/90">
          {issue.title}
        </p>

        {/* Нижняя строка: приоритет + аватар исполнителя */}
        <div className="flex items-center justify-between">
          <Badge
            variant="secondary"
            className={cn(
              "text-[10px] font-semibold px-1.5 py-0 h-[18px] rounded-md border-0",
              priority.bg,
              priority.text,
            )}
          >
            {priority.label}
          </Badge>

          {issue.assignee ? (
            <Avatar className="h-5 w-5 ring-1 ring-border/50">
              <AvatarImage src={issue.assignee.avatar_url} />
              <AvatarFallback className="text-[9px] font-medium bg-primary/10 text-primary">
                {issue.assignee.name.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
          ) : (
            <div className="h-5 w-5 rounded-full border border-dashed border-border/40" />
          )}
        </div>
      </div>
    </div>
  );
}
