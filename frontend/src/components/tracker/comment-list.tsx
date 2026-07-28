import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { Value } from "@udecode/plate";
import { useState } from "react";
import { toast } from "sonner";
import {
	type CommentResponse,
	createComment,
	deleteComment,
	listComments,
	toggleReaction,
} from "@/api/tracker";
import { PlateEditor } from "@/components/tracker/plate-editor";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import i18n from "@/i18n";
import { toastApiError } from "@/lib/format-api-error";
import { cn } from "@/lib/utils";

/** Plate Value → JSON-объект для API без double-cast. */
function plateValueToBodyJson(
	value: Value | undefined,
): Record<string, unknown> {
	return { type: "doc", children: value ?? [] };
}

/** body_json из API → Plate Value с runtime-guard. */
function bodyJsonToPlateValue(body: unknown): Value {
	if (Array.isArray(body)) return body as Value;
	if (!body || typeof body !== "object") return [];
	const children = (body as Record<string, unknown>).children;
	if (Array.isArray(children)) return children as Value;
	return [];
}

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
	const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

	const { data: comments, isLoading } = useQuery({
		queryKey: ["comments", projectId, issueKey],
		queryFn: () => listComments(projectId, issueKey),
	});

	const createMutation = useMutation({
		mutationFn: () =>
			createComment(projectId, issueKey, {
				body_json: plateValueToBodyJson(newBody),
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
		onError: (err) => toastApiError(err, "Не удалось добавить комментарий"),
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
		onError: (err) => toastApiError(err, "Не удалось удалить комментарий"),
	});

	const reactionMutation = useMutation({
		mutationFn: ({ commentId, emoji }: { commentId: string; emoji: string }) =>
			toggleReaction(projectId, issueKey, commentId, { emoji }),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["comments", projectId, issueKey],
			});
		},
	});

	const handleSubmit = () => {
		if (!newBody || extractPlainText(newBody).trim() === "") {
			toast.error(i18n.t("common:validation.enterComment"));
			return;
		}
		createMutation.mutate();
	};

	if (isLoading) {
		return <p className="text-sm text-t2">Загрузка комментариев...</p>;
	}

	return (
		<div className="space-y-4">
			{/* Диалог подтверждения удаления */}
			<AlertDialog
				open={!!deleteTarget}
				onOpenChange={(open) => {
					if (!open) setDeleteTarget(null);
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Удалить комментарий?</AlertDialogTitle>
						<AlertDialogDescription>
							Это действие нельзя отменить. Комментарий будет удалён навсегда.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Отмена</AlertDialogCancel>
						<AlertDialogAction
							onClick={() => {
								if (deleteTarget) deleteMutation.mutate(deleteTarget);
								setDeleteTarget(null);
							}}
						>
							Удалить
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>

			{/* Список комментариев */}
			{comments?.map((comment) => (
				<CommentItem
					key={comment.id}
					comment={comment}
					onDelete={(id) => setDeleteTarget(id)}
					onReaction={(commentId, emoji) =>
						reactionMutation.mutate({ commentId, emoji })
					}
				/>
			))}

			{comments?.length === 0 && (
				<p className="text-sm text-t2">Нет комментариев</p>
			)}

			{/* Форма нового комментария */}
			<div className="space-y-3 border-t border-b1 pt-4 mt-4">
				<PlateEditor
					value={newBody}
					onChange={setNewBody}
					placeholder="Написать комментарий..."
					className="w-full min-h-[44px] px-4 py-3 bg-input border border-b1 rounded-[var(--r3)] text-[14px] text-t1 resize-none outline-none transition-[border-color] duration-250 focus-within:border-[rgba(91,156,245,0.25)] placeholder:text-t3"
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
				"flex gap-[14px] py-4 border-b border-b1",
				comment.is_internal && "border-yellow-300/20 bg-yellow-950/10",
			)}
		>
			{/* Аватар */}
			<Avatar className="w-[34px] h-[34px] rounded-full shrink-0">
				<AvatarImage src={comment.author.avatar_url} />
				<AvatarFallback className="w-[34px] h-[34px] rounded-full flex items-center justify-center text-[12px] font-bold text-white bg-gradient-to-br from-[var(--blueS)] to-[var(--purpleS)]">
					{comment.author.name.charAt(0).toUpperCase()}
				</AvatarFallback>
			</Avatar>

			<div className="flex-1">
				{/* Заголовок */}
				<div className="mb-[5px] flex items-center justify-between">
					<div className="flex items-center gap-2">
						<span className="text-[13.5px] font-bold">
							{comment.author.name}
						</span>
						<span className="text-[10.5px] text-t4">{timeAgo}</span>
						{comment.is_internal && (
							<span className="text-[10.5px] text-yellow-600">Внутренний</span>
						)}
					</div>
					<Button
						variant="ghost"
						size="sm"
						className="h-6 px-2 text-xs text-t4"
						onClick={() => onDelete(comment.id)}
					>
						Удалить
					</Button>
				</div>

				{/* Тело */}
				<div className="text-[14px] leading-[1.65] text-t2">
					<PlateEditor
						value={bodyJsonToPlateValue(comment.body_json)}
						readOnly
						className="border-none bg-transparent p-0"
					/>
				</div>

				{/* Реакции */}
				{comment.reactions && comment.reactions.length > 0 && (
					<div className="mt-2 flex flex-wrap gap-1">
						{comment.reactions.map((reaction) => (
							<button
								key={reaction.emoji}
								type="button"
								className={cn(
									"inline-flex items-center gap-1 rounded-full border border-b1 px-2 py-0.5 text-xs transition-colors hover:bg-[var(--hover)]",
									reaction.has_reacted && "border-blue bg-blue/10",
								)}
								onClick={() => onReaction(comment.id, reaction.emoji)}
							>
								{reaction.emoji} {reaction.count}
							</button>
						))}
					</div>
				)}
			</div>
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
