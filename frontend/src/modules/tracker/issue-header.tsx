/**
 * Заголовок задачи -- хлебные крошки, тип, ключ, кнопка удаления.
 */
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

/** Конфигурация иконок типов задач. */
export const TYPE_ICONS: Record<string, string> = {
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

export function IssueHeader({
  projectId,
  issueKey,
  issueType,
  onDelete,
  isDeleting,
}: {
  projectId: string;
  issueKey: string;
  issueType: string;
  onDelete: () => void;
  isDeleting: boolean;
}) {
  return (
    <>
      {/* Хлебные крошки */}
      <nav className="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
        <Link
          to="/projects/$projectId"
          params={{ projectId }}
          className="hover:text-foreground"
        >
          Проект
        </Link>
        <span>/</span>
        <Link
          to="/projects/$projectId/issues"
          params={{ projectId }}
          className="hover:text-foreground"
        >
          Задачи
        </Link>
        <span>/</span>
        <span className="text-foreground">{issueKey}</span>
      </nav>

      {/* Заголовок с типом и ключом */}
      <div className="mb-6 flex items-start justify-between">
        <div className="flex items-center gap-3">
          <span className="text-2xl">
            {TYPE_ICONS[issueType] ?? "\u{1F4CB}"}
          </span>
          <span className="text-lg font-bold text-muted-foreground">
            {issueKey}
          </span>
        </div>
        <Button
          variant="destructive"
          size="sm"
          onClick={onDelete}
          disabled={isDeleting}
        >
          Удалить
        </Button>
      </div>
    </>
  );
}
