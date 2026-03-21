import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Value } from "@udecode/plate";
import {
  createIssue,
  updateIssue,
  type CreateIssueRequest,
  type IssueResponse,
  type IssueType,
  type IssuePriority,
} from "@/api/tracker";
import { PlateEditor } from "@/components/tracker/plate-editor";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

/** Типы задач для выбора. */
const ISSUE_TYPES: { value: IssueType; label: string }[] = [
  { value: "bug", label: "Баг" },
  { value: "task", label: "Задача" },
  { value: "story", label: "История" },
  { value: "epic", label: "Эпик" },
  { value: "sub_task", label: "Подзадача" },
  { value: "improvement", label: "Улучшение" },
  { value: "tech_debt", label: "Техдолг" },
  { value: "spike", label: "Исследование" },
  { value: "release", label: "Релиз" },
];

/** Приоритеты задач для выбора. */
const PRIORITIES: { value: IssuePriority; label: string }[] = [
  { value: "critical", label: "Критический" },
  { value: "high", label: "Высокий" },
  { value: "medium", label: "Средний" },
  { value: "low", label: "Низкий" },
  { value: "trivial", label: "Тривиальный" },
];

/** Пропсы компонента IssueForm. */
interface IssueFormProps {
  /** ID проекта. */
  projectId: string;
  /** Колбэк при успешном создании/обновлении. */
  onSuccess?: (issue: IssueResponse) => void;
  /** Начальные данные для режима редактирования. */
  initialData?: IssueResponse;
  /** Режим формы: создание или редактирование. */
  mode: "create" | "edit";
}

/**
 * Форма создания и редактирования задачи.
 * Поддерживает все 9 типов задач, приоритеты и rich-text описание через PlateEditor.
 * В режиме редактирования использует оптимистичную блокировку через поле version.
 */
export function IssueForm({
  projectId,
  onSuccess,
  initialData,
  mode,
}: IssueFormProps) {
  const queryClient = useQueryClient();

  const [title, setTitle] = useState(initialData?.title ?? "");
  const [issueType, setIssueType] = useState<IssueType>(
    initialData?.issue_type ?? "task",
  );
  const [priority, setPriority] = useState<IssuePriority>(
    initialData?.priority ?? "medium",
  );
  const [description, setDescription] = useState<Value | undefined>(
    initialData?.description_json as Value | undefined,
  );
  const [assigneeId, setAssigneeId] = useState(
    initialData?.assignee?.id ?? "",
  );
  const [estimateMinutes, setEstimateMinutes] = useState<string>(
    initialData?.estimate_minutes?.toString() ?? "",
  );
  const [dueDate, setDueDate] = useState(initialData?.due_date ?? "");

  const createMutation = useMutation({
    mutationFn: (data: CreateIssueRequest) => createIssue(projectId, data),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["board", projectId] });
      void queryClient.invalidateQueries({
        queryKey: ["issues", projectId],
      });
      toast.success("Задача создана");
      onSuccess?.(result);
    },
    onError: () => toast.error("Не удалось создать задачу"),
  });

  const updateMutation = useMutation({
    mutationFn: () => {
      if (!initialData) throw new Error("No initial data for edit");
      return updateIssue(projectId, initialData.issue_key, {
        title,
        description_json: description as Record<string, unknown> | undefined,
        priority,
        assignee_id: assigneeId || undefined,
        estimate_minutes: estimateMinutes
          ? Number(estimateMinutes)
          : undefined,
        due_date: dueDate || undefined,
        version: initialData.version,
      });
    },
    onSuccess: (result) => {
      void queryClient.invalidateQueries({
        queryKey: ["issue", projectId, initialData?.issue_key],
      });
      void queryClient.invalidateQueries({ queryKey: ["board", projectId] });
      toast.success("Задача обновлена");
      onSuccess?.(result);
    },
    onError: () => toast.error("Не удалось обновить задачу"),
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      toast.error("Введите название задачи");
      return;
    }

    if (mode === "create") {
      createMutation.mutate({
        title: title.trim(),
        issue_type: issueType,
        priority,
        description_json: description as Record<string, unknown> | undefined,
        assignee_id: assigneeId || undefined,
        estimate_minutes: estimateMinutes
          ? Number(estimateMinutes)
          : undefined,
        due_date: dueDate || undefined,
      });
    } else {
      updateMutation.mutate();
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Название */}
      <div className="space-y-2">
        <Label htmlFor="title">Название</Label>
        <Input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Введите название задачи"
          required
        />
      </div>

      {/* Тип и приоритет */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="issue_type">Тип</Label>
          <select
            id="issue_type"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            value={issueType}
            onChange={(e) => setIssueType(e.target.value as IssueType)}
          >
            {ISSUE_TYPES.map((t) => (
              <option key={t.value} value={t.value}>
                {t.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="priority">Приоритет</Label>
          <select
            id="priority"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            value={priority}
            onChange={(e) => setPriority(e.target.value as IssuePriority)}
          >
            {PRIORITIES.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Описание */}
      <div className="space-y-2">
        <Label>Описание</Label>
        <PlateEditor
          value={description}
          onChange={setDescription}
          placeholder="Опишите задачу..."
        />
      </div>

      {/* Исполнитель */}
      <div className="space-y-2">
        <Label htmlFor="assignee">Исполнитель (ID)</Label>
        <Input
          id="assignee"
          value={assigneeId}
          onChange={(e) => setAssigneeId(e.target.value)}
          placeholder="ID исполнителя"
        />
      </div>

      {/* Оценка и срок */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="estimate">Оценка (минуты)</Label>
          <Input
            id="estimate"
            type="number"
            min={0}
            value={estimateMinutes}
            onChange={(e) => setEstimateMinutes(e.target.value)}
            placeholder="0"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="due_date">Срок</Label>
          <Input
            id="due_date"
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
          />
        </div>
      </div>

      <Button type="submit" disabled={isPending} className="w-full">
        {isPending
          ? "Сохранение..."
          : mode === "create"
            ? "Создать задачу"
            : "Сохранить изменения"}
      </Button>
    </form>
  );
}
