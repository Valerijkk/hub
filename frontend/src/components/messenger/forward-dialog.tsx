import { useCallback, useState } from "react";
import {
	canChannel,
	type Channel,
	forwardMessage,
	type Message,
} from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useMessengerStore } from "@/stores/messenger-store";

interface ForwardDialogProps {
	message: Message;
	channelId: string;
	canForward: boolean;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

/** Диалог пересылки: source CanForward, target CanPost, оба не revoked. */
export function ForwardDialog({
	message,
	channelId,
	canForward,
	open,
	onOpenChange,
}: ForwardDialogProps) {
	const channels = useMessengerStore((state) => state.channels);
	const revokedChannelIds = useMessengerStore(
		(state) => state.revokedChannelIds,
	);
	const [isForwarding, setIsForwarding] = useState(false);
	const sourceRevoked = revokedChannelIds.has(channelId);

	const canForwardNow = useCallback(
		(targetChannelId: string) => {
			const state = useMessengerStore.getState();
			if (
				!canForward ||
				state.isChannelRevoked(channelId) ||
				state.isChannelRevoked(targetChannelId)
			) {
				return false;
			}
			const source = state.channels.find((item) => item.id === channelId);
			const target = state.channels.find(
				(item) => item.id === targetChannelId,
			);
			return (
				source !== undefined &&
				target !== undefined &&
				canChannel(source, "can_forward") &&
				canChannel(target, "can_post")
			);
		},
		[canForward, channelId],
	);

	const handleForward = useCallback(
		async (targetChannel: Channel) => {
			if (isForwarding || !canForwardNow(targetChannel.id)) return;
			setIsForwarding(true);
			try {
				await forwardMessage(channelId, message.id, targetChannel.id);
				if (!canForwardNow(targetChannel.id)) return;
				onOpenChange(false);
			} finally {
				setIsForwarding(false);
			}
		},
		[
			canForwardNow,
			channelId,
			isForwarding,
			message.id,
			onOpenChange,
		],
	);

	const targetChannels = channels.filter(
		(channel) =>
			channel.id !== channelId &&
			!revokedChannelIds.has(channel.id) &&
			canChannel(channel, "can_post"),
	);
	const effectiveOpen = open && canForward && !sourceRevoked;

	return (
		<Dialog
			open={effectiveOpen}
			onOpenChange={(nextOpen) => {
				if (nextOpen && (!canForward || sourceRevoked)) return;
				onOpenChange(nextOpen);
			}}
		>
			<DialogContent className="max-w-sm">
				<DialogHeader>
					<DialogTitle>Переслать сообщение</DialogTitle>
				</DialogHeader>
				<div className="space-y-3 pt-2">
					<div className="rounded-md bg-raised p-3">
						<p className="text-xs font-medium text-t2">{message.sender_id}</p>
						<p className="mt-1 line-clamp-3 text-sm">
							{message.content_text || "---"}
						</p>
					</div>
					<ScrollArea className="max-h-60">
						<div className="space-y-1">
							{targetChannels.map((channel) => (
								<Button
									key={channel.id}
									variant="ghost"
									className="w-full justify-start text-sm"
									disabled={isForwarding}
									onClick={() => void handleForward(channel)}
								>
									{channel.name || "DM"}
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