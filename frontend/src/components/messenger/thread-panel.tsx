import { useQuery } from "@tanstack/react-query";
import type { Message } from "@/api/messenger";
import { getMessages } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useThreadMessages } from "@/hooks/use-messages";
import { MessageComposer } from "./message-composer";
import { MessageItem } from "./message-item";

/** Пропсы компонента ThreadPanel. */
interface ThreadPanelProps {
	/** ID канала. */
	channelId: string;
	/** ID сообщения (родительского). */
	messageId: string;
	/** Колбэк закрытия панели. */
	onClose: () => void;
}

/**
 * ThreadPanel -- панель просмотра и ответа в тред.
 * Показывает исходное сообщение, список ответов и composer для нового ответа.
 */
export function ThreadPanel({
	channelId,
	messageId,
	onClose,
}: ThreadPanelProps) {
	const { messages: replies, isLoading } = useThreadMessages(
		channelId,
		messageId,
	);

	// Загрузка родительского сообщения.
	const { data: parentMessages } = useQuery({
		queryKey: ["messenger", "parent-message", channelId, messageId],
		queryFn: () => getMessages(channelId, { limit: 1 }),
		enabled: !!channelId && !!messageId,
	});

	// Ищем родительское сообщение из кеша или первого запроса.
	const parentMessage =
		parentMessages?.find((m: Message) => m.id === messageId) ?? null;

	return (
		<div className="flex h-full flex-col">
			{/* Заголовок */}
			<div className="flex h-12 shrink-0 items-center justify-between border-b px-4">
				<span className="text-sm font-semibold">Тред</span>
				<Button
					variant="ghost"
					size="icon"
					className="h-7 w-7"
					onClick={onClose}
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
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</Button>
			</div>

			{/* Исходное сообщение */}
			{parentMessage && (
				<div className="border-b bg-raised/30">
					<MessageItem message={parentMessage} channelId={channelId} />
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
					<div>
						{replies.map((reply) => (
							<MessageItem
								key={reply.id}
								message={reply}
								channelId={channelId}
							/>
						))}
					</div>
				)}
			</ScrollArea>

			{/* Composer для ответа в тред */}
			<div className="shrink-0 border-t p-3">
				<MessageComposer
					channelId={channelId}
					threadId={messageId}
					placeholder="Ответить в тред..."
				/>
			</div>
		</div>
	);
}
