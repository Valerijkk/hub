import {
	Bookmark,
	Forward,
	MessageCircle,
	Pencil,
	Pin,
	Reply,
	Trash2,
} from "lucide-react";
import type { ChannelCapabilities, Message } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ReactionPicker } from "./reaction-picker";

/** Пропсы плавающей панели действий над сообщением. */
interface MessageActionsProps {
	message: Message;
	channelId: string;
	isOwn: boolean;
	capabilities?: ChannelCapabilities;
	showReactionPicker: boolean;
	onReactionPickerChange: (open: boolean) => void;
	onReply: () => void;
	onOpenThread: () => void;
	onPin: () => void;
	onBookmark: () => void;
	onForward: () => void;
	onEdit: () => void;
	onDelete: () => void;
	deletePending?: boolean;
}

function ActionButton({
	label,
	onClick,
	children,
	destructive = false,
	disabled = false,
}: {
	label: string;
	onClick: () => void;
	children: React.ReactNode;
	destructive?: boolean;
	disabled?: boolean;
}) {
	return (
		<Button
			type="button"
			variant="ghost"
			size="icon"
			className={cn(
				"h-7 w-7 rounded-full hover:bg-raised",
				destructive &&
					"text-destructive hover:bg-destructive/10 hover:text-destructive",
			)}
			title={label}
			aria-label={label}
			onClick={onClick}
			disabled={disabled}
		>
			{children}
		</Button>
	);
}

/** Плавающая Slack-like панель, строго проецирующая server capabilities. */
export function MessageActions({
	message,
	channelId,
	isOwn,
	capabilities,
	showReactionPicker,
	onReactionPickerChange,
	onReply,
	onOpenThread,
	onPin,
	onBookmark,
	onForward,
	onEdit,
	onDelete,
	deletePending = false,
}: MessageActionsProps) {
	return (
		<div
			className={cn(
				"absolute -top-3 right-4 z-10",
				"flex items-center gap-0.5 rounded-full",
				"border border-border/60 bg-card/90 backdrop-blur-md",
				"px-1 py-0.5 shadow-lg animate-fade-in",
			)}
		>
			{capabilities?.can_reply === true && (
				<ActionButton label="Ответить" onClick={onReply}>
					<Reply className="h-3.5 w-3.5" />
				</ActionButton>
			)}
			{capabilities?.can_react === true && (
				<ReactionPicker
					channelId={channelId}
					messageId={message.id}
					canReact
					open={showReactionPicker}
					onOpenChange={onReactionPickerChange}
				/>
			)}
			{capabilities?.can_read === true && (
				<ActionButton label="Открыть тред" onClick={onOpenThread}>
					<MessageCircle className="h-3.5 w-3.5" />
				</ActionButton>
			)}
			{capabilities?.can_pin === true && (
				<ActionButton label="Закрепить сообщение" onClick={onPin}>
					<Pin className="h-3.5 w-3.5" />
				</ActionButton>
			)}
			{capabilities?.can_bookmark === true && (
				<ActionButton label="Сохранить в закладки" onClick={onBookmark}>
					<Bookmark className="h-3.5 w-3.5" />
				</ActionButton>
			)}
			{capabilities?.can_forward === true && (
				<ActionButton label="Переслать сообщение" onClick={onForward}>
					<Forward className="h-3.5 w-3.5" />
				</ActionButton>
			)}
			{isOwn && capabilities?.can_edit === true && (
				<ActionButton label="Редактировать сообщение" onClick={onEdit}>
					<Pencil className="h-3.5 w-3.5" />
				</ActionButton>
			)}
			{((isOwn && capabilities?.is_member === true) ||
			capabilities?.can_delete_message === true) && (
				<ActionButton
					label="Удалить сообщение"
					onClick={onDelete}
					destructive
					disabled={deletePending}
				>
					<Trash2 className="h-3.5 w-3.5" />
				</ActionButton>
			)}
		</div>
	);
}
