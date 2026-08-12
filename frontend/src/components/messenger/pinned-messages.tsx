import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import {
	getChannelPins,
	type Message,
	type PinnedMessage,
	togglePin,
} from "@/api/messenger";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";

/** Пропсы компонента PinnedMessages. */
interface PinnedMessagesProps {
	/** ID канала. */
	channelId: string;
}

/**
 * Preview текста пина: content_text из кэша сообщений канала, иначе нейтральный label.
 * A3W9-3-002: не показывать UUID как единственный смысл строки.
 */
export function resolvePinPreview(
	pages: Message[][] | undefined,
	messageId: string,
): string {
	if (pages) {
		for (const page of pages) {
			const msg = page.find((m) => m.id === messageId);
			const text = msg?.content_text?.trim();
			if (text) {
				return text.length > 120 ? `${text.slice(0, 120)}…` : text;
			}
		}
	}
	return "Закреплённое сообщение";
}

/**
 * PinnedMessages -- панель закреплённых сообщений канала.
 * Открывается кнопкой pin в заголовке чата.
 */
export function PinnedMessages({ channelId }: PinnedMessagesProps) {
	const [open, setOpen] = useState(false);
	const queryClient = useQueryClient();

	const { data: pins, refetch } = useQuery({
		queryKey: ["messenger", "pins", channelId],
		queryFn: () => getChannelPins(channelId),
		enabled: open,
	});

	const messagesCache = queryClient.getQueryData<{ pages: Message[][] }>([
		"messenger",
		"messages",
		channelId,
	]);

	const handleUnpin = useCallback(
		async (pin: PinnedMessage) => {
			await togglePin(pin.channel_id, pin.message_id);
			await refetch();
		},
		[refetch],
	);

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button
					variant="ghost"
					size="icon"
					className="h-8 w-8"
					title="Закреплённые"
					aria-label="Закреплённые сообщения"
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
							d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
						/>
					</svg>
				</Button>
			</DialogTrigger>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle>Закреплённые сообщения</DialogTitle>
				</DialogHeader>
				<ScrollArea className="max-h-96">
					{pins && pins.length > 0 ? (
						<div className="space-y-2 pt-2">
							{pins.map((pin) => {
								const preview = resolvePinPreview(
									messagesCache?.pages,
									pin.message_id,
								);
								return (
									<div
										key={pin.id}
										className="flex items-start justify-between rounded-md border p-3"
									>
										<div className="min-w-0 flex-1">
											<p className="text-sm line-clamp-2" title={preview}>
												{preview}
											</p>
											<p className="text-xs text-t2">
												Закреплено:{" "}
												{new Date(pin.pinned_at).toLocaleDateString("ru-RU")}
											</p>
										</div>
										<Button
											variant="ghost"
											size="sm"
											className="shrink-0"
											onClick={() => void handleUnpin(pin)}
										>
											Открепить
										</Button>
									</div>
								);
							})}
						</div>
					) : (
						<p className="py-4 text-center text-sm text-t2">
							Нет закреплённых сообщений
						</p>
					)}
				</ScrollArea>
			</DialogContent>
		</Dialog>
	);
}
