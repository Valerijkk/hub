import { useQueryClient } from "@tanstack/react-query";
import { Check, X } from "lucide-react";
import { memo, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { ChannelCapabilities, ChannelCapability, Message } from "@/api/messenger";
import {
	addBookmark,
	deleteMessage,
	editMessage,
	toggleReaction,
	togglePin,
} from "@/api/messenger";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { messagesQueryKey } from "@/hooks/use-messages";
import { toastApiError } from "@/lib/format-api-error";
import { isSafeUrl } from "@/lib/safe-url";
import { cn } from "@/lib/utils";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { BotMarkdown, isHubAIMarkdown } from "./bot-markdown";
import { ForwardDialog } from "./forward-dialog";
import { LinkPreview } from "./link-preview";
import { MessageActions } from "./message-actions";
import { MessageAttachments } from "./message-attachments";
import { pinsQueryKey } from "./pinned-messages";
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
	/** Fresh server-authoritative action capabilities. */
	capabilities?: ChannelCapabilities;
	/** Предыдущее сообщение (для группировки). */
	previousMessage?: Message;
	/** Показать разделитель даты перед сообщением. */
	showDateSeparator?: boolean;
	/** Текст разделителя даты. */
	dateSeparatorText?: string;
	/** Показать разделитель новых сообщений. */
	showNewMessagesDivider?: boolean;
}

interface DeleteScope {
	lifecycle: number;
	mounted: boolean;
	userId: string | undefined;
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
}

interface EditScope {
	lifecycle: number;
	mounted: boolean;
	userId: string;
	channelId: string;
	messageId: string;
	messageChannelId: string | undefined;
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

/** Режим редактирования сообщения. */
function MessageEditMode({
	text,
	onSave,
	onCancel,
}: {
	text: string;
	onSave: (newText: string) => void;
	onCancel: () => void;
}) {
	const [editText, setEditText] = useState(text);
	const [saving, setSaving] = useState(false);
	const inputRef = useRef<HTMLTextAreaElement>(null);

	useEffect(() => {
		inputRef.current?.focus();
	}, []);

	const handleKeyDown = useCallback(
		(e: React.KeyboardEvent<HTMLTextAreaElement>) => {
			if (e.key === "Enter" && !e.shiftKey) {
				e.preventDefault();
				setSaving(true);
				onSave(editText);
			} else if (e.key === "Escape") {
				e.preventDefault();
				onCancel();
			}
		},
		[editText, onSave, onCancel],
	);

	return (
		<div className="flex flex-col gap-1.5">
			<textarea
				ref={inputRef}
				value={editText}
				onChange={(e) => setEditText(e.target.value)}
				onKeyDown={handleKeyDown}
				disabled={saving}
				className="w-full resize-none rounded-lg border border-input bg-base px-3 py-2 text-sm text-t1 shadow-sm focus:outline-none focus:ring-2 focus:ring-ring/50"
				rows={Math.min(editText.split("\n").length + 1, 6)}
			/>
			<div className="flex items-center gap-1 justify-end">
				<button
					type="button"
					className="inline-flex items-center justify-center h-6 w-6 rounded-full hover:bg-accent text-t2 hover:text-t1 transition-colors"
					title="Отменить (Esc)"
					aria-label="Отменить редактирование"
					onClick={onCancel}
					disabled={saving}
				>
					<X className="h-3.5 w-3.5" aria-hidden="true" />
				</button>
				<button
					type="button"
					className="inline-flex items-center justify-center h-6 w-6 rounded-full hover:bg-accent text-primary hover:text-primary transition-colors"
					title="Сохранить (Enter)"
					aria-label="Сохранить сообщение"
					onClick={() => {
						setSaving(true);
						onSave(editText);
					}}
					disabled={saving || !editText.trim()}
				>
					<Check className="h-3.5 w-3.5" aria-hidden="true" />
				</button>
			</div>
		</div>
	);
}

/**
 * MessageItem отображает одно сообщение чата в стиле Slack / Tinkoff Talk.
 * Flat layout без пузырей: аватар слева, имя + время + контент справа.
 * Поддерживает группировку, реакции, hover-действия, ответы и вложения.
 */
function MessageItemInner({
	message,
	channelId,
	capabilities,
	previousMessage,
	showDateSeparator,
	dateSeparatorText,
	showNewMessagesDivider,
}: MessageItemProps) {
	const { t } = useTranslation("messenger");
	const [showReactionPicker, setShowReactionPicker] = useState(false);
	const [showForwardDialog, setShowForwardDialog] = useState(false);
	const [isEditing, setIsEditing] = useState(false);
	const [editText, setEditText] = useState("");
	const [editPending, setEditPending] = useState(false);
	const [deletePending, setDeletePending] = useState(false);
	const queryClient = useQueryClient();
	const openThread = useMessengerStore((s) => s.openThread);
	const setReplyTo = useMessengerStore((s) => s.setReplyTo);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const currentUserId = useAuthStore((s) => s.user?.id);
	const liveDeleteCapabilities = useMessengerStore((state) =>
		state.channels.find((item) => item.id === channelId)?.capabilities,
	);
	const deleteScopeRef = useRef<DeleteScope>({
		lifecycle: 0,
		mounted: false,
		userId: undefined,
		channelId: "",
		messageId: "",
		messageChannelId: undefined,
	});
	const deleteRequestRef = useRef<object | null>(null);
	const editScopeRef = useRef<EditScope | null>(null);
	const editLifecycleRef = useRef(0);
	const editRequestRef = useRef<EditScope | null>(null);
	const capabilitiesRef = useRef(capabilities);
	capabilitiesRef.current = capabilities;
	const loadBotInfo = useAIStore((s) => s.loadBotInfo);
	const isBotMessage = useAIStore((s) => s.isBotMessage);
	const botDisplayName = useAIStore((s) => s.botDisplayName);

	useEffect(() => {
		void loadBotInfo();
	}, [loadBotInfo]);

	// NOTE: все хуки до early return (Rules of Hooks, FE-01).
	// Guard для битых сообщений — ПОСЛЕ всех хуков.

	// Стабильные id для хуков ниже; при битом message хуки всё равно вызываем,
	// чтобы порядок вызовов не менялся между renders.
	const safeMessageId =
		message && typeof message.id === "string" ? message.id : "";
	const messageContentText = message?.content_text;
	const messageContentJson = message?.content_json;

	useLayoutEffect(() => {
		const lifecycle = deleteScopeRef.current.lifecycle + 1;
		deleteScopeRef.current = {
			lifecycle,
			mounted: !channelRevoked,
			userId: currentUserId,
			channelId,
			messageId: safeMessageId,
			messageChannelId: message?.channel_id,
		};
		deleteRequestRef.current = null;
		setDeletePending(false);
		return () => {
			if (deleteScopeRef.current.lifecycle !== lifecycle) return;
			deleteScopeRef.current = {
				...deleteScopeRef.current,
				lifecycle: lifecycle + 1,
				mounted: false,
			};
			deleteRequestRef.current = null;
		};
	}, [
		capabilities?.can_delete_message,
		capabilities?.is_member,
		channelId,
		channelRevoked,
		currentUserId,
		liveDeleteCapabilities?.can_delete_message,
		liveDeleteCapabilities?.is_member,
		message?.channel_id,
		message?.sender_id,
		safeMessageId,
	]);

	useLayoutEffect(() => {
		const scope: EditScope = {
			lifecycle: editLifecycleRef.current + 1,
			mounted: !channelRevoked,
			userId: currentUserId ?? "",
			channelId,
			messageId: safeMessageId,
			messageChannelId: message?.channel_id,
		};
		editLifecycleRef.current = scope.lifecycle;
		editScopeRef.current = scope;
		editRequestRef.current = null;
		setEditPending(false);
		setIsEditing(false);
		setEditText("");
		return () => {
			if (editScopeRef.current !== scope) return;
			editScopeRef.current = { ...scope, mounted: false };
			editRequestRef.current = null;
		};
	}, [
		capabilities?.can_edit,
		channelId,
		channelRevoked,
		currentUserId,
		liveDeleteCapabilities?.can_edit,
		message?.channel_id,
		message?.sender_id,
		safeMessageId,
	]);

	const canPinNow = useCallback(() => {
		const state = useMessengerStore.getState();
		if (state.isChannelRevoked(channelId)) return false;
		const currentCapabilities = state.channels.find(
			(item) => item.id === channelId,
		)?.capabilities;
		return (
			capabilitiesRef.current?.can_pin === true &&
			currentCapabilities?.can_pin === true
		);
	}, [channelId]);

	const canAct = useCallback(
		(capability: ChannelCapability) => {
			const state = useMessengerStore.getState();
			if (state.isChannelRevoked(channelId)) return false;
			const currentChannel = state.channels.find((item) => item.id === channelId);
			return (
				capabilities?.[capability] === true &&
				currentChannel?.capabilities?.[capability] === true
			);
		},
		[capabilities, channelId],
	);
	const isOwnMessage =
		currentUserId !== undefined && currentUserId === message?.sender_id;
	const canReact = capabilities?.can_react === true && !channelRevoked;
	const canForward = capabilities?.can_forward === true && !channelRevoked;
	const canEdit =
		capabilities?.can_edit === true &&
		liveDeleteCapabilities?.can_edit === true &&
		!channelRevoked &&
		isOwnMessage &&
		message?.channel_id === channelId;
	const canDeleteMessage = useCallback(() => {
		const state = useMessengerStore.getState();
		if (state.isChannelRevoked(channelId)) return false;
		const currentCapabilities = state.channels.find(
			(item) => item.id === channelId,
		)?.capabilities;
		const canDeleteOwn =
			isOwnMessage &&
			capabilities?.is_member === true &&
			currentCapabilities?.is_member === true;
		const canDeleteAny =
			capabilities?.can_delete_message === true &&
			currentCapabilities?.can_delete_message === true;
		return canDeleteOwn || canDeleteAny;
	}, [capabilities, channelId, isOwnMessage]);

	const isCurrentDeleteScope = useCallback(
		(
			request: object,
			lifecycle: number,
			capturedUserId: string,
			capturedChannelId: string,
			capturedMessageId: string,
		) => {
			const scope = deleteScopeRef.current;
			return (
				deleteRequestRef.current === request &&
				scope.lifecycle === lifecycle &&
				scope.mounted &&
				scope.userId === capturedUserId &&
				scope.channelId === capturedChannelId &&
				scope.messageId === capturedMessageId &&
				scope.messageChannelId === capturedChannelId &&
				useAuthStore.getState().user?.id === capturedUserId &&
				!useMessengerStore.getState().isChannelRevoked(capturedChannelId)
			);
		},
		[],
	);

	useEffect(() => {
		if (channelRevoked || !canReact) setShowReactionPicker(false);
		if (channelRevoked || !canForward) setShowForwardDialog(false);
		if (channelRevoked || !canEdit) setIsEditing(false);
	}, [canEdit, canForward, canReact, channelRevoked]);

	const isCurrentEditScope = useCallback(
		(scope: EditScope) => {
			const state = useMessengerStore.getState();
			return (
				editScopeRef.current === scope &&
				scope.mounted &&
				editLifecycleRef.current === scope.lifecycle &&
				scope.userId.length > 0 &&
				scope.messageId.length > 0 &&
				scope.userId === (currentUserId ?? "") &&
				scope.channelId === channelId &&
				scope.messageId === safeMessageId &&
				scope.messageChannelId === scope.channelId &&
				useAuthStore.getState().user?.id === scope.userId &&
				!state.isChannelRevoked(scope.channelId) &&
				capabilities?.can_edit === true &&
				state.channels.find((item) => item.id === scope.channelId)?.capabilities
					?.can_edit === true
			);
		},
		[capabilities, channelId, currentUserId, safeMessageId],
	);

	/**
	 * Извлекает уникальные URL из текста сообщения.
	 */
	const extractedUrls = useMemo(() => {
		const urls = new Set<string>();
		const addUrl = (u: string) => {
			// A3W9-3-003: только whitelist-схемы попадают в LinkPreview.
			if (isSafeUrl(u)) urls.add(u);
		};
		if (messageContentText) {
			const matches = messageContentText.match(URL_REGEX);
			if (matches) {
				for (const u of matches) {
					addUrl(u);
				}
			}
		}
		if (messageContentJson) {
			const extractFromNode = (node: unknown) => {
				if (!node || typeof node !== "object") return;
				const n = node as Record<string, unknown>;
				if (n.type === "a" && typeof n.url === "string") {
					addUrl(n.url);
				}
				if (typeof n.text === "string") {
					const matches = n.text.match(URL_REGEX);
					if (matches) {
						for (const u of matches) {
							addUrl(u);
						}
					}
				}
				if (Array.isArray(n.children)) {
					n.children.forEach(extractFromNode);
				}
			};
			extractFromNode(messageContentJson);
		}
		return Array.from(urls);
	}, [messageContentText, messageContentJson]);

	const handleDelete = useCallback(async () => {
		if (deleteRequestRef.current) return;
		const capturedUserId = currentUserId;
		const capturedChannelId = channelId;
		const capturedMessageId = safeMessageId;
		if (!capturedUserId || !capturedMessageId || message?.channel_id !== capturedChannelId) {
			return;
		}
		const request = {};
		const lifecycle = deleteScopeRef.current.lifecycle;
		deleteRequestRef.current = request;
		if (
			!isCurrentDeleteScope(request, lifecycle, capturedUserId, capturedChannelId, capturedMessageId) ||
			!canDeleteMessage()
		) {
			if (deleteRequestRef.current === request) deleteRequestRef.current = null;
			return;
		}
		setDeletePending(true);
		try {
			await deleteMessage(capturedChannelId, capturedMessageId);
		} catch (error) {
			if (
				isCurrentDeleteScope(request, lifecycle, capturedUserId, capturedChannelId, capturedMessageId) &&
				canDeleteMessage()
			) {
				toastApiError(error, t("messageDelete.error"));
			}
		} finally {
			if (isCurrentDeleteScope(request, lifecycle, capturedUserId, capturedChannelId, capturedMessageId)) {
				deleteRequestRef.current = null;
				setDeletePending(false);
			}
		}
	}, [
		canDeleteMessage,
		channelId,
		currentUserId,
		isCurrentDeleteScope,
		message?.channel_id,
		safeMessageId,
		t,
	]);

	const handlePin = useCallback(async () => {
		const capturedUserId = currentUserId;
		const capturedChannelId = channelId;
		if (
			!capturedUserId ||
			safeMessageId.length === 0 ||
			useAuthStore.getState().user?.id !== capturedUserId ||
			!canPinNow()
		) {
			return;
		}

		await togglePin(capturedChannelId, safeMessageId);
		if (
			useAuthStore.getState().user?.id !== capturedUserId ||
			!canPinNow()
		) {
			return;
		}
		await queryClient.invalidateQueries({
			queryKey: pinsQueryKey(capturedChannelId, capturedUserId),
			exact: true,
		});
	}, [
		canPinNow,
		channelId,
		currentUserId,
		queryClient,
		safeMessageId,
	]);

	const handleBookmark = useCallback(async () => {
		if (!canAct("can_bookmark")) return;
		await addBookmark(safeMessageId, channelId);
	}, [canAct, channelId, safeMessageId]);
	/** Ответить на сообщение (установить reply-to). */
	const handleReply = useCallback(() => {
		if (message && canAct("can_reply")) setReplyTo(message);
	}, [canAct, message, setReplyTo]);

	const handleOpenThread = useCallback(() => {
		if (canAct("can_read")) openThread(channelId, safeMessageId);
	}, [canAct, channelId, openThread, safeMessageId]);

	const handleReaction = useCallback(
		async (emoji: string) => {
			if (!canAct("can_react")) return;
			await toggleReaction(channelId, safeMessageId, emoji);
		},
		[canAct, channelId, safeMessageId],
	);

	/** Начать редактирование сообщения. */
	const handleStartEdit = useCallback(() => {
		const scope = editScopeRef.current;
		if (editRequestRef.current || isEditing || !scope) return;
		if (!canEdit || !isCurrentEditScope(scope)) return;
		setEditText(messageContentText ?? "");
		setIsEditing(true);
	}, [canEdit, isCurrentEditScope, isEditing, messageContentText]);
	/** Сохранить отредактированное сообщение. */
	const handleSaveEdit = useCallback(
		async (newText?: string) => {
			const scope = editScopeRef.current;
			if (
				!scope ||
				editRequestRef.current ||
				!canEdit ||
				!isCurrentEditScope(scope)
			) {
				setIsEditing(false);
				return;
			}
			const trimmed = (newText ?? editText).trim();
			if (!trimmed || trimmed === (messageContentText ?? "").trim()) {
				setIsEditing(false);
				return;
			}
			const queryKey = messagesQueryKey(scope.channelId, scope.userId);
			const contentJson = {
				type: "doc",
				children: [{ type: "p", children: [{ text: trimmed }] }],
			};
			editRequestRef.current = scope;
			setEditPending(true);
			setIsEditing(false);
			try {
				queryClient.setQueryData(queryKey, (old: unknown) => {
					if (!old) return old;
					const typed = old as { pages: Message[][]; pageParams: unknown[] };
					return {
						...typed,
						pages: typed.pages.map((page) =>
							page.map((msg) =>
								msg.id === scope.messageId
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
				if (!isCurrentEditScope(scope)) return;
				await editMessage(scope.channelId, scope.messageId, {
					content_json: contentJson,
					content_text: trimmed,
				});
			} catch (error) {
				if (
					editRequestRef.current === scope &&
					isCurrentEditScope(scope)
				) {
					void queryClient.invalidateQueries({ queryKey, exact: true });
					toastApiError(error, t("messageEdit.error"));
				}
			} finally {
				if (
					editRequestRef.current === scope &&
					isCurrentEditScope(scope)
				) {
					editRequestRef.current = null;
					setEditPending(false);
				}
			}
		},
		[
			canEdit,
			editText,
			isCurrentEditScope,
			messageContentText,
			queryClient,
			t,
		],
	);

	/** Отменить редактирование. */
	const handleCancelEdit = useCallback(() => {
		setIsEditing(false);
		setEditText("");
	}, []);

	// Защита: не рендерим битое сообщение (частичный WS event).
	// Только ПОСЛЕ всех хуков — иначе нарушим Rules of Hooks.
	if (channelRevoked || !message || !message.id || typeof message.id !== "string") {
		return null;
	}

	const isOwn = isOwnMessage;
	const isGrouped = shouldGroupWithPrevious(message, previousMessage);
	const fromBot =
		isBotMessage(message.sender_id) ||
		isHubAIMarkdown(
			message.content_json as Record<string, unknown> | undefined,
		);

	return (
		<>
			{/* Разделитель дат — pill по центру */}
			{showDateSeparator && dateSeparatorText && (
				<div className="flex items-center justify-center py-4 px-5">
					<div className="text-[11.5px] font-semibold text-[var(--t3)] bg-[var(--raised)] px-4 py-1.5 rounded-full border border-[var(--b1)] shadow-[var(--sh2)]">
						{dateSeparatorText}
					</div>
				</div>
			)}

			{/* Разделитель «Новые сообщения» */}
			{showNewMessagesDivider && (
				<div className="flex items-center gap-3.5 py-3 px-5">
					<div className="flex-1 h-px bg-[var(--red)] opacity-50" />
					<span className="text-[11.5px] font-semibold text-[var(--red)] whitespace-nowrap">
						Новые сообщения
					</span>
					<div className="flex-1 h-px bg-[var(--red)] opacity-50" />
				</div>
			)}

			{/* Сообщение — плоская Slack-раскладка */}
			<div
				className={cn(
					"group msg-row flex gap-3 py-2 px-2.5 rounded-[var(--r3)] transition-[background] relative",
					"hover:bg-[var(--hover)]",
				)}
			>
				<div className="flex gap-3 w-full min-w-0">
					{/* Колонка аватара 38px — только у первого в группе */}
					<div
						className={cn("w-[38px] shrink-0", isGrouped && "invisible h-0")}
					>
						{!isGrouped && (
							<Avatar className="w-[38px] h-[38px]">
								<AvatarFallback className="w-[38px] h-[38px] rounded-full flex items-center justify-center text-[13px] font-bold text-white hover:scale-[1.06] hover:shadow-[0_2px_12px_rgba(0,0,0,.3)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] bg-primary/10 text-primary">
									{message.sender_id?.charAt(0)?.toUpperCase() ?? "?"}
								</AvatarFallback>
							</Avatar>
						)}
						{/* Время при hover у сгруппированных сообщений */}
						{isGrouped && (
							<span className="hidden group-hover:flex items-center justify-center h-full text-[10px] text-t2/60">
								{formatTime(message.created_at)}
							</span>
						)}
					</div>

					{/* Контент сообщения */}
					<div className="min-w-0 flex-1">
						{/* Имя + время — только у первого в группе */}
						{!isGrouped && (
							<div className="flex items-baseline gap-2 mb-0.5">
								<span
									className={cn(
										"text-[13.5px] font-bold tracking-[-0.01em] hover:underline cursor-pointer",
										fromBot && "text-[var(--purple)]",
									)}
								>
									{fromBot ? botDisplayName : message.sender_id}
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

						{/* Бейдж «Переслано» */}
						{message.forwarded_from_id && (
							<div className="mb-1 flex items-center gap-1 text-xs text-t2/70">
								<svg
									aria-hidden="true"
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
							<PollMessage pollId={message.poll_id} channelId={channelId} />
						)}

						{/* Контент сообщения (или inline-редактирование) — скрываем для опросов */}
						{message.message_type !== "poll" &&
							(isEditing ? (
								<MessageEditMode
									text={editText}
									onSave={(newText) => void handleSaveEdit(newText)}
									onCancel={handleCancelEdit}
								/>
							) : (
								<div className="text-sm leading-relaxed text-[var(--t1)] break-words">
									{fromBot && message.content_text ? (
										<BotMarkdown text={message.content_text} />
									) : (
										<>
											{message.content_json && (
												<RenderContent contentJson={message.content_json} />
											)}
											{!message.content_json && message.content_text && (
												<p className="whitespace-pre-wrap">
													{message.content_text}
												</p>
											)}
										</>
									)}
								</div>
							))}

						{/* Голосовое сообщение */}
						{message.voice_url && (
							<div className="mt-1.5 flex items-center gap-2 rounded-lg bg-raised/50 p-2 max-w-sm">
								<audio
									controls
									preload="metadata"
									className="h-8 max-w-xs"
									aria-label="Голосовое сообщение"
								>
									<source src={message.voice_url} type="audio/webm" />
									<track kind="captions" />
								</audio>
								{message.voice_duration_ms && (
									<span className="text-xs text-t2">
										{Math.floor(message.voice_duration_ms / 1000)}с
									</span>
								)}
							</div>
						)}

						{/* Вложения */}
						{message.attachments && message.attachments.length > 0 && (
							<MessageAttachments
								attachments={message.attachments}
								isOwn={isOwn}
							/>
						)}

						{/* Превью ссылок */}
						{extractedUrls.length > 0 && (
							<div className="mt-1.5 space-y-1.5">
								{extractedUrls.slice(0, 2).map((url) => (
									<LinkPreview key={url} url={url} />
								))}
							</div>
						)}

						{/* Реакции — компактные emoji-чипы */}
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
												hasReacted &&
													"border-[color-mix(in_srgb,var(--green)_20%,transparent)] bg-[var(--gG)]",
											)}
											disabled={!canReact || channelRevoked}
											onClick={() => void handleReaction(reaction.emoji)}
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

						{/* Индикатор треда */}
						{message.reply_count > 0 && (
							<button
								type="button"
								className="mt-1.5 inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium text-primary hover:bg-primary/10 transition-colors"
								onClick={handleOpenThread}
							>
								<svg
									aria-hidden="true"
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

				{/* Hover-действия — pill-toolbar справа */}
				{!isEditing && !channelRevoked && (
					<div
						role="toolbar"
						aria-label="Действия с сообщением"
						className="pointer-events-none opacity-0 transition-opacity group-hover:pointer-events-auto group-hover:opacity-100 group-focus-within:pointer-events-auto group-focus-within:opacity-100 focus-within:pointer-events-auto focus-within:opacity-100"
						onMouseLeave={() => setShowReactionPicker(false)}
					>
						<MessageActions
							message={message}
							channelId={channelId}
							isOwn={isOwn}
							capabilities={capabilities}
							showReactionPicker={showReactionPicker}
							onReactionPickerChange={setShowReactionPicker}
							onReply={handleReply}
							onOpenThread={handleOpenThread}
							onPin={handlePin}
							onBookmark={handleBookmark}
							onForward={() => {
								if (canAct("can_forward")) setShowForwardDialog(true);
							}}
							onEdit={handleStartEdit}
							onDelete={handleDelete}
							editPending={editPending}
							deletePending={deletePending}
						/>
					</div>
				)}

				{/* Диалог пересылки */}
				<ForwardDialog
					message={message}
					channelId={channelId}
					canForward={canForward}
					open={showForwardDialog}
					onOpenChange={setShowForwardDialog}
				/>
			</div>
		</>
	);
}

export const MessageItem = memo(MessageItemInner);
