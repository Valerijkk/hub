import { useEffect, useRef, useState } from "react";
import { Link } from "@tanstack/react-router";
import { draggable } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { IssueBriefResponse } from "@/api/tracker";
import { cn } from "@/lib/utils";

/** Конфигурация цветов приоритетов задач. */
const PRIORITY_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  critical: { bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Критический" },
  high: { bg: "bg-orange-500/10", text: "text-orange-600 dark:text-orange-400", label: "Высокий" },
  medium: { bg: "bg-yellow-500/10", text: "text-yellow-700 dark:text-yellow-400", label: "Средний" },
  low: { bg: "bg-blue-500/10", text: "text-blue-600 dark:text-blue-400", label: "Низкий" },
  trivial: { bg: "bg-gray-500/10", text: "text-gray-500", label: "Тривиальный" },
};

/** Конфигурация иконок типов задач. */
const TYPE_ICONS: Record<string, string> = {
  bug: "\u{1F41B}",
  task: "\u{2705}",
  story: "\u{1F4D6}",
  epic: "\u{26A1}",
  sub_task: "\u{1F4CB}",
  improvement: "\u{1F680}",
  tech_debt: "\u{1F527}",
  spike: "\u{1F50D}",
  release: "\u{1F3F7}\uFE0F",
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
 * Отображает ключ, название, приоритет, тип и исполнителя.
 */
export function IssueCard({ issue, projectId }: IssueCardProps) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);

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

  const priority = PRIORITY_CONFIG[issue.priority] ?? PRIORITY_CONFIG.medium;

  return (
    <div ref={cardRef} className={cn("cursor-grab active:cursor-grabbing", isDragging && "opacity-40")}>
      <Link
        to="/projects/$projectId/issues/$issueKey"
        params={{ projectId, issueKey: issue.issue_key }}
        className="block"
        onClick={(e) => {
          if (isDragging) e.preventDefault();
        }}
      >
        <Card className="gap-2.5 p-3 shadow-card hover-lift bg-card border-border/50">
          {/* Верхняя строка: тип + ключ */}
          <div className="flex items-center gap-2">
            <span className="text-sm" title={issue.issue_type}>
              {TYPE_ICONS[issue.issue_type] ?? "\u{1F4CB}"}
            </span>
            <span className="font-mono text-[11px] font-bold text-muted-foreground tracking-wide">
              {issue.issue_key}
            </span>
          </div>

          {/* Название */}
          <p className="text-sm font-medium leading-snug line-clamp-2">{issue.title}</p>

          {/* Нижняя строка: приоритет + лейблы + аватар */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1.5">
              <Badge
                variant="secondary"
                className={cn("text-[10px] font-semibold px-1.5 py-0 h-5", priority.bg, priority.text)}
              >
                {priority.label}
              </Badge>
              {issue.labels?.map((label) => (
                <Badge key={label} variant="outline" className="text-[10px] px-1.5 py-0 h-5">
                  {label}
                </Badge>
              ))}
            </div>

            {issue.assignee && (
              <Avatar className="h-6 w-6 ring-2 ring-background">
                <AvatarImage src={issue.assignee.avatar_url} />
                <AvatarFallback className="text-[10px] font-medium bg-primary/10 text-primary">
                  {issue.assignee.name.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            )}
          </div>
        </Card>
      </Link>
    </div>
  );
}
