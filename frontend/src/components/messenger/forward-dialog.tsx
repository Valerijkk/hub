import { useCallback, useState } from "react";
import { type Channel, forwardMessage, type Message } from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMessengerStore } from "@/stores/messenger-store";

/** Пропсы компонента ForwardDialog. */
interface ForwardDialogProps {
	/** Сообщение для пересылки. */
	message: Message;
	/** ID канала-источника. */
	channelId: string;
	/** Открыто ли окно. */
	open: boolean;
	/** Обработчик изменения видимости. */
	onOpenChange: (open: boolean) => void;
}

/**
 * ForwardDialog -- диалог пересылки сообщения в другой канал.
 * Отображает список каналов пользователя для выбора цели.
 */
export function ForwardDialog({
	message,
	channelId,
	open,
	onOpenChange,
}: ForwardDialogProps) {
	const channels = useMessengerStore((s) => s.channels);
	const [isForwarding, setIsForwarding] = useState(false);

	const handleForward = useCallback(
		async (targetChannel: Channel) => {
			setIsForwarding(true);
			try {
				await forwardMessage(channelId, message.id, targetChannel.id);
				onOpenChange(false);
			} finally {
				setIsForwarding(false);
			}
		},
		[channelId, message.id, onOpenChange],
	);

	// Фильтруем текущий канал из списка целей.
	const targetChannels = channels.filter((ch) => ch.id !== channelId);

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-sm">
				<DialogHeader>
					<DialogTitle>Переслать сообщение</DialogTitle>
				</DialogHeader>
				<div className="space-y-3 pt-2">
					{/* Превью */}
					<div className="rounded-md bg-raised p-3">
						<p className="text-xs font-medium text-t2">{message.sender_id}</p>
						<p className="mt-1 line-clamp-3 text-sm">
							{message.content_text || "---"}
						</p>
					</div>

					{/* Список каналов */}
					<ScrollArea className="max-h-60">
						<div className="space-y-1">
							{targetChannels.map((ch) => (
								<Button
									key={ch.id}
									variant="ghost"
									className="w-full justify-start text-sm"
									disabled={isForwarding}
									onClick={() => void handleForward(ch)}
								>
									{ch.name || "DM"}
								</Button>
							))}
							{targetChannels.length === 0 && (
								<p className="py-4 text-center text-sm text-t2">
									Нет доступных каналов
								</p>
							)}
						</div>
					</ScrollArea>
				</div>
			</DialogContent>
		</Dialog>
	);
}
