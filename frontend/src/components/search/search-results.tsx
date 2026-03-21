import {
  Bug,
  FileText,
  User,
  FolderKanban,
  Hash,
  File,
  MessageSquare,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { SearchResult } from "@/api/search";

/** Иконка по типу сущности. */
const typeIcons: Record<string, React.ElementType> = {
  issue: Bug,
  test_case: FileText,
  user: User,
  project: FolderKanban,
  channel: Hash,
  file: File,
  message: MessageSquare,
};

/** Русские названия типов сущностей. */
const typeLabels: Record<string, string> = {
  issue: "Задача",
  test_case: "Тест-кейс",
  user: "Пользователь",
  project: "Проект",
  channel: "Канал",
  file: "Файл",
  message: "Сообщение",
};

/** Русские заголовки групп. */
export const groupLabels: Record<string, string> = {
  issue: "Задачи",
  test_case: "Тест-кейсы",
  user: "Пользователи",
  project: "Проекты",
  channel: "Каналы",
  file: "Файлы",
  message: "Сообщения",
};

interface SearchResultItemProps {
  /** Результат поиска. */
  result: SearchResult;
  /** Выделен ли элемент. */
  selected?: boolean;
}

/** Элемент результата поиска с иконкой, заголовком и подзаголовком. */
export function SearchResultItem({
  result,
  selected,
}: SearchResultItemProps) {
  const Icon = typeIcons[result.type] ?? File;

  return (
    <div
      className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm ${
        selected ? "bg-accent text-accent-foreground" : ""
      }`}
    >
      <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
      <div className="flex-1 overflow-hidden">
        <p className="truncate font-medium">{result.title}</p>
        {result.subtitle && (
          <p className="truncate text-xs text-muted-foreground">
            {result.subtitle}
          </p>
        )}
      </div>
      <Badge variant="secondary" className="shrink-0 text-[10px]">
        {typeLabels[result.type] ?? result.type}
      </Badge>
    </div>
  );
}
