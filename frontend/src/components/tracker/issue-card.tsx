import { useEffect, useRef, useState } from "react";
import { Link } from "@tanstack/react-router";
import { draggable } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { IssueBriefResponse } from "@/api/tracker";
import { cn } from "@/lib/utils";

/** Конфигурация цветов приоритетов задач. */
const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-500 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-500 text-black",
  low: "bg-blue-500 text-white",
  trivial: "bg-gray-400 text-white",
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

  return (
    <div ref={cardRef} className={cn("cursor-grab", isDragging && "opacity-50")}>
      <Link
        to="/projects/$projectId/issues/$issueKey"
        params={{ projectId, issueKey: issue.issue_key }}
        className="block"
        onClick={(e) => {
          if (isDragging) e.preventDefault();
        }}
      >
        <Card className="gap-2 p-3 transition-shadow hover:shadow-md">
          <div className="flex items-center gap-2">
            <span className="text-sm" title={issue.issue_type}>
              {TYPE_ICONS[issue.issue_type] ?? "\u{1F4CB}"}
            </span>
            <span className="text-xs font-bold text-muted-foreground">
              {issue.issue_key}
            </span>
          </div>

          <p className="text-sm font-medium leading-tight">{issue.title}</p>

          <div className="flex items-center justify-between">
            <Badge
              variant="secondary"
              className={cn("text-xs", PRIORITY_COLORS[issue.priority])}
            >
              {issue.priority}
            </Badge>

            {issue.assignee && (
              <Avatar className="h-5 w-5">
                <AvatarImage src={issue.assignee.avatar_url} />
                <AvatarFallback className="text-[10px]">
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
