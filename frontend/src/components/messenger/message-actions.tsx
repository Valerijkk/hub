import type { Message } from "@/api/messenger";
import { Pencil } from "lucide-react";
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
	onEdit: () => void;
	onDelete: () => void;
}

/**
 * Плавающая горизонтальная панель действий при наведении на сообщение.
 * Отображается справа от сообщения в виде pill-shaped toolbar с glass-эффектом.
 */
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
	onEdit,
	onDelete,
}: MessageActionsProps) {
	return (
		<div
			className={cn(
				"absolute -top-3 right-4 z-10",
				"flex items-center gap-0.5 rounded-full",
				"border border-border/60 bg-card/90 backdrop-blur-md",
				"px-1 py-0.5 shadow-lg",
				"animate-fade-in",
			)}
		>
			{/* Ответить */}
			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7 rounded-full hover:bg-raised"
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

			{/* Реакция */}
			<ReactionPicker
				channelId={channelId}
				messageId={message.id}
				open={showReactionPicker}
				onOpenChange={onReactionPickerChange}
			/>

			{/* Тред */}
			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7 rounded-full hover:bg-raised"
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

			{/* Закрепить */}
			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7 rounded-full hover:bg-raised"
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

			{/* Сохранить */}
			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7 rounded-full hover:bg-raised"
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

			{/* Переслать */}
			<Button
				variant="ghost"
				size="icon"
				className="h-7 w-7 rounded-full hover:bg-raised"
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

			{/* Редактировать (только свои) */}
			{isOwn && (
				<Button
					variant="ghost"
					size="icon"
					className="h-7 w-7 rounded-full hover:bg-raised"
					title="Редактировать"
					onClick={onEdit}
				>
					<Pencil className="h-3.5 w-3.5" />
				</Button>
			)}

			{/* Удалить (только свои) */}
			{isOwn && (
				<Button
					variant="ghost"
					size="icon"
					className="h-7 w-7 rounded-full text-destructive hover:text-destructive hover:bg-destructive/10"
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
