import { useQuery } from "@tanstack/react-query";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Channel, Message } from "@/api/messenger";
import { getChannel } from "@/api/messenger";
import { CallButton } from "@/components/calls/CallButton";
import { PinnedMessages } from "./pinned-messages";
import { SearchDialog } from "./search-dialog";
import { useMessages } from "@/hooks/use-messages";
import { useTypingIndicator } from "@/hooks/use-typing";
import { useMarkRead } from "@/hooks/use-unread";
import { cn } from "@/lib/utils";
import { useMessengerStore } from "@/stores/messenger-store";
import { MessageComposer } from "./message-composer";
import { MessageItem } from "./message-item";
import { ThreadPanel } from "./thread-panel";

/** Пропсы компонента ChatView. */
interface ChatViewProps {
	/** ID канала для отображения. */
	channelId: string;
}

/**
 * Возвращает текст разделителя даты на русском языке.
 * "Сегодня", "Вчера" или "15 марта 2026".
 */
function getDateSeparatorText(dateStr: string | undefined | null): string {
	if (!dateStr || typeof dateStr !== "string") return "";
	const date = new Date(dateStr);
	if (Number.isNaN(date.getTime())) return "";
	const now = new Date();
	const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const yesterday = new Date(today);
	yesterday.setDate(yesterday.getDate() - 1);
	const msgDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());

	if (msgDate.getTime() === today.getTime()) return "Сегодня";
	if (msgDate.getTime() === yesterday.getTime()) return "Вчера";

	return date.toLocaleDateString("ru-RU", {
		day: "numeric",
		month: "long",
		year: date.getFullYear() !== now.getFullYear() ? "numeric" : undefined,
	});
}

/**
 * Проверяет, нужен ли разделитель даты между двумя сообщениями.
 */
function needsDateSeparator(current: Message, previous?: Message): boolean {
	if (!previous) return true;
	if (!current.created_at || !previous.created_at) return true;
	const curDate = new Date(current.created_at);
	const prevDate = new Date(previous.created_at);
	if (Number.isNaN(curDate.getTime()) || Number.isNaN(prevDate.getTime())) return true;
	return curDate.toDateString() !== prevDate.toDateString();
}

/**
 * ChatView -- основной компонент чата в стиле Tinkoff Talk.
 * Виртуализация, разделители дат, индикатор печати,
 * кнопка прокрутки вниз со счётчиком.
 */
export function ChatView({ channelId }: ChatViewProps) {
	const { messages, loadMore, hasMore, isLoading } = useMessages(channelId);
	const markRead = useMarkRead();
	const activeThreadMessageId = useMessengerStore(
		(s) => s.activeThreadMessageId,
	);
	const closeThread = useMessengerStore((s) => s.closeThread);
	const typingUsers = useMessengerStore((s) => s.typingUsers[channelId]) ?? [];

	// Подписка на typing-события для текущего канала.
	useTypingIndicator(channelId);

	// Пометить канал как прочитанный при открытии.
	useEffect(() => {
		if (channelId) {
			void markRead(channelId);
		}
	}, [channelId, markRead]);

	const { data: channel } = useQuery<Channel>({
		queryKey: ["messenger", "channel", channelId],
		queryFn: () => getChannel(channelId),
	});

	const parentRef = useRef<HTMLDivElement>(null);
	const [isAtBottom, setIsAtBottom] = useState(true);
	const [unreadBelowCount, setUnreadBelowCount] = useState(0);

	const virtualizer = useVirtualizer({
		count: messages.length,
		getScrollElement: () => parentRef.current,
		estimateSize: () => 80,
		overscan: 5,
	});

	// Автоскролл вниз при новых сообщениях (если пользователь внизу).
	const prevCountRef = useRef(messages.length);
	useEffect(() => {
		if (messages.length > prevCountRef.current) {
			if (isAtBottom) {
				virtualizer.scrollToIndex(messages.length - 1, { align: "end" });
			} else {
				setUnreadBelowCount(
					(c) => c + (messages.length - prevCountRef.current),
				);
			}
		}
		prevCountRef.current = messages.length;
	}, [messages.length, isAtBottom, virtualizer]);

	// Первоначальная прокрутка вниз.
	useEffect(() => {
		if (messages.length > 0 && prevCountRef.current === messages.length) {
			virtualizer.scrollToIndex(messages.length - 1, { align: "end" });
		}
	}, [channelId]); // eslint-disable-line react-hooks/exhaustive-deps

	// Определение позиции скролла.
	const handleScroll = useCallback(() => {
		const el = parentRef.current;
		if (!el) return;
		const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50;
		setIsAtBottom(atBottom);
		if (atBottom) setUnreadBelowCount(0);
		if (el.scrollTop < 100 && hasMore) {
			loadMore();
		}
	}, [hasMore, loadMore]);

	/** Прокрутка вниз по клику. */
	const scrollToBottom = useCallback(() => {
		if (messages.length > 0) {
			virtualizer.scrollToIndex(messages.length - 1, { align: "end" });
			setUnreadBelowCount(0);
		}
	}, [messages.length, virtualizer]);

	/** Текст индикатора печати. */
	const typingText = useMemo(() => {
		if (typingUsers.length === 0) return null;
		if (typingUsers.length === 1)
			return `${typingUsers[0]!.displayName} печатает`;
		if (typingUsers.length === 2)
			return `${typingUsers[0]!.displayName} и ${typingUsers[1]!.displayName} печатают`;
		return `${typingUsers[0]!.displayName} и ещё ${typingUsers.length - 1} печатают`;
	}, [typingUsers]);

	return (
		<div className="flex flex-1 flex-col overflow-hidden">
			{/* Заголовок канала */}
			<div className="flex h-14 shrink-0 items-center justify-between border-b bg-card/60 backdrop-blur-sm px-5">
				<div className="flex items-center gap-3">
					{channel && (
						<div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
							{channel.channel_type === "public" ? (
								<span className="text-sm font-bold">#</span>
							) : channel.channel_type === "dm" ? (
								<svg
									className="h-4 w-4"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={2}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
									/>
								</svg>
							) : (
								<svg
									className="h-4 w-4"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={2}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"
									/>
								</svg>
							)}
						</div>
					)}
					<div>
						<h2 className="text-sm font-semibold leading-tight">
							{channel?.name ?? "Загрузка..."}
						</h2>
						{channel && channel.channel_type !== "dm" && (
							<span className="text-[11px] text-muted-foreground">
								{channel.channel_type === "public"
									? "Публичный"
									: channel.channel_type === "private"
										? "Приватный"
										: "Группа"}
							</span>
						)}
					</div>
				</div>
				<div className="flex items-center gap-1">
					<CallButton channelId={channelId} />
					<SearchDialog channelId={channelId} />
					<PinnedMessages channelId={channelId} />
				</div>
			</div>

			{/* Область сообщений */}
			<div className="flex flex-1 overflow-hidden">
				<div className="flex flex-1 flex-col overflow-hidden relative">
					{isLoading ? (
						/* Скелетон загрузки -- имитация пузырей */
						<div className="flex flex-1 flex-col justify-end p-4 space-y-4">
							{[1, 2, 3, 4, 5].map((i) => (
								<div
									key={i}
									className={cn(
										"flex gap-3",
										i % 3 === 0 ? "justify-end" : "justify-start",
									)}
								>
									{i % 3 !== 0 && (
										<div className="h-9 w-9 rounded-full animate-shimmer shrink-0" />
									)}
									<div className="space-y-1.5">
										<div className="h-3 w-20 rounded animate-shimmer" />
										<div
											className={cn(
												"rounded-2xl animate-shimmer",
												i % 3 === 0 ? "h-10 w-48" : "h-12 w-56",
											)}
										/>
									</div>
								</div>
							))}
						</div>
					) : messages.length === 0 ? (
						/* Пустое состояние */
						<div className="flex flex-1 items-center justify-center">
							<div className="text-center max-w-xs animate-fade-in">
								<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500/10 to-teal-500/10">
									<svg
										className="h-8 w-8 text-emerald-500/50"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={1.5}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
										/>
									</svg>
								</div>
								<h3 className="text-base font-medium text-foreground mb-1">
									Нет сообщений
								</h3>
								<p className="text-sm text-muted-foreground">
									Начните общение! Напишите первое сообщение в этом канале.
								</p>
							</div>
						</div>
					) : (
						<div
							ref={parentRef}
							className="flex-1 overflow-y-auto scroll-smooth"
							onScroll={handleScroll}
						>
							<div
								className="relative w-full"
								style={{ height: `${virtualizer.getTotalSize()}px` }}
							>
								{virtualizer.getVirtualItems().map((virtualItem) => {
									const message = messages[virtualItem.index];
									if (!message) return null;
									const prevMessage =
										virtualItem.index > 0
											? messages[virtualItem.index - 1]
											: undefined;
									const showDateSep = needsDateSeparator(message, prevMessage);
									return (
										<div
											key={virtualItem.key}
											data-index={virtualItem.index}
											ref={virtualizer.measureElement}
											className="absolute left-0 top-0 w-full"
											style={{
												transform: `translateY(${virtualItem.start}px)`,
											}}
										>
											<MessageItem
												message={message}
												channelId={channelId}
												previousMessage={prevMessage}
												showDateSeparator={showDateSep}
												dateSeparatorText={
													showDateSep
														? getDateSeparatorText(message.created_at)
														: undefined
												}
											/>
										</div>
									);
								})}
							</div>
						</div>
					)}

					{/* Кнопка "Прокрутить вниз" -- плавающая */}
					{!isAtBottom && messages.length > 0 && (
						<button
							type="button"
							onClick={scrollToBottom}
							className="absolute bottom-20 right-6 z-20 flex items-center gap-1.5 rounded-full bg-card border shadow-lg px-3 py-2 text-sm font-medium text-foreground hover:bg-accent transition-all animate-fade-in"
						>
							<svg
								className="h-4 w-4"
								fill="none"
								viewBox="0 0 24 24"
								stroke="currentColor"
								strokeWidth={2}
							>
								<path
									strokeLinecap="round"
									strokeLinejoin="round"
									d="M19 14l-7 7m0 0l-7-7m7 7V3"
								/>
							</svg>
							{unreadBelowCount > 0 && (
								<span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground px-1.5">
									{unreadBelowCount > 99 ? "99+" : unreadBelowCount}
								</span>
							)}
						</button>
					)}

					{/* Индикатор печати */}
					{typingText && (
						<div className="shrink-0 px-5 py-1.5 text-xs text-muted-foreground flex items-center gap-1.5">
							<span className="typing-dots">
								<span className="dot" />
								<span className="dot" />
								<span className="dot" />
							</span>
							<span>{typingText}...</span>
						</div>
					)}

					{/* Композер */}
					{!channel?.is_readonly && (
						<div className="shrink-0 border-t glass-subtle px-4 py-3">
							<MessageComposer channelId={channelId} />
						</div>
					)}
				</div>

				{/* Thread panel */}
				{activeThreadMessageId && (
					<div
						className={cn(
							"w-96 shrink-0 border-l",
							"max-md:fixed max-md:inset-y-0 max-md:right-0 max-md:z-50 max-md:w-full max-md:bg-background",
						)}
					>
						<ThreadPanel
							channelId={channelId}
							messageId={activeThreadMessageId}
							onClose={closeThread}
						/>
					</div>
				)}
			</div>
		</div>
	);
}
