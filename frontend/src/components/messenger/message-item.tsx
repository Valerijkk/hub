import { memo, useCallback, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Check, X } from "lucide-react";
import type { Message } from "@/api/messenger";
import { addBookmark, deleteMessage, editMessage, togglePin } from "@/api/messenger";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
import { useInvalidateMessages } from "@/hooks/use-invalidate";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { ForwardDialog } from "./forward-dialog";
import { LinkPreview } from "./link-preview";
import { MessageActions } from "./message-actions";
import { MessageAttachments } from "./message-attachments";
import { PollMessage } from "./poll-message";
import { RenderContent } from "./render-content";

/**
 * Регулярное выражение для обнаружения URL в тексте сообщения.
 * Поддерживает http/https ссылки.
 */
const URL_REGEX = /https?:\/\/[^\s<>"')\]]+/g;

/** Пропсы компонента MessageItem. */
interface MessageItemProps {
	/** Сообщение для отображения. */
	message: Message;
	/** ID канала. */
	channelId: string;
	/** Предыдущее сообщение (для группировки). */
	previousMessage?: Message;
	/** Показать разделитель даты перед сообщением. */
	showDateSeparator?: boolean;
	/** Текст разделителя даты. */
	dateSeparatorText?: string;
	/** Показать разделитель новых сообщений. */
	showNewMessagesDivider?: boolean;
}

/** Форматирует время из ISO-строки в "ЧЧ:ММ". */
function formatTime(iso: string | undefined | null): string {
	if (!iso || typeof iso !== "string") return "";
	const date = new Date(iso);
	if (Number.isNaN(date.getTime())) return "";
	return date.toLocaleTimeString("ru-RU", {
		hour: "2-digit",
		minute: "2-digit",
	});
}

/**
 * Проверяет, нужно ли группировать сообщение с предыдущим.
 * Группировка: один отправитель, менее 2 минут разницы.
 */
function shouldGroupWithPrevious(
	message: Message,
	previous?: Message,
): boolean {
	if (!previous) return false;
	if (previous.sender_id !== message.sender_id) return false;
	if (!message.created_at || !previous.created_at) return false;
	const curTime = new Date(message.created_at).getTime();
	const prevTime = new Date(previous.created_at).getTime();
	if (Number.isNaN(curTime) || Number.isNaN(prevTime)) return false;
	return curTime - prevTime < 2 * 60 * 1000; // 2 minutes
}

/**
 * MessageItem отображает одно сообщение чата в стиле Slack / Tinkoff Talk.
 * Flat layout без пузырей: аватар слева, имя + время + контент справа.
 * Поддерживает группировку, реакции, hover-действия, ответы и вложения.
 */
function MessageItemInner({
	message,
	channelId,
	previousMessage,
	showDateSeparator,
	dateSeparatorText,
	showNewMessagesDivider,
}: MessageItemProps) {
	const [showActions, setShowActions] = useState(false);
	const [showReactionPicker, setShowReactionPicker] = useState(false);
	const [showForwardDialog, setShowForwardDialog] = useState(false);
	const [isEditing, setIsEditing] = useState(false);
	const [editText, setEditText] = useState("");
	const [editSaving, setEditSaving] = useState(false);
	const editInputRef = useRef<HTMLTextAreaElement>(null);
	const queryClient = useQueryClient();
	const invalidateMessages = useInvalidateMessages(channelId);
	const openThread = useMessengerStore((s) => s.openThread);
	const setReplyTo = useMessengerStore((s) => s.setReplyTo);
	const currentUserId = useAuthStore((s) => s.user?.id);

	// Defensive: skip rendering if message is malformed (e.g. partial WS event)
	if (!message || !message.id || typeof message.id !== "string") {
		return null;
	}

	const isOwn = currentUserId === message.sender_id;
	const isGrouped = shouldGroupWithPrevious(message, previousMessage);

	/**
	 * Извлекает уникальные URL из текста сообщения.
	 */
	const extractedUrls = useMemo(() => {
		const urls = new Set<string>();
		if (message.content_text) {
			const matches = message.content_text.match(URL_REGEX);
			if (matches) matches.forEach((u) => urls.add(u));
		}
		if (message.content_json) {
			const extractFromNode = (node: unknown) => {
				if (!node || typeof node !== "object") return;
				const n = node as Record<string, unknown>;
				if (n.type === "a" && typeof n.url === "string") {
					urls.add(n.url);
				}
				if (typeof n.text === "string") {
					const matches = n.text.match(URL_REGEX);
					if (matches) matches.forEach((u) => urls.add(u));
				}
				if (Array.isArray(n.children)) {
					n.children.forEach(extractFromNode);
				}
			};
			extractFromNode(message.content_json);
		}
		return Array.from(urls);
	}, [message.content_text, message.content_json]);

	const handleDelete = useCallback(async () => {
		await deleteMessage(channelId, message.id);
	}, [channelId, message.id]);

	const handlePin = useCallback(async () => {
		await togglePin(channelId, message.id);
	}, [channelId, message.id]);

	const handleBookmark = useCallback(async () => {
		await addBookmark(message.id, channelId);
	}, [channelId, message.id]);

	/** Ответить на сообщение (установить reply-to). */
	const handleReply = useCallback(() => {
		setReplyTo(message);
	}, [message, setReplyTo]);

	/** Начать редактирование сообщения. */
	const handleStartEdit = useCallback(() => {
		const text = message.content_text ?? "";
		setEditText(text);
		setIsEditing(true);
		setShowActions(false);
		requestAnimationFrame(() => {
			editInputRef.current?.focus();
			editInputRef.current?.setSelectionRange(text.length, text.length);
		});
	}, [message.content_text]);

	/** Сохранить отредактированное сообщение. */
	const handleSaveEdit = useCallback(async () => {
		const trimmed = editText.trim();
		if (!trimmed || trimmed === (message.content_text ?? "").trim()) {
			setIsEditing(false);
			return;
		}
		setEditSaving(true);
		try {
			const contentJson = {
				type: "doc",
				children: [{ type: "p", children: [{ text: trimmed }] }],
			};
			const queryKey = ["messenger", "messages", channelId];
			queryClient.setQueryData(queryKey, (old: unknown) => {
				if (!old) return old;
				const typed = old as { pages: Message[][]; pageParams: unknown[] };
				return {
					...typed,
					pages: typed.pages.map((page) =>
						page.map((msg) =>
							msg.id === message.id
								? {
										...msg,
										content_json: contentJson,
										content_text: trimmed,
										edited_at: new Date().toISOString(),
									}
								: msg,
						),
					),
				};
			});
			setIsEditing(false);
			await editMessage(channelId, message.id, {
				content_json: contentJson,
				content_text: trimmed,
			});
		} catch {
			invalidateMessages();
		} finally {
			setEditSaving(false);
		}
	}, [editText, message.content_text, message.id, channelId, queryClient, invalidateMessages]);

	/** Отменить редактирование. */
	const handleCancelEdit = useCallback(() => {
		setIsEditing(false);
		setEditText("");
	}, []);

	/** Обработка клавиш в поле редактирования. */
	const handleEditKeyDown = useCallback(
		(e: React.KeyboardEvent<HTMLTextAreaElement>) => {
			if (e.key === "Enter" && !e.shiftKey) {
				e.preventDefault();
				void handleSaveEdit();
			} else if (e.key === "Escape") {
				e.preventDefault();
				handleCancelEdit();
			}
		},
		[handleSaveEdit, handleCancelEdit],
	);

	return (
		<>
			{/* Date separator -- centered pill */}
			{showDateSeparator && dateSeparatorText && (
				<div className="flex items-center justify-center py-4 px-5">
					<div className="text-[11.5px] font-semibold text-[var(--t3)] bg-[var(--raised)] px-4 py-1.5 rounded-full border border-[var(--b1)] shadow-[var(--sh2)]">
						{dateSeparatorText}
					</div>
				</div>
			)}

			{/* New messages divider */}
			{showNewMessagesDivider && (
				<div className="flex items-center gap-3.5 py-3 px-5">
					<div className="flex-1 h-px bg-[var(--red)] opacity-50" />
					<span className="text-[11.5px] font-semibold text-[var(--red)] whitespace-nowrap">
						Новые сообщения
					</span>
					<div className="flex-1 h-px bg-[var(--red)] opacity-50" />
				</div>
			)}

			{/* Message -- flat Slack-style layout */}
			<div
				className={cn(
					"msg-row flex gap-3 py-2 px-2.5 rounded-[var(--r3)] transition-[background] relative",
					"hover:bg-[var(--hover)]",
				)}
				onMouseEnter={() => setShowActions(true)}
				onMouseLeave={() => {
					setShowActions(false);
					setShowReactionPicker(false);
				}}
			>
				<div className="flex gap-3 w-full min-w-0">
					{/* Avatar column -- 38px, shown only on first message in group */}
					<div className={cn("w-[38px] shrink-0", isGrouped && "invisible h-0")}>
						{!isGrouped && (
							<Avatar className="w-[38px] h-[38px]">
								<AvatarFallback className="w-[38px] h-[38px] rounded-full flex items-center justify-center text-[13px] font-bold text-white hover:scale-[1.06] hover:shadow-[0_2px_12px_rgba(0,0,0,.3)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] bg-primary/10 text-primary">
									{message.sender_id?.charAt(0)?.toUpperCase() ?? "?"}
								</AvatarFallback>
							</Avatar>
						)}
						{/* Show time on hover for grouped messages */}
						{isGrouped && (
							<span className="hidden group-hover:flex items-center justify-center h-full text-[10px] text-t2/60">
								{formatTime(message.created_at)}
							</span>
						)}
					</div>

					{/* Content */}
					<div className="min-w-0 flex-1">
						{/* Name + time -- shown only on first message in group */}
						{!isGrouped && (
							<div className="flex items-baseline gap-2 mb-0.5">
								<span className="text-[13.5px] font-bold tracking-[-0.01em] hover:underline cursor-pointer">
									{message.sender_id}
								</span>
								<span className="text-[10.5px] text-[var(--t4)] font-medium">
									{formatTime(message.created_at)}
								</span>
								{message.edited_at && (
									<span className="text-[10.5px] text-[var(--t4)] italic">
										(изм.)
									</span>
								)}
							</div>
						)}

						{/* Forwarded badge */}
						{message.forwarded_from_id && (
							<div className="mb-1 flex items-center gap-1 text-xs text-t2/70">
								<svg
									className="h-3 w-3"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={2}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M13 5l7 7-7 7M5 5l7 7-7 7"
									/>
								</svg>
								<span className="italic">Переслано</span>
							</div>
						)}

						{/* Опрос — специальная карточка вместо обычного текста */}
						{message.message_type === "poll" && message.poll_id && (
							<PollMessage
								pollId={message.poll_id}
								channelId={channelId}
							/>
						)}

						{/* Message content (or inline edit) — не показываем для опросов */}
						{message.message_type !== "poll" && (
							isEditing ? (
								<div className="flex flex-col gap-1.5">
									<textarea
										ref={editInputRef}
										value={editText}
										onChange={(e) => setEditText(e.target.value)}
										onKeyDown={handleEditKeyDown}
										disabled={editSaving}
										className="w-full resize-none rounded-lg border border-input bg-base px-3 py-2 text-sm text-t1 shadow-sm focus:outline-none focus:ring-2 focus:ring-ring/50"
										rows={Math.min(
											editText.split("\n").length + 1,
											6,
										)}
									/>
									<div className="flex items-center gap-1 justify-end">
										<button
											type="button"
											className="inline-flex items-center justify-center h-6 w-6 rounded-full hover:bg-accent text-t2 hover:text-t1 transition-colors"
											title="Отменить (Esc)"
											onClick={handleCancelEdit}
											disabled={editSaving}
										>
											<X className="h-3.5 w-3.5" />
										</button>
										<button
											type="button"
											className="inline-flex items-center justify-center h-6 w-6 rounded-full hover:bg-accent text-primary hover:text-primary transition-colors"
											title="Сохранить (Enter)"
											onClick={() => void handleSaveEdit()}
											disabled={editSaving || !editText.trim()}
										>
											<Check className="h-3.5 w-3.5" />
										</button>
									</div>
								</div>
							) : (
								<div className="text-sm leading-relaxed text-[var(--t1)] break-words">
									{message.content_json && (
										<RenderContent contentJson={message.content_json} />
									)}
									{!message.content_json && message.content_text && (
										<p className="whitespace-pre-wrap">
											{message.content_text}
										</p>
									)}
								</div>
							)
						)}

						{/* Voice message */}
						{message.voice_url && (
							<div className="mt-1.5 flex items-center gap-2 rounded-lg bg-raised/50 p-2 max-w-sm">
								<audio controls preload="metadata" className="h-8 max-w-xs">
									<source src={message.voice_url} type="audio/webm" />
								</audio>
								{message.voice_duration_ms && (
									<span className="text-xs text-t2">
										{Math.floor(message.voice_duration_ms / 1000)}с
									</span>
								)}
							</div>
						)}

						{/* Attachments */}
						{message.attachments && message.attachments.length > 0 && (
							<MessageAttachments
								attachments={message.attachments}
								isOwn={isOwn}
							/>
						)}

						{/* Link previews */}
						{extractedUrls.length > 0 && (
							<div className="mt-1.5 space-y-1.5">
								{extractedUrls.slice(0, 2).map((url) => (
									<LinkPreview key={url} url={url} />
								))}
							</div>
						)}

						{/* Reactions -- small emoji chips */}
						{message.reactions && message.reactions.length > 0 && (
							<div className="flex gap-1.5 mt-2 flex-wrap">
								{message.reactions.map((reaction) => {
									const hasReacted = currentUserId
										? reaction.user_ids.includes(currentUserId)
										: false;
									return (
										<button
											key={reaction.emoji}
											type="button"
											className={cn(
												"flex items-center gap-1.5 px-2.5 py-0.5 border border-[var(--b2)] rounded-full text-[13px] cursor-pointer hover:bg-[var(--hover)] hover:border-[var(--b3)] hover:-translate-y-px transition-[color,background-color,border-color,box-shadow,transform,filter,opacity]",
												hasReacted && "border-[rgba(52,211,153,.2)] bg-[rgba(52,211,153,.04)]",
											)}
											onClick={() =>
												void import("@/api/messenger").then((m) =>
													m.toggleReaction(
														channelId,
														message.id,
														reaction.emoji,
													),
												)
											}
										>
											<span>{reaction.emoji}</span>
											<span className="text-[11.5px] font-semibold text-[var(--t2)]">
												{reaction.user_ids.length}
											</span>
										</button>
									);
								})}
							</div>
						)}

						{/* Thread indicator */}
						{message.reply_count > 0 && (
							<button
								type="button"
								className="mt-1.5 inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium text-primary hover:bg-primary/10 transition-colors"
								onClick={() => openThread(message.id)}
							>
								<svg
									className="h-3.5 w-3.5"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={2}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"
									/>
								</svg>
								{message.reply_count}{" "}
								{message.reply_count === 1
									? "ответ"
									: message.reply_count < 5
										? "ответа"
										: "ответов"}
							</button>
						)}
					</div>
				</div>

				{/* Hover actions -- floating pill toolbar on the right */}
				{showActions && !isEditing && (
					<MessageActions
						message={message}
						channelId={channelId}
						isOwn={isOwn}
						showReactionPicker={showReactionPicker}
						onReactionPickerChange={setShowReactionPicker}
						onReply={handleReply}
						onOpenThread={() => openThread(message.id)}
						onPin={handlePin}
						onBookmark={handleBookmark}
						onForward={() => setShowForwardDialog(true)}
						onEdit={handleStartEdit}
						onDelete={handleDelete}
					/>
				)}

				{/* Forward dialog */}
				<ForwardDialog
					message={message}
					channelId={channelId}
					open={showForwardDialog}
					onOpenChange={setShowForwardDialog}
				/>
			</div>
		</>
	);
}

export const MessageItem = memo(MessageItemInner);
