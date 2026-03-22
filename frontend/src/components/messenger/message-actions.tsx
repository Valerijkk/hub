import type { Message } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ReactionPicker } from "./reaction-picker";

/** Пропсы плавающей панели действий над сообщением. */
interface MessageActionsProps {
	message: Message;
	channelId: string;
	isOwn: boolean;
	showReactionPicker: boolean;
	onReactionPickerChange: (open: boolean) => void;
	onReply: () => void;
	onOpenThread: () => void;
	onPin: () => void;
	onBookmark: () => void;
	onForward: () => void;
	onDelete: () => void;
}

/** Плавающие кнопки действий при наведении на сообщение (ответ, тред, закрепить, сохранить, переслать, удалить). */
export function MessageActions({
	message,
	channelId,
	isOwn,
	showReactionPicker,
	onReactionPickerChange,
	onReply,
	onOpenThread,
	onPin,
	onBookmark,
	onForward,
	onDelete,
}: MessageActionsProps) {
	return (
		<div
			className={cn(
				"absolute -top-3 flex items-center gap-0.5 rounded-lg border bg-card p-1 shadow-lg animate-fade-in z-10",
				isOwn ? "left-4" : "right-4",
			)}
		>
			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7"
				title="Ответить"
				onClick={onReply}
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
			</Button>

			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7"
				title="Тред"
				onClick={onOpenThread}
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
						d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
					/>
				</svg>
			</Button>

			<ReactionPicker
				channelId={channelId}
				messageId={message.id}
				open={showReactionPicker}
				onOpenChange={onReactionPickerChange}
			/>

			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7"
				title="Закрепить"
				onClick={onPin}
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
						d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
					/>
				</svg>
			</Button>

			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7"
				title="Сохранить"
				onClick={onBookmark}
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
						d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
					/>
				</svg>
			</Button>

			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7"
				title="Переслать"
				onClick={onForward}
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
						d="M13 5l7 7-7 7M5 5l7 7-7 7"
					/>
				</svg>
			</Button>

			{isOwn && (
				<Button
					variant="ghost"
					size="icon"
					className="h-7 w-7 text-destructive hover:text-destructive"
					title="Удалить"
					onClick={onDelete}
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
							d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
						/>
					</svg>
				</Button>
			)}
		</div>
	);
}
