import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Value } from "@udecode/plate";
import {
  listComments,
  createComment,
  deleteComment,
  toggleReaction,
  type CommentResponse,
} from "@/api/tracker";
import { PlateEditor } from "@/components/tracker/plate-editor";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

/** Пропсы компонента CommentList. */
interface CommentListProps {
  /** ID проекта. */
  projectId: string;
  /** Ключ задачи (например, "HUB-42"). */
  issueKey: string;
}

/**
 * Список комментариев к задаче.
 * Отображает комментарии с аватарами, реакциями и действиями.
 * Содержит форму для создания нового комментария с PlateEditor.
 */
export function CommentList({ projectId, issueKey }: CommentListProps) {
  const queryClient = useQueryClient();
  const [newBody, setNewBody] = useState<Value | undefined>();
  const [isInternal, setIsInternal] = useState(false);

  const { data: comments, isLoading } = useQuery({
    queryKey: ["comments", projectId, issueKey],
    queryFn: () => listComments(projectId, issueKey),
  });

  const createMutation = useMutation({
    mutationFn: () =>
      createComment(projectId, issueKey, {
        body_json: (newBody ?? []) as unknown as Record<string, unknown>,
        body_text: extractPlainText(newBody),
        is_internal: isInternal,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["comments", projectId, issueKey],
      });
      setNewBody(undefined);
      setIsInternal(false);
      toast.success("Комментарий добавлен");
    },
    onError: () => toast.error("Не удалось добавить комментарий"),
  });

  const deleteMutation = useMutation({
    mutationFn: (commentId: string) =>
      deleteComment(projectId, issueKey, commentId),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["comments", projectId, issueKey],
      });
      toast.success("Комментарий удалён");
    },
    onError: () => toast.error("Не удалось удалить комментарий"),
  });

  const reactionMutation = useMutation({
    mutationFn: ({
      commentId,
      emoji,
    }: {
      commentId: string;
      emoji: string;
    }) => toggleReaction(projectId, issueKey, commentId, { emoji }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["comments", projectId, issueKey],
      });
    },
  });

  const handleSubmit = () => {
    if (!newBody || extractPlainText(newBody).trim() === "") {
      toast.error("Введите текст комментария");
      return;
    }
    createMutation.mutate();
  };

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">Загрузка комментариев...</p>;
  }

  return (
    <div className="space-y-4">
      {/* Список комментариев */}
      {comments?.map((comment) => (
        <CommentItem
          key={comment.id}
          comment={comment}
          onDelete={(id) => {
            if (window.confirm("Удалить комментарий?")) {
              deleteMutation.mutate(id);
            }
          }}
          onReaction={(commentId, emoji) =>
            reactionMutation.mutate({ commentId, emoji })
          }
        />
      ))}

      {comments?.length === 0 && (
        <p className="text-sm text-muted-foreground">Нет комментариев</p>
      )}

      {/* Форма нового комментария */}
      <div className="space-y-3 border-t pt-4">
        <PlateEditor
          value={newBody}
          onChange={setNewBody}
          placeholder="Написать комментарий..."
          className="min-h-[80px]"
        />
        <div className="flex items-center justify-between">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={isInternal}
              onChange={(e) => setIsInternal(e.target.checked)}
              className="rounded"
            />
            Внутренний комментарий
          </label>
          <Button
            size="sm"
            onClick={handleSubmit}
            disabled={createMutation.isPending}
          >
            {createMutation.isPending ? "Отправка..." : "Отправить"}
          </Button>
        </div>
      </div>
    </div>
  );
}

/** Пропсы элемента комментария. */
interface CommentItemProps {
  /** Данные комментария. */
  comment: CommentResponse;
  /** Колбэк удаления комментария. */
  onDelete: (commentId: string) => void;
  /** Колбэк переключения реакции. */
  onReaction: (commentId: string, emoji: string) => void;
}

/** Элемент комментария с аватаром, телом, реакциями и действиями. */
function CommentItem({ comment, onDelete, onReaction }: CommentItemProps) {
  const timeAgo = getTimeAgo(comment.created_at);

  return (
    <div
      className={cn(
        "rounded-md border p-4",
        comment.is_internal && "border-yellow-300 bg-yellow-50 dark:bg-yellow-950/20",
      )}
    >
      {/* Заголовок */}
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Avatar className="h-6 w-6">
            <AvatarImage src={comment.author.avatar_url} />
            <AvatarFallback className="text-[10px]">
              {comment.author.name.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <span className="text-sm font-medium">{comment.author.name}</span>
          <span className="text-xs text-muted-foreground">{timeAgo}</span>
          {comment.is_internal && (
            <span className="text-xs text-yellow-600">Внутренний</span>
          )}
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="h-6 px-2 text-xs text-muted-foreground"
          onClick={() => onDelete(comment.id)}
        >
          Удалить
        </Button>
      </div>

      {/* Тело */}
      <PlateEditor
        value={comment.body_json as unknown as Value}
        readOnly
        className="border-none bg-transparent p-0"
      />

      {/* Реакции */}
      {comment.reactions.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {comment.reactions.map((reaction) => (
            <button
              key={reaction.emoji}
              type="button"
              className={cn(
                "inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs transition-colors hover:bg-accent",
                reaction.has_reacted && "border-primary bg-primary/10",
              )}
              onClick={() => onReaction(comment.id, reaction.emoji)}
            >
              {reaction.emoji} {reaction.count}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * Извлекает plain text из Plate Value для отправки на бэкенд.
 */
function extractPlainText(value: Value | undefined): string {
  if (!value) return "";
  return value
    .map((node) => {
      if ("children" in node && Array.isArray(node.children)) {
        return (node.children as Array<{ text?: string }>)
          .map((child) => child.text ?? "")
          .join("");
      }
      return "";
    })
    .join("\n");
}

/**
 * Вычисляет относительное время ("2 часа назад").
 */
function getTimeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = Math.floor((now - then) / 1000);

  if (diff < 60) return "только что";
  if (diff < 3600) return `${Math.floor(diff / 60)} мин назад`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} ч назад`;
  if (diff < 2592000) return `${Math.floor(diff / 86400)} дн назад`;
  return new Date(dateStr).toLocaleDateString("ru-RU");
}
