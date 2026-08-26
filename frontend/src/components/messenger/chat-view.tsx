import { useQuery } from "@tanstack/react-query";
import { useVirtualizer } from "@tanstack/react-virtual";
import { Menu } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Channel, Message } from "@/api/messenger";
import { canChannel, getChannel } from "@/api/messenger";
import { CallButton } from "@/components/calls/CallButton";
import { useMessages } from "@/hooks/use-messages";
import { useTypingIndicator } from "@/hooks/use-typing";
import { useMarkRead } from "@/hooks/use-unread";
import { cn } from "@/lib/utils";
import { useMessengerStore } from "@/stores/messenger-store";
import { useChannelDrawer } from "./channel-drawer-context";
import { InviteLinkDialog } from "./invite-link-dialog";
import { MessageComposer } from "./message-composer";
import { MessageItem } from "./message-item";
import { PinnedMessages } from "./pinned-messages";
import { SearchDialog } from "./search-dialog";
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
	if (Number.isNaN(curDate.getTime()) || Number.isNaN(prevDate.getTime()))
		return true;
	return curDate.toDateString() !== prevDate.toDateString();
}

/** Иконка типа канала для заголовка. */
function ChannelHeaderIcon({ type }: { type: string }) {
	if (type === "public") {
		return <span className="text-base font-bold text-t2">#</span>;
	}
	if (type === "dm") {
		return (
			<svg
				aria-hidden="true"
				className="h-5 w-5 text-t2"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				strokeWidth={1.5}
			>
				<path
					strokeLinecap="round"
					strokeLinejoin="round"
					d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
				/>
			</svg>
		);
	}
	if (type === "private") {
		return (
			<svg
				aria-hidden="true"
				className="h-5 w-5 text-t2"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				strokeWidth={1.5}
			>
				<path
					strokeLinecap="round"
					strokeLinejoin="round"
					d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
				/>
			</svg>
		);
	}
	return (
		<svg
			aria-hidden="true"
			className="h-5 w-5 text-t2"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={1.5}
		>
			<path
				strokeLinecap="round"
				strokeLinejoin="round"
				d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"
			/>
		</svg>
	);
}

/**
 * ChatView -- основной компонент чата в стиле Slack / Tinkoff Talk.
 * Виртуализация, разделители дат, индикатор печати,
 * кнопка прокрутки вниз со счётчиком.
 */
export function ChatView({ channelId }: ChatViewProps) {
	const { messages, loadMore, hasMore, isLoading } = useMessages(channelId);
	const markRead = useMarkRead();
	const activeThread = useMessengerStore((state) => state.activeThread);
	const channelRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const closeThread = useMessengerStore((s) => s.closeThread);
	const typingUsers = useMessengerStore((s) => s.typingUsers[channelId]) ?? [];
	const { setOpen: setDrawerOpen } = useChannelDrawer();

	// Подписка на typing-события для текущего канала.
	useTypingIndicator(channelId);

	// Пометить канал как прочитанный при открытии.
	useEffect(() => {
		if (
			channelId &&
			!channelRevoked &&
			!useMessengerStore.getState().isChannelRevoked(channelId)
		) {
			void markRead(channelId);
		}
	}, [channelId, channelRevoked, markRead]);

	const { data: channel } = useQuery<Channel>({
		queryKey: ["messenger", "channel", channelId],
		queryFn: () => getChannel(channelId),
		enabled: !channelRevoked,
	});
	const canPost = channel ? canChannel(channel, "can_post") : false;
	const canReply = channel ? canChannel(channel, "can_reply") : false;

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

	// Первоначальная прокрутка вниз. Намеренно реагируем только на смену
	// channelId — messages.length и virtualizer читаются императивно, а
	// повторные срабатывания при каждом новом сообщении сломают UX.
	// biome-ignore lint/correctness/useExhaustiveDependencies: только смена channelId; остальное императивно.
	useEffect(() => {
		if (messages.length > 0 && prevCountRef.current === messages.length) {
			virtualizer.scrollToIndex(messages.length - 1, { align: "end" });
		}
	}, [channelId]);

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
		const first = typingUsers[0];
		if (!first) return null;
		if (typingUsers.length === 1) return `${first.displayName} печатает`;
		const second = typingUsers[1];
		if (typingUsers.length === 2 && second)
			return `${first.displayName} и ${second.displayName} печатают`;
		return `${first.displayName} и ещё ${typingUsers.length - 1} печатают`;
	}, [typingUsers]);

	if (channelRevoked) return null;

	return (
		<div className="flex flex-1 flex-col overflow-hidden">
			{/* Заголовок канала */}
			<div className="h-[58px] shrink-0 flex items-center px-4 md:px-[22px] gap-2.5 md:gap-3.5 border-b border-[var(--b1)] bg-[color-mix(in_srgb,var(--void)_70%,transparent)] backdrop-blur-[24px] backdrop-saturate-[180%] relative z-20">
				{/* Mobile: гамбургер для открытия списка каналов */}
				<button
					type="button"
					className="md:hidden w-[34px] h-[34px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t2)] hover:bg-[var(--hover)] hover:text-[var(--t1)] transition-colors shrink-0"
					onClick={() => setDrawerOpen(true)}
					aria-label="Открыть список каналов"
					title="Каналы"
				>
					<Menu className="h-5 w-5" />
				</button>
				<div className="flex items-center gap-3.5 min-w-0 flex-1">
					{channel && (
						<div className="w-[34px] h-[34px] rounded-full bg-[var(--gG)] text-[var(--green)] flex items-center justify-center text-sm font-extrabold shadow-[0_0_12px_color-mix(in_srgb,var(--green)_6%,transparent)] hover:scale-110 transition-transform">
							<ChannelHeaderIcon type={channel.channel_type} />
						</div>
					)}
					<div className="min-w-0">
						<h2 className="text-[15px] font-bold tracking-[-0.015em] truncate">
							{channel?.name ?? "Загрузка..."}
						</h2>
						{channel && channel.channel_type !== "dm" && (
							<span className="text-[11.5px] text-[var(--t3)] font-medium">
								{channel.channel_type === "public"
									? "Публичный канал"
									: channel.channel_type === "private"
										? "Приватный канал"
										: "Группа"}
							</span>
						)}
					</div>
				</div>
				<div className="flex items-center gap-0.5">
					{channel && channel.channel_type !== "dm" && (
						<InviteLinkDialog channelId={channelId} />
					)}
					<CallButton channelId={channelId} />
					<SearchDialog channelId={channelId} />
					<PinnedMessages
				channelId={channelId}
				capabilities={channel?.capabilities}
			/>
				</div>
				{/* Градиентная линия под заголовком */}
				<div className="absolute bottom-[-1px] left-[5%] right-[5%] h-px bg-gradient-to-r from-transparent via-[color-mix(in_srgb,var(--green)_6%,transparent)] to-transparent" />
			</div>

			{/* Область сообщений */}
			<div className="flex flex-1 overflow-hidden">
				<div className="flex flex-1 flex-col overflow-hidden relative">
					{isLoading ? (
						<>
							{/* Skeleton загрузки */}
							<div className="flex flex-1 flex-col justify-end p-4 space-y-3">
								{[1, 2, 3, 4, 5].map((i) => (
									<div key={`msg-sk-${i}`} className="flex gap-3">
										<div className="h-9 w-9 rounded-full animate-shimmer shrink-0" />
										<div className="space-y-1.5 flex-1">
											<div className="h-3 w-24 rounded animate-shimmer" />
											<div className="h-4 w-56 rounded animate-shimmer" />
										</div>
									</div>
								))}
							</div>
						</>
					) : messages.length === 0 ? (
						<>
							{/* Пустое состояние */}
							<div className="flex flex-1 items-center justify-center">
								<div className="text-center max-w-xs animate-fade-in">
									<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary/10 to-primary/5">
										<svg
											aria-hidden="true"
											className="h-8 w-8 text-primary/40"
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
									<h3 className="text-base font-medium text-t1 mb-1">
										Начните общение
									</h3>
									<p className="text-sm text-t2">
										Напишите первое сообщение в этом канале.
									</p>
								</div>
							</div>
						</>
					) : (
						<div
							ref={parentRef}
							className="flex-1 overflow-y-auto scroll-smooth px-[22px] py-[18px]"
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
												capabilities={channel?.capabilities}
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

					{/* FAB прокрутки вниз */}
					{!isAtBottom && messages.length > 0 && (
						<button
							type="button"
							onClick={scrollToBottom}
							aria-label={
								unreadBelowCount > 0
									? `Вниз, ${unreadBelowCount} новых`
									: "К последним сообщениям"
							}
							className={cn(
								"absolute bottom-20 right-6 z-20",
								"flex items-center justify-center",
								"h-10 w-10 rounded-full",
								"bg-card/90 backdrop-blur-md border border-border/60 shadow-lg",
								"text-t1 hover:bg-accent",
								"transition-colors animate-fade-in",
							)}
						>
							<svg
								aria-hidden="true"
								className="h-5 w-5"
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
								<span className="absolute -top-1.5 -right-1.5 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground px-1">
									{unreadBelowCount > 99 ? "99+" : unreadBelowCount}
								</span>
							)}
						</button>
					)}

					{/* Индикатор набора текста */}
					{typingText && (
						<div className="shrink-0 px-[22px] py-1.5 text-xs text-[var(--t3)] flex items-center gap-1.5">
							<span className="flex items-center gap-0.5">
								<span className="typing-dot" />
								<span className="typing-dot" />
								<span className="typing-dot" />
							</span>
							<span>{typingText}...</span>
						</div>
					)}

					{/* Композер сообщений */}
					{canPost && (
						<div className="px-[22px] pb-[18px] pt-3 shrink-0">
							<MessageComposer
								channelId={channelId}
								canSend={canPost}
								placeholder={
									channel
										? `Написать в ${channel.channel_type === "dm" ? channel.name : `#${channel.name}`}...`
										: "Написать сообщение..."
								}
							/>
						</div>
					)}
				</div>

				{/* Панель треда */}
				{activeThread?.channelId === channelId && (
					<div
						role="dialog"
						aria-label="Тред"
						tabIndex={-1}
						onKeyDown={(e) => {
							if (e.key === "Escape") {
								e.stopPropagation();
								closeThread();
							}
						}}
						className={cn(
							"w-96 shrink-0 border-l outline-none",
							"max-md:fixed max-md:inset-y-0 max-md:right-0 max-md:z-50 max-md:w-full max-md:bg-base",
						)}
					>
						<ThreadPanel
							channelId={channelId}
							canReply={canReply}
							canPost={canPost}
							messageId={activeThread.rootId}
							projectId={channel?.project_id}
							onClose={closeThread}
						/>
					</div>
				)}
			</div>
		</div>
	);
}
