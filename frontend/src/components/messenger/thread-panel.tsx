import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import type { ChannelCapabilities, Message } from "@/api/messenger";
import { getThread } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { threadDetailQueryKey, useThreadMessages } from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { MessageComposer } from "./message-composer";
import { MessageItem } from "./message-item";
import { ThreadFollowControl } from "./thread-follow-control";
import { ThreadIssueLinks } from "./thread-issue-links";
import { ThreadReadReceipts } from "./thread-read-receipts";

/** Пропсы компонента ThreadPanel. */
interface ThreadPanelProps {
	/** ID канала. */
	channelId: string;
	/** ID сообщения (родительского). */
	messageId: string;
	/** Fresh право отвечать в тред. */
	canReply?: boolean;
	/** Fresh право вывести ответ в основную ленту. */
	canPost?: boolean;
	/** Серверные права действий над сообщениями; отсутствие запрещает действия. */
	capabilities?: ChannelCapabilities;
	/** ID проекта для issue↔thread links (optional). */
	projectId?: string;
	/** Колбэк закрытия панели. */
	onClose: () => void;
}

/** Ищет сообщение в infinite-cache ленты канала. */
export function findParentInMessagesCache(
	cache: { pages: Message[][] } | undefined,
	messageId: string,
): Message | null {
	if (!cache?.pages) return null;
	for (const page of cache.pages) {
		const found = page.find((m) => m.id === messageId);
		if (found) return found;
	}
	return null;
}

/**
 * ThreadPanel -- панель просмотра и ответа в тред.
 * Показывает исходное сообщение, список ответов и composer для нового ответа.
 */
export function ThreadPanel({
	channelId,
	messageId,
	canReply = false,
	canPost = false,
	capabilities,
	projectId,
	onClose,
}: ThreadPanelProps) {
	const userId = useAuthStore((state) => state.user?.id ?? null);
	const isRevoked = useMessengerStore((state) =>
		state.revokedChannelIds.has(channelId),
	);
	const returnFocusRef = useRef<HTMLElement | null>(null);
	const closeButtonRef = useRef<HTMLButtonElement>(null);
	const replyContainerRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		returnFocusRef.current =
			document.activeElement instanceof HTMLElement
				? document.activeElement
				: null;
		closeButtonRef.current?.focus();
	}, []);

	const parentQuery = useQuery({
		queryKey: threadDetailQueryKey(
			channelId,
			messageId,
			userId ?? "__no-user__",
		),
		queryFn: () => getThread(channelId, messageId),
		enabled: !!channelId && !!messageId && !!userId && !isRevoked,
	});
	const threadMessagesQuery = useThreadMessages(channelId, messageId);
	const replies = threadMessagesQuery.messages;
	const isRepliesLoading = threadMessagesQuery.isLoading;
	const detailReady =
		parentQuery.isSuccess &&
		!parentQuery.isError &&
		!parentQuery.isFetching &&
		Boolean(parentQuery.data);
	const repliesReady =
		threadMessagesQuery.isSuccess &&
		!threadMessagesQuery.isError &&
		!threadMessagesQuery.isFetching &&
		Boolean(threadMessagesQuery.data);
	const parentMessage =
		userId && !isRevoked ? (parentQuery.data?.parent ?? null) : null;
	const isLoading = userId && !isRevoked ? isRepliesLoading : false;

	const handleClose = () => {
		const returnTarget = returnFocusRef.current;
		onClose();
		returnTarget?.focus();
	};

	return (
		<div className="flex h-full flex-col">
			{/* Заголовок */}
			<div className="flex h-12 shrink-0 items-center justify-between border-b px-4">
				<div className="flex items-center gap-2">
					<span className="text-sm font-semibold">Тред</span>
					{userId &&
						!isRevoked &&
						parentQuery.isSuccess &&
						!parentQuery.isError &&
						parentQuery.data && (
							<ThreadFollowControl
								key={JSON.stringify([channelId, messageId, userId])}
								channelId={channelId}
								rootMessageId={messageId}
								userId={userId}
								thread={parentQuery.data}
								detailIsSuccess={parentQuery.isSuccess}
								detailIsError={parentQuery.isError}
								detailIsFetching={parentQuery.isFetching}
							/>
						)}
				</div>
				<Button
					ref={closeButtonRef}
					variant="ghost"
					size="icon"
					className="h-7 w-7"
					aria-label="Закрыть тред"
					onClick={handleClose}
				>
					<svg
						aria-hidden="true"
						className="h-4 w-4"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</Button>
			</div>

			{projectId && userId && !isRevoked && (
				<ThreadIssueLinks
					projectId={projectId}
					channelId={channelId}
					messageId={messageId}
				/>
			)}

			{/* Исходное сообщение */}
			{parentMessage && (
				<div className="border-b bg-raised/30">
					<MessageItem
						message={parentMessage}
						channelId={channelId}
						capabilities={capabilities}
					/>
				</div>
			)}

			{/* Ответы */}
			<ScrollArea className="flex-1">
				{isLoading ? (
					<div className="flex items-center justify-center p-4">
						<span className="text-sm text-t2">Загрузка...</span>
					</div>
				) : replies.length === 0 ? (
					<div className="flex items-center justify-center p-4">
						<span className="text-sm text-t2">Нет ответов</span>
					</div>
				) : (
					<div ref={replyContainerRef}>
						{replies.map((reply) => (
							<div key={reply.id} className="relative">
								<MessageItem
									message={reply}
									channelId={channelId}
									capabilities={capabilities}
								/>
								<span
									aria-hidden="true"
									className="pointer-events-none absolute bottom-0 left-0 h-px w-px"
									data-thread-read-marker=""
									data-message-id={reply.id}
								/>
							</div>
						))}
					</div>
				)}
			</ScrollArea>

			{userId &&
				!isRevoked &&
				parentQuery.isSuccess &&
				!parentQuery.isError &&
				parentQuery.data && (
					<ThreadReadReceipts
						key={JSON.stringify([channelId, messageId, userId])}
						channelId={channelId}
						rootMessageId={messageId}
						userId={userId}
						replies={replies}
						containerRef={replyContainerRef}
						detail={parentQuery.data}
						detailReady={detailReady}
						repliesReady={repliesReady}
					/>
				)}

			{/* Composer для ответа в тред */}
			{userId && !isRevoked && (
				<div className="shrink-0 border-t p-3">
					<MessageComposer
						channelId={channelId}
						canSend={canReply}
						canAlsoSendToChannel={canPost}
						threadId={messageId}
						placeholder="Ответить в тред..."
					/>
				</div>
			)}
		</div>
	);
}
